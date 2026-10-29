package configuracion_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

type dnsFalso struct {
	cname string
	txt   []string
}

func (d dnsFalso) LookupCNAME(context.Context, string) (string, error) { return d.cname, nil }
func (d dnsFalso) LookupTXT(context.Context, string) ([]string, error) { return d.txt, nil }

func TestDominiosYDatos(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	sitio := "https://tulocalweb.com.ar"

	if err := configuracion.ConectarDominio(ctx, b.App, luna, "Luna.com.ar"); err != nil {
		t.Fatal(err)
	}
	if err := configuracion.ConectarDominio(ctx, b.App, sol, "luna.com.ar"); !errors.Is(err, configuracion.ErrRepetido) {
		t.Fatalf("otra tienda tomó el mismo dominio: %v", err)
	}
	if err := configuracion.ConectarDominio(ctx, b.App, luna, "no es dominio"); !errors.Is(err, configuracion.ErrInvalido) {
		t.Fatal(err)
	}
	configuracion.Resolver = dnsFalso{}
	d, _ := configuracion.Dominios(ctx, b.App, luna, sitio)
	ds := d["dominios"].([]configuracion.Dominio)
	if len(ds) != 2 || ds[1].Estado != "verificando" || !ds[0].Principal {
		t.Fatalf("antes del DNS: %+v", ds)
	}
	if err := configuracion.DominioPrincipal(ctx, b.App, luna, "luna.com.ar", sitio); !errors.Is(err, configuracion.ErrNoSePuede) {
		t.Fatalf("hizo principal uno sin verificar: %v", err)
	}
	configuracion.Resolver = dnsFalso{cname: "tiendas.tulocalweb.com.ar.", txt: []string{"mostrador-verificacion=luna"}}
	d, _ = configuracion.Dominios(ctx, b.App, luna, sitio)
	if ds = d["dominios"].([]configuracion.Dominio); ds[1].Estado != "activo" {
		t.Fatalf("con el DNS bien: %+v", ds)
	}
	if err := configuracion.DominioPrincipal(ctx, b.App, luna, "luna.com.ar", sitio); err != nil {
		t.Fatal(err)
	}
	d, _ = configuracion.Dominios(ctx, b.App, luna, sitio)
	if ds = d["dominios"].([]configuracion.Dominio); ds[0].Principal || !ds[1].Principal {
		t.Fatalf("principal: %+v", ds)
	}

	f, err := configuracion.Exportar(ctx, b.App, luna, "productos")
	if err != nil || len(f) != 1 || f[0][0] != "SKU" {
		t.Fatalf("exportar: %v %v", f, err)
	}
	if err := configuracion.CerrarCuenta(ctx, b.App, luna, "otra"); !errors.Is(err, configuracion.ErrInvalido) {
		t.Fatal("cerró sin confirmar bien")
	}
	if err := configuracion.CerrarCuenta(ctx, b.App, luna, "tienda LUNA"); err != nil {
		t.Fatal(err)
	}
}
