package configuracion_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func TestCuentaPermisosYEquipo(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	id := pruebadb.CrearTienda(t, b.App, "sol")
	var duenio, otro string
	if err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre) VALUES ('a@sol.com', 'Ana') RETURNING id::text`).Scan(&duenio); err != nil {
			return err
		}
		if err := tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre) VALUES ('b@sol.com', 'Beto') RETURNING id::text`).Scan(&otro); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO auth.membresias (usuario_id, tienda_id, rol) VALUES ($1, $3, 'duenio'), ($2, $3, 'ventas')`, duenio, otro, id)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	q := configuracion.Quien{UsuarioID: duenio}

	if err := configuracion.FijarPermiso(ctx, b.App, id, "finanzas", "ventas", true); err != nil {
		t.Fatal(err)
	}
	if err := configuracion.FijarPermiso(ctx, b.App, id, "cuenta", "admin", true); !errors.Is(err, configuracion.ErrNoSePuede) {
		t.Fatalf("tocó un permiso del dueño: %v", err)
	}
	ps, _ := configuracion.Permisos(ctx, b.App, id)
	for _, p := range ps {
		if p["id"] == "finanzas" && !p["roles"].(map[string]bool)["ventas"] {
			t.Fatal("no quedó el permiso")
		}
	}
	if err := configuracion.CambiarRol(ctx, b.App, id, q, duenio, "ventas"); !errors.Is(err, configuracion.ErrNoSePuede) {
		t.Fatalf("se cambió el rol a sí misma: %v", err)
	}
	if err := configuracion.CambiarRol(ctx, b.App, id, q, otro, "deposito"); err != nil {
		t.Fatal(err)
	}
	if err := configuracion.Quitar(ctx, b.App, id, q, otro); err != nil {
		t.Fatal(err)
	}
	if err := configuracion.Quitar(ctx, b.App, id, q, otro); !errors.Is(err, configuracion.ErrNoExiste) {
		t.Fatalf("quitó dos veces: %v", err)
	}
	p, err := configuracion.Plan(ctx, b.App, id, time.Now())
	if err != nil || p["enPrueba"] != true || p["uso"].(map[string]any)["personas"] != 1 {
		t.Fatalf("plan: %+v %v", p, err)
	}
}
