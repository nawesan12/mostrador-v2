package mercadolibre

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// Aviso es el cuerpo de una notificación de Mercado Libre.
type Aviso struct {
	Resource      string      `json:"resource"`
	UserID        json.Number `json:"user_id"`
	Topic         string      `json:"topic"`
	ApplicationID json.Number `json:"application_id"`
}

// Sólo estos recursos: el worker los trae con el token de la tienda.
var recursos = map[string]*regexp.Regexp{
	"orders_v2": regexp.MustCompile(`^/orders/\d{1,20}$`),
	"questions": regexp.MustCompile(`^/questions/\d{1,20}$`),
	"items":     regexp.MustCompile(`^/items/MLA\d{1,15}$`),
}

// Recibir encola el aviso para su tienda. Mercado Libre no firma: se exige que sea de nuestra
// aplicación y de una cuenta conectada, y el recurso se trae después con el token de esa cuenta.
func (s *Servicio) Recibir(ctx context.Context, a Aviso) error {
	if a.ApplicationID.String() != s.Config.AppID {
		return nil
	}
	re, ok := recursos[a.Topic]
	if !ok || !re.MatchString(a.Resource) {
		return nil
	}
	usuario, err := a.UserID.Int64()
	if err != nil {
		return nil
	}
	var tienda *string
	if err := s.Pool.EnTx(ctx, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT app.tienda_de_ml($1)::text`, usuario).Scan(&tienda)
	}); err != nil {
		return err
	}
	if tienda == nil {
		return nil
	}
	return s.Pool.EnTienda(ctx, *tienda, func(t pgx.Tx) error {
		return outbox.Publicar(ctx, t, *tienda, "ml.aviso", map[string]string{"topic": a.Topic, "recurso": a.Resource})
	})
}

// Manejadores del outbox para el worker.
func (s *Servicio) Manejadores() map[string]outbox.Manejador {
	return map[string]outbox.Manejador{
		"ml.sincronizar": func(ctx context.Context, ev outbox.Evento) error {
			var d struct{ Producto string }
			if err := json.Unmarshal(ev.Datos, &d); err != nil || d.Producto == "" {
				return nil
			}
			return s.Sincronizar(ctx, ev.TiendaID, d.Producto)
		},
		"ml.aviso": func(ctx context.Context, ev outbox.Evento) error {
			var d struct{ Topic, Recurso string }
			if err := json.Unmarshal(ev.Datos, &d); err != nil {
				return nil
			}
			return s.procesar(ctx, ev.TiendaID, d.Topic, d.Recurso)
		},
	}
}

func (s *Servicio) procesar(ctx context.Context, tiendaID, topic, recurso string) error {
	re, ok := recursos[topic]
	if !ok || !re.MatchString(recurso) {
		return nil
	}
	token, _, err := s.tokenDe(ctx, tiendaID)
	if errors.Is(err, ErrSinConexion) || errors.Is(err, ErrReautorizar) {
		return nil
	}
	if err != nil {
		return err
	}
	switch topic {
	case "orders_v2":
		return s.traerPedido(ctx, tiendaID, token, recurso)
	case "questions":
		return s.traerPregunta(ctx, tiendaID, token, recurso)
	case "items":
		return s.traerPublicacion(ctx, tiendaID, token, recurso)
	}
	return nil
}

type orden struct {
	ID         int64  `json:"id"`
	Status     string `json:"status"`
	OrderItems []struct {
		Item struct {
			ID          string `json:"id"`
			VariationID int64  `json:"variation_id"`
		} `json:"item"`
		Quantity  int     `json:"quantity"`
		UnitPrice float64 `json:"unit_price"`
	} `json:"order_items"`
	Buyer struct {
		Nickname  string `json:"nickname"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	} `json:"buyer"`
}

// traerPedido: una orden pagada entra una sola vez como pedido del canal ml (descuenta stock y eso
// se sincroniza solo con las demás publicaciones).
func (s *Servicio) traerPedido(ctx context.Context, tiendaID, token, recurso string) error {
	var o orden
	if err := s.pedir(ctx, http.MethodGet, recurso, token, nil, &o); err != nil {
		return err
	}
	if o.Status != "paid" || o.ID == 0 {
		return nil
	}
	// Se reserva la orden primero: dos avisos de la misma venta no crean dos pedidos.
	var reservada bool
	if err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `INSERT INTO ml_pedidos (orden_ml, tienda_id, numero) VALUES ($1, $2, 0) ON CONFLICT DO NOTHING`, o.ID, tiendaID)
		reservada = err == nil && tag.RowsAffected() == 1
		return err
	}); err != nil || !reservada {
		return err
	}
	var n pedidos.PedidoNuevo
	n.Pago, n.Canal, n.Envio = "mp", "ml", "mercadoenvios"
	n.Precios = map[string]int64{}
	nombre := strings.TrimSpace(o.Buyer.FirstName + " " + o.Buyer.LastName)
	if nombre == "" {
		nombre = o.Buyer.Nickname
	}
	n.Cliente = &nombre
	nota := fmt.Sprintf("Venta de Mercado Libre #%d", o.ID)
	n.Nota = &nota
	for _, it := range o.OrderItems {
		var producto, variante string
		err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
			// La variación vendida es una variante: se busca en el mapa que se guardó al publicar.
			return t.QueryRow(ctx, `SELECT producto_id::text, coalesce((SELECT key FROM jsonb_each_text(variaciones)
					WHERE value = $3 LIMIT 1), '')
				FROM ml_publicaciones WHERE tienda_id = $1 AND item_id = $2`,
				tiendaID, it.Item.ID, strconv.FormatInt(it.Item.VariationID, 10)).Scan(&producto, &variante)
		})
		if errors.Is(err, pgx.ErrNoRows) {
			continue // una publicación que no salió de Mostrador
		}
		if err != nil {
			return s.liberar(ctx, tiendaID, o.ID, err)
		}
		n.Items = append(n.Items, pedidos.Linea{ProductoID: producto, VarianteID: variante, Cantidad: it.Quantity})
		if variante != "" {
			n.Precios[producto+":"+variante] = int64(it.UnitPrice + 0.5)
		} else {
			n.Precios[producto] = int64(it.UnitPrice + 0.5)
		}
	}
	if len(n.Items) == 0 {
		return nil
	}
	numero, _, err := pedidos.CrearDesdeProductos(ctx, s.Pool, tiendaID, n)
	if errors.Is(err, pedidos.ErrSinStock) {
		// Se vendió en Mercado Libre algo que acá ya no estaba: queda anotado para que la tienda lo vea.
		s.log().Warn("mercadolibre: venta sin stock en Mostrador", "tienda", tiendaID, "orden", o.ID)
		return nil
	}
	if errors.Is(err, pedidos.ErrSinVariante) || errors.Is(err, pedidos.ErrVariante) {
		// Una variación que no corresponde a ninguna variante a la venta: reintentar no lo arregla.
		s.log().Warn("mercadolibre: venta de una variación sin variante en Mostrador", "tienda", tiendaID, "orden", o.ID)
		return nil
	}
	if err != nil {
		return s.liberar(ctx, tiendaID, o.ID, err)
	}
	return s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE ml_pedidos SET numero = $2 WHERE orden_ml = $1`, o.ID, numero)
		return err
	})
}

// liberar la reserva para que el reintento vuelva a probar.
func (s *Servicio) liberar(ctx context.Context, tiendaID string, orden int64, causa error) error {
	_ = s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `DELETE FROM ml_pedidos WHERE orden_ml = $1 AND numero = 0`, orden)
		return err
	})
	return causa
}

func (s *Servicio) traerPregunta(ctx context.Context, tiendaID, token, recurso string) error {
	var q struct {
		ID          int64  `json:"id"`
		Text        string `json:"text"`
		Status      string `json:"status"`
		ItemID      string `json:"item_id"`
		DateCreated string `json:"date_created"`
		Answer      *struct {
			Text string `json:"text"`
		} `json:"answer"`
	}
	if err := s.pedir(ctx, http.MethodGet, recurso, token, nil, &q); err != nil {
		var em *ErrorML
		if errors.As(err, &em) && em.Estado == http.StatusNotFound {
			return nil // borrada
		}
		return err
	}
	creada, err := time.Parse(time.RFC3339, q.DateCreated)
	if err != nil {
		creada = s.ahora()
	}
	var respuesta *string
	if q.Answer != nil {
		respuesta = &q.Answer.Text
	}
	return s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `INSERT INTO ml_preguntas (pregunta_id, tienda_id, item_id, texto, respuesta, estado, creada_en)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (pregunta_id) DO UPDATE SET respuesta = coalesce($5, ml_preguntas.respuesta), estado = $6`,
			q.ID, tiendaID, q.ItemID, recortar(q.Text, 2000), respuesta, q.Status, creada)
		return err
	})
}

// traerPublicacion: si se pausó o cerró en Mercado Libre, se refleja acá.
func (s *Servicio) traerPublicacion(ctx context.Context, tiendaID, token, recurso string) error {
	var it struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := s.pedir(ctx, http.MethodGet, recurso, token, nil, &it); err != nil {
		return err
	}
	return s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE ml_publicaciones SET estado = $3 WHERE tienda_id = $1 AND item_id = $2`, tiendaID, it.ID, it.Status)
		return err
	})
}

// Pregunta es components/schemas/PreguntaML.
type Pregunta struct {
	ID        int64     `json:"id"`
	ItemID    string    `json:"itemId"`
	Producto  string    `json:"producto"`
	Texto     string    `json:"texto"`
	Respuesta *string   `json:"respuesta"`
	Estado    string    `json:"estado"`
	Creada    time.Time `json:"creada"`
}

// Preguntas de la tienda, las sin responder primero.
func (s *Servicio) Preguntas(ctx context.Context, tiendaID string) ([]Pregunta, error) {
	out := []Pregunta{}
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT q.pregunta_id, q.item_id, coalesce(p.nombre, ''), q.texto, q.respuesta, q.estado, q.creada_en
			FROM ml_preguntas q
			LEFT JOIN ml_publicaciones m ON m.tienda_id = q.tienda_id AND m.item_id = q.item_id
			LEFT JOIN productos p ON p.tienda_id = m.tienda_id AND p.id = m.producto_id
			WHERE q.tienda_id = $1
			ORDER BY q.estado <> 'UNANSWERED', q.creada_en DESC LIMIT 100`, tiendaID)
		if err != nil {
			return err
		}
		for filas.Next() {
			var q Pregunta
			if err := filas.Scan(&q.ID, &q.ItemID, &q.Producto, &q.Texto, &q.Respuesta, &q.Estado, &q.Creada); err != nil {
				return err
			}
			out = append(out, q)
		}
		return filas.Err()
	})
	return out, err
}

// Responder una pregunta (la tienda o la sugerencia de Hornero que la tienda aprobó).
func (s *Servicio) Responder(ctx context.Context, tiendaID string, preguntaID int64, texto string) error {
	texto = strings.TrimSpace(texto)
	if len(texto) < 2 || len(texto) > 2000 {
		return ErrInvalido
	}
	var estado string
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT estado FROM ml_preguntas WHERE pregunta_id = $1`, preguntaID).Scan(&estado)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNoExiste
	}
	if err != nil {
		return err
	}
	token, _, err := s.tokenDe(ctx, tiendaID)
	if err != nil {
		return err
	}
	if err := s.pedir(ctx, http.MethodPost, "/answers", token, map[string]any{"question_id": preguntaID, "text": texto}, nil); err != nil {
		return err
	}
	return s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE ml_preguntas SET respuesta = $2, estado = 'ANSWERED', respondida_en = now() WHERE pregunta_id = $1`,
			preguntaID, texto)
		return err
	})
}
