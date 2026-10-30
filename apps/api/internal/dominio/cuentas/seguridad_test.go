package cuentas_test

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

func TestDispositivo(t *testing.T) {
	for _, c := range []struct{ agente, tipo, nombre string }{
		{"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Safari/537.36", "compu", "Chrome en Mac"},
		{"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Safari/537.36 Edg/130.0", "compu", "Edge en Windows"},
		{"Mozilla/5.0 (X11; Linux x86_64; rv:131.0) Gecko/20100101 Firefox/131.0", "compu", "Firefox en Linux"},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Mobile/15E148 Safari/604.1", "celu", "Safari en iPhone"},
		{"Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/130.0 Mobile/15E148 Safari/604.1", "celu", "Chrome en iPhone"},
		{"Mozilla/5.0 (iPad; CPU OS 18_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.0 Mobile/15E148 Safari/604.1", "tablet", "Safari en iPad"},
		{"Mozilla/5.0 (Linux; Android 14; SM-S911B) AppleWebKit/537.36 (KHTML, like Gecko) SamsungBrowser/25.0 Chrome/121.0 Mobile Safari/537.36", "celu", "Samsung Internet en Android"},
		{"Mozilla/5.0 (Linux; Android 14; SM-X710) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Safari/537.36", "tablet", "Chrome en Android"},
		{"TuLocalWeb/12 CFNetwork/1568.100.1 Darwin/24.0.0", "celu", "App de Tu Local Web en iPhone"},
		{"okhttp/4.12.0", "celu", "App de Tu Local Web en Android"},
		{"", "compu", "Dispositivo sin identificar"},
	} {
		a := cuentas.Dispositivo(c.agente)
		if a.Tipo != c.tipo || a.Nombre() != c.nombre {
			t.Errorf("%q: %s / %q, quería %s / %q", c.agente, a.Tipo, a.Nombre(), c.tipo, c.nombre)
		}
	}
}

func TestIPCorta(t *testing.T) {
	for ip, want := range map[string]string{
		"190.2.3.4": "190.2.3.4", "::ffff:190.2.3.4": "190.2.3.4", "2800:810:4a2:1f00:aa:bb:cc:dd": "2800:810:4a2:1f00:…",
		"2001:db8::1": "2001:db8:0:0:…", "": "", "basura": "",
	} {
		if got := cuentas.IPCorta(ip); got != want {
			t.Errorf("IPCorta(%q) = %q, quería %q", ip, got, want)
		}
	}
}

// La app de iPhone le pega directo a la API (Cloudflare → Caddy → Go): la IP es la de Cloudflare.
func TestClienteDeDetrasDeCaddy(t *testing.T) {
	r := httptest.NewRequest("POST", "/v1/auth/clerk", nil)
	r.RemoteAddr = "[::1]:41000"
	r.Header.Set("Cf-Connecting-Ip", "190.2.3.4")
	if c := cuentas.ClienteDe(r, "secreto"); c.IP != "190.2.3.4" {
		t.Fatalf("desde Caddy: %s", c.IP)
	}
	r.RemoteAddr = "172.18.0.1:41000"
	if c := cuentas.ClienteDe(r, "secreto"); c.IP != "190.2.3.4" {
		t.Fatalf("desde la red de Docker: %s", c.IP)
	}
	r.RemoteAddr = "200.1.1.1:41000"
	if c := cuentas.ClienteDe(r, "secreto"); c.IP != "200.1.1.1" {
		t.Fatalf("una conexión de internet no elige su IP: %s", c.IP)
	}
}

type lugares map[string]string

func (l lugares) Ubicar(ip string) string { return l[ip] }

func avisos(t *testing.T, b *pruebadb.Bases) []outbox.Evento {
	t.Helper()
	filas, err := b.Worker.Query(ctx, `SELECT id, tipo, datos FROM outbox WHERE tipo = 'cuenta.ingreso' AND tienda_id IS NULL ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer filas.Close()
	var out []outbox.Evento
	for filas.Next() {
		var e outbox.Evento
		if err := filas.Scan(&e.ID, &e.Tipo, &e.Datos); err != nil {
			t.Fatal(err)
		}
		out = append(out, e)
	}
	return out
}

func TestAvisoDeIngreso(t *testing.T) {
	s, _, b := nuevo(t)
	s.Geo = lugares{"190.2.3.4": "Rosario, Santa Fe"}
	if _, err := s.Registrar(ctx, alta("lucia@laesquina.com.ar"), cl); err != nil {
		t.Fatal(err)
	}
	if n := len(avisos(t, b)); n != 0 {
		t.Fatalf("el registro no avisa: %d", n)
	}
	iphone := cuentas.Cliente{IP: "190.2.3.4", Agente: "Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) Version/18.0 Mobile Safari/604.1"}
	ses, err := s.Ingresar(ctx, "lucia@laesquina.com.ar", "Abcdefghijk1", true, iphone)
	if err != nil {
		t.Fatal(err)
	}
	ev := avisos(t, b)
	if len(ev) != 1 {
		t.Fatalf("ingresar avisa una vez: %d", len(ev))
	}

	m := &correo.Memoria{}
	if err := cuentas.Manejadores(m, "https://tulocalweb.com.ar")["cuenta.ingreso"](ctx, ev[0]); err != nil {
		t.Fatal(err)
	}
	if len(m.Enviados) != 1 {
		t.Fatalf("mails: %d", len(m.Enviados))
	}
	mail := m.Enviados[0]
	hora := cuentas.HoraAR(time.Now())[:10]
	for _, quiero := range []string{"Safari en iPhone", "Rosario, Santa Fe", "190.2.3.4", hora, "hora de Argentina",
		"https://tulocalweb.com.ar/panel/configuracion?tab=seguridad"} {
		if !strings.Contains(mail.Texto, quiero) || !strings.Contains(mail.HTML, quiero) {
			t.Errorf("al mail le falta %q:\n%s", quiero, mail.Texto)
		}
	}
	if mail.Para != "lucia@laesquina.com.ar" || mail.Asunto != "Entraron a tu cuenta de Tu Local Web" {
		t.Fatalf("%+v", mail)
	}

	t.Run("apagado no avisa", func(t *testing.T) {
		d, _ := s.Sesion(ctx, ses.Token)
		if _, err := s.Pool.Exec(ctx, `UPDATE auth.usuarios SET seguridad = '{"avisoIngreso": false}' WHERE id = $1`, d.Usuario.ID); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Ingresar(ctx, "lucia@laesquina.com.ar", "Abcdefghijk1", false, cl); err != nil {
			t.Fatal(err)
		}
		if n := len(avisos(t, b)); n != 1 {
			t.Fatalf("avisó con el aviso apagado: %d", n)
		}
	})

	t.Run("sin lugar ni IP, el mail no los inventa", func(t *testing.T) {
		m := &correo.Memoria{}
		e := outbox.Evento{Datos: []byte(`{"para":"a@b.co","nombre":"Ana","aparato":"Chrome en Mac","hora":"30/10/2026 a las 14:05"}`)}
		if err := cuentas.Manejadores(m, "https://x")["cuenta.ingreso"](ctx, e); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(m.Enviados[0].Texto, "Lugar") || strings.Contains(m.Enviados[0].Texto, "IP:") {
			t.Fatalf("%s", m.Enviados[0].Texto)
		}
	})
}

func TestSesionesInactivas(t *testing.T) {
	s, _, _ := nuevo(t)
	ses, err := s.Registrar(ctx, alta("lucia@laesquina.com.ar"), cl)
	if err != nil {
		t.Fatal(err)
	}
	d, _ := s.Sesion(ctx, ses.Token)
	viejo := func() {
		if _, err := s.Pool.Exec(ctx, `UPDATE auth.sesiones SET visto_en = now() - interval '8 days' WHERE usuario_id = $1`, d.Usuario.ID); err != nil {
			t.Fatal(err)
		}
	}
	viejo()
	if _, err := s.Sesion(ctx, ses.Token); err != nil {
		t.Fatalf("con la opción apagada, la sesión sigue valiendo: %v", err)
	}
	viejo()
	if _, err := s.Pool.Exec(ctx, `UPDATE auth.usuarios SET seguridad = '{"inactivas": true}' WHERE id = $1`, d.Usuario.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Sesion(ctx, ses.Token); err != cuentas.ErrSinSesion {
		t.Fatalf("inactiva hace 8 días: %v", err)
	}
	var n int
	_ = s.Pool.QueryRow(ctx, `SELECT count(*) FROM auth.sesiones WHERE usuario_id = $1`, d.Usuario.ID).Scan(&n)
	if n != 0 {
		t.Fatal("la sesión inactiva no se borró")
	}
}
