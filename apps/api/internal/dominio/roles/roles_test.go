package roles

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// Todas las rutas de /v1/panel/ que se registran en el código tienen su permiso en la tabla: una
// ruta nueva sin permiso rompe acá (y mientras tanto, sólo la usa la dueña).
func TestCadaRutaDelPanelTienePermiso(t *testing.T) {
	re := regexp.MustCompile(`"((?:GET|POST|PUT|PATCH|DELETE) /v1/panel/[^"]*)"`)
	enCodigo := map[string]bool{}
	err := filepath.WalkDir("../..", func(ruta string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(ruta, ".go") || strings.HasSuffix(ruta, "_test.go") ||
			strings.Contains(ruta, string(filepath.Separator)+"roles"+string(filepath.Separator)) {
			return err
		}
		b, err := os.ReadFile(ruta)
		if err != nil {
			return err
		}
		for _, m := range re.FindAllSubmatch(b, -1) {
			if strings.Contains(string(b), "HandleFunc(\""+string(m[1])) || strings.Contains(string(b), "Handle(\""+string(m[1])) {
				enCodigo[string(m[1])] = true
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(enCodigo) < 100 {
		t.Fatalf("se encontraron sólo %d rutas: ¿cambió la forma de registrarlas?", len(enCodigo))
	}
	for r := range enCodigo {
		if _, ok := Rutas[r]; !ok {
			t.Errorf("la ruta %q no tiene permiso en roles.Rutas", r)
		}
	}
}

func TestPorRolIgualQueLaWeb(t *testing.T) {
	b, err := os.ReadFile("testdata/roles.json")
	if err != nil {
		t.Fatal(err)
	}
	var esperado map[string][]Permiso
	if err := json.Unmarshal(b, &esperado); err != nil {
		t.Fatal(err)
	}
	if len(esperado) != len(porRol) {
		t.Fatalf("roles: %v", esperado)
	}
	for rol, ps := range esperado {
		if !slices.Equal(ps, porRol[rol]) {
			t.Errorf("%s: testdata %v, Go %v", rol, ps, porRol[rol])
		}
	}
}

func TestGuardia(t *testing.T) {
	llego := false
	h := Guardia(func(r *http.Request) (string, bool) {
		rol := r.Header.Get("X-Rol")
		return rol, rol != ""
	}, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { llego = true; w.WriteHeader(http.StatusNoContent) }))
	casos := []struct {
		rol, metodo, ruta string
		pasa              bool
	}{
		{"local", "GET", "/v1/panel/pedidos", true},
		{"local", "POST", "/v1/panel/pedidos/12/estado", true},
		{"local", "GET", "/v1/panel/listas/productos", true},
		{"local", "GET", "/v1/panel/listas/clientes", false},
		{"local", "GET", "/v1/panel/listas/facturacion", false},
		{"ventas", "GET", "/v1/panel/listas/clientes", true},
		{"ventas", "PUT", "/v1/panel/productos/abc", true},
		{"ventas", "POST", "/v1/panel/productos", false},
		{"ventas", "GET", "/v1/panel/finanzas", false},
		{"ventas", "PUT", "/v1/panel/configuracion/negocio", false},
		{"ventas", "PUT", "/v1/panel/configuracion/seguridad/inactivas", true},
		{"ventas", "PUT", "/v1/panel/configuracion/avisos/venta", false},
		{"admin", "GET", "/v1/panel/finanzas", true},
		{"admin", "PUT", "/v1/panel/configuracion/negocio", true},
		{"admin", "POST", "/v1/panel/equipo", false},
		{"admin", "PUT", "/v1/panel/plan", false},
		{"admin", "GET", "/v1/panel/ruta-que-no-existe", false},
		{"duenio", "GET", "/v1/panel/ruta-que-no-existe", true},
		{"duenio", "POST", "/v1/panel/equipo", true},
		{"", "GET", "/v1/panel/finanzas", true}, // sin sesión: responde el handler (401)
		{"local", "GET", "/v1/rubros", true},
	}
	for _, c := range casos {
		llego = false
		r := httptest.NewRequest(c.metodo, c.ruta, nil)
		r.Header.Set("X-Rol", c.rol)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		if llego != c.pasa || (!c.pasa && rec.Code != http.StatusForbidden) {
			t.Errorf("%s %s %s: pasó=%v código=%d", c.rol, c.metodo, c.ruta, llego, rec.Code)
		}
	}
}
