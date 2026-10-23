package ventas

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

var (
	ErrCarritoInvalido = errors.New("ventas: carrito inválido")
	ErrNoExisteCarrito = errors.New("ventas: ese carrito no existe")
	ErrRecienRecordado = errors.New("ventas: ya se le recordó hace menos de un día")
)

var emailCarrito = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)

// LineaCarrito: lo que manda el checkout (los precios salen de la tienda, no del navegador).
type LineaCarrito struct {
	ID         string  `json:"id"`
	VarianteID *string `json:"varianteId,omitempty"`
	Cantidad   int     `json:"cantidad"`
}

func nuevoToken() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// GuardarCarrito guarda (o actualiza) el carrito abierto de ese email. Sólo productos publicados de
// la tienda; el total se calcula con sus precios.
func GuardarCarrito(ctx context.Context, pool tx, tienda, email, nombre string, lineas []LineaCarrito) error {
	email = strings.ToLower(strings.TrimSpace(email))
	nombre = strings.TrimSpace(nombre)
	if !emailCarrito.MatchString(email) || len(email) > 120 || len([]rune(nombre)) > 80 || len(lineas) == 0 || len(lineas) > 50 {
		return ErrCarritoInvalido
	}
	return pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var total int64
		var validas []LineaCarrito
		for _, l := range lineas {
			if l.Cantidad < 1 || l.Cantidad > 99 {
				return ErrCarritoInvalido
			}
			var precio int64
			err := t.QueryRow(ctx, `SELECT precio FROM productos WHERE id::text = $1 AND publicado`, l.ID).Scan(&precio)
			if err == nil && l.VarianteID != nil {
				// Con variante, su precio (si tiene uno propio); una que ya no se vende, se ignora.
				var propio *int64
				err = t.QueryRow(ctx, `SELECT precio FROM variantes WHERE producto_id::text = $1 AND id::text = $2
					AND activa`, l.ID, *l.VarianteID).Scan(&propio)
				if propio != nil {
					precio = *propio
				}
			}
			if errors.Is(err, pgx.ErrNoRows) {
				continue // un producto (o variante) que ya no está: se ignora
			}
			if err != nil {
				return err
			}
			total += precio * int64(l.Cantidad)
			validas = append(validas, l)
		}
		if len(validas) == 0 {
			return ErrCarritoInvalido
		}
		items, _ := json.Marshal(validas)
		_, err := t.Exec(ctx, `INSERT INTO carritos (tienda_id, email, nombre, items, total, token)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (tienda_id, lower(email)) WHERE recuperado_en IS NULL
			DO UPDATE SET items = excluded.items, total = excluded.total, actualizado = now(),
				nombre = CASE WHEN excluded.nombre <> '' THEN excluded.nombre ELSE carritos.nombre END`,
			tienda, email, nombre, items, total, nuevoToken())
		return err
	})
}

// CarritoRecuperado cierra el carrito de ese email cuando compra (lo llama la vidriera).
func CarritoRecuperado(ctx context.Context, t pgx.Tx, email string) error {
	_, err := t.Exec(ctx, `UPDATE carritos SET recuperado_en = now() WHERE lower(email) = lower($1) AND recuperado_en IS NULL`, email)
	return err
}

// CarritoPorToken devuelve las líneas para el link del recordatorio (restaura el carrito).
func CarritoPorToken(ctx context.Context, pool tx, tienda, token string) ([]LineaCarrito, error) {
	var out []LineaCarrito
	err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var b []byte
		if err := t.QueryRow(ctx, `SELECT items FROM carritos WHERE token = $1`, token).Scan(&b); err != nil {
			return err
		}
		return json.Unmarshal(b, &out)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNoExisteCarrito
	}
	return out, err
}

// CarritoPanel es un carrito como lo muestra el panel (components/schemas/Carritos).
type CarritoPanel struct {
	ID        string    `json:"id"`
	Nombre    string    `json:"nombre"`
	Foto      *string   `json:"foto"`
	Desde     string    `json:"desde"`
	Creado    time.Time `json:"creado"`
	Fotos     []*string `json:"fotos"`
	Cantidad  int       `json:"cantidad"`
	Total     int64     `json:"total"`
	Recordado bool      `json:"recordado"`
}

// CarritosPanel: los abiertos de los últimos 30 días y los números del mes.
func CarritosPanel(ctx context.Context, pool tx, tienda string, ahora time.Time) (cs []CarritoPanel, abiertos int, monto, recuperado int64, tasa float64, err error) {
	cs = []CarritoPanel{}
	err = pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT c.id::text, coalesce(nullif(c.nombre, ''), split_part(c.email, '@', 1)), c.creado, c.items, c.total,
				c.recordado_en IS NOT NULL
			FROM carritos c WHERE c.tienda_id = $1 AND c.recuperado_en IS NULL AND c.actualizado > $2
			ORDER BY c.actualizado DESC LIMIT 200`, tienda, ahora.Add(-30*24*time.Hour))
		if err != nil {
			return err
		}
		for filas.Next() {
			var c CarritoPanel
			var items []byte
			if err := filas.Scan(&c.ID, &c.Nombre, &c.Creado, &items, &c.Total, &c.Recordado); err != nil {
				filas.Close()
				return err
			}
			var ls []LineaCarrito
			_ = json.Unmarshal(items, &ls)
			for _, l := range ls {
				c.Cantidad += l.Cantidad
			}
			c.Desde = "Tienda online"
			cs = append(cs, c)
			abiertos++
			monto += c.Total
		}
		filas.Close()
		if err := filas.Err(); err != nil {
			return err
		}
		// Fotos de los productos de cada carrito (hasta 3).
		for i := range cs {
			var items []byte
			if err := t.QueryRow(ctx, `SELECT items FROM carritos WHERE id::text = $1`, cs[i].ID).Scan(&items); err != nil {
				return err
			}
			var ls []LineaCarrito
			_ = json.Unmarshal(items, &ls)
			for _, l := range ls {
				if len(cs[i].Fotos) == 3 {
					break
				}
				var f *string
				_ = t.QueryRow(ctx, `SELECT foto FROM productos WHERE id::text = $1`, l.ID).Scan(&f)
				cs[i].Fotos = append(cs[i].Fotos, f)
			}
			if len(cs[i].Fotos) > 0 {
				cs[i].Foto = cs[i].Fotos[0]
			}
			if cs[i].Fotos == nil {
				cs[i].Fotos = []*string{}
			}
		}
		inicioMes := time.Date(ahora.Year(), ahora.Month(), 1, 0, 0, 0, 0, ahora.Location())
		// Recuperado este mes: los que se cerraron con compra después de un recordatorio.
		var recordados, recup int
		if err := t.QueryRow(ctx, `SELECT count(*) FILTER (WHERE recordado_en IS NOT NULL),
				count(*) FILTER (WHERE recuperado_en IS NOT NULL AND recordado_en IS NOT NULL),
				coalesce(sum(total) FILTER (WHERE recuperado_en IS NOT NULL AND recordado_en IS NOT NULL), 0)
			FROM carritos WHERE tienda_id = $1 AND creado >= $2`, tienda, inicioMes).Scan(&recordados, &recup, &recuperado); err != nil {
			return err
		}
		if recordados > 0 {
			tasa = float64(recup) * 100 / float64(recordados)
		}
		return nil
	})
	return
}

// RecordarCarrito encola el mail con el link que restaura el carrito. Uno por día como mucho.
func RecordarCarrito(ctx context.Context, pool tx, tienda, id string, ahora time.Time) error {
	return pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var recordado *time.Time
		err := t.QueryRow(ctx, `SELECT recordado_en FROM carritos WHERE id::text = $1 AND recuperado_en IS NULL FOR UPDATE`, id).Scan(&recordado)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExisteCarrito
		}
		if err != nil {
			return err
		}
		if recordado != nil && ahora.Sub(*recordado) < 24*time.Hour {
			return ErrRecienRecordado
		}
		if _, err := t.Exec(ctx, `UPDATE carritos SET recordado_en = $2, recordatorios = recordatorios + 1 WHERE id::text = $1`, id, ahora); err != nil {
			return err
		}
		return outbox.Publicar(ctx, t, tienda, "carrito.recordar", map[string]string{"id": id})
	})
}

// RecordatoriosAutomaticos encola el mail de los carritos de 3 días de las tiendas que lo prendieron.
func RecordatoriosAutomaticos(ctx context.Context, pool interface {
	tx
	EnTx(context.Context, func(pgx.Tx) error) error
}, ahora time.Time) (int, error) {
	type par struct{ tienda, id string }
	var pares []par
	err := pool.EnTx(ctx, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT tienda_id::text, id::text FROM app.carritos_para_recordar(100)`)
		if err != nil {
			return err
		}
		defer filas.Close()
		for filas.Next() {
			var p par
			if err := filas.Scan(&p.tienda, &p.id); err != nil {
				return err
			}
			pares = append(pares, p)
		}
		return filas.Err()
	})
	if err != nil {
		return 0, err
	}
	n := 0
	for _, p := range pares {
		if RecordarCarrito(ctx, pool, p.tienda, p.id, ahora) == nil {
			n++
		}
	}
	return n, nil
}

// Mails de carritos (el worker): "¿Te olvidaste algo?" con el link que restaura el carrito.
type MailsCarritos struct {
	Pool   tx
	Correo correo.Correo
	Sitio  string
}

func (m *MailsCarritos) Manejadores() map[string]outbox.Manejador {
	return map[string]outbox.Manejador{"carrito.recordar": m.recordar}
}

func (m *MailsCarritos) recordar(ctx context.Context, ev outbox.Evento) error {
	var d struct{ ID string }
	if json.Unmarshal(ev.Datos, &d) != nil || m.Correo == nil {
		return nil
	}
	var email, nombre, token, tienda, slug string
	var items []byte
	var total int64
	var recuperado bool
	err := m.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT c.email, c.nombre, c.token, c.items, c.total, c.recuperado_en IS NOT NULL, t.nombre, t.slug
			FROM carritos c JOIN tiendas t ON t.id = c.tienda_id WHERE c.id::text = $1`, d.ID).
			Scan(&email, &nombre, &token, &items, &total, &recuperado, &tienda, &slug)
	})
	if errors.Is(err, pgx.ErrNoRows) || recuperado {
		return nil // compró mientras tanto: no se manda
	}
	if err != nil {
		return err
	}
	var ls []LineaCarrito
	_ = json.Unmarshal(items, &ls)
	var nombres []string
	_ = m.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		for _, l := range ls {
			var n string
			if t.QueryRow(ctx, `SELECT nombre FROM productos WHERE id::text = $1`, l.ID).Scan(&n) == nil {
				nombres = append(nombres, fmt.Sprintf("%d × %s", l.Cantidad, n))
			}
		}
		return nil
	})
	link := fmt.Sprintf("%s/t/%s/carrito/recuperar/%s", strings.TrimRight(m.Sitio, "/"), slug, token)
	hola := "Hola"
	if nombre != "" {
		hola += ", " + strings.Split(nombre, " ")[0]
	}
	texto := fmt.Sprintf("%s.\n\nDejaste esto en tu carrito de %s:\n\n%s\n\nSi lo querés, está guardado acá: %s\n\n%s",
		hola, tienda, strings.Join(nombres, "\n"), link, tienda)
	var b strings.Builder
	fmt.Fprintf(&b, `<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:Helvetica,Arial,sans-serif;color:#2A140C"><div style="max-width:560px;margin:0 auto;padding:32px 24px"><div style="font-size:20px;font-weight:700">%s</div><h1 style="margin:24px 0 0;font-size:26px;font-weight:600">¿Te olvidaste algo?</h1><p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">%s. Dejaste esto en tu carrito:</p><div style="margin:14px 0 0;padding:14px 16px;border-radius:14px;background:#fff;font-size:14.5px;line-height:1.7">`,
		html.EscapeString(tienda), html.EscapeString(hola))
	for _, n := range nombres {
		fmt.Fprintf(&b, "<div>%s</div>", html.EscapeString(n))
	}
	fmt.Fprintf(&b, `</div><p style="margin:22px 0 0"><a href="%s" style="display:inline-block;padding:12px 18px;border-radius:999px;background:#2A140C;color:#fff;text-decoration:none;font-weight:600">Volver a mi carrito</a></p><p style="margin:28px 0 0;font-size:12px;color:#9C8676">Te escribimos porque dejaste tu email en el checkout de %s.</p></div></body></html>`,
		html.EscapeString(link), html.EscapeString(tienda))
	return m.Correo.Enviar(ctx, correo.Mensaje{Para: email, Asunto: "¿Te olvidaste algo en " + tienda + "?", Texto: texto, HTML: b.String()})
}

// RecordarTodos: los abiertos de la semana que no se recordaron en el último día.
func RecordarTodos(ctx context.Context, pool tx, tienda string, ahora time.Time) (int, error) {
	var ids []string
	err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT id::text FROM carritos WHERE recuperado_en IS NULL AND actualizado > $1
			AND (recordado_en IS NULL OR recordado_en < $2) LIMIT 200`, ahora.Add(-7*24*time.Hour), ahora.Add(-24*time.Hour))
		if err != nil {
			return err
		}
		defer filas.Close()
		for filas.Next() {
			var id string
			if err := filas.Scan(&id); err != nil {
				return err
			}
			ids = append(ids, id)
		}
		return filas.Err()
	})
	n := 0
	for _, id := range ids {
		if RecordarCarrito(ctx, pool, tienda, id, ahora) == nil {
			n++
		}
	}
	return n, err
}
