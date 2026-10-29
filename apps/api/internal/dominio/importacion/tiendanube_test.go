package importacion_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/importacion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/tiendanube"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// canjeFalso responde el canje del código como Tiendanube, sin salir a la red.
type canjeFalso struct{ pedidos []map[string]string }

func (c *canjeFalso) RoundTrip(r *http.Request) (*http.Response, error) {
	var b map[string]string
	_ = json.NewDecoder(r.Body).Decode(&b)
	c.pedidos = append(c.pedidos, b)
	cuerpo := `{"error":"invalid_grant"}`
	estado := http.StatusBadRequest
	if r.URL.Host == "www.tiendanube.com" && b["code"] == "codigo-bueno" && b["client_secret"] == "secreto" {
		cuerpo, estado = `{"access_token":"tok-123","token_type":"bearer","scope":"read_products,read_customers,read_orders","user_id":4567}`, http.StatusOK
	}
	return &http.Response{StatusCode: estado, Body: io.NopCloser(strings.NewReader(cuerpo)), Header: http.Header{}}, nil
}

func stateDe(t *testing.T, u string) string {
	t.Helper()
	p, err := url.Parse(u)
	if err != nil || !strings.HasPrefix(u, "https://www.tiendanube.com/apps/321/authorize?") {
		t.Fatalf("URL de autorización: %q", u)
	}
	return p.Query().Get("state")
}

func TestOAuthTiendanube(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	f := fuenteCompleta()
	s := servicio(t, b, f)
	const duenia, otra = "00000000-0000-4000-8000-000000000001", "00000000-0000-4000-8000-000000000002"

	// Sin la app configurada, no está disponible y no se arma nada.
	if s.TiendanubeDisponible() {
		t.Fatal("disponible sin credenciales")
	}
	if _, err := s.ConectarTiendanube(ctx, luna, "duenio", duenia); !errors.Is(err, importacion.ErrTiendanubeApagado) {
		t.Fatalf("apagado: %v", err)
	}
	canje := &canjeFalso{}
	s.Tiendanube = tiendanube.App{ID: "321", Secreto: "secreto", HTTP: &http.Client{Transport: canje}}
	if _, err := s.ConectarTiendanube(ctx, luna, "ventas", duenia); !errors.Is(err, importacion.ErrSinPermiso) {
		t.Fatalf("ventas pudo conectar: %v", err)
	}

	u, err := s.ConectarTiendanube(ctx, luna, "duenio", duenia)
	if err != nil {
		t.Fatal(err)
	}
	estado := stateDe(t, u)
	// El state es de la tienda y de la persona que lo pidió.
	if _, err := s.VolverTiendanube(ctx, sol, "duenio", duenia, "codigo-bueno", estado); !errors.Is(err, importacion.ErrEstado) {
		t.Fatalf("otra tienda: %v", err)
	}
	if _, err := s.VolverTiendanube(ctx, luna, "duenio", otra, "codigo-bueno", estado); !errors.Is(err, importacion.ErrEstado) {
		t.Fatalf("otra persona: %v", err)
	}
	if len(canje.pedidos) != 0 {
		t.Fatalf("canjeó con un state ajeno: %v", canje.pedidos)
	}

	imp, err := s.VolverTiendanube(ctx, luna, "duenio", duenia, "codigo-bueno", estado)
	if err != nil || imp.Plataforma != "tiendanube" || imp.Estado != "pendiente" {
		t.Fatalf("vuelta: %+v %v", imp, err)
	}
	if p := canje.pedidos[0]; p["client_id"] != "321" || p["grant_type"] != "authorization_code" {
		t.Fatalf("canje: %v", p)
	}
	// Un solo uso.
	if _, err := s.VolverTiendanube(ctx, luna, "duenio", duenia, "codigo-bueno", estado); !errors.Is(err, importacion.ErrEstado) {
		t.Fatalf("el state sirvió dos veces: %v", err)
	}

	// La fuente recibe la credencial descifrada, con el token y el id de la tienda.
	correr(t, b, s)
	var c struct{ Token, Tienda string }
	if err := json.Unmarshal([]byte(f.cred), &c); err != nil || c.Token != "tok-123" || c.Tienda != "4567" {
		t.Fatalf("credencial: %q %v", f.cred, err)
	}
	var cred *[]byte
	if err := b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT credencial FROM importaciones WHERE id::text = $1`, imp.ID).Scan(&cred)
	}); err != nil || cred != nil {
		t.Fatalf("la credencial quedó guardada: %v %v", cred, err)
	}

	// Un código que Tiendanube rechaza no crea nada.
	u, _ = s.ConectarTiendanube(ctx, sol, "admin", duenia)
	if _, err := s.VolverTiendanube(ctx, sol, "admin", duenia, "codigo-malo", stateDe(t, u)); !errors.Is(err, importacion.ErrCanje) {
		t.Fatalf("canje malo: %v", err)
	}
	if a, err := s.Actual(ctx, sol); err != nil || a != nil {
		t.Fatalf("creó una importación sin token: %+v %v", a, err)
	}
}
