package vidriera

import (
	"encoding/json"
	"os"
	"testing"
)

// Los mismos casos que la web (packages/contrato/casos/combinar.json).
func TestCombinarDescuentos(t *testing.T) {
	b, err := os.ReadFile("../../../../../packages/contrato/casos/combinar.json")
	if err != nil {
		t.Fatal(err)
	}
	var c struct {
		Casos []struct {
			Nombre                 string
			Subtotal, Combo, Cupon int64
			Combinable             bool
			Esperado               [2]int64
		}
	}
	if err := json.Unmarshal(b, &c); err != nil {
		t.Fatal(err)
	}
	for _, x := range c.Casos {
		combo, cupon := combinarDescuentos(x.Subtotal, x.Combo, x.Cupon, x.Combinable)
		if combo != x.Esperado[0] || cupon != x.Esperado[1] {
			t.Errorf("%s: %d %d, quería %v", x.Nombre, combo, cupon, x.Esperado)
		}
	}
}
