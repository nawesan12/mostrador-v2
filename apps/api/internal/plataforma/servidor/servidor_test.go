package servidor

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func TestSaludVivo(t *testing.T) {
	rec := httptest.NewRecorder()
	Nuevo(Dependencias{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/salud/vivo", nil))
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("estado=%d tipo=%q", rec.Code, rec.Header().Get("Content-Type"))
	}
}

func TestSaludListoConBase(t *testing.T) {
	b := pruebadb.Nueva(t)
	rec := httptest.NewRecorder()
	Nuevo(Dependencias{App: b.App}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/salud/listo", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("estado=%d %s", rec.Code, rec.Body)
	}
}

func TestRutaInexistente(t *testing.T) {
	rec := httptest.NewRecorder()
	Nuevo(Dependencias{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/no-existe", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("estado = %d, quería 404", rec.Code)
	}
}

func TestUnPanicDevuelveProblema(t *testing.T) {
	h := recuperar(silencio(), http.HandlerFunc(func(http.ResponseWriter, *http.Request) { panic("bug") }))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/x", nil))
	if rec.Code != 500 || rec.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("estado=%d tipo=%q", rec.Code, rec.Header().Get("Content-Type"))
	}
}

func silencio() *slog.Logger { return slog.New(slog.NewTextHandler(io.Discard, nil)) }
