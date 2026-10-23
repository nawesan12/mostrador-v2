package tiendas_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func TestOrigenSeFijaUnaSolaVez(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	if _, err := tiendas.FijarOrigen(ctx, b.App, luna, tiendas.Origen{Ref: "<script>"}); err != tiendas.ErrOrigen {
		t.Fatalf("valor raro: %v", err)
	}
	if ok, err := tiendas.FijarOrigen(ctx, b.App, luna, tiendas.Origen{Fuente: "flyer", Ref: "BRUMA-REF"}); err != nil || !ok {
		t.Fatalf("primera vez: %v %v", ok, err)
	}
	if ok, _ := tiendas.FijarOrigen(ctx, b.App, luna, tiendas.Origen{Ref: "OTRO"}); ok {
		t.Fatal("se pisó el origen")
	}
	var ref string
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT origen->>'ref' FROM tiendas WHERE id = $1`, luna).Scan(&ref)
	})
	if ref != "BRUMA-REF" {
		t.Fatalf("ref: %q", ref)
	}
}
