package telegram_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/telegram"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/servidor"
)

const claveBFF = "clave-del-bff-de-prueba"

type api struct {
	t *testing.T
	h http.Handler
}

func (a api) pedir(metodo, ruta, cuerpo, token string, bff bool) *httptest.ResponseRecorder {
	r := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	if bff {
		r.Header.Set("X-Mostrador-BFF", claveBFF)
	}
	rec := httptest.NewRecorder()
	a.h.ServeHTTP(rec, r)
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

func (a api) registrar(email, tienda string) cuentas.Sesion {
	a.t.Helper()
	return leer[cuentas.Sesion](a.t, a.pedir("POST", "/v1/auth/registro",
		`{"nombre":"Lucía","tienda":"`+tienda+`","email":"`+email+`","clave":"Abcdefghijk1"}`, "", false), 201)
}

func (a api) codigo(token string) string {
	a.t.Helper()
	return leer[telegram.Codigo](a.t, a.pedir("POST", "/v1/panel/telegram/codigo", "", token, false), 201).Codigo
}

func (a api) vincular(codigo string, chat string) *httptest.ResponseRecorder {
	return a.pedir("POST", "/v1/telegram/vincular", `{"codigo":"`+codigo+`","chat":`+chat+`,"nombre":"Lucía"}`, "", true)
}

func codigoDe(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var p struct{ Code string }
	_ = json.Unmarshal(rec.Body.Bytes(), &p)
	return p.Code
}

func nuevo(t *testing.T) (api, *pruebadb.Bases) {
	b := pruebadb.Nueva(t)
	h := servidor.Nuevo(servidor.Dependencias{App: b.App, ClaveBFF: claveBFF,
		Cuentas:  &cuentas.Servicio{Pool: b.App, Correo: &correo.Memoria{}, Sitio: "https://tulocalweb.com.ar"},
		Telegram: &telegram.Servicio{Pool: b.App}})
	return api{t, h}, b
}

func TestVincularYDesvincular(t *testing.T) {
	a, b := nuevo(t)
	esquina := a.registrar("lucia@laesquina.com.ar", "La Esquina")
	otra := a.registrar("ana@otra.com.ar", "Otra Tienda")

	e := leer[telegram.Estado](t, a.pedir("GET", "/v1/panel/telegram", "", esquina.Token, false), 200)
	if e.Vinculado {
		t.Fatal("arrancó vinculada")
	}
	cod := a.codigo(esquina.Token)

	// Sin la clave del BFF no se vincula ni se lee nada.
	for _, ruta := range []string{"/v1/telegram/vincular", "/v1/telegram/contexto", "/v1/telegram/desvincular"} {
		if rec := a.pedir("POST", ruta, `{"codigo":"`+cod+`","chat":111}`, "", false); rec.Code != 403 {
			t.Fatalf("%s sin clave: %d", ruta, rec.Code)
		}
	}
	// Ni con una sesión de la tienda: la ruta es del bot.
	if rec := a.pedir("POST", "/v1/telegram/contexto", `{"chat":111}`, esquina.Token, false); rec.Code != 403 {
		t.Fatalf("contexto con sesión: %d", rec.Code)
	}

	// Código válido: vincula y devuelve el nombre de la tienda.
	v := leer[telegram.Vinculado](t, a.vincular(strings.ToLower(cod), "111"), 200)
	if v.Tienda != "La Esquina" {
		t.Fatalf("vinculado a %q", v.Tienda)
	}
	e = leer[telegram.Estado](t, a.pedir("GET", "/v1/panel/telegram", "", esquina.Token, false), 200)
	if !e.Vinculado || e.Nombre != "Lucía" || e.Desde == nil {
		t.Fatalf("estado: %+v", e)
	}
	// La otra tienda no ve el vínculo de La Esquina.
	if e := leer[telegram.Estado](t, a.pedir("GET", "/v1/panel/telegram", "", otra.Token, false), 200); e.Vinculado {
		t.Fatal("la otra tienda ve el vínculo ajeno")
	}

	// Usado: no sirve dos veces.
	if rec := a.vincular(cod, "999"); rec.Code != 409 || codigoDe(t, rec) != "codigo_invalido" {
		t.Fatalf("código usado: %d %s", rec.Code, rec.Body)
	}
	// Inventado: tampoco.
	if rec := a.vincular("ABCDEFGH", "999"); rec.Code != 409 {
		t.Fatalf("código inventado: %d", rec.Code)
	}

	// Vencido: no vincula (y se descarta).
	vencido := a.codigo(otra.Token)
	admin, err := pgx.Connect(context.Background(), b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(context.Background())
	if _, err := admin.Exec(context.Background(), `UPDATE telegram_codigos SET vence_en = now() - interval '1 minute'`); err != nil {
		t.Fatal(err)
	}
	if rec := a.vincular(vencido, "222"); rec.Code != 409 || codigoDe(t, rec) != "codigo_invalido" {
		t.Fatalf("código vencido: %d %s", rec.Code, rec.Body)
	}
	var quedan int
	_ = admin.QueryRow(context.Background(), `SELECT count(*) FROM telegram_codigos`).Scan(&quedan)
	if quedan != 0 {
		t.Fatalf("quedaron %d códigos", quedan)
	}

	// Un chat no puede estar en dos tiendas.
	if rec := a.vincular(a.codigo(otra.Token), "111"); rec.Code != 409 || codigoDe(t, rec) != "chat_ocupado" {
		t.Fatalf("chat en dos tiendas: %d %s", rec.Code, rec.Body)
	}
	// Generar uno nuevo invalida el anterior.
	viejo := a.codigo(otra.Token)
	nuevoCod := a.codigo(otra.Token)
	if rec := a.vincular(viejo, "222"); rec.Code != 409 {
		t.Fatalf("el código anterior sigue sirviendo: %d", rec.Code)
	}
	leer[telegram.Vinculado](t, a.vincular(nuevoCod, "222"), 200)

	// Una tienda, un chat: vincular otro reemplaza al anterior.
	leer[telegram.Vinculado](t, a.vincular(a.codigo(esquina.Token), "333"), 200)
	if rec := a.pedir("POST", "/v1/telegram/contexto", `{"chat":111}`, "", true); rec.Code != 404 {
		t.Fatalf("el chat reemplazado sigue vinculado: %d", rec.Code)
	}

	// Aislamiento: cada chat ve sólo su tienda, y los mensajes se cuentan por día.
	c := leer[telegram.Contexto](t, a.pedir("POST", "/v1/telegram/contexto", `{"chat":333}`, "", true), 200)
	if c.Tienda.Nombre != "La Esquina" || c.Usos != 1 || c.Hoy.Periodo != "hoy" || c.Semana.Periodo != "7d" || c.Contexto == nil {
		t.Fatalf("contexto: %+v", c)
	}
	c = leer[telegram.Contexto](t, a.pedir("POST", "/v1/telegram/contexto", `{"chat":333}`, "", true), 200)
	if c.Usos != 2 {
		t.Fatalf("usos: %d", c.Usos)
	}
	c = leer[telegram.Contexto](t, a.pedir("POST", "/v1/telegram/contexto", `{"chat":222}`, "", true), 200)
	if c.Tienda.Nombre != "Otra Tienda" || c.Usos != 1 {
		t.Fatalf("contexto de la otra: %+v", c)
	}

	// /desvincular desde el chat.
	if rec := a.pedir("POST", "/v1/telegram/desvincular", `{"chat":222}`, "", true); rec.Code != 204 {
		t.Fatalf("desvincular chat: %d", rec.Code)
	}
	if rec := a.pedir("POST", "/v1/telegram/desvincular", `{"chat":222}`, "", true); rec.Code != 404 {
		t.Fatalf("desvincular dos veces: %d", rec.Code)
	}
	if e := leer[telegram.Estado](t, a.pedir("GET", "/v1/panel/telegram", "", otra.Token, false), 200); e.Vinculado {
		t.Fatal("sigue vinculada después de /desvincular")
	}
	// Desde el panel: devuelve el chat para avisarle; la segunda vez no hay nada.
	des := leer[map[string]int64](t, a.pedir("DELETE", "/v1/panel/telegram", "", esquina.Token, false), 200)
	if des["chat"] != 333 {
		t.Fatalf("desvincular panel: %v", des)
	}
	if rec := a.pedir("DELETE", "/v1/panel/telegram", "", esquina.Token, false); rec.Code != 204 {
		t.Fatalf("desvincular panel dos veces: %d", rec.Code)
	}
}

// telegramFalso es la Bot API de mentira: guarda lo que le mandan.
type telegramFalso struct {
	mu       sync.Mutex
	rutas    []string
	mensajes []struct {
		Chat  int64  `json:"chat_id"`
		Texto string `json:"text"`
	}
	estado int
}

func (f *telegramFalso) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	cuerpo, _ := io.ReadAll(r.Body)
	var m struct {
		Chat  int64  `json:"chat_id"`
		Texto string `json:"text"`
	}
	_ = json.Unmarshal(cuerpo, &m)
	f.rutas = append(f.rutas, r.URL.Path)
	if f.estado != 0 {
		w.WriteHeader(f.estado)
		return
	}
	f.mensajes = append(f.mensajes, m)
	_, _ = w.Write([]byte(`{"ok":true}`))
}

func TestAvisoDeVenta(t *testing.T) {
	ctx := context.Background()
	a, b := nuevo(t)
	esquina := a.registrar("lucia@laesquina.com.ar", "La Esquina")
	otra := a.registrar("ana@otra.com.ar", "Otra Tienda")
	falso := &telegramFalso{}
	srv := httptest.NewServer(falso)
	defer srv.Close()
	tg := &telegram.Servicio{Pool: b.App, Bot: &telegram.Bot{Token: "123:abc", Base: srv.URL}}
	w := &outbox.Worker{Pool: b.Worker, Manejadores: tg.Manejadores()}
	vender := func(tienda string, canal string) int {
		n, err := pedidos.Crear(ctx, b.App, tienda, pedidos.Nuevo{Comprador: pedidos.Comprador{Nombre: "Juan"},
			Canal: canal, Pago: "efectivo", Envio: "retiro", Items: []pedidos.Item{{Nombre: "Mate", Cantidad: 2, Precio: 22300}}})
		if err != nil {
			t.Fatal(err)
		}
		return n
	}

	// Sin chat vinculado no se encola nada.
	vender(esquina.TiendaID, "pos")
	if n, _ := w.Pasada(ctx); n != 0 {
		t.Fatalf("procesó %d eventos sin chat vinculado", n)
	}

	leer[telegram.Vinculado](t, a.vincular(a.codigo(esquina.Token), "4242"), 200)
	n := vender(esquina.TiendaID, "pos")
	vender(otra.TiendaID, "web") // la otra tienda no tiene chat: a 4242 no le llega
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	if len(falso.mensajes) != 1 || falso.mensajes[0].Chat != 4242 {
		t.Fatalf("mensajes: %+v", falso.mensajes)
	}
	if want := "🛒 Nueva venta en el local: pedido #1002 por $ 44.600."; falso.mensajes[0].Texto != want || n != 1002 {
		t.Fatalf("texto %q (pedido %d), esperaba %q", falso.mensajes[0].Texto, n, want)
	}
	if falso.rutas[0] != "/bot123:abc/sendMessage" {
		t.Fatalf("ruta %q", falso.rutas[0])
	}

	// Al menos una vez del outbox, una sola vez el aviso.
	if err := b.App.EnTienda(ctx, esquina.TiendaID, func(tx pgx.Tx) error {
		return outbox.Publicar(ctx, tx, esquina.TiendaID, "telegram.venta", map[string]any{"numero": n})
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	if len(falso.mensajes) != 1 {
		t.Fatalf("se avisó dos veces: %d", len(falso.mensajes))
	}

	// El bot bloqueado (403) no se reintenta para siempre.
	falso.estado = http.StatusForbidden
	vender(esquina.TiendaID, "web")
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	var pendientes int
	_ = b.App.EnTienda(ctx, esquina.TiendaID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT count(*) FROM outbox WHERE publicado_en IS NULL`).Scan(&pendientes)
	})
	if pendientes != 0 {
		t.Fatalf("quedaron %d eventos pendientes", pendientes)
	}

	// Sin token, el manejador no manda nada.
	falso.estado = 0
	sinToken := &telegram.Servicio{Pool: b.App, Bot: &telegram.Bot{Base: srv.URL}}
	w2 := &outbox.Worker{Pool: b.Worker, Manejadores: sinToken.Manejadores()}
	vender(esquina.TiendaID, "wa")
	if _, err := w2.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	if len(falso.mensajes) != 1 {
		t.Fatalf("sin token igual mandó: %d", len(falso.mensajes))
	}
}

func TestTextoVenta(t *testing.T) {
	if got := telegram.TextoVenta("ml", 7, 1234567); got != "🛒 Nueva venta en Mercado Libre: pedido #7 por $ 1.234.567." {
		t.Fatal(got)
	}
	if got := telegram.TextoVenta("otro", 7, 500); got != "🛒 Nueva venta: pedido #7 por $ 500." {
		t.Fatal(got)
	}
}
