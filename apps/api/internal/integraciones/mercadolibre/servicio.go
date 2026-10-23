// Package mercadolibre: Mercado Libre como canal de venta de la tienda.
//
//   - Cada tienda conecta su cuenta por OAuth (PKCE, state de un solo uso) y Mostrador usa su token,
//     guardado cifrado. El access token dura 6 horas y el refresh token se usa una sola vez: se
//     reemplazan los dos juntos.
//   - Publicar: el producto de Mostrador pasa a una publicación (categoría sugerida por el título,
//     precio con el recargo de la tienda, stock, fotos y descripción).
//   - Stock y precio: un trigger encola ml.sincronizar cuando cambian; el worker lo manda.
//   - Avisos (orders_v2, questions, items): sin firma, así que se verifica la aplicación y se trae el
//     recurso con el token de la tienda (el aviso nunca es la fuente de verdad).
package mercadolibre

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
)

// Errores que el panel traduce.
var (
	ErrApagado     = errors.New("mercadolibre: sin credenciales de la aplicación")
	ErrSinConexion = errors.New("mercadolibre: la tienda no conectó su cuenta")
	ErrReautorizar = errors.New("mercadolibre: hay que volver a conectar la cuenta")
	ErrEstado      = errors.New("mercadolibre: state inválido o vencido")
	ErrNoExiste    = errors.New("mercadolibre: no existe")
	ErrYaPublicado = errors.New("mercadolibre: ya está publicado")
	ErrInvalido    = errors.New("mercadolibre: datos inválidos")
)

// ErrorML es una respuesta no exitosa de la API. Mensaje es el "message" que manda Mercado Libre
// (sirve para decirle a la tienda qué falta en una publicación); nunca lleva tokens.
type ErrorML struct {
	Estado  int
	Mensaje string
	Cuerpo  string
}

func (e *ErrorML) Error() string { return fmt.Sprintf("mercadolibre: HTTP %d %s", e.Estado, e.Mensaje) }

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
	EnTx(context.Context, func(pgx.Tx) error) error
}

// Config de la aplicación de Mostrador en Mercado Libre.
type Config struct {
	AppID, Secreto string
	// Sitio público (la vuelta del OAuth).
	Sitio string
}

// Servicio de Mercado Libre.
type Servicio struct {
	Pool   tx
	Config Config
	Cifra  *cifrado.Cifrador
	// Base de la API y de la autorización (se cambian en los tests).
	API, Auth string
	HTTP      *http.Client
	Log       *slog.Logger
	Ahora     func() time.Time
}

// DesdeConfig arma el servicio (lo usan la API y el worker).
func DesdeConfig(pool tx, appID, secreto, sitio, claveCifrado string, log *slog.Logger) *Servicio {
	s := &Servicio{Pool: pool, Log: log, Config: Config{AppID: appID, Secreto: secreto, Sitio: sitio}}
	if claveCifrado != "" {
		if c, err := cifrado.Nuevo(claveCifrado); err == nil {
			s.Cifra = c
		}
	}
	return s
}

// Activo: hay credenciales para conectar tiendas.
func (s *Servicio) Activo() bool {
	return s != nil && s.Config.AppID != "" && s.Config.Secreto != "" && s.Cifra != nil
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

func (s *Servicio) api() string {
	if s.API != "" {
		return s.API
	}
	return "https://api.mercadolibre.com"
}

func (s *Servicio) auth() string {
	if s.Auth != "" {
		return s.Auth
	}
	return "https://auth.mercadolibre.com.ar"
}

func (s *Servicio) http() *http.Client {
	if s.HTTP != nil {
		return s.HTTP
	}
	return &http.Client{Timeout: 15 * time.Second}
}

// pedir a la API con el token dado.
func (s *Servicio) pedir(ctx context.Context, metodo, ruta, token string, cuerpo, salida any) error {
	var r io.Reader
	if cuerpo != nil {
		b, err := json.Marshal(cuerpo)
		if err != nil {
			return err
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, metodo, s.api()+ruta, r)
	if err != nil {
		return err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Accept", "application/json")
	if cuerpo != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return s.hacer(req, salida)
}

func (s *Servicio) formulario(ctx context.Context, ruta string, v url.Values, salida any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.api()+ruta, strings.NewReader(v.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	return s.hacer(req, salida)
}

func (s *Servicio) hacer(req *http.Request, salida any) error {
	res, err := s.http().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(res.Body, 2<<20))
	if res.StatusCode >= 300 {
		var m struct {
			Message string `json:"message"`
			Cause   []struct {
				Message string `json:"message"`
			} `json:"cause"`
		}
		_ = json.Unmarshal(b, &m)
		msg := m.Message
		if len(m.Cause) > 0 && m.Cause[0].Message != "" {
			msg = m.Cause[0].Message
		}
		return &ErrorML{Estado: res.StatusCode, Mensaje: msg, Cuerpo: string(b)}
	}
	if salida == nil || len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, salida)
}

func azar() string {
	b := make([]byte, 24)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func hash(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func desafio(verificador string) string {
	h := sha256.Sum256([]byte(verificador))
	return base64.RawURLEncoding.EncodeToString(h[:])
}
