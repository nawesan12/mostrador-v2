package vidriera

import (
	"context"
	"sync"
	"time"
)

// Caché en memoria de la vidriera pública (GET /v1/publico/tiendas/{slug}): cada vista de una tienda
// armaba Publica entera, con el ranking de 30 días. Dura VidaCache y se olvida la tienda con
// cualquier aviso de tiempo real suyo (Olvidar, desde el hub): un precio o un stock nuevo se ve en el
// momento. Comprar no la usa: el stock lo valida CrearDesdeProductos en la transacción.
const VidaCache = 30 * time.Second

type entrada struct {
	t      Tienda
	expira time.Time
}

var cache = struct {
	sync.Mutex
	porSlug map[string]entrada
	slugDe  map[string]string // id de la tienda → slug
}{porSlug: map[string]entrada{}, slugDe: map[string]string{}}

// PublicaConCache: Publica con el caché (para la vista pública).
func PublicaConCache(ctx context.Context, pool tx, slug string) (Tienda, error) {
	ahora := time.Now()
	cache.Lock()
	e, ok := cache.porSlug[slug]
	cache.Unlock()
	if ok && ahora.Before(e.expira) {
		return e.t, nil
	}
	t, err := Publica(ctx, pool, slug)
	if err != nil {
		return t, err
	}
	id, _ := tiendaDe(ctx, pool, slug)
	cache.Lock()
	// Tope de tamaño simple: con muchas tiendas, se vacía y vuelve a llenarse.
	if len(cache.porSlug) > 2000 {
		cache.porSlug, cache.slugDe = map[string]entrada{}, map[string]string{}
	}
	cache.porSlug[slug] = entrada{t: t, expira: ahora.Add(VidaCache)}
	if id != "" {
		cache.slugDe[id] = slug
	}
	cache.Unlock()
	return t, nil
}

// Olvidar saca del caché la tienda de ese id (cambió algo suyo).
func Olvidar(tiendaID string) {
	cache.Lock()
	if slug, ok := cache.slugDe[tiendaID]; ok {
		delete(cache.porSlug, slug)
	}
	cache.Unlock()
}
