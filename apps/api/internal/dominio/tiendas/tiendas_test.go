package tiendas_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// Carpeta del contrato compartido con el front.
var contrato = filepath.Join("..", "..", "..", "..", "..", "packages", "contrato")

func leerJSON(t *testing.T, archivo string, v any) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(contrato, archivo))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatal(err)
	}
}

func TestRubrosIgualAlContrato(t *testing.T) {
	var esperados []tiendas.Rubro
	leerJSON(t, "rubros.json", &esperados)
	if !reflect.DeepEqual(esperados, tiendas.Rubros) {
		t.Fatalf("los rubros de Go no coinciden con packages/contrato/rubros.json")
	}
}

var uuid = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// El mismo caso que corre contra el sandbox de la demo (apps/web/src/lib/server/sandbox).
func TestContratoTienda(t *testing.T) {
	var caso struct {
		Siembra  struct{ Rubro, Slug string }
		Esperado map[string]any
		Forma    map[string]string
	}
	leerJSON(t, "casos/tienda.json", &caso)

	b := pruebadb.Nueva(t)
	r, _ := tiendas.BuscarRubro(caso.Siembra.Rubro)
	id, err := tiendas.Crear(context.Background(), b.App, caso.Siembra.Slug, r.Tienda, caso.Siembra.Rubro)
	if err != nil {
		t.Fatal(err)
	}

	mux := http.NewServeMux()
	tiendas.Rutas(mux, b.App, func(r *http.Request) string { return r.Header.Get("X-Tienda") })
	pedir := func(tienda string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/v1/panel/tienda", nil)
		req.Header.Set("X-Tienda", tienda)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		return rec
	}

	rec := pedir(id)
	if rec.Code != 200 {
		t.Fatalf("estado %d: %s", rec.Code, rec.Body)
	}
	var got map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	for k, v := range caso.Esperado {
		if got[k] != v {
			t.Errorf("%s = %v, quería %v", k, got[k], v)
		}
	}
	if !uuid.MatchString(got["id"].(string)) {
		t.Errorf("id %v no es uuid", got["id"])
	}
	if _, err := time.Parse(time.RFC3339Nano, got["creadaEn"].(string)); err != nil {
		t.Errorf("creadaEn: %v", err)
	}
	var claves, esperadas []string
	for k := range got {
		claves = append(claves, k)
	}
	for k := range caso.Esperado {
		esperadas = append(esperadas, k)
	}
	for k := range caso.Forma {
		esperadas = append(esperadas, k)
	}
	sort.Strings(claves)
	sort.Strings(esperadas)
	if !reflect.DeepEqual(claves, esperadas) {
		t.Errorf("campos %v, el contrato define %v", claves, esperadas)
	}

	t.Run("sin tienda es 401 problem+json", func(t *testing.T) {
		rec := pedir("")
		if rec.Code != 401 || rec.Header().Get("Content-Type") != "application/problem+json" {
			t.Fatalf("estado %d tipo %q", rec.Code, rec.Header().Get("Content-Type"))
		}
	})
	t.Run("una tienda que no existe es 401", func(t *testing.T) {
		if rec := pedir("00000000-0000-4000-8000-000000000000"); rec.Code != 401 {
			t.Fatalf("estado %d", rec.Code)
		}
	})
}

func TestRubroDesconocido(t *testing.T) {
	b := pruebadb.Nueva(t)
	if _, err := tiendas.Crear(context.Background(), b.App, "x", "X", "casino"); err == nil {
		t.Fatal("tendría que rechazar un rubro desconocido")
	}
}
