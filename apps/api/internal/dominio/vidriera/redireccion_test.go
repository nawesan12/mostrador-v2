package vidriera_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// La misma tabla corre en la web (lib/server/importacion/rutas.spec.ts).
//
//go:embed testdata/rutas.json
var casosRutas []byte

func TestNormalizarRutaYHacia(t *testing.T) {
	var c struct {
		Normalizar []struct{ Entrada, Sale string }
		Hacia      []struct {
			Entrada string
			Valido  bool
		}
	}
	if err := json.Unmarshal(casosRutas, &c); err != nil {
		t.Fatal(err)
	}
	for _, x := range c.Normalizar {
		if got := vidriera.NormalizarRuta(x.Entrada); got != x.Sale {
			t.Errorf("NormalizarRuta(%q) = %q, quería %q", x.Entrada, got, x.Sale)
		}
	}
	for _, x := range c.Hacia {
		if got := vidriera.HaciaValido(x.Entrada); got != x.Valido {
			t.Errorf("HaciaValido(%q) = %v", x.Entrada, got)
		}
	}
}

func TestRutaDeRedireccion(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	if err := b.App.EnTienda(ctx, luna, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `INSERT INTO redirecciones (tienda_id, desde, hacia, origen) VALUES ($1, '/productos/mate', '/producto/abc', 'importado:tiendanube')`, luna)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	vidriera.Rutas(mux, b.App, nil, "https://mostrador.test", nil, nil, nil)
	pedir := func(url string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, url, nil))
		return w
	}
	w := pedir("/v1/publico/tiendas/luna/redireccion?path=" + "%2Fproductos%2FMate%2F")
	if w.Code != http.StatusMovedPermanently || w.Header().Get("Location") != "/producto/abc" {
		t.Fatalf("redirección: %d %q", w.Code, w.Header().Get("Location"))
	}
	for _, u := range []string{"/v1/publico/tiendas/luna/redireccion?path=/otra", "/v1/publico/tiendas/sol/redireccion?path=/productos/mate",
		"/v1/publico/tiendas/luna/redireccion"} {
		if w := pedir(u); w.Code != http.StatusNotFound {
			t.Fatalf("%s: %d", u, w.Code)
		}
	}
}
