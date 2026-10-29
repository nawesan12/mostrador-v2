// Package arrepentimiento es el botón de arrepentimiento de los planes de Tu Local Web (Res. 424/2020):
// cualquiera, sin iniciar sesión, pide revocar una contratación hecha a distancia dentro de los 10
// días corridos. Se registra con un código de trámite y se confirma por mail; el backoffice lo procesa.
package arrepentimiento

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/limite"
)

// Errores del dominio.
var ErrDemasiadosPedidos = errors.New("demasiados pedidos")

// ErrInvalido es un dato del formulario que no sirve, con el campo para marcarlo.
type ErrInvalido struct{ Campo, Mensaje string }

func (e ErrInvalido) Error() string { return e.Campo + ": " + e.Mensaje }

// Pedido de revocación, tal como lo completa la persona.
type Pedido struct {
	Email, Nombre, Documento, Plan, Motivo string
}

// Servicio de arrepentimientos.
type Servicio struct {
	Pool   *db.Pool
	Correo correo.Correo
	// Bandeja del equipo, que procesa la devolución; vacía, no se avisa.
	Equipo string
	// Ahora, para el año del código (los tests lo fijan).
	Ahora func() time.Time
}

var soloDigitos = regexp.MustCompile(`\D`)

func (s *Servicio) ahora() time.Time {
	if s.Ahora != nil {
		return s.Ahora()
	}
	return time.Now()
}

func validar(p Pedido) (Pedido, error) {
	email := strings.ToLower(strings.TrimSpace(p.Email))
	if a, err := mail.ParseAddress(email); err != nil || a.Address != email || len(email) > 254 {
		return p, ErrInvalido{"email", "Revisá el email: tiene que ser el de la cuenta."}
	}
	nombre := strings.Join(strings.Fields(p.Nombre), " ")
	if nombre == "" || utf8.RuneCountInString(nombre) > 80 {
		return p, ErrInvalido{"nombre", "Escribí tu nombre y apellido."}
	}
	// DNI (7 u 8 cifras) o CUIT (11), con o sin guiones ni puntos.
	doc := soloDigitos.ReplaceAllString(p.Documento, "")
	if n := len(doc); n != 7 && n != 8 && n != 11 {
		return p, ErrInvalido{"documento", "Poné tu DNI o tu CUIT, con o sin guiones."}
	}
	plan := strings.TrimSpace(p.Plan)
	if plan == "" || utf8.RuneCountInString(plan) > 120 {
		return p, ErrInvalido{"plan", "Contanos qué plan contrataste y cuándo."}
	}
	motivo := strings.TrimSpace(p.Motivo)
	if utf8.RuneCountInString(motivo) > 1000 {
		return p, ErrInvalido{"motivo", "El motivo puede tener hasta 1.000 caracteres."}
	}
	return Pedido{Email: email, Nombre: nombre, Documento: doc, Plan: plan, Motivo: motivo}, nil
}

// Pedir registra la revocación y manda la confirmación. Devuelve el código de trámite.
func (s *Servicio) Pedir(ctx context.Context, p Pedido, ip string) (string, error) {
	p, err := validar(p)
	if err != nil {
		return "", err
	}
	for _, l := range []struct {
		clave string
		max   int
	}{{"arrepentimiento:ip:" + ip, 10}, {"arrepentimiento:email:" + p.Email, 3}} {
		ok, err := limite.Permitido(ctx, s.Pool, l.clave, l.max, time.Hour)
		if err != nil {
			return "", err
		}
		if !ok {
			return "", ErrDemasiadosPedidos
		}
	}
	var codigo string
	var motivo *string
	if p.Motivo != "" {
		motivo = &p.Motivo
	}
	err = s.Pool.QueryRow(ctx, `
		INSERT INTO soporte.arrepentimientos (codigo, email, nombre, documento, plan, motivo, ip)
		VALUES (format('ARR-%s-%s', $1::int, nextval('soporte.arrepentimientos_numero')), $2, $3, $4, $5, $6, nullif($7, '')::inet)
		RETURNING codigo`,
		s.ahora().In(argentina).Year(), p.Email, p.Nombre, p.Documento, p.Plan, motivo, ip).Scan(&codigo)
	if err != nil {
		return "", fmt.Errorf("registrar arrepentimiento: %w", err)
	}
	if s.Equipo != "" {
		if err := s.Correo.Enviar(ctx, mailEquipo(s.Equipo, codigo, p)); err != nil {
			slog.Warn("arrepentimiento: no salió el aviso al equipo", "codigo", codigo, "error", err)
		}
	}
	// El trámite ya quedó registrado: si el mail falla, el código vale igual y se ve en pantalla.
	if err := s.Correo.Enviar(ctx, mailRecibido(p.Email, p.Nombre, codigo, p.Plan)); err != nil {
		return codigo, fmt.Errorf("%w: %w", ErrMailPendiente, err)
	}
	return codigo, nil
}

// ErrMailPendiente: el pedido se registró pero el mail de confirmación no salió.
var ErrMailPendiente = errors.New("mail de confirmación pendiente")

var argentina = func() *time.Location {
	l, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	if err != nil {
		return time.FixedZone("ART", -3*3600)
	}
	return l
}()
