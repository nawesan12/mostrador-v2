package vidriera

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

// TiendaYPedido: la tienda abierta de un slug y, si viene un código público, el número de ese
// pedido de esa tienda (el tiempo real de la tienda online, docs/decisiones.md §42). Un código que
// no es de esa tienda da ErrNoExiste, igual que uno que no existe.
func TiendaYPedido(ctx context.Context, pool tx, slug, codigo string) (tienda, pedido string, err error) {
	tienda, err = tiendaDe(ctx, pool, slug)
	if err != nil || codigo == "" {
		return tienda, "", err
	}
	err = pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT numero::text FROM pedidos WHERE codigo_publico = $1`, codigo).Scan(&pedido)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", "", ErrNoExiste
	}
	return tienda, pedido, err
}
