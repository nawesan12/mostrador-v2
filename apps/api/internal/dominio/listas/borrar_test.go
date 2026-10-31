package listas_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func TestDuplicarYBorrar(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	id, _ := tiendas.Crear(ctx, b.App, "kiosco", "Kiosco", "almacen")
	otra, _ := tiendas.Crear(ctx, b.App, "otra", "Otra", "almacen")
	p, err := listas.Guardar(ctx, b.App, id, "", listas.Editable{Nombre: "Remera", Precio: 9000,
		Canales:   map[string]bool{"web": true},
		Opciones:  []listas.Opcion{{Nombre: "Talle", Valores: []string{"S", "M"}}},
		Variantes: []listas.VarianteEditable{{Valores: map[string]string{"Talle": "M"}, Stock: ptr(4)}}})
	if err != nil {
		t.Fatal(err)
	}

	c, err := listas.Duplicar(ctx, b.App, id, p.ID, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if c.ID == p.ID || c.Nombre != "Remera (copia)" || c.Publicado || c.Stock != 0 || len(c.Variantes) != 2 {
		t.Fatalf("copia: %+v", c)
	}

	// Otra tienda no puede borrar ni duplicar lo ajeno.
	if err := listas.Borrar(ctx, b.App, otra, p.ID); !errors.Is(err, listas.ErrNoExiste) {
		t.Fatalf("borró lo de otra tienda: %v", err)
	}
	if err := listas.Borrar(ctx, b.App, id, p.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := listas.Obtener(ctx, b.App, id, p.ID, time.Now()); !errors.Is(err, listas.ErrNoExiste) {
		t.Fatalf("sigue estando: %v", err)
	}
	if err := listas.Borrar(ctx, b.App, id, "no-es-un-id"); !errors.Is(err, listas.ErrNoExiste) {
		t.Fatalf("id inválido: %v", err)
	}
}
