package vidriera_test

import (
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
)

func TestAplicarCuponIgualQueLaWeb(t *testing.T) {
	b, err := os.ReadFile("testdata/cupones.json")
	if err != nil {
		t.Fatal(err)
	}
	var d struct {
		Ahora time.Time `json:"ahora"`
		Casos []struct {
			Nombre    string                  `json:"nombre"`
			Cupon     vidriera.CuponTienda    `json:"cupon"`
			Lineas    []vidriera.LineaCupon   `json:"lineas"`
			Historial vidriera.Historial      `json:"historial"`
			Esperado  vidriera.ResultadoCupon `json:"esperado"`
		} `json:"casos"`
	}
	if err := json.Unmarshal(b, &d); err != nil {
		t.Fatal(err)
	}
	for _, c := range d.Casos {
		t.Run(c.Nombre, func(t *testing.T) {
			if got := vidriera.AplicarCupon(c.Cupon, c.Lineas, c.Historial, d.Ahora, "web"); got != c.Esperado {
				t.Fatalf("got %+v, quería %+v", got, c.Esperado)
			}
		})
	}
}
