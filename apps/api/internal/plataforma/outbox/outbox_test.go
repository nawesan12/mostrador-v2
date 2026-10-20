package outbox_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

var ctx = context.Background()

func TestElEventoViveYMuereConSuTransaccion(t *testing.T) {
	b := pruebadb.Nueva(t)
	tienda := pruebadb.CrearTienda(t, b.App, "luna")

	if err := b.App.EnTienda(ctx, tienda, func(tx pgx.Tx) error {
		return outbox.Publicar(ctx, tx, tienda, "pedido.pagado", map[string]int{"numero": 1})
	}); err != nil {
		t.Fatal(err)
	}
	falla := errors.New("se cayó el cambio")
	err := b.App.EnTienda(ctx, tienda, func(tx pgx.Tx) error {
		if err := outbox.Publicar(ctx, tx, tienda, "pedido.pagado", map[string]int{"numero": 2}); err != nil {
			return err
		}
		return falla
	})
	if !errors.Is(err, falla) {
		t.Fatalf("err = %v", err)
	}
	var n int
	if err := b.Worker.QueryRow(ctx, "SELECT count(*) FROM outbox").Scan(&n); err != nil || n != 1 {
		t.Fatalf("eventos = %d (err %v), quería 1: el revertido no tiene que quedar", n, err)
	}
}

func TestLaAppNoPuedeMarcarEventos(t *testing.T) {
	b := pruebadb.Nueva(t)
	tienda := pruebadb.CrearTienda(t, b.App, "luna")
	err := b.App.EnTienda(ctx, tienda, func(tx pgx.Tx) error {
		if err := outbox.Publicar(ctx, tx, tienda, "pedido.pagado", nil); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, "UPDATE outbox SET publicado_en = now()")
		return err
	})
	if err == nil {
		t.Fatal("la app pudo marcar el outbox como publicado")
	}
}

func publicar(t *testing.T, b *pruebadb.Bases, tienda string, n int) {
	t.Helper()
	if err := b.App.EnTienda(ctx, tienda, func(tx pgx.Tx) error {
		for i := 0; i < n; i++ {
			if err := outbox.Publicar(ctx, tx, tienda, "pedido.pagado", map[string]int{"i": i}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
}

func TestReintentaConEsperaYAparta(t *testing.T) {
	b := pruebadb.Nueva(t)
	tienda := pruebadb.CrearTienda(t, b.App, "luna")
	publicar(t, b, tienda, 1)

	w := &outbox.Worker{Pool: b.Worker, Manejadores: map[string]outbox.Manejador{
		"pedido.pagado": func(context.Context, outbox.Evento) error { return errors.New("el proveedor no responde") },
	}}
	if n, err := w.Pasada(ctx); err != nil || n != 1 {
		t.Fatalf("primera pasada: n=%d err=%v", n, err)
	}
	// Con la espera puesta, una pasada inmediata no lo vuelve a tomar.
	if n, _ := w.Pasada(ctx); n != 0 {
		t.Fatalf("lo retomó sin esperar (n=%d)", n)
	}
	var intentos int
	var errTxt string
	if err := b.Worker.QueryRow(ctx, "SELECT intentos, ultimo_error FROM outbox").Scan(&intentos, &errTxt); err != nil {
		t.Fatal(err)
	}
	if intentos != 1 || errTxt != "el proveedor no responde" {
		t.Fatalf("intentos=%d error=%q", intentos, errTxt)
	}
	// Llegando al tope queda apartado (infinity) con su error, sin publicarse.
	if _, err := b.Worker.Exec(ctx, "UPDATE outbox SET intentos = $1, proximo_intento = now()", outbox.MaxIntentos-1); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	var apartado bool
	if err := b.Worker.QueryRow(ctx, "SELECT proximo_intento = 'infinity' AND publicado_en IS NULL FROM outbox").Scan(&apartado); err != nil || !apartado {
		t.Fatalf("no quedó apartado (err %v)", err)
	}
}

func TestUnPanicNoTiraElWorker(t *testing.T) {
	b := pruebadb.Nueva(t)
	tienda := pruebadb.CrearTienda(t, b.App, "luna")
	publicar(t, b, tienda, 1)
	w := &outbox.Worker{Pool: b.Worker, Manejadores: map[string]outbox.Manejador{
		"pedido.pagado": func(context.Context, outbox.Evento) error { panic("bug") },
	}}
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatalf("el panic se escapó: %v", err)
	}
}

// Varios workers en paralelo procesan cada evento una sola vez.
func TestDosWorkersNoSePisan(t *testing.T) {
	b := pruebadb.Nueva(t)
	tienda := pruebadb.CrearTienda(t, b.App, "luna")
	const total = 300
	publicar(t, b, tienda, total)

	var mu sync.Mutex
	vistos := map[int64]int{}
	var procesados atomic.Int64
	man := func(_ context.Context, ev outbox.Evento) error {
		mu.Lock()
		vistos[ev.ID]++
		mu.Unlock()
		procesados.Add(1)
		return nil
	}
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			w := &outbox.Worker{Pool: b.Worker, Lote: 25, Manejadores: map[string]outbox.Manejador{"pedido.pagado": man}}
			for {
				n, err := w.Pasada(ctx)
				if err != nil {
					t.Error(err)
					return
				}
				if n == 0 {
					return
				}
			}
		}()
	}
	wg.Wait()
	if procesados.Load() != total || len(vistos) != total {
		t.Fatalf("procesados=%d distintos=%d, quería %d", procesados.Load(), len(vistos), total)
	}
	for id, veces := range vistos {
		if veces != 1 {
			t.Fatalf("el evento %d se procesó %d veces", id, veces)
		}
	}
}

func TestEsperaCreciente(t *testing.T) {
	if outbox.Espera(1).Seconds() != 2 || outbox.Espera(3).Seconds() != 8 || outbox.Espera(30).Hours() != 1 {
		t.Fatalf("esperas: %v %v %v", outbox.Espera(1), outbox.Espera(3), outbox.Espera(30))
	}
}
