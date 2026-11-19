package tincho

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/roles"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
)

// Tope de filas de las herramientas de lectura (el modelo pide menos si le alcanza).
const maxFilas = 20

func esquema(props map[string]any, requeridos ...string) anthropic.ToolInputSchemaParam {
	return anthropic.ToolInputSchemaParam{Properties: props, Required: requeridos}
}

var limite = map[string]any{"type": "integer", "minimum": 1, "maximum": maxFilas, "description": "Cuántas filas (hasta 20; por defecto 10)."}
var estadosPedido = []string{"Sin pagar", "Para preparar", "Listo para despachar", "En camino", "Entregado", "Cancelado"}

// Herramientas: siempre las mismas para todas las tiendas y roles (cambiarlas invalida el caché);
// el permiso se mira al ejecutarlas. Las de acción tienen los mismos argumentos que las del MCP
// del BFF, que es quien las ejecuta cuando la persona confirma.
func Herramientas() []anthropic.ToolParam {
	return []anthropic.ToolParam{
		{Name: "resumen_del_dia", Description: anthropic.String("Cómo viene hoy: pedidos y ventas contra ayer, ticket, lo que hay para preparar, transferencias por confirmar, sin stock, la meta del mes y los últimos pedidos."),
			InputSchema: esquema(map[string]any{})},
		{Name: "informe_de_ventas", Description: anthropic.String("Ventas, pedidos, ticket y clientes nuevos de un período contra el anterior del mismo largo, por canal y los más vendidos."),
			InputSchema: esquema(map[string]any{"periodo": map[string]any{"type": "string", "enum": []string{"hoy", "7d", "30d"}}}, "periodo")},
		{Name: "listar_pedidos", Description: anthropic.String("Pedidos del más nuevo al más viejo, con filtros: número, cliente, estado, canal, total, fecha y productos."),
			InputSchema: esquema(map[string]any{
				"texto":  map[string]any{"type": "string", "description": "Busca por número, cliente, ciudad o producto."},
				"estado": map[string]any{"type": "string", "enum": estadosPedido},
				"canal":  map[string]any{"type": "string", "enum": []string{"web", "wa", "ml", "ig", "pos", "tg"}, "description": "web (tienda online), wa (WhatsApp), ml (Mercado Libre), ig (Instagram), pos (local), tg (Telegram)."},
				"fecha":  map[string]any{"type": "string", "enum": []string{"hoy", "7d", "30d", "todo"}},
				"limite": limite,
			})},
		{Name: "ver_pedido", Description: anthropic.String("El detalle de un pedido: productos, total, envío, pago, estado e historial."),
			InputSchema: esquema(map[string]any{"numero": map[string]any{"type": "integer", "minimum": 1}}, "numero")},
		{Name: "listar_productos", Description: anthropic.String("Productos con id, precio, stock, vendidos en 30 días, categoría y estado. El id sirve para ajustar_producto."),
			InputSchema: esquema(map[string]any{
				"texto":  map[string]any{"type": "string", "description": "Busca por nombre, SKU o categoría."},
				"estado": map[string]any{"type": "string", "enum": []string{"activo", "sin_stock", "pausado", "borrador"}},
				"orden":  map[string]any{"type": "string", "enum": []string{"mas_vendidos", "menos_stock", "mayor_precio"}},
				"limite": limite,
			})},
		{Name: "listar_clientes", Description: anthropic.String("Clientes con ciudad, canal, pedidos, total gastado, días desde la última compra y segmento."),
			InputSchema: esquema(map[string]any{
				"texto":    map[string]any{"type": "string", "description": "Busca por nombre o ciudad."},
				"segmento": map[string]any{"type": "string", "enum": []string{"vip", "frecuentes", "nuevos", "en_riesgo", "dormidos"}},
				"orden":    map[string]any{"type": "string", "enum": []string{"mas_gastaron", "compra_reciente", "mas_pedidos"}},
				"limite":   limite,
			})},
		{Name: "cambiar_estado_pedido", Description: anthropic.String("PROPONE pasar un pedido a otro estado (avanza una etapa, vuelve una o se cancela). No lo cambia: la persona confirma."),
			InputSchema: esquema(map[string]any{"numero": map[string]any{"type": "integer", "minimum": 1}, "estado": map[string]any{"type": "string", "enum": estadosPedido}}, "numero", "estado")},
		{Name: "ajustar_producto", Description: anthropic.String("PROPONE cambiar el precio (en pesos) o el stock de un producto sin variantes: \"stock\" con el total nuevo o \"sumar_stock\" con lo que entra (negativo si sale). No lo cambia: la persona confirma."),
			InputSchema: esquema(map[string]any{
				"id":          map[string]any{"type": "string", "description": "Id del producto (de listar_productos)."},
				"precio":      map[string]any{"type": "integer", "minimum": 0},
				"stock":       map[string]any{"type": "integer", "minimum": 0},
				"sumar_stock": map[string]any{"type": "integer"},
			}, "id")},
		{Name: "ajustar_precios", Description: anthropic.String("PROPONE subir o bajar en un porcentaje el precio de varios productos (con sus variantes), redondeado a $ 10: los ids de listar_productos o una categoría. Una sola propuesta para todos. No lo cambia: la persona confirma."),
			InputSchema: esquema(map[string]any{
				"ids":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}, "maxItems": 200, "description": "Ids de listar_productos."},
				"categoria":  map[string]any{"type": "string", "description": "O todos los de esta categoría."},
				"porcentaje": map[string]any{"type": "number", "minimum": -50, "maximum": 100, "description": "Ej.: 10 sube un 10 %, -5 baja un 5 %."},
			}, "porcentaje")},
		{Name: "crear_cupon", Description: anthropic.String("PROPONE un cupón de descuento para toda la tienda. No lo crea: la persona confirma."),
			InputSchema: esquema(map[string]any{
				"codigo":      map[string]any{"type": "string", "description": "Mayúsculas, números o guiones, de 3 a 20. Ej.: VERANO15."},
				"tipo":        map[string]any{"type": "string", "enum": []string{"porcentaje", "monto_fijo", "envio_gratis"}},
				"valor":       map[string]any{"type": "integer", "minimum": 0, "description": "Porcentaje (1 a 90) o pesos (desde 100). 0 en envío gratis."},
				"minimo":      map[string]any{"type": "integer", "minimum": 0, "description": "Compra mínima en pesos."},
				"limite_usos": map[string]any{"type": "integer", "minimum": 0, "description": "Usos en total (0: sin límite)."},
				"vigencia":    map[string]any{"type": "string", "enum": []string{"7", "30", "finde", "sin"}},
			}, "codigo", "tipo")},
		{Name: "crear_link_de_pago", Description: anthropic.String("PROPONE un link de pago por un monto en pesos. No lo crea: la persona confirma."),
			InputSchema: esquema(map[string]any{
				"monto":    map[string]any{"type": "integer", "minimum": 1},
				"concepto": map[string]any{"type": "string", "description": "Qué se cobra (lo ve quien paga)."},
				"cuotas":   map[string]any{"type": "boolean"},
				"pedido":   map[string]any{"type": "integer", "minimum": 1, "description": "Número de pedido que cobra."},
			}, "monto", "concepto")},
	}
}

// herramientas de una pregunta: la tienda, el rol de quien pregunta y las propuestas que salieron.
type herramientas struct {
	s          *Servicio
	tienda     string
	rol        string
	ahora      time.Time
	propuestas []Propuesta
}

// errHerramienta: el texto va al modelo como error de la herramienta.
type errHerramienta string

func (e errHerramienta) Error() string { return string(e) }

// Permiso de cada herramienta (el mismo que el de la ruta del panel que usa).
var permisos = map[string]roles.Permiso{
	"resumen_del_dia": roles.Panel, "informe_de_ventas": roles.Panel,
	"listar_pedidos": roles.Pedidos, "ver_pedido": roles.Pedidos, "listar_productos": roles.Pedidos,
	"listar_clientes": roles.Clientes, "cambiar_estado_pedido": roles.Pedidos, "ajustar_producto": roles.Productos,
	"ajustar_precios": roles.Catalogo, "crear_cupon": roles.Catalogo, "crear_link_de_pago": roles.Pedidos,
}

func (h *herramientas) ejecutar(ctx context.Context, nombre string, entrada json.RawMessage) (string, bool) {
	p, ok := permisos[nombre]
	if !ok {
		return "No existe esa herramienta.", true
	}
	if !roles.Puede(h.rol, p) {
		return "El rol de quien pregunta no tiene permiso para esto: " + roles.Mensaje(p), true
	}
	var a map[string]any
	if len(entrada) > 0 {
		if err := json.Unmarshal(entrada, &a); err != nil {
			return "Los argumentos no son válidos.", true
		}
	}
	v, err := h.correr(ctx, nombre, a)
	if err != nil {
		var e errHerramienta
		if errors.As(err, &e) {
			return string(e), true
		}
		h.s.log().Error("tincho: herramienta", "nombre", nombre, "error", err)
		return "No pude leer eso ahora.", true
	}
	b, _ := json.Marshal(sinControl(v))
	return string(b), false
}

// sinControl saca los saltos de línea y los caracteres de control de todo texto que vuelve al modelo:
// los nombres de clientes y productos los escriben terceros (el checkout, una importación) y no
// pueden armar renglones que parezcan instrucciones. El resto de la defensa está en Sistema y en que
// toda acción pasa por «Confirmar».
func sinControl(v any) any {
	switch x := v.(type) {
	case string:
		return limpiarTexto(x)
	case map[string]any:
		for k, y := range x {
			x[k] = sinControl(y)
		}
		return x
	case []map[string]any:
		for i := range x {
			x[i] = sinControl(x[i]).(map[string]any)
		}
		return x
	case []any:
		for i := range x {
			x[i] = sinControl(x[i])
		}
		return x
	case []string:
		for i := range x {
			x[i] = limpiarTexto(x[i])
		}
		return x
	}
	return v
}

func limpiarTexto(s string) string {
	return strings.Join(strings.FieldsFunc(s, func(r rune) bool { return r < 0x20 || r == 0x7f || r == 0x2028 || r == 0x2029 }), " ")
}

func (h *herramientas) montos() bool { return roles.Puede(h.rol, roles.Finanzas) }

func txt(a map[string]any, k string, max int) string {
	s, _ := a[k].(string)
	s = strings.TrimSpace(s)
	if len([]rune(s)) > max {
		s = string([]rune(s)[:max])
	}
	return s
}

func entero(a map[string]any, k string) (int64, bool) {
	f, ok := a[k].(float64)
	if !ok || f != float64(int64(f)) {
		return 0, false
	}
	return int64(f), true
}

func filas(a map[string]any) int {
	n, ok := entero(a, "limite")
	if !ok || n < 1 {
		return 10
	}
	return int(min(n, maxFilas))
}

func (h *herramientas) correr(ctx context.Context, nombre string, a map[string]any) (any, error) {
	pool := h.s.Pool
	switch nombre {
	case "resumen_del_dia":
		t, err := tiendas.ObtenerTablero(ctx, pool, h.tienda, "hoy", h.ahora)
		if err != nil {
			return nil, err
		}
		if !h.montos() {
			t.SinMontos()
		}
		r := map[string]any{"pedidos": t.Kpis.Pedidos, "paraHacer": t.ParaHacer, "ultimosPedidos": primeros(t.UltimosPedidos, 5)}
		if t.Ventas != nil {
			r["ventas"], r["ticket"], r["meta"] = t.Ventas, t.Kpis.Ticket, t.Meta
		}
		return r, nil

	case "informe_de_ventas":
		periodo := txt(a, "periodo", 4)
		if periodo != "hoy" && periodo != "7d" && periodo != "30d" {
			periodo = "7d"
		}
		t, err := tiendas.ObtenerTablero(ctx, pool, h.tienda, periodo, h.ahora)
		if err != nil {
			return nil, err
		}
		if !h.montos() {
			t.SinMontos()
		}
		r := map[string]any{"periodo": periodo, "pedidos": t.Kpis.Pedidos, "clientesNuevos": t.Kpis.ClientesNuevos,
			"masVendidos": primeros(t.MasVendidos, 10)}
		if t.Ventas != nil {
			r["ventas"], r["ticket"], r["canales"] = t.Ventas, t.Kpis.Ticket, t.Canales
		}
		return r, nil

	case "listar_pedidos":
		f := pedidos.Filtros{Q: txt(a, "texto", 100), Fecha: txt(a, "fecha", 4), Cuantos: filas(a)}
		if f.Fecha == "todo" {
			f.Fecha = ""
		}
		if e := txt(a, "estado", 30); e != "" {
			f.Estado = []string{e}
		}
		if c := txt(a, "canal", 4); c != "" {
			f.Canal = []string{c}
		}
		l, err := pedidos.Listar(ctx, pool, h.tienda, f, h.ahora)
		if err != nil {
			return nil, err
		}
		out := make([]map[string]any, 0, len(l.Filas))
		for _, p := range l.Filas {
			prods := []string{}
			for _, x := range primeros(p.Productos, 3) {
				prods = append(prods, x.Nombre)
			}
			m := map[string]any{"numero": p.Numero, "cliente": p.Cliente, "estado": p.Estado, "canal": p.Canal,
				"envio": p.Envio, "fecha": p.Creado.In(horaArgentina).Format("2006-01-02 15:04"), "productos": prods,
				"cantidadProductos": p.CantidadProductos}
			if h.montos() {
				m["total"] = p.Total
			}
			out = append(out, m)
		}
		return map[string]any{"total": l.Total, "pedidos": out}, nil

	case "ver_pedido":
		n, ok := entero(a, "numero")
		if !ok || n < 1 {
			return nil, errHerramienta("Falta el número de pedido.")
		}
		d, err := pedidos.Obtener(ctx, pool, h.tienda, int(n))
		if errors.Is(err, pedidos.ErrNoExiste) {
			return nil, errHerramienta(fmt.Sprintf("No existe el pedido #%d.", n))
		}
		if err != nil {
			return nil, err
		}
		items := []map[string]any{}
		for _, i := range d.Items {
			m := map[string]any{"nombre": i.Nombre, "cantidad": i.Cantidad}
			if h.montos() {
				m["precio"] = i.Precio
			}
			items = append(items, m)
		}
		// Del comprador, sólo lo que hace falta para hablar del pedido: nada de contacto.
		r := map[string]any{"numero": d.Numero, "cliente": d.Cliente, "ciudad": d.Ciudad, "estado": d.Estado,
			"estadoPago": d.EstadoPago, "canal": d.Canal, "pago": d.Pago, "envio": d.Envio,
			"fecha": d.Creado.In(horaArgentina).Format("2006-01-02 15:04"), "items": items,
			"pedidosDelCliente": d.Comprador.Pedidos, "seguimiento": d.Seguimiento,
			"siguientes": siguientes(d.Estado, d.Envio)}
		if h.montos() {
			r["total"], r["costoEnvio"] = d.Total, d.CostoEnvio
		}
		return r, nil

	case "listar_productos":
		regs, err := listas.Productos(ctx, pool, h.tienda, h.ahora)
		if err != nil {
			return nil, err
		}
		f := listas.Filtros{Q: txt(a, "texto", 100), Cuantos: filas(a),
			Orden: map[string]string{"mas_vendidos": "sold", "menos_stock": "stock", "mayor_precio": "price"}[txt(a, "orden", 20)]}
		if s := map[string]string{"activo": "active", "sin_stock": "out", "pausado": "paused", "borrador": "draft"}[txt(a, "estado", 20)]; s != "" {
			f.Seg = []string{s}
		}
		l := listas.Aplicar(regs, listas.DefProductos, f)
		out := make([]map[string]any, 0, len(l.Filas))
		for _, r := range l.Filas {
			m := map[string]any{"id": r["id"], "nombre": r["nombre"], "categoria": r["categoria"], "precio": r["precio"],
				"stock": r["stock"], "vendidos30d": r["vendidos30"], "estado": r["estado"]}
			if v, _ := r["variantes"].([]listas.Registro); len(v) > 0 {
				m["variantes"] = len(v)
			}
			if roles.Puede(h.rol, roles.Catalogo) {
				m["margenPct"] = r["margen"]
			}
			out = append(out, m)
		}
		return map[string]any{"total": l.Total, "productos": out}, nil

	case "listar_clientes":
		regs, err := listas.Clientes(ctx, pool, h.tienda, h.ahora)
		if err != nil {
			return nil, err
		}
		f := listas.Filtros{Q: txt(a, "texto", 100), Cuantos: filas(a),
			Orden: map[string]string{"mas_gastaron": "ltv", "compra_reciente": "last", "mas_pedidos": "orders"}[txt(a, "orden", 20)]}
		if s := map[string]string{"vip": "vip", "frecuentes": "freq", "nuevos": "new", "en_riesgo": "risk", "dormidos": "sleep"}[txt(a, "segmento", 20)]; s != "" {
			f.Seg = []string{s}
		}
		l := listas.Aplicar(regs, listas.DefClientes, f)
		out := make([]map[string]any, 0, len(l.Filas))
		for _, r := range l.Filas {
			// Sin email ni teléfono: al modelo va lo justo para hablar de los clientes.
			m := map[string]any{"nombre": r["nombre"], "ciudad": r["ciudad"], "canal": r["canal"], "pedidos": r["pedidos"],
				"diasDesdeUltimaCompra": r["ultima"], "segmento": r["seg"], "tieneWhatsApp": r["wa"]}
			if h.montos() {
				m["gastado"] = r["gastado"]
			}
			out = append(out, m)
		}
		return map[string]any{"total": l.Total, "clientes": out}, nil

	case "cambiar_estado_pedido":
		return h.proponerEstado(ctx, a)
	case "ajustar_producto":
		return h.proponerProducto(ctx, a)
	case "ajustar_precios":
		return h.proponerPrecios(ctx, a)
	case "crear_cupon":
		return h.proponerCupon(a)
	case "crear_link_de_pago":
		return h.proponerLink(ctx, a)
	}
	return nil, errHerramienta("No existe esa herramienta.")
}

// siguientes: a qué estados puede pasar el pedido.
func siguientes(estado, envio string) []string {
	out := []string{}
	for _, e := range pedidos.Estados {
		if pedidos.TransicionValida(estado, e, envio) {
			out = append(out, e)
		}
	}
	return out
}

// proponer guarda la propuesta y le dice al modelo que todavía no se hizo.
func (h *herramientas) proponer(p Propuesta) (any, error) {
	if len(h.propuestas) >= 3 {
		return nil, errHerramienta("Ya hay 3 propuestas en esta respuesta: que la persona confirme esas primero.")
	}
	p.ID = idPropuesta()
	p.Titulo, p.Detalle = limpiarTexto(p.Titulo), limpiarTexto(p.Detalle)
	h.propuestas = append(h.propuestas, p)
	return map[string]any{"propuesta": "lista", "titulo": p.Titulo,
		"importante": "Todavía NO se hizo: la persona ve un botón «Confirmar». No digas que quedó hecho."}, nil
}

func (h *herramientas) proponerEstado(ctx context.Context, a map[string]any) (any, error) {
	n, ok := entero(a, "numero")
	estado := txt(a, "estado", 30)
	if !ok || n < 1 || estado == "" {
		return nil, errHerramienta("Falta el número de pedido o el estado nuevo.")
	}
	d, err := pedidos.Obtener(ctx, h.s.Pool, h.tienda, int(n))
	if errors.Is(err, pedidos.ErrNoExiste) {
		return nil, errHerramienta(fmt.Sprintf("No existe el pedido #%d.", n))
	}
	if err != nil {
		return nil, err
	}
	if !pedidos.TransicionValida(d.Estado, estado, d.Envio) {
		return nil, errHerramienta(fmt.Sprintf("El pedido #%d está «%s»: puede pasar a %s.", n, d.Estado, strings.Join(siguientes(d.Estado, d.Envio), ", ")))
	}
	return h.proponer(Propuesta{Herramienta: "cambiar_estado_pedido", Argumentos: map[string]any{"numero": n, "estado": estado},
		Titulo: fmt.Sprintf("Pasar el pedido #%d a «%s»", n, estado), Detalle: fmt.Sprintf("%s · ahora está «%s»", d.Cliente, d.Estado)})
}

func (h *herramientas) proponerProducto(ctx context.Context, a map[string]any) (any, error) {
	id := txt(a, "id", 80)
	if id == "" {
		return nil, errHerramienta("Falta el id del producto (buscalo con listar_productos).")
	}
	precio, conPrecio := entero(a, "precio")
	stock, conStock := entero(a, "stock")
	sumar, conSumar := entero(a, "sumar_stock")
	if !conPrecio && !conStock && !conSumar {
		return nil, errHerramienta("Decime el precio nuevo, el stock nuevo o cuánto sumar al stock.")
	}
	if conStock && conSumar {
		return nil, errHerramienta("Pasá \"stock\" o \"sumar_stock\", no los dos.")
	}
	if conPrecio && !roles.Puede(h.rol, roles.Catalogo) {
		return nil, errHerramienta("El rol de quien pregunta no puede cambiar precios: " + roles.Mensaje(roles.Catalogo))
	}
	if conPrecio && (precio < 0 || precio > 100_000_000) || conStock && (stock < 0 || stock > 1_000_000) {
		return nil, errHerramienta("El precio o el stock están fuera de rango.")
	}
	p, err := listas.Obtener(ctx, h.s.Pool, h.tienda, id, h.ahora)
	if errors.Is(err, listas.ErrNoExiste) {
		return nil, errHerramienta("No existe ese producto. Buscalo con listar_productos.")
	}
	if err != nil {
		return nil, err
	}
	if (conStock || conSumar) && len(p.Opciones) > 0 {
		return nil, errHerramienta(fmt.Sprintf("«%s» tiene variantes: su stock se ajusta por variante en Inventario.", p.Nombre))
	}
	args := map[string]any{"id": id}
	cambios := []string{}
	if conPrecio {
		args["precio"] = precio
		cambios = append(cambios, fmt.Sprintf("precio de $ %s a $ %s", miles(int(p.Precio)), miles(int(precio))))
	}
	if conStock || conSumar {
		nuevo := stock
		if conSumar {
			nuevo = int64(p.Stock) + sumar
			args["sumar_stock"] = sumar
		} else {
			args["stock"] = stock
		}
		if nuevo < 0 {
			return nil, errHerramienta(fmt.Sprintf("El stock no puede quedar negativo (hay %d).", p.Stock))
		}
		cambios = append(cambios, fmt.Sprintf("stock de %d a %d", p.Stock, nuevo))
	}
	return h.proponer(Propuesta{Herramienta: "ajustar_producto", Argumentos: args,
		Titulo: "Ajustar «" + p.Nombre + "»", Detalle: strings.Join(cambios, " · ")})
}

// proponerPrecios: el aumento (o la baja) en lote, como «Cambiar precio» del panel: una sola
// propuesta con el resumen («12 productos, +10 %: Coca $ 3.500 → $ 3.850…»).
func (h *herramientas) proponerPrecios(ctx context.Context, a map[string]any) (any, error) {
	pct, ok := a["porcentaje"].(float64)
	if !ok || pct == 0 || pct < -50 || pct > 100 {
		return nil, errHerramienta("Decime el porcentaje (por ejemplo 10 para subir un 10 %, -5 para bajar un 5 %).")
	}
	pct = math.Round(pct*10) / 10
	pedidos := map[string]bool{}
	if xs, ok := a["ids"].([]any); ok {
		for _, x := range xs {
			if id, ok := x.(string); ok && id != "" && len(pedidos) < 200 {
				pedidos[id] = true
			}
		}
	}
	cat := strings.ToLower(txt(a, "categoria", 80))
	if len(pedidos) == 0 && cat == "" {
		return nil, errHerramienta("Decime qué productos (los ids de listar_productos) o una categoría.")
	}
	regs, err := listas.Productos(ctx, h.s.Pool, h.tienda, h.ahora)
	if err != nil {
		return nil, err
	}
	type prod struct {
		id, nombre string
		precio     int64
	}
	var elegidos []prod
	for _, r := range regs {
		id, _ := r["id"].(string)
		c, _ := r["categoria"].(string)
		if pedidos[id] || (cat != "" && strings.ToLower(strings.TrimSpace(c)) == cat) {
			n, _ := r["nombre"].(string)
			p, _ := r["precio"].(int64)
			if f, ok := r["precio"].(float64); ok {
				p = int64(f)
			}
			elegidos = append(elegidos, prod{id, n, p})
		}
	}
	if len(elegidos) == 0 {
		return nil, errHerramienta("No encontré esos productos. Buscalos con listar_productos.")
	}
	ids := make([]string, len(elegidos))
	muestra := []string{}
	for i, p := range elegidos {
		ids[i] = p.id
		if i < 3 {
			nuevo := int64(math.Round(float64(p.precio)*(1+pct/100)/10) * 10)
			muestra = append(muestra, fmt.Sprintf("%s $ %s → $ %s", p.nombre, miles(int(p.precio)), miles(int(max(0, nuevo)))))
		}
	}
	signo, verbo := "+", "Subir"
	if pct < 0 {
		signo, verbo = "−", "Bajar"
	}
	cuantos := "1 producto"
	if len(elegidos) > 1 {
		cuantos = fmt.Sprintf("%d productos", len(elegidos))
	}
	detalle := fmt.Sprintf("%s, %s%s %%: %s", cuantos, signo, strconv.FormatFloat(math.Abs(pct), 'f', -1, 64), strings.Join(muestra, " · "))
	if len(elegidos) > 3 {
		detalle += "…"
	}
	return h.proponer(Propuesta{Herramienta: "ajustar_precios", Argumentos: map[string]any{"ids": ids, "porcentaje": pct},
		Titulo: fmt.Sprintf("%s %s %% el precio de %s", verbo, strconv.FormatFloat(math.Abs(pct), 'f', -1, 64), cuantos), Detalle: detalle})
}

var codigoCupon = regexp.MustCompile(`^[A-Z0-9-]{3,20}$`)

func (h *herramientas) proponerCupon(a map[string]any) (any, error) {
	codigo := strings.ToUpper(txt(a, "codigo", 30))
	if !codigoCupon.MatchString(codigo) {
		return nil, errHerramienta("El código va en mayúsculas, de 3 a 20 letras, números o guiones.")
	}
	tipo := txt(a, "tipo", 20)
	valor, _ := entero(a, "valor")
	var desc string
	switch tipo {
	case "porcentaje":
		if valor < 1 || valor > 90 {
			return nil, errHerramienta("El porcentaje va de 1 a 90.")
		}
		desc = fmt.Sprintf("%d %% de descuento", valor)
	case "monto_fijo":
		if valor < 100 {
			return nil, errHerramienta("El descuento fijo es de $ 100 para arriba.")
		}
		desc = fmt.Sprintf("$ %s de descuento", miles(int(valor)))
	case "envio_gratis":
		valor, desc = 0, "envío gratis"
	default:
		return nil, errHerramienta("Falta el tipo de descuento: porcentaje, monto_fijo o envio_gratis.")
	}
	args := map[string]any{"codigo": codigo, "tipo": tipo, "valor": valor}
	if m, ok := entero(a, "minimo"); ok && m > 0 {
		args["minimo"] = m
		desc += fmt.Sprintf(" desde $ %s", miles(int(m)))
	}
	if l, ok := entero(a, "limite_usos"); ok && l > 0 {
		args["limite_usos"] = l
		desc += fmt.Sprintf(" · %d usos", l)
	}
	vig := txt(a, "vigencia", 5)
	if vig != "7" && vig != "30" && vig != "finde" && vig != "sin" {
		vig = "7"
	}
	args["vigencia"] = vig
	desc += " · " + map[string]string{"7": "7 días", "30": "30 días", "finde": "este fin de semana", "sin": "sin vencimiento"}[vig]
	return h.proponer(Propuesta{Herramienta: "crear_cupon", Argumentos: args, Titulo: "Crear el cupón " + codigo, Detalle: desc})
}

func (h *herramientas) proponerLink(ctx context.Context, a map[string]any) (any, error) {
	monto, ok := entero(a, "monto")
	if !ok || monto < 1 || monto > 100_000_000 {
		return nil, errHerramienta("Falta el monto en pesos.")
	}
	concepto := txt(a, "concepto", 140)
	if len([]rune(concepto)) < 2 {
		return nil, errHerramienta("Escribí qué se cobra.")
	}
	args := map[string]any{"monto": monto, "concepto": concepto, "cuotas": a["cuotas"] != false}
	if n, ok := entero(a, "pedido"); ok && n > 0 {
		if _, err := pedidos.Obtener(ctx, h.s.Pool, h.tienda, int(n)); errors.Is(err, pedidos.ErrNoExiste) {
			return nil, errHerramienta(fmt.Sprintf("No existe el pedido #%d.", n))
		} else if err != nil {
			return nil, err
		}
		args["pedido"] = n
	}
	return h.proponer(Propuesta{Herramienta: "crear_link_de_pago", Argumentos: args,
		Titulo: fmt.Sprintf("Link de pago por $ %s", miles(int(monto))), Detalle: concepto})
}
