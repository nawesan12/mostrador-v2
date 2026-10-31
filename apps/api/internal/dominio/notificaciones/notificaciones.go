// Package notificaciones: la campanita del panel (docs/decisiones.md §45).
//
// El worker escribe una notificación por evento y por persona del equipo que puede verlo (Generador,
// que envuelve los manejadores del outbox que ya mandan mail y push). La API las lista, las cuenta y
// las marca leídas, siempre las de la persona de la sesión en la tienda de la sesión (RLS por tienda
// y por persona, migración 00062). Se guardan 90 días.
package notificaciones

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// Pool: el de la app (rol mostrador_app). *db.Pool lo cumple.
type Pool interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
	EnTx(context.Context, func(pgx.Tx) error) error
}

// Notificacion es una fila de la campanita (components/schemas/Notificacion).
type Notificacion struct {
	ID      string    `json:"id"`
	Tipo    string    `json:"tipo"`
	Titulo  string    `json:"titulo"`
	Detalle string    `json:"detalle"`
	Creada  time.Time `json:"creada"`
	Leida   bool      `json:"leida"`
	Grupo   string    `json:"grupo"`
	Ruta    string    `json:"ruta"`
}

// Pagina es lo que responde la lista (components/schemas/PaginaNotificaciones).
type Pagina struct {
	Notificaciones []Notificacion `json:"notificaciones"`
	SinLeer        int            `json:"sinLeer"`
	Siguiente      *string        `json:"siguiente"`
}

// Retencion: lo que dura una notificación.
const Retencion = 90 * 24 * time.Hour

// PorPagina y MaxPorPagina: cuántas trae la lista.
const (
	PorPagina    = 20
	MaxPorPagina = 50
)

// ErrCursor: el cursor no tiene la forma que da la API.
var ErrCursor = errors.New("cursor inválido")

// enPersona corre fn con la tienda y la persona fijadas: RLS deja ver y marcar sólo las suyas.
func enPersona(ctx context.Context, p Pool, tienda, usuario string, fn func(pgx.Tx) error) error {
	if usuario == "" {
		return errors.New("notificaciones: sin persona")
	}
	return p.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		if _, err := t.Exec(ctx, "SELECT set_config('app.usuario_id', $1, true)", usuario); err != nil {
			return err
		}
		return fn(t)
	})
}

// cursor: «<micro segundos de creada_en>.<id>» de la última fila de la página.
func armarCursor(n Notificacion) string {
	return strconv.FormatInt(n.Creada.UnixMicro(), 10) + "." + n.ID
}

func leerCursor(c string) (time.Time, string, error) {
	us, id, ok := strings.Cut(c, ".")
	n, err := strconv.ParseInt(us, 10, 64)
	if !ok || err != nil || len(id) != 36 {
		return time.Time{}, "", ErrCursor
	}
	return time.UnixMicro(n), id, nil
}

// Listar: las de la persona, de la más nueva a la más vieja, de a limite, con cuántas no leyó.
func Listar(ctx context.Context, p Pool, tienda, usuario, cursor string, limite int) (Pagina, error) {
	pg := Pagina{Notificaciones: []Notificacion{}}
	if limite <= 0 {
		limite = PorPagina
	}
	limite = min(limite, MaxPorPagina)
	var desde time.Time
	var desdeID string
	if cursor != "" {
		var err error
		if desde, desdeID, err = leerCursor(cursor); err != nil {
			return pg, err
		}
	}
	err := enPersona(ctx, p, tienda, usuario, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT id::text, tipo, titulo, detalle, creada_en, leida_en IS NOT NULL, grupo, ruta
			FROM notificaciones
			WHERE tienda_id = $1 AND usuario_id = $2 AND creada_en > now() - $5::interval
				AND ($3::text = '' OR (creada_en, id) < ($4::timestamptz, $3::uuid))
			ORDER BY creada_en DESC, id DESC LIMIT $6`,
			tienda, usuario, desdeID, desde, Retencion.String(), limite+1)
		if err != nil {
			return err
		}
		ns, err := pgx.CollectRows(filas, func(f pgx.CollectableRow) (Notificacion, error) {
			var n Notificacion
			err := f.Scan(&n.ID, &n.Tipo, &n.Titulo, &n.Detalle, &n.Creada, &n.Leida, &n.Grupo, &n.Ruta)
			return n, err
		})
		if err != nil {
			return err
		}
		if len(ns) > limite {
			ns = ns[:limite]
			c := armarCursor(ns[len(ns)-1])
			pg.Siguiente = &c
		}
		pg.Notificaciones = append(pg.Notificaciones, ns...)
		pg.SinLeer, err = contar(ctx, t, tienda, usuario)
		return err
	})
	return pg, err
}

func contar(ctx context.Context, t pgx.Tx, tienda, usuario string) (int, error) {
	var n int
	err := t.QueryRow(ctx, `SELECT count(*) FROM notificaciones WHERE tienda_id = $1 AND usuario_id = $2
		AND leida_en IS NULL AND creada_en > now() - $3::interval`, tienda, usuario, Retencion.String()).Scan(&n)
	return n, err
}

// SinLeer: cuántas no leyó la persona (la insignia de la campanita).
func SinLeer(ctx context.Context, p Pool, tienda, usuario string) (int, error) {
	var n int
	err := enPersona(ctx, p, tienda, usuario, func(t pgx.Tx) error {
		var err error
		n, err = contar(ctx, t, tienda, usuario)
		return err
	})
	return n, err
}

// ErrNoExiste: la notificación no existe o no es de la persona.
var ErrNoExiste = errors.New("notificación inexistente")

// MarcarLeida marca una de la persona. Ya leída no cambia la hora.
func MarcarLeida(ctx context.Context, p Pool, tienda, usuario, id string) error {
	if len(id) != 36 {
		return ErrNoExiste
	}
	return enPersona(ctx, p, tienda, usuario, func(t pgx.Tx) error {
		var existe bool
		if err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM notificaciones WHERE id::text = $1)`, id).Scan(&existe); err != nil {
			return err
		}
		if !existe {
			return ErrNoExiste
		}
		_, err := t.Exec(ctx, `UPDATE notificaciones SET leida_en = now() WHERE id::text = $1 AND leida_en IS NULL`, id)
		return err
	})
}

// MarcarTodas marca leídas todas las de la persona en la tienda. Devuelve cuántas cambió.
func MarcarTodas(ctx context.Context, p Pool, tienda, usuario string) (int64, error) {
	var n int64
	err := enPersona(ctx, p, tienda, usuario, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `UPDATE notificaciones SET leida_en = now()
			WHERE tienda_id = $1 AND usuario_id = $2 AND leida_en IS NULL`, tienda, usuario)
		n = tag.RowsAffected()
		return err
	})
	return n, err
}

// Purgar borra las de más de 90 días de todas las tiendas (app.purgar_notificaciones).
func Purgar(ctx context.Context, p Pool) (int64, error) {
	var n int64
	err := p.EnTx(ctx, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT app.purgar_notificaciones()`).Scan(&n)
	})
	if err != nil {
		return 0, fmt.Errorf("notificaciones: purgar: %w", err)
	}
	return n, nil
}
