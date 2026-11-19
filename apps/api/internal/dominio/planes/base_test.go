package planes_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/planes"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// app.plan_incluye (migración 00097) es la copia de la regla para lo que decide la base (facturas,
// colas del worker, sitemap): tiene que coincidir con Go en cada plan, función y estado de la prueba.
func TestIgualQueLaBase(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	ahora := time.Now()
	for _, p := range planes.Todos {
		for _, prueba := range []bool{false, true} {
			fin := ahora.Add(-time.Hour)
			if prueba {
				fin = ahora.Add(24 * time.Hour)
			}
			vigente := planes.Efectivo(p.ID, fin, ahora)
			for _, f := range planes.De("negocio").Funciones {
				var tiene bool
				if err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
					return tx.QueryRow(ctx, `SELECT app.plan_incluye($1, $2, $3)`, p.ID, fin, string(f)).Scan(&tiene)
				}); err != nil {
					t.Fatal(err)
				}
				if tiene != vigente.Tiene(f) {
					t.Errorf("%s (prueba %v) %s: base %v, Go %v", p.ID, prueba, f, tiene, vigente.Tiene(f))
				}
			}
		}
	}
}
