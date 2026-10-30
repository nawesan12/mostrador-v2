package mercadopago

import (
	"context"
	"regexp"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

var esIDPago = regexp.MustCompile(`^[0-9]{1,20}$`)

// AvisarVuelta: quien compra volvió de Mercado Pago con el id del pago. Se encola lo mismo que
// manda el webhook (mp.pago): el worker lo consulta en Mercado Pago con el token de la tienda y
// recién ahí marca el pedido. Así el pago se ve aunque el aviso de Mercado Pago tarde, y un id
// inventado no hace nada (Mercado Pago no lo encuentra en la cuenta de esa tienda).
func (s *Servicio) AvisarVuelta(ctx context.Context, slug, pago string) error {
	if !esIDPago.MatchString(pago) {
		return ErrInvalido
	}
	var tienda *string
	if err := s.Pool.EnTx(ctx, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT app.tienda_de_slug($1)::text`, slug).Scan(&tienda)
	}); err != nil {
		return err
	}
	if tienda == nil {
		return ErrNoExiste
	}
	return s.Pool.EnTienda(ctx, *tienda, func(t pgx.Tx) error {
		return outbox.Publicar(ctx, t, *tienda, "mp.pago", map[string]string{"id": pago})
	})
}
