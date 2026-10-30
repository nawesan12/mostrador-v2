package importacion_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/importacion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

var ctx = context.Background()

// falsa es una fuente en memoria: cada etapa tiene sus páginas y el cursor es el número de página.
type falsa struct {
	paginas map[string][]importacion.Pagina
	// Errores a devolver antes de la página (uno por llamada).
	errores []error
	pedidas int
	cred    string
}

func (f *falsa) Pagina(_ context.Context, etapa string, cursor json.RawMessage) (importacion.Pagina, error) {
	f.pedidas++
	if len(f.errores) > 0 {
		e := f.errores[0]
		f.errores = f.errores[1:]
		return importacion.Pagina{}, e
	}
	n := 0
	if len(cursor) > 0 {
		_ = json.Unmarshal(cursor, &n)
	}
	ps := f.paginas[etapa]
	if n >= len(ps) {
		return importacion.Pagina{}, nil
	}
	p := ps[n]
	if n+1 < len(ps) {
		p.Siguiente = json.RawMessage(fmt.Sprint(n + 1))
	}
	return p, nil
}

func servicio(t *testing.T, b *pruebadb.Bases, f *falsa) *importacion.Servicio {
	t.Helper()
	c, err := cifrado.Nuevo(base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32))))
	if err != nil {
		t.Fatal(err)
	}
	return &importacion.Servicio{Pool: b.App, Cifra: c, Fuentes: map[string]importacion.Abrir{
		"tiendanube": func(_ context.Context, o importacion.Origen) (importacion.Fuente, error) {
			f.cred = string(o.Credencial)
			return f, nil
		},
	}}
}

// correr hace lo que el worker: procesa el outbox hasta que no queda nada.
func correr(t *testing.T, b *pruebadb.Bases, s *importacion.Servicio) int {
	t.Helper()
	w := &outbox.Worker{Pool: b.Worker, Manejadores: s.Manejadores(), Lote: 1}
	pasos := 0
	for range 100 {
		n, err := w.Pasada(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if n == 0 {
			return pasos
		}
		pasos += n
	}
	t.Fatal("la importación no terminaba")
	return 0
}

func ptr[T any](v T) *T { return &v }

func fuenteCompleta() *falsa {
	return &falsa{paginas: map[string][]importacion.Pagina{
		"productos": {
			{Productos: []importacion.ProductoViejo{
				{Externo: "p1", URL: "https://luna.mitiendanube.com/productos/mate-imperial/", Producto: listas.Editable{
					Nombre: "Mate imperial", Precio: 32000, SKU: ptr("MI-1"), Stock: ptr(4),
					Fotos: []string{"https://cdn.ejemplo.com/1.jpg", "https://cdn.ejemplo.com/2.jpg"}}},
				// No pasa las validaciones del panel: queda en errores y el resto entra.
				{Externo: "p2", URL: "/productos/x", Producto: listas.Editable{Nombre: "x", Precio: 100}},
			}},
			{Productos: []importacion.ProductoViejo{
				{Externo: "p3", URL: "/productos/remera", Producto: listas.Editable{
					Nombre: "Remera " + strings.Repeat("larga ", 30), Precio: 15000,
					Opciones:  []listas.Opcion{{Nombre: "Talle", Valores: []string{"S", "M"}}},
					Variantes: []listas.VarianteEditable{{Valores: map[string]string{"Talle": "M"}, Stock: ptr(3)}}}},
				// El mismo SKU que el primero: repetido.
				{Externo: "p4", URL: "/productos/mate-imperial-2", Producto: listas.Editable{Nombre: "Mate imperial", Precio: 1, SKU: ptr("MI-1")}},
			}},
		},
		"clientes": {{Clientes: []importacion.ClienteViejo{
			{Externo: "c1", Nombre: "Lucía Fernández", Email: "Lucia@Example.com", Telefono: "341 555-0101", Ciudad: "Rosario"},
			{Externo: "c2", Nombre: "Sin Compras", Email: "nadie@example.com"},
			{Externo: "c3", Nombre: "Don Sin Email", Telefono: "11 5555-0000"},
			{Externo: "c4", Email: "malo"},
		}}},
		"pedidos": {{Pedidos: []importacion.PedidoViejo{
			{Externo: "o1", Numero: "#101", Fecha: time.Now().AddDate(0, -2, 0), Email: "LUCIA@example.com", Nombre: "Lucía F.",
				Total: 50000, Estado: "paid", Items: []importacion.ItemViejo{{Nombre: "Mate imperial", Cantidad: 1, Precio: 50000}}},
			{Externo: "o2", Numero: "#102", Fecha: time.Now().AddDate(0, -1, 0), Email: "otra@example.com", Nombre: "Otra",
				Total: 20000, Estado: "shipped"},
			{Externo: "o1", Numero: "#101", Fecha: time.Now(), Email: "x@example.com", Total: 1},
		}}},
		"redirecciones": {{Redirecciones: []importacion.URLVieja{
			{Desde: "/categorias/mates/", Hacia: "/catalogo?categoria=Mates"},
			{Desde: "/contacto", Hacia: "//otro.com"},
		}}},
	}}
}

func TestImportacionCompletaPorElOutbox(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	f := fuenteCompleta()
	s := servicio(t, b, f)

	if _, err := s.Crear(ctx, luna, "ventas", importacion.Nueva{Plataforma: "tiendanube"}); !errors.Is(err, importacion.ErrSinPermiso) {
		t.Fatalf("ventas pudo importar: %v", err)
	}
	if _, err := s.Crear(ctx, luna, "duenio", importacion.Nueva{Plataforma: "shopify"}); !errors.Is(err, importacion.ErrSinFuente) {
		t.Fatalf("sin fuente: %v", err)
	}
	if _, err := s.Crear(ctx, luna, "duenio", importacion.Nueva{Plataforma: "tiendanube", URL: "javascript:alert(1)"}); !errors.Is(err, importacion.ErrInvalido) {
		t.Fatalf("url: %v", err)
	}
	// Un cliente que ya compró en Tu Local Web con el mismo email (en otra caja).
	email := "lucia@example.com"
	if _, err := pedidos.Crear(ctx, b.App, luna, pedidos.Nuevo{Comprador: pedidos.Comprador{Nombre: "Lucía Fernández", Email: &email},
		Canal: "pos", Pago: "efectivo", Envio: "retiro", Ciudad: "Rosario", Estado: "Entregado",
		Items: []pedidos.Item{{Nombre: "Yerba", Cantidad: 1, Precio: 10000}}}); err != nil {
		t.Fatal(err)
	}

	imp, err := s.Crear(ctx, luna, "duenio", importacion.Nueva{Plataforma: "tiendanube", URL: "https://luna.mitiendanube.com", Credencial: "tn-secreto"})
	if err != nil || imp.Estado != "pendiente" || imp.Etapa != "productos" {
		t.Fatalf("crear: %+v %v", imp, err)
	}
	if _, err := s.Crear(ctx, luna, "admin", importacion.Nueva{Plataforma: "tiendanube"}); !errors.Is(err, importacion.ErrActiva) {
		t.Fatalf("dos activas: %v", err)
	}
	var cifrada []byte
	_ = b.App.EnTienda(ctx, luna, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT credencial FROM importaciones WHERE id = $1`, imp.ID).Scan(&cifrada)
	})
	if len(cifrada) == 0 || strings.Contains(string(cifrada), "tn-secreto") {
		t.Fatal("la credencial no quedó cifrada")
	}

	// productos (2 páginas) + clientes + pedidos + redirecciones = 5 pasos encadenados.
	if n := correr(t, b, s); n != 5 {
		t.Fatalf("pasos: %d", n)
	}
	if f.cred != "tn-secreto" {
		t.Fatalf("la fuente no recibió la credencial: %q", f.cred)
	}
	a, err := s.Actual(ctx, luna)
	if err != nil || a == nil {
		t.Fatal(err)
	}
	want := importacion.Progreso{Productos: 2, Clientes: 3, Pedidos: 2, Redirecciones: 3, Repetidos: 2, Recortados: 1}
	if a.Estado != "lista" || a.Terminada == nil || a.Progreso != want {
		t.Fatalf("estado: %+v", a)
	}
	// p2 (nombre corto), c4 (sin nombre ni email válido) y la redirección a otro dominio.
	if len(a.Errores) != 3 || a.Errores[0].Externo != "p2" || a.Errores[0].Etapa != "productos" {
		t.Fatalf("errores: %+v", a.Errores)
	}
	_ = b.App.EnTienda(ctx, luna, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT credencial FROM importaciones WHERE id = $1`, imp.ID).Scan(&cifrada)
	})
	if cifrada != nil {
		t.Fatal("la credencial sigue guardada después de terminar")
	}

	// Los productos entraron con la función del panel (variantes y galería) y su URL vieja redirige.
	prods, _ := listas.Productos(ctx, b.App, luna, time.Now())
	if len(prods) != 2 {
		t.Fatalf("productos: %d", len(prods))
	}
	var mate, remera listas.Producto
	for _, r := range prods {
		p := r["_ficha"].(listas.Producto)
		if p.SKU == "MI-1" {
			mate = p
		} else {
			remera = p
		}
	}
	if len(mate.Fotos) != 2 || mate.Stock != 4 || len(remera.Variantes) != 2 || remera.Stock != 3 || len([]rune(remera.Nombre)) != 120 {
		t.Fatalf("productos: %+v / %+v", mate, remera)
	}
	if h, err := vidriera.Redireccion(ctx, b.App, "luna", "/productos/Mate-Imperial?ref=ig"); err != nil || h != "/producto/mate-imperial" {
		t.Fatalf("redirección del producto: %q %v", h, err)
	}
	if h, _ := vidriera.Redireccion(ctx, b.App, "luna", "/categorias/mates"); h != "/catalogo?categoria=Mates" {
		t.Fatalf("redirección de categoría: %q", h)
	}
	if _, err := vidriera.Redireccion(ctx, b.App, "luna", "/contacto"); !errors.Is(err, vidriera.ErrNoExiste) {
		t.Fatalf("redirigió a otro dominio: %v", err)
	}

	// Clientes: el del pedido y su ficha importada son uno (email en minúsculas), con el historial
	// sumado; la ficha sin compras y la sin email aparecen igual; el del historial sin ficha también.
	regs, err := listas.Clientes(ctx, b.App, luna, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	por := map[string]listas.Registro{}
	for _, r := range regs {
		por[r["id"].(string)] = r
	}
	if len(regs) != 4 {
		t.Fatalf("clientes: %d %+v", len(regs), regs)
	}
	lu := por["lucia@example.com"]
	if lu["pedidos"] != 2 || lu["gastado"] != int64(60000) || lu["telefono"] == nil || lu["nombre"] != "Lucía Fernández" {
		t.Fatalf("Lucía: %+v", lu)
	}
	if por["nadie@example.com"]["pedidos"] != 0 || por["otra@example.com"]["gastado"] != int64(20000) {
		t.Fatalf("fichas e historial: %+v", por)
	}
	h, err := listas.Historial(ctx, b.App, luna, "lucia@example.com")
	if err != nil || len(h) != 1 || h[0].Plataforma != "tiendanube" || h[0].Numero != "#101" || h[0].Total != 50000 {
		t.Fatalf("historial: %+v %v", h, err)
	}
	// El historial no es un pedido de Tu Local Web: no aparece en Pedidos ni toca el stock.
	l, err := pedidos.Listar(ctx, b.App, luna, pedidos.Filtros{}, time.Now())
	if err != nil || l.Total != 1 {
		t.Fatalf("pedidos: %+v %v", l.Total, err)
	}

	// Otra tienda no ve nada de esto.
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	if a, _ := s.Actual(ctx, sol); a != nil {
		t.Fatal("la otra tienda ve la importación")
	}
	if regs, _ := listas.Clientes(ctx, b.App, sol, time.Now()); len(regs) != 0 {
		t.Fatalf("la otra tienda ve clientes: %d", len(regs))
	}
	if _, err := vidriera.Redireccion(ctx, b.App, "sol", "/categorias/mates"); !errors.Is(err, vidriera.ErrNoExiste) {
		t.Fatal("la otra tienda redirige")
	}

	// Importar de nuevo lo mismo no duplica: todo repetido (menos lo que no tiene cómo reconocerse).
	if _, err := s.Crear(ctx, luna, "duenio", importacion.Nueva{Plataforma: "tiendanube"}); err != nil {
		t.Fatal(err)
	}
	correr(t, b, s)
	a, _ = s.Actual(ctx, luna)
	if a.Progreso.Productos != 0 || a.Progreso.Pedidos != 0 || a.Progreso.Clientes != 0 {
		t.Fatalf("la segunda vez duplicó: %+v", a.Progreso)
	}
}

func TestUnPasoRepetidoOCanceladoNoHaceNada(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	f := fuenteCompleta()
	s := servicio(t, b, f)
	imp, err := s.Crear(ctx, luna, "duenio", importacion.Nueva{Plataforma: "tiendanube"})
	if err != nil {
		t.Fatal(err)
	}
	ev := func(paso int) outbox.Evento {
		d, _ := json.Marshal(map[string]any{"id": imp.ID, "paso": paso})
		return outbox.Evento{TiendaID: luna, Tipo: importacion.TipoPaso, Datos: d}
	}
	if err := s.Paso(ctx, ev(0)); err != nil {
		t.Fatal(err)
	}
	// El outbox entrega al menos una vez: el mismo paso otra vez no guarda la página de nuevo.
	if err := s.Paso(ctx, ev(0)); err != nil {
		t.Fatal(err)
	}
	a, _ := s.Actual(ctx, luna)
	if a.Progreso.Productos != 1 || a.Estado != "corriendo" || a.Etapa != "productos" {
		t.Fatalf("después del paso repetido: %+v", a)
	}
	if _, err := s.Cancelar(ctx, luna, "ventas", imp.ID); !errors.Is(err, importacion.ErrSinPermiso) {
		t.Fatalf("ventas canceló: %v", err)
	}
	c, err := s.Cancelar(ctx, luna, "admin", imp.ID)
	if err != nil || c.Estado != "cancelada" || c.Terminada == nil {
		t.Fatalf("cancelar: %+v %v", c, err)
	}
	if _, err := s.Cancelar(ctx, luna, "admin", imp.ID); !errors.Is(err, importacion.ErrTerminada) {
		t.Fatalf("cancelar dos veces: %v", err)
	}
	if _, err := s.Cancelar(ctx, luna, "admin", "00000000-0000-0000-0000-000000000000"); !errors.Is(err, importacion.ErrNoExiste) {
		t.Fatalf("cancelar una que no existe: %v", err)
	}
	pedidas := f.pedidas
	// El paso que quedó encolado no hace nada con la importación cancelada.
	if err := s.Paso(ctx, ev(1)); err != nil || f.pedidas != pedidas {
		t.Fatalf("siguió después de cancelar: %v (%d pedidos a la fuente)", err, f.pedidas-pedidas)
	}
	// Cancelada, se puede empezar otra.
	if _, err := s.Crear(ctx, luna, "duenio", importacion.Nueva{Plataforma: "tiendanube"}); err != nil {
		t.Fatal(err)
	}
}

func TestErroresDeLaFuente(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	f := fuenteCompleta()
	s := servicio(t, b, f)
	imp, _ := s.Crear(ctx, luna, "duenio", importacion.Nueva{Plataforma: "tiendanube"})
	d, _ := json.Marshal(map[string]any{"id": imp.ID, "paso": 0})
	ev := outbox.Evento{TiendaID: luna, Tipo: importacion.TipoPaso, Datos: d}

	// Un corte de red se reintenta (el error vuelve al outbox) y la importación sigue activa.
	f.errores = []error{errors.New("se cortó la conexión")}
	if err := s.Paso(ctx, ev); err == nil {
		t.Fatal("un error transitorio no volvió al outbox")
	}
	if a, _ := s.Actual(ctx, luna); a.Estado != "pendiente" {
		t.Fatalf("estado después del corte: %s", a.Estado)
	}
	// En el último intento, la importación queda fallida con el error.
	f.errores = []error{errors.New("timeout")}
	ev.Intentos = outbox.MaxIntentos - 1
	if err := s.Paso(ctx, ev); err != nil {
		t.Fatal(err)
	}
	a, _ := s.Actual(ctx, luna)
	if a.Estado != "fallida" || len(a.Errores) != 1 || a.Errores[0].Detalle != "timeout" {
		t.Fatalf("último intento: %+v", a)
	}

	// Un error definitivo corta en el primer intento.
	imp, _ = s.Crear(ctx, luna, "duenio", importacion.Nueva{Plataforma: "tiendanube"})
	d, _ = json.Marshal(map[string]any{"id": imp.ID, "paso": 0})
	f.errores = []error{fmt.Errorf("%w: la clave de Tiendanube ya no sirve", importacion.ErrDefinitivo)}
	if err := s.Paso(ctx, outbox.Evento{TiendaID: luna, Datos: d}); err != nil {
		t.Fatal(err)
	}
	a, _ = s.Actual(ctx, luna)
	if a.Estado != "fallida" || a.Errores[0].Detalle != "la clave de Tiendanube ya no sirve" {
		t.Fatalf("definitivo: %+v", a)
	}
}

func TestClienteCargadoEnElPanel(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	alta := json.RawMessage(`{"nombre": "Ana Paz", "telefono": "11 5555-1234", "email": "Ana@Example.com", "ciudad": "Salta"}`)
	if n, err := listas.CrearCliente(ctx, b.App, luna, alta); n != 1 || err != nil {
		t.Fatalf("alta: %d %v", n, err)
	}
	// El mismo email no crea otra ficha.
	if n, _ := listas.CrearCliente(ctx, b.App, luna, alta); n != 0 {
		t.Fatalf("repetida: %d", n)
	}
	if n, _ := listas.CrearCliente(ctx, b.App, luna, json.RawMessage(`{"nombre": "A"}`)); n != 0 {
		t.Fatal("aceptó un nombre de una letra")
	}
	regs, _ := listas.Clientes(ctx, b.App, luna, time.Now())
	if len(regs) != 1 || regs[0]["id"] != "ana@example.com" || regs[0]["seg"] != "new" || regs[0]["canal"] != "pos" || regs[0]["wa"] != true {
		t.Fatalf("lista: %+v", regs)
	}
}
