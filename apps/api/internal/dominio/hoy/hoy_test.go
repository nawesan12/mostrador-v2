package hoy_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/hoy"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

var ctx = context.Background()

func TestTareasYActividad(t *testing.T) {
	b := pruebadb.Nueva(t)
	id := pruebadb.CrearTienda(t, b.App, "luna")
	stock := 6
	p, _ := listas.Guardar(ctx, b.App, id, "", listas.Editable{Nombre: "Termo", Precio: 40000, Stock: &stock})
	for _, pago := range []string{"transf", "mp"} {
		var n pedidos.PedidoNuevo
		n.Pago, n.Canal, n.Envio = pago, "web", "andreani"
		f := false
		if pago == "transf" {
			n.Pagado = &f
		}
		n.Items = append(n.Items, struct {
			ProductoID string `json:"productoId"`
			Cantidad   int    `json:"cantidad"`
		}{p.ID, 1})
		if _, _, err := pedidos.CrearDesdeProductos(ctx, b.App, id, n); err != nil {
			t.Fatal(err)
		}
	}
	ahora := time.Now()
	g, err := hoy.ParaHacer(ctx, b.App, id, ahora)
	if err != nil {
		t.Fatal(err)
	}
	ids := map[string]hoy.Tarea{}
	for _, x := range g {
		for _, y := range x.Tareas {
			ids[y.ID] = y
		}
	}
	if ids["preparar"].Valor == nil || *ids["preparar"].Valor != 40000 || ids["transferencias"].Titulo != "Conciliar 1 transferencia" {
		t.Fatalf("tareas: %+v", ids)
	}
	if ids["reponer"].Titulo != "Reponer 1 producto" {
		t.Fatalf("stock bajo (quedan 4): %+v", ids["reponer"])
	}
	if err := hoy.Marcar(ctx, b.App, id, "preparar", true, ahora); err != nil {
		t.Fatal(err)
	}
	g, _ = hoy.ParaHacer(ctx, b.App, id, ahora)
	if !g[0].Tareas[0].Hecha {
		t.Fatal("no quedó marcada")
	}
	// Al otro día, arranca sin marcar.
	g, _ = hoy.ParaHacer(ctx, b.App, id, ahora.Add(24*time.Hour))
	if g[0].Tareas[0].Hecha {
		t.Fatal("la marca pasó al día siguiente")
	}

	_ = b.App.EnTienda(ctx, id, func(tx pgx.Tx) error {
		_, err := auditoria.Anotar(ctx, tx, id, auditoria.Actor{Tipo: "persona", ID: "x"}, auditoria.Registro{Accion: "pausó", Entidad: "el cupón", EntidadID: "FINDE10"})
		return err
	})
	a, err := hoy.Actividad(ctx, b.App, id)
	if err != nil || len(a) != 1 || a[0].Texto != "pausó el cupón FINDE10" || a[0].Quien != "Vos" {
		t.Fatalf("actividad: %+v %v", a, err)
	}
}
