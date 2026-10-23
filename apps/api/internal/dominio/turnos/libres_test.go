package turnos_test

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/turnos"
)

// La tabla de casos vive en testdata/libres.json: la web (lib/turnos/libres.spec.ts) corre los
// mismos, así la tienda de la demo y la API ofrecen exactamente los mismos horarios.
type casoLibres struct {
	Nombre   string                `json:"nombre"`
	Entrada  turnos.Consulta       `json:"entrada"`
	Esperado []turnos.HorarioLibre `json:"esperado"`
}

func casos(t *testing.T) []casoLibres {
	t.Helper()
	b, err := os.ReadFile("testdata/libres.json")
	if err != nil {
		t.Fatal(err)
	}
	var doc struct {
		Casos []casoLibres `json:"casos"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Casos) < 15 {
		t.Fatalf("la tabla tiene %d casos", len(doc.Casos))
	}
	return doc.Casos
}

func TestLibres(t *testing.T) {
	for _, c := range casos(t) {
		t.Run(c.Nombre, func(t *testing.T) {
			got := turnos.Libres(c.Entrada)
			if c.Esperado == nil {
				c.Esperado = []turnos.HorarioLibre{}
			}
			if !reflect.DeepEqual(got, c.Esperado) {
				g, _ := json.Marshal(got)
				e, _ := json.Marshal(c.Esperado)
				t.Fatalf("\nobtuvo:  %s\nesperado: %s", g, e)
			}
		})
	}
}

// Asignar acepta exactamente los horarios que ofrece Libres, y con "cualquiera" elige el primer
// profesional libre por orden.
func TestAsignarCoincideConLibres(t *testing.T) {
	for _, c := range casos(t) {
		libres := map[string]bool{}
		for _, h := range c.Esperado {
			libres[h.Inicio] = true
		}
		dia, _ := time.ParseInLocation("2006-01-02", c.Entrada.Dia, turnos.Zona)
		for m := 0; m < 24*60; m += turnos.Paso {
			ti := dia.Add(time.Duration(m) * time.Minute)
			p := turnos.Asignar(c.Entrada, ti)
			if (p != "") != libres[ti.Format(time.RFC3339)] {
				t.Errorf("%s: Asignar(%s) = %q", c.Nombre, ti.Format("15:04"), p)
			}
		}
	}
	// El caso de "cualquiera": a las 9 Ana (orden 0) está ocupada y atiende Beto; a las 9:30, Ana.
	for _, c := range casos(t) {
		if c.Nombre != "cualquiera: sirve si algún profesional está libre" {
			continue
		}
		nueve := time.Date(2026, 10, 26, 9, 0, 0, 0, turnos.Zona)
		if p := turnos.Asignar(c.Entrada, nueve); p != "p-beto" {
			t.Fatalf("09:00: %q", p)
		}
		if p := turnos.Asignar(c.Entrada, nueve.Add(30*time.Minute)); p != "p-ana" {
			t.Fatalf("09:30: %q", p)
		}
		return
	}
	t.Fatal("falta el caso de cualquiera")
}

func TestSenaDe(t *testing.T) {
	for _, c := range []struct {
		precio int64
		pct    int
		sena   int64
	}{{12000, 30, 3600}, {9999, 30, 3000}, {9995, 10, 1000}, {5000, 0, 0}, {8000, 100, 8000}} {
		if got := turnos.SenaDe(c.precio, c.pct); got != c.sena {
			t.Errorf("SenaDe(%d, %d) = %d, quería %d", c.precio, c.pct, got, c.sena)
		}
	}
}
