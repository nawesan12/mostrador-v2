package mitienda_test

import (
	"context"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mitienda"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func TestBorradorYPublicado(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	id := pruebadb.CrearTienda(t, b.App, "luna")
	v, err := mitienda.Leer(ctx, b.App, id, "https://tulocalweb.com.ar")
	if err != nil || v["sinPublicar"] != false || v["direccion"] != "tulocalweb.com.ar/t/luna" {
		t.Fatalf("inicial: %+v %v", v, err)
	}
	d := mitienda.Inicial()
	d.Secciones[0], d.Secciones[1] = d.Secciones[1], d.Secciones[0]
	d.Secciones[5].Visible = false
	d.Paleta = 3
	if err := mitienda.Guardar(ctx, b.App, id, d); err != nil {
		t.Fatal(err)
	}
	if v, _ = mitienda.Leer(ctx, b.App, id, ""); v["sinPublicar"] != true {
		t.Fatal("no quedó sin publicar")
	}
	if err := mitienda.Publicar(ctx, b.App, id); err != nil {
		t.Fatal(err)
	}
	if v, _ = mitienda.Leer(ctx, b.App, id, ""); v["sinPublicar"] != false {
		t.Fatal("no se publicó")
	}
	malo := mitienda.Inicial()
	malo.Secciones = malo.Secciones[:3]
	if err := mitienda.Guardar(ctx, b.App, id, malo); err == nil {
		t.Fatal("aceptó un diseño sin todas las secciones")
	}
}
