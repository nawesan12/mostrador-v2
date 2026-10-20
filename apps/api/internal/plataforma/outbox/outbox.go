// Package outbox publica eventos de dominio en la misma transacción que el cambio que los causa:
// no existe "se guardó el pedido pero el aviso nunca salió". Un worker los consume con reintentos.
package outbox

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
)

// Publicar encola un evento en la transacción (con la tienda fijada). tipo es "entidad.hecho",
// por ejemplo "pedido.pagado".
func Publicar(ctx context.Context, tx pgx.Tx, tiendaID, tipo string, datos any) error {
	b, err := json.Marshal(datos)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO outbox (tienda_id, tipo, datos) VALUES ($1, $2, $3)`, tiendaID, tipo, b)
	return err
}
