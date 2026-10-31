package configuracion_test

import (
	"context"
	"errors"
	"strings"
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
		_, err := tx.Exec(ctx, `INSERT INTO auth.sesiones (token_hash, usuario_id, tienda_id, vence_en, agente, ip, visto_en) VALUES
			('aaaaaaaaaaaaaaaa1111', $1, $2, now() + interval '1 day', 'Mozilla/5.0 (Macintosh; Intel Mac OS X) Chrome/130', '190.210.32.1', now() - interval '2 hours'),
			('bbbbbbbbbbbbbbbb2222', $1, $2, now() + interval '1 day', 'Mozilla/5.0 (iPhone) Safari', '::1', now()),
			('cccccccccccccccc3333', $1, $2, now() + interval '1 day', 'Mozilla/5.0 (X11; Linux x86_64) Firefox/131', '2800:810:4a2:1f00::1', now() - interval '10 days'),
			('dddddddddddddddd4444', $1, $2, now() - interval '1 day', 'Mozilla/5.0 (iPad) Safari', '8.8.8.8', now() - interval '1 day')`, usuario, id)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	q := configuracion.Quien{UsuarioID: usuario, SesionID: "aaaaaaaaaaaaaaaa", Rol: "duenio"}
	geo := lugares{"190.210.32.1": "Ciudad de Buenos Aires"}
	c, err := configuracion.Leer(ctx, b.App, geo, id, q, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	eq := c["equipo"].([]configuracion.Miembro)
	if len(eq) != 1 || !eq[0].Vos || eq[0].Rol != "duenio" {
		t.Fatalf("equipo: %+v", eq)
	}
	ses := c["sesiones"].([]configuracion.Sesion)
	if len(ses) != 3 {
		t.Fatalf("todas las vigentes (la vencida no): %+v", ses)
	}
	if s := ses[0]; s.ID != "aaaaaaaaaaaaaaaa" || !s.Actual || s.Nombre != "Chrome en Mac" || s.Tipo != "compu" ||
		s.Ubicacion != "Ciudad de Buenos Aires" || s.IP != "190.210.32.1" || s.Creada.IsZero() {
		t.Fatalf("la actual va primero, con lugar e IP: %+v", s)
	}
	if s := ses[1]; s.ID != "bbbbbbbbbbbbbbbb" || s.Nombre != "Safari en iPhone" || s.Tipo != "celu" || s.Ubicacion != "" {
		t.Fatalf("después, la más reciente; sin lugar para ::1: %+v", s)
	}
	if s := ses[2]; s.Nombre != "Firefox en Linux" || s.IP != "2800:810:4a2:1f00:…" {
		t.Fatalf("IPv6 abreviada: %+v", s)
	}
	if seg := c["seguridad"].(map[string]bool); len(seg) != 2 || !seg["avisoIngreso"] || seg["inactivas"] {
		t.Fatalf("seguridad por defecto: %+v", seg)
	}
	if err := configuracion.FijarOpcion(ctx, b.App, id, q, "seguridad", "dobleFactor", true); !errors.Is(err, configuracion.ErrNoExiste) {
		t.Fatalf("una opción que no existe: %v", err)
	}

	t.Run("prender inactivas cierra las que no se usan hace 7 días", func(t *testing.T) {
		if err := configuracion.FijarOpcion(ctx, b.App, id, q, "seguridad", "inactivas", true); err != nil {
			t.Fatal(err)
		}
		c, _ := configuracion.Leer(ctx, b.App, nil, id, q, time.Now())
		if ses := c["sesiones"].([]configuracion.Sesion); len(ses) != 2 || !c["seguridad"].(map[string]bool)["inactivas"] {
			t.Fatalf("después: %+v %+v", ses, c["seguridad"])
		}
		var n int
		_ = b.App.QueryRow(ctx, `SELECT count(*) FROM auth.sesiones WHERE token_hash LIKE 'cccc%'`).Scan(&n)
		if n != 0 {
			t.Fatal("la inactiva sigue en la base")
		}
	})
	if c["plan"].(map[string]any)["enPrueba"] != true {
		t.Fatal("una tienda nueva está en prueba")
	}

	if err := configuracion.CerrarSesion(ctx, b.App, id, q, "aaaaaaaaaaaaaaaa"); !errors.Is(err, configuracion.ErrNoSePuede) {
		t.Fatalf("cerró la sesión actual: %v", err)
	}
	if err := configuracion.CerrarSesion(ctx, b.App, id, q, "bbbbbbbbbbbbbbbb"); err != nil {
		t.Fatal(err)
	}
	if _, err := configuracion.Invitar(ctx, b.App, "https://tulocalweb.test", id, q, "vale@luna.com", "ventas"); !errors.Is(err, configuracion.ErrRepetido) {
		t.Fatalf("invitó a alguien del equipo: %v", err)
	}
	if _, err := configuracion.Invitar(ctx, b.App, "https://tulocalweb.test", id, q, "juli@mail.com", "local"); err != nil {
		t.Fatal(err)
	}
	if err := configuracion.GuardarNegocio(ctx, b.App, id, configuracion.Negocio{Nombre: "Luna Deco", Horario: "9 a 18"}); err != nil {
		t.Fatal(err)
	}
	if err := configuracion.FijarOpcion(ctx, b.App, id, q, "avisos", "novedades", true); err != nil {
		t.Fatal(err)
	}
	if err := configuracion.FijarOpcion(ctx, b.App, id, q, "avisos", "nada", true); !errors.Is(err, configuracion.ErrNoExiste) {
		t.Fatal(err)
	}
	if err := configuracion.CambiarPlan(ctx, b.App, id, "emprende", "anio"); err != nil {
		t.Fatal(err)
	}
	c, _ = configuracion.Leer(ctx, b.App, nil, id, q, time.Now())
	n := c["negocio"].(configuracion.Negocio)
	if n.Nombre != "Luna Deco" || n.Horario != "9 a 18" || !c["avisos"].(map[string]bool)["novedades"] ||
		len(c["sesiones"].([]configuracion.Sesion)) != 1 || len(c["equipo"].([]configuracion.Miembro)) != 2 ||
		c["plan"].(map[string]any)["actual"] != "emprende" {
		t.Fatalf("después: %+v", c)
	}
}

type lugares map[string]string

func (l lugares) Ubicar(ip string) string { return l[ip] }

func TestCerrarTodasLasDemas(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	id := pruebadb.CrearTienda(t, b.App, "luna")
	var usuario, otro string
	err := b.App.EnTx(ctx, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre) VALUES ('vale@luna.com', 'Valeria') RETURNING id::text`).Scan(&usuario); err != nil {
			return err
		}
		if err := tx.QueryRow(ctx, `INSERT INTO auth.usuarios (email, nombre) VALUES ('juli@luna.com', 'Julia') RETURNING id::text`).Scan(&otro); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `INSERT INTO auth.sesiones (token_hash, usuario_id, vence_en) VALUES
			('aaaaaaaaaaaaaaaa1111', $1, now() + interval '1 day'), ('bbbbbbbbbbbbbbbb2222', $1, now() + interval '1 day'),
			('cccccccccccccccc3333', $1, now() + interval '1 day'), ('dddddddddddddddd4444', $2, now() + interval '1 day')`, usuario, otro)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	n, err := configuracion.CerrarOtras(ctx, b.App, id, configuracion.Quien{UsuarioID: usuario, SesionID: "aaaaaaaaaaaaaaaa"})
	if err != nil || n != 2 {
		t.Fatalf("cerradas = %d, err %v", n, err)
	}
	var quedan []string
	filas, _ := b.App.Query(ctx, `SELECT left(token_hash, 4) FROM auth.sesiones ORDER BY 1`)
	for filas.Next() {
		var s string
		_ = filas.Scan(&s)
		quedan = append(quedan, s)
	}
	if strings.Join(quedan, ",") != "aaaa,dddd" {
		t.Fatalf("quedan %v: tiene que seguir la actual y la de otra persona", quedan)
	}
}
