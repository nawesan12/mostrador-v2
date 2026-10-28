package equipo_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/equipo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

const clave = "una clave larga del equipo"

// alta da de alta a alguien del equipo como `mostradorctl equipo-alta` (con el superusuario).
func alta(t *testing.T, b *pruebadb.Bases, email, rol string) {
	t.Helper()
	hash, err := cuentas.HashClave(clave)
	if err != nil {
		t.Fatal(err)
	}
	conn, err := pgx.Connect(context.Background(), b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(context.Background())
	if _, err := conn.Exec(context.Background(),
		`INSERT INTO equipo.miembros (email, nombre, clave_hash, rol) VALUES ($1, $2, $3, $4)`,
		email, "Persona "+rol, hash, rol); err != nil {
		t.Fatal(err)
	}
}

type api struct {
	t *testing.T
	h http.Handler
}

func (a api) pedir(metodo, ruta, token string, cuerpo any) (int, map[string]any) {
	a.t.Helper()
	var r *http.Request
	if cuerpo != nil {
		js, _ := json.Marshal(cuerpo)
		r = httptest.NewRequest(metodo, ruta, bytes.NewReader(js))
	} else {
		r = httptest.NewRequest(metodo, ruta, nil)
	}
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	a.h.ServeHTTP(w, r)
	var v map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &v)
	return w.Code, v
}

func (a api) ingresar(email string) string {
	a.t.Helper()
	c, v := a.pedir("POST", "/v1/equipo/ingresar", "", map[string]string{"email": email, "clave": clave})
	if c != http.StatusOK {
		a.t.Fatalf("ingresar %s: %d %v", email, c, v)
	}
	return v["token"].(string)
}

func TestPermisosYAuditoria(t *testing.T) {
	b := pruebadb.Nueva(t)
	tienda := pruebadb.CrearTienda(t, b.App, "luna")
	alta(t, b, "soporte@mostradorapp.com.ar", "soporte")
	alta(t, b, "duenio@mostradorapp.com.ar", "duenio")
	mux := http.NewServeMux()
	(&equipo.Servicio{Staff: b.Staff, App: b.App, Sitio: "https://mostradorapp.com.ar"}).Rutas(mux)
	a := api{t, mux}

	// Una clave incorrecta, un email que no es del equipo y un token inventado no entran.
	if c, _ := a.pedir("POST", "/v1/equipo/ingresar", "", map[string]string{"email": "soporte@mostradorapp.com.ar", "clave": "otra"}); c != http.StatusUnauthorized {
		t.Fatalf("clave mala: %d", c)
	}
	if c, _ := a.pedir("POST", "/v1/equipo/ingresar", "", map[string]string{"email": "nadie@x.com", "clave": clave}); c != http.StatusUnauthorized {
		t.Fatalf("email ajeno: %d", c)
	}
	if c, _ := a.pedir("GET", "/v1/equipo/inicio", "inventado", nil); c != http.StatusUnauthorized {
		t.Fatalf("token inventado: %d", c)
	}

	soporte := a.ingresar("soporte@mostradorapp.com.ar")
	duenio := a.ingresar("duenio@mostradorapp.com.ar")

	// Soporte ve tiendas y extiende pruebas, pero no ve facturación ni cambia planes.
	if c, _ := a.pedir("GET", "/v1/equipo/tiendas/"+tienda, soporte, nil); c != http.StatusOK {
		t.Fatalf("soporte ve la ficha: %d", c)
	}
	if c, _ := a.pedir("GET", "/v1/equipo/facturacion", soporte, nil); c != http.StatusForbidden {
		t.Fatalf("soporte no ve facturación: %d", c)
	}
	if c, _ := a.pedir("POST", "/v1/equipo/tiendas/"+tienda+"/acciones", soporte, map[string]any{"accion": "cambiar_plan", "plan": "pro"}); c != http.StatusForbidden {
		t.Fatalf("soporte no cambia planes: %d", c)
	}
	if c, v := a.pedir("POST", "/v1/equipo/tiendas/"+tienda+"/acciones", soporte, map[string]any{"accion": "extender_prueba", "dias": 7}); c != http.StatusNoContent {
		t.Fatalf("soporte extiende la prueba: %d %v", c, v)
	}
	if c, _ := a.pedir("POST", "/v1/equipo/tiendas/"+tienda+"/acciones", soporte, map[string]any{"accion": "extender_prueba", "dias": 90}); c != http.StatusBadRequest {
		t.Fatalf("más de 30 días: %d", c)
	}
	if c, _ := a.pedir("GET", "/v1/equipo/facturacion", duenio, nil); c != http.StatusOK {
		t.Fatalf("el dueño ve facturación: %d", c)
	}

	// La auditoría tiene lo que soporte vio y lo que hizo, con la tienda.
	c, v := a.pedir("GET", "/v1/equipo/auditoria", duenio, nil)
	if c != http.StatusOK {
		t.Fatalf("auditoría: %d", c)
	}
	var acciones []string
	for _, r := range v["registros"].([]any) {
		reg := r.(map[string]any)
		if reg["tiendaId"] == tienda && reg["miembro"].(map[string]any)["nombre"] == "Persona soporte" {
			acciones = append(acciones, reg["accion"].(string))
		}
	}
	if len(acciones) < 2 {
		t.Fatalf("auditoría de soporte en la tienda: %v", acciones)
	}
	if c, _ := a.pedir("GET", "/v1/equipo/auditoria", soporte, nil); c != http.StatusForbidden {
		t.Fatalf("soporte no ve la auditoría: %d", c)
	}

	// Salir corta la sesión.
	if c, _ := a.pedir("POST", "/v1/equipo/salir", soporte, nil); c != http.StatusNoContent {
		t.Fatalf("salir: %d", c)
	}
	if c, _ := a.pedir("GET", "/v1/equipo/sesion", soporte, nil); c != http.StatusUnauthorized {
		t.Fatalf("después de salir: %d", c)
	}
}

// La matriz de permisos de Go es la misma que la de testdata (y la de la web y la de SQL).
func TestMatrizDePermisos(t *testing.T) {
	js, err := os.ReadFile("testdata/permisos.json")
	if err != nil {
		t.Fatal(err)
	}
	var m map[string][]string
	if err := json.Unmarshal(js, &m); err != nil {
		t.Fatal(err)
	}
	for rol, ps := range m {
		got := slices.Clone(equipo.Permisos[rol])
		slices.Sort(got)
		slices.Sort(ps)
		if !slices.Equal(got, ps) {
			t.Fatalf("%s: Go %v, testdata %v", rol, got, ps)
		}
	}
	if len(m) != len(equipo.Permisos) {
		t.Fatalf("roles: Go %d, testdata %d", len(equipo.Permisos), len(m))
	}
}
