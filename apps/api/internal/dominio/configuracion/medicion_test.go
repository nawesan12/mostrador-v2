package configuracion_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func ptr(s string) *string { return &s }

func TestMedicion(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")

	m, err := configuracion.LeerMedicion(ctx, b.App, luna)
	if err != nil || m.GA4 != nil || m.Pixel != nil || m.Feeds.GM || m.Feeds.Meta {
		t.Fatalf("una tienda nueva arranca sin medición: %+v %v", m, err)
	}

	// Se guarda limpio (sin espacios, GA4 en mayúsculas).
	err = configuracion.GuardarMedicion(ctx, b.App, luna, configuracion.Medicion{
		GA4: ptr(" g-ab12cd34ef "), Pixel: ptr("123456789012345"), Feeds: configuracion.Feeds{GM: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	m, err = configuracion.LeerMedicion(ctx, b.App, luna)
	if err != nil || m.GA4 == nil || *m.GA4 != "G-AB12CD34EF" || m.Pixel == nil || *m.Pixel != "123456789012345" ||
		!m.Feeds.GM || m.Feeds.Meta {
		t.Fatalf("leída: %+v %v", m, err)
	}

	// IDs inválidos: no se guarda nada.
	for _, malo := range []configuracion.Medicion{
		{GA4: ptr("UA-1234-1")},
		{GA4: ptr("G-<script>")},
		{Pixel: ptr("12345")},
		{Pixel: ptr("12345678901234a")},
	} {
		if err := configuracion.GuardarMedicion(ctx, b.App, luna, malo); !errors.Is(err, configuracion.ErrInvalido) {
			t.Fatalf("aceptó %+v: %v", malo, err)
		}
	}

	// Aislamiento: la otra tienda no ve ni pisa lo de luna.
	if m, _ := configuracion.LeerMedicion(ctx, b.App, sol); m.GA4 != nil || m.Pixel != nil || m.Feeds.GM {
		t.Fatalf("sol ve la medición de luna: %+v", m)
	}
	if err := configuracion.GuardarMedicion(ctx, b.App, sol, configuracion.Medicion{Pixel: ptr("9999999999999999")}); err != nil {
		t.Fatal(err)
	}
	pub, err := vidriera.Publica(ctx, b.App, "luna")
	if err != nil || pub.Medicion.GA4 == nil || *pub.Medicion.GA4 != "G-AB12CD34EF" || *pub.Medicion.Pixel != "123456789012345" {
		t.Fatalf("la tienda online de luna: %+v %v", pub.Medicion, err)
	}
	pub, _ = vidriera.Publica(ctx, b.App, "sol")
	if pub.Medicion.GA4 != nil || pub.Medicion.Pixel == nil || *pub.Medicion.Pixel != "9999999999999999" {
		t.Fatalf("la tienda online de sol: %+v", pub.Medicion)
	}

	// Vacío borra.
	if err := configuracion.GuardarMedicion(ctx, b.App, luna, configuracion.Medicion{GA4: ptr(""), Pixel: nil}); err != nil {
		t.Fatal(err)
	}
	if m, _ := configuracion.LeerMedicion(ctx, b.App, luna); m.GA4 != nil || m.Pixel != nil || m.Feeds.GM {
		t.Fatalf("no se borró: %+v", m)
	}
}
