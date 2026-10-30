// Package devoluciones: el botón de arrepentimiento de cada tienda online (Ley 24.240, art. 34, y
// Res. 424/2020) y las devoluciones del panel. Quien compró pide la revocación con el código de su
// pedido (o el número y su email), recibe un código de trámite en pantalla y por mail, y la tienda
// la ve en Ventas › Devoluciones.
package devoluciones

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

var (
	ErrInvalido     = errors.New("devoluciones: datos inválidos")
	ErrNoEncontrado = errors.New("devoluciones: no encontramos ese pedido")
	ErrFueraDePlazo = errors.New("devoluciones: pasaron más de 10 días desde la entrega")
)

// Plazo de la revocación: 10 días corridos desde la entrega (o desde la compra si todavía no llegó).
const plazo = 10 * 24 * time.Hour

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}

type Servicio struct {
	Pool   tx
	Correo correo.Correo
}

// Pedido de arrepentimiento desde la tienda online.
type Pedido struct {
	Codigo string `json:"codigo"`
	Numero int    `json:"numero"`
	Email  string `json:"email"`
	Nombre string `json:"nombre"`
	Motivo string `json:"motivo"`
}

var emailValido = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// Arrepentirse registra la revocación y devuelve el código de trámite. Un segundo pedido sobre la
// misma compra devuelve el mismo código (no duplica el trámite).
func (s *Servicio) Arrepentirse(ctx context.Context, tienda, nombreTienda string, p Pedido, ahora time.Time) (string, error) {
	p.Nombre, p.Email, p.Motivo = strings.TrimSpace(p.Nombre), strings.ToLower(strings.TrimSpace(p.Email)), strings.TrimSpace(p.Motivo)
	if len([]rune(p.Nombre)) < 2 || len([]rune(p.Nombre)) > 80 || !emailValido.MatchString(p.Email) || len([]rune(p.Motivo)) > 500 ||
		(p.Codigo == "" && p.Numero <= 0) {
		return "", ErrInvalido
	}
	var codigo string
	nuevo := true
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var numero int
		var total int64
		var emailPedido *string
		var creado time.Time
		var entregado *time.Time
		q := `SELECT numero, total, lower(comprador->>'email'), creado_en, (estados_en->>'Entregado')::timestamptz FROM pedidos WHERE `
		var err error
		if p.Codigo != "" {
			err = t.QueryRow(ctx, q+`codigo_publico = $1`, p.Codigo).Scan(&numero, &total, &emailPedido, &creado, &entregado)
		} else {
			err = t.QueryRow(ctx, q+`numero = $1`, p.Numero).Scan(&numero, &total, &emailPedido, &creado, &entregado)
		}
		// Con el número hace falta que el email coincida (si no, cualquiera revoca compras ajenas).
		if errors.Is(err, pgx.ErrNoRows) || (p.Codigo == "" && (emailPedido == nil || *emailPedido != p.Email)) {
			return ErrNoEncontrado
		}
		if err != nil {
			return err
		}
		desde := creado
		if entregado != nil {
			desde = *entregado
		}
		if ahora.Sub(desde) > plazo {
			return ErrFueraDePlazo
		}
		if err := t.QueryRow(ctx, `SELECT codigo FROM devoluciones WHERE pedido_numero = $1 AND origen = 'arrepentimiento'`, numero).Scan(&codigo); err == nil {
			nuevo = false
			return nil
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		// Dos pedidos a la vez por la misma compra: el segundo no choca, devuelve el mismo código.
		err = t.QueryRow(ctx, `INSERT INTO devoluciones (tienda_id, codigo, pedido_numero, nombre, email, motivo, monto)
			VALUES ($1, format('ARR-%s-%s', $2::int, nextval('devoluciones_numero')), $3, $4, $5, $6, $7)
			ON CONFLICT (tienda_id, pedido_numero) WHERE origen = 'arrepentimiento' DO NOTHING RETURNING codigo`,
			tienda, ahora.Year(), numero, p.Nombre, p.Email, p.Motivo, total).Scan(&codigo)
		if errors.Is(err, pgx.ErrNoRows) {
			nuevo = false
			return t.QueryRow(ctx, `SELECT codigo FROM devoluciones WHERE pedido_numero = $1 AND origen = 'arrepentimiento'`, numero).Scan(&codigo)
		}
		return err
	})
	if err != nil {
		return "", err
	}
	// El mail sale una vez, con el trámite nuevo (no en cada repetición).
	if s.Correo != nil && nuevo {
		if err := s.Correo.Enviar(ctx, mailRecibido(p.Email, p.Nombre, nombreTienda, codigo)); err != nil {
			slog.Warn("devoluciones: no salió el mail", "codigo", codigo, "error", err)
		}
	}
	return codigo, nil
}

func mailRecibido(para, nombre, tienda, codigo string) correo.Mensaje {
	n, t, c := html.EscapeString(nombre), html.EscapeString(tienda), html.EscapeString(codigo)
	return correo.Mensaje{
		Para:   para,
		Asunto: "Recibimos tu pedido de arrepentimiento · " + codigo,
		Texto: fmt.Sprintf("Hola, %s.\n\n%s recibió tu pedido de arrepentimiento de la compra.\nTu código de trámite es %s.\n\n"+
			"La tienda se va a comunicar con vos para coordinar la devolución del producto y el reintegro.\n\nTu Local Web, en nombre de %s", nombre, tienda, codigo, tienda),
		HTML: fmt.Sprintf(`<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:'Instrument Sans',Helvetica,Arial,sans-serif;color:#2A140C">
<div style="max-width:560px;margin:0 auto;padding:32px 24px">
<div style="font-size:20px;font-weight:700;letter-spacing:-.02em">%s</div>
<h1 style="margin:28px 0 0;font-size:28px;line-height:1.1;letter-spacing:-.03em;font-weight:600">Recibimos tu pedido</h1>
<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">Hola, %s. La tienda recibió tu pedido de arrepentimiento de la compra.</p>
<p style="margin:18px 0 0;padding:14px 16px;border-radius:14px;background:#fff;font-size:15px">Código de trámite: <b>%s</b></p>
<p style="margin:18px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">La tienda se va a comunicar con vos para coordinar la devolución del producto y el&nbsp;reintegro.</p>
</div></body></html>`, t, n, c),
	}
}

// Devolucion es components/schemas/Devolucion.
type Devolucion struct {
	ID         string    `json:"id"`
	Codigo     string    `json:"codigo"`
	Producto   string    `json:"producto"`
	Foto       *string   `json:"foto"`
	Cliente    string    `json:"cliente"`
	Pedido     int       `json:"pedido"`
	Creada     time.Time `json:"creada"`
	Motivo     string    `json:"motivo"`
	Monto      int64     `json:"monto"`
	Resolucion string    `json:"resolucion"`
	Estado     string    `json:"estado"`
}

// Panel: las devoluciones de la tienda y la tasa (devoluciones / pedidos de los últimos 90 días, %).
func (s *Servicio) Panel(ctx context.Context, tienda string) ([]Devolucion, float64, error) {
	out := []Devolucion{}
	var tasa float64
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT d.id::text, d.codigo, coalesce(p.items->0->>'nombre', 'Pedido') ||
				CASE WHEN jsonb_array_length(p.items) > 1 THEN ' y ' || (jsonb_array_length(p.items) - 1) || ' más' ELSE '' END,
				p.items->0->>'foto', d.nombre, d.pedido_numero, d.creada,
				CASE WHEN d.motivo = '' THEN 'Arrepentimiento (no hace falta motivo)' ELSE d.motivo END,
				d.monto, d.resolucion, d.estado
			FROM devoluciones d JOIN pedidos p ON p.tienda_id = d.tienda_id AND p.numero = d.pedido_numero
			WHERE d.tienda_id = $1 ORDER BY d.creada DESC LIMIT 300`, tienda)
		if err != nil {
			return err
		}
		for filas.Next() {
			var d Devolucion
			if err := filas.Scan(&d.ID, &d.Codigo, &d.Producto, &d.Foto, &d.Cliente, &d.Pedido, &d.Creada, &d.Motivo, &d.Monto, &d.Resolucion, &d.Estado); err != nil {
				filas.Close()
				return err
			}
			out = append(out, d)
		}
		filas.Close()
		if err := filas.Err(); err != nil {
			return err
		}
		return t.QueryRow(ctx, `SELECT coalesce(round(100.0 * (SELECT count(*) FROM devoluciones WHERE tienda_id = $1 AND creada > now() - interval '90 days')
			/ nullif((SELECT count(*) FROM pedidos WHERE tienda_id = $1 AND creado_en > now() - interval '90 days'), 0), 1), 0)`, tienda).Scan(&tasa)
	})
	return out, tasa, err
}

// Resolver: aprobar, ofrecer cambio, o deshacer la última decisión.
func (s *Servicio) Resolver(ctx context.Context, tienda, id, accion string) error {
	return s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var q string
		switch accion {
		case "aprobar":
			q = `UPDATE devoluciones SET estado_anterior = estado, estado = 'Aprobada' WHERE id::text = $1 AND estado <> 'Aprobada'`
		case "cambio":
			q = `UPDATE devoluciones SET estado_anterior = estado, estado = 'Cambio ofrecido' WHERE id::text = $1 AND estado = 'Pendiente'`
		case "deshacer":
			q = `UPDATE devoluciones SET estado = coalesce(estado_anterior, 'Pendiente'), estado_anterior = NULL WHERE id::text = $1`
		default:
			return ErrInvalido
		}
		tag, err := t.Exec(ctx, q, id)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrNoEncontrado
		}
		return err
	})
}

// Rutas del panel (con la sesión de la tienda).
func (s *Servicio) Rutas(mux *http.ServeMux, tienda func(*http.Request) string) {
	con := func(f func(w http.ResponseWriter, r *http.Request, id string)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id := tienda(r)
			if id == "" {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			f(w, r, id)
		}
	}
	mux.HandleFunc("GET /v1/panel/devoluciones", con(func(w http.ResponseWriter, r *http.Request, id string) {
		ds, tasa, err := s.Panel(r.Context(), id)
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(map[string]any{"devoluciones": ds, "tasa": tasa})
	}))
	mux.HandleFunc("POST /v1/panel/devoluciones/{id}", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var b struct{ Accion string }
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&b) != nil {
			problema.Escribir(w, problema.Invalido("Pedido inválido."))
			return
		}
		switch err := s.Resolver(r.Context(), id, r.PathValue("id"), b.Accion); {
		case errors.Is(err, ErrInvalido):
			problema.Escribir(w, problema.Invalido("Esa acción no existe."))
			return
		case errors.Is(err, ErrNoEncontrado):
			problema.Escribir(w, problema.NoEncontrado("Esa devolución no existe o ya no admite esa acción."))
			return
		case err != nil:
			problema.Escribir(w, problema.Interno())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
}
