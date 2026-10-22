// Package precios: listas de precio y el aumento por inflación (design/PanelCatalogo.dc.html).
// Las listas son reglas sobre el precio base; el aumento toca el precio base de todos los productos
// en una transacción y guarda los de antes para deshacer.
package precios

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}

// ErrNoExiste: no hay aumento para deshacer o cancelar.
var ErrNoExiste = errors.New("no hay aumento")

type Lista struct {
	ID      string `json:"id"`
	Nombre  string `json:"nombre"`
	Quien   string `json:"quien"`
	Ajuste  int    `json:"ajuste"`
	Detalle string `json:"detalle"`
}

type Muestra struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
	Precio int64  `json:"precio"`
}

type Programado struct {
	Pct      int       `json:"pct"`
	Redondeo string    `json:"redondeo"`
	Para     time.Time `json:"para"`
}

// Listas es components/schemas/ListasPrecio.
type Listas struct {
	Listas     []Lista     `json:"listas"`
	Productos  int         `json:"productos"`
	Muestra    []Muestra   `json:"muestra"`
	Programado *Programado `json:"programado"`
}

// redondeoSQL: el mismo redondeo que el panel (lib/panel/precios.ts) sobre $1 (pct) y $2 (redondeo).
const nuevoPrecio = `CASE $3::text
	WHEN '90' THEN GREATEST(90, round(precio * (1 + $2::numeric / 100) / 100) * 100 - 10)
	WHEN '500' THEN GREATEST(500, round(precio * (1 + $2::numeric / 100) / 500) * 500)
	ELSE round(precio * (1 + $2::numeric / 100)) END`

// ProximoLunes: el lunes que viene a las 00:00 de Argentina (UTC−3).
func ProximoLunes(ahora time.Time) time.Time {
	ar := ahora.UTC().Add(-3 * time.Hour)
	dias := (8 - int(ar.Weekday())) % 7
	if dias == 0 {
		dias = 7
	}
	return time.Date(ar.Year(), ar.Month(), ar.Day()+dias, 3, 0, 0, 0, time.UTC)
}

func aplicar(ctx context.Context, t pgx.Tx, tiendaID string, pct int, redondeo string) (int, error) {
	var antes []byte
	var n int
	if err := t.QueryRow(ctx, `SELECT coalesce(jsonb_object_agg(id::text, precio), '{}'), count(*)
		FROM (SELECT id, precio FROM productos WHERE tienda_id = $1 FOR UPDATE) p`, tiendaID).Scan(&antes, &n); err != nil {
		return 0, err
	}
	if _, err := t.Exec(ctx, `UPDATE productos SET precio = `+nuevoPrecio+` WHERE tienda_id = $1`, tiendaID, pct, redondeo); err != nil {
		return 0, err
	}
	_, err := t.Exec(ctx, `INSERT INTO aumentos (tienda_id, pct, redondeo, aplicado_en, antes) VALUES ($1, $2, $3, now(), $4)`,
		tiendaID, pct, redondeo, antes)
	return n, err
}

// Aplicar sube los precios ya; si programar, lo deja para el lunes (lo aplica el primero que mire
// las listas después de esa hora, o el worker).
func Aplicar(ctx context.Context, pool tx, tiendaID string, pct int, redondeo string, programar bool, ahora time.Time) (int, *time.Time, error) {
	if pct < 1 || pct > 50 || (redondeo != "no" && redondeo != "90" && redondeo != "500") {
		return 0, nil, errors.New("aumento inválido")
	}
	var n int
	var para *time.Time
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if !programar {
			var err error
			n, err = aplicar(ctx, t, tiendaID, pct, redondeo)
			return err
		}
		p := ProximoLunes(ahora)
		para = &p
		if _, err := t.Exec(ctx, `DELETE FROM aumentos WHERE tienda_id = $1 AND aplicado_en IS NULL`, tiendaID); err != nil {
			return err
		}
		if _, err := t.Exec(ctx, `INSERT INTO aumentos (tienda_id, pct, redondeo, para) VALUES ($1, $2, $3, $4)`, tiendaID, pct, redondeo, p); err != nil {
			return err
		}
		return t.QueryRow(ctx, `SELECT count(*) FROM productos WHERE tienda_id = $1`, tiendaID).Scan(&n)
	})
	return n, para, err
}

// Deshacer vuelve los precios del último aumento aplicado.
func Deshacer(ctx context.Context, pool tx, tiendaID string) error {
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var id int64
		var antes []byte
		err := t.QueryRow(ctx, `SELECT id, antes FROM aumentos WHERE tienda_id = $1 AND aplicado_en IS NOT NULL
			AND deshecho_en IS NULL ORDER BY id DESC LIMIT 1 FOR UPDATE`, tiendaID).Scan(&id, &antes)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExiste
		}
		if err != nil {
			return err
		}
		if _, err := t.Exec(ctx, `UPDATE productos p SET precio = (a.value)::bigint
			FROM jsonb_each_text($2::jsonb) a WHERE p.tienda_id = $1 AND p.id::text = a.key`, tiendaID, antes); err != nil {
			return err
		}
		_, err = t.Exec(ctx, `UPDATE aumentos SET deshecho_en = now() WHERE id = $1`, id)
		return err
	})
}

// Cancelar borra el aumento programado.
func Cancelar(ctx context.Context, pool tx, tiendaID string) error {
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `DELETE FROM aumentos WHERE tienda_id = $1 AND aplicado_en IS NULL`, tiendaID)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrNoExiste
		}
		return err
	})
}

// Leer devuelve las listas; si el programado ya venció, lo aplica antes.
func Leer(ctx context.Context, pool tx, tiendaID string, ahora time.Time) (Listas, error) {
	l := Listas{Muestra: []Muestra{}}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var p Programado
		var id int64
		err := t.QueryRow(ctx, `SELECT id, pct, redondeo, para FROM aumentos WHERE tienda_id = $1 AND aplicado_en IS NULL
			FOR UPDATE`, tiendaID).Scan(&id, &p.Pct, &p.Redondeo, &p.Para)
		switch {
		case errors.Is(err, pgx.ErrNoRows):
		case err != nil:
			return err
		case !p.Para.After(ahora):
			if _, err := t.Exec(ctx, `DELETE FROM aumentos WHERE id = $1`, id); err != nil {
				return err
			}
			if _, err := aplicar(ctx, t, tiendaID, p.Pct, p.Redondeo); err != nil {
				return err
			}
		default:
			l.Programado = &p
		}
		var mayoristas int
		if err := t.QueryRow(ctx, `SELECT count(DISTINCT comprador->>'nombre') FROM pedidos WHERE tienda_id = $1 AND mayorista`, tiendaID).Scan(&mayoristas); err != nil {
			return err
		}
		detalle := "clientes mayoristas"
		if mayoristas == 1 {
			detalle = "cliente mayorista"
		}
		l.Listas = []Lista{
			{"minorista", "Minorista", "Tienda y local", 0, "Precio de referencia"},
			{"mayorista", "Mayorista", "Desde 10 unidades", -25, itoa(mayoristas) + " " + detalle},
			{"ml", "Mercado Libre", "Cubre la comisión", 10, "Se actualiza solo"},
			{"revendedoras", "Revendedoras", "Para quienes revenden", -15, "Con link propio"},
		}
		if err := t.QueryRow(ctx, `SELECT count(*) FROM productos WHERE tienda_id = $1`, tiendaID).Scan(&l.Productos); err != nil {
			return err
		}
		filas, err := t.Query(ctx, `SELECT id::text, nombre, precio FROM productos WHERE tienda_id = $1 ORDER BY nombre LIMIT 4`, tiendaID)
		if err != nil {
			return err
		}
		l.Muestra, err = pgx.CollectRows(filas, pgx.RowToStructByPos[Muestra])
		return err
	})
	return l, err
}

func itoa(n int) string { return strconv.Itoa(n) }

// Rutas del panel.
func Rutas(mux *http.ServeMux, pool tx, tienda func(*http.Request) string) {
	con := func(h func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id := tienda(r)
			if id == "" {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			h(w, r, id)
		}
	}
	escribir := func(w http.ResponseWriter, v any) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}
	sinContenido := func(w http.ResponseWriter, err error, msg string) {
		switch {
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado(msg))
		case err != nil:
			problema.Escribir(w, problema.Interno())
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}
	mux.HandleFunc("GET /v1/panel/listas-precio", con(func(w http.ResponseWriter, r *http.Request, id string) {
		l, err := Leer(r.Context(), pool, id, time.Now())
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		escribir(w, l)
	}))
	mux.HandleFunc("POST /v1/panel/listas-precio/aumento", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var c struct {
			Pct       int    `json:"pct"`
			Redondeo  string `json:"redondeo"`
			Programar bool   `json:"programar"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&c); err != nil {
			problema.Escribir(w, problema.Invalido("Mandá el porcentaje y el redondeo."))
			return
		}
		n, para, err := Aplicar(r.Context(), pool, id, c.Pct, c.Redondeo, c.Programar, time.Now())
		if err != nil {
			problema.Escribir(w, problema.Invalido("Revisá el porcentaje y el redondeo."))
			return
		}
		escribir(w, map[string]any{"productos": n, "para": para})
	}))
	mux.HandleFunc("POST /v1/panel/listas-precio/aumento/deshacer", con(func(w http.ResponseWriter, r *http.Request, id string) {
		sinContenido(w, Deshacer(r.Context(), pool, id), "No hay un aumento para deshacer.")
	}))
	mux.HandleFunc("DELETE /v1/panel/listas-precio/programado", con(func(w http.ResponseWriter, r *http.Request, id string) {
		sinContenido(w, Cancelar(r.Context(), pool, id), "No hay un aumento programado.")
	}))
}
