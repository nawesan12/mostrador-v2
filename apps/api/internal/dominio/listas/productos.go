package listas

import (
	"context"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// ErrNoExiste: el producto no es de esta tienda.
var ErrNoExiste = errors.New("el producto no existe")

// Editable es lo que se carga en el editor (components/schemas/ProductoEditable).
type Editable struct {
	Nombre      string          `json:"nombre"`
	Descripcion *string         `json:"descripcion"`
	Categoria   *string         `json:"categoria"`
	Precio      int64           `json:"precio"`
	Costo       *int64          `json:"costo"`
	SKU         *string         `json:"sku"`
	Stock       *int            `json:"stock"`
	Foto        *string         `json:"foto"`
	Canales     map[string]bool `json:"canales"`
}

// Producto es la ficha (components/schemas/Producto).
type Producto struct {
	ID          string          `json:"id"`
	Nombre      string          `json:"nombre"`
	Descripcion string          `json:"descripcion"`
	Categoria   *string         `json:"categoria"`
	Precio      int64           `json:"precio"`
	Costo       int64           `json:"costo"`
	SKU         string          `json:"sku"`
	Stock       int             `json:"stock"`
	Foto        *string         `json:"foto"`
	Canales     map[string]bool `json:"canales"`
	Publicado   bool            `json:"publicado"`
	Vendidos30d int             `json:"vendidos30d"`
	Slug        string          `json:"slug"`
}

var noSlug = regexp.MustCompile(`[^a-z0-9]+`)

// Slug arma la parte de la URL del producto.
func Slug(nombre string) string {
	s := strings.Trim(noSlug.ReplaceAllString(normal(nombre), "-"), "-")
	if len(s) > 60 {
		s = s[:60]
	}
	return s
}

// Validar revisa lo que se carga.
func (e Editable) Validar() error {
	n := len([]rune(strings.TrimSpace(e.Nombre)))
	if n < 2 || n > 120 {
		return errors.New("el nombre tiene que tener entre 2 y 120 letras")
	}
	if e.Precio < 0 || (e.Costo != nil && *e.Costo < 0) || (e.Stock != nil && *e.Stock < 0) {
		return errors.New("precio, costo y stock no pueden ser negativos")
	}
	return nil
}

func valor[T any](p *T, def T) T {
	if p == nil {
		return def
	}
	return *p
}

// Guardar crea (id vacío) o actualiza un producto y devuelve su ficha.
func Guardar(ctx context.Context, pool tx, tiendaID, id string, e Editable) (Producto, error) {
	if err := e.Validar(); err != nil {
		return Producto{}, err
	}
	canales, _ := json.Marshal(e.Canales)
	if e.Canales == nil {
		canales = []byte(`{"web": true}`)
	}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if id == "" {
			return t.QueryRow(ctx, `INSERT INTO productos (tienda_id, nombre, descripcion, categoria, precio, costo, sku, stock, foto, canales)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`,
				tiendaID, strings.TrimSpace(e.Nombre), valor(e.Descripcion, ""), e.Categoria, e.Precio, valor(e.Costo, 0),
				valor(e.SKU, ""), valor(e.Stock, 0), e.Foto, canales).Scan(&id)
		}
		tag, err := t.Exec(ctx, `UPDATE productos SET nombre = $3, descripcion = coalesce($4, descripcion),
				categoria = $5, precio = $6, costo = coalesce($7, costo), sku = coalesce($8, sku),
				stock = coalesce($9, stock), foto = $10, canales = $11, publicado = true
			WHERE tienda_id = $1 AND id::text = $2`,
			tiendaID, id, strings.TrimSpace(e.Nombre), e.Descripcion, e.Categoria, e.Precio, e.Costo, e.SKU, e.Stock, e.Foto, canales)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrNoExiste
		}
		return err
	})
	if err != nil {
		return Producto{}, err
	}
	return Obtener(ctx, pool, tiendaID, id, time.Now())
}

// Obtener devuelve la ficha de un producto.
func Obtener(ctx context.Context, pool tx, tiendaID, id string, ahora time.Time) (Producto, error) {
	regs, err := Productos(ctx, pool, tiendaID, ahora)
	if err != nil {
		return Producto{}, err
	}
	for _, r := range regs {
		if r["id"] == id {
			p := r["_ficha"].(Producto)
			return p, nil
		}
	}
	return Producto{}, ErrNoExiste
}

// Productos arma la lista con lo vendido (por nombre, de los pedidos de las últimas 8 semanas).
func Productos(ctx context.Context, pool tx, tiendaID string, ahora time.Time) ([]Registro, error) {
	regs := []Registro{}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		semanas := map[string][]int{}
		ultimos30 := map[string]int{}
		v, err := t.Query(ctx, `SELECT i->>'nombre', (i->>'cantidad')::int, creado_en
			FROM pedidos, jsonb_array_elements(items) i
			WHERE tienda_id = $1 AND estado <> 'Cancelado' AND creado_en > $2`, tiendaID, ahora.Add(-56*24*time.Hour))
		if err != nil {
			return err
		}
		for v.Next() {
			var n string
			var c int
			var cu time.Time
			if err := v.Scan(&n, &c, &cu); err != nil {
				return err
			}
			if semanas[n] == nil {
				semanas[n] = make([]int, 8)
			}
			if k := 7 - int(ahora.Sub(cu).Hours()/24/7); k >= 0 && k < 8 {
				semanas[n][k] += c
			}
			if ahora.Sub(cu) < 30*24*time.Hour {
				ultimos30[n] += c
			}
		}
		if err := v.Err(); err != nil {
			return err
		}
		filas, err := t.Query(ctx, `SELECT id, nombre, descripcion, categoria, precio, costo, sku, stock, foto, canales, publicado
			FROM productos WHERE tienda_id = $1 ORDER BY creado_en DESC`, tiendaID)
		if err != nil {
			return err
		}
		for filas.Next() {
			var p Producto
			if err := filas.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Categoria, &p.Precio, &p.Costo, &p.SKU, &p.Stock, &p.Foto, &p.Canales, &p.Publicado); err != nil {
				return err
			}
			p.Vendidos30d, p.Slug = ultimos30[p.Nombre], Slug(p.Nombre)
			ventas := semanas[p.Nombre]
			if ventas == nil {
				ventas = make([]int, 8)
			}
			margen := 0
			if p.Precio > 0 {
				margen = int((p.Precio - p.Costo) * 100 / p.Precio)
			}
			canales := []string{}
			for k, on := range p.Canales {
				if on {
					canales = append(canales, k)
				}
			}
			estado := "active"
			switch {
			case !p.Publicado && p.Foto == nil:
				estado = "draft"
			case !p.Publicado:
				estado = "paused"
			case p.Stock == 0:
				estado = "out"
			}
			cat := "Sin categoría"
			if p.Categoria != nil && *p.Categoria != "" {
				cat = *p.Categoria
			}
			regs = append(regs, Registro{
				"id": p.ID, "nombre": p.Nombre, "descripcion": p.Descripcion, "foto": p.Foto, "sku": p.SKU,
				"categoria": cat, "precio": p.Precio, "costo": p.Costo, "margen": margen, "stock": p.Stock,
				"ventas": ventas, "vendidos30": p.Vendidos30d, "canales": canales, "publicado": p.Publicado,
				"estado": estado, "lowm": margen < 25, "star": p.Vendidos30d > 60, "_ficha": p,
			})
		}
		return filas.Err()
	})
	return regs, err
}

// DefProductos: cómo filtra y ordena Productos.
var DefProductos = Def{
	Buscar: func(r Registro) string {
		return r["nombre"].(string) + " " + r["sku"].(string) + " " + r["categoria"].(string)
	},
	Segmento: "estado", Facetas: []string{"categoria", "canales"}, Rango: "precio", Banderas: []string{"lowm", "star"},
	Ordenes: []Orden{
		{"sold", func(a, b Registro) bool { return numero(a["vendidos30"]) > numero(b["vendidos30"]) }},
		{"stock", func(a, b Registro) bool { return numero(a["stock"]) < numero(b["stock"]) }},
		{"price", func(a, b Registro) bool { return numero(a["precio"]) > numero(b["precio"]) }},
		{"margin", func(a, b Registro) bool { return numero(a["margen"]) > numero(b["margen"]) }},
	},
	Resumen: func(todos, l []Registro) map[string]float64 {
		r := map[string]float64{"precioMax": 10000}
		for _, x := range l {
			r["unidades30"] += numero(x["vendidos30"])
		}
		for _, x := range todos {
			r["precioMax"] = max(r["precioMax"], numero(x["precio"]))
			cs, _ := x["canales"].([]string)
			for _, c := range cs {
				if c == "ml" && x["estado"] == "out" {
					r["sinStockEnMl"]++
				}
			}
		}
		return r
	},
}

// Publicar prende o apaga productos; Precio los sube o baja un porcentaje (redondeo a $ 10).
func Publicar(ctx context.Context, pool tx, tiendaID string, ids []string, on bool) (int, error) {
	var n int64
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `UPDATE productos SET publicado = $3 WHERE tienda_id = $1 AND id::text = ANY($2)`, tiendaID, ids, on)
		n = tag.RowsAffected()
		return err
	})
	return int(n), err
}

// Precio cambia el precio un porcentaje.
func Precio(ctx context.Context, pool tx, tiendaID string, ids []string, pct float64) (int, error) {
	if pct <= -90 || pct > 200 {
		return 0, errors.New("porcentaje fuera de rango")
	}
	var n int64
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `UPDATE productos SET precio = greatest(0, round(precio * (1 + $3 / 100.0) / 10) * 10)
			WHERE tienda_id = $1 AND id::text = ANY($2)`, tiendaID, ids, pct)
		n = tag.RowsAffected()
		return err
	})
	return int(n), err
}
