// Package formularios recibe los formularios públicos del sitio, sin sesión: la consulta de ventas
// (/contacto), la postulación al programa de partners (/partners) y la suscripción a las novedades
// (/novedades). Cada uno queda registrado en `soporte`, con límite por IP y por email, confirmación
// por mail a quien lo mandó y, si hay bandeja del equipo, un aviso.
package formularios

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/mail"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/limite"
)

// ErrDemasiados: se pasó del límite de envíos.
var ErrDemasiados = errors.New("demasiados envíos")

// ErrTokenInvalido: el link de baja no existe o ya se usó.
var ErrTokenInvalido = errors.New("link de baja inválido")

// ErrInvalido es un dato que no sirve, con el campo para marcarlo en el formulario.
type ErrInvalido struct{ Campo, Mensaje string }

func (e ErrInvalido) Error() string { return e.Campo + ": " + e.Mensaje }

// Servicio de formularios.
type Servicio struct {
	Pool   *db.Pool
	Correo correo.Correo
	// URL pública del sitio, para el link de baja.
	Sitio string
	// Bandeja del equipo; vacía, no se avisa.
	Equipo string
}

// Consulta de ventas (/contacto).
type Consulta struct {
	Nombre, Email, Whatsapp, Web, Facturacion, Origen, Mensaje string
}

// Postulacion al programa de partners (/partners).
type Postulacion struct {
	Nombre, Email, Agencia, Web string
	Especialidades              []string
}

// Opciones válidas (las mismas que los chips del handoff y los CHECK de la base).
var (
	Facturaciones  = []string{"empiezo", "hasta-5", "5-a-20", "20-a-100", "mas-de-100"}
	Origenes       = []string{"tiendanube", "empretienda", "mercadolibre", "shopify", "woocommerce", "ninguna"}
	Especialidades = []string{"diseno", "desarrollo", "marketing", "contabilidad", "fotografia"}
)

var noDigito = regexp.MustCompile(`\D`)

func email(v string) (string, error) {
	e := strings.ToLower(strings.TrimSpace(v))
	if a, err := mail.ParseAddress(e); err != nil || a.Address != e || len(e) > 254 {
		return "", ErrInvalido{"email", "Revisá el email."}
	}
	return e, nil
}

func texto(campo, v string, max int, vacio string) (string, error) {
	t := strings.Join(strings.Fields(v), " ")
	if t == "" && vacio != "" {
		return "", ErrInvalido{campo, vacio}
	}
	if utf8.RuneCountInString(t) > max {
		return "", ErrInvalido{campo, fmt.Sprintf("Puede tener hasta %d caracteres.", max)}
	}
	return t, nil
}

func opcion(campo, v string, validas []string) (string, error) {
	if v == "" || slices.Contains(validas, v) {
		return v, nil
	}
	return "", ErrInvalido{campo, "Elegí una de las opciones."}
}

func nulo(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (s *Servicio) limitar(ctx context.Context, tipo, ip, mail string) error {
	for _, l := range []struct {
		clave string
		max   int
	}{{tipo + ":ip:" + ip, 10}, {tipo + ":email:" + mail, 3}} {
		ok, err := limite.Permitido(ctx, s.Pool, l.clave, l.max, time.Hour)
		if err != nil {
			return err
		}
		if !ok {
			return ErrDemasiados
		}
	}
	return nil
}

// avisar manda un mail; si falla, el envío ya quedó registrado y se loguea.
func (s *Servicio) avisar(ctx context.Context, m correo.Mensaje) {
	if m.Para == "" {
		return
	}
	if err := s.Correo.Enviar(ctx, m); err != nil {
		slog.Warn("formularios: no salió un mail", "asunto", m.Asunto, "error", err)
	}
}

// Consultar registra una consulta de ventas.
func (s *Servicio) Consultar(ctx context.Context, c Consulta, ip string) error {
	var err error
	if c.Nombre, err = texto("nombre", c.Nombre, 80, "Escribí tu nombre."); err != nil {
		return err
	}
	if c.Email, err = email(c.Email); err != nil {
		return err
	}
	if c.Whatsapp = noDigito.ReplaceAllString(c.Whatsapp, ""); c.Whatsapp != "" && (len(c.Whatsapp) < 8 || len(c.Whatsapp) > 15) {
		return ErrInvalido{"whatsapp", "Revisá el número: con característica, sin 0 ni 15."}
	}
	if c.Web, err = texto("web", c.Web, 200, ""); err != nil {
		return err
	}
	if c.Facturacion, err = opcion("facturacion", c.Facturacion, Facturaciones); err != nil {
		return err
	}
	if c.Origen, err = opcion("origen", c.Origen, Origenes); err != nil {
		return err
	}
	if c.Mensaje, err = texto("mensaje", c.Mensaje, 2000, ""); err != nil {
		return err
	}
	if err := s.limitar(ctx, "consulta", ip, c.Email); err != nil {
		return err
	}
	if _, err := s.Pool.Exec(ctx, `
		INSERT INTO soporte.consultas (nombre, email, whatsapp, web, facturacion, origen, mensaje, ip)
		VALUES ($1, $2, $3, $4, $5, $6, $7, nullif($8, '')::inet)`,
		c.Nombre, c.Email, nulo(c.Whatsapp), nulo(c.Web), nulo(c.Facturacion), nulo(c.Origen), nulo(c.Mensaje), ip); err != nil {
		return fmt.Errorf("registrar consulta: %w", err)
	}
	s.avisar(ctx, mailConsultaRecibida(c))
	if s.Equipo != "" {
		s.avisar(ctx, mailConsultaEquipo(s.Equipo, c))
	}
	return nil
}

// Postular registra una postulación al programa de partners. Postularse dos veces con el mismo
// email no es un error: queda la primera.
func (s *Servicio) Postular(ctx context.Context, p Postulacion, ip string) error {
	var err error
	if p.Nombre, err = texto("nombre", p.Nombre, 80, "Escribí tu nombre."); err != nil {
		return err
	}
	if p.Email, err = email(p.Email); err != nil {
		return err
	}
	if p.Agencia, err = texto("agencia", p.Agencia, 120, "¿Cómo se llama tu agencia o estudio?"); err != nil {
		return err
	}
	if p.Web, err = texto("web", p.Web, 200, ""); err != nil {
		return err
	}
	esp := []string{}
	for _, e := range p.Especialidades {
		if !slices.Contains(Especialidades, e) {
			return ErrInvalido{"especialidades", "Elegí de las opciones."}
		}
		if !slices.Contains(esp, e) {
			esp = append(esp, e)
		}
	}
	if err := s.limitar(ctx, "partner", ip, p.Email); err != nil {
		return err
	}
	// Sin ON CONFLICT: necesitaría leer los emails, y la app no puede listar a los postulantes.
	_, err = s.Pool.Exec(ctx, `
		INSERT INTO soporte.postulaciones_partner (nombre, email, agencia, web, especialidades, ip)
		VALUES ($1, $2, $3, $4, $5, nullif($6, '')::inet)`, p.Nombre, p.Email, p.Agencia, nulo(p.Web), esp, ip)
	if pe := (*pgconn.PgError)(nil); errors.As(err, &pe) && pe.Code == "23505" {
		return nil
	}
	if err != nil {
		return fmt.Errorf("registrar postulación: %w", err)
	}
	p.Especialidades = esp
	s.avisar(ctx, mailPostulacionRecibida(p))
	if s.Equipo != "" {
		s.avisar(ctx, mailPostulacionEquipo(s.Equipo, p))
	}
	return nil
}

func tokenBaja() (string, string) {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	t := base64.RawURLEncoding.EncodeToString(b)
	h := sha256.Sum256([]byte(t))
	return t, hex.EncodeToString(h[:])
}

func hash(t string) string {
	h := sha256.Sum256([]byte(t))
	return hex.EncodeToString(h[:])
}

// Suscribir suma un email al resumen mensual de novedades (o lo reactiva si se había dado de baja).
// Si ya está suscripto, no hace nada ni manda otro mail.
func (s *Servicio) Suscribir(ctx context.Context, mailRaw, ip string) error {
	e, err := email(mailRaw)
	if err != nil {
		return err
	}
	if err := s.limitar(ctx, "novedades", ip, e); err != nil {
		return err
	}
	token, h := tokenBaja()
	tag, err := s.Pool.Exec(ctx, `
		INSERT INTO soporte.suscripciones_novedades (email, token_baja_hash) VALUES ($1, $2)
		ON CONFLICT (email) DO UPDATE SET creada_en = now(), baja_en = NULL, token_baja_hash = EXCLUDED.token_baja_hash
		WHERE soporte.suscripciones_novedades.baja_en IS NOT NULL`, e, h)
	if err != nil {
		return fmt.Errorf("suscribir: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return nil
	}
	s.avisar(ctx, mailSuscripcion(e, s.Sitio+"/novedades/baja?token="+token))
	return nil
}

// Baja da de baja la suscripción del link del mail.
func (s *Servicio) Baja(ctx context.Context, token string) error {
	if token == "" {
		return ErrTokenInvalido
	}
	tag, err := s.Pool.Exec(ctx, `
		UPDATE soporte.suscripciones_novedades SET baja_en = now()
		WHERE token_baja_hash = $1 AND baja_en IS NULL`, hash(token))
	if err != nil {
		return fmt.Errorf("baja: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrTokenInvalido
	}
	return nil
}
