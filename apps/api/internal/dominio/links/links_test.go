package links_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/links"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

var ctx = context.Background()

func TestLinkSeVeSinSesionYNoCruzaTiendas(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	cod, err := links.Crear(ctx, b.App, luna, links.Nuevo{Monto: 42900, Concepto: "Termo para Lucía"})
	if err != nil {
		t.Fatal(err)
	}
	l, err := links.Ver(ctx, b.App, cod)
	if err != nil || l.Monto != 42900 || !l.Cuotas || l.Pagado || l.Tienda == "" {
		t.Fatalf("link: %+v %v", l, err)
	}
	if _, err := links.Ver(ctx, b.App, "ZZZZZZ"); !errors.Is(err, links.ErrNoExiste) {
		t.Fatalf("código inexistente: %v", err)
	}
	if _, err := links.Crear(ctx, b.App, sol, links.Nuevo{Monto: 0, Concepto: "x"}); err == nil {
		t.Fatal("aceptó un link sin monto")
	}
	// Desde la otra tienda, el link no se ve.
	var n int
	_ = b.App.EnTienda(ctx, sol, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT count(*) FROM links`).Scan(&n)
	})
	if n != 0 {
		t.Fatalf("la otra tienda ve %d links", n)
	}
}
