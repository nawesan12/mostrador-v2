package cobros

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
)

// Cambio de adónde llega la plata: los datos de transferencia (alias, CBU, titular) o la cuenta de
// Mercado Pago. Cada uno le avisa a la dueña por mail y push (también si lo hizo ella), con un link
// para volverlo atrás y cerrar todas las sesiones del equipo (auth.cobros_cambios, 00092).
const (
	CambioTransferencia = "transferencia"
	CambioMercadoPago   = "mercadopago"
	// TipoCambio es el evento del mail; TipoCambioPush, el del push.
	TipoCambio     = "cobros.cambio"
	TipoCambioPush = "push.cobros_cambio"
)

// AvisoCambio son los datos de los eventos: qué cambió (sólo lo que se puede mostrar: el CBU va
// con sus últimos 4 números) y el token del link para volverlo atrás.
type AvisoCambio struct {
	Tipo    string            `json:"tipo"`
	Antes   map[string]string `json:"antes"`
	Despues map[string]string `json:"despues"`
	Quien   string            `json:"quien"`
	Token   string            `json:"token,omitempty"`
}

func hashToken(t string) string {
	h := sha256.Sum256([]byte(t))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// AvisarCambio guarda el cambio (con lo de antes entero, para volverlo atrás) y encola los avisos,
// en la transacción del cambio. mostrarAntes/mostrarDespues son lo que dice el mail.
func AvisarCambio(ctx context.Context, t pgx.Tx, tiendaID, usuarioID, tipo string, antes, despues, mostrarAntes, mostrarDespues map[string]string) error {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	token := base64.RawURLEncoding.EncodeToString(b)
	a, _ := json.Marshal(antes)
	d, _ := json.Marshal(despues)
	if _, err := t.Exec(ctx, `INSERT INTO auth.cobros_cambios (tienda_id, tipo, antes, despues, usuario_id, token_hash)
		VALUES ($1, $2, $3, $4, nullif($5, '')::uuid, $6)`, tiendaID, tipo, a, d, usuarioID, hashToken(token)); err != nil {
		return err
	}
	var quien string
	if usuarioID != "" {
		_ = t.QueryRow(ctx, `SELECT nombre FROM auth.usuarios WHERE id = $1`, usuarioID).Scan(&quien)
	}
	ev := AvisoCambio{Tipo: tipo, Antes: mostrarAntes, Despues: mostrarDespues, Quien: quien, Token: token}
	datos, _ := json.Marshal(ev)
	if _, err := t.Exec(ctx, `INSERT INTO outbox (tienda_id, tipo, datos) VALUES ($1, $2, $3)`, tiendaID, TipoCambio, datos); err != nil {
		return err
	}
	ev.Token = "" // el push no lleva el link: se toca y abre el panel
	datos, _ = json.Marshal(ev)
	_, err := t.Exec(ctx, `INSERT INTO outbox (tienda_id, tipo, datos) VALUES ($1, $2, $3)`, tiendaID, TipoCambioPush, datos)
	return err
}

// ErrCambioVencido: el link para volver atrás no existe, ya se usó o pasaron 7 días.
var ErrCambioVencido = errors.New("cobros: el link para volver atrás venció")

// CambioARevertir lee el cambio del link (sin tienda fijada: auth no tiene RLS) y lo marca usado.
func CambioARevertir(ctx context.Context, t pgx.Tx, token string) (tiendaID, tipo string, antes map[string]string, err error) {
	if token == "" || len(token) > 200 {
		return "", "", nil, ErrCambioVencido
	}
	var a []byte
	err = t.QueryRow(ctx, `UPDATE auth.cobros_cambios SET revertido_en = now()
		WHERE token_hash = $1 AND revertido_en IS NULL AND creado_en > now() - interval '7 days'
		RETURNING tienda_id::text, tipo, antes`, hashToken(token)).Scan(&tiendaID, &tipo, &a)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", nil, ErrCambioVencido
	}
	if err != nil {
		return "", "", nil, err
	}
	err = json.Unmarshal(a, &antes)
	return tiendaID, tipo, antes, err
}
