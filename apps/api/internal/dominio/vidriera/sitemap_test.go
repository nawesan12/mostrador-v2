package vidriera_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// El sitemap trae las tiendas y sus productos publicados, de todas las tiendas; nada pausado.
func TestSitemap(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	una := pruebadb.CrearTienda(t, b.App, "una")
	otra := pruebadb.CrearTienda(t, b.App, "otra")
	stock := 3
	if _, err := listas.Guardar(ctx, b.App, una, "", listas.Editable{Nombre: "Yerba suave", Precio: 5000, Stock: &stock}); err != nil {
		t.Fatal(err)
	}
	p, _ := listas.Guardar(ctx, b.App, otra, "", listas.Editable{Nombre: "Termo", Precio: 9000, Stock: &stock})
	// Pausado: no va al sitemap.
	if err := b.App.EnTienda(ctx, otra, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE productos SET publicado = false WHERE id = $1`, p.ID)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	e, err := vidriera.Sitemap(ctx, b.App)
	if err != nil {
		t.Fatal(err)
	}
	tiendas, productos := map[string]bool{}, map[string]bool{}
	for _, x := range e {
		if x.Producto == nil {
			tiendas[x.Tienda] = true
		} else {
			productos[x.Tienda+"/"+*x.Producto] = true
		}
	}
	if !tiendas["una"] || !tiendas["otra"] || len(productos) != 1 || !productos["una/yerba-suave"] {
		t.Fatalf("sitemap: tiendas %v, productos %v", tiendas, productos)
	}
}
