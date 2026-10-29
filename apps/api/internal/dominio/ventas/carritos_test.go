package ventas_test

import (
	"context"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/ventas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

func TestCarritosAbandonados(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	// Con datos para transferir (sin ellos la tienda no ofrece transferencia).
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Alias: "luna.tienda"}); err != nil {
		t.Fatal(err)
	}
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	mails := &correo.Memoria{}
	w := &outbox.Worker{Pool: b.Worker, Manejadores: (&ventas.MailsCarritos{Pool: b.App, Correo: mails, Sitio: "https://tulocalweb.com.ar"}).Manejadores()}
	stock := 10
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	linea := []ventas.LineaCarrito{{ID: p.ID, Cantidad: 2}}
	ahora := time.Now()

	if err := ventas.GuardarCarrito(ctx, b.App, luna, "no-es-un-email", "", linea); err != ventas.ErrCarritoInvalido {
		t.Fatalf("email inválido: %v", err)
	}
	// Un producto de otra tienda no cuenta: el carrito queda vacío y no se guarda.
	if err := ventas.GuardarCarrito(ctx, b.App, sol, "ana@ejemplo.com", "Ana", linea); err != ventas.ErrCarritoInvalido {
		t.Fatalf("producto ajeno: %v", err)
	}
	if err := ventas.GuardarCarrito(ctx, b.App, luna, "Lucia@Ejemplo.com", "Lucía", linea); err != nil {
		t.Fatal(err)
	}
	// El mismo email actualiza el mismo carrito (no hay dos abiertos).
	if err := ventas.GuardarCarrito(ctx, b.App, luna, "lucia@ejemplo.com", "", []ventas.LineaCarrito{{ID: p.ID, Cantidad: 3}}); err != nil {
		t.Fatal(err)
	}
	cs, abiertos, monto, _, _, err := ventas.CarritosPanel(ctx, b.App, luna, ahora)
	if err != nil || len(cs) != 1 || abiertos != 1 || monto != 30000 || cs[0].Cantidad != 3 || cs[0].Nombre != "Lucía" {
		t.Fatalf("panel: %+v %d %d %v", cs, abiertos, monto, err)
	}
	if cs2, _, _, _, _, _ := ventas.CarritosPanel(ctx, b.App, sol, ahora); len(cs2) != 0 {
		t.Fatal("otra tienda ve carritos ajenos")
	}

	// Recordar: mail con el link que restaura el carrito; uno por día.
	if err := ventas.RecordarCarrito(ctx, b.App, luna, cs[0].ID, ahora); err != nil {
		t.Fatal(err)
	}
	if err := ventas.RecordarCarrito(ctx, b.App, luna, cs[0].ID, ahora.Add(time.Hour)); err != ventas.ErrRecienRecordado {
		t.Fatalf("dos el mismo día: %v", err)
	}
	if err := ventas.RecordarCarrito(ctx, b.App, sol, cs[0].ID, ahora); err != ventas.ErrNoExisteCarrito {
		t.Fatalf("otra tienda recordó: %v", err)
	}
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	m, ok := mails.Ultimo("lucia@ejemplo.com")
	if !ok || !strings.Contains(m.Texto, "3 × Mate") || !strings.Contains(m.Texto, "/t/luna/carrito/recuperar/") {
		t.Fatalf("mail: %+v", m)
	}
	token := m.Texto[strings.Index(m.Texto, "/recuperar/")+len("/recuperar/"):]
	token = strings.Fields(token)[0]
	if ls, err := ventas.CarritoPorToken(ctx, b.App, luna, token); err != nil || len(ls) != 1 || ls[0].Cantidad != 3 {
		t.Fatalf("restaurar: %+v %v", ls, err)
	}

	// Compra con el mismo email: el carrito queda recuperado y deja de aparecer.
	c := vidriera.Compra{Email: "lucia@ejemplo.com", Nombre: "Lucía Pérez", Telefono: "11 5555-0101", CP: "1414",
		Direccion: "Gorriti 4500", Envio: "andreani", Pago: "transferencia"}
	c.Carrito = append(c.Carrito, vidriera.LineaCarrito{ID: p.ID, Cantidad: 1})
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", c, ahora); err != nil {
		t.Fatal(err)
	}
	if cs, _, _, _, _, _ := ventas.CarritosPanel(ctx, b.App, luna, ahora); len(cs) != 0 {
		t.Fatalf("sigue abierto después de comprar: %+v", cs)
	}

	// Automático al tercer día, sólo si la tienda lo prendió.
	if err := ventas.GuardarCarrito(ctx, b.App, luna, "ana@ejemplo.com", "Ana", linea); err != nil {
		t.Fatal(err)
	}
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE carritos SET actualizado = now() - interval '3 days 12 hours' WHERE email = 'ana@ejemplo.com'`)
		return err
	})
	if n, _ := ventas.RecordatoriosAutomaticos(ctx, b.App, ahora); n != 0 {
		t.Fatalf("recordó con el automático apagado: %d", n)
	}
	if err := ventas.GuardarAutomaticos(ctx, b.App, luna, ventas.Automaticos{Email: true}); err != nil {
		t.Fatal(err)
	}
	if n, err := ventas.RecordatoriosAutomaticos(ctx, b.App, ahora); err != nil || n != 1 {
		t.Fatalf("automático: %d %v", n, err)
	}
}
