// Package roles: qué puede hacer cada persona del equipo de una tienda (docs/decisiones.md §36).
//
// Cuatro roles fijos, sin matriz editable: dueña (todo), administración (todo menos el plan, la
// facturación de Tu Local Web y el equipo), ventas (pedidos, clientes y productos, sin precios,
// costos ni finanzas) y local (punto de venta y pedidos). Guardia lo aplica a cada ruta del panel
// antes de llegar al handler; la web sólo oculta lo que el rol no puede.
package roles

import "slices"

// Roles de auth.membresias.
const (
	Duenia = "duenio"
	Admin  = "admin"
	Ventas = "ventas"
	Local  = "local"
)

// Todos los roles, de más a menos.
var Todos = []string{Duenia, Admin, Ventas, Local}

// Invitables: a nadie se lo invita (ni se lo pasa) a dueña.
var Invitables = []string{Admin, Ventas, Local}

// Nombre de cada rol, como lo lee una persona (mails).
var Nombre = map[string]string{Duenia: "Dueña", Admin: "Administración", Ventas: "Ventas", Local: "Local"}

// Permiso es un área del panel.
type Permiso string

const (
	// Panel: lo de cualquiera del equipo (armazón, resumen, avisos, sus propias sesiones).
	Panel Permiso = "panel"
	// Pedidos: pedidos, punto de venta, envíos, devoluciones, turnos y ver los productos (sin costos).
	Pedidos Permiso = "pedidos"
	// Clientes: clientes, WhatsApp, carritos abandonados, reseñas y preguntas.
	Clientes Permiso = "clientes"
	// Productos: editar productos y stock, sin tocar precios ni costos.
	Productos Permiso = "productos"
	// Catalogo: altas y bajas de productos, precios, costos, listas de precio, cupones y promociones.
	Catalogo Permiso = "catalogo"
	// Finanzas: finanzas, informes, facturación con ARCA, Mercado Pago, reembolsos y señas.
	Finanzas Permiso = "finanzas"
	// Tienda: configuración (negocio y CBU, entregas, diseño, dominio, integraciones, marketing,
	// datos y exportaciones).
	Tienda Permiso = "tienda"
	// Cuenta: plan y facturación de Tu Local Web, equipo y cerrar la tienda.
	Cuenta Permiso = "cuenta"
)

// porRol: qué puede cada rol. Es la misma tabla que apps/web/src/lib/panel/permisos.ts (un test
// de la web la compara con testdata/roles.json, que también lee el test de Go).
var porRol = map[string][]Permiso{
	Duenia: {Panel, Pedidos, Clientes, Productos, Catalogo, Finanzas, Tienda, Cuenta},
	Admin:  {Panel, Pedidos, Clientes, Productos, Catalogo, Finanzas, Tienda},
	Ventas: {Panel, Pedidos, Clientes, Productos},
	Local:  {Panel, Pedidos},
}

// Puede dice si el rol tiene el permiso. Un rol desconocido no puede nada.
func Puede(rol string, p Permiso) bool { return slices.Contains(porRol[rol], p) }

// De devuelve los permisos del rol (vacío si no existe).
func De(rol string) []Permiso { return slices.Clone(porRol[rol]) }

// Valido dice si es uno de los cuatro roles.
func Valido(rol string) bool { return slices.Contains(Todos, rol) }

// Invitable dice si se puede invitar (o pasar) a alguien a ese rol.
func Invitable(rol string) bool { return slices.Contains(Invitables, rol) }

// mensajes del 403, por permiso.
var mensajes = map[Permiso]string{
	Clientes:  "Tu rol no ve clientes ni mensajes. Pedíselo a la dueña de la tienda.",
	Productos: "Tu rol no edita productos. Pedíselo a la dueña de la tienda.",
	Catalogo:  "Tu rol no puede tocar precios, costos ni el catálogo. Pedíselo a la dueña de la tienda.",
	Finanzas:  "Tu rol no ve finanzas, cobros ni facturación. Pedíselo a la dueña de la tienda.",
	Tienda:    "Tu rol no cambia la configuración de la tienda. Pedíselo a la dueña de la tienda.",
	Cuenta:    "Sólo la dueña de la tienda maneja el plan y el equipo.",
}

// Mensaje es el detalle del 403 para ese permiso.
func Mensaje(p Permiso) string {
	if m, ok := mensajes[p]; ok {
		return m
	}
	return "Tu rol no puede hacer esto. Pedíselo a la dueña de la tienda."
}
