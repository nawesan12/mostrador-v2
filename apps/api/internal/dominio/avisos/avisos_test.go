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
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

func TestMailsAQuienCompra(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	mails := &correo.Memoria{}
	av := &avisos.Servicio{Pool: b.App, Correo: mails, Sitio: "https://mostradorapp.com.ar"}
	w := &outbox.Worker{Pool: b.Worker, Manejadores: av.Manejadores()}
	stock := 5
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	compra := vidriera.Compra{Email: "lucia@ejemplo.com", Nombre: "Lucía Pérez", Telefono: "11 5555-0101", CP: "1414",
		Direccion: "Gorriti 4500", Envio: "andreani", Pago: "transferencia"}
	compra.Carrito = append(compra.Carrito, struct {
		ID       string `json:"id"`
		Cantidad int    `json:"cantidad"`
	}{p.ID, 1})

	// Sin datos de transferencia, la tienda no la ofrece.
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", compra, time.Now()); !errors.Is(err, vidriera.ErrInvalido) {
		t.Fatalf("transferencia sin alias: %v", err)
	}
	// CBU con el dígito verificador mal: no se guarda.
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", CBU: "2850590940090418135209"}); err == nil {
		t.Fatal("aceptó un CBU inválido")
	}
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Alias: "luna.mates", CBU: "2850590940090418135201", Titular: "Luna SRL"}); err != nil {
		t.Fatal(err)
	}
	r, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", compra, time.Now())
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
	for _, want := range []string{"luna.mates", "2850590940090418135201", "Luna SRL", "1 × Mate", "/t/luna/pedido/" + r.Codigo} {
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
}
