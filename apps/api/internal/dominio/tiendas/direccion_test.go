package tiendas_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

type contratoDireccion struct {
	Reservados []string `json:"reservados"`
	Casos      []struct {
		Slug   string `json:"slug"`
		Motivo string `json:"motivo"`
	} `json:"casos"`
}

func leerContratoDireccion(t *testing.T) contratoDireccion {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(contrato, "direccion.json"))
	if err != nil {
		t.Fatal(err)
	}
	var c contratoDireccion
	if err := json.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestReservadosIgualAlContrato(t *testing.T) {
	c := leerContratoDireccion(t)
	a, b := slices.Clone(c.Reservados), slices.Clone(tiendas.Reservados)
	slices.Sort(a)
	slices.Sort(b)
	if !slices.Equal(a, b) {
		t.Fatalf("contrato %v\ngo %v", a, b)
	}
	// Las vitrinas (tiendas de ejemplo de cada rubro) tienen que estar reservadas.
	for _, r := range tiendas.Rubros {
		if !slices.Contains(tiendas.Reservados, tiendas.Slug(r.Tienda)) {
			t.Errorf("la vitrina %q no está reservada", tiendas.Slug(r.Tienda))
		}
	}
}

func TestDisponibilidadCasosDelContrato(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	id := pruebadb.CrearTienda(t, b.App, "la-esquina")
	for _, c := range leerContratoDireccion(t).Casos {
		d, err := tiendas.DisponibilidadDireccion(ctx, b.App, id, c.Slug)
		if err != nil {
			t.Fatal(err)
		}
		if d.Motivo != c.Motivo || d.Disponible != (c.Motivo == "libre") || d.Mensaje == "" {
			t.Errorf("%q: %+v, quería %s", c.Slug, d, c.Motivo)
		}
	}
	if d, _ := tiendas.DisponibilidadDireccion(ctx, b.App, id, "la-esquina"); d.Motivo != "actual" || d.Disponible {
		t.Errorf("la propia: %+v", d)
	}
}

func slugDe(t *testing.T, b *pruebadb.Bases, id string) string {
	t.Helper()
	ti, err := tiendas.Obtener(context.Background(), b.App, id)
	if err != nil {
		t.Fatal(err)
	}
	return ti.Slug
}

func TestCambiarDireccion(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	ana := pruebadb.CrearTienda(t, b.App, "la-tienda-de-ana")
	otra := pruebadb.CrearTienda(t, b.App, "otra")

	ant, err := tiendas.CambiarDireccion(ctx, b.App, ana, "cafe-de-ana")
	if err != nil || ant != "la-tienda-de-ana" {
		t.Fatalf("cambio: %q %v", ant, err)
	}
	if s := slugDe(t, b, ana); s != "cafe-de-ana" {
		t.Fatalf("slug %q", s)
	}
	// La misma otra vez: no cuenta como cambio y devuelve la misma.
	if ant, err := tiendas.CambiarDireccion(ctx, b.App, ana, "cafe-de-ana"); err != nil || ant != "cafe-de-ana" {
		t.Fatalf("la misma: %q %v", ant, err)
	}

	// La vieja: para otra tienda está usada y no la puede tomar; para Ana es suya.
	if d, _ := tiendas.DisponibilidadDireccion(ctx, b.App, otra, "la-tienda-de-ana"); d.Motivo != "usado" || d.Disponible {
		t.Errorf("la vieja para otra: %+v", d)
	}
	if _, err := tiendas.CambiarDireccion(ctx, b.App, otra, "la-tienda-de-ana"); !errors.Is(err, tiendas.ErrDireccionUsada) {
		t.Errorf("otra toma la vieja: %v", err)
	}
	if _, err := tiendas.CambiarDireccion(ctx, b.App, otra, "cafe-de-ana"); !errors.Is(err, tiendas.ErrDireccionUsada) {
		t.Errorf("otra toma la nueva: %v", err)
	}
	if d, _ := tiendas.DisponibilidadDireccion(ctx, b.App, ana, "la-tienda-de-ana"); d.Motivo != "propio" || !d.Disponible {
		t.Errorf("la vieja para Ana: %+v", d)
	}

	// Formato y reservados, sin tocar la base.
	if _, err := tiendas.CambiarDireccion(ctx, b.App, ana, "Cafe"); !errors.Is(err, tiendas.ErrDireccionInvalida) {
		t.Errorf("inválida: %v", err)
	}
	if _, err := tiendas.CambiarDireccion(ctx, b.App, ana, "panel"); !errors.Is(err, tiendas.ErrDireccionReservada) {
		t.Errorf("reservada: %v", err)
	}

	// La vieja redirige a la vigente; un slug que nunca existió, no.
	if s, err := vidriera.SlugVigente(ctx, b.App, "la-tienda-de-ana"); err != nil || s != "cafe-de-ana" {
		t.Errorf("vigente: %q %v", s, err)
	}
	if _, err := vidriera.SlugVigente(ctx, b.App, "nunca-existio"); !errors.Is(err, vidriera.ErrNoExiste) {
		t.Errorf("inexistente: %v", err)
	}
	// La vigente misma no es una dirección vieja.
	if _, err := vidriera.SlugVigente(ctx, b.App, "cafe-de-ana"); !errors.Is(err, vidriera.ErrNoExiste) {
		t.Errorf("vigente como vieja: %v", err)
	}

	// Otro cambio: las dos viejas llevan a la última.
	if _, err := tiendas.CambiarDireccion(ctx, b.App, ana, "ana-cafe"); err != nil {
		t.Fatal(err)
	}
	for _, vieja := range []string{"la-tienda-de-ana", "cafe-de-ana"} {
		if s, err := vidriera.SlugVigente(ctx, b.App, vieja); err != nil || s != "ana-cafe" {
			t.Errorf("%s: %q %v", vieja, s, err)
		}
	}
	// Volver a una vieja: pasa a ser la vigente y deja de redirigir.
	if _, err := tiendas.CambiarDireccion(ctx, b.App, ana, "la-tienda-de-ana"); err != nil {
		t.Fatal(err)
	}
	if _, err := vidriera.SlugVigente(ctx, b.App, "la-tienda-de-ana"); !errors.Is(err, vidriera.ErrNoExiste) {
		t.Errorf("la que volvió a ser vigente: %v", err)
	}
	if s, _ := vidriera.SlugVigente(ctx, b.App, "ana-cafe"); s != "la-tienda-de-ana" {
		t.Errorf("ana-cafe → %q", s)
	}
	// La tienda pública se encuentra por la nueva y no por la vieja.
	if _, err := vidriera.Publica(ctx, b.App, "ana-cafe"); !errors.Is(err, vidriera.ErrNoExiste) {
		t.Errorf("la vieja sigue siendo la tienda: %v", err)
	}
}

func TestCambiarDireccionTope(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	id := pruebadb.CrearTienda(t, b.App, "tope")
	otra := pruebadb.CrearTienda(t, b.App, "sin-tope")
	for i, s := range []string{"uno", "dos", "tres", "cuatro", "cinco"} {
		if _, err := tiendas.CambiarDireccion(ctx, b.App, id, s); err != nil {
			t.Fatalf("cambio %d: %v", i+1, err)
		}
	}
	if _, err := tiendas.CambiarDireccion(ctx, b.App, id, "seis"); !errors.Is(err, tiendas.ErrDireccionTope) {
		t.Fatalf("sexto: %v", err)
	}
	if s := slugDe(t, b, id); s != "cinco" {
		t.Fatalf("quedó %q", s)
	}
	// El tope es de cada tienda.
	if _, err := tiendas.CambiarDireccion(ctx, b.App, otra, "seis"); err != nil {
		t.Fatalf("otra tienda: %v", err)
	}
}

// Dos tiendas piden la misma dirección a la vez: una sola la consigue.
func TestCambiarDireccionCarrera(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	ids := []string{pruebadb.CrearTienda(t, b.App, "una"), pruebadb.CrearTienda(t, b.App, "dos")}
	errs := make([]error, len(ids))
	var wg sync.WaitGroup
	for i, id := range ids {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, errs[i] = tiendas.CambiarDireccion(ctx, b.App, id, "la-misma")
		}()
	}
	wg.Wait()
	ok := 0
	for _, err := range errs {
		switch {
		case err == nil:
			ok++
		case !errors.Is(err, tiendas.ErrDireccionUsada):
			t.Errorf("error inesperado: %v", err)
		}
	}
	if ok != 1 {
		t.Fatalf("la consiguieron %d tiendas", ok)
	}
}

func TestDireccionHTTP(t *testing.T) {
	b := pruebadb.Nueva(t)
	id := pruebadb.CrearTienda(t, b.App, "http-uno")
	pruebadb.CrearTienda(t, b.App, "http-dos")
	mux := http.NewServeMux()
	tiendas.Rutas(mux, b.App, func(*http.Request) string { return id })
	pedir := func(metodo, ruta, cuerpo string) (int, map[string]any) {
		t.Helper()
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo)))
		var m map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &m)
		return w.Code, m
	}
	if c, m := pedir("GET", "/v1/panel/tienda/direccion?slug=http-dos", ""); c != 200 || m["motivo"] != "usado" || m["disponible"] != false {
		t.Errorf("usada: %d %v", c, m)
	}
	if c, m := pedir("GET", "/v1/panel/tienda/direccion?slug=mi-local", ""); c != 200 || m["disponible"] != true {
		t.Errorf("libre: %d %v", c, m)
	}
	if c, m := pedir("PUT", "/v1/panel/tienda/direccion", `{"slug":"http-dos"}`); c != 409 || m["code"] != "usado" {
		t.Errorf("PUT usada: %d %v", c, m)
	}
	if c, m := pedir("PUT", "/v1/panel/tienda/direccion", `{"slug":"Mi Local"}`); c != 400 {
		t.Errorf("PUT inválida: %d %v", c, m)
	}
	if c, m := pedir("PUT", "/v1/panel/tienda/direccion", `{"slug":"panel"}`); c != 409 || m["code"] != "reservado" {
		t.Errorf("PUT reservada: %d %v", c, m)
	}
	if c, m := pedir("PUT", "/v1/panel/tienda/direccion", `{"slug":"mi-local"}`); c != 200 || m["slug"] != "mi-local" || m["anterior"] != "http-uno" {
		t.Errorf("PUT: %d %v", c, m)
	}
}

// La app no toca las direcciones viejas ni el slug de la tienda sino por las funciones, y las
// funciones sólo actúan sobre la tienda fijada en la transacción.
func TestDireccionRLS(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	ana := pruebadb.CrearTienda(t, b.App, "ana")
	otra := pruebadb.CrearTienda(t, b.App, "otra")
	if _, err := tiendas.CambiarDireccion(ctx, b.App, ana, "ana-nueva"); err != nil {
		t.Fatal(err)
	}
	prohibido := func(nombre, sql string, args ...any) {
		t.Helper()
		err := b.App.EnTienda(ctx, otra, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, sql, args...)
			return err
		})
		if err == nil {
			t.Errorf("%s: la app pudo", nombre)
		}
	}
	prohibido("leer anteriores", `SELECT * FROM app.slugs_anteriores`)
	prohibido("robar una vieja", `INSERT INTO app.slugs_anteriores (slug, tienda_id) VALUES ('ana', $1)`, otra)
	prohibido("borrar una vieja", `DELETE FROM app.slugs_anteriores`)
	prohibido("leer cambios", `SELECT * FROM app.slugs_cambios`)
	prohibido("borrar el tope", `DELETE FROM app.slugs_cambios`)
	prohibido("liberar un slug", `DELETE FROM app.slugs WHERE slug = 'ana'`)
	prohibido("cambiar el slug a mano", `UPDATE tiendas SET slug = 'x-otra' WHERE id = $1`, otra)

	// Sin tienda fijada, cambiar no hace nada.
	err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `SELECT app.cambiar_slug('sin-tienda')`)
		return err
	})
	if err == nil {
		t.Error("cambiar_slug sin tienda anduvo")
	}
	// Con la tienda de otra fijada, sólo cambia la de otra: la de Ana sigue igual.
	if _, err := tiendas.CambiarDireccion(ctx, b.App, otra, "otra-nueva"); err != nil {
		t.Fatal(err)
	}
	if s := slugDe(t, b, ana); s != "ana-nueva" {
		t.Errorf("Ana quedó en %q", s)
	}
}
