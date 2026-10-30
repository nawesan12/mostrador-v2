package avisos_test

import (
	"context"
	"strconv"
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

// ventas cuenta los «Nueva venta» que le llegaron a la dueña.
func ventas(m *correo.Memoria, para string) []correo.Mensaje {
	var out []correo.Mensaje
	for _, x := range m.Enviados {
		if x.Para == para && strings.HasPrefix(x.Asunto, "Nueva venta #") {
			out = append(out, x)
		}
	}
	return out
}

func TestVentaNuevaALaDuenia(t *testing.T) {
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
	// Sin email en Negocio: va a la cuenta dueña (no a la de ventas).
	if err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
		var duenia, otra string
		if err := tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre) VALUES ('ana@luna.com', 'Ana') RETURNING id::text`).Scan(&duenia); err != nil {
			return err
		}
		if err := tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre) VALUES ('beto@luna.com', 'Beto') RETURNING id::text`).Scan(&otra); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO auth.membresias (usuario_id, tienda_id, rol) VALUES ($2, $3, 'ventas'), ($1, $3, 'duenio')`, duenia, otra, luna)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	stock := 20
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Alias: "luna.mates"}); err != nil {
		t.Fatal(err)
	}
	if err := configuracion.GuardarEntregas(ctx, b.App, luna, configuracion.Entregas{Retiro: true, RetiroDireccion: "Gorriti 4500"}); err != nil {
		t.Fatal(err)
	}
	compra := vidriera.Compra{Email: "lucia@ejemplo.com", Nombre: "Lucía Pérez", Telefono: "11 5555-0101", CP: "1414",
		Envio: "retiro", Pago: "transferencia"}
	compra.Carrito = append(compra.Carrito, vidriera.LineaCarrito{ID: p.ID, Cantidad: 2})

	// Transferencia: sale al crearse, aclarando que falta el comprobante.
	r, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", compra, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	pasada()
	v := ventas(mails, "ana@luna.com")
	if len(v) != 1 || len(ventas(mails, "beto@luna.com")) != 0 {
		t.Fatalf("transferencia: %d mails a la dueña", len(v))
	}
	m := v[0]
	if m.Asunto != "Nueva venta #"+strconv.Itoa(r.Numero)+" · $ 20.000" {
		t.Errorf("asunto %q", m.Asunto)
	}
	for _, want := range []string{"2 × Mate — $ 20.000", "Total — $ 20.000", "Transferencia · pendiente de comprobante",
		"todavía no está acreditada", "Retira en el local", "Lucía Pérez", "lucia@ejemplo.com", "11 5555-0101",
		"https://tulocalweb.com.ar/panel/pedidos/" + strconv.Itoa(r.Numero)} {
		if !strings.Contains(m.Texto, want) {
			t.Errorf("falta %q en el texto:\n%s", want, m.Texto)
		}
	}
	if !strings.Contains(m.HTML, "Ver el pedido") || m.ResponderA != "lucia@ejemplo.com" {
		t.Errorf("html o responder a: %q", m.ResponderA)
	}
	// El outbox entrega al menos una vez: reprocesar no manda otro. Marcarlo pagado tampoco.
	if err := b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return outbox.Publicar(ctx, tx, luna, "tienda.venta", map[string]any{"numero": r.Numero})
	}); err != nil {
		t.Fatal(err)
	}
	if err := pedidos.CambiarEstado(ctx, b.App, luna, r.Numero, "Para preparar"); err != nil {
		t.Fatal(err)
	}
	pasada()
	if n := len(ventas(mails, "ana@luna.com")); n != 1 {
		t.Fatalf("se mandó %d veces", n)
	}

	// Mercado Pago: nada al crearse; sale cuando se acredita, una sola vez aunque vuelva a acreditarse.
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Alias: "luna.mates", Email: "hola@lunamates.com.ar"}); err != nil {
		t.Fatal(err)
	}
	cliente, f := "Tomás Gil", false
	mp, _, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, pedidos.PedidoNuevo{Canal: "web", Pago: "mp", Envio: "retiro",
		Cliente: &cliente, Email: "tomas@ejemplo.com", Pagado: &f, Items: []pedidos.Linea{{ProductoID: p.ID, Cantidad: 1}}})
	if err != nil {
		t.Fatal(err)
	}
	pasada()
	if n := len(ventas(mails, "hola@lunamates.com.ar")); n != 0 {
		t.Fatalf("Mercado Pago sin acreditar: %d mails", n)
	}
	estadoPago := func(e string) {
		t.Helper()
		if err := b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `UPDATE pedidos SET estado_pago = $2 WHERE numero = $1`, mp, e)
			return err
		}); err != nil {
			t.Fatal(err)
		}
	}
	estadoPago("Aprobado")
	pasada()
	v = ventas(mails, "hola@lunamates.com.ar")
	if len(v) != 1 || !strings.Contains(v[0].Texto, "Mercado Pago · acreditado") || strings.Contains(v[0].Texto, "comprobante") {
		t.Fatalf("Mercado Pago acreditado: %+v", v)
	}
	estadoPago("Devuelto")
	estadoPago("Aprobado")
	pasada()
	if n := len(ventas(mails, "hola@lunamates.com.ar")); n != 1 {
		t.Fatalf("reacreditado: %d mails", n)
	}

	// Ni el punto de venta ni lo que carga la tienda a mano le avisan a ella misma.
	efectivo := "efectivo"
	if _, _, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, pedidos.PedidoNuevo{Canal: "pos", Pago: efectivo, Envio: "retiro",
		Items: []pedidos.Linea{{ProductoID: p.ID, Cantidad: 1}}}); err != nil {
		t.Fatal(err)
	}
	pasada()
	if n := len(ventas(mails, "hola@lunamates.com.ar")); n != 1 {
		t.Fatalf("punto de venta: %d mails", n)
	}

	// Apagado en Notificaciones › Venta nueva › Email: no sale.
	if err := b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE tiendas SET preferencias = jsonb_set(preferencias, '{canales}',
			'{"matriz": {"venta:email": false}, "silencio": false, "resumen": "off"}') WHERE id = $1`, luna)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", compra, time.Now()); err != nil {
		t.Fatal(err)
	}
	pasada()
	if n := len(ventas(mails, "hola@lunamates.com.ar")); n != 1 {
		t.Fatalf("apagado: %d mails", n)
	}
}
