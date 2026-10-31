package mercadopago

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
)

// RutaVuelta del OAuth de conexión: tiene que coincidir exacto con la de la aplicación.
const RutaVuelta = "/integraciones/mercadopago/callback"

func (s *Servicio) vuelta() string { return s.Config.Sitio + RutaVuelta }

// Estado es lo que ve el panel (components/schemas/ConexionMP).
type Estado struct {
	Disponible  bool       `json:"disponible"`
	Conectada   bool       `json:"conectada"`
	Usuario     string     `json:"usuario,omitempty"`
	Desde       *time.Time `json:"desde,omitempty"`
	Reautorizar bool       `json:"reautorizar"`
	CajaQR      bool       `json:"cajaQR"`
}

// EstadoDe la conexión de la tienda.
func (s *Servicio) EstadoDe(ctx context.Context, tiendaID string) (Estado, error) {
	e := Estado{Disponible: s.Activo()}
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var usuario int64
		var desde time.Time
		var caja *string
		err := t.QueryRow(ctx, `SELECT mp_usuario, conectada_en, reautorizar, caja_externa FROM mp_conexiones WHERE tienda_id = $1`,
			tiendaID).Scan(&usuario, &desde, &e.Reautorizar, &caja)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		e.Conectada, e.Usuario, e.Desde, e.CajaQR = true, strconv.FormatInt(usuario, 10), &desde, caja != nil
		return nil
	})
	return e, err
}

// Conectar guarda un state de un solo uso (atado a la tienda y a la persona) y devuelve la URL de
// autorización de Mercado Pago.
func (s *Servicio) Conectar(ctx context.Context, tiendaID, usuarioID string) (string, error) {
	if !s.Activo() {
		return "", ErrApagado
	}
	estado, verificador := azar(), azar()+azar()
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, _ = t.Exec(ctx, `DELETE FROM mp_estados WHERE tienda_id = $1 AND vence_en < now()`, tiendaID)
		_, err := t.Exec(ctx, `INSERT INTO mp_estados (estado_hash, tienda_id, usuario_id, verificador, vence_en)
			VALUES ($1, $2, $3, $4, now() + interval '10 minutes')`, hash(estado), tiendaID, usuarioID, verificador)
		return err
	})
	if err != nil {
		return "", err
	}
	q := url.Values{
		"client_id":     {s.Config.AppID},
		"response_type": {"code"},
		"platform_id":   {"mp"},
		"redirect_uri":  {s.vuelta()},
		"state":         {estado},
		// PKCE S256 (la aplicación lo exige).
		"code_challenge":        {desafio(verificador)},
		"code_challenge_method": {"S256"},
	}
	return "https://auth.mercadopago.com/authorization?" + q.Encode(), nil
}

type tokens struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	Scope        string `json:"scope"`
	UserID       int64  `json:"user_id"`
}

// Volver consume el state, canjea el código y guarda los tokens cifrados.
func (s *Servicio) Volver(ctx context.Context, tiendaID, usuarioID, codigo, estado string) error {
	if !s.Activo() {
		return ErrApagado
	}
	if codigo == "" || estado == "" {
		return ErrEstado
	}
	var verificador string
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		err := t.QueryRow(ctx, `UPDATE mp_estados SET usado_en = now()
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
	if err := s.cliente().Formulario(ctx, "/oauth/token", url.Values{
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
		return errors.New("mercadopago: el canje no trajo tokens")
	}
	return s.guardarTokens(ctx, tiendaID, tk, true)
}

func (s *Servicio) guardarTokens(ctx context.Context, tiendaID string, tk tokens, nueva bool) error {
	acceso, err := s.Cifra.Cifrar("mp:"+tiendaID, []byte(tk.AccessToken))
	if err != nil {
		return err
	}
	renov, err := s.Cifra.Cifrar("mp:"+tiendaID, []byte(tk.RefreshToken))
	if err != nil {
		return err
	}
	vence := s.ahora().Add(time.Duration(tk.ExpiresIn) * time.Second)
	return s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if nueva {
			// Otra cuenta de Mercado Pago reemplaza a la anterior (y la caja de QR era de esa cuenta).
			_, err := t.Exec(ctx, `INSERT INTO mp_conexiones (tienda_id, mp_usuario, acceso, renovacion, vence_en, alcances)
				VALUES ($1, $2, $3, $4, $5, $6)
				ON CONFLICT (tienda_id) DO UPDATE SET mp_usuario = $2, acceso = $3, renovacion = $4, vence_en = $5,
					alcances = $6, conectada_en = now(), reautorizar = false,
					sucursal_mp = CASE WHEN mp_conexiones.mp_usuario = $2 THEN mp_conexiones.sucursal_mp END,
					caja_externa = CASE WHEN mp_conexiones.mp_usuario = $2 THEN mp_conexiones.caja_externa END`,
				tiendaID, tk.UserID, acceso, renov, vence, tk.Scope)
			if err != nil {
				return err
			}
			return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "integracion.mercadopago_conectar",
				Entidad: "integracion", EntidadID: "mercadopago", Texto: "conectó Mercado Pago"})
		}
		// Un refresh puede rotar el refresh token: se reemplazan los dos juntos.
		_, err := t.Exec(ctx, `UPDATE mp_conexiones SET acceso = $2, renovacion = $3, vence_en = $4, reautorizar = false
			WHERE tienda_id = $1`, tiendaID, acceso, renov, vence)
		return err
	})
}

// Desconectar borra los tokens. (La autorización se revoca desde la cuenta de Mercado Pago.)
func (s *Servicio) Desconectar(ctx context.Context, tiendaID string) error {
	return s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `DELETE FROM mp_conexiones WHERE tienda_id = $1`, tiendaID)
		if err != nil || tag.RowsAffected() == 0 {
			return err
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "integracion.mercadopago_desconectar",
			Entidad: "integracion", EntidadID: "mercadopago", Texto: "desconectó Mercado Pago"})
	})
}

// tokenDe devuelve el token de la tienda, renovándolo si vence en menos de una semana. Si el refresh
// falla, la conexión queda para reautorizar: nunca se usa otro token.
func (s *Servicio) tokenDe(ctx context.Context, tiendaID string) (string, int64, error) {
	if !s.Activo() {
		return "", 0, ErrApagado
	}
	var acceso, renov []byte
	var vence time.Time
	var reaut bool
	var usuario int64
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT acceso, renovacion, vence_en, reautorizar, mp_usuario FROM mp_conexiones WHERE tienda_id = $1`,
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
	if vence.After(s.ahora().Add(7 * 24 * time.Hour)) {
		tk, err := s.Cifra.Descifrar("mp:"+tiendaID, acceso)
		return string(tk), usuario, err
	}
	r, err := s.Cifra.Descifrar("mp:"+tiendaID, renov)
	if err != nil {
		return "", 0, err
	}
	var tk tokens
	if err := s.cliente().Formulario(ctx, "/oauth/token", url.Values{
		"grant_type":    {"refresh_token"},
		"client_id":     {s.Config.AppID},
		"client_secret": {s.Config.Secreto},
		"refresh_token": {string(r)},
	}, &tk); err != nil || tk.AccessToken == "" {
		s.log().Warn("mercadopago: no se pudo renovar el token", "tienda", tiendaID, "error", err)
		_ = s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
			_, err := t.Exec(ctx, `UPDATE mp_conexiones SET reautorizar = true WHERE tienda_id = $1`, tiendaID)
			return err
		})
		return "", 0, ErrReautorizar
	}
	if tk.RefreshToken == "" {
		tk.RefreshToken = string(r)
	}
	if err := s.guardarTokens(ctx, tiendaID, tk, false); err != nil {
		return "", 0, err
	}
	return tk.AccessToken, usuario, nil
}
