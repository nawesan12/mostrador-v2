package formularios_test

import (
	"context"
	"errors"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/formularios"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/servidor"
)

var ctx = context.Background()

func nuevo(t *testing.T) (*formularios.Servicio, *correo.Memoria, *pruebadb.Bases) {
	b := pruebadb.Nueva(t)
	m := &correo.Memoria{}
	return &formularios.Servicio{Pool: b.App, Correo: m, Sitio: "https://tulocalweb.com.ar", Equipo: "equipo@tulocalweb.com.ar"}, m, b
}

func TestConsultaDeVentas(t *testing.T) {
	s, m, b := nuevo(t)
	err := s.Consultar(ctx, formularios.Consulta{
		Nombre: " Lucía  Fernández ", Email: "Lucia@TuTienda.com.ar", Whatsapp: "11 5555-0000",
		Web: "tutienda.com.ar", Facturacion: "5-a-20", Origen: "tiendanube", Mensaje: "Tengo 2 locales.",
	}, "190.2.3.4")
	if err != nil {
		t.Fatal(err)
	}
	var nombre, wa, fact string
	if err := b.Staff.QueryRow(ctx, `SELECT nombre, whatsapp, facturacion FROM soporte.consultas`).Scan(&nombre, &wa, &fact); err != nil {
		t.Fatal(err)
	}
	if nombre != "Lucía Fernández" || wa != "1155550000" || fact != "5-a-20" {
		t.Fatalf("%s %s %s", nombre, wa, fact)
	}
	if msj, ok := m.Ultimo("lucia@tutienda.com.ar"); !ok || !strings.Contains(msj.Texto, "a la brevedad") {
		t.Fatalf("confirmación %+v", msj)
	}
	if msj, ok := m.Ultimo("equipo@tulocalweb.com.ar"); !ok || !strings.Contains(msj.Texto, "$ 5 a 20 M") || !strings.Contains(msj.Texto, "Tiendanube") {
		t.Fatalf("aviso al equipo %+v", msj)
	}
	if _, err := b.App.Exec(ctx, `SELECT email FROM soporte.consultas`); err == nil {
		t.Fatal("la app no debería leer las consultas")
	}
}

func TestConsultaInvalida(t *testing.T) {
	s, _, _ := nuevo(t)
	casos := map[string]formularios.Consulta{
		"nombre":      {Email: "a@b.com"},
		"email":       {Nombre: "Lucía", Email: "lucia"},
		"whatsapp":    {Nombre: "Lucía", Email: "a@b.com", Whatsapp: "123"},
		"facturacion": {Nombre: "Lucía", Email: "a@b.com", Facturacion: "mucho"},
		"origen":      {Nombre: "Lucía", Email: "a@b.com", Origen: "otra"},
	}
	for campo, c := range casos {
		var inv formularios.ErrInvalido
		if err := s.Consultar(ctx, c, "190.2.3.4"); !errors.As(err, &inv) || inv.Campo != campo {
			t.Errorf("%s: %v", campo, err)
		}
	}
}

func TestPostulacionPartner(t *testing.T) {
	s, m, b := nuevo(t)
	p := formularios.Postulacion{Nombre: "Juan Gómez", Email: "juan@agencia.com", Agencia: "Estudio Brote",
		Especialidades: []string{"diseno", "desarrollo", "diseno"}}
	if err := s.Postular(ctx, p, "190.2.3.4"); err != nil {
		t.Fatal(err)
	}
	// La segunda vez no es error ni manda otro mail.
	if err := s.Postular(ctx, p, "190.2.3.4"); err != nil {
		t.Fatal(err)
	}
	var n int
	var esp []string
	if err := b.Staff.QueryRow(ctx, `SELECT count(*), max(especialidades::text)::text[] FROM soporte.postulaciones_partner`).Scan(&n, &esp); err != nil {
		t.Fatal(err)
	}
	if n != 1 || len(esp) != 2 {
		t.Fatalf("n=%d esp=%v", n, esp)
	}
	if len(m.Enviados) != 2 {
		t.Fatalf("mails %d", len(m.Enviados))
	}
	var inv formularios.ErrInvalido
	if err := s.Postular(ctx, formularios.Postulacion{Nombre: "Ana", Email: "ana@x.com", Agencia: "X", Especialidades: []string{"magia"}}, "1.1.1.1"); !errors.As(err, &inv) {
		t.Fatalf("err = %v", err)
	}
}

func TestNovedadesSuscripcionYBaja(t *testing.T) {
	s, m, _ := nuevo(t)
	if err := s.Suscribir(ctx, "Lucia@TuTienda.com.ar", "190.2.3.4"); err != nil {
		t.Fatal(err)
	}
	if err := s.Suscribir(ctx, "lucia@tutienda.com.ar", "190.2.3.4"); err != nil {
		t.Fatal(err)
	}
	if len(m.Enviados) != 1 {
		t.Fatalf("una sola confirmación, hubo %d", len(m.Enviados))
	}
	link := m.Enviados[0].Texto[strings.Index(m.Enviados[0].Texto, "https://"):]
	link = strings.Fields(link)[0]
	u, _ := url.Parse(link)
	token := u.Query().Get("token")
	if u.Path != "/novedades/baja" || token == "" {
		t.Fatalf("link %q", link)
	}
	if err := s.Baja(ctx, token); err != nil {
		t.Fatal(err)
	}
	if err := s.Baja(ctx, token); !errors.Is(err, formularios.ErrTokenInvalido) {
		t.Fatalf("la segunda baja: %v", err)
	}
	// Volver a sumarse la reactiva con un link nuevo.
	if err := s.Suscribir(ctx, "lucia@tutienda.com.ar", "190.2.3.4"); err != nil || len(m.Enviados) != 2 {
		t.Fatalf("resuscribir: %v, mails %d", err, len(m.Enviados))
	}
}

func TestLimite(t *testing.T) {
	s, _, _ := nuevo(t)
	for i := range 3 {
		if err := s.Consultar(ctx, formularios.Consulta{Nombre: "Lucía", Email: "a@b.com"}, "190.2.3.4"); err != nil {
			t.Fatalf("%d: %v", i, err)
		}
	}
	if err := s.Consultar(ctx, formularios.Consulta{Nombre: "Lucía", Email: "a@b.com"}, "190.2.3.9"); !errors.Is(err, formularios.ErrDemasiados) {
		t.Fatalf("err = %v", err)
	}
}

func TestHTTP(t *testing.T) {
	s, _, b := nuevo(t)
	h := servidor.Nuevo(servidor.Dependencias{App: b.App, Formularios: s})
	pedir := func(ruta, cuerpo string) *httptest.ResponseRecorder {
		r := httptest.NewRequest("POST", ruta, strings.NewReader(cuerpo))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		return rec
	}
	if rec := pedir("/v1/contacto", `{"nombre":"Lucía","email":"lucia@x.com","facturacion":"hasta-5"}`); rec.Code != 202 {
		t.Fatalf("contacto %d %s", rec.Code, rec.Body)
	}
	if rec := pedir("/v1/partners/postulaciones", `{"nombre":"Juan","email":"juan@x.com","agencia":"Brote","especialidades":["marketing"]}`); rec.Code != 202 {
		t.Fatalf("partners %d %s", rec.Code, rec.Body)
	}
	if rec := pedir("/v1/novedades/suscripciones", `{"email":"juan@x.com"}`); rec.Code != 202 {
		t.Fatalf("novedades %d %s", rec.Code, rec.Body)
	}
	if rec := pedir("/v1/novedades/baja", `{"token":"nada"}`); rec.Code != 422 {
		t.Fatalf("baja %d %s", rec.Code, rec.Body)
	}
	if rec := pedir("/v1/contacto", `{"nombre":"","email":"lucia@x.com"}`); rec.Code != 400 || !strings.Contains(rec.Body.String(), `"path":"nombre"`) {
		t.Fatalf("inválido %d %s", rec.Code, rec.Body)
	}
}
