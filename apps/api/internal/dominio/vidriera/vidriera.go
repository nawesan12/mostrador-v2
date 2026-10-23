// Package vidriera: la tienda online pública (/t/{slug}). Lo publicado, sin costos ni datos de
// clientes; comprar arma el pedido con precios y envío calculados acá, y con Mercado Pago devuelve el
// checkout de la cuenta de la tienda.
package vidriera

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
)

var (
	ErrNoExiste = errors.New("vidriera: no existe")
	ErrInvalido = errors.New("vidriera: pedido inválido")
	ErrSinStock = errors.New("vidriera: sin stock")
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
	EnTx(context.Context, func(pgx.Tx) error) error
}

// Pagador cobra un pedido con Mercado Pago (el paquete mercadopago).
type Pagador interface {
	PagarPedido(ctx context.Context, tiendaID string, numero int, vuelta string) (string, error)
}

type ProductoPublico struct {
	ID          string  `json:"id"`
	Nombre      string  `json:"nombre"`
	Descripcion string  `json:"descripcion"`
	Categoria   *string `json:"categoria"`
	Precio      int64   `json:"precio"`
	Anterior    *int64  `json:"anterior"`
	Foto        *string `json:"foto"`
	Etiqueta    *string `json:"etiqueta"`
	Stock       int     `json:"stock"`
}

type Categoria struct {
	Nombre   string  `json:"nombre"`
	Cantidad int     `json:"cantidad"`
	Foto     *string `json:"foto"`
}

// Tienda es components/schemas/TiendaPublica (lib/tienda/tipos.ts).
type Tienda struct {
	Slug        string            `json:"slug"`
	Nombre      string            `json:"nombre"`
	Plantilla   string            `json:"plantilla"`
	Paleta      int               `json:"paleta"`
	Secciones   []string          `json:"secciones"`
	Portada     map[string]any    `json:"portada"`
	Ciudad      string            `json:"ciudad"`
	Direccion   string            `json:"direccion"`
	Whatsapp    string            `json:"whatsapp"`
	Categorias  []Categoria       `json:"categorias"`
	Productos   []ProductoPublico `json:"productos"`
	MasVendidos []string          `json:"masVendidos"`
	Resenas     []any             `json:"resenas"`
	Promedio    *float64          `json:"promedio"`
	Fotos       []string          `json:"fotos"`
	Cobros      map[string]any    `json:"cobros"`
	Envios      Envios            `json:"envios"`
	Vitrina     bool              `json:"vitrina"`
}

type Envios struct {
	GratisDesde *int64 `json:"gratisDesde"`
	Retiro      bool   `json:"retiro"`
	Moto        bool   `json:"moto"`
	Correo      bool   `json:"correo"`
}

// Cobros por defecto hasta que la tienda los configure.
const (
	transferencia = 15
	cuotas        = 6
)

var slugValido = regexp.MustCompile(`^[a-z0-9-]{2,60}$`)

func tiendaDe(ctx context.Context, pool tx, slug string) (string, error) {
	if !slugValido.MatchString(slug) {
		return "", ErrNoExiste
	}
	var id *string
	if err := pool.EnTx(ctx, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT app.tienda_de_slug($1)::text`, slug).Scan(&id)
	}); err != nil {
		return "", err
	}
	if id == nil {
		return "", ErrNoExiste
	}
	return *id, nil
}

// Publica arma la tienda como la ve quien compra.
func Publica(ctx context.Context, pool tx, slug string) (Tienda, error) {
	id, err := tiendaDe(ctx, pool, slug)
	if err != nil {
		return Tienda{}, err
	}
	out := Tienda{Slug: slug, Secciones: []string{"hero", "cats", "prods", "revs", "ig", "wa"}, Resenas: []any{},
		Envios: Envios{Retiro: true, Correo: true}, Fotos: []string{}, MasVendidos: []string{}}
	// Los más vendidos de los últimos 30 días (los ítems de los pedidos guardan el nombre).
	vendidos := map[string]int{}
	err = pool.EnTienda(ctx, id, func(t pgx.Tx) error {
		var diseno, negocio []byte
		var mp bool
		if err := t.QueryRow(ctx, `SELECT nombre, plantilla, diseno_publicado, negocio,
				EXISTS (SELECT 1 FROM mp_conexiones WHERE tienda_id = $1 AND NOT reautorizar)
			FROM tiendas WHERE id = $1`, id).Scan(&out.Nombre, &out.Plantilla, &diseno, &negocio, &mp); err != nil {
			return err
		}
		var d struct {
			Secciones []struct {
				ID      string `json:"id"`
				Visible bool   `json:"visible"`
			} `json:"secciones"`
			Paleta int `json:"paleta"`
		}
		if json.Unmarshal(diseno, &d) == nil && len(d.Secciones) > 0 {
			out.Secciones = out.Secciones[:0]
			for _, s := range d.Secciones {
				if s.Visible {
					out.Secciones = append(out.Secciones, s.ID)
				}
			}
			out.Paleta = d.Paleta
		}
		var n struct{ Direccion, Whatsapp string }
		_ = json.Unmarshal(negocio, &n)
		out.Direccion, out.Whatsapp = n.Direccion, n.Whatsapp
		out.Cobros = map[string]any{"transferencia": transferencia, "cuotas": cuotas, "mercadoPago": mp}

		filas, err := t.Query(ctx, `SELECT id::text, nombre, descripcion, categoria, precio, foto, stock FROM productos
			WHERE tienda_id = $1 AND publicado AND coalesce((canales->>'web')::boolean, true) ORDER BY creado_en DESC LIMIT 500`, id)
		if err != nil {
			return err
		}
		for filas.Next() {
			var p ProductoPublico
			if err := filas.Scan(&p.ID, &p.Nombre, &p.Descripcion, &p.Categoria, &p.Precio, &p.Foto, &p.Stock); err != nil {
				return err
			}
			if p.Stock > 0 && p.Stock <= 3 {
				u := "Últimos"
				p.Etiqueta = &u
			}
			out.Productos = append(out.Productos, p)
		}
		if err := filas.Err(); err != nil {
			return err
		}
		rows, err := t.Query(ctx, `SELECT i->>'nombre', sum((i->>'cantidad')::int) FROM pedidos, jsonb_array_elements(items) i
			WHERE tienda_id = $1 AND estado <> 'Cancelado' AND creado_en > now() - interval '30 days' GROUP BY 1`, id)
		if err != nil {
			return err
		}
		for rows.Next() {
			var nombre string
			var c int
			if err := rows.Scan(&nombre, &c); err != nil {
				return err
			}
			vendidos[nombre] = c
		}
		return rows.Err()
	})
	if err != nil {
		return Tienda{}, err
	}
	if out.Productos == nil {
		out.Productos = []ProductoPublico{}
	}
	orden := append([]ProductoPublico(nil), out.Productos...)
	sort.SliceStable(orden, func(i, j int) bool { return vendidos[orden[i].Nombre] > vendidos[orden[j].Nombre] })
	cats := map[string]*Categoria{}
	var nombres []string
	for _, p := range orden {
		out.MasVendidos = append(out.MasVendidos, p.ID)
		if p.Foto != nil && len(out.Fotos) < 6 && !contiene(out.Fotos, *p.Foto) {
			out.Fotos = append(out.Fotos, *p.Foto)
		}
		if p.Categoria != nil && *p.Categoria != "" {
			c, ok := cats[*p.Categoria]
			if !ok {
				c = &Categoria{Nombre: *p.Categoria}
				cats[*p.Categoria] = c
				nombres = append(nombres, *p.Categoria)
			}
			c.Cantidad++
			if c.Foto == nil {
				c.Foto = p.Foto
			}
		}
	}
	for _, n := range nombres {
		out.Categorias = append(out.Categorias, *cats[n])
	}
	if out.Categorias == nil {
		out.Categorias = []Categoria{}
	}
	var foto *string
	if len(out.Fotos) > 0 {
		foto = &out.Fotos[0]
	}
	out.Portada = map[string]any{"titulo": out.Nombre, "lema": "", "foto": foto}
	return out, nil
}

func contiene(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}

// Compra es lo que manda el checkout (precios y envío NO vienen: se calculan acá).
type Compra struct {
	Carrito []struct {
		ID       string `json:"id"`
		Cantidad int    `json:"cantidad"`
	} `json:"carrito"`
	Email     string `json:"email"`
	Nombre    string `json:"nombre"`
	Telefono  string `json:"telefono"`
	DNI       string `json:"dni"`
	CP        string `json:"cp"`
	Direccion string `json:"direccion"`
	Envio     string `json:"envio"`
	Pago      string `json:"pago"`
}

// Resultado de comprar: el pedido, su código público y (con Mercado Pago) adónde pagar.
type Resultado struct {
	Numero int    `json:"numero"`
	Codigo string `json:"codigo"`
	Total  int64  `json:"total"`
	URL    string `json:"url,omitempty"`
}

var emailValido = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// Comprar arma el pedido de la tienda online.
func Comprar(ctx context.Context, pool tx, mp Pagador, sitio, slug string, c Compra, hoy time.Time) (Resultado, error) {
	t, err := Publica(ctx, pool, slug)
	if err != nil {
		return Resultado{}, err
	}
	id, _ := tiendaDe(ctx, pool, slug)
	c.Nombre, c.Email = strings.TrimSpace(c.Nombre), strings.TrimSpace(c.Email)
	if len(c.Carrito) == 0 || len(c.Carrito) > 50 || len(c.Nombre) < 2 || len(c.Nombre) > 80 || !emailValido.MatchString(c.Email) ||
		(c.Pago != "transferencia" && c.Pago != "mercadopago") {
		return Resultado{}, ErrInvalido
	}
	if c.Pago == "mercadopago" && (mp == nil || t.Cobros["mercadoPago"] != true) {
		return Resultado{}, ErrInvalido
	}
	porID := map[string]ProductoPublico{}
	for _, p := range t.Productos {
		porID[p.ID] = p
	}
	var subtotal int64
	var n pedidos.PedidoNuevo
	for _, l := range c.Carrito {
		p, ok := porID[l.ID]
		if !ok || l.Cantidad < 1 || l.Cantidad > 99 {
			return Resultado{}, ErrInvalido
		}
		subtotal += p.Precio * int64(l.Cantidad)
		n.Items = append(n.Items, struct {
			ProductoID string `json:"productoId"`
			Cantidad   int    `json:"cantidad"`
		}{l.ID, l.Cantidad})
	}
	var opcion *Opcion
	for _, o := range Cotizar(c.CP, subtotal, t.Envios, hoy) {
		if o.ID == c.Envio {
			o := o
			opcion = &o
		}
	}
	if opcion == nil || (c.Envio != "retiro" && len(strings.TrimSpace(c.Direccion)) < 4) {
		return Resultado{}, ErrInvalido
	}
	n.Canal, n.Envio = "web", c.Envio
	n.Pago = "transf"
	if c.Pago == "mercadopago" {
		n.Pago = "mp"
	} else {
		n.Descuento = subtotal * transferencia / 100
	}
	f := false
	n.Pagado = &f
	n.Cliente, n.Telefono = &c.Nombre, &c.Telefono
	n.Email, n.CostoEnvio, n.CiudadEnvio = c.Email, opcion.Precio, strings.TrimSpace(c.CP)
	nota := strings.TrimSpace(fmt.Sprintf("Tienda online · DNI %s · Entrega: %s (%s)", c.DNI, c.Direccion, c.CP))
	n.Nota = &nota
	b := make([]byte, 18)
	_, _ = rand.Read(b)
	n.CodigoPublico = base64.RawURLEncoding.EncodeToString(b)
	numero, total, err := pedidos.CrearDesdeProductos(ctx, pool, id, n)
	if errors.Is(err, pedidos.ErrSinStock) {
		return Resultado{}, ErrSinStock
	}
	if err != nil {
		return Resultado{}, err
	}
	r := Resultado{Numero: numero, Codigo: n.CodigoPublico, Total: total}
	if c.Pago == "mercadopago" {
		r.URL, err = mp.PagarPedido(ctx, id, numero, fmt.Sprintf("%s/t/%s/pedido/%s", sitio, slug, r.Codigo))
	}
	return r, err
}

// Confirmacion es lo que muestra la página del pedido.
type Confirmacion struct {
	Numero     int    `json:"numero"`
	Total      int64  `json:"total"`
	Pago       string `json:"pago"`
	Nombre     string `json:"nombre"`
	EstadoPago string `json:"estadoPago"`
}

// VerPedido por su código público.
func VerPedido(ctx context.Context, pool tx, slug, codigo string) (Confirmacion, error) {
	id, err := tiendaDe(ctx, pool, slug)
	if err != nil {
		return Confirmacion{}, err
	}
	var c Confirmacion
	err = pool.EnTienda(ctx, id, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT numero, total, CASE WHEN pago = 'mp' THEN 'mercadopago' ELSE 'transferencia' END,
				split_part(comprador->>'nombre', ' ', 1), estado_pago
			FROM pedidos WHERE codigo_publico = $1`, codigo).Scan(&c.Numero, &c.Total, &c.Pago, &c.Nombre, &c.EstadoPago)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return c, ErrNoExiste
	}
	return c, err
}
