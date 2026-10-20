package problema

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestEscribir(t *testing.T) {
	rec := httptest.NewRecorder()
	Escribir(rec, Invalido("Falta el nombre.", Campo{Ruta: "nombre", Mensaje: "Obligatorio"}))
	if rec.Code != 400 {
		t.Fatalf("estado %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/problem+json" {
		t.Fatalf("Content-Type %q", ct)
	}
	var p Problema
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatal(err)
	}
	if p.Codigo != "invalido" || p.Tipo != "about:blank" || len(p.Campos) != 1 || p.Campos[0].Ruta != "nombre" {
		t.Fatalf("problema mal armado: %+v", p)
	}
}
