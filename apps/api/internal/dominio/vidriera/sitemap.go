package vidriera

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// EntradaSitemap es una tienda (Producto vacío) o un producto publicado, con su dominio propio
// verificado si lo tiene.
type EntradaSitemap struct {
	Tienda      string    `json:"tienda"`
	Dominio     *string   `json:"dominio"`
	Producto    *string   `json:"producto"`
	Actualizado time.Time `json:"actualizado"`
}

// maxSitemap: lo que entra en un sitemap (el estándar pide 50.000 URLs por archivo).
const maxSitemap = 49_000

// Sitemap lista lo público de todas las tiendas (app.sitemap, migración 00050).
func Sitemap(ctx context.Context, pool tx) ([]EntradaSitemap, error) {
	out := []EntradaSitemap{}
	err := pool.EnTx(ctx, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT tienda, dominio, producto, actualizado FROM app.sitemap($1)`, maxSitemap)
		if err != nil {
			return err
		}
		out, err = pgx.CollectRows(filas, pgx.RowToStructByPos[EntradaSitemap])
		return err
	})
	return out, err
}
