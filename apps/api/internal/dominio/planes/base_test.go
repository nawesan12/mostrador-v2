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

// El aviso del fin de la prueba sale una vez a 3 días y otra el día que termina; si el equipo la
// extiende, se vuelve a avisar con la fecha nueva.
func TestAvisoDelFinDeLaPrueba(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	admin, err := pgx.Connect(ctx, b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)
	fijar := func(sql string) {
		t.Helper()
		if _, err := admin.Exec(ctx, `UPDATE tiendas SET `+sql+` WHERE id = $1`, luna); err != nil {
			t.Fatal(err)
		}
	}
	avisos := func() []planes.FinDePrueba {
		var out []planes.FinDePrueba
		filas, _ := admin.Query(ctx, `SELECT datos FROM outbox WHERE tienda_id = $1 AND tipo = 'prueba.termina' ORDER BY id`, luna)
		defer filas.Close()
		for filas.Next() {
			var d planes.FinDePrueba
			_ = filas.Scan(&d)
			out = append(out, d)
		}
		return out
	}
	pasar := func() {
		t.Helper()
		if _, err := planes.AvisarFinDePrueba(ctx, b.App, time.Now()); err != nil {
			t.Fatal(err)
		}
	}
	pasar() // recién creada: nada
	if len(avisos()) != 0 {
		t.Fatal("avisó con 14 días por delante")
	}
	fijar(`creada_en = now() - interval '12 days'`)
	pasar()
	pasar()
	if a := avisos(); len(a) != 1 || a[0].Dias != 2 {
		t.Fatalf("a 2 días: %+v", a)
	}
	fijar(`creada_en = now() - interval '15 days'`)
	pasar()
	pasar()
	if a := avisos(); len(a) != 2 || a[1].Dias != 0 {
		t.Fatalf("al terminar: %+v", a)
	}
	fijar(`prueba_hasta = now() + interval '2 days'`)
	pasar()
	if a := avisos(); len(a) != 3 || a[2].Dias != 2 {
		t.Fatalf("con la prueba extendida: %+v", a)
	}
	// Con un plan pago no se avisa.
	fijar(`plan = 'emprende', prueba_hasta = NULL, aviso_prueba = NULL, aviso_prueba_fin = NULL`)
	pasar()
	if len(avisos()) != 3 {
		t.Fatal("avisó a un plan pago")
	}
}
