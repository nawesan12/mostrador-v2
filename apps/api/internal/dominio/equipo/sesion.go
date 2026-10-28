package equipo

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/limite"
)

// Duraciones de las sesiones del equipo.
const (
	VidaSesion = 12 * time.Hour
	VidaMirada = 30 * time.Minute
)

// SesionNueva es el token que se le da al BFF una sola vez (components/schemas/SesionNueva).
type SesionNueva struct {
	Token   string    `json:"token"`
	VenceEn time.Time `json:"venceEn"`
}

// hashFalso: el mismo tiempo cuando el email no existe, para no delatar quién es del equipo.
var hashFalso, _ = cuentas.HashClave("una clave que nadie del equipo tiene")

// Ingresar con email y contraseña. Un email que no es del equipo, una clave incorrecta y una
// persona dada de baja responden igual.
func (s *Servicio) Ingresar(ctx context.Context, email, clave string, cl cuentas.Cliente) (SesionNueva, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" || !strings.Contains(email, "@") || len(email) > 254 {
		return SesionNueva{}, ErrInvalido{"email", "Escribí tu email."}
	}
	if clave == "" {
		return SesionNueva{}, ErrInvalido{"clave", "Escribí tu contraseña."}
	}
	for _, l := range []struct {
		clave string
		max   int
	}{{"equipo:email:" + email, 5}, {"equipo:ip:" + cl.IP, 30}} {
		ok, err := limite.Permitido(ctx, s.App, l.clave, l.max, 15*time.Minute)
		if err != nil {
			return SesionNueva{}, err
		}
		if !ok {
			return SesionNueva{}, ErrIntentos
		}
	}
	var m Miembro
	var hash string
	var activo bool
	err := s.Staff.QueryRow(ctx, `SELECT id::text, nombre, email::text, rol, clave_hash, activo FROM equipo.miembros WHERE email = $1`,
		email).Scan(&m.ID, &m.Nombre, &m.Email, &m.Rol, &hash, &activo)
	if errors.Is(err, pgx.ErrNoRows) {
		_, _ = cuentas.VerificarClave(clave, hashFalso)
		return SesionNueva{}, ErrCredenciales
	}
	if err != nil {
		return SesionNueva{}, err
	}
	ok, err := cuentas.VerificarClave(clave, hash)
	if err != nil {
		return SesionNueva{}, err
	}
	if !ok || !activo {
		return SesionNueva{}, ErrCredenciales
	}
	_, _ = s.App.Exec(ctx, `DELETE FROM auth.intentos WHERE clave = $1`, "equipo:email:"+email)
	token := nuevoToken()
	var ses SesionNueva
	err = s.Staff.EnTx(ctx, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `INSERT INTO equipo.sesiones (token_hash, miembro_id, vence_en, ip, agente)
			VALUES ($1, $2, now() + $3::interval, nullif($4, '')::inet, nullif($5, '')) RETURNING vence_en`,
			hashToken(token), m.ID, VidaSesion.String(), cl.IP, cl.Agente).Scan(&ses.VenceEn); err != nil {
			return err
		}
		return anotar(ctx, tx, m, cl, "ingreso", "Ingresó al backoffice", "", nil)
	})
	ses.Token = token
	return ses, err
}

// Sesion resuelve el token del equipo. Nunca mira auth.sesiones: un token de una tienda no es del
// equipo.
func (s *Servicio) Sesion(ctx context.Context, token string) (Miembro, error) {
	var m Miembro
	if token == "" {
		return m, ErrNoExiste
	}
	var visto time.Time
	err := s.Staff.QueryRow(ctx, `
		SELECT m.id::text, m.nombre, m.email::text, m.rol, s.vence_en, s.visto_en
		FROM equipo.sesiones s JOIN equipo.miembros m ON m.id = s.miembro_id
		WHERE s.token_hash = $1 AND s.vence_en > now() AND m.activo`, hashToken(token)).
		Scan(&m.ID, &m.Nombre, &m.Email, &m.Rol, &m.VenceEn, &visto)
	if errors.Is(err, pgx.ErrNoRows) {
		return m, ErrNoExiste
	}
	if err != nil {
		return m, err
	}
	if time.Since(visto) > time.Hour {
		_, _ = s.Staff.Exec(ctx, `UPDATE equipo.sesiones SET visto_en = now() WHERE token_hash = $1`, hashToken(token))
	}
	return m, nil
}

// Salir cierra la sesión del equipo.
func (s *Servicio) Salir(ctx context.Context, token string) error {
	_, err := s.Staff.Exec(ctx, `DELETE FROM equipo.sesiones WHERE token_hash = $1`, hashToken(token))
	return err
}

// Mirar abre una mirada de sólo lectura al panel de la tienda (30 min) y lo anota.
func (s *Servicio) Mirar(ctx context.Context, m Miembro, tiendaID string, cl cuentas.Cliente) (SesionNueva, error) {
	token := nuevoToken()
	var ses SesionNueva
	err := s.Staff.EnTx(ctx, func(tx pgx.Tx) error {
		var nombre string
		err := tx.QueryRow(ctx, `SELECT nombre FROM tiendas WHERE id::text = $1 AND cerrada_en IS NULL`, tiendaID).Scan(&nombre)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExiste
		}
		if err != nil {
			return err
		}
		if err := tx.QueryRow(ctx, `INSERT INTO equipo.miradas (token_hash, miembro_id, tienda_id, vence_en)
			VALUES ($1, $2, $3::uuid, now() + $4::interval) RETURNING vence_en`,
			hashToken(token), m.ID, tiendaID, VidaMirada.String()).Scan(&ses.VenceEn); err != nil {
			return err
		}
		return anotar(ctx, tx, m, cl, "mirar_tienda", "Entró a ver la tienda "+nombre, tiendaID, nil)
	})
	ses.Token = token
	return ses, err
}

// ParaCuentas resuelve un token de mirada como una sesión del panel de esa tienda, de sólo lectura
// (cuentas.Servicio.SesionEquipo). Las escrituras las frena SoloLectura.
func (s *Servicio) ParaCuentas(ctx context.Context, token string) (cuentas.Datos, error) {
	var d cuentas.Datos
	var nombre string
	err := s.Staff.QueryRow(ctx, `
		SELECT mi.tienda_id::text, m.id::text, m.nombre, m.email::text, mi.vence_en
		FROM equipo.miradas mi JOIN equipo.miembros m ON m.id = mi.miembro_id
		WHERE mi.token_hash = $1 AND mi.vence_en > now() AND m.activo`, hashToken(token)).
		Scan(&d.TiendaID, &d.Usuario.ID, &nombre, &d.Usuario.Email, &d.VenceEn)
	if errors.Is(err, pgx.ErrNoRows) {
		return d, cuentas.ErrSinSesion
	}
	if err != nil {
		return d, err
	}
	d.Usuario.Nombre = nombre + " (equipo)"
	d.Rol = "duenio"
	d.Mirada = true
	d.SesionID = "equipo-" + hashToken(token)[:8]
	return d, nil
}
