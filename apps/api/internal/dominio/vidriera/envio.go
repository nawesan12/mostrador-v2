package vidriera

import (
	"strconv"
	"strings"
	"time"
)

// Opcion de envío cotizada (lib/tienda/envio.ts hace lo mismo para mostrarlo: mantenerlos iguales).
type Opcion struct {
	ID      string `json:"id"`
	Nombre  string `json:"nombre"`
	Detalle string `json:"detalle"`
	Precio  int64  `json:"precio"`
}

// Cotizar las entregas propias de la tienda para un código postal: la moto (CABA) y el retiro. Correo
// Argentino no está acá: lo cotiza la API de MiCorreo (correo.go), nunca una tabla. Andreani no se
// ofrece: no es una integración (pide contrato).
func Cotizar(cp string, subtotal int64, e Envios, hoy time.Time) []Opcion {
	_, _ = subtotal, hoy
	n, err := strconv.Atoi(strings.TrimSpace(cp))
	if err != nil || len(strings.TrimSpace(cp)) != 4 {
		n = -1
	}
	var out []Opcion
	if e.Moto && n >= 1000 && n <= 1499 {
		out = append(out, Opcion{"moto", "Moto en el día", "Si comprás antes de las 16 h", 4500})
	}
	if e.Retiro {
		out = append(out, Opcion{"retiro", "Retiro en el local", "Te avisamos cuando esté listo", 0})
	}
	return out
}
