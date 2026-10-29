package equipo_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/equipo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/servidor"
)

const clave = "una-clave-larga-del-equipo"

func pedir(h http.Handler, metodo, ruta, cuerpo, token string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec
}

func leer[T any](t *testing.T, rec *httptest.ResponseRecorder, estado int) T {
	t.Helper()
	var v T
	if rec.Code != estado {
		t.Fatalf("esperaba %d, vino %d: %s", estado, rec.Code, rec.Body)
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &v)
	return v
}

func estado(t *testing.T, rec *httptest.ResponseRecorder, esperado int) {
	t.Helper()
	if rec.Code != esperado {
		t.Fatalf("esperaba %d, vino %d: %s", esperado, rec.Code, rec.Body)
	}
}

type entorno struct {
	h     http.Handler
	admin *pgx.Conn
	s     *equipo.Servicio
	b     *pruebadb.Bases
	// Tokens del equipo por rol y miembros.
	tok map[string]string
	ids map[string]string
	// La tienda de prueba y el token de su dueña.
	tienda, slug, dueña string
}

func armar(t *testing.T) *entorno {
	t.Helper()
	b := pruebadb.Nueva(t)
	ctx := context.Background()
	admin, err := pgx.Connect(ctx, b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { admin.Close(context.Background()) })
	e := &entorno{admin: admin, b: b, tok: map[string]string{}, ids: map[string]string{}}
	e.s = &equipo.Servicio{Staff: b.Staff, App: b.App, Sitio: "https://tulocalweb.com.ar"}
	e.h = servidor.Nuevo(servidor.Dependencias{App: b.App, Equipo: e.s,
		Cuentas: &cuentas.Servicio{Pool: b.App, Correo: &correo.Memoria{}, Sitio: "https://tulocalweb.com.ar"}})
	hash, _ := cuentas.HashClave(clave)
	for _, rol := range []string{"soporte", "administracion", "duenio"} {
		var id string
		if err := admin.QueryRow(ctx, `INSERT INTO equipo.miembros (email, nombre, clave_hash, rol) VALUES ($1, $2, $3, $4) RETURNING id::text`,
			rol+"@tulocalweb.com.ar", "Persona "+rol, hash, rol).Scan(&id); err != nil {
			t.Fatal(err)
		}
		e.ids[rol] = id
		ses := leer[equipo.SesionNueva](t, pedir(e.h, "POST", "/v1/equipo/ingresar",
			`{"email":"`+rol+`@tulocalweb.com.ar","clave":"`+clave+`"}`, ""), 200)
		e.tok[rol] = ses.Token
	}
	s := leer[cuentas.Sesion](t, pedir(e.h, "POST", "/v1/auth/registro",
		`{"nombre":"Lucía","tienda":"La Esquina","email":"lucia@laesquina.com.ar","clave":"Abcdefghijk1"}`, ""), 201)
	e.dueña, e.tienda = s.Token, s.TiendaID
	if err := admin.QueryRow(ctx, `SELECT slug FROM tiendas WHERE id = $1`, e.tienda).Scan(&e.slug); err != nil {
		t.Fatal(err)
	}
	return e
}

func (e *entorno) acciones(t *testing.T) []string {
	t.Helper()
	filas, _ := e.admin.Query(context.Background(), `SELECT accion FROM equipo.auditoria ORDER BY id`)
	a, err := pgx.CollectRows(filas, pgx.RowTo[string])
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestLaMatrizDePermisosEsLaMismaEnGoJSONySQL(t *testing.T) {
	b := pruebadb.Nueva(t)
	crudo, err := os.ReadFile("testdata/permisos.json")
	if err != nil {
		t.Fatal(err)
	}
	var js map[string][]string
	if err := json.Unmarshal(crudo, &js); err != nil {
		t.Fatal(err)
	}
	filas, _ := b.Staff.Query(context.Background(), `SELECT rol, permiso FROM equipo.permisos`)
	sql := map[string][]string{}
	for filas.Next() {
		var r, p string
		_ = filas.Scan(&r, &p)
		sql[r] = append(sql[r], p)
	}
	ordenar := func(m map[string][]string) map[string][]string {
		for k := range m {
			m[k] = slices.Clone(m[k])
			sort.Strings(m[k])
		}
		return m
	}
	g, j, s := ordenar(maps(equipo.Permisos)), ordenar(js), ordenar(sql)
	for _, rol := range []string{"soporte", "administracion", "duenio"} {
		if !slices.Equal(g[rol], j[rol]) || !slices.Equal(g[rol], s[rol]) {
			t.Errorf("%s: go=%v json=%v sql=%v", rol, g[rol], j[rol], s[rol])
		}
	}
	if len(g) != 3 || len(j) != 3 || len(s) != 3 {
		t.Fatalf("roles: go=%d json=%d sql=%d", len(g), len(j), len(s))
	}
}

func maps(m map[string][]string) map[string][]string {
	c := map[string][]string{}
	for k, v := range m {
		c[k] = slices.Clone(v)
	}
	return c
}

func TestSesionesSeparadas(t *testing.T) {
	e := armar(t)
	// Un token de una tienda no abre el backoffice.
	estado(t, pedir(e.h, "GET", "/v1/equipo/sesion", "", e.dueña), 401)
	estado(t, pedir(e.h, "GET", "/v1/equipo/tiendas", "", e.dueña), 401)
	// Uno del equipo no abre el panel de ninguna tienda.
	estado(t, pedir(e.h, "GET", "/v1/auth/sesion", "", e.tok["duenio"]), 401)
	estado(t, pedir(e.h, "GET", "/v1/panel/cuenta/plan", "", e.tok["duenio"]), 401)

	ses := leer[map[string]any](t, pedir(e.h, "GET", "/v1/equipo/sesion", "", e.tok["soporte"]), 200)
	if ses["rol"] != "soporte" || len(ses["permisos"].([]any)) != 7 || ses["miembro"].(map[string]any)["email"] != "soporte@tulocalweb.com.ar" {
		t.Fatalf("sesión: %v", ses)
	}
	// Clave incorrecta y persona dada de baja responden igual.
	rec := pedir(e.h, "POST", "/v1/equipo/ingresar", `{"email":"soporte@tulocalweb.com.ar","clave":"otra-cosa"}`, "")
	if rec.Code != 401 || !strings.Contains(rec.Body.String(), "credenciales_invalidas") {
		t.Fatalf("clave mala: %d %s", rec.Code, rec.Body)
	}
	_, _ = e.admin.Exec(context.Background(), `UPDATE equipo.miembros SET activo = false WHERE id = $1`, e.ids["soporte"])
	estado(t, pedir(e.h, "POST", "/v1/equipo/ingresar", `{"email":"soporte@tulocalweb.com.ar","clave":"`+clave+`"}`, ""), 401)
	// Y su sesión abierta deja de valer.
	estado(t, pedir(e.h, "GET", "/v1/equipo/sesion", "", e.tok["soporte"]), 401)
	estado(t, pedir(e.h, "POST", "/v1/equipo/salir", "", e.tok["duenio"]), 204)
	estado(t, pedir(e.h, "GET", "/v1/equipo/sesion", "", e.tok["duenio"]), 401)
	if a := e.acciones(t); a[0] != "ingreso" {
		t.Fatalf("el ingreso no quedó auditado: %v", a)
	}
}

func TestPermisosPorRolYAuditoria(t *testing.T) {
	e := armar(t)
	ruta := "/v1/equipo/tiendas/" + e.tienda + "/acciones"
	accion := func(rol, cuerpo string) *httptest.ResponseRecorder {
		return pedir(e.h, "POST", ruta, cuerpo, e.tok[rol])
	}

	// Soporte: extiende la prueba y regala packs; no cambia el plan, no aplica crédito, no pausa ni cierra.
	estado(t, accion("soporte", `{"accion":"extender_prueba","dias":7}`), 204)
	estado(t, accion("soporte", `{"accion":"regalar_pack","cantidad":2}`), 204)
	for _, c := range []string{`{"accion":"cambiar_plan","plan":"pro"}`, `{"accion":"aplicar_credito","monto":500000,"motivo":"Compensación"}`,
		`{"accion":"pausar","motivo":"Pedido de la dueña"}`, `{"accion":"cerrar","motivo":"Pedido de la dueña","confirmacion":"` + e.slug + `"}`} {
		estado(t, accion("soporte", c), 403)
	}
	for _, r := range []string{"/v1/equipo/facturacion", "/v1/equipo/auditoria", "/v1/equipo/partners"} {
		estado(t, pedir(e.h, "GET", r, "", e.tok["soporte"]), 403)
	}
	// Administración: plan, crédito (con motivo) y pausa; no cierra.
	estado(t, accion("administracion", `{"accion":"cambiar_plan","plan":"pro"}`), 204)
	estado(t, accion("administracion", `{"accion":"aplicar_credito","monto":500000}`), 400)
	estado(t, accion("administracion", `{"accion":"aplicar_credito","monto":500000,"motivo":"Compensación por la caída"}`), 204)
	estado(t, accion("administracion", `{"accion":"pausar","motivo":"Pedido de la dueña"}`), 204)
	estado(t, accion("administracion", `{"accion":"reactivar"}`), 204)
	estado(t, accion("administracion", `{"accion":"cerrar","motivo":"Pedido de la dueña","confirmacion":"`+e.slug+`"}`), 403)
	estado(t, pedir(e.h, "GET", "/v1/equipo/facturacion", "", e.tok["administracion"]), 200)

	ctx := context.Background()
	var plan string
	var credito int64
	var hasta time.Time
	var pausada *time.Time
	if err := e.admin.QueryRow(ctx, `SELECT plan, credito, prueba_hasta, pausada_en FROM tiendas WHERE id = $1`, e.tienda).
		Scan(&plan, &credito, &hasta, &pausada); err != nil {
		t.Fatal(err)
	}
	if plan != "pro" || credito != 500000 || hasta.Before(time.Now().Add(20*24*time.Hour)) || pausada != nil {
		t.Fatalf("tienda: plan=%s credito=%d prueba=%s pausada=%v", plan, credito, hasta, pausada)
	}

	// Dueño: cierra, con la dirección escrita a mano.
	estado(t, accion("duenio", `{"accion":"cerrar","motivo":"Pedido de la dueña","confirmacion":"otra"}`), 400)
	estado(t, accion("duenio", `{"accion":"cerrar","motivo":"Pedido de la dueña","confirmacion":"`+e.slug+`"}`), 204)
	estado(t, accion("duenio", `{"accion":"extender_prueba","dias":3}`), 409)

	// Cada acción quedó en la auditoría del equipo y en la de la tienda (como staff).
	want := []string{"extender_prueba", "regalar_pack", "cambiar_plan", "aplicar_credito", "pausar", "reactivar", "cerrar"}
	var hechas []string
	for _, a := range e.acciones(t) {
		if a != "ingreso" {
			hechas = append(hechas, a)
		}
	}
	if !slices.Equal(hechas, want) {
		t.Fatalf("auditoría del equipo: %v", hechas)
	}
	var n int
	_ = e.admin.QueryRow(ctx, `SELECT count(*) FROM auditoria WHERE tienda_id = $1 AND actor_tipo = 'staff' AND accion LIKE 'equipo.%'`, e.tienda).Scan(&n)
	if n != len(want) {
		t.Fatalf("auditoría de la tienda: %d registros del equipo", n)
	}
	reg := leer[map[string]any](t, pedir(e.h, "GET", "/v1/equipo/auditoria?accion=aplicar_credito", "", e.tok["duenio"]), 200)
	r := reg["registros"].([]any)
	if len(r) != 1 || !strings.Contains(r[0].(map[string]any)["texto"].(string), "$ 5.000 de crédito") {
		t.Fatalf("auditoría filtrada: %v", r)
	}

	// La función SQL también rechaza a soporte aunque alguien se saltee el chequeo de Go.
	err := e.b.Staff.QueryRow(ctx, `SELECT equipo.accion_tienda($1, $2, 'cambiar_plan', '{"plan":"negocio"}', 'x', '', '')`,
		e.ids["soporte"], e.tienda).Scan(new(int64))
	var pg *pgconn.PgError
	if !errors.As(err, &pg) || pg.Code != "42501" {
		t.Fatalf("la función dejó cambiar el plan a soporte: %v", err)
	}
	// Y el rol staff no escribe en las tiendas directo.
	if _, err := e.b.Staff.Exec(ctx, `UPDATE tiendas SET plan = 'negocio'`); !errors.As(err, &pg) || pg.Code != "42501" {
		t.Fatalf("staff actualizó tiendas: %v", err)
	}
}

func TestFichaMiradaYNotas(t *testing.T) {
	e := armar(t)
	ctx := context.Background()
	f := leer[map[string]any](t, pedir(e.h, "GET", "/v1/equipo/tiendas/"+e.tienda, "", e.tok["soporte"]), 200)
	if f["tienda"].(map[string]any)["estado"] != "prueba" || f["titular"].(map[string]any)["email"] != "lucia@laesquina.com.ar" {
		t.Fatalf("ficha: %v", f)
	}
	estado(t, pedir(e.h, "GET", "/v1/equipo/tiendas/00000000-0000-0000-0000-000000000000", "", e.tok["soporte"]), 404)
	estado(t, pedir(e.h, "POST", "/v1/equipo/tiendas/"+e.tienda+"/notas", `{"texto":"Llamó por el dominio"}`, e.tok["soporte"]), 201)

	ses := leer[equipo.SesionNueva](t, pedir(e.h, "POST", "/v1/equipo/tiendas/"+e.tienda+"/mirar", "", e.tok["soporte"]), 200)
	// La mirada lee el panel de esa tienda…
	d := leer[map[string]any](t, pedir(e.h, "GET", "/v1/auth/sesion", "", ses.Token), 200)
	if d["tiendaId"] != e.tienda || d["rol"] != "duenio" || !strings.HasSuffix(d["usuario"].(map[string]any)["nombre"].(string), "(equipo)") {
		t.Fatalf("mirada: %v", d)
	}
	estado(t, pedir(e.h, "GET", "/v1/panel/cuenta/plan", "", ses.Token), 200)
	// …y no puede escribir nada.
	rec := pedir(e.h, "POST", "/v1/panel/cuenta/pack", `{}`, ses.Token)
	if rec.Code != 403 || !strings.Contains(rec.Body.String(), "modo lectura") {
		t.Fatalf("la mirada escribió: %d %s", rec.Code, rec.Body)
	}
	// No abre el backoffice, y vence.
	estado(t, pedir(e.h, "GET", "/v1/equipo/sesion", "", ses.Token), 401)
	_, _ = e.admin.Exec(ctx, `UPDATE equipo.miradas SET vence_en = now() - interval '1 second'`)
	estado(t, pedir(e.h, "GET", "/v1/panel/cuenta/plan", "", ses.Token), 401)

	a := e.acciones(t)
	for _, x := range []string{"ver_ficha", "nota", "mirar_tienda"} {
		if !slices.Contains(a, x) {
			t.Fatalf("falta %s en la auditoría: %v", x, a)
		}
	}
	var texto string
	_ = e.admin.QueryRow(ctx, `SELECT texto FROM equipo.auditoria WHERE accion = 'mirar_tienda'`).Scan(&texto)
	if texto != "Entró a ver la tienda La Esquina" {
		t.Fatalf("texto: %q", texto)
	}
	f = leer[map[string]any](t, pedir(e.h, "GET", "/v1/equipo/tiendas/"+e.tienda, "", e.tok["soporte"]), 200)
	if n := f["notas"].([]any); len(n) != 1 || n[0].(map[string]any)["autor"] != "Persona soporte" {
		t.Fatalf("notas: %v", n)
	}

	l := leer[map[string]any](t, pedir(e.h, "GET", "/v1/equipo/tiendas?seg=prueba&q=esquina", "", e.tok["soporte"]), 200)
	if l["total"].(float64) != 1 || l["filas"].([]any)[0].(map[string]any)["titular"] != "Lucía" {
		t.Fatalf("lista: %v", l)
	}
	in := leer[map[string]any](t, pedir(e.h, "GET", "/v1/equipo/inicio", "", e.tok["soporte"]), 200)
	if in["kpis"].(map[string]any)["tiendasActivas"].(float64) != 1 {
		t.Fatalf("inicio: %v", in)
	}
}

func TestConsultas(t *testing.T) {
	e := armar(t)
	ctx := context.Background()
	var contacto string
	if err := e.admin.QueryRow(ctx, `INSERT INTO soporte.consultas (nombre, email, mensaje) VALUES ('Ana', 'ana@vivero.com.ar', 'Quiero migrar')
		RETURNING id::text`).Scan(&contacto); err != nil {
		t.Fatal(err)
	}
	if _, err := e.admin.Exec(ctx, `INSERT INTO soporte.arrepentimientos (codigo, email, nombre, documento, plan, creado_en)
		VALUES ('ARR-2026-100001', 'beto@mail.com', 'Beto', '30111222', 'Pro mensual', '2026-10-23 12:00-03')`); err != nil {
		t.Fatal(err)
	}
	b := leer[equipo.Bandeja](t, pedir(e.h, "GET", "/v1/equipo/consultas", "", e.tok["soporte"]), 200)
	if len(b.Consultas) != 2 || len(b.Miembros) != 3 {
		t.Fatalf("bandeja: %+v", b)
	}
	for _, c := range b.Consultas {
		if c.Estado != "nueva" {
			t.Fatalf("estado: %+v", c)
		}
		// Viernes 23/10 + 10 días hábiles: viernes 6/11.
		if c.Tipo == "arrepentimiento" && (c.Plazo != "2026-11-06" || c.Codigo != "ARR-2026-100001") {
			t.Fatalf("arrepentimiento: %+v", c)
		}
	}
	estado(t, pedir(e.h, "POST", "/v1/equipo/consultas/contacto/"+contacto+"/responder", `{"asunto":"Hola","texto":""}`, e.tok["soporte"]), 400)
	estado(t, pedir(e.h, "POST", "/v1/equipo/consultas/contacto/"+contacto+"/responder",
		`{"asunto":"Tu migración","texto":"Hola Ana,\nte ayudamos a migrar."}`, e.tok["soporte"]), 201)
	var tipo string
	var datos map[string]string
	var sinTienda bool
	if err := e.admin.QueryRow(ctx, `SELECT tipo, datos, tienda_id IS NULL FROM outbox WHERE tipo LIKE 'equipo.%'`).Scan(&tipo, &datos, &sinTienda); err != nil {
		t.Fatal(err)
	}
	if tipo != "equipo.correo" || !sinTienda || datos["para"] != "ana@vivero.com.ar" || !strings.Contains(datos["html"], "<br>") {
		t.Fatalf("outbox: %s %v %v", tipo, sinTienda, datos)
	}
	estado(t, pedir(e.h, "POST", "/v1/equipo/consultas/arrepentimiento/ARR-2026-100001",
		`{"estado":"resuelta","asignadoA":"`+e.ids["administracion"]+`"}`, e.tok["soporte"]), 204)
	estado(t, pedir(e.h, "POST", "/v1/equipo/consultas/arrepentimiento/ARR-2026-100001", `{"asignadoA":"nadie"}`, e.tok["soporte"]), 400)
	d := leer[equipo.ConsultaDetalle](t, pedir(e.h, "GET", "/v1/equipo/consultas/contacto/"+contacto, "", e.tok["soporte"]), 200)
	if d.Estado != "en_curso" || len(d.Respuestas) != 1 || d.Respuestas[0].Autor != "Persona soporte" {
		t.Fatalf("detalle: %+v", d)
	}
	a := leer[equipo.ConsultaDetalle](t, pedir(e.h, "GET", "/v1/equipo/consultas/arrepentimiento/ARR-2026-100001", "", e.tok["soporte"]), 200)
	if a.Estado != "resuelta" || a.AsignadoA == nil || *a.AsignadoA != e.ids["administracion"] || a.Documento != "30111222" {
		t.Fatalf("arrepentimiento: %+v", a)
	}
	estado(t, pedir(e.h, "GET", "/v1/equipo/consultas/contacto/no-existe", "", e.tok["soporte"]), 404)
	acc := e.acciones(t)
	for _, x := range []string{"responder", "consulta_estado", "consulta_asignar", "ver_consulta"} {
		if !slices.Contains(acc, x) {
			t.Fatalf("falta %s: %v", x, acc)
		}
	}

	// El worker manda el mail encolado.
	mem := &correo.Memoria{}
	var ev outbox.Evento
	_ = e.admin.QueryRow(ctx, `SELECT id, datos FROM outbox WHERE tipo = 'equipo.correo'`).Scan(&ev.ID, &ev.Datos)
	if err := equipo.Manejadores(mem)["equipo.correo"](ctx, ev); err != nil || len(mem.Enviados) != 1 || mem.Enviados[0].Asunto != "Tu migración" {
		t.Fatalf("worker: %v %+v", err, mem.Enviados)
	}
}

func TestPartners(t *testing.T) {
	e := armar(t)
	ctx := context.Background()
	alta := func(email string) string {
		var id string
		if err := e.admin.QueryRow(ctx, `INSERT INTO soporte.postulaciones_partner (nombre, email, agencia, especialidades)
			VALUES ('Ana', $1, 'Pampa Code', '{diseno}') RETURNING id::text`, email).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	uno, dos := alta("ana@pampacode.com.ar"), alta("beto@estudio.com.ar")
	estado(t, pedir(e.h, "POST", "/v1/equipo/partners/"+uno, `{"decision":"aprobar"}`, e.tok["soporte"]), 403)
	estado(t, pedir(e.h, "POST", "/v1/equipo/partners/"+uno, `{"decision":"rechazar"}`, e.tok["administracion"]), 400)
	estado(t, pedir(e.h, "POST", "/v1/equipo/partners/"+uno, `{"decision":"rechazar","motivo":"Todavía no trabajan con tiendas"}`, e.tok["administracion"]), 204)
	estado(t, pedir(e.h, "POST", "/v1/equipo/partners/"+uno, `{"decision":"aprobar"}`, e.tok["administracion"]), 409)

	// Si el alta del partner falla, la postulación sigue nueva.
	e.s.Aprobador = func(context.Context, equipo.PostulacionAprobada) error { return errors.New("sin cuenta") }
	estado(t, pedir(e.h, "POST", "/v1/equipo/partners/"+dos, `{"decision":"aprobar"}`, e.tok["administracion"]), 500)
	var aprobada equipo.PostulacionAprobada
	e.s.Aprobador = func(_ context.Context, p equipo.PostulacionAprobada) error { aprobada = p; return nil }
	estado(t, pedir(e.h, "POST", "/v1/equipo/partners/"+dos, `{"decision":"aprobar"}`, e.tok["administracion"]), 204)
	if aprobada.Email != "beto@estudio.com.ar" || aprobada.Agencia != "Pampa Code" {
		t.Fatalf("aprobador: %+v", aprobada)
	}
	p := leer[equipo.Partners](t, pedir(e.h, "GET", "/v1/equipo/partners", "", e.tok["administracion"]), 200)
	est := map[string]string{}
	for _, x := range p.Postulaciones {
		est[x.Email] = x.Estado
	}
	if est["ana@pampacode.com.ar"] != "rechazada" || est["beto@estudio.com.ar"] != "aprobada" {
		t.Fatalf("postulaciones: %v", est)
	}
	var n int
	_ = e.admin.QueryRow(ctx, `SELECT count(*) FROM outbox WHERE tipo = 'equipo.correo'`).Scan(&n)
	if n != 2 {
		t.Fatalf("mails encolados: %d", n)
	}
	if a := e.acciones(t); !slices.Contains(a, "partner_aprobar") || !slices.Contains(a, "partner_rechazar") {
		t.Fatalf("auditoría: %v", a)
	}
}
