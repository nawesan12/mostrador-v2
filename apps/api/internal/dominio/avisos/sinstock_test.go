package avisos_test

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/avisos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// sinStockPara: a quién le llegó «Te quedaste sin stock».
func sinStockPara(m *correo.Memoria) []string {
	var out []string
	for _, x := range m.Enviados {
		if strings.HasPrefix(x.Asunto, "Te quedaste sin stock") {
			out = append(out, x.Para)
		}
	}
	slices.Sort(out)
	return out
}

func TestSinStockPorMail(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	mails := &correo.Memoria{}
	av := &avisos.Servicio{Pool: b.App, Correo: mails, Sitio: "https://tulocalweb.com.ar"}
	w := &outbox.Worker{Pool: b.Worker, Manejadores: av.Manejadores()}
	pasada := func() {
		t.Helper()
		if _, err := w.Pasada(ctx); err != nil {
			t.Fatal(err)
		}
	}
	admin := func(sql string, args ...any) {
		t.Helper()
		conn, err := pgx.Connect(ctx, b.AdminURL)
		if err != nil {
			t.Fatal(err)
		}
		defer conn.Close(ctx)
		if _, err := conn.Exec(ctx, sql, args...); err != nil {
			t.Fatal(err)
		}
	}
	// Las cuatro personas del equipo: Local no tiene Productos.
	if err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
		for _, rol := range []string{"duenio", "admin", "ventas", "local"} {
			var id string
			if err := tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre) VALUES ($1, $2) RETURNING id::text`, rol+"@luna.com", rol).Scan(&id); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `INSERT INTO auth.membresias (usuario_id, tienda_id, rol) VALUES ($1, $2, $3)`, id, luna, rol); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	// La dueña recibe en el email del negocio, como «Nueva venta».
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Alias: "luna.mates", Email: "hola@luna.com"}); err != nil {
		t.Fatal(err)
	}
	if err := configuracion.GuardarEntregas(ctx, b.App, luna, configuracion.Entregas{Retiro: true, RetiroDireccion: "Gorriti 4500"}); err != nil {
		t.Fatal(err)
	}
	nuevo := func(nombre string, stock int) listas.Producto {
		t.Helper()
		p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: nombre, Precio: 10000, Stock: &stock})
		if err != nil {
			t.Fatal(err)
		}
		return p
	}
	// Se vende la última por la tienda online.
	mate := nuevo("Mate imperial", 1)
	c := vidriera.Compra{Email: "lucia@ejemplo.com", Nombre: "Lucía Pérez", Telefono: "11 5555-0101", CP: "1414", Envio: "retiro", Pago: "transferencia"}
	c.Carrito = append(c.Carrito, vidriera.LineaCarrito{ID: mate.ID, Cantidad: 1})
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c, time.Now()); err != nil {
		t.Fatal(err)
	}
	pasada()
	if got := sinStockPara(mails); !slices.Equal(got, []string{"admin@luna.com", "hola@luna.com", "ventas@luna.com"}) {
		t.Fatalf("destinatarios: %v", got)
	}
	var m correo.Mensaje
	for _, x := range mails.Enviados {
		if x.Para == "hola@luna.com" && strings.HasPrefix(x.Asunto, "Te quedaste sin stock") {
			m = x
		}
	}
	if m.Asunto != "Te quedaste sin stock de Mate imperial" || !strings.Contains(m.HTML, "/panel/productos?editar="+mate.ID) ||
		!strings.Contains(m.HTML, "Reponer stock") || !strings.Contains(m.Texto, "«Sin stock» por email") || strings.Contains(m.HTML, "Lucía") {
		t.Fatalf("mail: %q\n%s", m.Asunto, m.Texto)
	}

	// Si repusieron antes de que salga, no sale.
	antes := len(sinStockPara(mails))
	bombilla := nuevo("Bombilla", 1)
	if _, _, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, pedidos.PedidoNuevo{Pago: "efectivo", Canal: "pos",
		Items: []pedidos.Linea{{ProductoID: bombilla.ID, Cantidad: 1}}}); err != nil {
		t.Fatal(err)
	}
	admin(`UPDATE productos SET stock = 5 WHERE id = $1`, bombilla.ID)
	pasada()
	if n := len(sinStockPara(mails)); n != antes {
		t.Fatalf("con stock repuesto salieron %d", n-antes)
	}

	// Con Sin stock › Email apagado, no sale (el push tiene su propio interruptor).
	admin(`UPDATE tiendas SET preferencias = jsonb_set(coalesce(preferencias, '{}'), '{canales}', '{"matriz":{"sinstock:email":false}}') WHERE id = $1`, luna)
	yerba := nuevo("Yerba", 1)
	if _, _, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, pedidos.PedidoNuevo{Pago: "efectivo", Canal: "pos",
		Items: []pedidos.Linea{{ProductoID: yerba.ID, Cantidad: 1}}}); err != nil {
		t.Fatal(err)
	}
	pasada()
	if n := len(sinStockPara(mails)); n != antes {
		t.Fatalf("con el mail apagado salieron %d", n-antes)
	}
}
