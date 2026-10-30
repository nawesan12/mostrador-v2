package cuentas_test

import (
	"context"
	"errors"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

var ctx = context.Background()

func nuevo(t *testing.T) (*cuentas.Servicio, *correo.Memoria, *pruebadb.Bases) {
	b := pruebadb.Nueva(t)
	m := &correo.Memoria{}
	return &cuentas.Servicio{Pool: b.App, Correo: m, Sitio: "https://tulocalweb.com.ar"}, m, b
}

var cl = cuentas.Cliente{IP: "190.2.3.4", Agente: "Chrome · macOS"}

func alta(email string) cuentas.Alta {
	return cuentas.Alta{Nombre: "Lucía", Tienda: "La Esquina", Email: email, Clave: "Abcdefghijk1"}
}

func TestRegistroCreaCuentaTiendaYSesion(t *testing.T) {
	s, _, _ := nuevo(t)
	ses, err := s.Registrar(ctx, alta("  Lucia@LaEsquina.com.ar "), cl)
	if err != nil {
		t.Fatal(err)
	}
	if ses.Token == "" || ses.TiendaID == "" || time.Until(ses.VenceEn) < 29*24*time.Hour {
		t.Fatalf("sesión %+v", ses)
	}
	d, err := s.Sesion(ctx, ses.Token)
	if err != nil {
		t.Fatal(err)
	}
	if d.Usuario.Email != "lucia@laesquina.com.ar" || d.Rol != "duenio" || d.TiendaID != ses.TiendaID {
		t.Fatalf("datos %+v", d)
	}

	t.Run("el mismo email no se registra dos veces", func(t *testing.T) {
		if _, err := s.Registrar(ctx, alta("lucia@laesquina.com.ar"), cl); !errors.Is(err, cuentas.ErrEmailEnUso) {
			t.Fatalf("err = %v", err)
		}
	})

	t.Run("dos tiendas con el mismo nombre tienen slugs distintos", func(t *testing.T) {
		otra, err := s.Registrar(ctx, alta("otra@laesquina.com.ar"), cl)
		if err != nil {
			t.Fatal(err)
		}
		if otra.TiendaID == ses.TiendaID {
			t.Fatal("misma tienda")
		}
	})
}

func TestRegistroValida(t *testing.T) {
	s, _, _ := nuevo(t)
	for _, c := range []struct {
		a     cuentas.Alta
		campo string
	}{
		{cuentas.Alta{Tienda: "X", Email: "a@b.co", Clave: "Abcdefghijk1"}, "nombre"},
		{cuentas.Alta{Nombre: "L", Email: "a@b.co", Clave: "Abcdefghijk1"}, "tienda"},
		{cuentas.Alta{Nombre: "L", Tienda: "X", Email: "no-es-email", Clave: "Abcdefghijk1"}, "email"},
		{cuentas.Alta{Nombre: "L", Tienda: "X", Email: "a@b.co", Clave: "corta"}, "clave"},
	} {
		var inv cuentas.ErrInvalido
		if _, err := s.Registrar(ctx, c.a, cl); !errors.As(err, &inv) || inv.Campo != c.campo {
			t.Errorf("%s: err = %v", c.campo, err)
		}
	}
}

func TestIngresar(t *testing.T) {
	s, _, _ := nuevo(t)
	if _, err := s.Registrar(ctx, alta("lucia@laesquina.com.ar"), cl); err != nil {
		t.Fatal(err)
	}
	t.Run("con la clave correcta", func(t *testing.T) {
		corta, err := s.Ingresar(ctx, "LUCIA@laesquina.com.ar", "Abcdefghijk1", false, cl)
		if err != nil || time.Until(corta.VenceEn) > cuentas.SesionCorta {
			t.Fatalf("sesión %+v err %v", corta, err)
		}
		larga, _ := s.Ingresar(ctx, "lucia@laesquina.com.ar", "Abcdefghijk1", true, cl)
		if time.Until(larga.VenceEn) < 29*24*time.Hour {
			t.Fatalf("recordarme: vence %v", larga.VenceEn)
		}
	})
	t.Run("clave incorrecta y email inexistente dan el mismo error", func(t *testing.T) {
		_, e1 := s.Ingresar(ctx, "lucia@laesquina.com.ar", "otra", false, cl)
		_, e2 := s.Ingresar(ctx, "nadie@laesquina.com.ar", "otra", false, cl)
		if !errors.Is(e1, cuentas.ErrCredenciales) || !errors.Is(e2, cuentas.ErrCredenciales) {
			t.Fatalf("e1=%v e2=%v", e1, e2)
		}
	})
	t.Run("a los 5 intentos por email se frena", func(t *testing.T) {
		var err error
		for i := 0; i < 6; i++ {
			_, err = s.Ingresar(ctx, "frenado@laesquina.com.ar", "mal", false, cl)
		}
		if !errors.Is(err, cuentas.ErrDemasiadosIntentos) {
			t.Fatalf("err = %v", err)
		}
	})
}

func TestSalirCierraLaSesion(t *testing.T) {
	s, _, _ := nuevo(t)
	ses, _ := s.Registrar(ctx, alta("lucia@laesquina.com.ar"), cl)
	if err := s.Salir(ctx, ses.Token); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Sesion(ctx, ses.Token); !errors.Is(err, cuentas.ErrSinSesion) {
		t.Fatalf("err = %v", err)
	}
	if _, err := s.Sesion(ctx, "inventado"); !errors.Is(err, cuentas.ErrSinSesion) {
		t.Fatal("un token inventado no puede tener sesión")
	}
}

var linkRe = regexp.MustCompile(`https://tulocalweb\.com\.ar/ingresar\?m=restablecer&token=([A-Za-z0-9_-]+)`)

func TestRecuperarYRestablecer(t *testing.T) {
	s, m, _ := nuevo(t)
	ses, _ := s.Registrar(ctx, alta("lucia@laesquina.com.ar"), cl)
	if b, ok := m.Ultimo("lucia@laesquina.com.ar"); !ok || !strings.Contains(b.Asunto, "Bienvenida") ||
		!strings.Contains(b.Texto, "/onboarding") {
		t.Fatalf("al registrarse sale el mail de bienvenida: %+v", b)
	}

	if err := s.Recuperar(ctx, "nadie@laesquina.com.ar", cl); err != nil {
		t.Fatalf("un email sin cuenta responde igual: %v", err)
	}
	if _, ok := m.Ultimo("nadie@laesquina.com.ar"); ok {
		t.Fatal("no se manda nada a un email sin cuenta")
	}

	if err := s.Recuperar(ctx, "lucia@laesquina.com.ar", cl); err != nil {
		t.Fatal(err)
	}
	msj, ok := m.Ultimo("lucia@laesquina.com.ar")
	if !ok || !strings.Contains(msj.Asunto, "contraseña") {
		t.Fatalf("mail %+v", msj)
	}
	g := linkRe.FindStringSubmatch(msj.Texto)
	if g == nil {
		t.Fatalf("el mail no trae el link: %s", msj.Texto)
	}
	tok, _ := url.QueryUnescape(g[1])

	if err := s.VerificarRestablecer(ctx, tok); err != nil {
		t.Fatalf("un link sin usar tiene que servir: %v", err)
	}
	if err := s.VerificarRestablecer(ctx, "otro"); !errors.Is(err, cuentas.ErrTokenInvalido) {
		t.Fatalf("un link inventado: %v", err)
	}
	if err := s.Restablecer(ctx, tok, "Nuevaclave123"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Sesion(ctx, ses.Token); !errors.Is(err, cuentas.ErrSinSesion) {
		t.Fatal("cambiar la contraseña tiene que cerrar las sesiones abiertas")
	}
	if _, err := s.Ingresar(ctx, "lucia@laesquina.com.ar", "Nuevaclave123", false, cl); err != nil {
		t.Fatalf("la clave nueva no anda: %v", err)
	}
	if err := s.Restablecer(ctx, tok, "Otraclave1234"); !errors.Is(err, cuentas.ErrTokenInvalido) {
		t.Fatalf("el link se usa una sola vez: %v", err)
	}
	if err := s.VerificarRestablecer(ctx, tok); !errors.Is(err, cuentas.ErrTokenInvalido) {
		t.Fatalf("un link usado ya no sirve al abrirlo: %v", err)
	}
}

func TestRestablecerVencido(t *testing.T) {
	s, m, b := nuevo(t)
	_, _ = s.Registrar(ctx, alta("lucia@laesquina.com.ar"), cl)
	_ = s.Recuperar(ctx, "lucia@laesquina.com.ar", cl)
	msj, _ := m.Ultimo("lucia@laesquina.com.ar")
	tok := linkRe.FindStringSubmatch(msj.Texto)[1]

	// La app no puede tocar el vencimiento (permiso de columna): se vence desde el admin del test.
	if _, err := b.App.Exec(ctx, "UPDATE auth.restablecimientos SET vence_en = now()"); err == nil {
		t.Fatal("la app pudo mover el vencimiento")
	}
	admin, err := pgx.Connect(ctx, b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)
	if _, err := admin.Exec(ctx, "UPDATE auth.restablecimientos SET vence_en = now() - interval '1 minute'"); err != nil {
		t.Fatal(err)
	}
	if err := s.Restablecer(ctx, tok, "Nuevaclave123"); !errors.Is(err, cuentas.ErrTokenInvalido) {
		t.Fatalf("un link vencido no sirve: %v", err)
	}
}
