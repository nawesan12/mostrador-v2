package tiemporeal

import (
	"errors"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func hubDePrueba() *Hub {
	h := NuevoHub()
	h.Buffer = 32
	h.MaxPorTipo = 5
	return h
}

func registrar(t *testing.T, h *Hub, c *Cliente) *Cliente {
	t.Helper()
	if err := h.Registrar(c); err != nil {
		t.Fatal(err)
	}
	return c
}

// recibidos vacía la cola del cliente.
func recibidos(c *Cliente) []Mensaje {
	var ms []Mensaje
	for {
		select {
		case m := <-c.envio:
			ms = append(ms, m)
		default:
			return ms
		}
	}
}

func corte(c *Cliente) (Corte, bool) {
	select {
	case k := <-c.corte:
		return k, true
	default:
		return Corte{}, false
	}
}

func tipos(ms []Mensaje) []string {
	var s []string
	for _, m := range ms {
		s = append(s, m.Tipo)
	}
	return s
}

func TestLeer(t *testing.T) {
	e, ok := Leer(`{"t":"a","k":"pedido.nuevo","id":"12"}`)
	if !ok || e.Tienda != "a" || e.Tipo != "pedido.nuevo" || e.ID != "12" {
		t.Fatalf("leer: %+v %v", e, ok)
	}
	for _, malo := range []string{``, `{`, `{"t":"a"}`, `{"k":""}`} {
		if _, ok := Leer(malo); ok {
			t.Errorf("%q no tendría que leerse", malo)
		}
	}
}

func TestFiltroPorRolYTienda(t *testing.T) {
	h := hubDePrueba()
	duenia := registrar(t, h, &Cliente{Tienda: "sol", Usuario: "ana", Sesion: "s1", Rol: "duenio"})
	admin := registrar(t, h, &Cliente{Tienda: "sol", Usuario: "beto", Sesion: "s2", Rol: "admin"})
	ventas := registrar(t, h, &Cliente{Tienda: "sol", Usuario: "caro", Sesion: "s3", Rol: "ventas"})
	local := registrar(t, h, &Cliente{Tienda: "sol", Usuario: "dani", Sesion: "s4", Rol: "local"})
	otra := registrar(t, h, &Cliente{Tienda: "luna", Usuario: "eva", Sesion: "s5", Rol: "duenio"})

	for _, tipo := range []string{"pedido.nuevo", "pago.acreditado", "producto.stock", "cliente.nuevo", "cupon.cambio",
		"comprobante.cambio", "finanzas.cambio", "actividad.nuevo", "plan.cambio", "equipo.cambio", "visitas.cambio",
		"integracion.cambio", "mitienda.publicada", "tienda.cambio"} {
		h.Publicar(Evento{Tienda: "sol", Tipo: tipo, ID: "1"})
	}
	h.repartir(time.Now())

	esperado := map[*Cliente][]string{
		duenia: {"actividad.nuevo", "cliente.nuevo", "comprobante.cambio", "cupon.cambio", "equipo.cambio", "finanzas.cambio",
			"integracion.cambio", "mitienda.publicada", "pago.acreditado", "pedido.nuevo", "plan.cambio", "producto.stock", "tienda.cambio", "visitas.cambio"},
		admin: {"actividad.nuevo", "cliente.nuevo", "comprobante.cambio", "cupon.cambio", "finanzas.cambio",
			"integracion.cambio", "mitienda.publicada", "pago.acreditado", "pedido.nuevo", "producto.stock", "tienda.cambio", "visitas.cambio"},
		ventas: {"cliente.nuevo", "pago.acreditado", "pedido.nuevo", "producto.stock", "tienda.cambio", "visitas.cambio"},
		local:  {"pago.acreditado", "pedido.nuevo", "producto.stock", "tienda.cambio", "visitas.cambio"},
		otra:   nil,
	}
	for c, quiere := range esperado {
		got := map[string]bool{}
		for _, m := range recibidos(c) {
			got[m.Tipo] = true
		}
		if len(got) != len(quiere) {
			t.Errorf("%s (%s) recibió %v, quería %v", c.Usuario, c.Rol, got, quiere)
			continue
		}
		for _, q := range quiere {
			if !got[q] {
				t.Errorf("%s (%s) no recibió %s", c.Usuario, c.Rol, q)
			}
		}
	}
}

func TestTipoDesconocidoNoLlega(t *testing.T) {
	h := hubDePrueba()
	c := registrar(t, h, &Cliente{Tienda: "sol", Usuario: "ana", Sesion: "s1", Rol: "duenio"})
	h.Publicar(Evento{Tienda: "sol", Tipo: "secreto.cambio"})
	h.repartir(time.Now())
	if ms := recibidos(c); len(ms) != 0 {
		t.Fatalf("un tipo sin destinatarios no tendría que llegar: %v", ms)
	}
}

func TestSesionesSoloDeLaPersona(t *testing.T) {
	h := hubDePrueba()
	ana1 := registrar(t, h, &Cliente{Tienda: "sol", Usuario: "ana", Sesion: "s1", Rol: "duenio"})
	ana2 := registrar(t, h, &Cliente{Tienda: "luna", Usuario: "ana", Sesion: "s2", Rol: "ventas"})
	beto := registrar(t, h, &Cliente{Tienda: "sol", Usuario: "beto", Sesion: "s3", Rol: "duenio"})

	h.Publicar(Evento{Tipo: "sesion.cambio", Usuario: "ana"})
	h.repartir(time.Now())
	if len(recibidos(ana1)) != 1 || len(recibidos(ana2)) != 1 || len(recibidos(beto)) != 0 {
		t.Fatal("el cambio de sesiones tiene que llegar a cada pestaña de esa persona y a nadie más")
	}

	// Cerrar la sesión s1 corta esa conexión (y sólo esa).
	h.Publicar(Evento{Tienda: "sol", Tipo: "sesion.cerrada", Usuario: "ana", Sesion: "s1"})
	if k, ok := corte(ana1); !ok || k.Codigo != CierreSinSesion {
		t.Fatalf("sesión cerrada: %+v %v", k, ok)
	}
	if _, ok := corte(ana2); ok {
		t.Fatal("la otra sesión de la misma persona sigue")
	}
	// Cambiar de tienda: resincronizar.
	h.Publicar(Evento{Tienda: "sol", Tipo: "sesion.tienda", Usuario: "ana", Sesion: "s2"})
	if k, ok := corte(ana2); !ok || k.Codigo != CierreResincronizar {
		t.Fatalf("cambio de tienda: %+v %v", k, ok)
	}
	if _, ok := corte(beto); ok {
		t.Fatal("otra persona no se toca")
	}
}

// La campanita (§45) es de cada persona: el aviso llega sólo a sus pestañas en esa tienda, sin
// importar el rol, y a nadie más (ni a la otra tienda de la misma persona, ni a la tienda online).
func TestNotificacionSoloALaPersonaEnEsaTienda(t *testing.T) {
	h := hubDePrueba()
	anaSol := registrar(t, h, &Cliente{Tienda: "sol", Usuario: "ana", Sesion: "s1", Rol: "local"})
	anaLuna := registrar(t, h, &Cliente{Tienda: "luna", Usuario: "ana", Sesion: "s2", Rol: "duenio"})
	beto := registrar(t, h, &Cliente{Tienda: "sol", Usuario: "beto", Sesion: "s3", Rol: "duenio"})
	pub := registrar(t, h, &Cliente{Tienda: "sol", Publico: true, IP: "1.1.1.1"})

	h.Publicar(Evento{Tienda: "sol", Tipo: "notificacion.nueva", Usuario: "ana"})
	h.Publicar(Evento{Tienda: "sol", Tipo: "notificacion.nueva", Usuario: "ana"})
	h.repartir(time.Now())
	if ms := recibidos(anaSol); len(ms) != 1 || ms[0].Tipo != "notificacion.nueva" || ms[0].ID != "" {
		t.Fatalf("ana en sol: %v", ms)
	}
	if len(recibidos(anaLuna)) != 0 || len(recibidos(beto)) != 0 || len(recibidos(pub)) != 0 {
		t.Fatal("la notificación de ana en sol le llegó a otra pestaña")
	}
	// Sin persona no le llega a nadie.
	h.Publicar(Evento{Tienda: "sol", Tipo: "notificacion.nueva"})
	h.repartir(time.Now())
	if len(recibidos(anaSol)) != 0 || len(recibidos(beto)) != 0 {
		t.Fatal("sin persona no se reparte")
	}
}

func TestCambioDeRolCortaALaPersonaAfectada(t *testing.T) {
	h := hubDePrueba()
	duenia := registrar(t, h, &Cliente{Tienda: "sol", Usuario: "ana", Sesion: "s1", Rol: "duenio"})
	caro := registrar(t, h, &Cliente{Tienda: "sol", Usuario: "caro", Sesion: "s2", Rol: "ventas"})
	caroLuna := registrar(t, h, &Cliente{Tienda: "luna", Usuario: "caro", Sesion: "s3", Rol: "admin"})
	h.Publicar(Evento{Tienda: "sol", Tipo: "equipo.cambio", ID: "caro", Usuario: "caro"})
	if k, ok := corte(caro); !ok || k.Codigo != CierreResincronizar {
		t.Fatalf("a quien le cambian el rol se le refresca el panel: %+v %v", k, ok)
	}
	if _, ok := corte(caroLuna); ok {
		t.Fatal("su sesión en otra tienda sigue")
	}
	h.repartir(time.Now())
	if got := tipos(recibidos(duenia)); len(got) != 1 || got[0] != "equipo.cambio" {
		t.Fatalf("la dueña ve el cambio del equipo: %v", got)
	}
}

func TestCoalescencia(t *testing.T) {
	h := hubDePrueba()
	c := registrar(t, h, &Cliente{Tienda: "sol", Usuario: "ana", Sesion: "s1", Rol: "duenio"})
	// El mismo aviso repetido en la ventana sale una vez.
	for range 10 {
		h.Publicar(Evento{Tienda: "sol", Tipo: "pedido.cambio", ID: "7"})
	}
	h.repartir(time.Now())
	if ms := recibidos(c); len(ms) != 1 || ms[0].ID != "7" {
		t.Fatalf("repetidos: %v", ms)
	}
	// Muchos distintos del mismo tipo: un lote.
	for i := range 40 {
		h.Publicar(Evento{Tienda: "sol", Tipo: "producto.cambio", ID: string(rune('a' + i))})
	}
	h.Publicar(Evento{Tienda: "sol", Tipo: "pedido.nuevo", ID: "8"})
	h.repartir(time.Now())
	ms := recibidos(c)
	if len(ms) != 2 {
		t.Fatalf("un lote y el pedido: %v", ms)
	}
	for _, m := range ms {
		if m.Tipo == "producto.cambio" && (!m.Lote || m.ID != "") {
			t.Fatalf("el lote va sin id: %+v", m)
		}
	}
	// Un lote que ya viene de Postgres también se junta en uno.
	h.Publicar(Evento{Tienda: "sol", Tipo: "producto.stock", Lote: true})
	h.Publicar(Evento{Tienda: "sol", Tipo: "producto.stock", Lote: true})
	h.repartir(time.Now())
	if ms := recibidos(c); len(ms) != 1 || !ms[0].Lote {
		t.Fatalf("lotes repetidos: %v", ms)
	}
}

func TestEspaciar(t *testing.T) {
	h := hubDePrueba()
	h.Espaciar = map[string]time.Duration{"visitas": 10 * time.Second}
	c := registrar(t, h, &Cliente{Tienda: "sol", Usuario: "ana", Sesion: "s1", Rol: "duenio"})
	t0 := time.Now()
	h.Publicar(Evento{Tienda: "sol", Tipo: "visitas.cambio"})
	h.repartir(t0)
	if len(recibidos(c)) != 1 {
		t.Fatal("el primero sale enseguida")
	}
	h.Publicar(Evento{Tienda: "sol", Tipo: "visitas.cambio"})
	h.repartir(t0.Add(time.Second))
	if len(recibidos(c)) != 0 {
		t.Fatal("dentro del intervalo espera")
	}
	// No se pierde: sale al cumplirse el intervalo.
	h.repartir(t0.Add(11 * time.Second))
	if len(recibidos(c)) != 1 {
		t.Fatal("lo espaciado sale al final del intervalo")
	}
	// Otra tienda no espera por esta.
	d := registrar(t, h, &Cliente{Tienda: "luna", Usuario: "eva", Sesion: "s2", Rol: "duenio"})
	h.Publicar(Evento{Tienda: "luna", Tipo: "visitas.cambio"})
	h.repartir(t0.Add(12 * time.Second))
	if len(recibidos(d)) != 1 {
		t.Fatal("cada tienda tiene su intervalo")
	}
}

func TestClienteLentoSeCorta(t *testing.T) {
	h := hubDePrueba()
	h.MaxPorTipo = 1000
	lento := registrar(t, h, &Cliente{Tienda: "sol", Usuario: "ana", Sesion: "s1", Rol: "duenio"})
	rapido := registrar(t, h, &Cliente{Tienda: "sol", Usuario: "beto", Sesion: "s2", Rol: "duenio"})
	for i := range 40 {
		h.Publicar(Evento{Tienda: "sol", Tipo: "pedido.cambio", ID: string(rune('a' + i))})
		h.repartir(time.Now())
		recibidos(rapido)
	}
	if k, ok := corte(lento); !ok || k.Codigo != CierreLento {
		t.Fatalf("al llenarse la cola se corta: %+v %v", k, ok)
	}
	if _, ok := corte(rapido); ok {
		t.Fatal("el que lee sigue")
	}
}

func TestLimites(t *testing.T) {
	h := hubDePrueba()
	h.MaxPorSesion, h.MaxPorIP = 2, 3
	var cs []*Cliente
	for range 2 {
		cs = append(cs, registrar(t, h, &Cliente{Tienda: "sol", Usuario: "ana", Sesion: "s1", Rol: "duenio", IP: "1.1.1.1"}))
	}
	if err := h.Registrar(&Cliente{Tienda: "sol", Usuario: "ana", Sesion: "s1", Rol: "duenio", IP: "2.2.2.2"}); !errors.Is(err, ErrLimite) {
		t.Fatalf("tope por sesión: %v", err)
	}
	registrar(t, h, &Cliente{Tienda: "sol", Usuario: "beto", Sesion: "s2", Rol: "duenio", IP: "1.1.1.1"})
	if err := h.Registrar(&Cliente{Tienda: "sol", Usuario: "caro", Sesion: "s3", Rol: "duenio", IP: "1.1.1.1"}); !errors.Is(err, ErrLimite) {
		t.Fatalf("tope por IP: %v", err)
	}
	h.Quitar(cs[0])
	registrar(t, h, &Cliente{Tienda: "sol", Usuario: "ana", Sesion: "s1", Rol: "duenio", IP: "1.1.1.1"})
	if h.Conexiones() != 3 {
		t.Fatalf("conexiones: %d", h.Conexiones())
	}
}

func TestPublicoSoloLoSuyo(t *testing.T) {
	h := hubDePrueba()
	vidriera := registrar(t, h, &Cliente{Tienda: "sol", Publico: true, IP: "9.9.9.9"})
	comprador := registrar(t, h, &Cliente{Tienda: "sol", Publico: true, Pedido: "12", IP: "9.9.9.9"})
	otraTienda := registrar(t, h, &Cliente{Tienda: "luna", Publico: true, Pedido: "12", IP: "9.9.9.9"})
	for _, e := range []Evento{
		{Tienda: "sol", Tipo: "producto.stock", ID: "p1"},
		{Tienda: "sol", Tipo: "pedido.nuevo", ID: "13"},
		{Tienda: "sol", Tipo: "pedido.cambio", ID: "13"},
		{Tienda: "sol", Tipo: "pago.acreditado", ID: "12"},
		{Tienda: "sol", Tipo: "cliente.nuevo", ID: "c1"},
		{Tienda: "sol", Tipo: "finanzas.cambio"},
		{Tienda: "sol", Tipo: "sesion.cambio", Usuario: "ana"},
	} {
		h.Publicar(e)
	}
	h.repartir(time.Now())
	if got := tipos(recibidos(vidriera)); len(got) != 1 || got[0] != "producto.stock" {
		t.Fatalf("la vidriera sólo ve productos: %v", got)
	}
	ms := recibidos(comprador)
	pago := false
	for _, m := range ms {
		if m.Tipo == "pago.acreditado" {
			pago = true
			if m.ID != "" {
				t.Fatal("al comprador no le viaja ni el número")
			}
		}
		if m.Tipo == "pedido.cambio" || m.Tipo == "pedido.nuevo" {
			t.Fatalf("un pedido ajeno no le llega: %v", ms)
		}
	}
	if !pago || len(ms) != 2 {
		t.Fatalf("el comprador ve su pago y el stock: %v", ms)
	}
	if len(recibidos(otraTienda)) != 0 {
		t.Fatal("el mismo número en otra tienda no es su pedido")
	}
}

func TestResincronizarYCerrar(t *testing.T) {
	h := hubDePrueba()
	a := registrar(t, h, &Cliente{Tienda: "sol", Usuario: "ana", Sesion: "s1", Rol: "local"})
	p := registrar(t, h, &Cliente{Tienda: "luna", Publico: true})
	h.Resincronizar()
	for _, c := range []*Cliente{a, p} {
		if got := tipos(recibidos(c)); len(got) != 1 || got[0] != TipoResincronizar {
			t.Fatalf("resincronizar: %v", got)
		}
	}
	h.Cerrar()
	for _, c := range []*Cliente{a, p} {
		if k, ok := corte(c); !ok || k.Codigo != websocket.StatusServiceRestart {
			t.Fatalf("al apagar: %+v %v", k, ok)
		}
	}
	if err := h.Registrar(&Cliente{Tienda: "sol", Usuario: "x", Sesion: "s9", Rol: "duenio"}); err == nil {
		t.Fatal("apagado no acepta conexiones nuevas")
	}
}
