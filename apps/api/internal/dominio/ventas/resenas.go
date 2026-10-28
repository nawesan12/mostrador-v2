package ventas

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

var (
	ErrResenaInvalida = errors.New("ventas: reseña inválida")
	ErrYaReseno       = errors.New("ventas: ese pedido ya tiene reseña")
)

// ResenaPanel es una reseña como la ve la tienda.
type ResenaPanel struct {
	ID        string    `json:"id"`
	Autor     string    `json:"autor"`
	Estrellas int       `json:"estrellas"`
	Texto     string    `json:"texto"`
	Origen    string    `json:"origen"`
	Creada    time.Time `json:"creada"`
	Producto  *string   `json:"producto"`
	ConCompra bool      `json:"conCompra"`
	Oculta    bool      `json:"oculta"`
	Respuesta *string   `json:"respuesta"`
}

// NuevaResena la deja quien compró (o quien sea: sin código de pedido no queda "con compra").
type NuevaResena struct {
	Codigo     string `json:"codigo"`
	ProductoID string `json:"productoId"`
	Autor      string `json:"autor"`
	Estrellas  int    `json:"estrellas"`
	Texto      string `json:"texto"`
}

// DejarResena guarda la reseña de la tienda online. Con el código público de un pedido cobrado
// queda con compra (una por pedido y producto).
func DejarResena(ctx context.Context, pool tx, tienda string, n NuevaResena) error {
	n.Autor, n.Texto = strings.TrimSpace(n.Autor), strings.TrimSpace(n.Texto)
	if len([]rune(n.Autor)) < 2 || len([]rune(n.Autor)) > 60 || n.Estrellas < 1 || n.Estrellas > 5 || len([]rune(n.Texto)) > 1000 {
		return ErrResenaInvalida
	}
	return pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var pedido *int
		if n.Codigo != "" {
			var num int
			err := t.QueryRow(ctx, `SELECT numero FROM pedidos WHERE codigo_publico = $1
				AND estado_pago IN ('Aprobado', 'Acreditada', 'Cobrado')`, n.Codigo).Scan(&num)
			if err == nil {
				pedido = &num
			} else if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
		}
		var prod *string
		if n.ProductoID != "" {
			var id string
			if err := t.QueryRow(ctx, `SELECT id::text FROM productos WHERE id::text = $1`, n.ProductoID).Scan(&id); err == nil {
				prod = &id
			}
		}
		_, err := t.Exec(ctx, `INSERT INTO resenas (tienda_id, pedido_numero, producto_id, autor, estrellas, texto)
			VALUES ($1, $2, $3::uuid, $4, $5, $6)`, tienda, pedido, prod, n.Autor, n.Estrellas, n.Texto)
		if err != nil && strings.Contains(err.Error(), "resenas_por_pedido") {
			return ErrYaReseno
		}
		return err
	})
}

// Preguntar deja una pregunta sobre un producto publicado.
func Preguntar(ctx context.Context, pool tx, tienda, productoID, nombre, email, pregunta string) error {
	nombre, pregunta, email = strings.TrimSpace(nombre), strings.TrimSpace(pregunta), strings.TrimSpace(email)
	if len([]rune(nombre)) < 2 || len([]rune(nombre)) > 60 || len([]rune(pregunta)) < 3 || len([]rune(pregunta)) > 500 || len(email) > 120 {
		return ErrResenaInvalida
	}
	return pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `INSERT INTO preguntas (tienda_id, producto_id, nombre, email, pregunta)
			SELECT $1, id, $3, nullif($4, ''), $5 FROM productos WHERE id::text = $2 AND publicado`, tienda, productoID, nombre, email, pregunta)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrResenaInvalida
		}
		return err
	})
}

// ResenasPanel: las reseñas de la tienda con su resumen, y las preguntas de la tienda online.
func ResenasPanel(ctx context.Context, pool tx, tienda string) (resenas []ResenaPanel, preguntas []PreguntaPanel, err error) {
	resenas, preguntas = []ResenaPanel{}, []PreguntaPanel{}
	err = pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT r.id::text, r.autor, r.estrellas, r.texto, r.origen, r.creada, p.nombre,
				r.pedido_numero IS NOT NULL, r.oculta, r.respuesta
			FROM resenas r LEFT JOIN productos p ON p.tienda_id = r.tienda_id AND p.id = r.producto_id
			WHERE r.tienda_id = $1 ORDER BY r.creada DESC LIMIT 500`, tienda)
		if err != nil {
			return err
		}
		for filas.Next() {
			var r ResenaPanel
			if err := filas.Scan(&r.ID, &r.Autor, &r.Estrellas, &r.Texto, &r.Origen, &r.Creada, &r.Producto, &r.ConCompra, &r.Oculta, &r.Respuesta); err != nil {
				filas.Close()
				return err
			}
			resenas = append(resenas, r)
		}
		filas.Close()
		if err := filas.Err(); err != nil {
			return err
		}
		filas, err = t.Query(ctx, `SELECT 't-' || q.id::text, p.nombre, p.foto, q.pregunta, q.creada, q.respuesta
			FROM preguntas q JOIN productos p ON p.tienda_id = q.tienda_id AND p.id = q.producto_id
			WHERE q.tienda_id = $1 ORDER BY (q.respuesta IS NULL) DESC, q.creada DESC LIMIT 300`, tienda)
		if err != nil {
			return err
		}
		defer filas.Close()
		for filas.Next() {
			var q PreguntaPanel
			if err := filas.Scan(&q.ID, &q.Producto, &q.Foto, &q.Pregunta, &q.Creada, &q.Respuesta); err != nil {
				return err
			}
			preguntas = append(preguntas, q)
		}
		return filas.Err()
	})
	return
}

// ResponderPregunta de la tienda online (id "t-<uuid>").
func ResponderPregunta(ctx context.Context, pool tx, tienda, id, respuesta string) error {
	respuesta = strings.TrimSpace(respuesta)
	if !strings.HasPrefix(id, "t-") || len([]rune(respuesta)) < 1 || len([]rune(respuesta)) > 1000 {
		return ErrResenaInvalida
	}
	return pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		// La primera respuesta le llega por mail a quien preguntó (si dejó su email).
		var email string
		var primera bool
		err := t.QueryRow(ctx, `SELECT coalesce(email, ''), respuesta IS NULL FROM preguntas WHERE id::text = $1 FOR UPDATE`,
			strings.TrimPrefix(id, "t-")).Scan(&email, &primera)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrResenaInvalida
		}
		if err != nil {
			return err
		}
		if _, err := t.Exec(ctx, `UPDATE preguntas SET respuesta = $2, respondida_en = now() WHERE id::text = $1`,
			strings.TrimPrefix(id, "t-"), respuesta); err != nil {
			return err
		}
		if primera && email != "" {
			return outbox.Publicar(ctx, t, tienda, "pregunta.respondida", map[string]string{"id": strings.TrimPrefix(id, "t-")})
		}
		return nil
	})
}

// AccionResena: ocultar/mostrar o responder públicamente.
func AccionResena(ctx context.Context, pool tx, tienda, id, accion, respuesta string) error {
	return pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var tag interface{ RowsAffected() int64 }
		var err error
		switch accion {
		case "ocultar", "mostrar":
			tag, err = t.Exec(ctx, `UPDATE resenas SET oculta = $2 WHERE id::text = $1`, id, accion == "ocultar")
		case "responder":
			r := strings.TrimSpace(respuesta)
			if len([]rune(r)) > 1000 {
				return ErrResenaInvalida
			}
			tag, err = t.Exec(ctx, `UPDATE resenas SET respuesta = nullif($2, '') WHERE id::text = $1`, id, r)
		default:
			return ErrResenaInvalida
		}
		if err == nil && tag.RowsAffected() == 0 {
			return ErrResenaInvalida
		}
		return err
	})
}

// Resumen de las visibles: promedio, cantidad, por estrellas (de 5 a 1) y % que recomendaría.
func Resumen(rs []ResenaPanel) (prom *float64, total int, por []int, recomienda int) {
	por = []int{0, 0, 0, 0, 0}
	suma, buenas := 0, 0
	for _, r := range rs {
		if r.Oculta {
			continue
		}
		total++
		suma += r.Estrellas
		por[5-r.Estrellas]++
		if r.Estrellas >= 4 {
			buenas++
		}
	}
	if total > 0 {
		p := float64(suma) / float64(total)
		prom = &p
		recomienda = buenas * 100 / total
	}
	return
}
