// Package cloudflare: lo justo de la API de Cloudflare para servir una tienda en su dominio
// propio. El dominio es una zona de la cuenta de Grupo Vimo (los nameservers que da Cloudflare se
// cargan en NIC.ar); cuando la zona está activa, el dominio y su www apuntan al VPS con un registro
// A con proxy, y Caddy hace el certificado de origen al primer pedido.
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

// Apuntar deja `host` sirviéndose desde el VPS: un registro A con proxy a `ip` (Caddy hace el
// certificado al primer pedido, si la API confirma que es el dominio de una tienda). Borra lo que
// haya quedado del esquema viejo (AAAA 100:: y la route del Worker). Se puede correr de nuevo.
func (c *Cliente) Apuntar(ctx context.Context, zona, host, ip string) error {
	if err := c.limpiarViejo(ctx, zona, host); err != nil {
		return err
	}
	var rs []registro
	if err := c.pedir(ctx, http.MethodGet, "/zones/"+zona+"/dns_records?"+url.Values{"name": {host}}.Encode(), nil, &rs); err != nil {
		return err
	}
	for _, r := range rs {
		if r.Tipo == "A" && r.Valor == ip && r.Proxied {
			return nil
		}
		if r.Tipo == "A" || r.Tipo == "AAAA" || r.Tipo == "CNAME" {
			return fmt.Errorf("%s ya apunta a otro lado (%s %s): sacalo a mano antes", host, r.Tipo, r.Valor)
		}
	}
	return c.pedir(ctx, http.MethodPost, "/zones/"+zona+"/dns_records", registro{Tipo: "A", Nombre: host, Valor: ip, Proxied: true}, nil)
}

// Desapuntar saca el registro que puso Apuntar (y lo del esquema viejo, si quedó).
func (c *Cliente) Desapuntar(ctx context.Context, zona, host, ip string) error {
	if err := c.limpiarViejo(ctx, zona, host); err != nil {
		return err
	}
	var rs []registro
	if err := c.pedir(ctx, http.MethodGet, "/zones/"+zona+"/dns_records?"+url.Values{"name": {host}, "type": {"A"}}.Encode(), nil, &rs); err != nil {
		return err
	}
	for _, r := range rs {
		if r.Valor == ip {
			if err := c.pedir(ctx, http.MethodDelete, "/zones/"+zona+"/dns_records/"+r.ID, nil, nil); err != nil {
				return err
			}
		}
	}
	return nil
}

// limpiarViejo borra la route del Worker y el AAAA 100:: de cuando la web corría en Cloudflare.
func (c *Cliente) limpiarViejo(ctx context.Context, zona, host string) error {
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

// SSLCompleto deja la zona en SSL "full": Cloudflare cifra hasta el VPS y acepta el certificado que
// Caddy emite solo (en "strict" exigiría uno de una autoridad pública o de origen de Cloudflare).
func (c *Cliente) SSLCompleto(ctx context.Context, zona string) error {
	return c.pedir(ctx, http.MethodPatch, "/zones/"+zona+"/settings/ssl", map[string]string{"value": "full"}, nil)
}
