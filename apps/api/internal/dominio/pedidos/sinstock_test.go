package pedidos_test

import (
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// avisosSinStock lee del outbox (como superusuario) los avisos de sin stock encolados.
func avisosSinStock(t *testing.T, b *pruebadb.Bases) (mails, push []pedidos.SinStock) {
	t.Helper()
	conn, err := pgx.Connect(ctx, b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)
	filas, err := conn.Query(ctx, `SELECT tipo, datos FROM outbox WHERE tipo IN ('tienda.sin_stock', 'push.sin_stock') ORDER BY id`)
	if err != nil {
		t.Fatal(err)
	}
	defer filas.Close()
	for filas.Next() {
		var tipo string
		var datos []byte
		if err := filas.Scan(&tipo, &datos); err != nil {
			t.Fatal(err)
		}
		var d pedidos.SinStock
		_ = json.Unmarshal(datos, &d)
		if tipo == "tienda.sin_stock" {
			mails = append(mails, d)
		} else {
			push = append(push, d)
		}
	}
	return mails, push
}

func adminExec(t *testing.T, b *pruebadb.Bases, sql string, args ...any) {
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

func TestSinStockSeEncolaAlLlegarACero(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	mate, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate de calabaza", Precio: 12000, Stock: ptr(3)})
	if err != nil {
		t.Fatal(err)
	}
	vender := func(cant int) {
		t.Helper()
		if _, _, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, venta(pedidos.Linea{ProductoID: mate.ID, Cantidad: cant})); err != nil {
			t.Fatal(err)
		}
	}

	// Baja sin llegar a 0: nada.
	vender(2)
	if m, p := avisosSinStock(t, b); len(m)+len(p) != 0 {
		t.Fatalf("con stock 1 se encolaron %d mails y %d push", len(m), len(p))
	}
	// Llega a 0: un mail y un push, con el producto.
	vender(1)
	m, p := avisosSinStock(t, b)
	if len(m) != 1 || len(p) != 1 || m[0].Producto != mate.ID || m[0].Variante != "" || p[0] != m[0] {
		t.Fatalf("al llegar a 0: mails %+v, push %+v", m, p)
	}

	// Repone y vuelve a 0 el mismo día: no se repite.
	adminExec(t, b, `UPDATE productos SET stock = 2 WHERE id = $1`, mate.ID)
	vender(2)
	if m, p := avisosSinStock(t, b); len(m) != 1 || len(p) != 1 {
		t.Fatalf("el mismo día se volvió a encolar: %d mails, %d push", len(m), len(p))
	}

	// Al otro día sí.
	adminExec(t, b, `UPDATE productos SET stock = 1, sin_stock_avisado = sin_stock_avisado - 1 WHERE id = $1`, mate.ID)
	vender(1)
	if m, _ := avisosSinStock(t, b); len(m) != 2 {
		t.Fatalf("al otro día: %d mails, quería 2", len(m))
	}

	// Un producto pausado que llega a 0 (lo vende el local) no avisa.
	pausado, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Bombilla", Precio: 5000, Stock: ptr(1)})
	if err != nil {
		t.Fatal(err)
	}
	adminExec(t, b, `UPDATE productos SET publicado = false WHERE id = $1`, pausado.ID)
	if _, _, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, venta(pedidos.Linea{ProductoID: pausado.ID, Cantidad: 1})); err != nil {
		t.Fatal(err)
	}
	if m, _ := avisosSinStock(t, b); len(m) != 2 {
		t.Fatalf("un producto pausado avisó: %d mails", len(m))
	}

	// Si la venta falla (no alcanza el stock de otro producto), no queda nada encolado.
	yerba, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Yerba", Precio: 4000, Stock: ptr(1)})
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = pedidos.CrearDesdeProductos(ctx, b.App, luna, venta(pedidos.Linea{ProductoID: yerba.ID, Cantidad: 1}, pedidos.Linea{ProductoID: pausado.ID, Cantidad: 1}))
	if err == nil {
		t.Fatal("vendió sin stock")
	}
	if m, _ := avisosSinStock(t, b); len(m) != 2 {
		t.Fatalf("una venta que no se hizo encoló el aviso: %d mails", len(m))
	}
}

func TestSinStockDeUnaVariante(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	p := remera(t, b, luna, map[string]int{"MNegro": 1, "SBlanco": 2})
	mNegro, sBlanco := idDe(p, "M", "Negro"), idDe(p, "S", "Blanco")
	// S Blanco baja a 1: nada.
	if _, _, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, venta(pedidos.Linea{ProductoID: p.ID, VarianteID: sBlanco, Cantidad: 1})); err != nil {
		t.Fatal(err)
	}
	if m, _ := avisosSinStock(t, b); len(m) != 0 {
		t.Fatalf("sin llegar a 0: %d", len(m))
	}
	// M Negro llega a 0 aunque al producto le quede stock: avisa con la variante.
	if _, _, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, venta(pedidos.Linea{ProductoID: p.ID, VarianteID: mNegro, Cantidad: 1})); err != nil {
		t.Fatal(err)
	}
	m, _ := avisosSinStock(t, b)
	if len(m) != 1 || m[0].Producto != p.ID || m[0].Variante != mNegro {
		t.Fatalf("variante: %+v", m)
	}
	var a pedidos.Agotado
	if err := b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		var err error
		a, err = pedidos.LeerSinStock(ctx, tx, m[0])
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if a.Nombre != "Remera lisa" || a.Variante != "Talle M · Color Negro" || !a.Sigue {
		t.Fatalf("agotado: %+v", a)
	}
	// Otra variante a 0 el mismo día: uno por producto por día.
	if _, _, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, venta(pedidos.Linea{ProductoID: p.ID, VarianteID: sBlanco, Cantidad: 1})); err != nil {
		t.Fatal(err)
	}
	if m, _ := avisosSinStock(t, b); len(m) != 1 {
		t.Fatalf("otra variante del mismo producto, mismo día: %d", len(m))
	}
}
