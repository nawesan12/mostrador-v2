package mercadolibre

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// RutaVuelta del OAuth: tiene que coincidir exacto con la de la aplicación.
const RutaVuelta = "/integraciones/mercadolibre/callback"

func (s *Servicio) vuelta() string { return s.Config.Sitio + RutaVuelta }

// Estado de la conexión (components/schemas/ConexionML).
type Estado struct {
	Disponible     bool       `json:"disponible"`
	Conectada      bool       `json:"conectada"`
	Apodo          string     `json:"apodo,omitempty"`
	Desde          *time.Time `json:"desde,omitempty"`
	Reautorizar    bool       `json:"reautorizar"`
	Recargo        int        `json:"recargo"`
	PausarSinStock bool       `json:"pausarSinStock"`
	Publicaciones  int        `json:"publicaciones"`
	Preguntas      int        `json:"preguntas"`
}

// EstadoDe la conexión de la tienda.
func (s *Servicio) EstadoDe(ctx context.Context, tiendaID string) (Estado, error) {
	e := Estado{Disponible: s.Activo(), Recargo: 10, PausarSinStock: true}
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var desde time.Time
		err := t.QueryRow(ctx, `SELECT apodo, conectada_en, reautorizar, recargo, pausar_sin_stock,
				(SELECT count(*) FROM ml_publicaciones WHERE tienda_id = $1),
				(SELECT count(*) FROM ml_preguntas WHERE tienda_id = $1 AND estado = 'UNANSWERED')
			FROM ml_conexiones WHERE tienda_id = $1`, tiendaID).
			Scan(&e.Apodo, &desde, &e.Reautorizar, &e.Recargo, &e.PausarSinStock, &e.Publicaciones, &e.Preguntas)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		e.Conectada, e.Desde = true, &desde
		return nil
	})
	return e, err
}

// Conectar guarda el state (con el verificador PKCE) y devuelve la URL de autorización.
func (s *Servicio) Conectar(ctx context.Context, tiendaID, usuarioID string) (string, error) {
	if !s.Activo() {
		return "", ErrApagado
	}
	estado, verificador := azar(), azar()+azar()
	if err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, _ = t.Exec(ctx, `DELETE FROM ml_estados WHERE tienda_id = $1 AND vence_en < now()`, tiendaID)
		_, err := t.Exec(ctx, `INSERT INTO ml_estados (estado_hash, tienda_id, usuario_id, verificador, vence_en)
			VALUES ($1, $2, $3, $4, now() + interval '10 minutes')`, hash(estado), tiendaID, usuarioID, verificador)
		return err
	}); err != nil {
		return "", err
	}
	q := url.Values{
		"response_type":         {"code"},
		"client_id":             {s.Config.AppID},
		"redirect_uri":          {s.vuelta()},
		"state":                 {estado},
		"code_challenge":        {desafio(verificador)},
		"code_challenge_method": {"S256"},
	}
	return s.auth() + "/authorization?" + q.Encode(), nil
}

type tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	UserID       int64  `json:"user_id"`
}

// Volver consume el state, canjea el código y guarda la conexión.
func (s *Servicio) Volver(ctx context.Context, tiendaID, usuarioID, codigo, estado string) error {
	if !s.Activo() {
		return ErrApagado
	}
	if codigo == "" || estado == "" {
		return ErrEstado
	}
	var verificador string
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		err := t.QueryRow(ctx, `UPDATE ml_estados SET usado_en = now()
			WHERE estado_hash = $1 AND tienda_id = $2 AND usuario_id = $3 AND usado_en IS NULL AND vence_en > now()
			RETURNING verificador`, hash(estado), tiendaID, usuarioID).Scan(&verificador)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrEstado
		}
		return err
	})
	if err != nil {
		return err
	}
	var tk tokens
	if err := s.formulario(ctx, "/oauth/token", url.Values{
		"grant_type":    {"authorization_code"},
		"client_id":     {s.Config.AppID},
		"client_secret": {s.Config.Secreto},
		"code":          {codigo},
		"redirect_uri":  {s.vuelta()},
		"code_verifier": {verificador},
	}, &tk); err != nil {
		return err
	}
	if tk.AccessToken == "" || tk.RefreshToken == "" || tk.UserID == 0 {
		return errors.New("mercadolibre: el canje no trajo tokens (¿falta offline_access?)")
	}
	var yo struct {
		Nickname string `json:"nickname"`
	}
	_ = s.pedir(ctx, http.MethodGet, "/users/me", tk.AccessToken, nil, &yo)
	return s.guardar(ctx, tiendaID, tk, yo.Nickname)
}

// guardar: apodo vacío es un refresh (no cambia la cuenta).
func (s *Servicio) guardar(ctx context.Context, tiendaID string, tk tokens, apodo string) error {
	acceso, err := s.Cifra.Cifrar("ml:"+tiendaID, []byte(tk.AccessToken))
	if err != nil {
		return err
	}
	renov, err := s.Cifra.Cifrar("ml:"+tiendaID, []byte(tk.RefreshToken))
	if err != nil {
		return err
	}
	vence := s.ahora().Add(time.Duration(tk.ExpiresIn) * time.Second)
	return s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if apodo == "" {
			_, err := t.Exec(ctx, `UPDATE ml_conexiones SET acceso = $2, renovacion = $3, vence_en = $4, reautorizar = false
				WHERE tienda_id = $1`, tiendaID, acceso, renov, vence)
			return err
		}
		_, err := t.Exec(ctx, `INSERT INTO ml_conexiones (tienda_id, ml_usuario, apodo, acceso, renovacion, vence_en)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (tienda_id) DO UPDATE SET ml_usuario = $2, apodo = $3, acceso = $4, renovacion = $5,
				vence_en = $6, conectada_en = now(), reautorizar = false`,
			tiendaID, tk.UserID, apodo, acceso, renov, vence)
		if err != nil {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "integracion.mercadolibre_conectar",
			Entidad: "integracion", EntidadID: "mercadolibre", Texto: "conectó Mercado Libre (cuenta " + apodo + ")"})
	})
}

// Desconectar borra los tokens. Las publicaciones quedan en Mercado Libre, sin sincronizar.
func (s *Servicio) Desconectar(ctx context.Context, tiendaID string) error {
	return s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `DELETE FROM ml_conexiones WHERE tienda_id = $1`, tiendaID)
		if err != nil || tag.RowsAffected() == 0 {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "integracion.mercadolibre_desconectar",
			Entidad: "integracion", EntidadID: "mercadolibre", Texto: "desconectó Mercado Libre"})
	})
}

// Ajustar el recargo y la pausa sin stock.
func (s *Servicio) Ajustar(ctx context.Context, tiendaID string, recargo int, pausar bool) error {
	if recargo < 0 || recargo > 60 {
		return ErrInvalido
	}
	return s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `UPDATE ml_conexiones SET recargo = $2, pausar_sin_stock = $3 WHERE tienda_id = $1`,
			tiendaID, recargo, pausar)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrSinConexion
		}
		return err
	})
}

// tokenDe: el token de la tienda, renovado si le quedan menos de 30 minutos. Un refresh fallido deja
// la conexión para reautorizar.
func (s *Servicio) tokenDe(ctx context.Context, tiendaID string) (string, int64, error) {
	if !s.Activo() {
		return "", 0, ErrApagado
	}
	var acceso, renov []byte
	var vence time.Time
	var reaut bool
	var usuario int64
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT acceso, renovacion, vence_en, reautorizar, ml_usuario FROM ml_conexiones WHERE tienda_id = $1`,
			tiendaID).Scan(&acceso, &renov, &vence, &reaut, &usuario)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", 0, ErrSinConexion
	}
	if err != nil {
		return "", 0, err
	}
	if reaut {
		return "", 0, ErrReautorizar
	}
	if vence.After(s.ahora().Add(30 * time.Minute)) {
		tk, err := s.Cifra.Descifrar("ml:"+tiendaID, acceso)
		return string(tk), usuario, err
	}
	r, err := s.Cifra.Descifrar("ml:"+tiendaID, renov)
	if err != nil {
		return "", 0, err
	}
	var tk tokens
	if err := s.formulario(ctx, "/oauth/token", url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {s.Config.AppID},
		"client_secret": {s.Config.Secreto},
		"refresh_token": {string(r)},
	}, &tk); err != nil || tk.AccessToken == "" {
		s.log().Warn("mercadolibre: no se pudo renovar el token", "tienda", tiendaID, "error", err)
		// Sólo la primera vez que se cae avisa (la campanita, §45): los reintentos no repiten.
		_ = s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
			tag, err := t.Exec(ctx, `UPDATE ml_conexiones SET reautorizar = true WHERE tienda_id = $1 AND NOT reautorizar`, tiendaID)
			if err != nil || tag.RowsAffected() == 0 {
				return err
			}
			return outbox.Publicar(ctx, t, tiendaID, "integracion.reautorizar", map[string]string{"canal": "ml"})
		})
		return "", 0, ErrReautorizar
	}
	if tk.RefreshToken == "" {
		tk.RefreshToken = string(r)
	}
	if err := s.guardar(ctx, tiendaID, tk, ""); err != nil {
		return "", 0, err
	}
	return tk.AccessToken, usuario, nil
}
