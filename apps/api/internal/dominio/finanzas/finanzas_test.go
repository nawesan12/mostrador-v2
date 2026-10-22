package finanzas_test

import (
	"context"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/finanzas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func TestFinanzasConLoQueSeSabe(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	id := pruebadb.CrearTienda(t, b.App, "luna")
	stock, costo := 5, int64(24000)
	p, _ := listas.Guardar(ctx, b.App, id, "", listas.Editable{Nombre: "Termo", Precio: 40000, Costo: &costo, Stock: &stock})
	var n pedidos.PedidoNuevo
	n.Pago, n.Canal, n.Envio = "transf", "web", "retiro"
	n.Items = append(n.Items, struct {
		ProductoID string `json:"productoId"`
		Cantidad   int    `json:"cantidad"`
	}{p.ID, 2})
	if _, _, err := pedidos.CrearDesdeProductos(ctx, b.App, id, n); err != nil {
		t.Fatal(err)
	}
	f, err := finanzas.Leer(ctx, b.App, id, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	k := f["kpis"].(map[string]any)
	if *k["margen"].(*int) != 40 || k["ganancia"].(int64) != 32000 {
		t.Fatalf("kpis: %+v", k)
	}
	if m := f["movimientos"].([]finanzas.Movimiento); len(m) != 1 || m[0].Monto != 80000 {
		t.Fatalf("movimientos: %+v", m)
	}
	if f["caja"] != nil || f["disponible"] != 0 {
		t.Fatalf("inventó plata sin Mercado Pago: %+v", f)
	}
}
