// Package mercadopago: la integración con Mercado Pago, según las guías oficiales del plugin de
// Mercado Pago (marketplace con OAuth + Checkout Pro, QR con la Orders API, suscripciones sin plan
// en estado pending y webhooks firmados).
//
//   - Cada tienda conecta su cuenta por OAuth y cobra con su propio token (cifrado en la base). Nunca
//     se cae al token de la plataforma: sin conexión, no hay cobro.
//   - Mostrador no cobra comisión: marketplace_fee no se manda.
//   - El abono de Mostrador es una suscripción (preapproval) con el token de la plataforma.
package mercadopago

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ErrorMP es una respuesta no exitosa de la API. Nunca se le muestra el cuerpo al comprador.
type ErrorMP struct {
	Estado int
	Cuerpo string
}

func (e *ErrorMP) Error() string { return fmt.Sprintf("mercadopago: HTTP %d", e.Estado) }

// Cliente habla con api.mercadopago.com (Base se cambia en los tests).
type Cliente struct {
	Base string
	HTTP *http.Client
}

func (c *Cliente) base() string {
	if c.Base != "" {
		return c.Base
	}
	return "https://api.mercadopago.com"
}

func (c *Cliente) http() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 15 * time.Second}
}

// JSON hace un pedido con el token dado. idem, si no está vacío, va como X-Idempotency-Key.
func (c *Cliente) JSON(ctx context.Context, metodo, ruta, token, idem string, cuerpo, salida any) error {
	var r io.Reader
	if cuerpo != nil {
		b, err := json.Marshal(cuerpo)
		if err != nil {
			return err
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, metodo, c.base()+ruta, r)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if cuerpo != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if idem != "" {
		req.Header.Set("X-Idempotency-Key", idem)
	}
	return c.hacer(req, salida)
}

// Formulario: /oauth/token va con application/x-www-form-urlencoded.
func (c *Cliente) Formulario(ctx context.Context, ruta string, v url.Values, salida any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base()+ruta, strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	return c.hacer(req, salida)
}

func (c *Cliente) hacer(req *http.Request, salida any) error {
	res, err := c.http().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode >= 300 {
		return &ErrorMP{Estado: res.StatusCode, Cuerpo: string(b)}
	}
	if salida == nil || len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, salida)
}

// monto: los montos de Mostrador son pesos enteros; la Orders API los pide como texto con dos decimales.
func monto(pesos int64) string { return fmt.Sprintf("%d.00", pesos) }
