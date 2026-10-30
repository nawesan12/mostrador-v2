// Package cloudflare: lo justo de la API de Cloudflare para servir una tienda en su dominio
// propio. El dominio es una zona de la cuenta de Grupo Vimo (los nameservers que da Cloudflare se
// cargan en NIC.ar); cuando la zona está activa, el dominio y su www apuntan al Worker de la web
// con una route (no un custom domain: `wrangler deploy` borra los custom domains que no están en
// su config, y las routes las deja).
package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Cliente struct {
	Token, Cuenta string
	// Base: la de la API (otra en los tests).
	Base string
	HTTP *http.Client
}

func Nuevo(token, cuenta string) *Cliente {
	return &Cliente{Token: token, Cuenta: cuenta, Base: "https://api.cloudflare.com/client/v4",
		HTTP: &http.Client{Timeout: 20 * time.Second}}
}

type Zona struct {
	ID          string   `json:"id"`
	Nombre      string   `json:"name"`
	Estado      string   `json:"status"`
	Nameservers []string `json:"name_servers"`
}

type respuesta struct {
	Success bool `json:"success"`
	Errors  []struct {
		Code    int
		Message string
	} `json:"errors"`
	Result json.RawMessage `json:"result"`
}

// ErrAPI: lo que Cloudflare contestó como error (con su código, p. ej. 1061 = la zona ya existe).
type ErrAPI struct {
	Codigo  int
	Mensaje string
}

func (e ErrAPI) Error() string { return fmt.Sprintf("cloudflare %d: %s", e.Codigo, e.Mensaje) }

func (c *Cliente) pedir(ctx context.Context, metodo, path string, cuerpo, out any) error {
	var b []byte
	if cuerpo != nil {
		var err error
		if b, err = json.Marshal(cuerpo); err != nil {
			return err
		}
	}
	req, err := http.NewRequestWithContext(ctx, metodo, c.Base+path, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	res, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	var r respuesta
	if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
		return fmt.Errorf("cloudflare %s %s: %d", metodo, path, res.StatusCode)
	}
	if !r.Success {
		if len(r.Errors) > 0 {
			return ErrAPI{r.Errors[0].Code, r.Errors[0].Message}
		}
		return ErrAPI{res.StatusCode, "sin detalle"}
	}
	if out != nil {
		return json.Unmarshal(r.Result, out)
	}
	return nil
}

// ZonaDe: la zona de la cuenta que contiene al dominio (la del dominio o la de un padre), o nil.
func (c *Cliente) ZonaDe(ctx context.Context, dominio string) (*Zona, error) {
	partes := strings.Split(dominio, ".")
	for i := 0; i+1 < len(partes); i++ {
		var zs []Zona
		q := url.Values{"name": {strings.Join(partes[i:], ".")}, "account.id": {c.Cuenta}}
		if err := c.pedir(ctx, http.MethodGet, "/zones?"+q.Encode(), nil, &zs); err != nil {
			return nil, err
		}
		if len(zs) > 0 {
			return &zs[0], nil
		}
	}
	return nil, nil
}

// CrearZona agrega el dominio a la cuenta (plan gratis); si ya estaba, devuelve la que hay.
func (c *Cliente) CrearZona(ctx context.Context, dominio string) (Zona, error) {
	var z Zona
	err := c.pedir(ctx, http.MethodPost, "/zones", map[string]any{
		"name": dominio, "account": map[string]string{"id": c.Cuenta}, "type": "full"}, &z)
	if e, ok := err.(ErrAPI); ok && e.Codigo == 1061 {
		ya, err := c.ZonaDe(ctx, dominio)
		if err != nil || ya == nil {
			return Zona{}, fmt.Errorf("la zona %s ya existe en otra cuenta de Cloudflare", dominio)
		}
		return *ya, nil
	}
	return z, err
}

type registro struct {
	ID      string `json:"id,omitempty"`
	Tipo    string `json:"type"`
	Nombre  string `json:"name"`
	Valor   string `json:"content"`
	Proxied bool   `json:"proxied"`
}

type route struct {
	ID     string `json:"id,omitempty"`
	Patron string `json:"pattern"`
	Script string `json:"script"`
}

// Apuntar deja `host` sirviéndose del Worker `script`: un AAAA a 100:: con proxy (no va a ningún
// origen: lo atiende el Worker) y la route host/*. Se puede correr de nuevo: lo que ya está queda.
func (c *Cliente) Apuntar(ctx context.Context, zona, host, script string) error {
	var rs []registro
	if err := c.pedir(ctx, http.MethodGet, "/zones/"+zona+"/dns_records?"+url.Values{"name": {host}}.Encode(), nil, &rs); err != nil {
		return err
	}
	if len(rs) == 0 {
		if err := c.pedir(ctx, http.MethodPost, "/zones/"+zona+"/dns_records",
			registro{Tipo: "AAAA", Nombre: host, Valor: "100::", Proxied: true}, nil); err != nil {
			return err
		}
	}
	var routes []route
	if err := c.pedir(ctx, http.MethodGet, "/zones/"+zona+"/workers/routes", nil, &routes); err != nil {
		return err
	}
	patron := host + "/*"
	for _, r := range routes {
		if r.Patron == patron {
			if r.Script == script {
				return nil
			}
			return c.pedir(ctx, http.MethodPut, "/zones/"+zona+"/workers/routes/"+r.ID, route{Patron: patron, Script: script}, nil)
		}
	}
	return c.pedir(ctx, http.MethodPost, "/zones/"+zona+"/workers/routes", route{Patron: patron, Script: script}, nil)
}

// Desapuntar saca la route y el registro que puso Apuntar (si están).
func (c *Cliente) Desapuntar(ctx context.Context, zona, host string) error {
	var routes []route
	if err := c.pedir(ctx, http.MethodGet, "/zones/"+zona+"/workers/routes", nil, &routes); err != nil {
		return err
	}
	for _, r := range routes {
		if r.Patron == host+"/*" {
			if err := c.pedir(ctx, http.MethodDelete, "/zones/"+zona+"/workers/routes/"+r.ID, nil, nil); err != nil {
				return err
			}
		}
	}
	var rs []registro
	if err := c.pedir(ctx, http.MethodGet, "/zones/"+zona+"/dns_records?"+url.Values{"name": {host}, "type": {"AAAA"}}.Encode(), nil, &rs); err != nil {
		return err
	}
	for _, r := range rs {
		if r.Valor == "100::" {
			if err := c.pedir(ctx, http.MethodDelete, "/zones/"+zona+"/dns_records/"+r.ID, nil, nil); err != nil {
				return err
			}
		}
	}
	return nil
}
