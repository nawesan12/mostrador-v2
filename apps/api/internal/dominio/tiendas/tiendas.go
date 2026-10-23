package tiendas

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Tienda tal como la define el contrato (components/schemas/Tienda).
type Tienda struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Nombre    string    `json:"nombre"`
	Rubro     string    `json:"rubro"`
	Plantilla string    `json:"plantilla"`
	Plan      string    `json:"plan"`
	CreadaEn  time.Time `json:"creadaEn"`
}

// Crear da de alta una tienda del rubro con la plantilla y el nombre de ejemplo como punto de partida.
func Crear(ctx context.Context, pool *db.Pool, slug, nombre, rubro string) (string, error) {
	r, ok := BuscarRubro(rubro)
	if !ok {
		return "", errors.New("rubro desconocido")
	}
	var id string
	err := pool.EnTx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, "SELECT app.crear_tienda($1, $2, $3, $4)", slug, nombre, r.ID, r.Plantilla).Scan(&id)
	})
	return id, err
}

// Obtener lee la tienda (con la tienda ya fijada por EnTienda).
func Obtener(ctx context.Context, pool *db.Pool, tiendaID string) (Tienda, error) {
	var t Tienda
	err := pool.EnTienda(ctx, tiendaID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT id, slug, nombre, rubro, plantilla, plan, creada_en FROM tiendas WHERE id = $1`, tiendaID).
			Scan(&t.ID, &t.Slug, &t.Nombre, &t.Rubro, &t.Plantilla, &t.Plan, &t.CreadaEn)
	})
	return t, err
}

// Rutas registra los endpoints del dominio. tienda resuelve la tienda del pedido (sale de la sesión).
func Rutas(mux *http.ServeMux, pool *db.Pool, tienda func(*http.Request) string) {
	mux.HandleFunc("GET /v1/rubros", func(w http.ResponseWriter, _ *http.Request) {
		escribir(w, Rubros)
	})
	mux.HandleFunc("PUT /v1/panel/tienda/origen", func(w http.ResponseWriter, r *http.Request) {
		id := tienda(r)
		if id == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		var o Origen
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&o) != nil {
			problema.Escribir(w, problema.Invalido("Origen inválido."))
			return
		}
		fijado, err := FijarOrigen(r.Context(), pool, id, o)
		if errors.Is(err, ErrOrigen) {
			problema.Escribir(w, problema.Invalido("Origen inválido."))
			return
		}
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		escribir(w, map[string]bool{"fijado": fijado})
	})
	mux.HandleFunc("GET /v1/panel/tienda", func(w http.ResponseWriter, r *http.Request) {
		id := tienda(r)
		if id == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		t, err := Obtener(r.Context(), pool, id)
		if errors.Is(err, pgx.ErrNoRows) {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		escribir(w, t)
	})
	rutasOnboarding(mux, pool, tienda)
	rutasResumen(mux, pool, tienda)
	rutasTablero(mux, pool, tienda)
}

func escribir(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}
