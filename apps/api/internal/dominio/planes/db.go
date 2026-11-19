package planes

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/abono"
)

// Vigente: el plan que rige hoy para la tienda (dentro de su transacción).
func Vigente(ctx context.Context, t pgx.Tx, tiendaID string, ahora time.Time) (Plan, error) {
	e, err := abono.Leer(ctx, t, tiendaID)
	if err != nil {
		return Plan{}, err
	}
	return DeEstado(e, ahora), nil
}

// Exigir: nil si el plan vigente de la tienda incluye la función.
func Exigir(ctx context.Context, t pgx.Tx, tiendaID string, f Funcion) error {
	p, err := Vigente(ctx, t, tiendaID, time.Now())
	if err != nil {
		return err
	}
	return ExigirFuncion(p, f)
}

// ExigirProductoNuevo: antes de crear un producto. Toma la fila de la tienda para que dos altas a la
// vez no pasen el tope entre las dos.
func ExigirProductoNuevo(ctx context.Context, t pgx.Tx, tiendaID string) error {
	p, err := Vigente(ctx, t, tiendaID, time.Now())
	if err != nil || p.Productos == SinTope {
		return err
	}
	if _, err := t.Exec(ctx, `SELECT 1 FROM tiendas WHERE id = $1 FOR NO KEY UPDATE`, tiendaID); err != nil {
		return err
	}
	var n int
	if err := t.QueryRow(ctx, `SELECT count(*) FROM productos WHERE tienda_id = $1`, tiendaID).Scan(&n); err != nil {
		return err
	}
	return ExigirProductos(p, n+1)
}

// ExigirEditarProducto: un producto que quedó fuera del plan (después de bajar) no se edita; sí se
// puede borrar.
func ExigirEditarProducto(ctx context.Context, t pgx.Tx, tiendaID, productoID string) error {
	p, err := Vigente(ctx, t, tiendaID, time.Now())
	if err != nil || p.Productos == SinTope {
		return err
	}
	fuera, err := ProductoFuera(ctx, t, tiendaID, productoID, p)
	if err != nil || !fuera {
		return err
	}
	return productoFuera(p)
}

// ordenDelPlan: qué productos ocupan los lugares del plan. Primero los publicados en la tienda, y
// entre ellos los más viejos: un borrador nunca le quita el lugar a uno que se vende. Es el mismo
// orden en el panel (sólo lectura) y en la tienda (ocultos).
const ordenDelPlan = `ORDER BY (publicado AND coalesce((canales->>'web')::boolean, true)) DESC, creado_en, id`

// ProductoFuera: el producto está después de los primeros N del plan (ordenDelPlan).
func ProductoFuera(ctx context.Context, t pgx.Tx, tiendaID, productoID string, p Plan) (bool, error) {
	if p.Productos == SinTope {
		return false, nil
	}
	var fuera bool
	err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM (SELECT id FROM productos WHERE tienda_id = $1
			`+ordenDelPlan+` OFFSET $3) f WHERE f.id::text = $2)`, tiendaID, productoID, p.Productos).Scan(&fuera)
	return fuera, err
}

// FueraDelPlan: los ids de los productos que quedaron fuera del plan (vacío si entran todos).
func FueraDelPlan(ctx context.Context, t pgx.Tx, tiendaID string, p Plan) (map[string]bool, error) {
	out := map[string]bool{}
	if p.Productos == SinTope {
		return out, nil
	}
	filas, err := t.Query(ctx, `SELECT id::text FROM productos WHERE tienda_id = $1 `+ordenDelPlan+` OFFSET $2`,
		tiendaID, p.Productos)
	if err != nil {
		return nil, err
	}
	defer filas.Close()
	for filas.Next() {
		var id string
		if err := filas.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, filas.Err()
}

// SQLVisibles: la condición que deja en la tienda sólo los productos dentro del plan. Va con la tienda
// en $1 y el tope en el parámetro que se indique (tope 0 o negativo: todos). Ej.:
// `AND ` + SQLVisibles("$2").
func SQLVisibles(param string) string {
	return `(` + param + `::int <= 0 OR id NOT IN (SELECT id FROM productos WHERE tienda_id = $1 ` + ordenDelPlan + ` OFFSET ` +
		param + `::int))`
}

// ExigirPersonaNueva: antes de invitar a alguien. Cuenta el equipo y las invitaciones pendientes.
func ExigirPersonaNueva(ctx context.Context, t pgx.Tx, tiendaID string) error {
	p, err := Vigente(ctx, t, tiendaID, time.Now())
	if err != nil || p.Personas == SinTope {
		return err
	}
	if _, err := t.Exec(ctx, `SELECT 1 FROM tiendas WHERE id = $1 FOR NO KEY UPDATE`, tiendaID); err != nil {
		return err
	}
	// Las invitaciones sin aceptar (vencidas incluidas) reservan su lugar: se pueden reenviar sin
	// volver a pasar por el tope. Si son ellas las que lo llenan, se dice cómo liberar lugar.
	var miembros, invitaciones int
	if err := t.QueryRow(ctx, `SELECT (SELECT count(*) FROM auth.membresias WHERE tienda_id = $1),
			(SELECT count(*) FROM invitaciones WHERE tienda_id = $1)`, tiendaID).Scan(&miembros, &invitaciones); err != nil {
		return err
	}
	err = ExigirPersonas(p, miembros+invitaciones+1)
	var tope *Tope
	if invitaciones > 0 && errors.As(err, &tope) {
		cuales := "1 invitación sin aceptar: cancelala"
		if invitaciones > 1 {
			cuales = strconv.Itoa(invitaciones) + " invitaciones sin aceptar: cancelá alguna"
		}
		tope.Detalle = fmt.Sprintf("Tenés %s para liberar lugar, o pasá a %s.", cuales, tope.Sube.Nombre)
	}
	return err
}

// PersonaFuera: la persona entró después de las que entran en el plan (la dueña siempre entra). Sus
// cambios se frenan hasta que la dueña suba de plan o libere un lugar.
func PersonaFuera(ctx context.Context, t pgx.Tx, tiendaID, usuarioID string, p Plan) (bool, error) {
	if p.Personas == SinTope || usuarioID == "" {
		return false, nil
	}
	var fuera bool
	err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM (SELECT usuario_id FROM auth.membresias
			WHERE tienda_id = $1 AND rol <> 'duenio' ORDER BY creada_en, usuario_id OFFSET greatest($3::int - 1, 0)) f
		WHERE f.usuario_id::text = $2)`, tiendaID, usuarioID, p.Personas).Scan(&fuera)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return fuera, err
}
