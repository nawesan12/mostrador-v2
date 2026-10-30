package cuentas

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

// Errores del ingreso con proveedores.
var (
	ErrProveedorApagado = errors.New("proveedor sin configurar")
	ErrEstadoInvalido   = errors.New("state inválido o vencido")
	ErrSinEmail         = errors.New("el proveedor no dio un email")
	// ErrVincular: hay una cuenta con ese email pero el proveedor no lo verifica: no se vincula sola.
	ErrVincular = errors.New("email en uso sin verificar")
)

// Proveedor OAuth 2 (con PKCE S256). Perfil lee la respuesta de PerfilURL.
type Proveedor struct {
	ID, Nombre                   string
	AuthURL, TokenURL, PerfilURL string
	ClientID, Secreto            string
	Alcances                     []string
	// Parámetros extra de la autorización (p. ej. platform_id de Mercado Pago).
	Extra  map[string]string
	Perfil func(map[string]any) Perfil
}

// Perfil de la persona según el proveedor.
type Perfil struct {
	Sujeto, Email, Nombre string
	// Verificado: el proveedor garantiza que el email es de esta persona (sólo así se vincula a una
	// cuenta existente).
	Verificado bool
}

func texto2(m map[string]any, k string) string {
	switch v := m[k].(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	}
	return ""
}

// Proveedores con sus URLs; ClientID y Secreto los completa la configuración. Sin ClientID, el
// proveedor está apagado.
func Proveedores(cfg map[string][2]string) map[string]*Proveedor {
	ps := map[string]*Proveedor{
		"google": {
			ID: "google", Nombre: "Google",
			AuthURL: "https://accounts.google.com/o/oauth2/v2/auth", TokenURL: "https://oauth2.googleapis.com/token",
			PerfilURL: "https://openidconnect.googleapis.com/v1/userinfo", Alcances: []string{"openid", "email", "profile"},
			Extra: map[string]string{"prompt": "select_account"},
			Perfil: func(m map[string]any) Perfil {
				v, _ := m["email_verified"].(bool)
				return Perfil{Sujeto: texto2(m, "sub"), Email: texto2(m, "email"), Nombre: texto2(m, "given_name"), Verificado: v}
			},
		},
		"mercadopago": {
			ID: "mercadopago", Nombre: "Mercado Pago",
			AuthURL: "https://auth.mercadopago.com/authorization", TokenURL: "https://api.mercadopago.com/oauth/token",
			PerfilURL: "https://api.mercadopago.com/users/me", Extra: map[string]string{"platform_id": "mp"},
			Perfil: func(m map[string]any) Perfil {
				return Perfil{Sujeto: texto2(m, "id"), Email: texto2(m, "email"), Nombre: texto2(m, "first_name")}
			},
		},
		"mercadolibre": {
			ID: "mercadolibre", Nombre: "Mercado Libre",
			AuthURL: "https://auth.mercadolibre.com.ar/authorization", TokenURL: "https://api.mercadolibre.com/oauth/token",
			PerfilURL: "https://api.mercadolibre.com/users/me",
			Perfil: func(m map[string]any) Perfil {
				n := texto2(m, "first_name")
				if n == "" {
					n = texto2(m, "nickname")
				}
				return Perfil{Sujeto: texto2(m, "id"), Email: texto2(m, "email"), Nombre: n}
			},
		},
	}
	for id, p := range ps {
		p.ClientID, p.Secreto = cfg[id][0], cfg[id][1]
	}
	return ps
}

func (s *Servicio) proveedor(id string) (*Proveedor, error) {
	p, ok := s.OAuth[id]
	if !ok || p.ClientID == "" {
		return nil, ErrProveedorApagado
	}
	return p, nil
}

// Disponibles: qué proveedores están configurados.
func (s *Servicio) Disponibles() []string {
	out := []string{}
	for _, id := range []string{"google", "mercadopago", "mercadolibre"} {
		if p, ok := s.OAuth[id]; ok && p.ClientID != "" {
			out = append(out, id)
		}
	}
	return out
}

func (s *Servicio) volverA(proveedor string) string {
	return s.Sitio + "/ingresar/" + proveedor + "/vuelta"
}

// IniciarOAuth guarda el state y el verificador PKCE y devuelve la URL de autorización.
func (s *Servicio) IniciarOAuth(ctx context.Context, proveedor string, cl Cliente) (string, error) {
	p, err := s.proveedor(proveedor)
	if err != nil {
		return "", err
	}
	if ok, err := s.permitido(ctx, "oauth:ip:"+cl.IP, 30, 15*time.Minute); err != nil || !ok {
		return "", errOr(err, ErrDemasiadosIntentos)
	}
	estado, verificador := nuevoToken(), nuevoToken()
	// De paso, se limpian los que ya vencieron (barato: índice por vencimiento).
	_, _ = s.Pool.Exec(ctx, `DELETE FROM auth.oauth_estados WHERE vence_en < now() - interval '1 day'`)
	if _, err := s.Pool.Exec(ctx, `INSERT INTO auth.oauth_estados (estado_hash, proveedor, verificador, vence_en)
		VALUES ($1, $2, $3, now() + interval '10 minutes')`, hashToken(estado), proveedor, verificador); err != nil {
		return "", err
	}
	desafio := sha256.Sum256([]byte(verificador))
	q := url.Values{
		"client_id":             {p.ClientID},
		"response_type":         {"code"},
		"redirect_uri":          {s.volverA(proveedor)},
		"state":                 {estado},
		"code_challenge":        {base64.RawURLEncoding.EncodeToString(desafio[:])},
		"code_challenge_method": {"S256"},
	}
	if len(p.Alcances) > 0 {
		q.Set("scope", strings.Join(p.Alcances, " "))
	}
	for k, v := range p.Extra {
		q.Set(k, v)
	}
	return p.AuthURL + "?" + q.Encode(), nil
}

// IngresoSocial: la sesión y si la cuenta es nueva (va al onboarding).
type IngresoSocial struct {
	Sesion
	Nueva bool `json:"nueva"`
}

// VolverOAuth canjea el código, lee el perfil y entra: con la identidad ya vinculada, vinculándola
// a la cuenta del mismo email (sólo si el proveedor lo verifica) o creando una cuenta nueva.
func (s *Servicio) VolverOAuth(ctx context.Context, proveedor, codigo, estado string, cl Cliente) (IngresoSocial, error) {
	p, err := s.proveedor(proveedor)
	if err != nil {
		return IngresoSocial{}, err
	}
	var verificador string
	err = s.Pool.QueryRow(ctx, `UPDATE auth.oauth_estados SET usado_en = now()
		WHERE estado_hash = $1 AND proveedor = $2 AND usado_en IS NULL AND vence_en > now()
		RETURNING verificador`, hashToken(estado), proveedor).Scan(&verificador)
	if errors.Is(err, pgx.ErrNoRows) {
		return IngresoSocial{}, ErrEstadoInvalido
	}
	if err != nil {
		return IngresoSocial{}, err
	}
	token, err := s.canjear(ctx, p, codigo, verificador)
	if err != nil {
		return IngresoSocial{}, err
	}
	perfil, err := s.perfil(ctx, p, token)
	if err != nil {
		return IngresoSocial{}, err
	}
	if perfil.Sujeto == "" {
		return IngresoSocial{}, errors.New("oauth: el perfil no trae id")
	}
	return s.entrarConPerfil(ctx, proveedor, perfil, cl)
}

func (s *Servicio) entrarConPerfil(ctx context.Context, proveedor string, perfil Perfil, cl Cliente) (IngresoSocial, error) {
	var usuario string
	err := s.Pool.QueryRow(ctx, `SELECT usuario_id FROM auth.identidades WHERE proveedor = $1 AND sujeto = $2`,
		proveedor, perfil.Sujeto).Scan(&usuario)
	if err == nil {
		ses, err := s.sesionDeUsuario(ctx, usuario, SesionLarga, cl)
		return IngresoSocial{Sesion: ses}, err
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return IngresoSocial{}, err
	}
	email, err := normalizarEmail(perfil.Email)
	if err != nil {
		return IngresoSocial{}, ErrSinEmail
	}
	err = s.Pool.QueryRow(ctx, `SELECT id FROM auth.usuarios WHERE email = $1`, email).Scan(&usuario)
	switch {
	case err == nil && !perfil.Verificado:
		return IngresoSocial{}, ErrVincular
	case err == nil:
		if _, err := s.Pool.Exec(ctx, `INSERT INTO auth.identidades (proveedor, sujeto, usuario_id, email) VALUES ($1, $2, $3, $4)
			ON CONFLICT DO NOTHING`, proveedor, perfil.Sujeto, usuario, email); err != nil {
			return IngresoSocial{}, err
		}
		ses, err := s.sesionDeUsuario(ctx, usuario, SesionLarga, cl)
		return IngresoSocial{Sesion: ses}, err
	case !errors.Is(err, pgx.ErrNoRows):
		return IngresoSocial{}, err
	}
	// Cuenta nueva: sin contraseña; la tienda toma un nombre provisorio que el onboarding cambia.
	nombre := strings.TrimSpace(perfil.Nombre)
	if nombre == "" {
		nombre = strings.Split(email, "@")[0]
	}
	if len([]rune(nombre)) > 80 {
		nombre = string([]rune(nombre)[:80])
	}
	if ok, err := s.permitido(ctx, "registro:ip:"+cl.IP, 10, time.Hour); err != nil || !ok {
		return IngresoSocial{}, errOr(err, ErrDemasiadosIntentos)
	}
	var ses Sesion
	err = s.Pool.EnTx(ctx, func(tx pgx.Tx) error {
		u, tiendaID, err := crearCuenta(ctx, tx, email, nombre, nil, "La tienda de "+strings.Fields(nombre)[0])
		if err != nil {
			return err
		}
		if perfil.Verificado {
			if _, err := tx.Exec(ctx, `UPDATE auth.usuarios SET email_verificado_en = now() WHERE id = $1`, u); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, `INSERT INTO auth.identidades (proveedor, sujeto, usuario_id, email) VALUES ($1, $2, $3, $4)`,
			proveedor, perfil.Sujeto, u, email); err != nil {
			return err
		}
		ses, err = crearSesion(ctx, tx, u, tiendaID, SesionLarga, cl)
		return err
	})
	if err == nil {
		s.bienvenida(ctx, email, nombre, "La tienda de "+strings.Fields(nombre)[0])
	}
	return IngresoSocial{Sesion: ses, Nueva: err == nil}, err
}

func (s *Servicio) cliente() *http.Client {
	if s.HTTP != nil {
		return s.HTTP
	}
	return &http.Client{Timeout: 10 * time.Second}
}

func (s *Servicio) canjear(ctx context.Context, p *Proveedor, codigo, verificador string) (string, error) {
	form := url.Values{
		"grant_type": {"authorization_code"}, "code": {codigo}, "redirect_uri": {s.volverA(p.ID)},
		"client_id": {p.ClientID}, "client_secret": {p.Secreto}, "code_verifier": {verificador},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.TokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	var r struct {
		AccessToken string `json:"access_token"`
	}
	if err := s.pedirJSON(req, &r); err != nil {
		return "", fmt.Errorf("oauth %s: canje: %w", p.ID, err)
	}
	if r.AccessToken == "" {
		return "", fmt.Errorf("oauth %s: sin access_token", p.ID)
	}
	return r.AccessToken, nil
}

func (s *Servicio) perfil(ctx context.Context, p *Proveedor, token string) (Perfil, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.PerfilURL, nil)
	if err != nil {
		return Perfil{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")
	var m map[string]any
	if err := s.pedirJSON(req, &m); err != nil {
		return Perfil{}, fmt.Errorf("oauth %s: perfil: %w", p.ID, err)
	}
	return p.Perfil(m), nil
}

func (s *Servicio) pedirJSON(req *http.Request, v any) error {
	res, err := s.cliente().Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(res.Body, 64<<10))
	if res.StatusCode >= 300 {
		return fmt.Errorf("respondió %d", res.StatusCode)
	}
	return json.Unmarshal(b, v)
}
