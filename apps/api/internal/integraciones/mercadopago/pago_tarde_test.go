package mercadopago_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cupones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadopago"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// Un pago que llega con el pedido ya cancelado no se factura y avisa «pago.tarde»; un segundo pago
// aprobado no pisa al primero y avisa «pago.doble»; «Pagar» dos veces reusa la misma preferencia.
func TestPagoTardeYDoble(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	var ana string
	if err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre) VALUES ('ana@luna.com', 'Ana') RETURNING id::text`).Scan(&ana)
	}); err != nil {
		t.Fatal(err)
	}
	f := &mpFalso{cuerpos: map[string]map[string]any{}, pagos: map[string]map[string]any{}, susc: map[string]map[string]any{}, usuario: 779, refreshOK: true}
	srv := httptest.NewServer(f)
	defer srv.Close()
	k := make([]byte, 32)
	_, _ = rand.Read(k)
	cif, _ := cifrado.Nuevo(base64.StdEncoding.EncodeToString(k))
	ahora := time.Now()
	s := &mercadopago.Servicio{Pool: b.App, Cifra: cif, Cliente: &mercadopago.Cliente{Base: srv.URL},
		Ahora:  func() time.Time { return ahora },
		Config: mercadopago.Config{AppID: "123", Secreto: "sec", SecretoWebhook: "whsec", Sitio: "https://tulocalweb.com.ar"}}
	mux := http.NewServeMux()
	s.Rutas(mux)
	u, err := s.Conectar(ctx, luna, ana)
	if err != nil {
		t.Fatal(err)
	}
	q, _ := url.Parse(u)
	if err := s.Volver(ctx, luna, ana, "codigo", q.Query().Get("state")); err != nil {
		t.Fatal(err)
	}
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Alias: "luna.mates"}); err != nil {
		t.Fatal(err)
	}
	if err := configuracion.GuardarEntregas(ctx, b.App, luna, configuracion.Entregas{Retiro: true, RetiroDireccion: "Gorriti 4500"}); err != nil {
		t.Fatal(err)
	}
	stock := 5
	mate, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	admin, err := pgx.Connect(ctx, b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)
	compra := func() vidriera.Resultado {
		t.Helper()
		c := vidriera.Compra{Email: "lucia@ejemplo.com", Nombre: "Lucía Pérez", Telefono: "11 5555-0101", CP: "1414",
			Envio: "retiro", Pago: "mercadopago"}
		c.Carrito = append(c.Carrito, vidriera.LineaCarrito{ID: mate.ID, Cantidad: 1})
		r, err := vidriera.Comprar(ctx, b.App, s, s.Config.Sitio, "luna", c, ahora)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}
	contar := func(sql string, args ...any) int {
		t.Helper()
		var n int
		if err := admin.QueryRow(ctx, sql, args...).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	wk := &outbox.Worker{Pool: b.Worker, Manejadores: s.Manejadores()}
	webhook := func(id int, ref string, monto int64) {
		t.Helper()
		f.mu.Lock()
		f.pagos[strconv.Itoa(id)] = map[string]any{"id": id, "status": "approved", "external_reference": ref, "transaction_amount": float64(monto)}
		f.mu.Unlock()
		req := "req-" + strconv.Itoa(id)
		rq := httptest.NewRequest(http.MethodPost, "/v1/webhooks/mercadopago?data.id="+strconv.Itoa(id)+"&type=payment",
			strings.NewReader(`{"type":"payment","action":"payment.created","user_id":779,"data":{"id":"`+strconv.Itoa(id)+`"}}`))
		rq.Header.Set("x-request-id", req)
		rq.Header.Set("x-signature", firmar("whsec", strconv.Itoa(id), req, ahora.UnixMilli()))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, rq)
		if w.Code != http.StatusOK {
			t.Fatalf("webhook: %d", w.Code)
		}
		if _, err := wk.Pasada(ctx); err != nil {
			t.Fatal(err)
		}
	}

	// 1. Se cancela (vence a las 24 h) y después llega el pago aprobado.
	r := compra()
	ref1 := f.cuerpos["POST /checkout/preferences"]["external_reference"].(string)
	if _, err := admin.Exec(ctx, `UPDATE pedidos SET creado_en = now() - interval '25 hours' WHERE numero = $1`, r.Numero); err != nil {
		t.Fatal(err)
	}
	if _, err := admin.Exec(ctx, `UPDATE mp_cobros SET creado_en = now() - interval '25 hours' WHERE referencia = $1`, ref1); err != nil {
		t.Fatal(err)
	}
	if n, err := pedidos.VencerSinPago(ctx, b.App); err != nil || n != 1 {
		t.Fatalf("vencer: %d %v", n, err)
	}
	d, _ := pedidos.Obtener(ctx, b.App, luna, r.Numero)
	webhook(901, ref1, d.Total)
	d, _ = pedidos.Obtener(ctx, b.App, luna, r.Numero)
	if d.Estado != "Cancelado" || d.EstadoPago != "Aprobado" {
		t.Fatalf("pago tarde: %s / %s", d.Estado, d.EstadoPago)
	}
	if n := contar(`SELECT count(*) FROM comprobantes WHERE pedido_numero = $1`, r.Numero); n != 0 {
		t.Fatalf("se encoló factura de un pedido cancelado: %d", n)
	}
	if n := contar(`SELECT count(*) FROM outbox WHERE tipo = 'pago.tarde' AND (datos->>'numero')::int = $1`, r.Numero); n != 1 {
		t.Fatalf("avisos pago.tarde: %d", n)
	}
	if n := contar(`SELECT count(*) FROM outbox WHERE tipo IN ('tienda.venta', 'pago.aprobado') AND (datos->>'numero')::int = $1`, r.Numero); n != 0 {
		t.Fatalf("el pago tarde se anunció como venta: %d", n)
	}
	// Reabrirlo con el pago ya aprobado sí lo factura (comprobante pendiente o «no facturado»).
	if err := b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE pedidos SET estado = 'Para preparar' WHERE numero = $1`, r.Numero)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if n := contar(`SELECT count(*) FROM comprobantes WHERE pedido_numero = $1`, r.Numero); n != 1 {
		t.Fatalf("al reabrir no se encoló el comprobante: %d", n)
	}

	// 2. «Pagar» otra vez (volver del checkout, otra pestaña) reusa la preferencia del checkout; dos
	// pagos aprobados de la misma no se pisan.
	v := compra()
	prefs := len(f.vio("POST /checkout/preferences"))
	ref2 := f.cuerpos["POST /checkout/preferences"]["external_reference"].(string)
	u1, err := s.PagarPedido(ctx, luna, v.Numero, "https://luna.tulocalweb.com.ar/pedido/x")
	if err != nil {
		t.Fatal(err)
	}
	u2, err := s.PagarPedido(ctx, luna, v.Numero, "https://luna.tulocalweb.com.ar/pedido/x")
	if err != nil || u1 != u2 || len(f.vio("POST /checkout/preferences")) != prefs {
		t.Fatalf("pagar otra vez armó otra preferencia: %v %q %q", err, u1, u2)
	}
	dv, _ := pedidos.Obtener(ctx, b.App, luna, v.Numero)
	webhook(911, ref2, dv.Total)
	webhook(912, ref2, dv.Total)
	var pago int64
	var extra []int64
	if err := admin.QueryRow(ctx, `SELECT pago, pagos_extra FROM mp_cobros WHERE referencia = $1`, ref2).Scan(&pago, &extra); err != nil {
		t.Fatal(err)
	}
	if pago != 911 || len(extra) != 1 || extra[0] != 912 {
		t.Fatalf("el segundo pago pisó al primero: pago %d extra %v", pago, extra)
	}
	if n := contar(`SELECT count(*) FROM outbox WHERE tipo = 'pago.doble' AND (datos->>'numero')::int = $1`, v.Numero); n != 1 {
		t.Fatalf("avisos pago.doble: %d", n)
	}
	// Y la otra preferencia (la del checkout) pagada también: es un pago de más, no otra venta.
	if _, err := admin.Exec(ctx, `INSERT INTO mp_cobros (referencia, tienda_id, tipo, objetivo, monto) VALUES ('OTRA1', $1, 'pedido', $2, $3)`,
		luna, strconv.Itoa(v.Numero), dv.Total); err != nil {
		t.Fatal(err)
	}
	webhook(913, "OTRA1", dv.Total)
	if n := contar(`SELECT count(*) FROM outbox WHERE tipo = 'pago.doble' AND (datos->>'numero')::int = $1`, v.Numero); n != 2 {
		t.Fatalf("pago de otra preferencia del mismo pedido: %d avisos", n)
	}
}

// Volver de Mercado Pago y tocar Pagar otra vez no arma otro pedido; cambiar el carrito cancela el
// que esperaba (vuelven el stock y el cupón de un uso). Un cupón que cubre todo deja el pedido pago,
// sin Mercado Pago. Cancelar una venta con cupón le resta al cupón lo vendido.
func TestPagarOtraVezYTotalCero(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	var ana string
	if err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre) VALUES ('ana@luna.com', 'Ana') RETURNING id::text`).Scan(&ana)
	}); err != nil {
		t.Fatal(err)
	}
	f := &mpFalso{cuerpos: map[string]map[string]any{}, pagos: map[string]map[string]any{}, susc: map[string]map[string]any{}, usuario: 779, refreshOK: true}
	srv := httptest.NewServer(f)
	defer srv.Close()
	k := make([]byte, 32)
	_, _ = rand.Read(k)
	cif, _ := cifrado.Nuevo(base64.StdEncoding.EncodeToString(k))
	ahora := time.Now()
	s := &mercadopago.Servicio{Pool: b.App, Cifra: cif, Cliente: &mercadopago.Cliente{Base: srv.URL},
		Ahora:  func() time.Time { return ahora },
		Config: mercadopago.Config{AppID: "123", Secreto: "sec", SecretoWebhook: "whsec", Sitio: "https://tulocalweb.com.ar"}}
	u, err := s.Conectar(ctx, luna, ana)
	if err != nil {
		t.Fatal(err)
	}
	q, _ := url.Parse(u)
	if err := s.Volver(ctx, luna, ana, "codigo", q.Query().Get("state")); err != nil {
		t.Fatal(err)
	}
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Alias: "luna.mates"}); err != nil {
		t.Fatal(err)
	}
	if err := configuracion.GuardarEntregas(ctx, b.App, luna, configuracion.Entregas{Retiro: true, RetiroDireccion: "Gorriti 4500"}); err != nil {
		t.Fatal(err)
	}
	stock := 1
	mate, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 3000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cupones.Crear(ctx, b.App, luna, cupones.Nuevo{Codigo: "UNO10", Tipo: "pct", Valor: 10, Alcance: "todo",
		Quien: "todos", Limite: 1, PorCliente: "1", Vigencia: "sin", Canales: []string{"web"}}, ahora); err != nil {
		t.Fatal(err)
	}
	compra := func(pago, cupon string, cant int) (vidriera.Resultado, error) {
		c := vidriera.Compra{Email: "lucia@ejemplo.com", Nombre: "Lucía Pérez", Telefono: "11 5555-0101", CP: "1414",
			Envio: "retiro", Pago: pago, Cupon: cupon}
		c.Carrito = append(c.Carrito, vidriera.LineaCarrito{ID: mate.ID, Cantidad: cant})
		return vidriera.Comprar(ctx, b.App, s, s.Config.Sitio, "luna", c, ahora)
	}
	// Stock 1 y cupón de un uso: el segundo intento con lo mismo paga el mismo pedido.
	r1, err := compra("mercadopago", "UNO10", 1)
	if err != nil {
		t.Fatal(err)
	}
	r2, err := compra("mercadopago", "UNO10", 1)
	if err != nil || r2.Numero != r1.Numero || r2.Codigo != r1.Codigo || r2.URL == "" {
		t.Fatalf("pagar otra vez: %+v %v (antes %+v)", r2, err, r1)
	}
	// Sin el cupón: el de antes se cancela (vuelven el stock y el cupón) y sale uno nuevo.
	r3, err := compra("mercadopago", "", 1)
	if err != nil || r3.Numero == r1.Numero {
		t.Fatalf("cambió el carrito: %+v %v", r3, err)
	}
	d1, _ := pedidos.Obtener(ctx, b.App, luna, r1.Numero)
	if d1.Estado != "Cancelado" {
		t.Fatalf("el pedido reemplazado quedó %s", d1.Estado)
	}
	admin, err := pgx.Connect(ctx, b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)
	var usos int
	_ = admin.QueryRow(ctx, `SELECT usos FROM cupones WHERE codigo = 'UNO10'`).Scan(&usos)
	if usos != 0 {
		t.Fatalf("el cupón del pedido reemplazado no volvió: %d usos", usos)
	}
	if _, err := pedidos.CancelarReemplazado(ctx, b.App, luna, r3.Numero); err != nil {
		t.Fatal(err)
	}

	// Un cupón fijo mayor que el carrito, con retiro: $ 0, pago y sin Mercado Pago.
	if _, err := cupones.Crear(ctx, b.App, luna, cupones.Nuevo{Codigo: "REGALO", Tipo: "fijo", Valor: 5000, Alcance: "todo",
		Quien: "todos", PorCliente: "libre", Vigencia: "sin", Canales: []string{"web"}}, ahora); err != nil {
		t.Fatal(err)
	}
	prefs := len(f.vio("POST /checkout/preferences"))
	r4, err := compra("mercadopago", "REGALO", 1)
	if err != nil || r4.URL != "" || r4.Total != 0 {
		t.Fatalf("total cero: %+v %v", r4, err)
	}
	if len(f.vio("POST /checkout/preferences")) != prefs {
		t.Fatal("armó una preferencia de $ 0")
	}
	d4, _ := pedidos.Obtener(ctx, b.App, luna, r4.Numero)
	if d4.Estado != "Para preparar" {
		t.Fatalf("total cero quedó %s", d4.Estado)
	}
	// Cancelarlo (ya era una venta) le resta al cupón lo vendido y lo descontado.
	var vendio, desconto int64
	_ = admin.QueryRow(ctx, `SELECT vendio, desconto FROM cupones WHERE codigo = 'REGALO'`).Scan(&vendio, &desconto)
	if desconto == 0 {
		t.Fatalf("el cupón no sumó lo descontado: %d", desconto)
	}
	if err := pedidos.CambiarEstado(ctx, b.App, luna, r4.Numero, "Cancelado"); err != nil {
		t.Fatal(err)
	}
	_ = admin.QueryRow(ctx, `SELECT vendio, desconto FROM cupones WHERE codigo = 'REGALO'`).Scan(&vendio, &desconto)
	if vendio != 0 || desconto != 0 {
		t.Fatalf("cancelado y el cupón sigue con vendió %d descontó %d", vendio, desconto)
	}
	if err := pedidos.CambiarEstado(ctx, b.App, luna, r4.Numero, "Para preparar"); err != nil {
		t.Fatal(err)
	}
	_ = admin.QueryRow(ctx, `SELECT desconto FROM cupones WHERE codigo = 'REGALO'`).Scan(&desconto)
	if desconto == 0 {
		t.Fatal("reabrirlo no le volvió a sumar al cupón")
	}
}
