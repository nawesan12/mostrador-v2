// Package telegram: Hornero en Telegram (@mostradorhornero_bot, para la dueña).
//
//   - Vínculo: el panel genera un código de un solo uso (10 min, se guarda su sha256) y el bot lo
//     canjea con el chat. Un chat por tienda y una tienda por chat.
//   - Respuestas: el modelo vive en el Worker (binding AI), así que el BFF le pide a Go el contexto de
//     la tienda del chat (lo mismo que arma el Hornero del panel) y responde él. Go no tiene IA.
//   - Avisos: un trigger encola "telegram.venta" por cada pedido de una tienda con chat vinculado y
//     el worker lo manda por la Bot API, una sola vez (pedidos.avisos).
//
// El webhook de Telegram lo atiende el Worker; las rutas de acá las llama el BFF con su clave.
package telegram

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/limite"
)

// Vigencia del código de vínculo (como la demo).
const Vigencia = 10 * time.Minute

var (
	ErrCodigo      = errors.New("telegram: el código venció, ya se usó o no existe")
	ErrChatOcupado = errors.New("telegram: el chat ya está vinculado a otra tienda")
	ErrSinVinculo  = errors.New("telegram: el chat no está vinculado")
	ErrSinPermiso  = errors.New("telegram: sin permiso")
	ErrInvalido    = errors.New("telegram: pedido inválido")
	ErrDemasiados  = errors.New("telegram: demasiados intentos")
)

var argentina = time.FixedZone("ART", -3*3600)

// Servicio del vínculo y el contexto. Bot manda los avisos (nil: no se avisa).
type Servicio struct {
	Pool *db.Pool
	Bot  *Bot
}

// alfabeto sin 0, 1, I ni O (como la demo): 32 letras, así que un byte módulo 32 no tiene sesgo.
const alfabeto = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

var codigoValido = regexp.MustCompile(`^[2-9A-HJ-NP-Z]{8}$`)

func nuevoCodigo() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = alfabeto[int(b[i])%len(alfabeto)]
	}
	return string(b)
}

func hash(t string) string {
	h := sha256.Sum256([]byte(t))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func esAdmin(rol string) bool { return rol == "duenio" || rol == "admin" }

// Codigo es lo que se le da al panel para armar el link al bot.
type Codigo struct {
	Codigo  string    `json:"codigo"`
	VenceEn time.Time `json:"venceEn"`
}

// Generar da un código nuevo para vincular un chat. El anterior (si había) deja de servir.
func (s *Servicio) Generar(ctx context.Context, tienda, rol string) (Codigo, error) {
	if !esAdmin(rol) {
		return Codigo{}, ErrSinPermiso
	}
	c := Codigo{Codigo: nuevoCodigo()}
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		if _, err := t.Exec(ctx, `DELETE FROM telegram_codigos WHERE tienda_id = $1`, tienda); err != nil {
			return err
		}
		return t.QueryRow(ctx, `INSERT INTO telegram_codigos (hash, tienda_id, vence_en)
			VALUES ($1, $2, now() + $3::interval) RETURNING vence_en`, hash(c.Codigo), tienda, Vigencia.String()).Scan(&c.VenceEn)
	})
	return c, err
}

// Estado del vínculo para el panel.
type Estado struct {
	Vinculado bool       `json:"vinculado"`
	Nombre    string     `json:"nombre,omitempty"`
	Desde     *time.Time `json:"desde,omitempty"`
}

func (s *Servicio) Estado(ctx context.Context, tienda string) (Estado, error) {
	var e Estado
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var desde time.Time
		err := t.QueryRow(ctx, `SELECT nombre, vinculado_en FROM telegram_vinculos WHERE tienda_id = $1`, tienda).
			Scan(&e.Nombre, &desde)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		e.Vinculado, e.Desde = err == nil, &desde
		return err
	})
	return e, err
}

// Vinculado es la respuesta del bot al canjear el código.
type Vinculado struct {
	Tienda string `json:"tienda"`
}

// Vincular canjea el código (un solo uso) y deja el chat como el de la tienda; si la tienda tenía
// otro chat, lo reemplaza. Un chat que ya está en otra tienda no se mueve: primero /desvincular.
func (s *Servicio) Vincular(ctx context.Context, codigo string, chat int64, nombre string) (Vinculado, error) {
	codigo = strings.ToUpper(strings.TrimSpace(codigo))
	if chat == 0 || !codigoValido.MatchString(codigo) {
		return Vinculado{}, ErrCodigo
	}
	nombre = strings.TrimSpace(nombre)
	if r := []rune(nombre); len(r) > 60 {
		nombre = string(r[:60])
	}
	if nombre == "" {
		nombre = "Telegram"
	}
	// 8 letras de 32 son 40 bits y el código dura 10 minutos; igual, 10 intentos por hora por chat.
	ok, err := limite.Permitido(ctx, s.Pool, "telegram:vincular:"+strconv.FormatInt(chat, 10), 10, time.Hour)
	if err != nil {
		return Vinculado{}, err
	}
	if !ok {
		return Vinculado{}, ErrDemasiados
	}
	tienda, err := s.tiendaDe(ctx, `SELECT app.tienda_de_codigo_telegram($1)::text`, hash(codigo))
	if err != nil {
		if errors.Is(err, ErrSinVinculo) {
			return Vinculado{}, ErrCodigo
		}
		return Vinculado{}, err
	}
	var v Vinculado
	vigente := false
	err = s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		err := t.QueryRow(ctx, `DELETE FROM telegram_codigos WHERE hash = $1 RETURNING vence_en > now()`, hash(codigo)).Scan(&vigente)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCodigo
		}
		if err != nil || !vigente {
			return err
		}
		if _, err := t.Exec(ctx, `INSERT INTO telegram_vinculos (tienda_id, chat, nombre) VALUES ($1, $2, $3)
			ON CONFLICT (tienda_id) DO UPDATE SET chat = EXCLUDED.chat, nombre = EXCLUDED.nombre, vinculado_en = now()`,
			tienda, chat, nombre); err != nil {
			var pg *pgconn.PgError
			if errors.As(err, &pg) && pg.Code == "23505" {
				return ErrChatOcupado
			}
			return err
		}
		return t.QueryRow(ctx, `SELECT nombre FROM tiendas WHERE id = $1`, tienda).Scan(&v.Tienda)
	})
	if err == nil && !vigente {
		// Vencido: se borró igual (no sirve más), pero no vincula.
		return Vinculado{}, ErrCodigo
	}
	return v, err
}

// tiendaDe resuelve la tienda con una de las funciones app.tienda_de_* (NULL: ErrSinVinculo).
func (s *Servicio) tiendaDe(ctx context.Context, consulta string, arg any) (string, error) {
	var tienda *string
	err := s.Pool.EnTx(ctx, func(t pgx.Tx) error {
		return t.QueryRow(ctx, consulta, arg).Scan(&tienda)
	})
	if err != nil {
		return "", err
	}
	if tienda == nil {
		return "", ErrSinVinculo
	}
	return *tienda, nil
}

// TiendaDelChat devuelve la tienda vinculada al chat.
func (s *Servicio) TiendaDelChat(ctx context.Context, chat int64) (string, error) {
	return s.tiendaDe(ctx, `SELECT app.tienda_de_telegram($1)::text`, chat)
}

// Desvincular saca el chat de la tienda desde el panel. Devuelve el chat que estaba (0 si ninguno),
// para que el BFF le avise.
func (s *Servicio) Desvincular(ctx context.Context, tienda, rol string) (int64, error) {
	if !esAdmin(rol) {
		return 0, ErrSinPermiso
	}
	var chat int64
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		err := t.QueryRow(ctx, `DELETE FROM telegram_vinculos WHERE tienda_id = $1 RETURNING chat`, tienda).Scan(&chat)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	})
	return chat, err
}

// DesvincularChat es el /desvincular desde el chat.
func (s *Servicio) DesvincularChat(ctx context.Context, chat int64) error {
	tienda, err := s.TiendaDelChat(ctx, chat)
	if err != nil {
		return err
	}
	return s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `DELETE FROM telegram_vinculos WHERE tienda_id = $1 AND chat = $2`, tienda, chat)
		return err
	})
}

// Contexto es lo que necesita Hornero para responder en el chat: lo mismo que arma el panel
// (tienda, tablero de hoy y de 7 días, contexto) y cuántos mensajes lleva hoy la tienda.
type Contexto struct {
	Tienda   tiendas.Tienda  `json:"tienda"`
	Hoy      tiendas.Tablero `json:"hoy"`
	Semana   tiendas.Tablero `json:"semana"`
	Contexto map[string]any  `json:"contexto"`
	Usos     int             `json:"usos"`
}

// ContextoDe suma un mensaje al día de la tienda del chat y devuelve su contexto. Sólo lectura de la
// tienda vinculada: el chat no puede pedir otra.
func (s *Servicio) ContextoDe(ctx context.Context, chat int64, ahora time.Time) (Contexto, error) {
	var c Contexto
	tienda, err := s.TiendaDelChat(ctx, chat)
	if err != nil {
		return c, err
	}
	dia := ahora.In(argentina).Format(time.DateOnly)
	if err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `UPDATE telegram_vinculos
			SET uso_n = CASE WHEN uso_dia = $2::date THEN uso_n + 1 ELSE 1 END, uso_dia = $2::date
			WHERE tienda_id = $1 RETURNING uso_n`, tienda, dia).Scan(&c.Usos)
	}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c, ErrSinVinculo
		}
		return c, err
	}
	if c.Tienda, err = tiendas.Obtener(ctx, s.Pool, tienda); err != nil {
		return c, err
	}
	if c.Hoy, err = tiendas.ObtenerTablero(ctx, s.Pool, tienda, "hoy", ahora); err != nil {
		return c, err
	}
	if c.Semana, err = tiendas.ObtenerTablero(ctx, s.Pool, tienda, "7d", ahora); err != nil {
		return c, err
	}
	c.Contexto = tiendas.ContextoHornero(ctx, s.Pool, tienda)
	return c, nil
}
