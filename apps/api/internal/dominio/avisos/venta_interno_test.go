package avisos

import (
	"strings"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
)

func TestMailVentaArmado(t *testing.T) {
	email, tel := "lucia@ejemplo.com", "11 5555-0101"
	v := venta{numero: 1042, tienda: "Luna <Mates>", para: "hola@luna.com", pago: "mp", estadoPago: "Aprobado", envio: "correo",
		comprador: pedidos.Comprador{Nombre: "Lucía Pérez", Email: &email, Telefono: &tel},
		nota:      "DNI 30111222 · Entrega: Gorriti 4500 (CP 1414)",
		entrega:   pedidos.Entrega{Direccion: "Gorriti 4500", Localidad: "CABA", CP: "1414", Tipo: "D"},
		items:     []item{{Nombre: "Mate", Cantidad: 2, Precio: 10000, Variante: "Negro"}},
		subtotal:  20000, envioCosto: 5500, total: 23500,
		descuentos: []pedidos.Descuento{{Concepto: "Cupón HOLA", Monto: 2000}}}
	m := mailVenta("https://tulocalweb.com.ar/", v)
	if m.Para != "hola@luna.com" || m.Asunto != "Nueva venta #1042 · $ 23.500" || m.ResponderA != email {
		t.Fatalf("encabezados: %+v", m)
	}
	for _, want := range []string{"2 × Mate (Negro) — $ 20.000", "Subtotal — $ 20.000", "Envío — $ 5.500",
		"Cupón HOLA — −$ 2.000", "Total — $ 23.500", "Pago: Mercado Pago · acreditado",
		"Entrega: Correo Argentino a domicilio: Gorriti 4500, CABA (CP 1414)", "Teléfono: 11 5555-0101", "Nota: DNI 30111222\n",
		"Ver el pedido: https://tulocalweb.com.ar/panel/pedidos/1042", "https://tulocalweb.com.ar/panel/notificaciones"} {
		if !strings.Contains(m.Texto, want) {
			t.Errorf("falta %q en:\n%s", want, m.Texto)
		}
	}
	if strings.Contains(m.HTML, "<Mates>") || !strings.Contains(m.HTML, "Luna &lt;Mates&gt;") {
		t.Error("el HTML no escapa el nombre de la tienda")
	}

	// Transferencia pendiente y retiro en sucursal; sin teléfono no hay fila vacía.
	v.pago, v.estadoPago, v.comprador.Telefono, v.nota = "transf", "Pendiente", nil, ""
	v.entrega = pedidos.Entrega{Tipo: "S", SucursalNombre: "Palermo"}
	m = mailVenta("https://tulocalweb.com.ar", v)
	for _, want := range []string{"Transferencia · pendiente de comprobante", "marcá el pedido como pagado",
		"Correo Argentino, retira en la sucursal Palermo"} {
		if !strings.Contains(m.Texto, want) {
			t.Errorf("falta %q", want)
		}
	}
	if strings.Contains(m.Texto, "Teléfono") || strings.Contains(m.Texto, "Nota:") {
		t.Error("filas vacías")
	}
	// Un email raro de quien compra no va al Reply-To.
	raro := "a@b.com, c@d.com"
	v.comprador.Email = &raro
	if mailVenta("", v).ResponderA != "" {
		t.Error("Reply-To con dos direcciones")
	}
	for envio, want := range map[string]string{"retiro": "Retira en el local", "moto": "Envío en moto"} {
		v.envio, v.entrega = envio, pedidos.Entrega{}
		if got := entregaDe(v); got != want {
			t.Errorf("%s: %q", envio, got)
		}
	}
}
