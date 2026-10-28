// Package tiendanube: importar una tienda de Tiendanube a Mostrador con su API oficial.
//
//   - La tienda autoriza la app de socio de Mostrador (OAuth sin PKCE; el state es de un solo uso y
//     lo guarda quien llama). El token no vence: sirve hasta que desinstalan la app o piden otro.
//   - Se lee de a páginas de 200 (productos, clientes y pedidos), siguiendo el header Link, con el
//     límite de la API: un balde de 40 pedidos que se vacía a 2 por segundo, por tienda.
//   - convertir.go y clientes.go pasan cada recurso a lo que guarda Mostrador; el caso compartido con
//     la web está en packages/contrato/casos/importar/tiendanube.json.
package tiendanube

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Errores que el panel traduce.
var (
	ErrApagado     = errors.New("tiendanube: sin credenciales de la aplicación")
	ErrReautorizar = errors.New("tiendanube: hay que volver a conectar la tienda")
	ErrInvalido    = errors.New("tiendanube: respuesta inválida")
)

// Versión de la API: va en la ruta y se cambia a mano cuando Tiendanube saque otra.
const version = "2025-03"

// Las URLs se pueden cambiar en los tests.
var (
	base       = "https://api.tiendanube.com/" + version
	autorizar  = "https://www.tiendanube.com/apps/%s/authorize"
	urlToken   = "https://www.tiendanube.com/apps/authorize/token"
	userAgent  = "Mostrador (soporte@mostradorapp.com.ar)"
	esperaMax  = 30 * time.Second
	tamanioMax = int64(20 << 20)
)

// App son las credenciales de la app de socio (MOSTRADOR_TN_CLIENT_ID y MOSTRADOR_TN_SECRETO).
type App struct {
	ID, Secreto string
	HTTP        *http.Client
}

func (a App) http() *http.Client {
	if a.HTTP != nil {
		return a.HTTP
	}
	return &http.Client{Timeout: 30 * time.Second}
}

// URLAutorizar es adonde se manda a la dueña para que autorice la app.
func (a App) URLAutorizar(state string) (string, error) {
	if a.ID == "" || a.Secreto == "" {
		return "", ErrApagado
	}
	return fmt.Sprintf(autorizar, url.PathEscape(a.ID)) + "?state=" + url.QueryEscape(state), nil
}

// Token es lo que devuelve el canje del código: el token y el id de la tienda (user_id).
type Token struct {
	Acceso   string
	TiendaID string
	Alcance  []string
}

// Canjear cambia el código de la vuelta del OAuth (dura 5 minutos) por el token de la tienda.
func (a App) Canjear(ctx context.Context, codigo string) (Token, error) {
	if a.ID == "" || a.Secreto == "" {
		return Token{}, ErrApagado
	}
	cuerpo, _ := json.Marshal(map[string]string{
		"client_id": a.ID, "client_secret": a.Secreto, "grant_type": "authorization_code", "code": codigo,
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, urlToken, bytes.NewReader(cuerpo))
	if err != nil {
		return Token{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", userAgent)
	res, err := a.http().Do(req)
	if err != nil {
		return Token{}, err
	}
	defer res.Body.Close()
	var r struct {
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
		UserID      any    `json:"user_id"`
		Error       string `json:"error"`
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&r); err != nil {
		return Token{}, ErrInvalido
	}
	// El id de la tienda llega como número o como texto, según la respuesta.
	id := strings.TrimSuffix(strings.TrimSpace(fmt.Sprint(r.UserID)), ".0")
	if res.StatusCode != http.StatusOK || r.AccessToken == "" || id == "" || id == "<nil>" {
		return Token{}, fmt.Errorf("tiendanube: no se pudo canjear el código (HTTP %d %s)", res.StatusCode, r.Error)
	}
	return Token{Acceso: r.AccessToken, TiendaID: id, Alcance: strings.Split(r.Scope, ",")}, nil
}

// API lee la API con el token de una tienda.
type API struct {
	Token, TiendaID string
	HTTP            *http.Client
	// Dormir espera entre reintentos; los tests lo reemplazan.
	Dormir func(context.Context, time.Duration) error
}

// Pagina es una página de un recurso: los registros crudos y la URL de la siguiente ("" si no hay).
type Pagina struct {
	Registros []json.RawMessage
	Siguiente string
	Total     int
	// Idioma principal de la tienda (X-Main-Language), para los campos por idioma.
	Idioma string
}

// Primera es la URL de la primera página de un recurso (products, customers, orders).
func (c API) Primera(recurso string, extra url.Values) string {
	q := url.Values{"per_page": {"200"}, "page": {"1"}}
	for k, v := range extra {
		q[k] = v
	}
	return fmt.Sprintf("%s/%s/%s?%s", base, url.PathEscape(c.TiendaID), recurso, q.Encode())
}

// Leer pide una página. Con 429 espera lo que dice la API y reintenta (hasta 3 veces); con 401,
// ErrReautorizar. Sólo sigue URLs de la API: el header Link no puede mandar el token a otro lado.
func (c API) Leer(ctx context.Context, u string) (Pagina, error) {
	if !strings.HasPrefix(u, base+"/"+c.TiendaID+"/") {
		return Pagina{}, fmt.Errorf("tiendanube: URL fuera de la API: %q", u)
	}
	dormir := c.Dormir
	if dormir == nil {
		dormir = esperar
	}
	hc := c.HTTP
	if hc == nil {
		hc = &http.Client{Timeout: 30 * time.Second}
	}
	for intento := 0; ; intento++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		if err != nil {
			return Pagina{}, err
		}
		req.Header.Set("Authorization", "Bearer "+c.Token)
		req.Header.Set("User-Agent", userAgent)
		res, err := hc.Do(req)
		if err != nil {
			return Pagina{}, err
		}
		cuerpo, err := io.ReadAll(io.LimitReader(res.Body, tamanioMax))
		res.Body.Close()
		if err != nil {
			return Pagina{}, err
		}
		switch {
		case res.StatusCode == http.StatusTooManyRequests && intento < 3:
			if err := dormir(ctx, espera(res.Header)); err != nil {
				return Pagina{}, err
			}
			continue
		case res.StatusCode == http.StatusUnauthorized || res.StatusCode == http.StatusForbidden:
			return Pagina{}, ErrReautorizar
		case res.StatusCode == http.StatusNotFound:
			// Una página después de la última responde 404 con "Last page is 0": no hay más.
			return Pagina{Registros: []json.RawMessage{}}, nil
		case res.StatusCode != http.StatusOK:
			return Pagina{}, fmt.Errorf("tiendanube: HTTP %d", res.StatusCode)
		}
		var regs []json.RawMessage
		if err := json.Unmarshal(cuerpo, &regs); err != nil {
			return Pagina{}, ErrInvalido
		}
		p := Pagina{Registros: regs, Idioma: res.Header.Get("X-Main-Language")}
		p.Total, _ = strconv.Atoi(res.Header.Get("X-Total-Count"))
		if sig := siguiente(res.Header.Get("Link")); strings.HasPrefix(sig, base+"/"+c.TiendaID+"/") {
			p.Siguiente = sig
		}
		return p, nil
	}
}

var reLink = regexp.MustCompile(`<([^>]+)>;\s*rel="next"`)

// siguiente saca la URL rel="next" del header Link.
func siguiente(link string) string {
	if m := reLink.FindStringSubmatch(link); m != nil {
		return m[1]
	}
	return ""
}

// espera es cuánto aguardar después de un 429: lo que falta para que se vacíe el balde
// (x-rate-limit-reset, en milisegundos), con un tope.
func espera(h http.Header) time.Duration {
	ms, err := strconv.Atoi(h.Get("X-Rate-Limit-Reset"))
	if err != nil || ms <= 0 {
		return time.Second
	}
	return min(time.Duration(ms)*time.Millisecond, esperaMax)
}

func esperar(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
