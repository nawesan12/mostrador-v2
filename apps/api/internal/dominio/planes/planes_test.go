package planes

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/abono"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/roles"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// La tabla es la misma que la de la web (apps/web/src/lib/panel/planes.ts lee testdata/planes.json).
// ACTUALIZAR=1 la reescribe.
func TestIgualQueLaWeb(t *testing.T) {
	if os.Getenv("ACTUALIZAR") == "1" {
		b, _ := json.MarshalIndent(Todos, "", "  ")
		if err := os.WriteFile("testdata/planes.json", append(b, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	b, err := os.ReadFile("testdata/planes.json")
	if err != nil {
		t.Fatal(err)
	}
	var esperado []Plan
	if err := json.Unmarshal(b, &esperado); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(esperado, Todos) {
		t.Fatalf("testdata/planes.json no es igual a planes.Todos (ACTUALIZAR=1 la reescribe)")
	}
}

func TestPreciosIgualQueAbono(t *testing.T) {
	for _, p := range Todos {
		m, ok := abono.PrecioMensual(p.ID)
		if !ok || m != p.Mensual {
			t.Errorf("%s: abono %d, planes %d", p.ID, m, p.Mensual)
		}
		if a, _ := abono.Precio(p.ID, "anio"); p.ID != "gratis" && a != p.Mensual*10 {
			t.Errorf("%s: el anual es %d, no 10 meses", p.ID, a)
		}
		if abono.Nombres[p.ID] != p.Nombre {
			t.Errorf("%s: nombre %q en abono", p.ID, abono.Nombres[p.ID])
		}
	}
}

func TestLoAcordado(t *testing.T) {
	g, e, c, n := De("gratis"), De("emprende"), De("pro"), De("negocio")
	casos := []struct {
		nombre string
		ok     bool
	}{
		{"gratis 20 productos, 1 persona, 3 fotos, 50 de Tincho", g.Productos == 20 && g.Personas == 1 && g.Fotos == 3 && g.Tincho == 50},
		{"gratis sin dominio ni nada pago", len(g.Funciones) == 0},
		{"emprende $14.900, 150, 2, dominio, ARCA, Correo, cupones, 500", e.Mensual == 14_900 && e.Productos == 150 && e.Personas == 2 &&
			e.Tiene(Dominio) && e.Tiene(ARCA) && e.Tiene(Correo) && e.Tiene(Cupones) && !e.Tiene(Roles) && e.Tincho == 500 && e.Fotos == 5},
		{"crece $34.900, 500, 5, roles, local, ML, newsletter, 2.000", c.Nombre == "Crece" && c.Mensual == 34_900 && c.Productos == 500 &&
			c.Personas == 5 && c.Tiene(Roles) && c.Tiene(Local) && c.Tiene(MercadoLibre) && c.Tiene(Newsletter) && !c.Tiene(Mayorista) &&
			c.Tincho == 2_000 && c.Fotos == 8},
		{"negocio $79.900, sin tope, mayorista, 8.000, 10 fotos", n.Mensual == 79_900 && n.Productos == SinTope && n.Personas == SinTope &&
			n.Tiene(Mayorista) && n.Tincho == 8_000 && n.Fotos == 10},
		{"pack +2.000 por $4.900", PackTincho == 2_000 && PrecioPackTincho == 4_900},
		{"cada plan incluye lo del anterior", func() bool {
			for i := 1; i < len(Todos); i++ {
				for _, f := range Todos[i-1].Funciones {
					if !Todos[i].Tiene(f) {
						return false
					}
				}
			}
			return true
		}()},
	}
	for _, c := range casos {
		if !c.ok {
			t.Error(c.nombre)
		}
	}
}

func TestPruebaRigeComoCrece(t *testing.T) {
	ahora := time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)
	if p := Efectivo("gratis", ahora.Add(time.Hour), ahora); p.ID != "pro" {
		t.Fatalf("en la prueba: %s", p.ID)
	}
	if p := Efectivo("gratis", ahora.Add(-time.Hour), ahora); p.ID != "gratis" {
		t.Fatalf("después de la prueba: %s", p.ID)
	}
	// Un plan pago rige como es, aun en la prueba.
	if p := Efectivo("emprende", ahora.Add(time.Hour), ahora); p.ID != "emprende" {
		t.Fatalf("emprende en la prueba: %s", p.ID)
	}
}

func TestTopes(t *testing.T) {
	g := De("gratis")
	casos := []struct {
		err       error
		que, sube string
		texto     string
	}{
		{ExigirProductos(g, 21), "productos", "emprende", "hasta 20 productos"},
		{ExigirProductos(De("emprende"), 151), "productos", "pro", "Con Crece cargás hasta 500"},
		{ExigirProductos(De("pro"), 501), "productos", "negocio", "sin tope"},
		{ExigirPersonas(g, 2), "personas", "emprende", "1 persona: vos"},
		{ExigirPersonas(De("emprende"), 3), "personas", "pro", "hasta 5"},
		{ExigirFotos(g, 4), "fotos", "emprende", "3 fotos por producto"},
		{ExigirFuncion(g, Dominio), "dominio", "emprende", "El dominio propio está desde el plan Emprende"},
		{ExigirFuncion(De("emprende"), MercadoLibre), "mercadolibre", "pro", "desde el plan Crece"},
		{ExigirFuncion(De("pro"), Mayorista), "mayorista", "negocio", "desde el plan Negocio"},
	}
	for _, c := range casos {
		p := problema.De(c.err)
		if p == nil || p.Estado != http.StatusPaymentRequired || p.Codigo != "tope_plan" || p.Funcion != c.que || p.Plan != c.sube ||
			!strings.Contains(p.Detalle, c.texto) {
			t.Errorf("%v → %+v", c.err, p)
		}
		var tope *Tope
		if !errors.As(c.err, &tope) {
			t.Errorf("no es un *Tope: %v", c.err)
		}
	}
	for _, err := range []error{ExigirProductos(g, 20), ExigirPersonas(g, 1), ExigirFotos(g, 3), ExigirProductos(De("negocio"), 100_000),
		ExigirFuncion(De("negocio"), Mayorista)} {
		if err != nil {
			t.Errorf("no debía topar: %v", err)
		}
	}
}

// Cada ruta con función del plan existe (está en la tabla de permisos, que un test compara con el
// código): un patrón mal escrito dejaría la función sin tope.
func TestRutasExisten(t *testing.T) {
	for r := range Rutas {
		if _, ok := roles.Rutas[r]; !ok {
			t.Errorf("%s no es una ruta del panel", r)
		}
	}
}
