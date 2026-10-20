package db_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// Toda tabla del esquema public tiene RLS habilitada y forzada, y si tiene tienda_id, la política
// por tienda. Una tabla nueva sin aislamiento rompe este test (y el CI).
func TestTodaTablaTieneAislamiento(t *testing.T) {
	b := pruebadb.Nueva(t)
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close(ctx)

	filas, err := conn.Query(ctx, `
		SELECT c.relname, c.relrowsecurity, c.relforcerowsecurity,
		       EXISTS (SELECT 1 FROM pg_attribute a WHERE a.attrelid = c.oid AND a.attname = 'tienda_id' AND NOT a.attisdropped),
		       EXISTS (SELECT 1 FROM pg_policy p WHERE p.polrelid = c.oid AND p.polname = 'tienda')
		FROM pg_class c JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname = 'public' AND c.relkind IN ('r', 'p')
		  AND c.relname <> 'goose_db_version'`)
	if err != nil {
		t.Fatal(err)
	}
	defer filas.Close()
	vistas := 0
	for filas.Next() {
		var tabla string
		var rls, forzada, conTienda, politica bool
		if err := filas.Scan(&tabla, &rls, &forzada, &conTienda, &politica); err != nil {
			t.Fatal(err)
		}
		vistas++
		if !rls || !forzada {
			t.Errorf("%s: RLS habilitada=%v forzada=%v; usá SELECT app.aplicar_rls('%s')", tabla, rls, forzada, tabla)
		}
		if conTienda && !politica {
			t.Errorf("%s tiene tienda_id y no tiene la política 'tienda'", tabla)
		}
	}
	if vistas < 4 {
		t.Fatalf("el catálogo devolvió %d tablas: la consulta está mal", vistas)
	}
}

// La app no puede saltear RLS ni ser dueña de nada.
func TestElRolDeLaAppNoTienePrivilegios(t *testing.T) {
	b := pruebadb.Nueva(t)
	ctx := context.Background()
	var super, bypass bool
	var duenias int
	err := b.App.QueryRow(ctx, `
		SELECT r.rolsuper, r.rolbypassrls,
		       (SELECT count(*) FROM pg_class c WHERE c.relowner = r.oid)
		FROM pg_roles r WHERE r.rolname = current_user`).Scan(&super, &bypass, &duenias)
	if err != nil {
		t.Fatal(err)
	}
	if super || bypass || duenias > 0 {
		t.Fatalf("mostrador_app: super=%v bypassrls=%v dueña de %d tablas", super, bypass, duenias)
	}
}

func TestDosTiendasNoSeCruzan(t *testing.T) {
	b := pruebadb.Nueva(t)
	ctx := context.Background()
	a := pruebadb.CrearTienda(t, b.App, "luna")
	o := pruebadb.CrearTienda(t, b.App, "sol")

	anotar := func(tienda, accion string) error {
		return b.App.EnTienda(ctx, tienda, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `INSERT INTO auditoria (tienda_id, actor_tipo, accion, entidad) VALUES ($1, 'sistema', $2, 'prueba')`, tienda, accion)
			return err
		})
	}
	if err := anotar(a, "de-luna"); err != nil {
		t.Fatal(err)
	}
	if err := anotar(o, "de-sol"); err != nil {
		t.Fatal(err)
	}

	contar := func(tienda, tabla string) (n int) {
		t.Helper()
		err := b.App.EnTienda(ctx, tienda, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, "SELECT count(*) FROM "+tabla).Scan(&n)
		})
		if err != nil {
			t.Fatal(err)
		}
		return n
	}

	t.Run("cada una ve sólo lo suyo", func(t *testing.T) {
		if n := contar(a, "auditoria"); n != 1 {
			t.Errorf("luna ve %d filas de auditoría, quería 1", n)
		}
		if n := contar(a, "tiendas"); n != 1 {
			t.Errorf("luna ve %d tiendas, quería 1", n)
		}
	})

	t.Run("no puede escribir en otra", func(t *testing.T) {
		err := b.App.EnTienda(ctx, a, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `INSERT INTO auditoria (tienda_id, actor_tipo, accion, entidad) VALUES ($1, 'sistema', 'intruso', 'prueba')`, o)
			return err
		})
		var pg *pgconn.PgError
		if !errors.As(err, &pg) || pg.Code != "42501" {
			t.Fatalf("insertar en otra tienda: err=%v, quería violación de RLS (42501)", err)
		}
	})

	t.Run("no puede tocar otra", func(t *testing.T) {
		var tocadas int64
		err := b.App.EnTienda(ctx, a, func(tx pgx.Tx) error {
			tag, err := tx.Exec(ctx, "UPDATE tiendas SET nombre = 'hackeada' WHERE id = $1", o)
			tocadas = tag.RowsAffected()
			return err
		})
		if err != nil || tocadas != 0 {
			t.Fatalf("update a otra tienda: err=%v filas=%d", err, tocadas)
		}
	})

	t.Run("sin tienda no ve nada", func(t *testing.T) {
		var n int
		err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
			return tx.QueryRow(ctx, "SELECT (SELECT count(*) FROM auditoria) + (SELECT count(*) FROM tiendas)").Scan(&n)
		})
		if err != nil || n != 0 {
			t.Fatalf("sin tienda: err=%v filas=%d, quería 0", err, n)
		}
		if err := b.App.EnTienda(ctx, "", func(pgx.Tx) error { return nil }); err == nil {
			t.Fatal("EnTienda sin tienda tendría que fallar")
		}
	})

	t.Run("staff lee todo pero no escribe", func(t *testing.T) {
		var n int
		if err := b.Staff.QueryRow(ctx, "SELECT count(*) FROM auditoria").Scan(&n); err != nil || n != 2 {
			t.Fatalf("staff: err=%v filas=%d, quería 2", err, n)
		}
		_, err := b.Staff.Exec(ctx, "DELETE FROM auditoria")
		if err == nil {
			t.Fatal("staff pudo borrar auditoría")
		}
	})
}
