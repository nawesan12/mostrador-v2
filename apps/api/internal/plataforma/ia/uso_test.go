package ia_test

import (
	"context"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/ia"
)

// La medición va a la base: cada respuesta suma a la tienda (por día y ruta) y al gasto del mes de
// la plataforma; el motor cuenta aparte y no gasta.
func TestRegistrarYTopes(t *testing.T) {
	b := pruebadb.Nueva(t)
	ctx := context.Background()
	una := pruebadb.CrearTienda(t, b.App, "una")
	otra := pruebadb.CrearTienda(t, b.App, "otra")
	ahora := time.Date(2026, 11, 18, 15, 0, 0, 0, time.UTC)
	ayer := ahora.Add(-24 * time.Hour)
	mesPasado := time.Date(2026, 10, 30, 15, 0, 0, 0, time.UTC)

	u := ia.Uso{Entrada: 1000, CacheLeido: 5000, Salida: 200}
	for _, r := range []struct {
		tienda, ruta string
		cuando       time.Time
	}{
		{una, ia.RutaChat, ahora}, {una, ia.RutaChat, ahora}, {una, ia.RutaMejorar, ayer},
		{una, ia.RutaMotor, ahora}, {una, ia.RutaChat, mesPasado}, {otra, ia.RutaTelegram, ahora},
	} {
		uso := u
		if r.ruta == ia.RutaMotor {
			uso = ia.Uso{}
		}
		if err := ia.Registrar(ctx, b.App, r.tienda, r.ruta, uso, r.cuando); err != nil {
			t.Fatal(err)
		}
	}

	corte, c, err := ia.Revisar(ctx, b.App, una, ia.Topes{Mes: 100, Dia: 2, Presupuesto: 40_000_000}, ahora)
	if err != nil {
		t.Fatal(err)
	}
	if c != (ia.Consumo{Mes: 3, Hoy: 2, SinIA: 1}) {
		t.Fatalf("consumo: %+v", c)
	}
	if corte != ia.CorteDia {
		t.Fatalf("corte: %q", corte)
	}
	// El gasto del mes de la plataforma suma las dos tiendas (4 respuestas del mes con IA).
	gastado, err := ia.Gastado(ctx, b.App, ahora)
	if err != nil {
		t.Fatal(err)
	}
	if gastado != 4*u.Costo() {
		t.Fatalf("gastado %d, quería %d", gastado, 4*u.Costo())
	}
	corte, _, _ = ia.Revisar(ctx, b.App, otra, ia.Topes{Mes: 100, Dia: 50, Presupuesto: gastado}, ahora)
	if corte != ia.CortePlataforma {
		t.Fatalf("tope de la plataforma: %q", corte)
	}
	// RLS: la otra tienda no ve el uso de la primera.
	_, c, _ = ia.Revisar(ctx, b.App, otra, ia.Topes{Mes: 100, Dia: 50}, ahora)
	if c.Mes != 1 {
		t.Fatalf("la otra tienda: %+v", c)
	}
}

// El cupo se reserva antes de llamar al modelo: dos preguntas a la vez no se pasan del tope, hay un
// máximo de preguntas en vuelo y liberar devuelve el lugar.
func TestReservarCupo(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	ahora := time.Now()
	topes := ia.Topes{Mes: 1, Dia: 25}
	if c, _, err := ia.Reservar(ctx, b.App, luna, topes, ahora); err != nil || c != ia.SinCorte {
		t.Fatalf("primera: %q %v", c, err)
	}
	// La primera todavía no terminó: con tope 1, la segunda ya no entra.
	if c, _, _ := ia.Reservar(ctx, b.App, luna, topes, ahora); c != ia.CortePlan {
		t.Fatalf("segunda en paralelo: %q", c)
	}
	_ = ia.Liberar(ctx, b.App, luna)
	topes.Mes = 100
	for i := 0; i < ia.EnVueloMax; i++ {
		if c, _, _ := ia.Reservar(ctx, b.App, luna, topes, ahora); c != ia.SinCorte {
			t.Fatalf("en vuelo %d: %q", i, c)
		}
	}
	if c, _, _ := ia.Reservar(ctx, b.App, luna, topes, ahora); c != ia.CorteOcupado {
		t.Fatalf("más de %d a la vez: %q", ia.EnVueloMax, c)
	}
	_ = ia.Liberar(ctx, b.App, luna)
	if c, _, _ := ia.Reservar(ctx, b.App, luna, topes, ahora); c != ia.SinCorte {
		t.Fatalf("después de liberar: %q", c)
	}
}
