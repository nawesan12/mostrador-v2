// Package mitienda: el editor de Mi tienda (design/PanelOperacion.dc.html). El diseño se guarda
// como borrador y se publica aparte; la tienda online lee el publicado.
package mitienda

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}

var ErrInvalido = errors.New("diseño inválido")

type Seccion struct {
	ID      string `json:"id"`
	Visible bool   `json:"visible"`
}

// Diseno es components/schemas/DisenoTienda.
type Diseno struct {
	Secciones []Seccion `json:"secciones"`
	Paleta    int       `json:"paleta"`
}

var ids = []string{"hero", "cats", "prods", "revs", "ig", "wa"}

// Inicial: las secciones como vienen en la plantilla.
func Inicial() Diseno {
	d := Diseno{}
	for _, id := range ids {
		d.Secciones = append(d.Secciones, Seccion{ID: id, Visible: true})
	}
	return d
}

func (d Diseno) valido() bool {
	if d.Paleta < 0 || d.Paleta > 3 || len(d.Secciones) != len(ids) {
		return false
	}
	vistos := map[string]bool{}
	for _, s := range d.Secciones {
		if !slices.Contains(ids, s.ID) || vistos[s.ID] {
			return false
		}
		vistos[s.ID] = true
	}
	return true
}

type producto struct {
	Nombre string  `json:"nombre"`
	Foto   *string `json:"foto"`
	Precio int64   `json:"precio"`
}

// Leer devuelve el editor con la vista previa armada con los productos de la tienda.
func Leer(ctx context.Context, pool tx, tiendaID, sitio string) (map[string]any, error) {
	var res map[string]any
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var nombre, slug, plantilla string
		var borrador, publicado *Diseno
		if err := t.QueryRow(ctx, `SELECT nombre, slug, plantilla, diseno_borrador, diseno_publicado FROM tiendas WHERE id = $1`,
			tiendaID).Scan(&nombre, &slug, &plantilla, &borrador, &publicado); err != nil {
			return err
		}
		ini := Inicial()
		if borrador == nil {
			borrador = &ini
		}
		if publicado == nil {
			publicado = &ini
		}
		a, _ := json.Marshal(borrador)
		b, _ := json.Marshal(publicado)
		filas, err := t.Query(ctx, `SELECT nombre, foto, precio FROM productos WHERE tienda_id = $1 AND publicado
			ORDER BY creado_en LIMIT 6`, tiendaID)
		if err != nil {
			return err
		}
		prods, err := pgx.CollectRows(filas, pgx.RowToStructByPos[producto])
		if err != nil {
			return err
		}
		var cats []string
		if err := t.QueryRow(ctx, `SELECT coalesce(array_agg(DISTINCT categoria), '{}') FROM productos
			WHERE tienda_id = $1 AND categoria IS NOT NULL`, tiendaID).Scan(&cats); err != nil {
			return err
		}
		fotos := []string{}
		for _, p := range prods {
			if p.Foto != nil {
				fotos = append(fotos, *p.Foto)
			}
		}
		var portada *string
		if len(fotos) > 0 {
			portada = &fotos[0]
		}
		cuatro := prods
		if len(cuatro) > 4 {
			cuatro = cuatro[:4]
		}
		res = map[string]any{
			"plantilla":   strings.ToUpper(plantilla[:1]) + plantilla[1:],
			"direccion":   strings.TrimPrefix(strings.TrimPrefix(sitio, "https://"), "http://") + "/t/" + slug,
			"diseno":      borrador,
			"sinPublicar": string(a) != string(b),
			"vista": map[string]any{
				"tienda": nombre, "categorias": cats, "productos": cuatro, "resena": nil, "fotos": fotos,
				"portada": map[string]any{"foto": portada, "titulo": nombre},
			},
		}
		return nil
	})
	return res, err
}

// Guardar el borrador.
func Guardar(ctx context.Context, pool tx, tiendaID string, d Diseno) error {
	if !d.valido() {
		return ErrInvalido
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE tiendas SET diseno_borrador = $2 WHERE id = $1`, tiendaID, d)
		return err
	})
}

// Publicar copia el borrador a lo que ve la gente.
func Publicar(ctx context.Context, pool tx, tiendaID string) error {
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE tiendas SET diseno_publicado = coalesce(diseno_borrador, diseno_publicado) WHERE id = $1`, tiendaID)
		return err
	})
}

// Rutas del panel.
func Rutas(mux *http.ServeMux, pool tx, tienda func(*http.Request) string, sitio string) {
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
	mux.HandleFunc("GET /v1/panel/mi-tienda", con(func(w http.ResponseWriter, r *http.Request, id string) {
		v, err := Leer(r.Context(), pool, id, sitio)
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}))
	mux.HandleFunc("PUT /v1/panel/mi-tienda", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var d Diseno
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&d); err != nil {
			problema.Escribir(w, problema.Invalido("Mandá las secciones y los colores."))
			return
		}
		if err := Guardar(r.Context(), pool, id, d); err != nil {
			problema.Escribir(w, problema.Invalido("Revisá las secciones."))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	mux.HandleFunc("POST /v1/panel/mi-tienda/publicar", con(func(w http.ResponseWriter, r *http.Request, id string) {
		if err := Publicar(r.Context(), pool, id); err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
}
