package importar

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"syscall"
	"time"
)

// Leer una tienda pública (Shopify, WooCommerce, Empretienda) desde el servidor: la URL la escribe
// la dueña, así que el cliente sólo sale a dominios públicos por https. La IP se revisa al conectar
// (no al resolver el nombre), para que un DNS que cambia entre una consulta y otra no la saltee.

// ErrNoPublica: la dirección no es una tienda pública a la que se pueda ir.
var ErrNoPublica = errors.New("importar: la dirección no es pública")

// userAgent es en ASCII: algunas tiendas cortan los pedidos con otro.
const userAgent = "Mostrador (importacion de la tienda de la duena; soporte@mostradorapp.com.ar)"

// TamanioMax es lo más que se lee de una respuesta.
const TamanioMax = 8 << 20

// Origen valida lo que escribió la dueña y devuelve el origen https ("https://tienda.com"). Sin IPs,
// sin puerto, sin usuario y con un punto en el nombre.
func Origen(entrada string) (string, error) {
	s := strings.TrimSpace(entrada)
	if !strings.Contains(s, "://") {
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.User != nil || u.Port() != "" {
		return "", ErrNoPublica
	}
	h := strings.TrimSuffix(strings.ToLower(u.Hostname()), ".")
	if h == "" || !strings.Contains(h, ".") || h == "localhost" || strings.HasSuffix(h, ".localhost") || strings.HasSuffix(h, ".local") {
		return "", ErrNoPublica
	}
	if _, err := netip.ParseAddr(h); err == nil {
		return "", ErrNoPublica
	}
	return "https://" + h, nil
}

// publica: ni privada, ni loopback, ni link-local, ni multicast, ni sin especificar.
func publica(a netip.Addr) bool {
	a = a.Unmap()
	return a.IsValid() && a.IsGlobalUnicast() && !a.IsPrivate() && !a.IsLoopback() && !a.IsLinkLocalUnicast() &&
		// 100.64.0.0/10 (CGNAT) tampoco es Internet.
		!netip.MustParsePrefix("100.64.0.0/10").Contains(a)
}

// Publico es el cliente para leer tiendas públicas: https, a IPs públicas, con tiempo y redirecciones
// acotadas (cada redirección vuelve a pasar por la misma revisión).
func Publico() *http.Client {
	d := &net.Dialer{
		Timeout: 10 * time.Second,
		Control: func(_, direccion string, _ syscall.RawConn) error {
			ap, err := netip.ParseAddrPort(direccion)
			if err != nil || !publica(ap.Addr()) {
				return ErrNoPublica
			}
			return nil
		},
	}
	tr := &http.Transport{
		DialContext:           d.DialContext,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 20 * time.Second,
		MaxIdleConnsPerHost:   2,
	}
	return &http.Client{
		Transport: tr,
		Timeout:   40 * time.Second,
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			if len(via) >= 3 || r.URL.Scheme != "https" || r.URL.Port() != "" {
				return ErrNoPublica
			}
			return nil
		},
	}
}

// ErrHTTP es una respuesta no exitosa de la tienda.
type ErrHTTP struct{ Estado int }

func (e *ErrHTTP) Error() string { return fmt.Sprintf("importar: la tienda respondió HTTP %d", e.Estado) }

// Pedir lee una URL con el cliente dado y devuelve el cuerpo y los headers. Una respuesta que no es
// 200 vuelve como *ErrHTTP.
func Pedir(ctx context.Context, c *http.Client, u string) ([]byte, http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json")
	res, err := c.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer res.Body.Close()
	b, err := io.ReadAll(io.LimitReader(res.Body, TamanioMax+1))
	if err != nil {
		return nil, nil, err
	}
	if len(b) > TamanioMax {
		return nil, nil, fmt.Errorf("importar: la respuesta de la tienda es demasiado grande")
	}
	if res.StatusCode != http.StatusOK {
		return nil, res.Header, &ErrHTTP{Estado: res.StatusCode}
	}
	return b, res.Header, nil
}
