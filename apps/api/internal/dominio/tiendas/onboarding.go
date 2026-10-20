package tiendas

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Plantillas son las 16 del handoff (apps/web/src/lib/plantillas.ts).
var Plantillas = []string{"ceibo", "mostaza", "yerbal", "pampa", "rosedal", "quebracho", "andes", "malbec",
	"patio", "miga", "taller", "hora", "litoral", "torno", "tijera", "lente"}

// ErrPlantilla: la plantilla no existe.
var ErrPlantilla = errors.New("plantilla desconocida")

// Onboarding es lo que manda el wizard: la plantilla elegida, las demás elecciones tal cual y si
// terminó. Las elecciones se guardan como vienen (jsonb, hasta 16 KB): las lee cada área del panel.
type Onboarding struct {
	Plantilla  string          `json:"plantilla"`
	Elecciones json.RawMessage `json:"elecciones"`
	Terminado  bool            `json:"terminado"`
}

// GuardarOnboarding guarda el avance del onboarding de la tienda.
func GuardarOnboarding(ctx context.Context, pool interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}, tiendaID string, o Onboarding) error {
	if o.Plantilla != "" && !slices.Contains(Plantillas, o.Plantilla) {
		return ErrPlantilla
	}
	if len(o.Elecciones) == 0 {
		o.Elecciones = json.RawMessage("{}")
	}
	if !json.Valid(o.Elecciones) || len(o.Elecciones) > 16<<10 {
		return errors.New("elecciones inválidas")
	}
	return pool.EnTienda(ctx, tiendaID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `
			UPDATE tiendas SET
				onboarding = $2::jsonb,
				plantilla = coalesce(nullif($3, ''), plantilla),
				onboarding_terminado_en = CASE WHEN $4 THEN coalesce(onboarding_terminado_en, now()) ELSE onboarding_terminado_en END
			WHERE id = $1`, tiendaID, string(o.Elecciones), o.Plantilla, o.Terminado)
		return err
	})
}

func rutasOnboarding(mux *http.ServeMux, pool interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}, tienda func(*http.Request) string) {
	mux.HandleFunc("PUT /v1/panel/onboarding", func(w http.ResponseWriter, r *http.Request) {
		id := tienda(r)
		if id == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		var o Onboarding
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 20<<10)).Decode(&o); err != nil {
			problema.Escribir(w, problema.Invalido("El pedido no es un JSON válido."))
			return
		}
		switch err := GuardarOnboarding(r.Context(), pool, id, o); {
		case err == nil:
			w.WriteHeader(http.StatusNoContent)
		case errors.Is(err, ErrPlantilla):
			problema.Escribir(w, problema.Invalido("Elegí una de las plantillas.", problema.Campo{Ruta: "plantilla", Mensaje: "Elegí una de las plantillas."}))
		default:
			problema.Escribir(w, problema.Invalido("No pudimos guardar el onboarding."))
		}
	})
}
