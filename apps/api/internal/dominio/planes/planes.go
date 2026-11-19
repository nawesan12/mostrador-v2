// Package planes: lo que incluye cada plan de Tu Local Web y los topes que se aplican en el servidor
// (acordado el 19/11/2026). Los precios y la cotización de un cambio viven en dominio/abono; acá está
// qué se puede hacer con cada plan. Su espejo en TypeScript es apps/web/src/lib/panel/planes.ts (un
// test compara las dos tablas con testdata/planes.json).
//
// Reglas:
//   - Durante la prueba gratis (14 días desde el alta) una tienda en Gratis usa Crece (el del medio).
//   - Bajar de plan nunca borra nada: lo que supera el tope queda oculto en la tienda y en sólo
//     lectura en el panel hasta subir de plan o liberar lugar (Productos: los primeros N por
//     antigüedad siguen visibles; Personas: las últimas en entrar quedan en sólo lectura; Fotos: la
//     tienda muestra las primeras N de cada producto).
//   - Un tope se informa con 402 y el código tope_plan, con el plan que lo habilita.
package planes

import (
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/abono"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Funcion es algo que un plan incluye o no.
type Funcion string

const (
	// Dominio propio (tutienda.com.ar). Sin él, la tienda vive en la dirección de Tu Local Web.
	Dominio Funcion = "dominio"
	// ARCA: factura electrónica.
	ARCA Funcion = "arca"
	// Correo Argentino: cotizar en el checkout y generar envíos.
	Correo Funcion = "correo"
	// Cupones, combos y recordatorios de carritos abandonados.
	Cupones Funcion = "cupones"
	// Roles y permisos: invitar con un rol distinto de Administración.
	Roles Funcion = "roles"
	// Venta en el local (punto de venta) y cobro con QR.
	Local Funcion = "local"
	// Mercado Libre como canal.
	MercadoLibre Funcion = "mercadolibre"
	// Newsletter y estadísticas (informes).
	Newsletter Funcion = "newsletter"
	// Mayorista y listas de precio.
	Mayorista Funcion = "mayorista"
)

// Nombre de cada función como la lee una persona («X está en el plan Y»).
var nombreFuncion = map[Funcion]string{
	Dominio:      "El dominio propio",
	ARCA:         "La factura electrónica con ARCA",
	Correo:       "Correo Argentino",
	Cupones:      "Los cupones, combos y carritos abandonados",
	Roles:        "Los roles y permisos del equipo",
	Local:        "La venta en el local y el cobro con QR",
	MercadoLibre: "Mercado Libre",
	Newsletter:   "El newsletter y las estadísticas",
	Mayorista:    "La venta mayorista y las listas de precio",
}

// SinTope en un tope numérico.
const SinTope = 0

// Plan con sus topes (0: sin tope, con uso razonable).
type Plan struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
	// Mensual: el precio por mes en pesos (precio final, IVA incluido). El anual cobra 10.
	Mensual int64 `json:"mensual"`
	// Productos cargados (publicados o pausados).
	Productos int `json:"productos"`
	// Personas del equipo, la dueña incluida (y las invitaciones pendientes).
	Personas int `json:"personas"`
	// Fotos por producto.
	Fotos int `json:"fotos"`
	// FotosMB: el espacio para fotos de la tienda, en MB (lo aplica la web, que guarda las fotos).
	FotosMB int `json:"fotosMB"`
	// Tincho: respuestas con IA por mes (más los packs).
	Tincho    int       `json:"tincho"`
	Funciones []Funcion `json:"funciones"`
}

// Todos, de menor a mayor. El id "pro" es Crece (quedó de antes y está en la base y en los débitos).
var Todos = []Plan{
	{ID: "gratis", Nombre: "Gratis", Productos: 20, Personas: 1, Fotos: 3, FotosMB: 15, Tincho: 50, Funciones: []Funcion{}},
	{ID: "emprende", Nombre: "Emprende", Mensual: 14_900, Productos: 150, Personas: 2, Fotos: 5, FotosMB: 150, Tincho: 500,
		Funciones: []Funcion{Dominio, ARCA, Correo, Cupones}},
	{ID: "pro", Nombre: "Crece", Mensual: 34_900, Productos: 500, Personas: 5, Fotos: 8, FotosMB: 600, Tincho: 2_000,
		Funciones: []Funcion{Dominio, ARCA, Correo, Cupones, Roles, Local, MercadoLibre, Newsletter}},
	{ID: "negocio", Nombre: "Negocio", Mensual: 79_900, Productos: SinTope, Personas: SinTope, Fotos: 10, FotosMB: 2_048, Tincho: 8_000,
		Funciones: []Funcion{Dominio, ARCA, Correo, Cupones, Roles, Local, MercadoLibre, Newsletter, Mayorista}},
}

// DePrueba: el plan que rige mientras dura la prueba gratis (el del medio, destacado).
const DePrueba = "pro"

// PackTincho: un pack extra de respuestas con IA, por mes, y su precio.
const (
	PackTincho       = 2_000
	PrecioPackTincho = 4_900
)

// De devuelve el plan por id (el gratis si no existe).
func De(id string) Plan {
	for _, p := range Todos {
		if p.ID == id {
			return p
		}
	}
	return Todos[0]
}

// Efectivo: el plan que rige. En la prueba gratis, una tienda en Gratis usa Crece.
func Efectivo(plan string, finPrueba, ahora time.Time) Plan {
	if plan == "gratis" && ahora.Before(finPrueba) {
		return De(DePrueba)
	}
	return De(plan)
}

// DeEstado: el plan que rige con el estado del abono.
func DeEstado(e abono.Estado, ahora time.Time) Plan { return Efectivo(e.Plan, e.FinPrueba, ahora) }

// Tiene dice si el plan incluye la función.
func (p Plan) Tiene(f Funcion) bool { return slices.Contains(p.Funciones, f) }

// Minimo: el plan más barato que incluye la función.
func Minimo(f Funcion) Plan {
	for _, p := range Todos {
		if p.Tiene(f) {
			return p
		}
	}
	return Todos[len(Todos)-1]
}

// cabe: n entra en el tope (0 es sin tope).
func cabe(tope, n int) bool { return tope == SinTope || n <= tope }

// minimoQue: el plan más barato en el que n entra según el tope que devuelve campo.
func minimoQue(campo func(Plan) int, n int) Plan {
	for _, p := range Todos {
		if cabe(campo(p), n) {
			return p
		}
	}
	return Todos[len(Todos)-1]
}

// Tope: lo que se quiso hacer no entra en el plan.
type Tope struct {
	// Que: productos, personas, fotos o la función.
	Que string
	// Plan: el que rige; Sube: el más barato que lo permite.
	Plan, Sube Plan
	Detalle    string
	// SoloLectura: no es algo nuevo sino algo que ya quedó fuera del plan (después de bajar).
	SoloLectura bool
}

func (t *Tope) Error() string { return "planes: tope de " + t.Que + " en " + t.Plan.ID }

// Problema: 402 con el código tope_plan, el plan que lo habilita y qué se topó.
func (t *Tope) Problema() *problema.Problema {
	titulo := "Llegaste al tope de tu plan"
	if t.SoloLectura {
		titulo = "Quedó fuera de tu plan"
	}
	return &problema.Problema{Estado: http.StatusPaymentRequired, Codigo: "tope_plan", Titulo: titulo,
		Detalle: t.Detalle, Plan: t.Sube.ID, Funcion: t.Que}
}

// miles: 2000 → «2.000».
func miles(n int) string {
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return s
}

func hasta(n int) string {
	if n == SinTope {
		return "sin tope"
	}
	return "hasta " + miles(n)
}

// ExigirFuncion: nil si el plan la incluye; si no, el tope con el plan que la habilita.
func ExigirFuncion(p Plan, f Funcion) error {
	if p.Tiene(f) {
		return nil
	}
	m := Minimo(f)
	return &Tope{Que: string(f), Plan: p, Sube: m,
		Detalle: fmt.Sprintf("%s está desde el plan %s. Tu plan es %s.", nombreFuncion[f], m.Nombre, p.Nombre)}
}

// ExigirProductos: nil si n productos entran en el plan.
func ExigirProductos(p Plan, n int) error {
	if cabe(p.Productos, n) {
		return nil
	}
	m := minimoQue(func(x Plan) int { return x.Productos }, n)
	return &Tope{Que: "productos", Plan: p, Sube: m,
		Detalle: fmt.Sprintf("Tu plan %s llega hasta %s productos. Con %s cargás %s.", p.Nombre, miles(p.Productos), m.Nombre,
			hasta(m.Productos))}
}

// ExigirPersonas: nil si n personas (la dueña incluida) entran en el plan.
func ExigirPersonas(p Plan, n int) error {
	if cabe(p.Personas, n) {
		return nil
	}
	m := minimoQue(func(x Plan) int { return x.Personas }, n)
	quien := "personas"
	if p.Personas == 1 {
		quien = "persona: vos"
	}
	return &Tope{Que: "personas", Plan: p, Sube: m,
		Detalle: fmt.Sprintf("Tu plan %s es para %d %s. Con %s suman %s.", p.Nombre, p.Personas, quien, m.Nombre,
			strings.Replace(hasta(m.Personas), "sin tope", "las que necesites", 1))}
}

// ExigirFotos: nil si n fotos por producto entran en el plan.
func ExigirFotos(p Plan, n int) error {
	if cabe(p.Fotos, n) {
		return nil
	}
	m := minimoQue(func(x Plan) int { return x.Fotos }, n)
	return &Tope{Que: "fotos", Plan: p, Sube: m,
		Detalle: fmt.Sprintf("Tu plan %s permite %d fotos por producto. Con %s, hasta %d.", p.Nombre, p.Fotos, m.Nombre, m.Fotos)}
}

// productoFuera: el tope de un producto que quedó fuera del plan (se puede borrar, no editar).
func productoFuera(p Plan) error {
	m := minimoQue(func(x Plan) int { return x.Productos }, p.Productos+1)
	return &Tope{Que: "productos", Plan: p, Sube: m, SoloLectura: true,
		Detalle: fmt.Sprintf("Este producto quedó fuera de tu plan %s (hasta %s productos): no se ve en la tienda y no se puede editar. Pasá a %s o borrá otro para liberar lugar.",
			p.Nombre, miles(p.Productos), m.Nombre)}
}

// personaFuera: el tope de una persona del equipo que quedó fuera del plan.
func personaFuera(p Plan) error {
	m := minimoQue(func(x Plan) int { return x.Personas }, p.Personas+1)
	return &Tope{Que: "personas", Plan: p, Sube: m, SoloLectura: true,
		Detalle: fmt.Sprintf("Tu lugar en el equipo quedó en sólo lectura: el plan %s es para %d. La dueña puede pasar a %s o liberar un lugar.",
			p.Nombre, p.Personas, m.Nombre)}
}
