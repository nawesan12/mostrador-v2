//go:build vivo

// Prueba contra la API real de Mercado Pago con la cuenta de prueba de una vendedora. No corre en el
// CI: go test -tags vivo ./internal/integraciones/mercadopago/ con MP_TOKEN_VENDEDOR_PRUEBA y
// MP_USUARIO_VENDEDOR_PRUEBA en el entorno. No cobra nada: crea una preferencia, una caja y una orden
// de QR, y la cancela.
package mercadopago

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/links"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func TestContraMercadoPagoReal(t *testing.T) {
	token := os.Getenv("MP_TOKEN_VENDEDOR_PRUEBA")
	usuario, _ := strconv.ParseInt(os.Getenv("MP_USUARIO_VENDEDOR_PRUEBA"), 10, 64)
	if token == "" || usuario == 0 {
		t.Skip("sin credenciales de prueba")
	}
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	tienda := pruebadb.CrearTienda(t, b.App, "prueba-mp")
	k := make([]byte, 32)
	_, _ = rand.Read(k)
	cif, _ := cifrado.Nuevo(base64.StdEncoding.EncodeToString(k))
	s := &Servicio{Pool: b.App, Cifra: cif, Cliente: &Cliente{},
		Config: Config{AppID: "app", Secreto: "sec", Sitio: "https://mostrador.iamnaw.dev"}}
	// La conexión como si la tienda hubiera vuelto del OAuth con estos tokens.
	if err := s.guardarTokens(ctx, tienda, tokens{AccessToken: token, RefreshToken: "sin-uso", ExpiresIn: 180 * 24 * 3600, UserID: usuario}, true); err != nil {
		t.Fatal(err)
	}

	// Link de pago → preferencia de Checkout Pro.
	cod, _ := links.Crear(ctx, b.App, tienda, links.Nuevo{Monto: 1500, Concepto: "Prueba de Tu Local Web"})
	u, err := s.PagarLink(ctx, tienda, cod)
	if err != nil {
		t.Fatalf("preferencia: %s", detalle(err))
	}
	if !strings.HasPrefix(u, "https://www.mercadopago.com.ar/checkout/") {
		t.Fatalf("init_point inesperado: %s", u)
	}
	t.Logf("checkout: %s", u)

	// QR: sucursal y caja en la cuenta de prueba, una orden dinámica, consultarla y cancelarla.
	if err := s.CrearCaja(ctx, tienda, Ubicacion{Calle: "Av. Colón", Numero: "2500", Ciudad: "Mar del Plata",
		Provincia: "Buenos Aires", Latitud: -38.0055, Longitud: -57.5426}); err != nil {
		t.Fatalf("caja: %s", detalle(err))
	}
	stock := 5
	p, err := listas.Guardar(ctx, b.App, tienda, "", listas.Editable{Nombre: "Termo", Precio: 2000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	var n pedidos.PedidoNuevo
	n.Pago, n.Canal, n.Envio = "mp", "pos", "retiro"
	f := false
	n.Pagado = &f
	n.Items = append(n.Items, pedidos.Linea{ProductoID: p.ID, Cantidad: 1})
	num, _, err := pedidos.CrearDesdeProductos(ctx, b.App, tienda, n)
	if err != nil {
		t.Fatal(err)
	}
	q, err := s.CobrarQR(ctx, tienda, num)
	if err != nil {
		t.Fatalf("orden QR: %s", detalle(err))
	}
	if q.QRData == "" {
		t.Fatalf("la orden no trajo qr_data: %+v", q)
	}
	t.Logf("orden QR: estado %s, qr de %d caracteres", q.Estado, len(q.QRData))
	time.Sleep(time.Second)
	e, err := s.EstadoQR(ctx, tienda, q.Referencia)
	if err != nil {
		t.Fatalf("consultar orden: %s", detalle(err))
	}
	t.Logf("estado al consultar: %s", e)
	if err := s.CancelarQR(ctx, tienda, q.Referencia); err != nil {
		t.Fatalf("cancelar orden: %s", detalle(err))
	}
	e, _ = s.EstadoQR(ctx, tienda, q.Referencia)
	t.Logf("estado después de cancelar: %s", e)
}

func detalle(err error) string {
	var e *ErrorMP
	if errors.As(err, &e) {
		return e.Error() + " " + e.Cuerpo
	}
	return err.Error()
}
