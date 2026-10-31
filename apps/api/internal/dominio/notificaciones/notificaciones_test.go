package notificaciones_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/notificaciones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/ventas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/servidor"
)

// --- sin base ---

func TestTextoVentaSinMontosParaVentasYLocal(t *testing.T) {
	for rol, conMonto := range map[string]bool{"duenio": true, "admin": true, "ventas": false, "local": false} {
		tx := notificaciones.TextoVenta(rol, 7, 15215, 1, "transf", "Pendiente")
		if strings.Contains(tx.Detalle, "$") != conMonto {
			t.Errorf("%s: %q (con monto=%v)", rol, tx.Detalle, conMonto)
		}
		if !strings.HasSuffix(tx.Detalle, "1 producto · Transferencia por confirmar") || tx.Titulo != "Nueva venta #7" ||
			tx.Ruta != "/panel/pedidos/7" || tx.Grupo != "ventas" || tx.Tipo != "transf" {
			t.Errorf("%s: %+v", rol, tx)
		}
	}
	if tx := notificaciones.TextoVenta("duenio", 12, 24020, 3, "mp", "Aprobado"); tx.Detalle != "$ 24.020 · 3 productos · Pago acreditado en Mercado Pago" || tx.Tipo != "mp" {
		t.Errorf("mp: %+v", tx)
	}
}

// --- con base ---

type escenario struct {
	b         *pruebadb.Bases
	h         http.Handler
	luna, sol string
	g         *notificaciones.Generador
	ses, ids  map[string]string
	llamado   map[string]int
	ms        map[string]outbox.Manejador
}

func silencio() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }

func armar(t *testing.T) *escenario {
	t.Helper()
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	cs := &cuentas.Servicio{Pool: b.App, Correo: &correo.Memoria{}, Sitio: "https://tulocalweb.test"}
	e := &escenario{b: b, ses: map[string]string{}, ids: map[string]string{}, llamado: map[string]int{},
		h: servidor.Nuevo(servidor.Dependencias{App: b.App, Cuentas: cs, Log: silencio()}),
		g: &notificaciones.Generador{Pool: b.App}}
	e.luna = pruebadb.CrearTienda(t, b.App, "luna")
	e.sol = pruebadb.CrearTienda(t, b.App, "sol")
	hash, err := cuentas.HashClave("clave-larga-1")
	if err != nil {
		t.Fatal(err)
	}
	for _, rol := range []string{"duenio", "admin", "ventas", "local"} {
		email := rol + "@luna.com"
		if err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
			var id string
			if err := tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre, clave_hash, email_verificado_en)
				VALUES ($1, $2, $3, now()) RETURNING id::text`, email, "Persona "+rol, hash).Scan(&id); err != nil {
				return err
			}
			e.ids[rol] = id
			_, err := tx.Exec(ctx, `INSERT INTO auth.membresias (usuario_id, tienda_id, rol) VALUES ($1, $2, $3)`, id, e.luna, rol)
			return err
		}); err != nil {
			t.Fatal(err)
		}
		s, err := cs.Ingresar(ctx, email, "clave-larga-1", false, cuentas.Cliente{IP: "190.210.32.1", Agente: "Mozilla/5.0"})
		if err != nil {
			t.Fatal(err)
		}
		e.ses[rol] = s.Token
	}
	// La dueña de Luna también es dueña de Sol (otra tienda, otras notificaciones).
	e.admin(t, `INSERT INTO auth.membresias (usuario_id, tienda_id, rol) VALUES ($1, $2, 'duenio')`, e.ids["duenio"], e.sol)
	// Los manejadores de siempre (mail, push) se cuentan para ver que siguen corriendo.
	e.ms = map[string]outbox.Manejador{}
	for _, tipo := range []string{"tienda.venta", "tienda.sin_stock", "push.pregunta_ml"} {
		e.ms[tipo] = func(context.Context, outbox.Evento) error { e.llamado[tipo]++; return nil }
	}
	e.g.Envolver(e.ms)
	return e
}

func (e *escenario) conn(t *testing.T) *pgx.Conn {
	t.Helper()
	c, err := pgx.Connect(context.Background(), e.b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close(context.Background()) })
	return c
}

func (e *escenario) enTienda(t *testing.T, tienda, sql string, args ...any) {
	t.Helper()
	if err := e.b.App.EnTienda(context.Background(), tienda, func(tx pgx.Tx) error {
		_, err := tx.Exec(context.Background(), sql, args...)
		return err
	}); err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
}

func (e *escenario) admin(t *testing.T, sql string, args ...any) {
	t.Helper()
	if _, err := e.conn(t).Exec(context.Background(), sql, args...); err != nil {
		t.Fatalf("%s: %v", sql, err)
	}
}

// evento corre el manejador envuelto como lo haría el worker.
func (e *escenario) evento(t *testing.T, id int64, tienda, tipo, datos string) {
	t.Helper()
	m, ok := e.ms[tipo]
	if !ok {
		t.Fatalf("%s no tiene manejador", tipo)
	}
	if err := m(context.Background(), outbox.Evento{ID: id, TiendaID: tienda, Tipo: tipo, Datos: []byte(datos)}); err != nil {
		t.Fatalf("%s: %v", tipo, err)
	}
}

type fila struct{ rol, titulo, detalle, ruta, tipo, grupo string }

// de: las notificaciones de una tienda por rol de su destinataria (con el admin, sin RLS).
func (e *escenario) de(t *testing.T, tienda, clave string) map[string]fila {
	t.Helper()
	filas, err := e.conn(t).Query(context.Background(), `SELECT m.rol, n.titulo, n.detalle, n.ruta, n.tipo, n.grupo
		FROM notificaciones n JOIN auth.membresias m ON m.usuario_id = n.usuario_id AND m.tienda_id = n.tienda_id
		WHERE n.tienda_id = $1 AND n.clave LIKE $2`, tienda, clave)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]fila{}
	for filas.Next() {
		var f fila
		if err := filas.Scan(&f.rol, &f.titulo, &f.detalle, &f.ruta, &f.tipo, &f.grupo); err != nil {
			t.Fatal(err)
		}
		if _, ya := out[f.rol]; ya {
			t.Fatalf("dos notificaciones %s para %s", clave, f.rol)
		}
		out[f.rol] = f
	}
	return out
}

func roles(m map[string]fila) string {
	var out []string
	for _, r := range []string{"duenio", "admin", "ventas", "local"} {
		if _, ok := m[r]; ok {
			out = append(out, r)
		}
	}
	return strings.Join(out, ",")
}

func TestGeneracionPorEventoYRol(t *testing.T) {
	e := armar(t)
	ctx := context.Background()

	t.Run("venta: a Pedidos, sin montos para Ventas y Local, idempotente", func(t *testing.T) {
		e.enTienda(t, e.luna, `INSERT INTO pedidos (tienda_id, numero, comprador, canal, pago, estado_pago, envio, items, subtotal, total)
			VALUES ($1, 5, '{"nombre": "Lu"}', 'web', 'transf', 'Pendiente', 'retiro', '[{"nombre": "Mate", "cantidad": 2}]', 15000, 15000)`, e.luna)
		for range 3 {
			e.evento(t, 101, e.luna, "tienda.venta", `{"numero":5}`)
		}
		ns := e.de(t, e.luna, "venta:5")
		if roles(ns) != "duenio,admin,ventas,local" {
			t.Fatalf("destinatarias: %s", roles(ns))
		}
		if d := ns["duenio"].detalle; d != "$ 15.000 · 2 productos · Transferencia por confirmar" {
			t.Errorf("dueña: %q", d)
		}
		for _, r := range []string{"ventas", "local"} {
			if strings.Contains(ns[r].detalle, "$") || ns[r].detalle != "2 productos · Transferencia por confirmar" {
				t.Errorf("%s ve montos: %q", r, ns[r].detalle)
			}
		}
		if ns["admin"].ruta != "/panel/pedidos/5" || ns["admin"].titulo != "Nueva venta #5" {
			t.Errorf("%+v", ns["admin"])
		}
		if e.llamado["tienda.venta"] != 3 {
			t.Errorf("el mail de la venta corrió %d veces", e.llamado["tienda.venta"])
		}
		if len(e.de(t, e.sol, "%")) != 0 {
			t.Error("la venta de Luna dejó notificaciones en Sol")
		}
	})

	t.Run("venta cancelada antes de procesar: nada", func(t *testing.T) {
		e.enTienda(t, e.luna, `INSERT INTO pedidos (tienda_id, numero, comprador, canal, pago, estado, envio, items, subtotal, total)
			VALUES ($1, 6, '{"nombre": "Lu"}', 'web', 'transf', 'Cancelado', 'retiro', '[{"nombre": "Mate", "cantidad": 1}]', 100, 100)`, e.luna)
		e.evento(t, 102, e.luna, "tienda.venta", `{"numero":6}`)
		if ns := e.de(t, e.luna, "venta:6"); len(ns) != 0 {
			t.Fatalf("cancelada: %v", ns)
		}
	})

	t.Run("sin stock: a Productos (Local no)", func(t *testing.T) {
		var prod string
		if err := e.conn(t).QueryRow(ctx, `INSERT INTO productos (tienda_id, nombre, precio, stock, publicado) VALUES ($1, 'Yerba', 1000, 0, true)
			RETURNING id::text`, e.luna).Scan(&prod); err != nil {
			t.Fatal(err)
		}
		e.evento(t, 103, e.luna, "tienda.sin_stock", `{"producto":"`+prod+`"}`)
		e.evento(t, 103, e.luna, "tienda.sin_stock", `{"producto":"`+prod+`"}`)
		ns := e.de(t, e.luna, "sin_stock:%")
		if roles(ns) != "duenio,admin,ventas" {
			t.Fatalf("destinatarias: %s", roles(ns))
		}
		if f := ns["ventas"]; f.ruta != "/panel/productos?editar="+prod || f.titulo != "Te quedaste sin stock" || !strings.HasPrefix(f.detalle, "Yerba ") {
			t.Errorf("%+v", f)
		}
	})

	t.Run("reseña: la encola la tienda online y va a Clientes", func(t *testing.T) {
		if err := ventas.DejarResena(ctx, e.b.App, e.luna, ventas.NuevaResena{Autor: "Lu", Estrellas: 5, Texto: "Muy rico"}); err != nil {
			t.Fatal(err)
		}
		var id int64
		var datos string
		if err := e.conn(t).QueryRow(ctx, `SELECT id, datos::text FROM outbox WHERE tienda_id = $1 AND tipo = 'tienda.resena'`, e.luna).Scan(&id, &datos); err != nil {
			t.Fatalf("la reseña no encoló el aviso: %v", err)
		}
		e.evento(t, id, e.luna, "tienda.resena", datos)
		ns := e.de(t, e.luna, "resena:%")
		if roles(ns) != "duenio,admin,ventas" || ns["admin"].detalle != "5 de 5 estrellas." || strings.Contains(ns["admin"].detalle, "rico") {
			t.Fatalf("reseña: %s %+v", roles(ns), ns["admin"])
		}
	})

	t.Run("pago de un link: sólo a Finanzas, con el monto", func(t *testing.T) {
		e.admin(t, `INSERT INTO links (codigo, tienda_id, monto, concepto) VALUES ('abc123', $1, 5000, 'Seña del pedido grande')`, e.luna)
		e.evento(t, 104, e.luna, "pago.aprobado", `{"referencia":"r-1","tipo":"link","objetivo":"abc123","monto":5000}`)
		ns := e.de(t, e.luna, "pago:%")
		if roles(ns) != "duenio,admin" || ns["duenio"].detalle != "$ 5.000 · Seña del pedido grande" {
			t.Fatalf("link: %s %+v", roles(ns), ns["duenio"])
		}
		// El pago de un pedido de la tienda online ya avisó con «Nueva venta».
		e.evento(t, 105, e.luna, "pago.aprobado", `{"referencia":"r-2","tipo":"pedido","objetivo":"5","monto":15000}`)
		if ns := e.de(t, e.luna, "pago:r-2"); len(ns) != 0 {
			t.Fatalf("pedido web duplicado: %v", ns)
		}
	})

	t.Run("integración desconectada: a Tienda o Finanzas", func(t *testing.T) {
		e.evento(t, 106, e.luna, "integracion.reautorizar", `{"canal":"mp"}`)
		ns := e.de(t, e.luna, "reautorizar:%")
		if roles(ns) != "duenio,admin" || ns["admin"].titulo != "Mercado Pago se desconectó" || ns["admin"].ruta != "/panel/integraciones" {
			t.Fatalf("reautorizar: %s %+v", roles(ns), ns["admin"])
		}
	})

	t.Run("invitación aceptada: a la dueña, no a quien se sumó", func(t *testing.T) {
		e.admin(t, `INSERT INTO auth.membresias (usuario_id, tienda_id, rol) VALUES ($1, $2, 'ventas')`, e.ids["admin"], e.sol)
		e.evento(t, 107, e.sol, "equipo.aceptada", `{"usuario":"`+e.ids["admin"]+`"}`)
		ns := e.de(t, e.sol, "aceptada:%")
		if roles(ns) != "duenio" || ns["duenio"].titulo != "Persona admin se sumó al equipo" || ns["duenio"].detalle != "Aceptó la invitación como Ventas." {
			t.Fatalf("aceptada: %s %+v", roles(ns), ns["duenio"])
		}
	})

	t.Run("plan: sólo a la dueña", func(t *testing.T) {
		e.evento(t, 108, e.luna, "plan.activado", `{"plan":"pro"}`)
		e.evento(t, 109, e.luna, "plan.pausado", `{}`)
		ns := e.de(t, e.luna, "plan:108")
		if roles(ns) != "duenio" || ns["duenio"].titulo != "Tu plan Pro está activo" {
			t.Fatalf("plan: %s %+v", roles(ns), ns)
		}
		if roles(e.de(t, e.luna, "plan:109")) != "duenio" {
			t.Fatal("plan pausado")
		}
	})
}

// RLS: cada persona ve y marca sólo las suyas de la tienda fijada; el worker (sin persona) inserta
// pero no lee, y no puede insertar para alguien que no es del equipo.
func TestRLSPorTiendaYPersona(t *testing.T) {
	e := armar(t)
	ctx := context.Background()
	e.admin(t, `INSERT INTO notificaciones (tienda_id, usuario_id, clave, tipo, grupo, titulo, ruta)
		SELECT $1, m.usuario_id, 'x', 'ok', 'sistema', 'Hola', '/panel' FROM auth.membresias m WHERE m.tienda_id = $1`, e.luna)
	e.admin(t, `INSERT INTO notificaciones (tienda_id, usuario_id, clave, tipo, grupo, titulo, ruta) VALUES ($1, $2, 'x', 'ok', 'sistema', 'Sol', '/panel')`,
		e.sol, e.ids["duenio"])

	ver := func(tienda, usuario string) (vistas int, marcadas int64) {
		t.Helper()
		err := e.b.App.EnTienda(ctx, tienda, func(tx pgx.Tx) error {
			if usuario != "" {
				if _, err := tx.Exec(ctx, "SELECT set_config('app.usuario_id', $1, true)", usuario); err != nil {
					return err
				}
			}
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM notificaciones`).Scan(&vistas); err != nil {
				return err
			}
			tag, err := tx.Exec(ctx, `UPDATE notificaciones SET leida_en = now()`)
			marcadas = tag.RowsAffected()
			return err
		})
		if err != nil {
			t.Fatal(err)
		}
		return
	}
	if v, m := ver(e.luna, e.ids["ventas"]); v != 1 || m != 1 {
		t.Errorf("ventas en Luna: ve %d, marca %d", v, m)
	}
	if v, m := ver(e.luna, ""); v != 0 || m != 0 {
		t.Errorf("sin persona: ve %d, marca %d", v, m)
	}
	if v, _ := ver(e.sol, e.ids["ventas"]); v != 0 {
		t.Errorf("ventas en Sol (no es del equipo): ve %d", v)
	}
	if v, _ := ver(e.sol, e.ids["duenio"]); v != 1 {
		t.Errorf("la dueña en Sol: ve %d", v)
	}
	// La app no inserta directo: sólo por app.notificar, que descarta a quien no es del equipo de la
	// tienda fijada.
	err := e.b.App.EnTienda(ctx, e.sol, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO notificaciones (tienda_id, usuario_id, clave, tipo, grupo, titulo, ruta)
			VALUES ($1, $2, 'y', 'ok', 'sistema', 'Intrusa', '/panel')`, e.sol, e.ids["duenio"])
		return err
	})
	if err == nil || !strings.Contains(err.Error(), "permission denied") {
		t.Fatalf("insertar directo: %v", err)
	}
	var creadas int64
	err = e.b.App.EnTienda(ctx, e.sol, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT app.notificar('y', $1::uuid[], '{ok,ok}', '{sistema,sistema}', '{Hola,Intrusa}', '{"",""}', '{/panel,/panel}')`,
			[]string{e.ids["duenio"], e.ids["ventas"]}).Scan(&creadas)
	})
	if err != nil || creadas != 1 {
		t.Fatalf("notificar en Sol: %d %v (sólo la dueña es del equipo)", creadas, err)
	}
	err = e.b.App.EnTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `SELECT app.notificar('z', $1::uuid[], '{ok}', '{sistema}', '{Hola}', '{""}', '{/panel}')`, []string{e.ids["duenio"]})
		return err
	})
	if err == nil {
		t.Fatal("notificar sin tienda")
	}
	// La app no borra (sólo la purga).
	err = e.b.App.EnTienda(ctx, e.luna, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `DELETE FROM notificaciones`)
		return err
	})
	if err == nil {
		t.Fatal("la app pudo borrar")
	}
}

func (e *escenario) pedir(t *testing.T, rol, metodo, ruta string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(metodo, ruta, nil)
	if rol != "" {
		r.Header.Set("Authorization", "Bearer "+e.ses[rol])
	}
	rec := httptest.NewRecorder()
	e.h.ServeHTTP(rec, r)
	return rec
}

func TestAPI(t *testing.T) {
	e := armar(t)
	// Tres para Ventas (una vieja de 100 días, que ya no se muestra) y una para la dueña.
	e.admin(t, `INSERT INTO notificaciones (tienda_id, usuario_id, clave, tipo, grupo, titulo, ruta, creada_en) VALUES
		($1, $2, 'a', 'ok', 'sistema', 'Primera', '/panel', now() - interval '3 minutes'),
		($1, $2, 'b', 'mp', 'ventas', 'Segunda', '/panel/pedidos/1', now() - interval '2 minutes'),
		($1, $2, 'c', 'ml', 'ventas', 'Tercera', '/panel/resenas-y-preguntas', now() - interval '1 minute'),
		($1, $2, 'vieja', 'ok', 'sistema', 'Vieja', '/panel', now() - interval '100 days'),
		($1, $3, 'd', 'ok', 'sistema', 'De la dueña', '/panel', now())`, e.luna, e.ids["ventas"], e.ids["duenio"])

	if rec := e.pedir(t, "", "GET", "/v1/panel/notificaciones"); rec.Code != http.StatusUnauthorized {
		t.Fatalf("sin sesión: %d", rec.Code)
	}
	leer := func(ruta string) notificaciones.Pagina {
		t.Helper()
		rec := e.pedir(t, "ventas", "GET", ruta)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s: %d %s", ruta, rec.Code, rec.Body)
		}
		var p notificaciones.Pagina
		if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
			t.Fatal(err)
		}
		return p
	}
	p := leer("/v1/panel/notificaciones")
	if len(p.Notificaciones) != 3 || p.SinLeer != 3 || p.Siguiente != nil || p.Notificaciones[0].Titulo != "Tercera" ||
		p.Notificaciones[0].Ruta != "/panel/resenas-y-preguntas" || p.Notificaciones[0].Leida {
		t.Fatalf("lista: %+v", p)
	}
	// De a una, con cursor.
	p1 := leer("/v1/panel/notificaciones?limite=2")
	if len(p1.Notificaciones) != 2 || p1.Siguiente == nil {
		t.Fatalf("página 1: %+v", p1)
	}
	p2 := leer("/v1/panel/notificaciones?limite=2&cursor=" + *p1.Siguiente)
	if len(p2.Notificaciones) != 1 || p2.Notificaciones[0].Titulo != "Primera" || p2.Siguiente != nil {
		t.Fatalf("página 2: %+v", p2)
	}
	if rec := e.pedir(t, "ventas", "GET", "/v1/panel/notificaciones?cursor=nada"); rec.Code != http.StatusBadRequest {
		t.Fatalf("cursor roto: %d", rec.Code)
	}
	// Marcar una: la de otra persona no existe para Ventas.
	var deLaDuenia string
	if err := e.conn(t).QueryRow(context.Background(), `SELECT id::text FROM notificaciones WHERE clave = 'd'`).Scan(&deLaDuenia); err != nil {
		t.Fatal(err)
	}
	if rec := e.pedir(t, "ventas", "POST", "/v1/panel/notificaciones/"+deLaDuenia+"/leida"); rec.Code != http.StatusNotFound {
		t.Fatalf("ajena: %d", rec.Code)
	}
	if rec := e.pedir(t, "ventas", "POST", "/v1/panel/notificaciones/"+p.Notificaciones[0].ID+"/leida"); rec.Code != http.StatusNoContent {
		t.Fatalf("marcar una: %d %s", rec.Code, rec.Body)
	}
	rec := e.pedir(t, "ventas", "GET", "/v1/panel/notificaciones/sin-leer")
	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != `{"sinLeer":2}` {
		t.Fatalf("sin leer: %d %s", rec.Code, rec.Body)
	}
	if p := leer("/v1/panel/notificaciones"); !p.Notificaciones[0].Leida || p.SinLeer != 2 {
		t.Fatalf("tras marcar: %+v", p)
	}
	// Todas: las de la dueña siguen sin leer.
	if rec := e.pedir(t, "ventas", "POST", "/v1/panel/notificaciones/leidas"); rec.Code != http.StatusNoContent {
		t.Fatalf("todas: %d", rec.Code)
	}
	if p := leer("/v1/panel/notificaciones"); p.SinLeer != 0 {
		t.Fatalf("tras todas: %+v", p)
	}
	rec = e.pedir(t, "duenio", "GET", "/v1/panel/notificaciones/sin-leer")
	if strings.TrimSpace(rec.Body.String()) != `{"sinLeer":1}` {
		t.Fatalf("la dueña: %s", rec.Body)
	}
	// Local también entra (roles.Panel) y no ve nada que no sea suyo.
	rec = e.pedir(t, "local", "GET", "/v1/panel/notificaciones")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"notificaciones":[]`) {
		t.Fatalf("local: %d %s", rec.Code, rec.Body)
	}
}

func TestPurga(t *testing.T) {
	e := armar(t)
	e.admin(t, `INSERT INTO notificaciones (tienda_id, usuario_id, clave, tipo, grupo, titulo, ruta, creada_en) VALUES
		($1, $2, 'vieja', 'ok', 'sistema', 'Vieja', '/panel', now() - interval '91 days'),
		($1, $2, 'casi', 'ok', 'sistema', 'Casi', '/panel', now() - interval '89 days'),
		($3, $2, 'sol', 'ok', 'sistema', 'Sol vieja', '/panel', now() - interval '200 days')`, e.luna, e.ids["duenio"], e.sol)
	n, err := notificaciones.Purgar(context.Background(), e.b.App)
	if err != nil || n != 2 {
		t.Fatalf("purgó %d: %v", n, err)
	}
	var quedan []string
	filas, err := e.conn(t).Query(context.Background(), `SELECT clave FROM notificaciones`)
	if err != nil {
		t.Fatal(err)
	}
	if quedan, err = pgx.CollectRows(filas, pgx.RowTo[string]); err != nil || len(quedan) != 1 || quedan[0] != "casi" {
		t.Fatalf("quedan %v: %v", quedan, err)
	}
}
