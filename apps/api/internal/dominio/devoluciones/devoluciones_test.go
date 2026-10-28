package devoluciones_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/devoluciones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func TestArrepentimientoDeLaTienda(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	mails := &correo.Memoria{}
	s := &devoluciones.Servicio{Pool: b.App, Correo: mails}
	stock := 5
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	n := pedidos.PedidoNuevo{Canal: "web", Pago: "mp", Envio: "retiro", Email: "lucia@ejemplo.com", CodigoPublico: "codigo-publico-del-pedido"}
	n.Items = append(n.Items, struct {
		ProductoID string `json:"productoId"`
		Cantidad   int    `json:"cantidad"`
	}{p.ID, 1})
	numero, _, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, n)
	if err != nil {
		t.Fatal(err)
	}
	ahora := time.Now()

	codigo, err := s.Arrepentirse(ctx, luna, "Luna", devoluciones.Pedido{Codigo: "codigo-publico-del-pedido", Nombre: "Lucía", Email: "lucia@ejemplo.com"}, ahora)
	if err != nil || !strings.HasPrefix(codigo, "ARR-") {
		t.Fatalf("arrepentirse: %q %v", codigo, err)
	}
	if m, ok := mails.Ultimo("lucia@ejemplo.com"); !ok || !strings.Contains(m.Texto, codigo) {
		t.Fatal("no salió el mail con el código")
	}
	// Por número: con el email de otra persona no se encuentra; con el de la compra, el mismo trámite.
	if _, err := s.Arrepentirse(ctx, luna, "Luna", devoluciones.Pedido{Numero: numero, Nombre: "Otra", Email: "otra@ejemplo.com"}, ahora); err != devoluciones.ErrNoEncontrado {
		t.Fatalf("email ajeno: %v", err)
	}
	if c2, err := s.Arrepentirse(ctx, luna, "Luna", devoluciones.Pedido{Numero: numero, Nombre: "Lucía", Email: "LUCIA@ejemplo.com"}, ahora); err != nil || c2 != codigo {
		t.Fatalf("mismo trámite: %q %v", c2, err)
	}
	// Repetir no manda otro mail.
	if len(mails.Enviados) != 1 {
		t.Fatalf("mails: %d", len(mails.Enviados))
	}
	// Desde otra tienda, ese código no existe.
	if _, err := s.Arrepentirse(ctx, sol, "Sol", devoluciones.Pedido{Codigo: "codigo-publico-del-pedido", Nombre: "Lucía", Email: "lucia@ejemplo.com"}, ahora); err != devoluciones.ErrNoEncontrado {
		t.Fatalf("otra tienda: %v", err)
	}
	// Pasados 10 días de la compra (sin entrega), ya no.
	if _, err := s.Arrepentirse(ctx, luna, "Luna", devoluciones.Pedido{Codigo: "codigo-publico-del-pedido", Nombre: "Lucía", Email: "lucia@ejemplo.com"}, ahora.Add(11*24*time.Hour)); err != devoluciones.ErrFueraDePlazo {
		t.Fatalf("fuera de plazo: %v", err)
	}

	if ds, _, _ := s.Panel(ctx, sol); len(ds) != 0 {
		t.Fatalf("sol ve %d devoluciones de luna", len(ds))
	}
	ds, tasa, err := s.Panel(ctx, luna)
	if err != nil || len(ds) != 1 || ds[0].Producto != "Mate" || ds[0].Monto != 10000 || ds[0].Estado != "Pendiente" || tasa != 100 {
		t.Fatalf("panel: %+v %v %v", ds, tasa, err)
	}
	if err := s.Resolver(ctx, luna, ds[0].ID, "aprobar"); err != nil {
		t.Fatal(err)
	}
	if err := s.Resolver(ctx, luna, ds[0].ID, "deshacer"); err != nil {
		t.Fatal(err)
	}
	if ds, _, _ := s.Panel(ctx, luna); ds[0].Estado != "Pendiente" {
		t.Fatalf("deshacer: %s", ds[0].Estado)
	}
	if err := s.Resolver(ctx, sol, ds[0].ID, "aprobar"); err != devoluciones.ErrNoEncontrado {
		t.Fatalf("otra tienda resolvió: %v", err)
	}
}

// Dos pedidos a la vez por la misma compra: los dos reciben el mismo código, sin error.
func TestArrepentimientoConcurrente(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	s := &devoluciones.Servicio{Pool: b.App}
	stock := 5
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	n := pedidos.PedidoNuevo{Canal: "web", Pago: "mp", Envio: "retiro", Email: "lucia@ejemplo.com", CodigoPublico: "codigo-concurrente"}
	n.Items = append(n.Items, struct {
		ProductoID string `json:"productoId"`
		Cantidad   int    `json:"cantidad"`
	}{p.ID, 1})
	if _, _, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, n); err != nil {
		t.Fatal(err)
	}
	type res struct {
		c   string
		err error
	}
	ch := make(chan res, 4)
	for range 4 {
		go func() {
			c, err := s.Arrepentirse(ctx, luna, "Luna", devoluciones.Pedido{Codigo: "codigo-concurrente", Nombre: "Lucía", Email: "lucia@ejemplo.com"}, time.Now())
			ch <- res{c, err}
		}()
	}
	primero := ""
	for range 4 {
		r := <-ch
		if r.err != nil {
			t.Fatal(r.err)
		}
		if primero == "" {
			primero = r.c
		} else if r.c != primero {
			t.Fatalf("códigos distintos: %s %s", primero, r.c)
		}
	}
}
