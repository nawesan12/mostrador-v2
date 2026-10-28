package partners

import (
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadopago"
)

// El programa publicado en /partners: Partner 20 %, Plata 25 % (desde 5 tiendas activas) y Oro 30 %
// (desde 20), recurrente durante 12 meses desde el alta de cada tienda. La misma regla vive en
// apps/web/src/lib/partners/programa.ts, con los mismos casos (testdata/comisiones.json).

// Niveles, de menor a mayor.
var Niveles = []string{"partner", "plata", "oro"}

var porcentaje = map[string]int64{"partner": 20, "plata": 25, "oro": 30}

// TiendasParaNivel: las tiendas activas que pide cada nivel.
var TiendasParaNivel = map[string]int{"partner": 0, "plata": 5, "oro": 20}

// MesesDeComision: cuántos meses desde el alta de la tienda cobra el partner.
const MesesDeComision = 12

// zona de las fechas de negocio.
var zona = func() *time.Location {
	l, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	if err != nil {
		return time.FixedZone("ART", -3*3600)
	}
	return l
}()

// Porcentaje de comisión de un nivel (0 si no existe).
func Porcentaje(nivel string) int64 { return porcentaje[nivel] }

// NivelEfectivo: el mayor entre el que dio el equipo y el que alcanzan las tiendas activas.
func NivelEfectivo(asignado string, activas int) string {
	mejor := 0
	for i, n := range Niveles {
		if n == asignado || activas >= TiendasParaNivel[n] {
			mejor = max(mejor, i)
		}
	}
	return Niveles[mejor]
}

// MesIndice: cuántos meses calendario (en hora de Argentina) pasaron entre el alta y el mes.
// 0 = el mes del alta.
func MesIndice(alta, mes time.Time) int {
	a, m := alta.In(zona), mes.In(zona)
	return (m.Year()*12 + int(m.Month())) - (a.Year()*12 + int(a.Month()))
}

// FinComision: el primer día del mes en que la tienda deja de generar comisión.
func FinComision(alta time.Time) time.Time {
	a := alta.In(zona)
	return time.Date(a.Year(), a.Month()+MesesDeComision, 1, 0, 0, 0, 0, zona)
}

// Estados de una tienda referida.
const (
	EstadoPrueba = "prueba" // dentro de los 14 días de prueba
	EstadoActiva = "activa" // con plan pago
	EstadoGratis = "gratis" // terminó la prueba y quedó sin plan pago
	EstadoBaja   = "baja"   // canceló el plan o cerró la cuenta
)

// Estado de una tienda referida a una hora.
func Estado(plan string, alta time.Time, baja bool, ahora time.Time) string {
	switch {
	case baja:
		return EstadoBaja
	case ahora.Sub(alta) < mercadopago.DiasDePrueba*24*time.Hour:
		return EstadoPrueba
	case plan != "gratis" && plan != "":
		return EstadoActiva
	default:
		return EstadoGratis
	}
}

// PrecioMensual de un plan en un ciclo: lo que la tienda paga por mes (el anual, dividido 12).
func PrecioMensual(plan, ciclo string) int64 {
	if ciclo == "anio" {
		p, ok := mercadopago.PrecioDe(plan, "anio")
		if !ok {
			return 0
		}
		return p / 12
	}
	p, _ := mercadopago.PrecioDe(plan, "mes")
	return p
}

// Comision de una tienda en un mes, en pesos: el porcentaje del nivel sobre lo que paga por mes,
// si el mes está dentro de los 12 desde el alta. Redondeado al peso.
func Comision(nivel, plan, ciclo string, alta, mes time.Time) int64 {
	if i := MesIndice(alta, mes); i < 0 || i >= MesesDeComision {
		return 0
	}
	return (PrecioMensual(plan, ciclo)*Porcentaje(nivel) + 50) / 100
}

// FechaDePago de la liquidación de un mes: el 10 del mes siguiente.
func FechaDePago(mes time.Time) time.Time {
	m := mes.In(zona)
	return time.Date(m.Year(), m.Month()+1, 10, 0, 0, 0, 0, zona)
}
