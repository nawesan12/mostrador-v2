package tiendas_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func TestVentana(t *testing.T) {
	ahora := time.Date(2026, 10, 20, 2, 30, 0, 0, time.UTC) // 23:30 del 19 en Argentina
	desde, hasta, err := tiendas.Ventana("hoy", ahora)
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Date(2026, 10, 19, 3, 0, 0, 0, time.UTC); !desde.Equal(want) || !hasta.Equal(ahora) {
		t.Fatalf("hoy: %s – %s", desde, hasta)
	}
	if d, _, _ := tiendas.Ventana("7d", ahora); !d.Equal(ahora.Add(-7 * 24 * time.Hour)) {
		t.Fatalf("7d: %s", d)
	}
	if _, _, err := tiendas.Ventana("anual", ahora); err == nil {
		t.Fatal("aceptó un período inválido")
	}
}

func TestMetaPorTienda(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	una, _ := tiendas.Crear(ctx, b.App, "una", "Una", "moda")
	otra, _ := tiendas.Crear(ctx, b.App, "otra", "Otra", "moda")
	if err := tiendas.GuardarMeta(ctx, b.App, una, 20_000_000); err != nil {
		t.Fatal(err)
	}
	if err := tiendas.GuardarMeta(ctx, b.App, una, -1); !errors.Is(err, tiendas.ErrMeta) {
		t.Fatalf("meta negativa: %v", err)
	}
	tu, err := tiendas.ObtenerTablero(ctx, b.App, una, "7d", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	to, _ := tiendas.ObtenerTablero(ctx, b.App, otra, "7d", time.Now())
	if tu.Meta.Objetivo != 20_000_000 || to.Meta.Objetivo != 0 {
		t.Fatalf("metas: %d / %d", tu.Meta.Objetivo, to.Meta.Objetivo)
	}
	if len(tu.Serie.Actual) != 14 || tu.UltimosPedidos == nil || tu.Avisos.CertificadoDias != nil {
		t.Fatalf("tablero vacío mal armado: %+v", tu)
	}
}
