package cuentas

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/whatsapp"
)

// VidaEnlace: lo que dura un link para entrar por WhatsApp.
const VidaEnlace = 15 * time.Minute

// ErrSinWhatsApp: la plataforma no tiene WhatsApp configurado.
var ErrSinWhatsApp = errors.New("whatsapp sin configurar")

func telefono(v string) (string, error) {
	t := whatsapp.Normalizar(v)
	if t == "" {
		return "", ErrInvalido{"telefono", "Revisá el número: con característica, sin 0 ni 15."}
	}
	return t, nil
}

// usuarioPorTelefono aplica los límites y busca la cuenta. "" sin error: no hay cuenta con ese
// número (no se dice, para no revelar qué números tienen cuenta).
func (s *Servicio) usuarioPorTelefono(ctx context.Context, tipo, tel string, cl Cliente) (string, error) {
	if s.WhatsApp == nil {
		return "", ErrSinWhatsApp
	}
	if ok, err := s.permitido(ctx, tipo+":tel:"+tel, 3, time.Hour); err != nil || !ok {
		return "", errOr(err, ErrDemasiadosIntentos)
	}
	if ok, err := s.permitido(ctx, tipo+":ip:"+cl.IP, 20, time.Hour); err != nil || !ok {
		return "", errOr(err, ErrDemasiadosIntentos)
	}
	var usuario string
	err := s.Pool.QueryRow(ctx, `SELECT id FROM auth.usuarios WHERE telefono = $1`, tel).Scan(&usuario)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return usuario, err
}

// PedirEnlace manda por WhatsApp un link para entrar sin contraseña. Responde igual tenga o no
// cuenta ese número.
func (s *Servicio) PedirEnlace(ctx context.Context, numero string, cl Cliente) error {
	tel, err := telefono(numero)
	if err != nil {
		return err
	}
	usuario, err := s.usuarioPorTelefono(ctx, "enlace", tel, cl)
	if err != nil || usuario == "" {
		return err
	}
	token := nuevoToken()
	if _, err := s.Pool.Exec(ctx, `INSERT INTO auth.enlaces (token_hash, usuario_id, vence_en) VALUES ($1, $2, now() + $3::interval)`,
		hashToken(token), usuario, VidaEnlace.String()); err != nil {
		return err
	}
	suf := "m=enlace&token=" + token
	return s.WhatsApp.Enviar(ctx, whatsapp.Mensaje{Para: tel, Tipo: "ingreso", Link: s.Sitio + "/ingresar?" + suf, Sufijo: suf})
}

// UsarEnlace entra con el link (un solo uso).
func (s *Servicio) UsarEnlace(ctx context.Context, token string, cl Cliente) (Sesion, error) {
	var usuario string
	err := s.Pool.QueryRow(ctx, `UPDATE auth.enlaces SET usado_en = now()
		WHERE token_hash = $1 AND usado_en IS NULL AND vence_en > now() RETURNING usuario_id`, hashToken(token)).Scan(&usuario)
	if errors.Is(err, pgx.ErrNoRows) {
		return Sesion{}, ErrTokenInvalido
	}
	if err != nil {
		return Sesion{}, err
	}
	return s.sesionDeUsuario(ctx, usuario, SesionLarga, cl)
}

// RecuperarPorWhatsApp manda el link de contraseña nueva al WhatsApp de la cuenta.
func (s *Servicio) RecuperarPorWhatsApp(ctx context.Context, numero string, cl Cliente) error {
	tel, err := telefono(numero)
	if err != nil {
		return err
	}
	usuario, err := s.usuarioPorTelefono(ctx, "recuperar", tel, cl)
	if err != nil || usuario == "" {
		return err
	}
	token := nuevoToken()
	if _, err := s.Pool.Exec(ctx, `INSERT INTO auth.restablecimientos (token_hash, usuario_id, vence_en) VALUES ($1, $2, now() + $3::interval)`,
		hashToken(token), usuario, VidaRestablecer.String()); err != nil {
		return err
	}
	suf := "m=restablecer&token=" + token
	return s.WhatsApp.Enviar(ctx, whatsapp.Mensaje{Para: tel, Tipo: "restablecer", Link: s.Sitio + "/ingresar?" + suf, Sufijo: suf})
}
