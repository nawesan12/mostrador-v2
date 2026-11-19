package tiendas

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/abono"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/planes"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// DiasPrueba: lo que dura la prueba gratis.
const DiasPrueba = 14

// Resumen es lo que muestra el armazón del panel (components/schemas/ResumenPanel). La insignia de
// pedidos y carritos ya son reales; las demás se llenan a medida que llegan los dominios (WhatsApp,
// tareas); hasta entonces van en cero, nunca inventadas. Las notificaciones van vacías: la campanita
// lee las suyas de /v1/panel/notificaciones (paquete notificaciones, §45).
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
	var e abono.Estado
	err := pool.EnTienda(ctx, tiendaID, func(tx pgx.Tx) error {
		var err error
		if e, err = abono.Leer(ctx, tx, tiendaID); err != nil {
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
	// El nombre que se lee (Crece, no el id «pro»), y los días que quedan hasta el fin de la prueba
	// (con la extensión del equipo, si la hubo).
	r.Prueba.Plan, r.Prueba.DiasTotales = planes.De(planes.DePrueba).Nombre, DiasPrueba
	r.Prueba.DiasRestantes = max(0, int(math.Ceil(e.FinPrueba.Sub(ahora).Hours()/24)))
	if e.Plan != "gratis" {
		r.Prueba.Plan, r.Prueba.DiasRestantes = planes.De(e.Plan).Nombre, 0
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
}
