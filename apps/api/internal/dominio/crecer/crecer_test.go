package crecer_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/crecer"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

var ctx = context.Background()

func TestCanalesEInformesSalenDeLosPedidos(t *testing.T) {
	b := pruebadb.Nueva(t)
	id := pruebadb.CrearTienda(t, b.App, "luna")
	stock := 10
	p, _ := listas.Guardar(ctx, b.App, id, "", listas.Editable{Nombre: "Termo", Precio: 40000, Stock: &stock})
	for _, canal := range []string{"pos", "pos", "web"} {
		var n pedidos.PedidoNuevo
		n.Pago, n.Canal = "efectivo", canal
		tel := "1155555555"
		n.Telefono = &tel
		n.Items = append(n.Items, pedidos.Linea{ProductoID: p.ID, Cantidad: 1})
		if _, _, err := pedidos.CrearDesdeProductos(ctx, b.App, id, n); err != nil {
			t.Fatal(err)
		}
	}
	ahora := time.Now().Add(time.Minute)
	c, err := crecer.Canales(ctx, b.App, id, ahora)
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range c["canales"].([]crecer.Canal) {
		if x.ID == "pos" && (x.Vendido != 80000 || x.Pedidos != 2 || !x.Activo) {
			t.Fatalf("local: %+v", x)
		}
		if x.ID == "ml" && x.Conectado {
			t.Fatalf("Mercado Libre figura conectado sin integración")
		}
	}
	if err := crecer.FijarCanal(ctx, b.App, id, "ml", false); !errors.Is(err, crecer.ErrSinConectar) {
		t.Fatalf("pausó un canal sin conectar: %v", err)
	}
	if err := crecer.FijarCanal(ctx, b.App, id, "pos", false); err != nil {
		t.Fatal(err)
	}
	c, _ = crecer.Canales(ctx, b.App, id, ahora)
	for _, x := range c["canales"].([]crecer.Canal) {
		if x.ID == "pos" && x.Activo {
			t.Fatal("no quedó pausado")
		}
	}
	inf, err := crecer.Informes(ctx, b.App, id, ahora)
	if err != nil {
		t.Fatal(err)
	}
	s := inf["series"].(map[string]any)
	if v := s["ventas"].([]int64); v[11] != 120000 || v[10] != 0 {
		t.Fatalf("ventas: %v", v)
	}
	if u := s["productos"].([]int64); u[11] != 3 {
		t.Fatalf("unidades: %v", u)
	}
	if inf["historia"] != 1 {
		t.Fatalf("historia: %v", inf["historia"])
	}
	// El 31 de octubre a la noche (ya 1 de noviembre en UTC): doce meses calendario sin repetir.
	fin := time.Date(2026, 11, 1, 2, 30, 0, 0, time.UTC)
	inf, _ = crecer.Informes(ctx, b.App, id, fin)
	want := []string{"nov", "dic", "ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sept", "oct"}
	if got := inf["meses"].([]string); strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("meses: %v", got)
	}
	if k := inf["kpis"].(map[string]any); k["mesAnterior"] != "septiembre" {
		t.Fatalf("mes anterior: %v", k["mesAnterior"])
	}
	m, err := crecer.Marketing(ctx, b.App, id)
	if err != nil || m["audiencias"].([]map[string]any)[0]["cantidad"] != 1 {
		t.Fatalf("marketing: %+v %v", m["audiencias"], err)
	}
}
