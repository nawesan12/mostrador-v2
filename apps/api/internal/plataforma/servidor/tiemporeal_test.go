package servidor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/tiemporeal"
)

type ws struct {
	c   *websocket.Conn
	msj chan tiemporeal.Mensaje
	fin chan websocket.StatusCode
}

func abrirWS(t *testing.T, srv *httptest.Server, token string) *ws {
	t.Helper()
	h := http.Header{}
	h.Set("Cookie", "__Host-tlw_sesion="+token)
	h.Set("Origin", "https://tulocalweb.test")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, _, err := websocket.Dial(ctx, "ws"+strings.TrimPrefix(srv.URL, "http")+"/v1/tiempo-real", &websocket.DialOptions{HTTPHeader: h})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.CloseNow() })
	w := &ws{c: c, msj: make(chan tiemporeal.Mensaje, 100), fin: make(chan websocket.StatusCode, 1)}
	go func() {
		for {
			_, b, err := c.Read(context.Background())
			if err != nil {
				w.fin <- websocket.CloseStatus(err)
				return
			}
			var m tiemporeal.Mensaje
			_ = json.Unmarshal(b, &m)
			w.msj <- m
		}
	}()
	select {
	case m := <-w.msj:
		if m.Tipo != tiemporeal.TipoHola {
			t.Fatalf("primero llega hola: %+v", m)
		}
	case code := <-w.fin:
		t.Fatalf("no entró: %d", code)
	case <-time.After(3 * time.Second):
		t.Fatal("no llegó hola")
	}
	return w
}

func (w *ws) junta(espera time.Duration) map[string]bool {
	got := map[string]bool{}
	for {
		select {
		case m := <-w.msj:
			got[m.Tipo] = true
		case <-time.After(espera):
			return got
		}
	}
}

func (w *ws) cierre(t *testing.T) websocket.StatusCode {
	t.Helper()
	for {
		select {
		case <-w.msj:
		case code := <-w.fin:
			return code
		case <-time.After(3 * time.Second):
			t.Fatal("no se cerró")
		}
	}
}

// El tiempo real con la API entera: sesiones de verdad por cookie, los roles de la base, las rutas
// reales escribiendo y los triggers avisando (docs/decisiones.md §42).
func TestTiempoRealConLaAPI(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	b := pruebadb.Nueva(t)
	hub := tiemporeal.NuevoHub()
	hub.Ventana = 20 * time.Millisecond
	go hub.Correr(ctx)
	esc := &tiemporeal.Escucha{URL: b.AdminURL, Hub: hub}
	go esc.Correr(ctx)
	cs := &cuentas.Servicio{Pool: b.App, Correo: &correo.Memoria{}, Sitio: "https://tulocalweb.test"}
	srv := httptest.NewServer(Nuevo(Dependencias{App: b.App, Cuentas: cs, Sitio: "https://tulocalweb.test", Log: silencio(), TiempoReal: hub}))
	defer srv.Close()
	for !esc.Conectada() {
		time.Sleep(10 * time.Millisecond)
	}

	cl := cuentas.Cliente{IP: "190.210.32.1"}
	ana, err := cs.Registrar(ctx, cuentas.Alta{Nombre: "Ana", Tienda: "Sol", Email: "ana@sol.com", Clave: "clave-larga-1"}, cl)
	if err != nil {
		t.Fatal(err)
	}
	eva, err := cs.Registrar(ctx, cuentas.Alta{Nombre: "Eva", Tienda: "Luna", Email: "eva@luna.com", Clave: "clave-larga-1"}, cl)
	if err != nil {
		t.Fatal(err)
	}
	var anaID string
	_ = b.App.QueryRow(ctx, `SELECT id::text FROM auth.usuarios WHERE email = 'ana@sol.com'`).Scan(&anaID)
	tokens := map[string]string{}
	ids := map[string]string{}
	for _, rol := range []string{"ventas", "local"} {
		email := rol + "@sol.com"
		if _, err := configuracion.Invitar(ctx, b.App, cs.Sitio, ana.TiendaID, configuracion.Quien{UsuarioID: anaID, Rol: "duenio"}, email, rol); err != nil {
			t.Fatal(err)
		}
		var raw []byte
		if err := b.Worker.QueryRow(ctx, `SELECT datos FROM outbox WHERE tipo = 'cuenta.invitacion' AND datos->>'para' = $1`, email).Scan(&raw); err != nil {
			t.Fatal(err)
		}
		var m cuentas.MailInvitacion
		_ = json.Unmarshal(raw, &m)
		u, _ := url.Parse(m.Link)
		ses, err := cs.RegistrarConInvitacion(ctx, u.Query().Get("token"), "Persona "+rol, "clave-larga-2", cl)
		if err != nil {
			t.Fatal(err)
		}
		tokens[rol] = ses.Token
		var id string
		_ = b.App.QueryRow(ctx, `SELECT id::text FROM auth.usuarios WHERE email = $1`, email).Scan(&id)
		ids[rol] = id
	}
	pedir := func(token, metodo, ruta, cuerpo string) *http.Response {
		r, _ := http.NewRequest(metodo, srv.URL+ruta, strings.NewReader(cuerpo))
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("Content-Type", "application/json")
		res, err := http.DefaultClient.Do(r)
		if err != nil {
			t.Fatal(err)
		}
		_ = res.Body.Close()
		return res
	}

	wDuenia := abrirWS(t, srv, ana.Token)
	wVentas := abrirWS(t, srv, tokens["ventas"])
	wLocal := abrirWS(t, srv, tokens["local"])
	wLuna := abrirWS(t, srv, eva.Token)
	for _, w := range []*ws{wDuenia, wVentas, wLocal, wLuna} {
		w.junta(150 * time.Millisecond)
	}

	// Un producto nuevo por la API: a todo el equipo de Sol (Pedidos ve productos), a Luna nada.
	if res := pedir(ana.Token, "POST", "/v1/panel/productos", `{"nombre":"Mate","precio":1000000,"costo":400000,"stock":3,"canales":{}}`); res.StatusCode != http.StatusCreated {
		t.Fatalf("crear producto: %d", res.StatusCode)
	}
	for nombre, w := range map[string]*ws{"duenia": wDuenia, "ventas": wVentas, "local": wLocal} {
		if got := w.junta(400 * time.Millisecond); !got["producto.nuevo"] {
			t.Errorf("%s no recibió producto.nuevo: %v", nombre, got)
		}
	}
	if got := wLuna.junta(50 * time.Millisecond); len(got) != 0 {
		t.Fatalf("otra tienda no recibe nada: %v", got)
	}

	// Un cupón: sólo quien tiene Catálogo (la dueña), no Ventas ni Local.
	if res := pedir(ana.Token, "POST", "/v1/panel/cupones", `{"codigo":"HOLA10","tipo":"pct","valor":10,"alcance":"todo","quien":"todos","minimo":0,"limite":0,"porCliente":"libre","vigencia":"sin","canales":["web"],"opciones":{"hornero":false,"combinable":false,"unicos":false,"link":false}}`); res.StatusCode >= 300 {
		t.Fatalf("crear cupón: %d", res.StatusCode)
	}
	if got := wDuenia.junta(400 * time.Millisecond); !got["cupon.nuevo"] {
		t.Fatalf("la dueña ve el cupón: %v", got)
	}
	if got := wVentas.junta(50 * time.Millisecond); got["cupon.nuevo"] {
		t.Fatal("ventas no ve cupones")
	}
	if got := wLocal.junta(10 * time.Millisecond); got["cupon.nuevo"] {
		t.Fatal("local no ve cupones")
	}

	// La dueña le cambia el rol a Ventas: a esa persona se le refresca el panel (4000).
	if res := pedir(ana.Token, "PUT", "/v1/panel/equipo/"+ids["ventas"], `{"rol":"admin"}`); res.StatusCode >= 300 {
		t.Fatalf("cambiar rol: %d", res.StatusCode)
	}
	if code := wVentas.cierre(t); code != tiemporeal.CierreResincronizar {
		t.Fatalf("cambio de rol: %d", code)
	}
	// Al reconectar entra con el rol nuevo: ahora ve los cupones.
	wAdmin := abrirWS(t, srv, tokens["ventas"])
	if res := pedir(ana.Token, "POST", "/v1/panel/cupones", `{"codigo":"CHAU20","tipo":"pct","valor":20,"alcance":"todo","quien":"todos","minimo":0,"limite":0,"porCliente":"libre","vigencia":"sin","canales":["web"],"opciones":{"hornero":false,"combinable":false,"unicos":false,"link":false}}`); res.StatusCode >= 300 {
		t.Fatalf("crear cupón: %d", res.StatusCode)
	}
	if got := wAdmin.junta(400 * time.Millisecond); !got["cupon.nuevo"] {
		t.Fatalf("con el rol nuevo ve cupones: %v", got)
	}

	// Quitar a Local del equipo cierra sus sesiones en la tienda: su conexión se corta.
	if res := pedir(ana.Token, "DELETE", "/v1/panel/equipo/"+ids["local"], ""); res.StatusCode >= 300 {
		t.Fatalf("quitar: %d", res.StatusCode)
	}
	if code := wLocal.cierre(t); code != tiemporeal.CierreSinSesion && code != tiemporeal.CierreResincronizar {
		t.Fatalf("quitada del equipo: %d", code)
	}

	// Salir (la sesión se borra): la conexión de esa sesión se corta con 4401.
	if res := pedir(eva.Token, "POST", "/v1/auth/salir", ""); res.StatusCode != http.StatusNoContent {
		t.Fatalf("salir: %d", res.StatusCode)
	}
	if code := wLuna.cierre(t); code != tiemporeal.CierreSinSesion {
		t.Fatalf("salir: %d", code)
	}
}
