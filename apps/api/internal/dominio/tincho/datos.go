package tincho

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
)

// Sistema es el mismo para todas las tiendas: sin nombres, fechas ni datos (eso va en el primer
// mensaje), así el prefijo con las herramientas se lee del caché entre todas.
const Sistema = `Sos Tincho, el asistente de Tu Local Web, una plataforma argentina para vender online, por WhatsApp, Instagram, Mercado Libre y en el local.

A quién le hablás:
- Sólo a la dueña de la tienda o a alguien de su equipo, dentro del panel, la app o su Telegram. Nunca le hablás a clientes de la tienda ni escribís mensajes como si fueras la tienda hablando con un cliente; si te piden un texto para mandarle a un cliente, lo armás como borrador para que la persona lo copie y lo mande ella.
- El primer mensaje trae los DATOS DEL NEGOCIO de esa tienda (resumidos, a una hora redondeada). Para el detalle o lo más nuevo, usá las herramientas.

Cómo respondés:
- Español rioplatense (vos), cálido y concreto, como alguien del equipo. De 1 a 4 oraciones (máximo 90 palabras), sin títulos ni listas largas.
- Marcá con **negrita** sólo el dato clave. La plata va así: $ 12.345 (pesos, sin centavos).
- Si el mensaje empieza con [telegram], respondé en texto plano: sin negritas, sin tablas, sin markdown.
- Usá sólo los datos y lo que devuelven las herramientas. Si algo no está, decí que todavía no lo tenés: nunca inventes números, nombres, fechas ni funciones del panel.
- Si los datos dicen que quien pregunta no ve montos, nunca des montos de dinero (ni los calcules): ofrecé pedidos, unidades o stock.
- Si una herramienta dice que el rol no tiene permiso, explicalo con amabilidad y decí que lo puede ver la dueña o administración.
- Si te preguntan algo que no tiene que ver con la tienda o con vender, decí con amabilidad que sólo ayudás con la tienda.

Acciones:
- Podés proponer: crear un cupón, cambiar el estado de un pedido, ajustar stock o precio de un producto y crear un link de pago. Las herramientas de acción NO hacen el cambio: arman una propuesta y la persona la confirma con el botón «Confirmar».
- Nunca digas que algo quedó hecho, guardado, creado o enviado. Decí que te quedó lista la propuesta para que la confirme. En Telegram no hay botón: decí que la confirme desde Tincho en la app o el panel.
- Antes de proponer, buscá lo que haga falta (el id del producto con listar_productos, el estado del pedido con ver_pedido). Si falta un dato clave (el código o el descuento de un cupón, el monto de un link), preguntalo.
- No podés mandar mensajes, mails ni campañas, borrar cosas ni tocar la configuración: contá en qué sección del panel se hace.

Herramientas: pedí sólo lo necesario, con pocas filas. Para «cómo vengo» usá resumen_del_dia; para comparar períodos, informe_de_ventas.`

// La foto del negocio se arma una vez por franja de 15 minutos: todas las preguntas de esa franja
// leen el mismo prefijo del caché (y la base no se consulta en cada pregunta).
const franja = 15 * time.Minute

type datosGuardados struct {
	texto string
	vence time.Time
}

var horaArgentina = time.FixedZone("ART", -3*3600)

func (s *Servicio) datosDe(ctx context.Context, tiendaID string, montos bool) (string, error) {
	ahora := s.ahora().In(horaArgentina)
	desde := ahora.Truncate(franja)
	clave := fmt.Sprintf("%s|%t|%d", tiendaID, montos, desde.Unix())
	s.mu.Lock()
	if g, ok := s.datos[clave]; ok && ahora.Before(g.vence) {
		s.mu.Unlock()
		return g.texto, nil
	}
	s.mu.Unlock()
	t, err := s.armarDatos(ctx, tiendaID, montos, desde)
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	if s.datos == nil || len(s.datos) > 2000 {
		s.datos = map[string]datosGuardados{}
	}
	s.datos[clave] = datosGuardados{t, desde.Add(franja)}
	s.mu.Unlock()
	return t, nil
}

// Foto es el resumen del negocio para el modelo: corto, con claves fijas y sin segundos.
type Foto struct {
	Tienda        map[string]string `json:"tienda"`
	VeMontos      bool              `json:"veMontos"`
	Hoy           map[string]any    `json:"hoy"`
	Ultimos7Dias  map[string]any    `json:"ultimos7Dias"`
	ParaHacer     map[string]int    `json:"paraHacer"`
	MasVendidos7d []any             `json:"masVendidos7d"`
	Catalogo      map[string]any    `json:"catalogo"`
	Clientes      any               `json:"clientes"`
	Meta          any               `json:"metaDelMes,omitempty"`
}

func (s *Servicio) armarDatos(ctx context.Context, tiendaID string, montos bool, cuando time.Time) (string, error) {
	tienda, err := tiendas.Obtener(ctx, s.Pool, tiendaID)
	if err != nil {
		return "", err
	}
	// Los tableros con la hora de la franja: el mismo resultado para todas las preguntas del rato.
	hoy, err := tiendas.ObtenerTablero(ctx, s.Pool, tiendaID, "hoy", cuando)
	if err != nil {
		return "", err
	}
	semana, err := tiendas.ObtenerTablero(ctx, s.Pool, tiendaID, "7d", cuando)
	if err != nil {
		return "", err
	}
	c := tiendas.ContextoHornero(ctx, s.Pool, tiendaID)
	if !montos {
		hoy.SinMontos()
		semana.SinMontos()
	}
	f := Foto{
		Tienda:   map[string]string{"nombre": tienda.Nombre, "rubro": tienda.Rubro},
		VeMontos: montos,
		Hoy:      map[string]any{"pedidos": hoy.Kpis.Pedidos.Actual},
		Ultimos7Dias: map[string]any{"pedidos": semana.Kpis.Pedidos.Actual, "pedidosSemanaAnterior": semana.Kpis.Pedidos.Anterior,
			"clientesNuevos": semana.Kpis.ClientesNuevos.Actual},
		ParaHacer: map[string]int{"preparar": semana.ParaHacer.Preparar, "transferencias": semana.ParaHacer.Transferencias,
			"sinStock": semana.ParaHacer.SinStock, "chatsSinResponder": semana.ParaHacer.Chats},
		MasVendidos7d: primeros(semana.MasVendidos, 5),
		Clientes:      c["clientes"],
	}
	if montos {
		if hoy.Ventas != nil {
			f.Hoy["ventas"] = int64(hoy.Ventas.Actual)
		}
		if semana.Ventas != nil {
			f.Ultimos7Dias["ventas"] = int64(semana.Ventas.Actual)
			f.Ultimos7Dias["ventasSemanaAnterior"] = int64(semana.Ventas.Anterior)
		}
		if semana.Kpis.Ticket != nil {
			f.Ultimos7Dias["ticketPromedio"] = int64(semana.Kpis.Ticket.Actual)
		}
		if semana.Meta != nil && semana.Meta.Objetivo > 0 {
			f.Meta = semana.Meta
		}
	}
	// Catálogo corto: cuántos hay, las categorías y los 15 que más salen (el resto, con herramientas).
	prods, _ := c["productos"].([]any)
	cats := []string{}
	top := []map[string]any{}
	for _, x := range prods {
		p, _ := x.(map[string]any)
		if cat, _ := p["categoria"].(*string); cat != nil && *cat != "" && !slices.Contains(cats, *cat) {
			cats = append(cats, *cat)
		}
		if len(top) < 15 {
			top = append(top, map[string]any{"nombre": p["nombre"], "precio": p["precio"], "stock": p["stock"], "vendidos30d": p["vendidos30d"]})
		}
	}
	slices.Sort(cats)
	f.Catalogo = map[string]any{"publicados": len(prods), "categorias": primeros(cats, 20), "masVendidos30d": top}
	b, err := json.Marshal(f)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("DATOS DEL NEGOCIO (al %s, hora de Argentina; lo de después, con herramientas)\n%s",
		cuando.Format("02/01/2006 15:04"), b), nil
}

func primeros[T any](l []T, n int) []T {
	if len(l) > n {
		return l[:n]
	}
	if l == nil {
		return []T{}
	}
	return l
}
