package tiendas

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// DiasPrueba: lo que dura la prueba gratis.
const DiasPrueba = 14

// Resumen es lo que muestra el armazón del panel (components/schemas/ResumenPanel). La insignia de
// pedidos ya es real; las demás y las notificaciones se llenan a medida que llegan los dominios
// (carritos, WhatsApp, tareas); hasta entonces van en cero y vacías, nunca inventadas.
type Resumen struct {
	Insignias struct {
		Pedidos   int `json:"pedidos"`
		Carritos  int `json:"carritos"`
		WhatsApp  int `json:"whatsapp"`
		ParaHacer int `json:"paraHacer"`
	} `json:"insignias"`
	Notificaciones []any    `json:"notificaciones"`
	Canales        []string `json:"canales"`
	Prueba         struct {
		Plan          string `json:"plan"`
		DiasRestantes int    `json:"diasRestantes"`
		DiasTotales   int    `json:"diasTotales"`
	} `json:"prueba"`
}

// ObtenerResumen arma el resumen de la tienda.
func ObtenerResumen(ctx context.Context, pool interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}, tiendaID string, ahora time.Time) (Resumen, error) {
	var r Resumen
	r.Notificaciones, r.Canales = []any{}, []string{}
	var creada time.Time
	var plan string
	err := pool.EnTienda(ctx, tiendaID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `SELECT plan, creada_en FROM tiendas WHERE id = $1`, tiendaID).Scan(&plan, &creada); err != nil {
			return err
		}
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM pedidos WHERE tienda_id = $1 AND estado = 'Para preparar'`, tiendaID).
			Scan(&r.Insignias.Pedidos); err != nil {
			return err
		}
		// Carritos abandonados abiertos de la última semana, sin recordar todavía.
		return tx.QueryRow(ctx, `SELECT count(*) FROM carritos WHERE tienda_id = $1 AND recuperado_en IS NULL
			AND recordado_en IS NULL AND actualizado > now() - interval '7 days'`, tiendaID).Scan(&r.Insignias.Carritos)
	})
	if err != nil {
		return r, err
	}
	r.Prueba.Plan, r.Prueba.DiasTotales = "Pro", DiasPrueba
	r.Prueba.DiasRestantes = max(0, DiasPrueba-int(math.Floor(ahora.Sub(creada).Hours()/24)))
	if plan != "gratis" {
		r.Prueba.Plan, r.Prueba.DiasRestantes = plan, 0
	}
	return r, nil
}

func rutasResumen(mux *http.ServeMux, pool interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}, tienda func(*http.Request) string) {
	mux.HandleFunc("GET /v1/panel/resumen", func(w http.ResponseWriter, r *http.Request) {
		id := tienda(r)
		if id == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		res, err := ObtenerResumen(r.Context(), pool, id, time.Now())
		if errors.Is(err, pgx.ErrNoRows) {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(res)
	})
	// Sin notificaciones todavía: marcarlas como leídas no tiene nada que hacer.
	mux.HandleFunc("POST /v1/panel/notificaciones/leidas", func(w http.ResponseWriter, r *http.Request) {
		if tienda(r) == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
