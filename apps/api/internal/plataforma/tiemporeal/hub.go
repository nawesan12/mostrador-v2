package tiemporeal

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// Códigos de cierre propios (4000-4999 son de la aplicación). El navegador decide con ellos.
const (
	// CierreResincronizar: cambió algo de la sesión (la tienda, el rol): reconectar ya y recargar todo.
	CierreResincronizar websocket.StatusCode = 4000
	// CierreSinSesion: no hay sesión válida (o se cerró, venció o ya no es del equipo): no reintentar.
	CierreSinSesion websocket.StatusCode = 4401
	// CierreSinPermiso: la sesión no puede usar el tiempo real (una IA conectada).
	CierreSinPermiso websocket.StatusCode = 4403
	// CierreLento: el navegador no leía y se le llenó la cola: reconecta y recarga.
	CierreLento websocket.StatusCode = 4408
	// CierreLimite: demasiadas conexiones de la misma sesión o IP: reintentar más tarde.
	CierreLimite websocket.StatusCode = 4429
)

// Corte es por qué se cierra una conexión.
type Corte struct {
	Codigo websocket.StatusCode
	Motivo string
}

// Cliente es una conexión abierta. Para el panel lleva la sesión (tienda, persona y rol de ese
// momento); para la tienda online, sólo la tienda y, en la página de un pedido, su número.
type Cliente struct {
	Tienda  string
	Usuario string
	Sesion  string
	Rol     string
	IP      string
	Publico bool
	// Pedido: el número del pedido que mira quien compró (sólo público).
	Pedido string

	envio chan Mensaje
	corte chan Corte
	una   sync.Once
}

// Envios es la cola de lo que hay que mandarle al navegador.
func (c *Cliente) Envios() <-chan Mensaje { return c.envio }

// Cortes avisa (una sola vez) que hay que cerrar la conexión.
func (c *Cliente) Cortes() <-chan Corte { return c.corte }

// Cortar pide cerrar la conexión. La primera razón gana.
func (c *Cliente) Cortar(codigo websocket.StatusCode, motivo string) {
	c.una.Do(func() { c.corte <- Corte{codigo, motivo} })
}

func (c *Cliente) mandar(m Mensaje) {
	select {
	case c.envio <- m:
	default:
		// Backpressure: un navegador que no lee no frena a los demás ni crece sin tope en memoria.
		c.Cortar(CierreLento, "no daba abasto")
	}
}

// ErrLimite: se pasó el tope de conexiones de la sesión o de la IP.
var ErrLimite = errors.New("demasiadas conexiones")

type clave struct{ tienda, tipo, id, usuario string }

// Hub reparte los avisos. Junta lo que llega en ráfaga durante una Ventana (un mismo aviso repetido
// sale una vez; muchos avisos distintos del mismo tipo en una tienda salen como uno por lote) y
// espacia las familias ruidosas (visitas, carritos, importación, actividad).
type Hub struct {
	// Ventana: cada cuánto se reparte lo juntado.
	Ventana time.Duration
	// Buffer: cuántos mensajes puede tener pendientes un navegador antes de cortarlo.
	Buffer int
	// MaxPorSesion y MaxPorIP: conexiones a la vez (una por pestaña; una oficina comparte IP).
	MaxPorSesion, MaxPorIP int
	// MaxPorTipo: más avisos distintos que esto del mismo tipo y tienda en una ventana → un lote.
	MaxPorTipo int
	// Espaciar: cada cuánto, como mucho, sale un aviso de esa familia por tienda.
	Espaciar map[string]time.Duration

	mu       sync.Mutex
	tiendas  map[string]map[*Cliente]struct{}
	usuarios map[string]map[*Cliente]struct{}
	publicos map[string]map[*Cliente]struct{}
	sesiones map[string]int
	ips      map[string]int
	pend     map[clave]Evento
	ultimo   map[string]time.Time
	cerrado  bool
}

// NuevoHub con los valores de producción.
func NuevoHub() *Hub {
	return &Hub{
		Ventana: 250 * time.Millisecond, Buffer: 256, MaxPorSesion: 12, MaxPorIP: 60, MaxPorTipo: 25,
		Espaciar: map[string]time.Duration{
			"visitas": 10 * time.Second, "carrito": 2 * time.Second, "importacion": time.Second, "actividad": 2 * time.Second,
		},
	}
}

func (h *Hub) iniciar() {
	if h.tiendas == nil {
		h.tiendas, h.usuarios, h.publicos = map[string]map[*Cliente]struct{}{}, map[string]map[*Cliente]struct{}{}, map[string]map[*Cliente]struct{}{}
		h.sesiones, h.ips, h.pend, h.ultimo = map[string]int{}, map[string]int{}, map[clave]Evento{}, map[string]time.Time{}
	}
}

func sumar(m map[string]map[*Cliente]struct{}, k string, c *Cliente) {
	if m[k] == nil {
		m[k] = map[*Cliente]struct{}{}
	}
	m[k][c] = struct{}{}
}

func sacar(m map[string]map[*Cliente]struct{}, k string, c *Cliente) {
	delete(m[k], c)
	if len(m[k]) == 0 {
		delete(m, k)
	}
}

// Registrar suma una conexión, si no pasa los topes.
func (h *Hub) Registrar(c *Cliente) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.iniciar()
	if h.cerrado {
		return ErrLimite
	}
	if c.IP != "" && h.ips[c.IP] >= h.MaxPorIP {
		return ErrLimite
	}
	if c.Sesion != "" && h.sesiones[c.Sesion] >= h.MaxPorSesion {
		return ErrLimite
	}
	c.envio = make(chan Mensaje, max(h.Buffer, 1))
	c.corte = make(chan Corte, 1)
	if c.IP != "" {
		h.ips[c.IP]++
	}
	if c.Publico {
		sumar(h.publicos, c.Tienda, c)
		return nil
	}
	if c.Sesion != "" {
		h.sesiones[c.Sesion]++
	}
	sumar(h.tiendas, c.Tienda, c)
	sumar(h.usuarios, c.Usuario, c)
	return nil
}

// Quitar saca una conexión (al cerrarse por cualquier motivo).
func (h *Hub) Quitar(c *Cliente) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.iniciar()
	if c.IP != "" {
		if h.ips[c.IP]--; h.ips[c.IP] <= 0 {
			delete(h.ips, c.IP)
		}
	}
	if c.Publico {
		sacar(h.publicos, c.Tienda, c)
		return
	}
	if c.Sesion != "" {
		if h.sesiones[c.Sesion]--; h.sesiones[c.Sesion] <= 0 {
			delete(h.sesiones, c.Sesion)
		}
	}
	sacar(h.tiendas, c.Tienda, c)
	sacar(h.usuarios, c.Usuario, c)
}

// Conexiones abiertas (panel y tienda online).
func (h *Hub) Conexiones() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	n := 0
	for _, m := range h.usuarios {
		n += len(m)
	}
	for _, m := range h.publicos {
		n += len(m)
	}
	return n
}

// Publicar recibe un aviso de Postgres. Lo que corta conexiones (sesión cerrada, cambio de tienda,
// cambio de rol) se aplica en el acto; el resto se junta hasta la próxima ventana.
func (h *Hub) Publicar(e Evento) {
	if !Conocido(e.Tipo) {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.iniciar()
	switch {
	case e.Tipo == "sesion.cerrada" && e.Sesion != "":
		for c := range h.usuarios[e.Usuario] {
			if c.Sesion == e.Sesion {
				c.Cortar(CierreSinSesion, "sesión cerrada")
			}
		}
	case e.Tipo == "sesion.tienda" && e.Sesion != "":
		for c := range h.usuarios[e.Usuario] {
			if c.Sesion == e.Sesion {
				c.Cortar(CierreResincronizar, "cambió la tienda")
			}
		}
	case e.Tipo == "equipo.cambio" && e.Usuario != "":
		// A la persona afectada (le cambiaron o quitaron el rol en esta tienda): que vuelva a entrar
		// con lo que tenga ahora. Si ya no es del equipo, al reconectar no tiene tienda.
		for c := range h.usuarios[e.Usuario] {
			if c.Tienda == e.Tienda {
				c.Cortar(CierreResincronizar, "cambió tu rol")
			}
		}
	}
	k := clave{e.Tienda, e.Tipo, e.ID, e.Usuario}
	if e.Lote {
		k.id = "*"
	}
	h.pend[k] = e
}

// Correr reparte lo juntado cada Ventana hasta que se cancele ctx.
func (h *Hub) Correr(ctx context.Context) {
	t := time.NewTicker(h.Ventana)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case ahora := <-t.C:
			h.repartir(ahora)
		}
	}
}

func (h *Hub) repartir(ahora time.Time) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.iniciar()
	if len(h.pend) == 0 {
		return
	}
	// Lo que sale ahora, por tienda y tipo; lo espaciado que todavía no toca queda pendiente.
	salen := map[[2]string][]Evento{}
	for k, e := range h.pend {
		f := familia(e.Tipo)
		if d := h.Espaciar[f]; d > 0 {
			if u, ok := h.ultimo[e.Tienda+"|"+f]; ok && ahora.Sub(u) < d {
				continue
			}
		}
		salen[[2]string{e.Tienda, e.Tipo}] = append(salen[[2]string{e.Tienda, e.Tipo}], e)
		delete(h.pend, k)
	}
	for tt, evs := range salen {
		if f := familia(tt[1]); h.Espaciar[f] > 0 {
			h.ultimo[tt[0]+"|"+f] = ahora
		}
		if len(evs) > h.MaxPorTipo {
			evs = []Evento{{Tienda: tt[0], Tipo: tt[1], Lote: true}}
		}
		for _, e := range evs {
			h.entregar(e)
		}
	}
	// Los «último envío» viejos no crecen sin tope.
	if len(h.ultimo) > 10000 {
		for k, u := range h.ultimo {
			if ahora.Sub(u) > time.Minute {
				delete(h.ultimo, k)
			}
		}
	}
}

// entregar manda un aviso a quien corresponde. Se llama con el candado tomado.
func (h *Hub) entregar(e Evento) {
	if familia(e.Tipo) == "sesion" {
		// Las sesiones son de la persona: le llega a cada pestaña suya (Configuración › Seguridad),
		// en cualquier tienda, y a nadie más.
		for c := range h.usuarios[e.Usuario] {
			c.mandar(Mensaje{Tipo: "sesion.cambio"})
		}
		return
	}
	if e.Tienda == "" {
		return
	}
	if familia(e.Tipo) == "notificacion" {
		// La campanita es de cada persona (§45): sólo a sus pestañas paradas en esa tienda.
		for c := range h.usuarios[e.Usuario] {
			if c.Tienda == e.Tienda && !c.Publico {
				c.mandar(Mensaje{Tipo: e.Tipo})
			}
		}
		return
	}
	m := Mensaje{Tipo: e.Tipo, ID: e.ID, Lote: e.Lote}
	for c := range h.tiendas[e.Tienda] {
		if Permitido(c.Rol, e.Tipo) {
			c.mandar(m)
		}
	}
	for c := range h.publicos[e.Tienda] {
		switch {
		case publico(e.Tipo):
			c.mandar(m)
		case c.Pedido != "" && (e.Tipo == "pedido.cambio" || e.Tipo == "pago.acreditado") && (e.ID == c.Pedido || e.Lote):
			// Quien compró sólo se entera de que su pedido cambió (ni el número viaja).
			c.mandar(Mensaje{Tipo: e.Tipo})
		}
	}
}

// Resincronizar avisa a todos que se pudieron perder avisos (se cayó y volvió la escucha).
func (h *Hub) Resincronizar() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.iniciar()
	for _, m := range []map[string]map[*Cliente]struct{}{h.usuarios, h.publicos} {
		for _, cs := range m {
			for c := range cs {
				c.mandar(Mensaje{Tipo: TipoResincronizar})
			}
		}
	}
}

// Cerrar corta todas las conexiones (al apagar la API): los navegadores reconectan solos.
func (h *Hub) Cerrar() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.iniciar()
	h.cerrado = true
	for _, m := range []map[string]map[*Cliente]struct{}{h.usuarios, h.publicos} {
		for _, cs := range m {
			for c := range cs {
				c.Cortar(websocket.StatusServiceRestart, "la API se reinicia")
			}
		}
	}
}
