package arrepentimiento_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/arrepentimiento"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/servidor"
)

var ctx = context.Background()

func nuevo(t *testing.T) (*arrepentimiento.Servicio, *correo.Memoria, *pruebadb.Bases) {
	b := pruebadb.Nueva(t)
	m := &correo.Memoria{}
	ahora := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	return &arrepentimiento.Servicio{Pool: b.App, Correo: m, Ahora: func() time.Time { return ahora }}, m, b
}

var pedido = arrepentimiento.Pedido{
	Email: " Lucia@LaEsquina.com.ar ", Nombre: "Lucía  Fernández", Documento: "27-38412905-4",
	Plan: "Pro mensual · 14 de octubre de 2026",
}

func TestPedidoRegistraCodigoYMandaElMail(t *testing.T) {
	s, m, b := nuevo(t)
	codigo, err := s.Pedir(ctx, pedido, "190.2.3.4")
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^ARR-2026-\d{6}$`).MatchString(codigo) {
		t.Fatalf("código %q", codigo)
	}
	otro, err := s.Pedir(ctx, pedido, "190.2.3.4")
	if err != nil || otro == codigo {
		t.Fatalf("el segundo pedido tiene que tener otro código: %q %v", otro, err)
	}

	msj, ok := m.Ultimo("lucia@laesquina.com.ar")
	if !ok || !strings.Contains(msj.Asunto, otro) || !strings.Contains(msj.Texto, "10 días hábiles") {
		t.Fatalf("mail %+v", msj)
	}

	t.Run("se guarda normalizado y sólo staff lo lee", func(t *testing.T) {
		var email, nombre, doc, estado string
		if err := b.Staff.QueryRow(ctx, `SELECT email, nombre, documento, estado FROM soporte.arrepentimientos WHERE codigo = $1`, codigo).
			Scan(&email, &nombre, &doc, &estado); err != nil {
			t.Fatal(err)
		}
		if email != "lucia@laesquina.com.ar" || nombre != "Lucía Fernández" || doc != "27384129054" || estado != "recibido" {
			t.Fatalf("%s %s %s %s", email, nombre, doc, estado)
		}
		if _, err := b.App.Exec(ctx, `SELECT email, documento FROM soporte.arrepentimientos`); err == nil {
			t.Fatal("la app no debería poder leer los datos de los pedidos")
		}
	})
}

func TestPedidoInvalido(t *testing.T) {
	s, _, _ := nuevo(t)
	casos := map[string]func(*arrepentimiento.Pedido){
		"email":     func(p *arrepentimiento.Pedido) { p.Email = "lucia" },
		"nombre":    func(p *arrepentimiento.Pedido) { p.Nombre = "  " },
		"documento": func(p *arrepentimiento.Pedido) { p.Documento = "123" },
		"plan":      func(p *arrepentimiento.Pedido) { p.Plan = "" },
		"motivo":    func(p *arrepentimiento.Pedido) { p.Motivo = strings.Repeat("a", 1001) },
	}
	for campo, cambiar := range casos {
		p := pedido
		cambiar(&p)
		var inv arrepentimiento.ErrInvalido
		if _, err := s.Pedir(ctx, p, "190.2.3.4"); !errors.As(err, &inv) || inv.Campo != campo {
			t.Errorf("%s: err = %v", campo, err)
		}
	}
}

func TestLimiteDePedidosPorEmail(t *testing.T) {
	s, _, _ := nuevo(t)
	for i := range 3 {
		if _, err := s.Pedir(ctx, pedido, "190.2.3.4"); err != nil {
			t.Fatalf("pedido %d: %v", i, err)
		}
	}
	if _, err := s.Pedir(ctx, pedido, "190.2.3.5"); !errors.Is(err, arrepentimiento.ErrDemasiadosPedidos) {
		t.Fatalf("err = %v", err)
	}
}

func TestHTTP(t *testing.T) {
	s, _, b := nuevo(t)
	h := servidor.Nuevo(servidor.Dependencias{App: b.App, Arrepentimiento: s})
	pedir := func(cuerpo string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", "/v1/arrepentimientos", strings.NewReader(cuerpo))
		r.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		return rec
	}
	rec := pedir(`{"email":"lucia@laesquina.com.ar","nombre":"Lucía Fernández","documento":"38412905","plan":"Pro mensual"}`)
	if rec.Code != 201 || !strings.Contains(rec.Body.String(), `"codigo":"ARR-2026-`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
	rec = pedir(`{"email":"lucia@laesquina.com.ar","nombre":"","documento":"38412905","plan":"Pro"}`)
	if rec.Code != 400 || !strings.Contains(rec.Body.String(), `"path":"nombre"`) {
		t.Fatalf("%d %s", rec.Code, rec.Body)
	}
}
