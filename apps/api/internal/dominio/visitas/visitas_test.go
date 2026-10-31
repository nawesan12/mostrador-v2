package visitas_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/visitas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// hash de prueba: 16 bytes en hexadecimal.
func h(c string) string { return strings.Repeat(c, 32) }

type fila struct {
	pagina, producto   string
	vistas, visitantes int
}

func filas(t *testing.T, b *pruebadb.Bases, tienda string) map[string]fila {
	t.Helper()
	ctx := context.Background()
	out := map[string]fila{}
	if err := b.App.EnTienda(ctx, tienda, func(tx pgx.Tx) error {
		r, err := tx.Query(ctx, `SELECT pagina, coalesce(p.slug, ''), vistas, visitantes FROM visitas v
			LEFT JOIN productos p ON p.tienda_id = v.tienda_id AND p.id = v.producto_id`)
		if err != nil {
			return err
		}
		defer r.Close()
		for r.Next() {
			var f fila
			if err := r.Scan(&f.pagina, &f.producto, &f.vistas, &f.visitantes); err != nil {
				return err
			}
			out[f.pagina+":"+f.producto] = f
		}
		return r.Err()
	}); err != nil {
		t.Fatal(err)
	}
	return out
}

func producto(t *testing.T, b *pruebadb.Bases, tienda, nombre string) string {
	t.Helper()
	ctx := context.Background()
	var slug string
	if err := b.App.EnTienda(ctx, tienda, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `INSERT INTO productos (tienda_id, nombre, precio) VALUES ($1, $2, 1000) RETURNING slug`,
			tienda, nombre).Scan(&slug)
	}); err != nil {
		t.Fatal(err)
	}
	return slug
}

// El agregado: vistas por página y producto, personas distintas por página y en el total del
// día, y nada de productos o tiendas que no existen.
func TestRegistrarAgregaYCuentaUnicos(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	id, _ := tiendas.Crear(ctx, b.App, "mate", "Mate", "regionales")
	mate := producto(t, b, id, "Mate de calabaza")
	ahora := time.Now()

	n, err := visitas.Registrar(ctx, b.App, []visitas.Vista{
		{Tienda: "mate", Pagina: "inicio", Visitante: h("a")},
		{Tienda: "mate", Pagina: "inicio", Visitante: h("a")},
		{Tienda: "mate", Pagina: "inicio", Visitante: h("b")},
		{Tienda: "mate", Pagina: "producto", Producto: mate, Visitante: h("a")},
		{Tienda: "mate", Pagina: "producto", Producto: "no-existe", Visitante: h("a")},
		{Tienda: "mate", Pagina: "admin", Visitante: h("a")},
		{Tienda: "mate", Pagina: "carrito", Visitante: "corto"},
		{Tienda: "no-existe", Pagina: "inicio", Visitante: h("a")},
	}, ahora)
	if err != nil {
		t.Fatal(err)
	}
	if n != 4 {
		t.Fatalf("contó %d, esperaba 4", n)
	}
	// Un segundo lote del mismo día: la misma persona no vuelve a sumar como única.
	if _, err := visitas.Registrar(ctx, b.App, []visitas.Vista{
		{Tienda: "mate", Pagina: "inicio", Visitante: h("b")},
		{Tienda: "mate", Pagina: "carrito", Visitante: h("c")},
	}, ahora); err != nil {
		t.Fatal(err)
	}
	f := filas(t, b, id)
	want := map[string]fila{
		"tienda:":          {"tienda", "", 6, 3},
		"inicio:":          {"inicio", "", 4, 2},
		"producto:" + mate: {"producto", mate, 1, 1},
		"carrito:":         {"carrito", "", 1, 1},
	}
	if len(f) != len(want) {
		t.Fatalf("filas: %+v", f)
	}
	for k, w := range want {
		if f[k] != w {
			t.Fatalf("%s: %+v, esperaba %+v", k, f[k], w)
		}
	}

	// Otro día, la misma persona vuelve a contar (la sal cambia: no se la reconoce).
	if _, err := visitas.Registrar(ctx, b.App, []visitas.Vista{{Tienda: "mate", Pagina: "inicio", Visitante: h("a")}},
		ahora.Add(24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	var dias, unicos int
	_ = b.App.EnTienda(ctx, id, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT count(*), sum(visitantes) FROM visitas WHERE pagina = 'tienda'`).Scan(&dias, &unicos)
	})
	if dias != 2 || unicos != 4 {
		t.Fatalf("dos días: %d filas, %d visitantes", dias, unicos)
	}
}

// La dueña (o su equipo) mirando su propia tienda con la sesión del panel no cuenta; con la
// sesión de otra tienda, sí.
func TestLaDuenaNoCuenta(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	id, _ := tiendas.Crear(ctx, b.App, "mate", "Mate", "regionales")
	otra, _ := tiendas.Crear(ctx, b.App, "otra", "Otra", "regionales")
	n, err := visitas.Registrar(ctx, b.App, []visitas.Vista{
		{Tienda: "mate", Pagina: "inicio", Visitante: h("a"), Duena: id},
		{Tienda: "mate", Pagina: "inicio", Visitante: h("b"), Duena: otra},
	}, time.Now())
	if err != nil || n != 1 {
		t.Fatalf("contó %d (%v), esperaba 1", n, err)
	}
}

// RLS: lo de una tienda no se ve desde otra, y sin tienda fijada no se ve nada.
func TestVisitasAisladasPorTienda(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	una, _ := tiendas.Crear(ctx, b.App, "una", "Una", "regionales")
	otra, _ := tiendas.Crear(ctx, b.App, "otra", "Otra", "regionales")
	if _, err := visitas.Registrar(ctx, b.App, []visitas.Vista{{Tienda: "una", Pagina: "inicio", Visitante: h("a")}}, time.Now()); err != nil {
		t.Fatal(err)
	}
	if len(filas(t, b, una)) != 2 || len(filas(t, b, otra)) != 0 {
		t.Fatal("las visitas de una tienda se ven desde otra")
	}
	var n int
	if err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT (SELECT count(*) FROM visitas) + (SELECT count(*) FROM visitas_vistos)`).Scan(&n)
	}); err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("sin tienda se ven %d filas", n)
	}
	// Desde otra tienda tampoco se puede escribir en la de al lado.
	err := b.App.EnTienda(ctx, otra, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO visitas (tienda_id, dia, pagina, vistas) VALUES ($1, current_date, 'inicio', 99)`, una)
		return err
	})
	if err == nil {
		t.Fatal("una tienda escribió visitas de otra")
	}
}

// La limpieza borra los vistos de días anteriores (de todas las tiendas) y deja las visitas.
func TestPurgar(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	id, _ := tiendas.Crear(ctx, b.App, "mate", "Mate", "regionales")
	ayer := time.Now().Add(-24 * time.Hour)
	if _, err := visitas.Registrar(ctx, b.App, []visitas.Vista{{Tienda: "mate", Pagina: "inicio", Visitante: h("a")}}, ayer); err != nil {
		t.Fatal(err)
	}
	if _, err := visitas.Registrar(ctx, b.App, []visitas.Vista{{Tienda: "mate", Pagina: "inicio", Visitante: h("b")}}, time.Now()); err != nil {
		t.Fatal(err)
	}
	n, err := visitas.Purgar(ctx, b.App)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 {
		t.Fatalf("borró %d vistos, esperaba 2 (los de ayer)", n)
	}
	var vistos, filasVisitas int
	_ = b.App.EnTienda(ctx, id, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT (SELECT count(*) FROM visitas_vistos), (SELECT count(*) FROM visitas)`).Scan(&vistos, &filasVisitas)
	})
	if vistos != 2 || filasVisitas != 4 {
		t.Fatalf("quedaron %d vistos y %d visitas", vistos, filasVisitas)
	}
}

// El Resumen: visitas de hoy, 7 y 30 días, los más vistos y la conversión real (pedidos web por
// cada 100 visitas). Sin visitas, `desde` en null y la conversión en cero.
func TestTableroConVisitas(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	id, _ := tiendas.Crear(ctx, b.App, "mate", "Mate", "regionales")
	ahora := time.Now()

	tb, err := tiendas.ObtenerTablero(ctx, b.App, id, "7d", ahora)
	if err != nil {
		t.Fatal(err)
	}
	if tb.Visitas.Desde != nil || tb.Kpis.Conversion.Actual != 0 || tb.Visitas.MasVistos == nil {
		t.Fatalf("sin visitas: %+v", tb.Visitas)
	}

	mate := producto(t, b, id, "Mate de calabaza")
	var vs []visitas.Vista
	for i := range 40 {
		vs = append(vs, visitas.Vista{Tienda: "mate", Pagina: "inicio", Visitante: strings.Repeat(string(rune('a'+i%6)), 31) + string(rune('0'+i/6))})
	}
	vs = append(vs, visitas.Vista{Tienda: "mate", Pagina: "producto", Producto: mate, Visitante: vs[0].Visitante})
	if _, err := visitas.Registrar(ctx, b.App, vs, ahora); err != nil {
		t.Fatal(err)
	}
	mail := "ana@ejemplo.com"
	for _, canal := range []string{"web", "web", "wa"} {
		if _, err := pedidos.Crear(ctx, b.App, id, pedidos.Nuevo{
			Comprador: pedidos.Comprador{Nombre: "Ana", Email: &mail}, Canal: canal, Pago: "transf", Envio: "retiro",
			Estado: "Entregado", Creado: ahora,
			Items: []pedidos.Item{{Nombre: "Mate", Cantidad: 1, Precio: 1000}},
		}); err != nil {
			t.Fatal(err)
		}
	}
	tb, err = tiendas.ObtenerTablero(ctx, b.App, id, "7d", ahora)
	if err != nil {
		t.Fatal(err)
	}
	v := tb.Visitas
	if v.Desde == nil || v.Hoy.Actual != 40 || v.Dias7.Actual != 40 || v.Dias30.Actual != 40 || v.Vistas.Actual != 41 {
		t.Fatalf("visitas: %+v", v)
	}
	if len(v.MasVistos) != 1 || v.MasVistos[0].Nombre != "Mate de calabaza" || v.MasVistos[0].Vistas != 1 {
		t.Fatalf("más vistos: %+v", v.MasVistos)
	}
	// 2 pedidos web / 40 visitas = 5 %; el de WhatsApp no cuenta.
	if tb.Kpis.Conversion.Actual != 5 {
		t.Fatalf("conversión: %v", tb.Kpis.Conversion)
	}
	c := tiendas.ContextoHornero(ctx, b.App, id)
	hv, ok := c["visitas"].(map[string]any)
	if !ok || hv["visitas30d"] != float64(40) || hv["conversion30d"] != float64(5) {
		t.Fatalf("contexto de Tincho: %+v", c["visitas"])
	}
}

// La ruta interna sólo le cree a la web (clave del BFF): sin la clave, nadie suma visitas.
func TestRutaSoloConClave(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	id, _ := tiendas.Crear(ctx, b.App, "mate", "Mate", "regionales")
	mux := http.NewServeMux()
	visitas.Rutas(mux, b.App, "clave-de-prueba", nil)
	cuerpo := `{"visitas":[{"tienda":"mate","pagina":"inicio","visitante":"` + h("a") + `"}]}`
	for _, c := range []struct {
		clave  string
		estado int
	}{{"", http.StatusNotFound}, {"otra", http.StatusNotFound}, {"clave-de-prueba", http.StatusNoContent}} {
		r := httptest.NewRequest("POST", "/v1/interno/visitas", strings.NewReader(cuerpo))
		r.Header.Set("X-Mostrador-BFF", c.clave)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		if w.Code != c.estado {
			t.Fatalf("clave %q: %d, esperaba %d", c.clave, w.Code, c.estado)
		}
	}
	if f := filas(t, b, id); f["tienda:"].vistas != 1 {
		t.Fatalf("visitas: %+v", f)
	}
}
