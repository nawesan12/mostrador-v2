package precios_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/precios"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

var ctx = context.Background()

func TestAumentoRedondeaYSeDeshace(t *testing.T) {
	b := pruebadb.Nueva(t)
	id := pruebadb.CrearTienda(t, b.App, "luna")
	otra := pruebadb.CrearTienda(t, b.App, "sol")
	p, _ := listas.Guardar(ctx, b.App, id, "", listas.Editable{Nombre: "Termo", Precio: 42900})
	q, _ := listas.Guardar(ctx, b.App, otra, "", listas.Editable{Nombre: "Termo", Precio: 42900})
	precio := func(tienda, prod string) int64 {
		x, err := listas.Obtener(ctx, b.App, tienda, prod, time.Now())
		if err != nil {
			t.Fatal(err)
		}
		return x.Precio
	}
	n, para, err := precios.Aplicar(ctx, b.App, id, 8, "90", false, time.Now())
	if err != nil || n != 1 || para != nil {
		t.Fatalf("aplicar: %d %v %v", n, para, err)
	}
	if got := precio(id, p.ID); got != 46290 {
		t.Fatalf("precio nuevo: %d", got)
	}
	if got := precio(otra, q.ID); got != 42900 {
		t.Fatalf("tocó otra tienda: %d", got)
	}
	if err := precios.Deshacer(ctx, b.App, id); err != nil {
		t.Fatal(err)
	}
	if got := precio(id, p.ID); got != 42900 {
		t.Fatalf("deshacer: %d", got)
	}
	if err := precios.Deshacer(ctx, b.App, id); !errors.Is(err, precios.ErrNoExiste) {
		t.Fatalf("deshizo dos veces: %v", err)
	}

	// Programado: se ve hasta el lunes; pasado el lunes, se aplica al mirar.
	jueves := time.Date(2026, 10, 22, 18, 0, 0, 0, time.UTC)
	if _, para, _ := precios.Aplicar(ctx, b.App, id, 10, "500", true, jueves); para == nil || !para.Equal(time.Date(2026, 10, 26, 3, 0, 0, 0, time.UTC)) {
		t.Fatalf("lunes: %v", para)
	}
	l, _ := precios.Leer(ctx, b.App, id, jueves)
	if l.Programado == nil || l.Programado.Pct != 10 || len(l.Listas) != 4 || l.Productos != 1 {
		t.Fatalf("listas: %+v", l)
	}
	l, _ = precios.Leer(ctx, b.App, id, jueves.Add(5*24*time.Hour))
	if l.Programado != nil || precio(id, p.ID) != 47000 {
		t.Fatalf("no se aplicó lo programado: %+v %d", l.Programado, precio(id, p.ID))
	}
	if err := precios.Cancelar(ctx, b.App, id); !errors.Is(err, precios.ErrNoExiste) {
		t.Fatalf("cancelar sin programado: %v", err)
	}
}
