package vidriera

import (
	"context"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
)

// El nombre de un dominio como lo guarda la tabla dominios (00020).
var dominioValido = regexp.MustCompile(`^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$`)

// SlugDeDominio: la tienda que se ve en un dominio propio ("www." vale como el dominio), sólo si
// el dominio está verificado y la tienda abierta. ErrNoExiste si no.
func SlugDeDominio(ctx context.Context, pool tx, dominio string) (string, error) {
	d := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(dominio)), "www.")
	if len(d) > 253 || !dominioValido.MatchString(d) {
		return "", ErrNoExiste
	}
	var slug *string
	if err := pool.EnTx(ctx, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT app.slug_de_dominio($1)`, d).Scan(&slug)
	}); err != nil {
		return "", err
	}
	if slug == nil {
		return "", ErrNoExiste
	}
	return *slug, nil
}
