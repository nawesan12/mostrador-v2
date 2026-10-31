package tiemporeal

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// Con Postgres de verdad: los triggers de la 00061, la escucha y el hub juntos.
type entorno struct {
	b     *pruebadb.Bases
	hub   *Hub
	esc   *Escucha
	admin *pgx.Conn
	sol   string
	luna  string
}

func nuevoEntorno(t *testing.T) *entorno {
	t.Helper()
	b := pruebadb.Nueva(t)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	h := NuevoHub()
	h.Ventana = 20 * time.Millisecond
	go h.Correr(ctx)
	sep := "?"
	if strings.Contains(b.AdminURL, "?") {
		sep = "&"
	}
	esc := &Escucha{URL: b.AdminURL + sep + "application_name=tr_escucha", Hub: h, Latido: 200 * time.Millisecond,
		EsperaMin: 20 * time.Millisecond, EsperaMax: 100 * time.Millisecond}
	go esc.Correr(ctx)
	admin, err := pgx.Connect(ctx, b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = admin.Close(context.Background()) })
	e := &entorno{b: b, hub: h, esc: esc, admin: admin,
		sol: pruebadb.CrearTienda(t, b.App, "sol"), luna: pruebadb.CrearTienda(t, b.App, "luna")}
	e.esperarConectada(t)
	return e
}

func (e *entorno) esperarConectada(t *testing.T) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !e.esc.Conectada() {
		if time.Now().After(deadline) {
			t.Fatal("la escucha no conectó")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (e *entorno) sql(t *testing.T, q string, args ...any) {
	t.Helper()
	if _, err := e.admin.Exec(context.Background(), q, args...); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
}

// enTienda corre q con el rol de la app y la tienda fijada (RLS), como la API.
func (e *entorno) enTienda(t *testing.T, tienda, q string, args ...any) {
	t.Helper()
	ctx := context.Background()
	if err := e.b.App.EnTienda(ctx, tienda, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, q, args...)
		return err
	}); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
}

// llegan junta lo que recibe el cliente hasta que pasa un rato sin nada.
func llegan(c *Cliente, espera time.Duration) []Mensaje {
	var ms []Mensaje
	for {
		select {
		case m := <-c.envio:
			ms = append(ms, m)
		case <-time.After(espera):
			return ms
		}
	}
}

func hay(ms []Mensaje, tipo, id string) bool {
	for _, m := range ms {
		if m.Tipo == tipo && (id == "*" || m.ID == id) {
			return true
		}
	}
	return false
}

func TestTriggersAvisanAlConfirmar(t *testing.T) {
	e := nuevoEntorno(t)
	ctx := context.Background()
	sol := registrar(t, e.hub, &Cliente{Tienda: e.sol, Usuario: "ana", Sesion: "s1", Rol: "duenio"})
	local := registrar(t, e.hub, &Cliente{Tienda: e.sol, Usuario: "dani", Sesion: "s2", Rol: "local"})
	luna := registrar(t, e.hub, &Cliente{Tienda: e.luna, Usuario: "eva", Sesion: "s3", Rol: "duenio"})
	vidriera := registrar(t, e.hub, &Cliente{Tienda: e.sol, Publico: true})
	comprador := registrar(t, e.hub, &Cliente{Tienda: e.sol, Publico: true, Pedido: "1"})

	// Con la app y RLS: un producto de Sol.
	var prod string
	if err := e.b.App.EnTienda(ctx, e.sol, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `INSERT INTO productos (tienda_id, nombre, precio, stock) VALUES ($1, 'Mate', 1000, 3) RETURNING id::text`, e.sol).Scan(&prod)
	}); err != nil {
		t.Fatal(err)
	}
	ms := llegan(sol, 300*time.Millisecond)
	if !hay(ms, "producto.nuevo", prod) {
		t.Fatalf("producto nuevo: %v", ms)
	}
	if len(llegan(luna, 50*time.Millisecond)) != 0 {
		t.Fatal("a otra tienda no le llega nada")
	}
	llegan(local, 10*time.Millisecond)
	llegan(vidriera, 10*time.Millisecond)

	// Lo que se revierte no avisa.
	_ = e.b.App.EnTienda(ctx, e.sol, func(tx pgx.Tx) error {
		_, _ = tx.Exec(ctx, `UPDATE productos SET stock = 0 WHERE id = $1`, prod)
		return context.Canceled
	})
	if ms := llegan(sol, 200*time.Millisecond); len(ms) != 0 {
		t.Fatalf("una transacción revertida no avisa: %v", ms)
	}

	// Stock: le llega al equipo (también a Local, por el punto de venta) y a la tienda online.
	e.sql(t, `UPDATE productos SET stock = 2 WHERE id = $1`, prod)
	for _, c := range []*Cliente{sol, local, vidriera} {
		if ms := llegan(c, 300*time.Millisecond); !hay(ms, "producto.stock", prod) || hay(ms, "producto.cambio", "*") {
			t.Fatalf("stock (%s): %v", c.Rol, ms)
		}
	}
	// Un cambio que no se mira (el día del aviso de sin stock) no avisa.
	e.sql(t, `UPDATE productos SET sin_stock_avisado = current_date WHERE id = $1`, prod)
	if ms := llegan(sol, 200*time.Millisecond); len(ms) != 0 {
		t.Fatalf("columna interna: %v", ms)
	}

	// Pedido nuevo y pago acreditado: al equipo con Pedidos y, el pago, al comprador de ese pedido.
	e.enTienda(t, e.sol, `INSERT INTO pedidos (tienda_id, numero, comprador, canal, pago, envio, items, subtotal, total)
		VALUES ($1, 1, '{"nombre": "Lu"}', 'pos', 'mp', 'retiro', '[{"nombre": "Mate"}]', 100, 100)`, e.sol)
	if ms := llegan(local, 300*time.Millisecond); !hay(ms, "pedido.nuevo", "1") {
		t.Fatalf("pedido nuevo: %v", ms)
	}
	llegan(sol, 10*time.Millisecond)
	e.enTienda(t, e.sol, `UPDATE pedidos SET estado_pago = 'Acreditada' WHERE tienda_id = $1 AND numero = 1`, e.sol)
	ms = llegan(sol, 300*time.Millisecond)
	if !hay(ms, "pago.acreditado", "1") || !hay(ms, "pedido.cambio", "1") {
		t.Fatalf("pago: %v", ms)
	}
	if ms := llegan(comprador, 100*time.Millisecond); !hay(ms, "pago.acreditado", "") {
		t.Fatalf("el comprador se entera del pago: %v", ms)
	}
	// Un pago ya acreditado que cambia de nombre (Acreditada → Cobrado) no vuelve a avisar el pago.
	e.enTienda(t, e.sol, `UPDATE pedidos SET estado_pago = 'Cobrado' WHERE tienda_id = $1 AND numero = 1`, e.sol)
	if ms := llegan(sol, 300*time.Millisecond); hay(ms, "pago.acreditado", "*") {
		t.Fatalf("el pago avisa una vez: %v", ms)
	}

	llegan(local, 10*time.Millisecond)
	// Plata: sólo a quien tiene Finanzas.
	e.sql(t, `INSERT INTO links (codigo, tienda_id, monto, concepto) VALUES ('abc123', $1, 100, 'Seña')`, e.sol)
	if ms := llegan(sol, 300*time.Millisecond); !hay(ms, "finanzas.cambio", "") {
		t.Fatalf("finanzas a la dueña: %v", ms)
	}
	if ms := llegan(local, 50*time.Millisecond); len(ms) != 0 {
		t.Fatalf("Local no ve plata: %v", ms)
	}
}

func TestLoteEnPostgres(t *testing.T) {
	e := nuevoEntorno(t)
	c := registrar(t, e.hub, &Cliente{Tienda: e.sol, Usuario: "ana", Sesion: "s1", Rol: "duenio"})
	// Una importación: 500 productos en una sentencia → un aviso.
	e.sql(t, `INSERT INTO productos (tienda_id, nombre, precio, stock) SELECT $1, 'P' || g, 100, 5 FROM generate_series(1, 500) g`, e.sol)
	ms := llegan(c, 400*time.Millisecond)
	if len(ms) != 1 || ms[0].Tipo != "producto.nuevo" || !ms[0].Lote {
		t.Fatalf("una importación sale como un lote: %v", ms)
	}
	// Un aumento masivo de precios: un aviso; el stock no cambió, así que no hay aviso de stock.
	e.sql(t, `UPDATE productos SET precio = precio * 2 WHERE tienda_id = $1`, e.sol)
	ms = llegan(c, 400*time.Millisecond)
	if len(ms) != 1 || ms[0].Tipo != "producto.cambio" || !ms[0].Lote {
		t.Fatalf("un aumento masivo sale como un lote: %v", ms)
	}
	// Pocas filas: una por fila.
	e.sql(t, `UPDATE productos SET stock = 0 WHERE tienda_id = $1 AND nombre IN ('P1', 'P2', 'P3')`, e.sol)
	ms = llegan(c, 400*time.Millisecond)
	if len(ms) != 3 {
		t.Fatalf("tres filas, tres avisos: %v", ms)
	}
}

func TestReconexionDeLaEscucha(t *testing.T) {
	e := nuevoEntorno(t)
	c := registrar(t, e.hub, &Cliente{Tienda: e.sol, Usuario: "ana", Sesion: "s1", Rol: "duenio"})
	// Se corta la conexión de la escucha (reinicio de Postgres, red): al volver, «resincronizá».
	var n int
	if err := e.admin.QueryRow(context.Background(),
		`SELECT count(pg_terminate_backend(pid)) FROM pg_stat_activity WHERE application_name = 'tr_escucha'`).Scan(&n); err != nil || n == 0 {
		t.Fatalf("cortar la escucha: %v %d", err, n)
	}
	deadline := time.Now().Add(10 * time.Second)
	for {
		ms := llegan(c, 100*time.Millisecond)
		if hay(ms, TipoResincronizar, "") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("al reconectar no se pidió resincronizar")
		}
	}
	e.esperarConectada(t)
	// Y sigue escuchando.
	e.sql(t, `INSERT INTO clientes (tienda_id, nombre, origen) VALUES ($1, 'Lu', 'panel')`, e.sol)
	if ms := llegan(c, 400*time.Millisecond); !hay(ms, "cliente.nuevo", "*") {
		t.Fatalf("después de reconectar: %v", ms)
	}
}

func TestLatidoSinAvisos(t *testing.T) {
	e := nuevoEntorno(t)
	c := registrar(t, e.hub, &Cliente{Tienda: e.sol, Usuario: "ana", Sesion: "s1", Rol: "duenio"})
	// Pasan varios latidos sin avisos (espera con plazo + ping) y la conexión sigue sirviendo.
	time.Sleep(700 * time.Millisecond)
	if ms := llegan(c, 50*time.Millisecond); hay(ms, TipoResincronizar, "") {
		t.Fatal("un rato sin avisos no es un corte")
	}
	e.sql(t, `INSERT INTO clientes (tienda_id, nombre, origen) VALUES ($1, 'Lu', 'panel')`, e.sol)
	if ms := llegan(c, 400*time.Millisecond); !hay(ms, "cliente.nuevo", "*") {
		t.Fatalf("después de los latidos: %v", ms)
	}
}

func TestSesionesYEquipoDesdeLaBase(t *testing.T) {
	e := nuevoEntorno(t)
	ctx := context.Background()
	var usuario string
	if err := e.admin.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre) VALUES ('ana@sol.com', 'Ana') RETURNING id::text`).Scan(&usuario); err != nil {
		t.Fatal(err)
	}
	e.sql(t, `INSERT INTO auth.membresias (usuario_id, tienda_id, rol) VALUES ($1, $2, 'duenio'), ($1, $3, 'ventas')`, usuario, e.sol, e.luna)
	e.sql(t, `INSERT INTO auth.sesiones (token_hash, usuario_id, tienda_id, vence_en) VALUES
		('hash-uno-0123456789abcdef', $1, $2, now() + interval '1 day'),
		('hash-dos-0123456789abcdef', $1, $2, now() + interval '1 day')`, usuario, e.sol)
	uno := registrar(t, e.hub, &Cliente{Tienda: e.sol, Usuario: usuario, Sesion: "hash-uno-0123456", Rol: "duenio"})
	dos := registrar(t, e.hub, &Cliente{Tienda: e.sol, Usuario: usuario, Sesion: "hash-dos-0123456", Rol: "duenio"})
	llegan(uno, 200*time.Millisecond)
	llegan(dos, 10*time.Millisecond)

	// Cambiar de tienda una sesión: esa conexión resincroniza; la otra sólo ve la lista cambiada.
	e.sql(t, `UPDATE auth.sesiones SET tienda_id = $1 WHERE token_hash = 'hash-uno-0123456789abcdef'`, e.luna)
	if k := esperarCorte(t, uno); k.Codigo != CierreResincronizar {
		t.Fatalf("cambio de tienda: %+v", k)
	}
	if ms := llegan(dos, 200*time.Millisecond); !hay(ms, "sesion.cambio", "") {
		t.Fatalf("la otra pestaña ve la lista de sesiones cambiada: %v", ms)
	}
	// El «visto» no avisa nada.
	e.sql(t, `UPDATE auth.sesiones SET visto_en = now()`)
	if ms := llegan(dos, 200*time.Millisecond); len(ms) != 0 {
		t.Fatalf("el visto no es un cambio: %v", ms)
	}
	// Cerrar la otra sesión: se corta con 4401.
	e.sql(t, `DELETE FROM auth.sesiones WHERE token_hash = 'hash-dos-0123456789abcdef'`)
	if k := esperarCorte(t, dos); k.Codigo != CierreSinSesion {
		t.Fatalf("sesión cerrada: %+v", k)
	}

	// Cambio de rol en Luna: la persona afectada en Luna resincroniza.
	enLuna := registrar(t, e.hub, &Cliente{Tienda: e.luna, Usuario: usuario, Sesion: "hash-uno-0123456", Rol: "ventas"})
	e.sql(t, `UPDATE auth.membresias SET rol = 'admin' WHERE usuario_id = $1 AND tienda_id = $2`, usuario, e.luna)
	if k := esperarCorte(t, enLuna); k.Codigo != CierreResincronizar {
		t.Fatalf("cambio de rol: %+v", k)
	}
}

func esperarCorte(t *testing.T, c *Cliente) Corte {
	t.Helper()
	select {
	case k := <-c.corte:
		return k
	case <-time.After(3 * time.Second):
		t.Fatal("no se cortó")
		return Corte{}
	}
}

// Cada tipo que escriben los triggers tiene destinatarios en el catálogo (si no, no le llegaría a
// nadie y nadie se daría cuenta).
func TestTiposDeLosTriggersEnElCatalogo(t *testing.T) {
	e := nuevoEntorno(t)
	rows, err := e.admin.Query(context.Background(), `
		SELECT tgname, encode(tgargs, 'escape') FROM pg_trigger
		WHERE tgname LIKE 'tiempo\_real\_%' AND NOT tgisinternal`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	n := 0
	for rows.Next() {
		var nombre, args string
		if err := rows.Scan(&nombre, &args); err != nil {
			t.Fatal(err)
		}
		n++
		tipo, _, _ := strings.Cut(args, `\000`)
		if tipo == "" {
			continue // los de función propia (pago, sesiones) se prueban aparte
		}
		if !strings.Contains(tipo, ".") {
			tipo += ".cambio"
		}
		if !Conocido(tipo) {
			t.Errorf("%s manda %q, que no tiene destinatarios", nombre, tipo)
		}
	}
	if n < 40 {
		t.Fatalf("faltan triggers: %d", n)
	}
	for _, tipo := range []string{"pago.acreditado", "sesion.cambio", "sesion.cerrada", "sesion.tienda"} {
		if !Conocido(tipo) {
			t.Errorf("%s sin destinatarios", tipo)
		}
	}
}
