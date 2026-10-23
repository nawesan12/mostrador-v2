package pedidos_test

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func ptr[T any](v T) *T { return &v }

func venta(lineas ...pedidos.Linea) pedidos.PedidoNuevo {
	return pedidos.PedidoNuevo{Pago: "efectivo", Canal: "pos", Items: lineas}
}

// remera arma un producto con Talle S/M/L × Color Negro/Blanco y stock en algunas variantes.
func remera(t *testing.T, b *pruebadb.Bases, tienda string, stock map[string]int) listas.Producto {
	t.Helper()
	e := listas.Editable{Nombre: "Remera lisa", Precio: 20000, Opciones: []listas.Opcion{
		{Nombre: "Talle", Valores: []string{"S", "M", "L"}}, {Nombre: "Color", Valores: []string{"Negro", "Blanco"}}}}
	for k, n := range stock {
		e.Variantes = append(e.Variantes, listas.VarianteEditable{Valores: map[string]string{"Talle": k[:1], "Color": k[1:]}, Stock: ptr(n)})
	}
	p, err := listas.Guardar(ctx, b.App, tienda, "", e)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func idDe(p listas.Producto, talle, color string) string {
	for _, v := range p.Variantes {
		if v.Valores["Talle"] == talle && v.Valores["Color"] == color {
			return v.ID
		}
	}
	return ""
}

func TestVentaPorVariante(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	p := remera(t, b, luna, map[string]int{"MNegro": 3, "SBlanco": 1})
	mNegro := idDe(p, "M", "Negro")
	// Precio propio en M Negro.
	if _, err := listas.Guardar(ctx, b.App, luna, p.ID, listas.Editable{Nombre: p.Nombre, Precio: 20000,
		Variantes: []listas.VarianteEditable{{Valores: map[string]string{"Talle": "M", "Color": "Negro"}, Precio: ptr(int64(22000))}}}); err != nil {
		t.Fatal(err)
	}

	num, total, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, venta(pedidos.Linea{ProductoID: p.ID, VarianteID: mNegro, Cantidad: 2}))
	if err != nil {
		t.Fatal(err)
	}
	if total != 44000 {
		t.Fatalf("total con el precio de la variante: %d", total)
	}
	d, _ := pedidos.Obtener(ctx, b.App, luna, num)
	if it := d.Items[0]; it.Variante == nil || *it.Variante != "Talle M · Color Negro" || *it.VarianteID != mNegro || it.Precio != 22000 {
		t.Fatalf("ítem: %+v", it)
	}
	p, _ = listas.Obtener(ctx, b.App, luna, p.ID, time.Now())
	if p.Stock != 2 {
		t.Fatalf("stock del producto (suma): %d", p.Stock)
	}
	for _, v := range p.Variantes {
		if v.ID == mNegro && v.Stock != 1 {
			t.Fatalf("stock de la variante: %d", v.Stock)
		}
	}

	casos := []struct {
		nombre string
		linea  pedidos.Linea
		quiero error
	}{
		{"sin variante", pedidos.Linea{ProductoID: p.ID, Cantidad: 1}, pedidos.ErrSinVariante},
		{"sin stock", pedidos.Linea{ProductoID: p.ID, VarianteID: mNegro, Cantidad: 2}, pedidos.ErrSinStock},
		{"variante inventada", pedidos.Linea{ProductoID: p.ID, VarianteID: "00000000-0000-0000-0000-000000000001", Cantidad: 1}, pedidos.ErrVariante},
	}
	// Variante de otro producto de la misma tienda.
	otro := remera(t, b, luna, map[string]int{"LNegro": 5})
	casos = append(casos, struct {
		nombre string
		linea  pedidos.Linea
		quiero error
	}{"variante de otro producto", pedidos.Linea{ProductoID: p.ID, VarianteID: idDe(otro, "L", "Negro"), Cantidad: 1}, pedidos.ErrVariante})
	// Variante de otra tienda (con su producto): para esta tienda no existe.
	ajena := remera(t, b, sol, map[string]int{"MNegro": 5})
	casos = append(casos, struct {
		nombre string
		linea  pedidos.Linea
		quiero error
	}{"variante de otra tienda", pedidos.Linea{ProductoID: ajena.ID, VarianteID: idDe(ajena, "M", "Negro"), Cantidad: 1}, pedidos.ErrProducto})
	for _, c := range casos {
		_, _, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, venta(c.linea))
		if c.quiero == pedidos.ErrProducto {
			if !errors.Is(err, pedidos.ErrVariante) && !errors.Is(err, pedidos.ErrProducto) {
				t.Errorf("%s: %v", c.nombre, err)
			}
			continue
		}
		if !errors.Is(err, c.quiero) {
			t.Errorf("%s: %v, quería %v", c.nombre, err, c.quiero)
		}
	}

	// Inactiva: no se vende aunque tenga stock.
	if _, err := listas.Guardar(ctx, b.App, luna, p.ID, listas.Editable{Nombre: p.Nombre, Precio: 20000,
		Variantes: []listas.VarianteEditable{{Valores: map[string]string{"Talle": "S", "Color": "Blanco"}, Activa: ptr(false)}}}); err != nil {
		t.Fatal(err)
	}
	if _, _, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, venta(pedidos.Linea{ProductoID: p.ID, VarianteID: idDe(p, "S", "Blanco"), Cantidad: 1})); !errors.Is(err, pedidos.ErrVariante) {
		t.Fatalf("vendió una variante inactiva: %v", err)
	}
	// Un rechazo en una línea no descuenta las otras.
	_, _, err = pedidos.CrearDesdeProductos(ctx, b.App, luna, venta(
		pedidos.Linea{ProductoID: otro.ID, VarianteID: idDe(otro, "L", "Negro"), Cantidad: 1},
		pedidos.Linea{ProductoID: p.ID, Cantidad: 1}))
	otro, _ = listas.Obtener(ctx, b.App, luna, otro.ID, time.Now())
	if !errors.Is(err, pedidos.ErrSinVariante) || otro.Stock != 5 {
		t.Fatalf("transacción: %v, stock %d", err, otro.Stock)
	}
}

// Dos cajas venden la última unidad de la misma variante a la vez: una sola lo logra.
func TestVariantesSinSobreventa(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	p := remera(t, b, luna, map[string]int{"MNegro": 3, "LNegro": 2})
	m, l := idDe(p, "M", "Negro"), idDe(p, "L", "Negro")

	var wg sync.WaitGroup
	var mu sync.Mutex
	vendidas, sinStock := 0, 0
	for i := range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// La mitad lleva las dos variantes en orden inverso (no se tienen que trabar).
			lineas := []pedidos.Linea{{ProductoID: p.ID, VarianteID: m, Cantidad: 1}, {ProductoID: p.ID, VarianteID: l, Cantidad: 1}}
			if i%2 == 1 {
				lineas[0], lineas[1] = lineas[1], lineas[0]
			}
			_, _, err := pedidos.CrearDesdeProductos(ctx, b.App, luna, venta(lineas...))
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err == nil:
				vendidas++
			case errors.Is(err, pedidos.ErrSinStock):
				sinStock++
			default:
				t.Errorf("venta: %v", err)
			}
		}()
	}
	wg.Wait()
	if vendidas != 2 || sinStock != 8 {
		t.Fatalf("vendidas %d, sin stock %d", vendidas, sinStock)
	}
	p, _ = listas.Obtener(ctx, b.App, luna, p.ID, time.Now())
	if p.Stock != 1 {
		t.Fatalf("queda %d (quería 1: M Negro)", p.Stock)
	}
}
