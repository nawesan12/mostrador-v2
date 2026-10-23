package mercadopago

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// nuevaReferencia: el external_reference del cobro, sólo letras y números (la Orders API no acepta
// otra cosa). Lleva el comienzo de la tienda para leerla de un vistazo en el panel de Mercado Pago,
// pero se concilia contra mp_cobros, nunca contra lo que diga el texto.
func nuevaReferencia(tiendaID string) string {
	b := make([]byte, 10)
	_, _ = rand.Read(b)
	return "MOS" + strings.ToUpper(tiendaID[:8]+hex.EncodeToString(b))
}

type item struct {
	Title      string `json:"title"`
	Quantity   int    `json:"quantity"`
	UnitPrice  int64  `json:"unit_price"`
	CurrencyID string `json:"currency_id"`
}

type preferencia struct {
	Items             []item            `json:"items"`
	ExternalReference string            `json:"external_reference"`
	BackURLs          map[string]string `json:"back_urls"`
	AutoReturn        string            `json:"auto_return,omitempty"`
	PaymentMethods    map[string]any    `json:"payment_methods,omitempty"`
	StatementDesc     string            `json:"statement_descriptor,omitempty"`
}

// PagarLink crea (o reusa) la preferencia de Checkout Pro de un link de pago, con el token de la
// tienda, y devuelve el init_point. El monto y el concepto salen de la base, nunca del navegador.
func (s *Servicio) PagarLink(ctx context.Context, tiendaID, codigo string) (string, error) {
	var concepto, tienda string
	var montoLink int64
	var cuotas, pagado bool
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT l.concepto, l.monto, l.cuotas, l.pagado_en IS NOT NULL, t.nombre
			FROM links l JOIN tiendas t ON t.id = l.tienda_id WHERE l.codigo = $1`, codigo).
			Scan(&concepto, &montoLink, &cuotas, &pagado, &tienda)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNoExiste
	}
	if err != nil {
		return "", err
	}
	if pagado {
		return "", ErrYaPagado
	}
	token, _, err := s.tokenDe(ctx, tiendaID)
	if err != nil {
		return "", err
	}
	// Si ya hay una preferencia pendiente por el mismo monto, se reusa: un doble clic no crea dos.
	var ref, url string
	_ = s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT referencia, coalesce(preferencia, '') FROM mp_cobros
			WHERE tienda_id = $1 AND tipo = 'link' AND objetivo = $2 AND monto = $3 AND estado = 'pendiente'
				AND preferencia IS NOT NULL AND creado_en > now() - interval '6 hours'
			ORDER BY creado_en DESC LIMIT 1`, tiendaID, codigo, montoLink).Scan(&ref, &url)
	})
	if url != "" {
		return url, nil
	}
	ref = nuevaReferencia(tiendaID)
	if err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `INSERT INTO mp_cobros (referencia, tienda_id, tipo, objetivo, monto) VALUES ($1, $2, 'link', $3, $4)`,
			ref, tiendaID, codigo, montoLink)
		return err
	}); err != nil {
		return "", err
	}
	vuelta := s.Config.Sitio + "/p/" + codigo
	p := preferencia{
		Items:             []item{{Title: concepto, Quantity: 1, UnitPrice: montoLink, CurrencyID: "ARS"}},
		ExternalReference: ref,
		BackURLs:          map[string]string{"success": vuelta + "?pago=ok", "pending": vuelta + "?pago=pendiente", "failure": vuelta + "?pago=error"},
		StatementDesc:     recortar(tienda, 22),
	}
	if s.publico() {
		p.AutoReturn = "approved"
	}
	if !cuotas {
		p.PaymentMethods = map[string]any{"installments": 1}
	}
	var r struct {
		ID        string `json:"id"`
		InitPoint string `json:"init_point"`
	}
	// /checkout/preferences, sin /v1 (la guía de Checkout Pro). Siempre init_point.
	if err := s.cliente().JSON(ctx, http.MethodPost, "/checkout/preferences", token, ref, p, &r); err != nil {
		return "", err
	}
	if r.InitPoint == "" {
		return "", errors.New("mercadopago: la preferencia no trajo init_point")
	}
	err = s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		// Se guarda el init_point (no es secreto) para reusarlo ante un doble clic.
		_, err := t.Exec(ctx, `UPDATE mp_cobros SET preferencia = $2, actualizado_en = now() WHERE referencia = $1`, ref, r.InitPoint)
		return err
	})
	return r.InitPoint, err
}

func recortar(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n])
	}
	return s
}

// Pago es lo que interesa de GET /v1/payments/{id}.
type Pago struct {
	ID                int64   `json:"id"`
	Status            string  `json:"status"`
	StatusDetail      string  `json:"status_detail"`
	ExternalReference string  `json:"external_reference"`
	TransactionAmount float64 `json:"transaction_amount"`
	PaymentTypeID     string  `json:"payment_type_id"`
}

// ConciliarPago trae el pago de Mercado Pago (con el token de la tienda: el aviso no es la fuente de
// verdad) y lo aplica: el cobro, el link y el pedido. Es idempotente.
func (s *Servicio) ConciliarPago(ctx context.Context, tiendaID string, pagoID int64) error {
	token, _, err := s.tokenDe(ctx, tiendaID)
	if err != nil {
		return err
	}
	var p Pago
	if err := s.cliente().JSON(ctx, http.MethodGet, "/v1/payments/"+strconv.FormatInt(pagoID, 10), token, "", nil, &p); err != nil {
		return err
	}
	return s.aplicar(ctx, tiendaID, p.ExternalReference, p.ID, p.Status, int64(p.TransactionAmount+0.5))
}

// aplicar el estado de un pago (o de una orden de QR) a lo que se cobró.
func (s *Servicio) aplicar(ctx context.Context, tiendaID, ref string, pagoID int64, estado string, pagado int64) error {
	return s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var tipo, objetivo, previo string
		var esperado int64
		err := t.QueryRow(ctx, `SELECT tipo, objetivo, monto, estado FROM mp_cobros WHERE referencia = $1 FOR UPDATE`, ref).
			Scan(&tipo, &objetivo, &esperado, &previo)
		if errors.Is(err, pgx.ErrNoRows) {
			// Un pago de la cuenta que no salió de Mostrador: no es nuestro.
			return nil
		}
		if err != nil {
			return err
		}
		if _, err := t.Exec(ctx, `UPDATE mp_cobros SET pago = $2, estado = $3, actualizado_en = now() WHERE referencia = $1`,
			ref, pagoID, estado); err != nil {
			return err
		}
		if estado == previo {
			return nil
		}
		aprobado := (estado == "approved" || estado == "processed") && pagado >= esperado
		devuelto := estado == "refunded" || estado == "charged_back"
		pedido := ""
		switch tipo {
		case "link":
			if aprobado {
				if _, err := t.Exec(ctx, `UPDATE links SET pagado_en = coalesce(pagado_en, now()) WHERE codigo = $1`, objetivo); err != nil {
					return err
				}
				var n *int
				_ = t.QueryRow(ctx, `SELECT pedido FROM links WHERE codigo = $1`, objetivo).Scan(&n)
				if n != nil {
					pedido = strconv.Itoa(*n)
				}
			}
		case "pedido", "qr":
			pedido = objetivo
		}
		if pedido != "" && (aprobado || devuelto) {
			if aprobado {
				if _, err := t.Exec(ctx, `UPDATE pedidos SET estado_pago = 'Aprobado',
						estado = CASE WHEN estado = 'Sin pagar' THEN 'Para preparar' ELSE estado END,
						estados_en = CASE WHEN estado = 'Sin pagar' THEN estados_en || jsonb_build_object('Para preparar', now()) ELSE estados_en END
					WHERE numero = $1::int`, pedido); err != nil {
					return err
				}
			} else if _, err := t.Exec(ctx, `UPDATE pedidos SET estado_pago = 'Devuelto' WHERE numero = $1::int`, pedido); err != nil {
				return err
			}
		}
		if aprobado {
			return outbox.Publicar(ctx, t, tiendaID, "pago.aprobado", map[string]any{"referencia": ref, "tipo": tipo, "objetivo": objetivo, "monto": pagado})
		}
		return nil
	})
}

// Reembolsar devuelve el pago de un pedido (total si monto es 0). Idempotente por clave. Los cobros
// con link o checkout se devuelven por Payments; los de QR, por la Orders API.
func (s *Servicio) Reembolsar(ctx context.Context, tiendaID string, pedido int, montoParcial int64, clave string) error {
	var tipo string
	var pagoID *int64
	var ordenID *string
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT c.tipo, c.pago, c.orden FROM mp_cobros c
			LEFT JOIN links l ON c.tipo = 'link' AND l.codigo = c.objetivo
			WHERE c.estado IN ('approved', 'processed')
				AND ((c.tipo IN ('pedido', 'qr') AND c.objetivo = $1::text) OR l.pedido = $1)
			ORDER BY c.actualizado_en DESC LIMIT 1`, pedido).Scan(&tipo, &pagoID, &ordenID)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoExiste
	}
	if err != nil {
		return err
	}
	token, _, err := s.tokenDe(ctx, tiendaID)
	if err != nil {
		return err
	}
	if tipo == "qr" && ordenID != nil {
		cuerpo := map[string]any{}
		if montoParcial > 0 {
			cuerpo["transactions"] = []map[string]any{{"amount": monto(montoParcial)}}
		}
		if err := s.cliente().JSON(ctx, http.MethodPost, "/v1/orders/"+*ordenID+"/refund", token, "reembolso:"+clave, cuerpo, nil); err != nil {
			return err
		}
		_, err = s.conciliarOrden(ctx, tiendaID, *ordenID)
		return err
	}
	if pagoID == nil {
		return ErrNoExiste
	}
	cuerpo := map[string]any{}
	if montoParcial > 0 {
		cuerpo["amount"] = montoParcial
	}
	if err := s.cliente().JSON(ctx, http.MethodPost, fmt.Sprintf("/v1/payments/%d/refunds", *pagoID), token,
		"reembolso:"+clave, cuerpo, nil); err != nil {
		return err
	}
	return s.ConciliarPago(ctx, tiendaID, *pagoID)
}
