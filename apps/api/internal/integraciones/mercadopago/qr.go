package mercadopago

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
)

// Cobro con QR en el local: Orders API con type "qr" y una caja (POS) de la cuenta de la tienda
// (LOCK 8 del plugin). La sucursal se crea con la dirección real que carga la tienda; nunca se
// inventa una.

// Ubicacion de la sucursal, como la pide /users/{id}/stores.
type Ubicacion struct {
	Calle     string  `json:"calle"`
	Numero    string  `json:"numero"`
	Ciudad    string  `json:"ciudad"`
	Provincia string  `json:"provincia"`
	Latitud   float64 `json:"latitud"`
	Longitud  float64 `json:"longitud"`
}

func (u Ubicacion) valida() bool {
	return strings.TrimSpace(u.Calle) != "" && strings.TrimSpace(u.Numero) != "" && strings.TrimSpace(u.Ciudad) != "" &&
		strings.TrimSpace(u.Provincia) != "" && u.Latitud >= -56 && u.Latitud <= -21 && u.Longitud >= -74 && u.Longitud <= -53
}

// CrearCaja da de alta la sucursal y la caja en la cuenta de la tienda (una vez).
func (s *Servicio) CrearCaja(ctx context.Context, tiendaID string, u Ubicacion) error {
	if !u.valida() {
		return ErrInvalido
	}
	token, usuario, err := s.tokenDe(ctx, tiendaID)
	if err != nil {
		return err
	}
	var nombre string
	if err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT nombre FROM tiendas WHERE id = $1`, tiendaID).Scan(&nombre)
	}); err != nil {
		return err
	}
	externo := "MOS" + strings.ReplaceAll(tiendaID, "-", "")[:12]
	var suc struct {
		ID json.Number `json:"id"`
	}
	if err := s.cliente().JSON(ctx, http.MethodPost, fmt.Sprintf("/users/%d/stores", usuario), token, "", map[string]any{
		"name":        recortar(nombre, 60),
		"external_id": externo,
		"location": map[string]any{
			"street_name": u.Calle, "street_number": u.Numero, "city_name": u.Ciudad, "state_name": u.Provincia,
			"latitude": u.Latitud, "longitude": u.Longitud,
		},
	}, &suc); err != nil {
		return err
	}
	// La caja en modo atendido (pdv): las órdenes las crea Tu Local Web.
	if err := s.cliente().JSON(ctx, http.MethodPost, "/v2/pos", token, "caja:"+externo, map[string]any{
		"name": "Caja 1", "store_id": suc.ID.String(), "external_id": externo + "C1",
		"config": map[string]any{"qr": map[string]any{"operating_mode": "pdv"}},
	}, nil); err != nil {
		return err
	}
	return s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE mp_conexiones SET sucursal_mp = $2, caja_externa = $3 WHERE tienda_id = $1`,
			tiendaID, suc.ID.String(), externo+"C1")
		return err
	})
}

// QR es la orden recién creada: qr_data se dibuja en el panel.
type QR struct {
	Referencia string `json:"referencia"`
	QRData     string `json:"qr"`
	Estado     string `json:"estado"`
}

type orden struct {
	ID                string `json:"id"`
	Status            string `json:"status"`
	ExternalReference string `json:"external_reference"`
	TotalAmount       string `json:"total_amount"`
	TypeResponse      struct {
		QRData string `json:"qr_data"`
	} `json:"type_response"`
}

// CobrarQR crea una orden dinámica por el total del pedido (que sale de la base).
func (s *Servicio) CobrarQR(ctx context.Context, tiendaID string, pedido int) (QR, error) {
	var total int64
	var caja *string
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if err := t.QueryRow(ctx, `SELECT total FROM pedidos WHERE numero = $1`, pedido).Scan(&total); err != nil {
			return err
		}
		return t.QueryRow(ctx, `SELECT caja_externa FROM mp_conexiones WHERE tienda_id = $1`, tiendaID).Scan(&caja)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return QR{}, ErrNoExiste
	}
	if err != nil {
		return QR{}, err
	}
	if caja == nil {
		return QR{}, ErrSinCaja
	}
	if total <= 0 {
		return QR{}, ErrInvalido
	}
	token, _, err := s.tokenDe(ctx, tiendaID)
	if err != nil {
		return QR{}, err
	}
	ref := nuevaReferencia(tiendaID)
	if err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `INSERT INTO mp_cobros (referencia, tienda_id, tipo, objetivo, monto) VALUES ($1, $2, 'qr', $3, $4)`,
			ref, tiendaID, fmt.Sprint(pedido), total)
		return err
	}); err != nil {
		return QR{}, err
	}
	var o orden
	if err := s.cliente().JSON(ctx, http.MethodPost, "/v1/orders", token, ref, map[string]any{
		"type":               "qr",
		"external_reference": ref,
		"total_amount":       monto(total),
		"description":        fmt.Sprintf("Pedido #%d", pedido),
		"config":             map[string]any{"qr": map[string]any{"external_pos_id": *caja, "mode": "dynamic"}},
		"transactions":       map[string]any{"payments": []map[string]any{{"amount": monto(total)}}},
	}, &o); err != nil {
		return QR{}, err
	}
	err = s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE mp_cobros SET orden = $2, actualizado_en = now() WHERE referencia = $1`, ref, o.ID)
		return err
	})
	return QR{Referencia: ref, QRData: o.TypeResponse.QRData, Estado: o.Status}, err
}

// ordenDe: la orden de un cobro con QR de la tienda.
func (s *Servicio) ordenDe(ctx context.Context, tiendaID, ref string) (string, error) {
	var id *string
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT orden FROM mp_cobros WHERE referencia = $1 AND tipo = 'qr'`, ref).Scan(&id)
	})
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && id == nil) {
		return "", ErrNoExiste
	}
	return *id, err
}

// EstadoQR consulta la orden y la aplica (sirve aunque el webhook no haya llegado).
func (s *Servicio) EstadoQR(ctx context.Context, tiendaID, ref string) (string, error) {
	id, err := s.ordenDe(ctx, tiendaID, ref)
	if err != nil {
		return "", err
	}
	return s.conciliarOrden(ctx, tiendaID, id)
}

func (s *Servicio) conciliarOrden(ctx context.Context, tiendaID, id string) (string, error) {
	token, _, err := s.tokenDe(ctx, tiendaID)
	if err != nil {
		return "", err
	}
	var o orden
	if err := s.cliente().JSON(ctx, http.MethodGet, "/v1/orders/"+id, token, "", nil, &o); err != nil {
		return "", err
	}
	var pagado int64
	_, _ = fmt.Sscanf(o.TotalAmount, "%d", &pagado)
	return o.Status, s.aplicar(ctx, tiendaID, o.ExternalReference, 0, o.Status, pagado)
}

// CancelarQR cancela una orden que todavía no se pagó.
func (s *Servicio) CancelarQR(ctx context.Context, tiendaID, ref string) error {
	id, err := s.ordenDe(ctx, tiendaID, ref)
	if err != nil {
		return err
	}
	token, _, err := s.tokenDe(ctx, tiendaID)
	if err != nil {
		return err
	}
	if err := s.cliente().JSON(ctx, http.MethodPost, "/v1/orders/"+id+"/cancel", token, "cancelar:"+ref, nil, nil); err != nil {
		return err
	}
	_, err = s.conciliarOrden(ctx, tiendaID, id)
	return err
}
