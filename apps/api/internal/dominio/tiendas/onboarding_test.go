package tiendas_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// La lista de Go tiene que ser la misma que la del front (que sale del handoff).
func TestPlantillasIgualAlFront(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "web", "src", "lib", "plantillas.ts"))
	if err != nil {
		t.Fatal(err)
	}
	var front []string
	for _, m := range regexp.MustCompile(`(?m)^\t\tid: '([a-z-]+)'`).FindAllStringSubmatch(string(b), -1) {
		front = append(front, m[1])
	}
	a, c := slices.Clone(front), slices.Clone(tiendas.Plantillas)
	slices.Sort(a)
	slices.Sort(c)
	if !slices.Equal(a, c) {
		t.Fatalf("front %v\ngo %v", a, c)
	}
}

func TestGuardarOnboarding(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	una, err := tiendas.Crear(ctx, b.App, "la-esquina", "La Esquina", "gastronomia")
	if err != nil {
		t.Fatal(err)
	}
	otra, err := tiendas.Crear(ctx, b.App, "otra", "Otra", "moda")
	if err != nil {
		t.Fatal(err)
	}
	el := json.RawMessage(`{"tono":"cercano","objetivos":[1,2]}`)
	if err := tiendas.GuardarOnboarding(ctx, b.App, una, tiendas.Onboarding{Plantilla: "yerbal", Elecciones: el, Terminado: true}); err != nil {
		t.Fatal(err)
	}
	tu, _ := tiendas.Obtener(ctx, b.App, una)
	to, _ := tiendas.Obtener(ctx, b.App, otra)
	if tu.Plantilla != "yerbal" || to.Plantilla == "yerbal" {
		t.Fatalf("plantillas: %s / %s", tu.Plantilla, to.Plantilla)
	}
	var tono string
	var terminado bool
	if err := b.Staff.QueryRow(ctx, `SELECT onboarding->>'tono', onboarding_terminado_en IS NOT NULL FROM tiendas WHERE id = $1`, una).Scan(&tono, &terminado); err != nil {
		t.Fatal(err)
	}
	if tono != "cercano" || !terminado {
		t.Fatalf("tono %q terminado %v", tono, terminado)
	}
	if err := tiendas.GuardarOnboarding(ctx, b.App, una, tiendas.Onboarding{Plantilla: "inventada"}); !errors.Is(err, tiendas.ErrPlantilla) {
		t.Fatalf("err = %v", err)
	}
	// El rubro elegido pasa a la tienda; uno desconocido no pisa el que tenía.
	if err := tiendas.GuardarOnboarding(ctx, b.App, una, tiendas.Onboarding{Elecciones: json.RawMessage(`{"rubro":"belleza"}`)}); err != nil {
		t.Fatal(err)
	}
	if tu, _ = tiendas.Obtener(ctx, b.App, una); tu.Rubro != "belleza" {
		t.Fatalf("rubro %q", tu.Rubro)
	}
	if err := tiendas.GuardarOnboarding(ctx, b.App, una, tiendas.Onboarding{Elecciones: json.RawMessage(`{"rubro":"inventado"}`)}); err != nil {
		t.Fatal(err)
	}
	if tu, _ = tiendas.Obtener(ctx, b.App, una); tu.Rubro != "belleza" {
		t.Fatalf("rubro pisado: %q", tu.Rubro)
	}
}

func TestResumenSinDatosTodavia(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	id, err := tiendas.Crear(ctx, b.App, "resumen", "Resumen", "moda")
	if err != nil {
		t.Fatal(err)
	}
	r, err := tiendas.ObtenerResumen(ctx, b.App, id, time.Now().Add(5*24*time.Hour+time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if r.Prueba.DiasRestantes != 9 || r.Prueba.DiasTotales != 14 || r.Insignias.Pedidos != 0 || r.Notificaciones == nil {
		t.Fatalf("%+v", r)
	}
}
