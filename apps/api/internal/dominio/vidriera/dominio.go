package vidriera

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/planes"
)

// El nombre de un dominio como lo guarda la tabla dominios (00020).
var dominioValido = regexp.MustCompile(`^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$`)

// SlugDeDominio: la tienda que se ve en un dominio propio ("www." vale como el dominio), sólo si
// el dominio está verificado y la tienda abierta. ErrNoExiste si no. propio es false si el plan que
// rige no incluye dominio propio (Gratis, o al terminar la prueba): la web responde 301 a la
// dirección de Tu Local Web con el mismo camino (los links compartidos siguen andando) y el dominio
// queda guardado para cuando vuelva a subir.
func SlugDeDominio(ctx context.Context, pool tx, dominio string) (slug string, propio bool, err error) {
	if slug, err = slugDeDominio(ctx, pool, dominio); err != nil {
		return "", false, err
	}
	id, err := tiendaDe(ctx, pool, slug)
	if err != nil {
		return "", false, err
	}
	err = pool.EnTienda(ctx, id, func(t pgx.Tx) error {
		p, err := planes.Vigente(ctx, t, id, time.Now())
		propio = p.Tiene(planes.Dominio)
		return err
	})
	return slug, propio, err
}

func slugDeDominio(ctx context.Context, pool tx, dominio string) (string, error) {
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
