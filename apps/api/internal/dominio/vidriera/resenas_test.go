package vidriera_test

import (
	"context"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/ventas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func TestResenasYPreguntas(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	stock := 10
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	// Una compra cobrada con su código público.
	n := pedidos.PedidoNuevo{Canal: "web", Pago: "mp", Envio: "retiro", CodigoPublico: "codigo-de-la-compra-123"}
	n.Items = append(n.Items, pedidos.Linea{ProductoID: p.ID, Cantidad: 1})
	if _, _, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, n); err != nil {
		t.Fatal(err)
	}

	if err := ventas.DejarResena(ctx, b.App, luna, ventas.NuevaResena{Codigo: "codigo-de-la-compra-123", ProductoID: p.ID, Autor: "Lucía", Estrellas: 5, Texto: "Hermoso"}); err != nil {
		t.Fatal(err)
	}
	if err := ventas.DejarResena(ctx, b.App, luna, ventas.NuevaResena{Codigo: "codigo-de-la-compra-123", ProductoID: p.ID, Autor: "Lucía", Estrellas: 1}); err != ventas.ErrYaReseno {
		t.Fatalf("dos por la misma compra: %v", err)
	}
	// Sin código: se guarda, pero no queda "con compra".
	if err := ventas.DejarResena(ctx, b.App, luna, ventas.NuevaResena{Autor: "Anónimo", Estrellas: 1, Texto: "Malo"}); err != nil {
		t.Fatal(err)
	}
	if err := ventas.DejarResena(ctx, b.App, luna, ventas.NuevaResena{Autor: "X", Estrellas: 6}); err != ventas.ErrResenaInvalida {
		t.Fatalf("6 estrellas: %v", err)
	}

	rs, _, err := ventas.ResenasPanel(ctx, b.App, luna)
	if err != nil || len(rs) != 2 {
		t.Fatalf("panel: %d %v", len(rs), err)
	}
	conCompra := map[string]bool{}
	for _, r := range rs {
		conCompra[r.Autor] = r.ConCompra
	}
	if !conCompra["Lucía"] || conCompra["Anónimo"] {
		t.Fatalf("con compra: %v", conCompra)
	}

	// Preguntas: sólo sobre productos de la tienda; la respuesta se publica.
	if err := ventas.Preguntar(ctx, b.App, sol, p.ID, "Ana", "", "¿Viene curado?"); err != ventas.ErrResenaInvalida {
		t.Fatalf("producto de otra tienda: %v", err)
	}
	if err := ventas.Preguntar(ctx, b.App, luna, p.ID, "Ana", "ana@ejemplo.com", "¿Viene curado?"); err != nil {
		t.Fatal(err)
	}
	_, ps, _ := ventas.ResenasPanel(ctx, b.App, luna)
	if len(ps) != 1 || ps[0].Respuesta != nil {
		t.Fatalf("preguntas: %+v", ps)
	}
	if err := ventas.ResponderPregunta(ctx, b.App, luna, ps[0].ID, "Sí, listo para usar."); err != nil {
		t.Fatal(err)
	}
	if err := ventas.ResponderPregunta(ctx, b.App, sol, ps[0].ID, "Ajena"); err == nil {
		t.Fatal("otra tienda respondió")
	}

	// Ocultar la de una estrella: sale de la tienda y del promedio.
	var anon string
	for _, r := range rs {
		if r.Autor == "Anónimo" {
			anon = r.ID
		}
	}
	if err := ventas.AccionResena(ctx, b.App, luna, anon, "ocultar", ""); err != nil {
		t.Fatal(err)
	}
	tienda, err := vidriera.Publica(ctx, b.App, "luna")
	if err != nil {
		t.Fatal(err)
	}
	if len(tienda.Resenas) != 1 || !tienda.Resenas[0].ConCompra || tienda.Promedio == nil || *tienda.Promedio != 5 {
		t.Fatalf("pública: %+v %v", tienda.Resenas, tienda.Promedio)
	}
	if len(tienda.Preguntas) != 1 || tienda.Preguntas[0].Respuesta != "Sí, listo para usar." {
		t.Fatalf("preguntas públicas: %+v", tienda.Preguntas)
	}
	if otra, _ := vidriera.Publica(ctx, b.App, "sol"); len(otra.Resenas) != 0 || len(otra.Preguntas) != 0 {
		t.Fatal("otra tienda ve reseñas ajenas")
	}
}
