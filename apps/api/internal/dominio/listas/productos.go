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
	// Fotos es la galería (la primera es la portada). Sin mandarla, manda Foto: reemplaza la portada
	// y deja el resto (o vacía la galería si Foto es null).
	Fotos []string `json:"fotos"`
	// Opciones nil deja las que había; [] las saca (y con ellas las variantes).
	Opciones []Opcion `json:"opciones"`
	// Variantes: sólo las combinaciones que se quieren tocar; las demás se conservan o se generan.
	Variantes []VarianteEditable `json:"variantes"`
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
	Fotos       []string        `json:"fotos"`
	Opciones    []Opcion        `json:"opciones"`
	Variantes   []Variante      `json:"variantes"`
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

// Guardar crea (id vacío) o actualiza un producto y devuelve su ficha. Con opciones, genera o ajusta
// las variantes y el stock del producto pasa a ser la suma de las activas.
func Guardar(ctx context.Context, pool tx, tiendaID, id string, e Editable) (Producto, error) {
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var err error
		id, err = GuardarEnTx(ctx, t, tiendaID, id, e)
		return err
	})
	if err != nil {
		return Producto{}, err
	}
	return Obtener(ctx, pool, tiendaID, id, time.Now())
}

// GuardarEnTx es Guardar dentro de una transacción con la tienda fijada, sin armar la ficha (que lee
// todo el catálogo): lo usa la importación para guardar un lote con las mismas reglas que el panel.
// Devuelve el id.
func GuardarEnTx(ctx context.Context, t pgx.Tx, tiendaID, id string, e Editable) (string, error) {
	if err := e.Validar(); err != nil {
		return "", err
	}
	canales, _ := json.Marshal(e.Canales)
	if e.Canales == nil {
		canales = []byte(`{"web": true}`)
	}
	var fotosPedidas []string
	if e.Fotos != nil {
		var err error
		if fotosPedidas, err = limpiarFotos(e.Fotos); err != nil {
			return "", err
		}
	}
	var opcionesPedidas []Opcion
	if e.Opciones != nil {
		var err error
		if opcionesPedidas, err = limpiarOpciones(e.Opciones); err != nil {
			return "", err
		}
	}
	opciones, fotos := []Opcion{}, []string{}
	if id != "" {
		err := t.QueryRow(ctx, `SELECT opciones, fotos FROM productos WHERE tienda_id = $1 AND id::text = $2 FOR UPDATE`,
			tiendaID, id).Scan(&opciones, &fotos)
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrNoExiste
		}
		if err != nil {
			return "", err
		}
	}
	switch {
	case e.Fotos != nil:
		fotos = fotosPedidas
	case e.Foto == nil:
		fotos = []string{}
	default:
		resto := []string{}
		for i, f := range fotos {
			if i > 0 && f != *e.Foto {
				resto = append(resto, f)
			}
		}
		fotos = append([]string{*e.Foto}, resto...)
	}
	var foto *string
	if len(fotos) > 0 {
		foto = &fotos[0]
	}
	if e.Opciones != nil {
		opciones = opcionesPedidas
	}
	if len(e.Variantes) > 0 && len(opciones) == 0 {
		return "", errors.New("para cargar variantes, el producto tiene que tener opciones")
	}
	// Con opciones, el stock lo pone la suma de las variantes (no el que manda el editor).
	stock := e.Stock
	if len(opciones) > 0 {
		stock = nil
	}
	if id == "" {
		if err := t.QueryRow(ctx, `INSERT INTO productos (tienda_id, nombre, descripcion, categoria, precio, costo, sku, stock, foto, canales, fotos, opciones)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id`,
			tiendaID, strings.TrimSpace(e.Nombre), valor(e.Descripcion, ""), e.Categoria, e.Precio, valor(e.Costo, 0),
			valor(e.SKU, ""), valor(stock, 0), foto, canales, fotos, opciones).Scan(&id); err != nil {
			return "", err
		}
	} else if _, err := t.Exec(ctx, `UPDATE productos SET nombre = $3, descripcion = coalesce($4, descripcion),
			categoria = $5, precio = $6, costo = coalesce($7, costo), sku = coalesce($8, sku),
			stock = coalesce($9, stock), foto = $10, canales = $11, publicado = true, fotos = $12, opciones = $13
		WHERE tienda_id = $1 AND id::text = $2`,
		tiendaID, id, strings.TrimSpace(e.Nombre), e.Descripcion, e.Categoria, e.Precio, e.Costo, e.SKU, stock, foto, canales,
		fotos, opciones); err != nil {
		return "", err
	}
	if err := sincronizarVariantes(ctx, t, tiendaID, id, opciones, fotos, e.Variantes); err != nil {
		return "", err
	}
	if len(opciones) == 0 {
		return id, nil
	}
	// El trigger suma con cada cambio de variante; esto cubre el caso de opciones nuevas sin cambios.
	_, err := t.Exec(ctx, `UPDATE productos SET stock = (SELECT coalesce(sum(stock), 0) FROM variantes
			WHERE tienda_id = $1 AND producto_id::text = $2 AND activa)
		WHERE tienda_id = $1 AND id::text = $2`, tiendaID, id)
	return id, err
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
		porProducto, err := variantesDe(ctx, t, tiendaID)
		if err != nil {
			return err
		}
		filas, err := t.Query(ctx, `SELECT id, nombre, descripcion, categoria, precio, costo, sku, stock, foto, canales, publicado, fotos, opciones
			FROM productos WHERE tienda_id = $1 ORDER BY creado_en DESC`, tiendaID)
		if err != nil {
			return err
		}
		for filas.Next() {
			var p Producto
			if err := filas.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Categoria, &p.Precio, &p.Costo, &p.SKU, &p.Stock, &p.Foto, &p.Canales, &p.Publicado, &p.Fotos, &p.Opciones); err != nil {
				return err
			}
			p.Variantes = porProducto[p.ID]
			if p.Variantes == nil {
				p.Variantes = []Variante{}
			}
			OrdenarVariantes(p.Opciones, p.Variantes)
			resumen := make([]Registro, 0, len(p.Variantes))
			for _, v := range p.Variantes {
				resumen = append(resumen, Registro{"id": v.ID, "texto": TextoVariante(p.Opciones, v.Valores), "sku": v.SKU,
					"precio": v.Precio, "stock": v.Stock, "activa": v.Activa})
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
				"estado": estado, "lowm": margen < 25, "star": p.Vendidos30d > 60, "variantes": resumen, "_ficha": p,
			})
		}
		return filas.Err()
	})
	return regs, err
}

// DefProductos: cómo filtra y ordena Productos.
var DefProductos = Def{
	Buscar: func(r Registro) string {
		return r["nombre"].(string) + " " + r["sku"].(string) + " " + r["categoria"].(string) + skusDe(r)
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
		if err != nil {
			return err
		}
		// Las variantes con precio propio suben (o bajan) igual.
		_, err = t.Exec(ctx, `UPDATE variantes SET precio = greatest(0, round(precio * (1 + $3 / 100.0) / 10) * 10)
			WHERE tienda_id = $1 AND producto_id::text = ANY($2) AND precio IS NOT NULL`, tiendaID, ids, pct)
		return err
	})
	return int(n), err
}

// skusDe suma los SKU de las variantes a la búsqueda.
func skusDe(r Registro) string {
	vs, _ := r["variantes"].([]Registro)
	var b strings.Builder
	for _, v := range vs {
		if s, _ := v["sku"].(string); s != "" {
			b.WriteString(" " + s)
		}
	}
	return b.String()
}
