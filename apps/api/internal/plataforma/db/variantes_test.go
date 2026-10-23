package db_test

import (
	"context"
	"database/sql"
	"io/fs"
	"net/url"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pressly/goose/v3"

	esquema "github.com/nawesan12/mostrador-v2/apps/api/db"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// La migración de variantes arranca la galería con la foto que ya tenía cada producto (aunque RLS
// esté forzada para el migrador).
func TestMigracionVariantesRellenaFotos(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	u, _ := url.Parse(b.AdminURL)
	u.User = url.User("mostrador_migrador")
	conn, err := sql.Open("pgx", u.String())
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	archivos, _ := fs.Sub(esquema.Migraciones, "migraciones")
	prov, err := goose.NewProvider(goose.DialectPostgres, conn, archivos)
	if err != nil {
		t.Fatal(err)
	}
	// Volver a antes de 00033 (y de lo que venga después), cargar productos y migrar de nuevo.
	if _, err := prov.DownTo(ctx, 32); err != nil {
		t.Fatal(err)
	}
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	if err := b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO productos (tienda_id, nombre, precio, foto) VALUES ($1, 'Con foto', 100, 'https://f/a.jpg'), ($1, 'Sin foto', 100, NULL)`, luna)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := prov.Up(ctx); err != nil {
		t.Fatal(err)
	}
	var con, sin []string
	if err := b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `SELECT fotos FROM productos WHERE nombre = 'Con foto'`).Scan(&con); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `SELECT fotos FROM productos WHERE nombre = 'Sin foto'`).Scan(&sin)
	}); err != nil {
		t.Fatal(err)
	}
	if len(con) != 1 || con[0] != "https://f/a.jpg" || len(sin) != 0 {
		t.Fatalf("fotos: %v %v", con, sin)
	}
}
