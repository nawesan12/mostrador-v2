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
