package servidor

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSaludVivo(t *testing.T) {
	rec := httptest.NewRecorder()
	Nuevo().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/salud/vivo", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("estado = %d, quería 200", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q", got)
	}
}

func TestRutaInexistente(t *testing.T) {
	rec := httptest.NewRecorder()
	Nuevo().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/no-existe", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("estado = %d, quería 404", rec.Code)
	}
}
