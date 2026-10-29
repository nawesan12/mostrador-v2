package mercadopago

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
)

// Errores que el panel y la página de pago traducen.
var (
	ErrApagado      = errors.New("mercadopago: sin credenciales de la aplicación")
	ErrSinConexion  = errors.New("mercadopago: la tienda no conectó su cuenta")
	ErrReautorizar  = errors.New("mercadopago: hay que volver a conectar la cuenta")
	ErrEstado       = errors.New("mercadopago: state inválido o vencido")
	ErrNoExiste     = errors.New("mercadopago: no existe")
	ErrYaPagado     = errors.New("mercadopago: ya está pagado")
	ErrSinCaja      = errors.New("mercadopago: falta la caja para cobrar con QR")
	ErrFirma        = errors.New("mercadopago: firma inválida")
	ErrPlanInvalido = errors.New("mercadopago: plan inválido")
	ErrInvalido     = errors.New("mercadopago: datos inválidos")
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
	EnTx(context.Context, func(pgx.Tx) error) error
}

// Config de la aplicación de Tu Local Web en Mercado Pago.
type Config struct {
	// OAuth de la aplicación (marketplace).
	AppID, Secreto string
	// Token de la cuenta de la plataforma: sólo para el abono de Tu Local Web.
	TokenPlataforma string
	// Clave secreta de los webhooks (Tus integraciones › Webhooks).
	SecretoWebhook string
	// Sitio público (vueltas del checkout y del OAuth).
	Sitio string
}

// Servicio de Mercado Pago.
type Servicio struct {
	Pool    tx
	Config  Config
	Cliente *Cliente
	Cifra   *cifrado.Cifrador
	Log     *slog.Logger
	Ahora   func() time.Time
}

func (s *Servicio) ahora() time.Time {
	if s.Ahora != nil {
		return s.Ahora()
	}
	return time.Now()
}

func (s *Servicio) log() *slog.Logger {
	if s.Log != nil {
		return s.Log
	}
	return slog.Default()
}

// Activo: la aplicación tiene credenciales para conectar tiendas.
func (s *Servicio) Activo() bool {
	return s != nil && s.Config.AppID != "" && s.Config.Secreto != "" && s.Cifra != nil
}

func (s *Servicio) cliente() *Cliente {
	if s.Cliente != nil {
		return s.Cliente
	}
	return &Cliente{}
}

// publico: auto_return sólo con un sitio https que no sea local (LOCK 6 del plugin).
func (s *Servicio) publico() bool {
	u := s.Config.Sitio
	return strings.HasPrefix(u, "https://") && !strings.Contains(u, "localhost") && !strings.Contains(u, "127.0.0.1")
}

func azar() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// desafio PKCE S256 del verificador.
func desafio(verificador string) string {
	h := sha256.Sum256([]byte(verificador))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

// DesdeConfig arma el servicio con la configuración de la API (lo usan la API y el worker). Sin
// clave de cifrado, las tiendas no pueden conectar su cuenta (Activo da false).
func DesdeConfig(pool tx, appID, secreto, tokenPlataforma, secretoWebhook, sitio, claveCifrado string, log *slog.Logger) *Servicio {
	s := &Servicio{Pool: pool, Log: log, Config: Config{AppID: appID, Secreto: secreto, TokenPlataforma: tokenPlataforma,
		SecretoWebhook: secretoWebhook, Sitio: sitio}}
	if claveCifrado != "" {
		c, err := cifrado.Nuevo(claveCifrado)
		if err != nil {
			log.Error("mercadopago: MOSTRADOR_CLAVE_CIFRADO inválida; las tiendas no van a poder conectar su cuenta")
		}
		s.Cifra = c
	}
	return s
}
