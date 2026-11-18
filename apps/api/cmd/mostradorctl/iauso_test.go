package main

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/ia"
)

// ia-uso lee con el rol staff (todas las tiendas) sin romperse.
func TestIAUso(t *testing.T) {
	b := pruebadb.Nueva(t)
	ahora := time.Date(2026, 11, 18, 15, 0, 0, 0, time.UTC)
	for _, slug := range []string{"una", "otra"} {
		id := pruebadb.CrearTienda(t, b.App, slug)
		if err := ia.Registrar(context.Background(), b.App, id, ia.RutaChat, ia.Uso{Entrada: 1000, Salida: 100}, ahora); err != nil {
			t.Fatal(err)
		}
	}
	u, _ := url.Parse(b.AdminURL)
	u.User = url.User("mostrador_staff")
	if err := iaUso(context.Background(), u.String(), "2026-11"); err != nil {
		t.Fatal(err)
	}
}
