package vidriera

import (
	"fmt"
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

// Hasta conectar Andreani y Correo Argentino, una tabla por región.
var regiones = []struct {
	desde, hasta int
	andreani     int64
	dias         int
}{{1000, 1999, 6900, 2}, {2000, 3999, 7900, 3}, {4000, 5999, 8900, 4}, {6000, 7999, 7500, 3}, {8000, 9499, 9900, 5}, {9500, 9999, 11900, 6}}

var diasSemana = []string{"domingo", "lunes", "martes", "miércoles", "jueves", "viernes", "sábado"}

func llega(dias int, desde time.Time) string {
	d := desde
	for habiles := 0; habiles < dias; {
		d = d.AddDate(0, 0, 1)
		if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
			habiles++
		}
	}
	return fmt.Sprintf("Llega el %s %d", diasSemana[d.Weekday()], d.Day())
}

// Cotizar el envío para un código postal.
func Cotizar(cp string, subtotal int64, e Envios, hoy time.Time) []Opcion {
	n, err := strconv.Atoi(strings.TrimSpace(cp))
	if err != nil || len(strings.TrimSpace(cp)) != 4 {
		n = -1
	}
	gratis := e.GratisDesde != nil && subtotal >= *e.GratisDesde
	var out []Opcion
	for _, r := range regiones {
		if n < r.desde || n > r.hasta {
			continue
		}
		precio := r.andreani
		if gratis {
			precio = 0
		}
		out = append(out, Opcion{"andreani", "Andreani a domicilio", llega(r.dias, hoy), precio})
		if e.Correo {
			c := (r.andreani*8/10 + 50) / 100 * 100
			if gratis {
				c = 0
			}
			out = append(out, Opcion{"correo", "Correo Argentino a sucursal", llega(r.dias+2, hoy), c})
		}
		if e.Moto && n >= 1000 && n <= 1499 {
			out = append(out, Opcion{"moto", "Moto en el día", "Si comprás antes de las 16 h", 4500})
		}
	}
	if e.Retiro {
		out = append(out, Opcion{"retiro", "Retiro en el local", "Te avisamos cuando esté listo", 0})
	}
	return out
}
