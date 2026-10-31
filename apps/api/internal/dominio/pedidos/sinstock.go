package pedidos

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// SinStock son los datos de los eventos tienda.sin_stock (mail) y push.sin_stock (iPhone).
type SinStock struct {
	Producto string `json:"producto"`
	// Variante: la que llegó a 0, si el producto tiene opciones.
	Variante string `json:"variante,omitempty"`
}

// avisarSinStock encola «Te quedaste sin stock» cuando una venta deja en 0 un producto publicado (o
// una de sus variantes), en la misma transacción que descontó. Uno por producto por día de
// Argentina: productos.sin_stock_avisado guarda el día y la fila ya está bloqueada por el
// descuento, así que de dos ventas a la vez avisa una sola. El worker decide a quién y por dónde
// (docs/decisiones.md §41).
func avisarSinStock(ctx context.Context, t pgx.Tx, tiendaID string, i Linea) error {
	var id string
	err := t.QueryRow(ctx, `UPDATE productos SET sin_stock_avisado = (now() AT TIME ZONE 'America/Argentina/Buenos_Aires')::date
		WHERE tienda_id = $1 AND id::text = $2 AND publicado
			AND sin_stock_avisado IS DISTINCT FROM (now() AT TIME ZONE 'America/Argentina/Buenos_Aires')::date
		RETURNING id::text`, tiendaID, i.ProductoID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	d := SinStock{Producto: id, Variante: i.VarianteID}
	if err := outbox.Publicar(ctx, t, tiendaID, "tienda.sin_stock", d); err != nil {
		return err
	}
	return outbox.Publicar(ctx, t, tiendaID, "push.sin_stock", d)
}

// Agotado es lo que el aviso dice del producto, leído al mandarlo.
type Agotado struct {
	Nombre string
	// Variante: «Talle M · Color Negro», si el aviso es de una variante.
	Variante string
	// Sigue: todavía publicado y en 0 (si ya repusieron o lo pausaron, el aviso no sale).
	Sigue bool
}

// LeerSinStock lee el producto (y la variante) del aviso con la tienda fijada. pgx.ErrNoRows si ya
// no existe.
func LeerSinStock(ctx context.Context, t pgx.Tx, d SinStock) (Agotado, error) {
	var a Agotado
	var stock int
	var publicado bool
	if d.Variante == "" {
		err := t.QueryRow(ctx, `SELECT nombre, stock, publicado FROM productos WHERE id::text = $1`, d.Producto).
			Scan(&a.Nombre, &stock, &publicado)
		a.Sigue = err == nil && publicado && stock == 0
		return a, err
	}
	var opciones []listas.Opcion
	var valores map[string]string
	var activa bool
	err := t.QueryRow(ctx, `SELECT p.nombre, p.opciones, v.valores, v.stock, v.activa, p.publicado
		FROM variantes v JOIN productos p ON p.tienda_id = v.tienda_id AND p.id = v.producto_id
		WHERE v.producto_id::text = $1 AND v.id::text = $2`, d.Producto, d.Variante).
		Scan(&a.Nombre, &opciones, &valores, &stock, &activa, &publicado)
	if err != nil {
		return a, err
	}
	a.Variante = listas.TextoVariante(opciones, valores)
	a.Sigue = publicado && activa && stock == 0
	return a, nil
}
