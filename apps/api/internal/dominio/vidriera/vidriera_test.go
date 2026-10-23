package vidriera_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cupones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

type mpFalso struct{ pedidos []int }

func (m *mpFalso) PagarPedido(_ context.Context, _ string, numero int, vuelta string) (string, error) {
	m.pedidos = append(m.pedidos, numero)
	return "https://www.mercadopago.com.ar/checkout/v1/redirect?pref_id=x&vuelta=" + vuelta, nil
}

func compra(id string, cantidad int, pago string) vidriera.Compra {
	c := vidriera.Compra{Email: "lucia@ejemplo.com", Nombre: "Lucía Pérez", Telefono: "11 5555-0101", CP: "1414",
		Direccion: "Gorriti 4500", Envio: "andreani", Pago: pago}
	c.Carrito = append(c.Carrito, struct {
		ID       string `json:"id"`
		Cantidad int    `json:"cantidad"`
	}{id, cantidad})
	return c
}

func TestTiendaOnline(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	// Sin alias ni CBU no se ofrece transferencia: la tienda los carga.
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Alias: "luna.tienda"}); err != nil {
		t.Fatal(err)
	}
	stock := 3
	cat := "Mates"
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate de calabaza", Categoria: &cat, Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}

	tienda, err := vidriera.Publica(ctx, b.App, "luna")
	if err != nil || len(tienda.Productos) != 1 || len(tienda.Categorias) != 1 || tienda.Categorias[0].Cantidad != 1 {
		t.Fatalf("pública: %+v %v", tienda, err)
	}
	if _, err := vidriera.Publica(ctx, b.App, "no-existe"); !errors.Is(err, vidriera.ErrNoExiste) {
		t.Fatal("una tienda que no existe")
	}

	// Transferencia: precio del servidor, envío por región, 15 % off y el código público.
	lunes := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	r, err := vidriera.Comprar(ctx, b.App, nil, "https://mostrador.iamnaw.dev", "luna", compra(p.ID, 2, "transferencia"), lunes)
	if err != nil {
		t.Fatal(err)
	}
	if want := int64(20000 + 6900 - 3000); r.Total != want || r.URL != "" || len(r.Codigo) < 16 {
		t.Fatalf("resultado: %+v (quería total %d)", r, want)
	}
	c, err := vidriera.VerPedido(ctx, b.App, "luna", r.Codigo)
	if err != nil || c.Numero != r.Numero || c.Total != r.Total || c.Nombre != "Lucía" || c.Pago != "transferencia" {
		t.Fatalf("confirmación: %+v %v", c, err)
	}
	if _, err := vidriera.VerPedido(ctx, b.App, "sol", r.Codigo); !errors.Is(err, vidriera.ErrNoExiste) {
		t.Fatal("otra tienda vio el pedido")
	}
	var quedan int
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT stock FROM productos WHERE id::text = $1`, p.ID).Scan(&quedan)
	})
	if quedan != 1 {
		t.Fatalf("stock: %d", quedan)
	}

	// Sin stock, carrito de otra tienda, y Mercado Pago sin cuenta conectada: no se arma el pedido.
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", compra(p.ID, 5, "transferencia"), lunes); !errors.Is(err, vidriera.ErrSinStock) {
		t.Fatalf("sin stock: %v", err)
	}
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "sol", compra(p.ID, 1, "transferencia"), lunes); !errors.Is(err, vidriera.ErrInvalido) {
		t.Fatalf("producto de otra tienda: %v", err)
	}
	mp := &mpFalso{}
	if _, err := vidriera.Comprar(ctx, b.App, mp, "", "luna", compra(p.ID, 1, "mercadopago"), lunes); !errors.Is(err, vidriera.ErrInvalido) {
		t.Fatalf("Mercado Pago sin conectar: %v", err)
	}

	// Una tienda cerrada no se ve.
	_ = b.App.EnTienda(ctx, sol, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE tiendas SET cerrada_en = now() WHERE id = $1`, sol)
		return err
	})
	if _, err := vidriera.Publica(ctx, b.App, "sol"); !errors.Is(err, vidriera.ErrNoExiste) {
		t.Fatal("se ve una tienda cerrada")
	}
}

func TestCotizar(t *testing.T) {
	lunes := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	gratis := int64(100000)
	e := vidriera.Envios{GratisDesde: &gratis, Retiro: true, Moto: true, Correo: true}
	o := vidriera.Cotizar("1414", 5000, e, lunes)
	if len(o) != 4 || o[0].Precio != 6900 || o[0].Detalle != "Llega el miércoles 28" || o[1].Precio != 5500 {
		t.Fatalf("CABA: %+v", o)
	}
	if o := vidriera.Cotizar("5000", 150000, e, lunes); o[0].Precio != 0 || len(o) != 3 {
		t.Fatalf("Córdoba con envío gratis: %+v", o)
	}
	if o := vidriera.Cotizar("x", 1, e, lunes); len(o) != 1 || o[0].ID != "retiro" {
		t.Fatalf("CP inválido: %+v", o)
	}
}

func TestCompraConCupon(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	stock := 50
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	ahora := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	nuevo := func(codigo, tipo string, valor int, limite int, porCliente string) {
		t.Helper()
		if _, err := cupones.Crear(ctx, b.App, luna, cupones.Nuevo{Codigo: codigo, Tipo: tipo, Valor: valor, Alcance: "todo",
			Quien: "todos", Limite: limite, PorCliente: porCliente, Vigencia: "sin", Canales: []string{"web"}}, ahora); err != nil {
			t.Fatal(err)
		}
	}
	nuevo("DIEZ", "pct", 10, 0, "1")
	nuevo("UNO", "fijo", 1000, 1, "libre")
	nuevo("ENVIO", "envio", 0, 0, "libre")

	// 10 % sobre $ 20.000 con transferencia: 2.000 de cupón y 15 % de lo que queda (2.700). Sin
	// "combinable" en el cupón, la transferencia no se suma: sólo el cupón.
	c := compra(p.ID, 2, "transferencia")
	c.Cupon = "diez"
	r, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c, ahora)
	if err != nil {
		t.Fatal(err)
	}
	if want := int64(20000 + 6900 - 2000); r.Total != want {
		t.Fatalf("total con cupón: %d, quería %d", r.Total, want)
	}
	// Una vez por cliente: el mismo email ya no puede.
	var ec vidriera.ErrCupon
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c, ahora); !errors.As(err, &ec) || ec.Motivo != "Ya usaste este cupón." {
		t.Fatalf("segundo uso: %v", err)
	}
	// Límite total de 1: el segundo, aunque sea otra persona, ya no.
	c2 := compra(p.ID, 1, "mercadopago")
	c2.Pago, c2.Cupon = "transferencia", "UNO"
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c2, ahora); err != nil {
		t.Fatal(err)
	}
	c2.Email = "otra@ejemplo.com"
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c2, ahora); !errors.As(err, &ec) {
		t.Fatalf("límite: %v", err)
	}
	// Envío gratis y código inexistente.
	c3 := compra(p.ID, 1, "transferencia")
	c3.Email, c3.Cupon = "tres@ejemplo.com", "ENVIO"
	r3, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c3, ahora)
	// No combinable: el envío sale gratis pero no se suma el 15 % de transferencia.
	if err != nil || r3.Total != 10000 {
		t.Fatalf("envío gratis: %+v %v", r3, err)
	}
	c3.Cupon = "NOEXISTE"
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c3, ahora); !errors.As(err, &ec) || !strings.Contains(ec.Motivo, "no existe") {
		t.Fatalf("inexistente: %v", err)
	}
	// El cupón suma usos y lo que descontó.
	var usos int
	var desconto int64
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT usos, desconto FROM cupones WHERE codigo = 'DIEZ'`).Scan(&usos, &desconto)
	})
	if usos != 1 || desconto != 2000 {
		t.Fatalf("contadores: %d %d", usos, desconto)
	}
}
