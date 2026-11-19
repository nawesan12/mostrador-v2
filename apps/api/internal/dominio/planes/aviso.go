package planes

import (
	"context"
	"math"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// TipoFinDePrueba: el evento del outbox que avisa a la dueña que la prueba termina (a 3 días) o
// terminó (el día que pasa a Gratis), con lo que se oculta. La campanita lo arma (notificaciones).
const TipoFinDePrueba = "prueba.termina"

// FinDePrueba son los datos del evento: los días que faltan (0: terminó) y lo que queda fuera de
// Gratis (productos ocultos y personas en sólo lectura).
type FinDePrueba struct {
	Dias      int `json:"dias"`
	Productos int `json:"productos"`
	Personas  int `json:"personas"`
}

type pools interface {
	EnTx(ctx context.Context, fn func(pgx.Tx) error) error
	EnTienda(ctx context.Context, tiendaID string, fn func(pgx.Tx) error) error
}

// AvisarFinDePrueba encola el aviso de cada tienda en Gratis cuya prueba termina en 3 días o ya
// terminó, una vez por momento (el worker, cada pocos minutos). Devuelve cuántos encoló.
func AvisarFinDePrueba(ctx context.Context, pool pools, ahora time.Time) (int, error) {
	var ids []string
	if err := pool.EnTx(ctx, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT id::text FROM app.pruebas_por_terminar(100) id`)
		if err != nil {
			return err
		}
		ids, err = pgx.CollectRows(filas, pgx.RowTo[string])
		return err
	}); err != nil {
		return 0, err
	}
	n := 0
	for _, id := range ids {
		err := pool.EnTienda(ctx, id, func(t pgx.Tx) error {
			var plan string
			var fin time.Time
			var avisado *int
			var avisadoFin *time.Time
			var productos, miembros int
			if err := t.QueryRow(ctx, `SELECT plan, coalesce(prueba_hasta, creada_en + interval '14 days'), aviso_prueba, aviso_prueba_fin,
					(SELECT count(*) FROM productos WHERE tienda_id = $1), (SELECT count(*) FROM auth.membresias WHERE tienda_id = $1)
				FROM tiendas WHERE id = $1 FOR UPDATE`, id).Scan(&plan, &fin, &avisado, &avisadoFin, &productos, &miembros); err != nil {
				return err
			}
			umbral := 3
			if !fin.After(ahora) {
				umbral = 0
			}
			mismoFin := avisadoFin != nil && avisadoFin.Equal(fin)
			if plan != "gratis" || (mismoFin && avisado != nil && *avisado <= umbral) {
				return nil
			}
			g := De("gratis")
			d := FinDePrueba{Dias: max(0, int(math.Ceil(fin.Sub(ahora).Hours()/24))),
				Productos: max(0, productos-g.Productos), Personas: max(0, miembros-g.Personas)}
			if _, err := t.Exec(ctx, `UPDATE tiendas SET aviso_prueba = $2, aviso_prueba_fin = $3 WHERE id = $1`, id, umbral, fin); err != nil {
				return err
			}
			n++
			return outbox.Publicar(ctx, t, id, TipoFinDePrueba, d)
		})
		if err != nil {
			return n, err
		}
	}
	return n, nil
}
