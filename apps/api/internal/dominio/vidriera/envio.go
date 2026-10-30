package vidriera

import (
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

// unir junta las partes que vinieron con " · " (nada de separadores colgando).
func unir(partes ...string) string {
	out := make([]string, 0, len(partes))
	for _, p := range partes {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, " · ")
}

// cpValido: los 4 números del código postal argentino.
func cpValido(cp string) bool {
	cp = strings.TrimSpace(cp)
	if len(cp) != 4 {
		return false
	}
	for _, r := range cp {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// Cotizar las entregas propias de la tienda: el envío en moto (con su costo fijo, a la zona que puso
// la tienda; pide un CP porque va a una dirección) y el retiro en el local. El envío gratis de la
// promoción automática vale también para la moto. Correo Argentino no está acá: lo cotiza la API de
// MiCorreo (correo.go), nunca una tabla. Andreani no se ofrece: no es una integración (pide contrato).
func Cotizar(cp string, subtotal int64, e Envios, hoy time.Time) []Opcion {
	_ = hoy
	gratis := e.GratisDesde != nil && subtotal >= *e.GratisDesde
	var out []Opcion
	if e.Moto && cpValido(cp) {
		precio := e.MotoCosto
		if gratis {
			precio = 0
		}
		zona := ""
		if z := strings.TrimSpace(e.MotoZona); z != "" {
			zona = "Llega a " + z
		}
		out = append(out, Opcion{"moto", "Envío en moto", unir(zona, e.MotoDemora), precio})
	}
	if e.Retiro {
		detalle := unir(e.RetiroDireccion, e.RetiroHorario)
		if detalle == "" {
			detalle = "Te avisamos cuando esté listo"
		}
		out = append(out, Opcion{"retiro", "Retiro en el local", detalle, 0})
	}
	return out
}
