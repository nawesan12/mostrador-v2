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

// SlugVigente: la dirección actual de la tienda que antes se llamaba slug (decisiones §38), para
// que /t/<viejo>/… responda 301 a /t/<nuevo>/…. ErrNoExiste si ese slug no fue de ninguna tienda o
// la tienda está cerrada o pausada.
func SlugVigente(ctx context.Context, pool tx, slug string) (string, error) {
	if !slugValido.MatchString(slug) {
		return "", ErrNoExiste
	}
	var nuevo *string
	if err := pool.EnTx(ctx, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT app.slug_vigente($1)`, slug).Scan(&nuevo)
	}); err != nil {
		return "", err
	}
	if nuevo == nil {
		return "", ErrNoExiste
	}
	return *nuevo, nil
}
