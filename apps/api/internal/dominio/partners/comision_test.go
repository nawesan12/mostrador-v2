package partners

import (
	"encoding/json"
	"os"
	"testing"
	"time"
)

// Los mismos casos corren en la web (apps/web/src/lib/partners/programa.spec.ts).
type casos struct {
	Comisiones []struct {
		Caso, Nivel, Plan, Ciclo, Alta, Mes string
		Comision                            int64
	}
	Niveles []struct {
		Asignado, Nivel string
		Activas         int
	}
}

func leerCasos(t *testing.T) casos {
	t.Helper()
	b, err := os.ReadFile("testdata/comisiones.json")
	if err != nil {
		t.Fatal(err)
	}
	var c casos
	if err := json.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}
	return c
}

func TestComisionPorNivelYDoceMeses(t *testing.T) {
	for _, c := range leerCasos(t).Comisiones {
		alta, err := time.Parse(time.RFC3339, c.Alta)
		if err != nil {
			t.Fatal(err)
		}
		mes, err := parsearMes(c.Mes)
		if err != nil {
			t.Fatal(err)
		}
		if got := Comision(c.Nivel, c.Plan, c.Ciclo, alta, mes); got != c.Comision {
			t.Errorf("%s: %d, quería %d", c.Caso, got, c.Comision)
		}
	}
}

func TestNivelEfectivo(t *testing.T) {
	for _, c := range leerCasos(t).Niveles {
		if got := NivelEfectivo(c.Asignado, c.Activas); got != c.Nivel {
			t.Errorf("%s con %d activas: %s, quería %s", c.Asignado, c.Activas, got, c.Nivel)
		}
	}
}

func TestElEjemploDeLaLanding(t *testing.T) {
	// "12 tiendas en plan Pro, nivel Plata, 25 % sobre $ 49.900 = $ 149.700 por mes".
	alta := time.Date(2026, 3, 10, 12, 0, 0, 0, zona)
	if got := 12 * Comision("plata", "pro", "mes", alta, alta); got != 149_700 {
		t.Fatalf("%d", got)
	}
}

func TestEstadoDeLaTienda(t *testing.T) {
	ahora := time.Date(2026, 10, 26, 12, 0, 0, 0, zona)
	for _, c := range []struct {
		plan string
		dias int
		baja bool
		want string
	}{
		{"pro", 3, false, EstadoPrueba},
		{"gratis", 13, false, EstadoPrueba},
		{"pro", 15, false, EstadoActiva},
		{"gratis", 15, false, EstadoGratis},
		{"pro", 200, true, EstadoBaja},
	} {
		if got := Estado(c.plan, ahora.AddDate(0, 0, -c.dias), c.baja, ahora); got != c.want {
			t.Errorf("%+v: %s", c, got)
		}
	}
}

func TestValidaciones(t *testing.T) {
	if !CuitValido("20123456786") || CuitValido("20123456780") || CuitValido("2012345678") {
		t.Fatal("cuit")
	}
	if _, err := validarCobro(Cobro{Tipo: "alias", Valor: "pampa.code.mp", Titular: "Pampa Code SRL", Cuit: "20-12345678-6", CondicionIva: "monotributo"}); err != nil {
		t.Fatal(err)
	}
	if _, err := validarCobro(Cobro{Tipo: "cbu", Valor: "123", Titular: "X", Cuit: "20123456786", CondicionIva: "monotributo"}); err == nil {
		t.Fatal("cbu corto")
	}
	if _, err := validarPerfil(Perfil{Nombre: "Pampa", Casos: []Caso{{"Bruma", "http://bruma.com"}}}); err == nil {
		t.Fatal("caso sin https")
	}
	if _, err := validarPerfil(Perfil{Nombre: "Pampa", Especialidades: []string{"magia"}}); err == nil {
		t.Fatal("especialidad")
	}
	p, err := validarPerfil(Perfil{Nombre: "  Pampa   Code ", Whatsapp: "+54 9 341 555-0123", Casos: []Caso{{"", ""}}})
	if err != nil || p.Nombre != "Pampa Code" || p.Whatsapp != "5493415550123" || len(p.Casos) != 0 {
		t.Fatalf("%+v %v", p, err)
	}
	if codigoDe("Estudio Brote & Co.") != "estudio-brote-y-co" || codigoDe("Ñu") != "nu0" {
		t.Fatal(codigoDe("Estudio Brote & Co."), codigoDe("Ñu"))
	}
}
