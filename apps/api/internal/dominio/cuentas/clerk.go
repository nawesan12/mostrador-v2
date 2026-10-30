package cuentas

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/clerk"
)

// ErrTokenClerk: el token de sesión de Clerk no sirve (vencido, de otra instancia, mal firmado).
var ErrTokenClerk = errors.New("token de Clerk inválido")

// IngresarConClerk entra con el token de sesión de Clerk que manda la web (después de Google o
// Apple) o la app de iPhone. Go lo verifica con las claves de la instancia, lee a la persona con la
// Backend API y sigue igual que los demás ingresos sociales: la identidad ya vinculada entra; si no,
// se vincula a la cuenta con ese email (Clerk sólo lo da si está verificado) o se crea la cuenta
// (onboarding y mail de bienvenida).
func (s *Servicio) IngresarConClerk(ctx context.Context, token string, cl Cliente) (IngresoSocial, error) {
	if s.Clerk == nil {
		return IngresoSocial{}, ErrProveedorApagado
	}
	if ok, err := s.permitido(ctx, "clerk:ip:"+cl.IP, 30, 15*time.Minute); err != nil || !ok {
		return IngresoSocial{}, errOr(err, ErrDemasiadosIntentos)
	}
	sujeto, err := s.Clerk.Verificar(ctx, strings.TrimSpace(token))
	if errors.Is(err, clerk.ErrToken) {
		return IngresoSocial{}, ErrTokenClerk
	}
	if err != nil {
		return IngresoSocial{}, err
	}
	u, err := s.Clerk.Usuario(ctx, sujeto)
	if err != nil {
		return IngresoSocial{}, err
	}
	if u.Email == "" {
		return IngresoSocial{}, ErrSinEmail
	}
	return s.entrarConPerfil(ctx, "clerk", Perfil{Sujeto: sujeto, Email: u.Email, Nombre: u.Nombre,
		Verificado: true, Foto: u.Foto}, cl)
}
