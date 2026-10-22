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

func TestConfiguracion(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	id := pruebadb.CrearTienda(t, b.App, "luna")
	var usuario string
	err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre) VALUES ('vale@luna.com', 'Valeria') RETURNING id::text`).Scan(&usuario); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO auth.membresias (usuario_id, tienda_id, rol) VALUES ($1, $2, 'duenio')`, usuario, id); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO auth.sesiones (token_hash, usuario_id, tienda_id, vence_en, agente) VALUES
			('aaaaaaaaaaaaaaaa1111', $1, $2, now() + interval '1 day', 'Mozilla/5.0 (Macintosh; Intel Mac OS X) Chrome/130'),
			('bbbbbbbbbbbbbbbb2222', $1, $2, now() + interval '1 day', 'Mozilla/5.0 (iPhone) Safari')`, usuario, id)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	q := configuracion.Quien{UsuarioID: usuario, SesionID: "aaaaaaaaaaaaaaaa"}
	c, err := configuracion.Leer(ctx, b.App, id, q, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	eq := c["equipo"].([]configuracion.Miembro)
	if len(eq) != 1 || !eq[0].Vos || eq[0].Rol != "duenio" {
		t.Fatalf("equipo: %+v", eq)
	}
	ses := c["sesiones"].([]configuracion.Sesion)
	if len(ses) != 2 {
		t.Fatalf("sesiones: %+v", ses)
	}
	for _, s := range ses {
		if s.ID == "aaaaaaaaaaaaaaaa" && (!s.Actual || s.Nombre != "Chrome en Mac") {
			t.Fatalf("sesión actual: %+v", s)
		}
	}
	if c["plan"].(map[string]any)["enPrueba"] != true {
		t.Fatal("una tienda nueva está en prueba")
	}

	if err := configuracion.CerrarSesion(ctx, b.App, id, q, "aaaaaaaaaaaaaaaa"); !errors.Is(err, configuracion.ErrNoSePuede) {
		t.Fatalf("cerró la sesión actual: %v", err)
	}
	if err := configuracion.CerrarSesion(ctx, b.App, id, q, "bbbbbbbbbbbbbbbb"); err != nil {
		t.Fatal(err)
	}
	if _, err := configuracion.Invitar(ctx, b.App, id, "vale@luna.com", "ventas"); !errors.Is(err, configuracion.ErrRepetido) {
		t.Fatalf("invitó a alguien del equipo: %v", err)
	}
	if _, err := configuracion.Invitar(ctx, b.App, id, "juli@mail.com", "deposito"); err != nil {
		t.Fatal(err)
	}
	if err := configuracion.GuardarNegocio(ctx, b.App, id, configuracion.Negocio{Nombre: "Luna Deco", Horario: "9 a 18"}); err != nil {
		t.Fatal(err)
	}
	if err := configuracion.FijarOpcion(ctx, b.App, id, "avisos", "novedades", true); err != nil {
		t.Fatal(err)
	}
	if err := configuracion.FijarOpcion(ctx, b.App, id, "avisos", "nada", true); !errors.Is(err, configuracion.ErrNoExiste) {
		t.Fatal(err)
	}
	if err := configuracion.CambiarPlan(ctx, b.App, id, "emprende", "anio"); err != nil {
		t.Fatal(err)
	}
	c, _ = configuracion.Leer(ctx, b.App, id, q, time.Now())
	n := c["negocio"].(configuracion.Negocio)
	if n.Nombre != "Luna Deco" || n.Horario != "9 a 18" || !c["avisos"].(map[string]bool)["novedades"] ||
		len(c["sesiones"].([]configuracion.Sesion)) != 1 || len(c["equipo"].([]configuracion.Miembro)) != 2 ||
		c["plan"].(map[string]any)["actual"] != "emprende" {
		t.Fatalf("después: %+v", c)
	}
}
