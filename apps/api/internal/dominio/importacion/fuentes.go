package importacion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/importar"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadolibre"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/tiendanube"
)

// Las fuentes reales, una por plataforma. Cada una lee una página por paso, en la etapa que le
// toca, y la pasa a lo que guarda Mostrador; las etapas que la plataforma no tiene terminan enseguida.

// Dependencias: lo que las fuentes necesitan del servidor.
type Dependencias struct {
	// Publico lee tiendas públicas (Shopify, WooCommerce): https a IPs públicas (importar.Publico).
	Publico *http.Client
	// MercadoLibre es el canal ya conectado de la tienda (su token).
	MercadoLibre *mercadolibre.Servicio
	// Tiendanube es el cliente HTTP para su API (nil: el de siempre).
	Tiendanube *http.Client
}

// Fuentes devuelve las fuentes por plataforma. Mientras falte una (CSV, Empretienda), crear una
// importación desde ella responde sin_fuente.
func Fuentes(d Dependencias) map[string]Abrir {
	if d.Publico == nil {
		d.Publico = importar.Publico()
	}
	f := map[string]Abrir{
		"shopify":     func(_ context.Context, o Origen) (Fuente, error) { return shopify{d.Publico, o.URL}, nil },
		"woocommerce": func(_ context.Context, o Origen) (Fuente, error) { return woo{d.Publico, o.URL}, nil },
		"tiendanube": func(_ context.Context, o Origen) (Fuente, error) {
			var c credencialTN
			if err := json.Unmarshal(o.Credencial, &c); err != nil || c.Token == "" || c.Tienda == "" {
				return nil, fmt.Errorf("%w: falta conectar la tienda de Tiendanube", ErrDefinitivo)
			}
			return tn{tiendanube.API{Token: c.Token, TiendaID: c.Tienda, HTTP: d.Tiendanube}}, nil
		},
	}
	f["mercadolibre"] = func(_ context.Context, o Origen) (Fuente, error) {
		if d.MercadoLibre == nil || !d.MercadoLibre.Activo() {
			return nil, fmt.Errorf("%w: Mercado Libre no está disponible en este servidor", ErrDefinitivo)
		}
		return ml{d.MercadoLibre, o.TiendaID}, nil
	}
	return f
}

// cursorPagina es el cursor de las fuentes que paginan por número.
type cursorPagina struct {
	Pagina int `json:"pagina"`
}

func paginaDe(cursor json.RawMessage) int {
	var c cursorPagina
	if json.Unmarshal(cursor, &c) != nil || c.Pagina < 1 {
		return 1
	}
	return c.Pagina
}

func siguientePagina(n int) json.RawMessage {
	b, _ := json.Marshal(cursorPagina{Pagina: n + 1})
	return b
}

// productosDe pasa lo importado a la página de la cola.
func productosDe(items []importar.Importado) []ProductoViejo {
	ps := make([]ProductoViejo, 0, len(items))
	for _, i := range items {
		ps = append(ps, ProductoViejo{Externo: i.Externo, URL: i.URL, Producto: i.Producto.Editable()})
	}
	return ps
}

// definitivo: una tienda que no existe o ya no publica no se arregla reintentando.
func definitivo(err error) error {
	var e *importar.ErrHTTP
	if errors.Is(err, importar.ErrNoPublica) || (errors.As(err, &e) && (e.Estado == http.StatusNotFound || e.Estado == http.StatusForbidden || e.Estado == http.StatusUnauthorized)) {
		return fmt.Errorf("%w: %v", ErrDefinitivo, err)
	}
	return err
}

// shopify lee /products.json de a 250. Sin cuenta no hay clientes ni pedidos.
type shopify struct {
	http   *http.Client
	origen string
}

const porPaginaShopify = 250

func (s shopify) Pagina(ctx context.Context, etapa string, cursor json.RawMessage) (Pagina, error) {
	if etapa != "productos" {
		return Pagina{}, nil
	}
	n := paginaDe(cursor)
	b, _, err := importar.Pedir(ctx, s.http, fmt.Sprintf("%s/products.json?limit=%d&page=%d", s.origen, porPaginaShopify, n))
	if err != nil {
		return Pagina{}, definitivo(err)
	}
	items, leidos, _, err := importar.PaginaShopify(b)
	if err != nil {
		return Pagina{}, fmt.Errorf("%w: la tienda no respondió como Shopify", ErrDefinitivo)
	}
	p := Pagina{Productos: productosDe(items)}
	if leidos == porPaginaShopify {
		p.Siguiente = siguientePagina(n)
	}
	return p, nil
}

// woo lee la Store API de a 100, con las variaciones de los productos variables de la página.
type woo struct {
	http   *http.Client
	origen string
}

const (
	apiWoo        = "/wp-json/wc/store/v1/products"
	porPaginaWoo  = 100
	maxPaginasVar = 10
)

func (w woo) Pagina(ctx context.Context, etapa string, cursor json.RawMessage) (Pagina, error) {
	if etapa != "productos" {
		return Pagina{}, nil
	}
	n := paginaDe(cursor)
	b, h, err := importar.Pedir(ctx, w.http, fmt.Sprintf("%s%s?per_page=%d&page=%d", w.origen, apiWoo, porPaginaWoo, n))
	if err != nil {
		var e *importar.ErrHTTP
		// Pasada la última página, la Store API responde 400 (rest_post_invalid_page_number).
		if errors.As(err, &e) && e.Estado == http.StatusBadRequest && n > 1 {
			return Pagina{}, nil
		}
		return Pagina{}, definitivo(err)
	}
	var productos []importar.ProductoWoo
	if err := json.Unmarshal(b, &productos); err != nil {
		return Pagina{}, fmt.Errorf("%w: la tienda no respondió como WooCommerce", ErrDefinitivo)
	}
	var padres []string
	for _, p := range productos {
		if p.Type == "variable" {
			padres = append(padres, strconv.FormatInt(p.ID, 10))
		}
	}
	var variaciones []importar.ProductoWoo
	for i := 1; len(padres) > 0 && i <= maxPaginasVar; i++ {
		u := fmt.Sprintf("%s%s?type=variation&parent=%s&per_page=100&page=%d", w.origen, apiWoo, url.QueryEscape(strings.Join(padres, ",")), i)
		vb, _, err := importar.Pedir(ctx, w.http, u)
		if err != nil {
			return Pagina{}, definitivo(err)
		}
		var vs []importar.ProductoWoo
		if err := json.Unmarshal(vb, &vs); err != nil {
			return Pagina{}, fmt.Errorf("%w: las variaciones no vinieron como WooCommerce", ErrDefinitivo)
		}
		variaciones = append(variaciones, vs...)
		if len(vs) < 100 {
			break
		}
	}
	items, _, _ := importar.PaginaWoo(productos, variaciones)
	p := Pagina{Productos: productosDe(items)}
	total, err := strconv.Atoi(h.Get("X-WP-TotalPages"))
	if (err == nil && n < total) || (err != nil && len(productos) == porPaginaWoo) {
		p.Siguiente = siguientePagina(n)
	}
	return p, nil
}

// credencialTN es lo que se guarda (cifrado) al conectar la tienda de Tiendanube.
type credencialTN struct {
	Token  string `json:"token"`
	Tienda string `json:"tienda"`
}

// CredencialTiendanube arma la credencial de una importación desde Tiendanube con el canje del OAuth.
func CredencialTiendanube(t tiendanube.Token) []byte {
	b, _ := json.Marshal(credencialTN{Token: t.Acceso, Tienda: t.TiendaID})
	return b
}

// tn lee la API de Tiendanube: productos, clientes y pedidos de a 200. El cursor es la URL de la
// página siguiente (la que manda el header Link) y el idioma principal de la tienda.
type tn struct{ api tiendanube.API }

type cursorTN struct {
	URL    string `json:"url"`
	Idioma string `json:"idioma"`
}

var recursoTN = map[string]string{"productos": "products", "clientes": "customers", "pedidos": "orders"}

func (t tn) Pagina(ctx context.Context, etapa string, cursor json.RawMessage) (Pagina, error) {
	recurso, ok := recursoTN[etapa]
	if !ok {
		return Pagina{}, nil
	}
	var c cursorTN
	_ = json.Unmarshal(cursor, &c)
	if c.URL == "" {
		c.URL = t.api.Primera(recurso, nil)
	}
	pag, err := t.api.Leer(ctx, c.URL)
	if errors.Is(err, tiendanube.ErrReautorizar) {
		return Pagina{}, fmt.Errorf("%w: Tiendanube ya no deja leer la tienda; conectala de nuevo", ErrDefinitivo)
	}
	if err != nil {
		return Pagina{}, err
	}
	if pag.Idioma != "" {
		c.Idioma = pag.Idioma
	}
	var p Pagina
	for _, crudo := range pag.Registros {
		switch etapa {
		case "productos":
			var x tiendanube.ProductoTN
			if json.Unmarshal(crudo, &x) != nil {
				continue
			}
			if i, ok := tiendanube.Producto(x, c.Idioma); ok {
				p.Productos = append(p.Productos, ProductoViejo{Externo: i.Externo, URL: i.URLVieja, Producto: i.Producto.Editable()})
			}
		case "clientes":
			var x tiendanube.ClienteTN
			if json.Unmarshal(crudo, &x) != nil {
				continue
			}
			if v, ok := tiendanube.ClienteDe(x); ok {
				p.Clientes = append(p.Clientes, ClienteViejo{Externo: v.Externo, Nombre: v.Nombre, Email: v.Email, Telefono: v.Telefono, Ciudad: v.Ciudad, Provincia: v.Provincia, Notas: v.Notas})
			}
		case "pedidos":
			var x tiendanube.PedidoTN
			if json.Unmarshal(crudo, &x) != nil {
				continue
			}
			if v, ok := tiendanube.PedidoDe(x); ok {
				pv := PedidoViejo{Externo: v.Externo, Numero: v.NumeroExterno, Fecha: v.Fecha, Email: v.ClienteEmail, Nombre: v.ClienteNombre, Total: v.Total, Estado: v.Estado}
				for _, it := range v.Items {
					pv.Items = append(pv.Items, ItemViejo(it))
				}
				p.Pedidos = append(p.Pedidos, pv)
			}
		}
	}
	if pag.Siguiente != "" {
		p.Siguiente, _ = json.Marshal(cursorTN{URL: pag.Siguiente, Idioma: c.Idioma})
	}
	return p, nil
}

// ml lee las publicaciones de la cuenta de Mercado Libre ya conectada. Mercado Libre no da los datos
// de quien compra, así que de acá sólo salen productos.
type ml struct {
	s      *mercadolibre.Servicio
	tienda string
}

type cursorML struct {
	Scroll string `json:"scroll"`
}

func (m ml) Pagina(ctx context.Context, etapa string, cursor json.RawMessage) (Pagina, error) {
	if etapa != "productos" {
		return Pagina{}, nil
	}
	var c cursorML
	_ = json.Unmarshal(cursor, &c)
	pag, err := m.s.LeerPublicaciones(ctx, m.tienda, c.Scroll)
	if errors.Is(err, mercadolibre.ErrSinConexion) || errors.Is(err, mercadolibre.ErrReautorizar) || errors.Is(err, mercadolibre.ErrApagado) {
		return Pagina{}, fmt.Errorf("%w: conectá tu cuenta de Mercado Libre en Integraciones", ErrDefinitivo)
	}
	if err != nil {
		return Pagina{}, err
	}
	var p Pagina
	for _, it := range pag.Items {
		if i, ok := importar.Ajustar(it.Crudo()); ok {
			p.Productos = append(p.Productos, ProductoViejo{Externo: it.ID, Producto: i.Producto.Editable()})
		}
	}
	if pag.Siguiente != "" {
		p.Siguiente, _ = json.Marshal(cursorML{Scroll: pag.Siguiente})
	}
	return p, nil
}
