// Package tiemporeal reparte a los navegadores, por WebSocket, los avisos de cambio que manda
// Postgres (docs/decisiones.md §44). El origen es la base: los triggers de la migración 00061 hacen
// NOTIFY al confirmarse cada transacción; una conexión dedicada (Escucha) los recibe y el Hub los
// reparte por tienda, filtrando por rol. Un aviso dice sólo qué cambió ({tipo, id}): el navegador
// vuelve a pedir lo suyo por los caminos de siempre, que aplican RLS y roles como cualquier pedido.
package tiemporeal

import (
	"encoding/json"
	"strings"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/roles"
)

// Canal de NOTIFY que escribe la migración 00061.
const Canal = "tiempo_real"

// Evento es lo que llega de Postgres. Tienda, Usuario y Sesion sirven para rutear y nunca salen
// hacia el navegador.
type Evento struct {
	Tienda  string `json:"t"`
	Tipo    string `json:"k"`
	ID      string `json:"id"`
	Usuario string `json:"u"`
	Sesion  string `json:"s"`
	Lote    bool   `json:"lote"`
}

// Mensaje es lo único que viaja al navegador.
type Mensaje struct {
	Tipo string `json:"tipo"`
	ID   string `json:"id,omitempty"`
	Lote bool   `json:"lote,omitempty"`
}

// Tipos de control que manda el propio servidor.
const (
	// Hola: la conexión quedó lista (el navegador aprovecha para ponerse al día).
	TipoHola = "hola"
	// Latido: cada Ping, para que el navegador sepa que la conexión sigue viva.
	TipoLatido = "latido"
	// Resincronizar: se perdieron avisos (se cayó la escucha de Postgres): hay que recargar todo.
	TipoResincronizar = "resincronizar"
)

// Leer interpreta un aviso de Postgres. Un aviso que no se entiende o sin tipo se descarta.
func Leer(payload string) (Evento, bool) {
	var e Evento
	if len(payload) > 2000 || json.Unmarshal([]byte(payload), &e) != nil || e.Tipo == "" || len(e.Tipo) > 40 {
		return Evento{}, false
	}
	if len(e.ID) > 80 {
		e.ID = e.ID[:80]
	}
	return e, true
}

// permisos: qué permiso hace falta para recibir cada familia de avisos (el prefijo del tipo). Con
// varios, alcanza con uno. Un tipo que no está acá no le llega a nadie (cerrado por defecto) y un
// test lo controla contra los tipos que escriben los triggers.
var permisos = map[string][]roles.Permiso{
	"pedido":      {roles.Pedidos},
	"pago":        {roles.Pedidos}, // el estado del pago, sin montos: lo ve todo el equipo con Pedidos
	"producto":    {roles.Pedidos}, // Pedidos incluye ver los productos (punto de venta, Local)
	"envio":       {roles.Pedidos},
	"turno":       {roles.Pedidos},
	"devolucion":  {roles.Pedidos},
	"cliente":     {roles.Clientes},
	"carrito":     {roles.Clientes},
	"resena":      {roles.Clientes},
	"pregunta":    {roles.Clientes},
	"cupon":       {roles.Catalogo},
	"comprobante": {roles.Finanzas},
	"finanzas":    {roles.Finanzas},
	"integracion": {roles.Tienda, roles.Finanzas},
	"importacion": {roles.Tienda},
	"actividad":   {roles.Tienda},
	"mitienda":    {roles.Tienda},
	"dominio":     {roles.Tienda},
	"plan":        {roles.Cuenta},
	"equipo":      {roles.Cuenta}, // y a la persona afectada (Hub)
	"tienda":      {roles.Panel},
	"visitas":     {roles.Panel},
	"tarea":       {roles.Panel},
	// "sesion": sólo a la persona de la sesión (Hub), nunca por permiso.
}

func familia(tipo string) string {
	f, _, _ := strings.Cut(tipo, ".")
	return f
}

// Conocido dice si el tipo tiene destinatarios definidos.
func Conocido(tipo string) bool {
	f := familia(tipo)
	_, ok := permisos[f]
	return ok || f == "sesion"
}

// Permitido dice si el rol puede recibir avisos de ese tipo.
func Permitido(rol, tipo string) bool {
	for _, p := range permisos[familia(tipo)] {
		if roles.Puede(rol, p) {
			return true
		}
	}
	return false
}

// publico: lo que ve una página de la tienda online sin sesión (sólo de esa tienda): que cambió el
// stock o la ficha de un producto. Del pedido de quien compra, sólo el suyo (Hub).
func publico(tipo string) bool {
	return tipo == "producto.stock" || tipo == "producto.cambio" || tipo == "producto.borrado"
}
