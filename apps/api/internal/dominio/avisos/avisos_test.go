package avisos_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/avisos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cupones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/correoargentino"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// correoFijo cotiza siempre lo mismo (Correo Argentino de verdad se prueba en su paquete).
type correoFijo struct{}

func (correoFijo) Cotizar(context.Context, string, string, int) (correoargentino.Cotizacion, error) {
	return correoargentino.Cotizacion{Domicilio: &correoargentino.Opcion{Precio: 5500, Producto: "Paq.ar Clásico"}}, nil
}

func (correoFijo) Sucursal(context.Context, string, string, string) (correoargentino.Sucursal, error) {
	return correoargentino.Sucursal{}, correoargentino.ErrInvalido{Motivo: "sin sucursales"}
}

func TestMailsAQuienCompra(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	mails := &correo.Memoria{}
	av := &avisos.Servicio{Pool: b.App, Correo: mails, Sitio: "https://tulocalweb.com.ar"}
	w := &outbox.Worker{Pool: b.Worker, Manejadores: av.Manejadores()}
	stock := 5
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	compra := vidriera.Compra{Email: "lucia@ejemplo.com", Nombre: "Lucía Pérez", Telefono: "11 5555-0101", CP: "1414",
		Direccion: "Gorriti 4500", Envio: vidriera.CorreoDomicilio, Provincia: "C", Localidad: "CABA", Pago: "transferencia"}
	compra.Carrito = append(compra.Carrito, vidriera.LineaCarrito{ID: p.ID, Cantidad: 1})

	// Sin datos de transferencia, la tienda no la ofrece.
	// La tienda despacha por Correo Argentino (conectado).
	if err := b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO ca_conexiones (tienda_id, cliente, email, origen, paquete)
			VALUES ($1, '0090000025', 'luna@ejemplo.com', '{"cp":"1414"}', '{"peso":500,"alto":10,"ancho":10,"largo":10}')`, luna)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := vidriera.ComprarConCorreo(ctx, b.App, nil, correoFijo{}, "", "luna", compra, time.Now()); !errors.Is(err, vidriera.ErrInvalido) {
		t.Fatalf("transferencia sin alias: %v", err)
	}
	// CBU con el dígito verificador mal: no se guarda.
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", CBU: "2850590940090418135209"}); err == nil {
		t.Fatal("aceptó un CBU inválido")
	}
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Alias: "luna.mates", CBU: "2850590940090418135201", Titular: "Luna SRL", Email: "hola@lunamates.com.ar"}); err != nil {
		t.Fatal(err)
	}
	// El 15 % por transferencia lo prende la tienda (arranca apagado).
	if err := cupones.FijarAutomatica(ctx, b.App, luna, "transf", true); err != nil {
		t.Fatal(err)
	}
	r, err := vidriera.ComprarConCorreo(ctx, b.App, nil, correoFijo{}, "", "luna", compra, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	// La confirmación: con los datos reales para transferir, el detalle y el link al pedido.
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	m, ok := mails.Ultimo("lucia@ejemplo.com")
	if !ok || !strings.Contains(m.Asunto, "Tu pedido #") {
		t.Fatalf("confirmación: %+v", m)
	}
	// Quien compra responde (con el comprobante) a la tienda, no a Tu Local Web.
	if m.ResponderA != "hola@lunamates.com.ar" {
		t.Errorf("responder a %q", m.ResponderA)
	}
	for _, want := range []string{"luna.mates", "2850590940090418135201", "Luna SRL", "1 × Mate", "Descuento por transferencia — −$ 1.500", "/t/luna/pedido/" + r.Codigo} {
		if !strings.Contains(m.Texto, want) {
			t.Errorf("falta %q en el mail", want)
		}
	}
	// Al menos una vez del outbox, una sola vez el mail: reencolar el mismo aviso no manda otro.
	antes := len(mails.Enviados)
	if err := b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return outbox.Publicar(ctx, tx, luna, "comprador.aviso", map[string]any{"numero": r.Numero, "aviso": "confirmacion"})
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	if len(mails.Enviados) != antes {
		t.Fatal("se mandó la confirmación dos veces")
	}

	// Despachado con seguimiento: "en camino" con el número.
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE pedidos SET seguimiento = 'AND-123456' WHERE numero = $1`, r.Numero)
		return err
	})
	for _, e := range []string{"Para preparar", "Listo para despachar", "En camino"} {
		if err := pedidos.CambiarEstado(ctx, b.App, luna, r.Numero, e); err != nil {
			t.Fatalf("%s: %v", e, err)
		}
	}
	_, _ = w.Pasada(ctx)
	m, _ = mails.Ultimo("lucia@ejemplo.com")
	if !strings.Contains(m.Asunto, "en camino") || !strings.Contains(m.Texto, "AND-123456") {
		t.Fatalf("en camino: %+v", m)
	}

	// Retiro en el local: Listo para retirar → Retirado, sin "En camino" ni su mail.
	compra.Envio, compra.Email = "retiro", "tomas@ejemplo.com"
	// Se retira donde dice Envíos y retiro (no en la dirección del negocio, que acá ni está cargada).
	if err := configuracion.GuardarEntregas(ctx, b.App, luna, configuracion.Entregas{Retiro: true,
		RetiroDireccion: "Gorriti 4500, Palermo", RetiroHorario: "Lunes a viernes de 10 a 18"}); err != nil {
		t.Fatal(err)
	}
	r2, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", compra, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range []string{"Para preparar", "Listo para despachar"} {
		if err := pedidos.CambiarEstado(ctx, b.App, luna, r2.Numero, e); err != nil {
			t.Fatalf("%s: %v", e, err)
		}
	}
	if err := pedidos.CambiarEstado(ctx, b.App, luna, r2.Numero, "En camino"); !errors.Is(err, pedidos.ErrTransicion) {
		t.Fatalf("retiro en camino: %v", err)
	}
	if err := pedidos.CambiarEstado(ctx, b.App, luna, r2.Numero, "Entregado"); err != nil {
		t.Fatal(err)
	}
	_, _ = w.Pasada(ctx)
	for _, x := range mails.Enviados {
		if x.Para == "tomas@ejemplo.com" && strings.Contains(x.Asunto, "en camino") {
			t.Fatal("al que retira le llegó «en camino»")
		}
	}
	if m, _ := mails.Ultimo("tomas@ejemplo.com"); !strings.Contains(m.Asunto, "listo para retirar") ||
		!strings.Contains(m.Texto, "Retiralo en Gorriti 4500, Palermo (Lunes a viernes de 10 a 18).") {
		t.Fatalf("retiro: %+v", m)
	}
	d, err := pedidos.Obtener(ctx, b.App, luna, r2.Numero)
	if err != nil || d.Nota != nil && strings.Contains(*d.Nota, "()") {
		t.Fatalf("nota: %v %v", d.Nota, err)
	}
	if len(d.Descuentos) != 1 || d.Subtotal+d.CostoEnvio-d.Descuentos[0].Monto != d.Total {
		t.Fatalf("descuentos: %+v", d.Descuentos)
	}
}
