package cuentas

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"html"
	"math/big"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/limite"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Volver a verificarse antes de cambiar adónde llega la plata (alias, CBU, titular, la cuenta de
// Mercado Pago). Una sesión robada o un «Recordarme» de 365 días no alcanzan: hace falta la
// contraseña, el código de la app autenticadora o un código que llega al mail de la cuenta (las
// cuentas que entran con Google, Apple o el link de WhatsApp no tienen contraseña). El permiso vale
// VidaReverificacion y sólo para la sesión que lo pidió.
const (
	VidaReverificacion = 5 * time.Minute
	vidaCodigoMail     = 10 * time.Minute
	// CabeceraReverificacion lleva el permiso en el pedido que lo necesita.
	CabeceraReverificacion = "Reverificacion"
)

// ErrReverificar: el cambio necesita volver a verificarse (o el permiso venció).
var ErrReverificar = errors.New("hace falta volver a verificarse")

// Metodos con los que la persona puede volver a verificarse. El código por mail está siempre.
type Metodos struct {
	Clave bool   `json:"clave"`
	TOTP  bool   `json:"totp"`
	Email string `json:"email"`
}

// MetodosReverificar dice qué puede usar la persona.
func (s *Servicio) MetodosReverificar(ctx context.Context, usuario string) (Metodos, error) {
	var m Metodos
	err := s.Pool.QueryRow(ctx, `SELECT clave_hash IS NOT NULL, totp_activo_en IS NOT NULL, email FROM auth.usuarios WHERE id = $1`,
		usuario).Scan(&m.Clave, &m.TOTP, &m.Email)
	m.Email = emailCorto(m.Email)
	return m, err
}

// emailCorto: «an•••@luna.com», para decir adónde llegó el código sin mostrarlo entero.
func emailCorto(e string) string {
	r := []rune(e)
	for i, c := range r {
		if c == '@' {
			if i <= 2 {
				return string(r[:1]) + "•••" + string(r[i:])
			}
			return string(r[:2]) + "•••" + string(r[i:])
		}
	}
	return e
}

func codigoSeis() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1_000_000))
	return fmt.Sprintf("%06d", n.Int64())
}

// PedirCodigoReverificar manda al mail de la cuenta un código de 6 números (vence en 10 minutos).
func (s *Servicio) PedirCodigoReverificar(ctx context.Context, d Datos) error {
	if d.SesionHash == "" {
		return ErrSinSesion
	}
	if ok, err := s.permitido(ctx, "reverificar:mail:"+d.Usuario.ID, 3, 15*time.Minute); err != nil || !ok {
		return errOr(err, ErrDemasiadosIntentos)
	}
	codigo := codigoSeis()
	if _, err := s.Pool.Exec(ctx, `INSERT INTO auth.reverificaciones (sesion_hash, codigo_hash, codigo_vence, intentos)
		VALUES ($1, $2, now() + $3::interval, 0)
		ON CONFLICT (sesion_hash) DO UPDATE SET codigo_hash = $2, codigo_vence = now() + $3::interval, intentos = 0`,
		d.SesionHash, hashCodigo(codigo), vidaCodigoMail.String()); err != nil {
		return err
	}
	if s.Correo == nil {
		return nil
	}
	return s.Correo.Enviar(ctx, mailCodigoReverificar(d.Usuario.Email, d.Usuario.Nombre, codigo))
}

// Reverificar comprueba la contraseña, el código de la app o el del mail y devuelve el permiso.
func (s *Servicio) Reverificar(ctx context.Context, d Datos, clave, codigo string) (string, time.Time, error) {
	if d.SesionHash == "" {
		return "", time.Time{}, ErrSinSesion
	}
	if ok, err := s.permitido(ctx, "reverificar:usuario:"+d.Usuario.ID, 10, 15*time.Minute); err != nil || !ok {
		return "", time.Time{}, errOr(err, ErrDemasiadosIntentos)
	}
	token := nuevoToken()
	var vence time.Time
	ok := false
	err := s.Pool.EnTx(ctx, func(tx pgx.Tx) error {
		var hash *string
		var totp bool
		if err := tx.QueryRow(ctx, `SELECT clave_hash, totp_activo_en IS NOT NULL FROM auth.usuarios WHERE id = $1`,
			d.Usuario.ID).Scan(&hash, &totp); err != nil {
			return err
		}
		switch {
		case clave != "":
			if hash != nil {
				v, err := VerificarClave(clave, *hash)
				if err != nil {
					return err
				}
				ok = v
			} else {
				_, _ = VerificarClave(clave, hashFalso)
			}
		case codigo != "":
			// El del mail (atado a esta sesión) o, si la tiene, el de la app autenticadora.
			var ch *string
			var intentos int
			err := tx.QueryRow(ctx, `SELECT codigo_hash, intentos FROM auth.reverificaciones
				WHERE sesion_hash = $1 AND codigo_vence > now() FOR UPDATE`, d.SesionHash).Scan(&ch, &intentos)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			if ch != nil && intentos < MaxIntentosCodigo && *ch == hashCodigo(soloNumeros(codigo)) {
				ok = true
				if _, err := tx.Exec(ctx, `UPDATE auth.reverificaciones SET codigo_hash = NULL WHERE sesion_hash = $1`, d.SesionHash); err != nil {
					return err
				}
			} else if totp {
				v, err := s.verificarCodigo(ctx, tx, d.Usuario.ID, codigo)
				if err != nil && !errors.Is(err, ErrNoActivo) {
					return err
				}
				ok = v
			}
			if !ok && ch != nil {
				if _, err := tx.Exec(ctx, `UPDATE auth.reverificaciones SET intentos = intentos + 1 WHERE sesion_hash = $1`, d.SesionHash); err != nil {
					return err
				}
			}
		}
		if !ok {
			return nil
		}
		return tx.QueryRow(ctx, `INSERT INTO auth.reverificaciones (sesion_hash, token_hash, token_vence)
			VALUES ($1, $2, now() + $3::interval)
			ON CONFLICT (sesion_hash) DO UPDATE SET token_hash = $2, token_vence = now() + $3::interval
			RETURNING token_vence`, d.SesionHash, hashToken(token), VidaReverificacion.String()).Scan(&vence)
	})
	if err != nil {
		return "", time.Time{}, err
	}
	if !ok {
		if clave != "" {
			return "", time.Time{}, ErrCredenciales
		}
		return "", time.Time{}, ErrCodigo
	}
	return token, vence, nil
}

// Reverificado dice si el permiso de la cabecera vale para esa sesión (corre en la transacción del
// cambio, así el permiso y el cambio van juntos).
func Reverificado(ctx context.Context, q limite.Consultor, sesionHash, token string) (bool, error) {
	if sesionHash == "" || token == "" || len(token) > 200 {
		return false, nil
	}
	var ok bool
	err := q.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM auth.reverificaciones
		WHERE sesion_hash = $1 AND token_hash = $2 AND token_vence > now())`, sesionHash, hashToken(token)).Scan(&ok)
	return ok, err
}

// ProblemaReverificar: el 403 que el panel y las apps reconocen para pedir la verificación.
func ProblemaReverificar() *problema.Problema {
	return &problema.Problema{Estado: http.StatusForbidden, Codigo: "reverificar", Titulo: "Confirmá que sos vos",
		Detalle: "Para cambiar adónde te llega la plata, volvé a verificarte."}
}

func (s *Servicio) rutasReverificar(mux *http.ServeMux) {
	persona := func(w http.ResponseWriter, r *http.Request) (Datos, bool) {
		d, ok := SesionDe(r)
		if !ok || d.SesionHash == "" || d.Mirada {
			problema.Escribir(w, problema.SinSesion())
			return d, false
		}
		return d, true
	}
	mux.HandleFunc("GET /v1/auth/reverificar", func(w http.ResponseWriter, r *http.Request) {
		d, ok := persona(w, r)
		if !ok {
			return
		}
		m, err := s.MetodosReverificar(r.Context(), d.Usuario.ID)
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusOK, m)
	})
	mux.HandleFunc("POST /v1/auth/reverificar/codigo", func(w http.ResponseWriter, r *http.Request) {
		d, ok := persona(w, r)
		if !ok {
			return
		}
		if err := s.PedirCodigoReverificar(r.Context(), d); err != nil {
			escribirError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("POST /v1/auth/reverificar", func(w http.ResponseWriter, r *http.Request) {
		d, ok := persona(w, r)
		if !ok {
			return
		}
		var a struct{ Clave, Codigo string }
		if !leer(w, r, &a) {
			return
		}
		t, vence, err := s.Reverificar(r.Context(), d, a.Clave, a.Codigo)
		switch {
		case errors.Is(err, ErrCredenciales):
			problema.Escribir(w, problema.Invalido("La contraseña no es correcta.",
				problema.Campo{Ruta: "clave", Mensaje: "La contraseña no es correcta."}))
		case errors.Is(err, ErrCodigo):
			problema.Escribir(w, problema.Invalido("El código no es correcto.",
				problema.Campo{Ruta: "codigo", Mensaje: "El código no es correcto."}))
		case err != nil:
			escribirError(w, err)
		default:
			escribir(w, http.StatusOK, map[string]any{"token": t, "venceEn": vence})
		}
	})
}

func mailCodigoReverificar(para, nombre, codigo string) correo.Mensaje {
	n := html.EscapeString(nombre)
	return correo.Mensaje{
		Para:   para,
		Asunto: "Tu código de Tu Local Web: " + codigo,
		Texto: fmt.Sprintf("Hola, %s.\n\nTu código para confirmar el cambio es %s. Vence en 10 minutos.\n\n"+
			"Si no fuiste vos, alguien entró a tu cuenta: cambiá tu contraseña y cerrá las otras sesiones desde Configuración › Seguridad.\n\nTu Local Web", nombre, codigo),
		HTML: fmt.Sprintf(`<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:'Instrument Sans',Helvetica,Arial,sans-serif;color:#2A140C">
<div style="max-width:560px;margin:0 auto;padding:32px 24px">
<div style="font-size:22px;font-weight:700;letter-spacing:-.03em">Tu Local Web</div>
<h1 style="margin:28px 0 0;font-size:28px;line-height:1.1;letter-spacing:-.03em;font-weight:600">Tu código</h1>
<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">Hola, %s. Usalo para confirmar el cambio. Vence en 10&nbsp;minutos.</p>
<p style="margin:24px 0;font-size:34px;font-weight:700;letter-spacing:.18em">%s</p>
<p style="margin:0;font-size:13.5px;line-height:1.5;color:#7A6252">Si no fuiste vos, alguien entró a tu cuenta: cambiá tu contraseña y cerrá las otras sesiones desde Configuración ›&nbsp;Seguridad.</p>
</div></body></html>`, n, codigo),
	}
}
