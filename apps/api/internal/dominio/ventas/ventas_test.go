package ventas_test

import (
	"context"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/ventas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func TestAutomaticosPorTienda(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	una := pruebadb.CrearTienda(t, b.App, "una")
	otra := pruebadb.CrearTienda(t, b.App, "otra")
	if a, _ := ventas.LeerAutomaticos(ctx, b.App, una); !a.Hora || !a.Dia || a.Email {
		t.Fatalf("por defecto: %+v", a)
	}
	if err := ventas.GuardarAutomaticos(ctx, b.App, una, ventas.Automaticos{Email: true}); err != nil {
		t.Fatal(err)
	}
	a, _ := ventas.LeerAutomaticos(ctx, b.App, una)
	o, _ := ventas.LeerAutomaticos(ctx, b.App, otra)
	if a.Hora || !a.Email || !o.Hora {
		t.Fatalf("una %+v, otra %+v", a, o)
	}
}

func TestTarjetaPorTienda(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	una := pruebadb.CrearTienda(t, b.App, "una")
	otra := pruebadb.CrearTienda(t, b.App, "otra")
	if g, err := ventas.LeerTarjeta(ctx, b.App, una); err != nil || g != ventas.TarjetaInicial {
		t.Fatalf("de fábrica: %+v %v", g, err)
	}
	nueva := ventas.TarjetaInicial
	nueva.Modelo, nueva.QR, nueva.Cupon = "sello", "cupon", "VOLVE10"
	if msg := nueva.Valida(); msg != "" {
		t.Fatal(msg)
	}
	if err := ventas.GuardarTarjeta(ctx, b.App, una, nueva); err != nil {
		t.Fatal(err)
	}
	g, _ := ventas.LeerTarjeta(ctx, b.App, una)
	o, _ := ventas.LeerTarjeta(ctx, b.App, otra)
	if g != nueva || o != ventas.TarjetaInicial {
		t.Fatalf("una %+v, otra %+v", g, o)
	}
	for _, mal := range []ventas.Tarjeta{
		{Modelo: "otro", Titulo: "Hola", QR: "resena"},
		{Modelo: "arco", Titulo: " ", QR: "resena"},
		{Modelo: "arco", Titulo: "Hola", QR: "resena", Cupon: "mal cupón"},
		{Modelo: "arco", Titulo: string(make([]rune, 41)), QR: "resena"},
	} {
		if mal.Valida() == "" {
			t.Fatalf("tendría que rechazar %+v", mal)
		}
	}
}
