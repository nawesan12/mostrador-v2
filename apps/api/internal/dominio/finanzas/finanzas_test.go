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
	n.Items = append(n.Items, pedidos.Linea{ProductoID: p.ID, Cantidad: 2})
	if _, _, err := pedidos.CrearDesdeProductos(ctx, b.App, id, n); err != nil {
		t.Fatal(err)
	}
	f, err := finanzas.Leer(ctx, b.App, id, time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	k := f["kpis"].(map[string]any)
	if *k["margen"].(*int) != 40 || *k["ganancia"].(*int64) != 32000 || f["conciliado"] != false {
		t.Fatalf("kpis: %+v", k)
	}
	// Un producto vendido sin costo: la ganancia no se estima (ni el margen da 100 %).
	sin, _ := listas.Guardar(ctx, b.App, id, "", listas.Editable{Nombre: "Bombilla", Precio: 5000, Stock: &stock})
	n.Items = []pedidos.Linea{{ProductoID: sin.ID, Cantidad: 1}}
	if _, _, err := pedidos.CrearDesdeProductos(ctx, b.App, id, n); err != nil {
		t.Fatal(err)
	}
	f, _ = finanzas.Leer(ctx, b.App, id, time.Now().Add(time.Minute))
	k = f["kpis"].(map[string]any)
	if k["ganancia"].(*int64) != nil || k["margen"].(*int) != nil || k["gananciaDetalle"] != "Falta el costo de 1 producto vendido" {
		t.Fatalf("sin costo: %+v", k)
	}
	if m := f["movimientos"].([]finanzas.Movimiento); len(m) != 2 || m[1].Monto != 80000 {
		t.Fatalf("movimientos: %+v", m)
	}
	if f["caja"] != nil || f["disponible"] != nil || k["aLiquidar"] != nil || k["gastos"] != nil {
		t.Fatalf("inventó plata sin Mercado Pago: %+v", f)
	}
}
