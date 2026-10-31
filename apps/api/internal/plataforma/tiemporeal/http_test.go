package tiemporeal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
)

// sesiones falsas: token → quién.
type sesiones struct {
	mu sync.Mutex
	m  map[string]Quien
}

func (s *sesiones) autenticar(_ context.Context, token string) (Quien, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	q, ok := s.m[token]
	if !ok {
		return Quien{}, ErrNoExiste
	}
	return q, nil
}

func (s *sesiones) poner(token string, q *Quien) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if q == nil {
		delete(s.m, token)
		return
	}
	s.m[token] = *q
}

const sitio = "https://tulocalweb.com.ar"

func servidorDePrueba(t *testing.T, ajustar func(*Config)) (*httptest.Server, *Hub, *sesiones) {
	t.Helper()
	h := NuevoHub()
	h.Ventana = 20 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go h.Correr(ctx)
	ses := &sesiones{m: map[string]Quien{
		"ana":  {Usuario: "u-ana", Sesion: "s-ana", Tienda: "sol", Rol: "duenio", VenceEn: time.Now().Add(time.Hour)},
		"dani": {Usuario: "u-dani", Sesion: "s-dani", Tienda: "sol", Rol: "local", VenceEn: time.Now().Add(time.Hour)},
		"eva":  {Usuario: "u-eva", Sesion: "s-eva", Tienda: "luna", Rol: "duenio", VenceEn: time.Now().Add(time.Hour)},
		"ia":   {Usuario: "u-ana", Sesion: "s-ia", Tienda: "sol", Rol: "duenio", IA: true, VenceEn: time.Now().Add(time.Hour)},
		"sin":  {Usuario: "u-sin", Sesion: "s-sin", VenceEn: time.Now().Add(time.Hour)},
	}}
	c := Config{
		Hub: h, Autenticar: ses.autenticar, Origenes: []string{sitio},
		Publico: func(_ context.Context, slug, codigo string) (string, string, error) {
			if slug != "sol" {
				return "", "", ErrNoExiste
			}
			if codigo == "" {
				return "sol", "", nil
			}
			if codigo == "codigo-del-pedido-12" {
				return "sol", "12", nil
			}
			return "", "", ErrNoExiste
		},
		SlugDeDominio: func(_ context.Context, d string) (string, error) {
			if d == "almacensol.com.ar" {
				return "sol", nil
			}
			return "", ErrNoExiste
		},
	}
	if ajustar != nil {
		ajustar(&c)
	}
	mux := http.NewServeMux()
	Rutas(mux, c)
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, h, ses
}

func wsURL(srv *httptest.Server, ruta string) string {
	return "ws" + strings.TrimPrefix(srv.URL, "http") + ruta
}

type conexion struct {
	ws  *websocket.Conn
	msj chan Mensaje
	fin chan websocket.StatusCode
}

// conectar abre la conexión y lee en segundo plano (así también contesta los pings).
func conectar(t *testing.T, url string, h http.Header) (*conexion, *http.Response, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ws, res, err := websocket.Dial(ctx, url, &websocket.DialOptions{HTTPHeader: h})
	if err != nil {
		return nil, res, err
	}
	c := &conexion{ws: ws, msj: make(chan Mensaje, 100), fin: make(chan websocket.StatusCode, 1)}
	go func() {
		for {
			_, b, err := ws.Read(context.Background())
			if err != nil {
				c.fin <- websocket.CloseStatus(err)
				return
			}
			var m Mensaje
			_ = json.Unmarshal(b, &m)
			c.msj <- m
		}
	}()
	t.Cleanup(func() { ws.CloseNow() })
	return c, res, nil
}

func conCookie(token, origen string) http.Header {
	h := http.Header{}
	h.Set("Cookie", "__Host-tlw_sesion="+token)
	if origen != "" {
		h.Set("Origin", origen)
	}
	return h
}

func (c *conexion) esperar(t *testing.T, tipo string) Mensaje {
	t.Helper()
	for {
		select {
		case m := <-c.msj:
			if m.Tipo == tipo {
				return m
			}
		case code := <-c.fin:
			t.Fatalf("se cerró (%d) esperando %s", code, tipo)
		case <-time.After(3 * time.Second):
			t.Fatalf("no llegó %s", tipo)
		}
	}
}

func (c *conexion) nada(t *testing.T, espera time.Duration) {
	t.Helper()
	select {
	case m := <-c.msj:
		t.Fatalf("no tenía que llegar nada y llegó %+v", m)
	case <-time.After(espera):
	}
}

func (c *conexion) cierre(t *testing.T) websocket.StatusCode {
	t.Helper()
	for {
		select {
		case <-c.msj:
		case code := <-c.fin:
			return code
		case <-time.After(3 * time.Second):
			t.Fatal("no se cerró")
		}
	}
}

func TestAutenticacionYOrigen(t *testing.T) {
	srv, _, _ := servidorDePrueba(t, nil)
	url := wsURL(srv, "/v1/tiempo-real")

	// Con la cookie, sin Origin o con otro Origin: 403 antes del upgrade (CSWSH).
	for _, origen := range []string{"", "https://malo.com", "https://tulocalweb.com.ar.malo.com"} {
		_, res, err := conectar(t, url, conCookie("ana", origen))
		if err == nil || res == nil || res.StatusCode != http.StatusForbidden {
			t.Fatalf("origen %q: tendría que responder 403 (%v)", origen, err)
		}
	}
	// Sin sesión o con una que no existe: se cierra con 4401 (el navegador no reintenta).
	for _, h := range []http.Header{conCookie("", sitio), conCookie("nadie", sitio), conCookie("sin", sitio)} {
		c, _, err := conectar(t, url, h)
		if err != nil {
			t.Fatal(err)
		}
		if code := c.cierre(t); code != CierreSinSesion {
			t.Fatalf("sin sesión: %d", code)
		}
	}
	// El token de una IA conectada no usa el tiempo real.
	c, _, err := conectar(t, url, conCookie("ia", sitio))
	if err != nil {
		t.Fatal(err)
	}
	if code := c.cierre(t); code != CierreSinPermiso {
		t.Fatalf("ia: %d", code)
	}
	// Con cookie y el Origin del panel: entra y recibe «hola».
	c, _, err = conectar(t, url, conCookie("ana", sitio))
	if err != nil {
		t.Fatal(err)
	}
	c.esperar(t, TipoHola)
	// Con Bearer (la app de iPhone, sin Origin): entra.
	h := http.Header{}
	h.Set("Authorization", "Bearer dani")
	c, _, err = conectar(t, url, h)
	if err != nil {
		t.Fatal(err)
	}
	c.esperar(t, TipoHola)
	// En producción la cookie sin __Host- no vale (la podría poner un subdominio).
	h = http.Header{}
	h.Set("Cookie", "tlw_sesion=ana")
	h.Set("Origin", sitio)
	c, _, err = conectar(t, url, h)
	if err != nil {
		t.Fatal(err)
	}
	if code := c.cierre(t); code != CierreSinSesion {
		t.Fatalf("cookie sin prefijo en producción: %d", code)
	}
}

func TestEntregaPorTiendaYRolDePuntaAPunta(t *testing.T) {
	srv, h, _ := servidorDePrueba(t, nil)
	url := wsURL(srv, "/v1/tiempo-real")
	ana, _, _ := conectar(t, url, conCookie("ana", sitio))
	dani, _, _ := conectar(t, url, conCookie("dani", sitio))
	eva, _, _ := conectar(t, url, conCookie("eva", sitio))
	for _, c := range []*conexion{ana, dani, eva} {
		c.esperar(t, TipoHola)
	}
	h.Publicar(Evento{Tienda: "sol", Tipo: "finanzas.cambio"})
	h.Publicar(Evento{Tienda: "sol", Tipo: "pedido.nuevo", ID: "12"})
	vistos := map[string]Mensaje{}
	for len(vistos) < 2 {
		select {
		case m := <-ana.msj:
			vistos[m.Tipo] = m
		case <-time.After(3 * time.Second):
			t.Fatalf("a la dueña le llegó %v", vistos)
		}
	}
	if vistos["pedido.nuevo"].ID != "12" || vistos["finanzas.cambio"].Tipo == "" {
		t.Fatalf("dueña: %v", vistos)
	}
	dani.esperar(t, "pedido.nuevo")
	dani.nada(t, 150*time.Millisecond) // Local no ve finanzas
	eva.nada(t, 50*time.Millisecond)   // otra tienda
}

func TestPingMantieneViva(t *testing.T) {
	var pings atomic.Int32
	srv, _, _ := servidorDePrueba(t, func(c *Config) { c.Ping = 40 * time.Millisecond })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ws, _, err := websocket.Dial(ctx, wsURL(srv, "/v1/tiempo-real"), &websocket.DialOptions{
		HTTPHeader: conCookie("ana", sitio),
		OnPingReceived: func(context.Context, []byte) bool {
			pings.Add(1)
			return true
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer ws.CloseNow()
	go func() {
		for {
			if _, _, err := ws.Read(context.Background()); err != nil {
				return
			}
		}
	}()
	time.Sleep(300 * time.Millisecond)
	if pings.Load() < 3 {
		t.Fatalf("el servidor hace ping seguido: %d", pings.Load())
	}
}

func TestSinPongSeCorta(t *testing.T) {
	srv, h, _ := servidorDePrueba(t, func(c *Config) { c.Ping = 40 * time.Millisecond })
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// Un cliente que no lee nunca contesta los pings: el servidor lo suelta.
	ws, _, err := websocket.Dial(ctx, wsURL(srv, "/v1/tiempo-real"), &websocket.DialOptions{HTTPHeader: conCookie("ana", sitio)})
	if err != nil {
		t.Fatal(err)
	}
	defer ws.CloseNow()
	deadline := time.Now().Add(15 * time.Second)
	for h.Conexiones() != 0 {
		if time.Now().After(deadline) {
			t.Fatal("una conexión que no responde los pings no se suelta")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestCorteAlCerrarseLaSesion(t *testing.T) {
	srv, _, ses := servidorDePrueba(t, func(c *Config) { c.Revalidar = 50 * time.Millisecond })
	url := wsURL(srv, "/v1/tiempo-real")
	ana, _, _ := conectar(t, url, conCookie("ana", sitio))
	ana.esperar(t, TipoHola)
	// Se cerró la sesión por un camino que no avisó (p. ej. inactiva): la revalidación lo ve.
	ses.poner("ana", nil)
	if code := ana.cierre(t); code != CierreSinSesion {
		t.Fatalf("sesión cerrada: %d", code)
	}
	// Le cambiaron el rol: resincronizar.
	dani, _, _ := conectar(t, url, conCookie("dani", sitio))
	dani.esperar(t, TipoHola)
	ses.poner("dani", &Quien{Usuario: "u-dani", Sesion: "s-dani", Tienda: "sol", Rol: "ventas", VenceEn: time.Now().Add(time.Hour)})
	if code := dani.cierre(t); code != CierreResincronizar {
		t.Fatalf("cambio de rol: %d", code)
	}
}

func TestCorteAlVencer(t *testing.T) {
	srv, _, ses := servidorDePrueba(t, nil)
	ses.poner("corta", &Quien{Usuario: "u", Sesion: "s", Tienda: "sol", Rol: "duenio", VenceEn: time.Now().Add(200 * time.Millisecond)})
	c, _, _ := conectar(t, wsURL(srv, "/v1/tiempo-real"), conCookie("corta", sitio))
	c.esperar(t, TipoHola)
	if code := c.cierre(t); code != CierreSinSesion {
		t.Fatalf("al vencer: %d", code)
	}
}

func TestCorteDesdeElHub(t *testing.T) {
	srv, h, _ := servidorDePrueba(t, nil)
	c, _, _ := conectar(t, wsURL(srv, "/v1/tiempo-real"), conCookie("ana", sitio))
	c.esperar(t, TipoHola)
	h.Publicar(Evento{Tienda: "sol", Tipo: "sesion.cerrada", Usuario: "u-ana", Sesion: "s-ana"})
	if code := c.cierre(t); code != CierreSinSesion {
		t.Fatalf("sesión cerrada en la base: %d", code)
	}
	deadline := time.Now().Add(3 * time.Second)
	for h.Conexiones() != 0 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if h.Conexiones() != 0 {
		t.Fatal("al cerrarse se saca del hub")
	}
}

func TestMensajeEntranteCorta(t *testing.T) {
	srv, _, _ := servidorDePrueba(t, nil)
	c, _, _ := conectar(t, wsURL(srv, "/v1/tiempo-real"), conCookie("ana", sitio))
	c.esperar(t, TipoHola)
	_ = c.ws.Write(context.Background(), websocket.MessageText, []byte(strings.Repeat("x", 2000)))
	if code := c.cierre(t); code != websocket.StatusPolicyViolation && code != websocket.StatusMessageTooBig {
		t.Fatalf("el navegador no manda datos: %d", code)
	}
}

func TestLimitePorSesion(t *testing.T) {
	srv, h, _ := servidorDePrueba(t, nil)
	h.MaxPorSesion = 2
	url := wsURL(srv, "/v1/tiempo-real")
	for range 2 {
		c, _, _ := conectar(t, url, conCookie("ana", sitio))
		c.esperar(t, TipoHola)
	}
	c, _, _ := conectar(t, url, conCookie("ana", sitio))
	if code := c.cierre(t); code != CierreLimite {
		t.Fatalf("tope por sesión: %d", code)
	}
}

func TestPublico(t *testing.T) {
	srv, h, _ := servidorDePrueba(t, nil)
	base := wsURL(srv, "/v1/tiempo-real/publico")
	origen := func(o string) http.Header {
		h := http.Header{}
		h.Set("Origin", o)
		return h
	}
	// Otro origen: 403. El dominio propio de esa misma tienda, sí.
	if _, res, err := conectar(t, base+"?tienda=sol", origen("https://malo.com")); err == nil || res.StatusCode != http.StatusForbidden {
		t.Fatal("origen ajeno")
	}
	if _, res, err := conectar(t, base+"?tienda=luna", origen("https://almacensol.com.ar")); err == nil || res.StatusCode != http.StatusForbidden {
		t.Fatal("el dominio de una tienda no abre otra")
	}
	vidriera, _, err := conectar(t, base+"?tienda=sol", origen("https://almacensol.com.ar"))
	if err != nil {
		t.Fatal(err)
	}
	vidriera.esperar(t, TipoHola)
	comprador, _, _ := conectar(t, base+"?tienda=sol&pedido=codigo-del-pedido-12", origen(sitio))
	comprador.esperar(t, TipoHola)
	// Un código que no es de la tienda (o no existe): se cierra sin decir más.
	malo, _, _ := conectar(t, base+"?tienda=sol&pedido=otro-codigo-cualquiera", origen(sitio))
	if code := malo.cierre(t); code != websocket.StatusPolicyViolation {
		t.Fatalf("código ajeno: %d", code)
	}
	h.Publicar(Evento{Tienda: "sol", Tipo: "pago.acreditado", ID: "12"})
	h.Publicar(Evento{Tienda: "sol", Tipo: "pedido.nuevo", ID: "13"})
	h.Publicar(Evento{Tienda: "sol", Tipo: "cliente.nuevo", ID: "c"})
	if m := comprador.esperar(t, "pago.acreditado"); m.ID != "" {
		t.Fatal("sin el número")
	}
	vidriera.nada(t, 150*time.Millisecond)
	h.Publicar(Evento{Tienda: "sol", Tipo: "producto.stock", ID: "p1"})
	if m := vidriera.esperar(t, "producto.stock"); m.ID != "p1" {
		t.Fatalf("stock: %+v", m)
	}
}

func TestOrigenPermitido(t *testing.T) {
	c := Config{Origenes: []string{sitio, "http://localhost:*"}}
	for o, quiere := range map[string]bool{
		sitio: true, "https://TULOCALWEB.com.ar": true, "http://tulocalweb.com.ar": false,
		"http://localhost:5173": true, "http://localhost": false, "https://a.tulocalweb.com.ar": false,
		"null": false, "": false,
	} {
		if got := c.origenPermitido(o); got != quiere {
			t.Errorf("%q: %v", o, got)
		}
	}
}
