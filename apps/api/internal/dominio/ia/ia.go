// Package ia: "Manejá Tu Local Web desde tu IA". OAuth 2.1 para los conectores MCP (registro dinámico,
// PKCE S256, código de un solo uso, acceso de 1 h y refresco que rota). El token de acceso entra a la
// API como una sesión más (cuentas.Autenticar), pero sólo a las rutas que usan las herramientas del
// MCP y, si la conexión es de lectura, sólo con GET (Guardia).
package ia

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
)

const (
	vidaCodigo  = 5 * time.Minute
	vidaAcceso  = time.Hour
	vidaCliente = 30 * 24 * time.Hour
)

var (
	ErrInvalido   = errors.New("ia: pedido inválido")
	ErrNoExiste   = errors.New("ia: no existe")
	ErrSinPermiso = errors.New("ia: sin permiso")
	// ErrConcesion es el invalid_grant de OAuth: código o refresco vencido, usado o de otro cliente.
	ErrConcesion = errors.New("ia: concesión inválida")
)

// Permisos que la tienda les da a las IA conectadas (PERMISOS_IA en apps/web). "leer" está siempre.
var PermisosDeFabrica = map[string]bool{"leer": true, "escribir": true, "responder": false, "plata": false}

type Servicio struct{ Pool *db.Pool }

type Cliente struct {
	ID            string   `json:"id"`
	Nombre        string   `json:"nombre"`
	Redirecciones []string `json:"redirecciones"`
}

// Tokens es la respuesta del endpoint de token (el BFF la traduce a los nombres de OAuth).
type Tokens struct {
	Acceso   string `json:"acceso"`
	Refresco string `json:"refresco"`
	VenceEn  int    `json:"venceEn"`
	Acciones bool   `json:"acciones"`
}

// Conexion es una IA conectada a una tienda.
type Conexion struct {
	ID       string    `json:"id"`
	Cliente  string    `json:"cliente"`
	Persona  string    `json:"persona"`
	Acciones bool      `json:"acciones"`
	CreadaEn time.Time `json:"creadaEn"`
	UsadaEn  time.Time `json:"usadaEn"`
}

// Sesion es lo que resuelve un token de acceso.
type Sesion struct {
	Conexion, UsuarioID, Nombre, Email, TiendaID, Rol string
	// Cliente es el nombre de la IA (Claude, ChatGPT…), para la Actividad.
	Cliente  string
	Acciones bool
	VenceEn  time.Time
}

func nuevoSecreto(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func hash(t string) string {
	h := sha256.Sum256([]byte(t))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// redireccionValida: https, o http sólo a localhost (los clientes de escritorio levantan un puerto).
func redireccionValida(r string) bool {
	if len(r) > 500 {
		return false
	}
	u, err := url.Parse(r)
	if err != nil || u.Host == "" || u.Fragment != "" || u.User != nil {
		return false
	}
	switch u.Scheme {
	case "https":
		return true
	case "http":
		h := u.Hostname()
		return h == "localhost" || h == "127.0.0.1" || h == "::1"
	}
	return false
}

// RegistrarCliente es el registro dinámico (RFC 7591): clientes públicos, sin secreto.
func (s *Servicio) RegistrarCliente(ctx context.Context, nombre string, redirecciones []string) (Cliente, error) {
	nombre = strings.TrimSpace(nombre)
	if nombre == "" {
		nombre = "IA sin nombre"
	}
	if len([]rune(nombre)) > 80 {
		nombre = string([]rune(nombre)[:80])
	}
	if len(redirecciones) == 0 || len(redirecciones) > 5 {
		return Cliente{}, ErrInvalido
	}
	for _, r := range redirecciones {
		if !redireccionValida(r) {
			return Cliente{}, ErrInvalido
		}
	}
	c := Cliente{ID: nuevoSecreto(18), Nombre: nombre, Redirecciones: redirecciones}
	_, err := s.Pool.Exec(ctx, `INSERT INTO auth.ia_clientes (id, nombre, redirecciones) VALUES ($1, $2, $3)`,
		c.ID, c.Nombre, c.Redirecciones)
	return c, err
}

// ClienteDe lee un cliente registrado (para la pantalla de consentimiento).
func (s *Servicio) ClienteDe(ctx context.Context, id string) (Cliente, error) {
	c := Cliente{ID: id}
	err := s.Pool.QueryRow(ctx, `SELECT nombre, redirecciones FROM auth.ia_clientes WHERE id = $1`, id).
		Scan(&c.Nombre, &c.Redirecciones)
	if errors.Is(err, pgx.ErrNoRows) {
		return c, ErrNoExiste
	}
	return c, err
}

var desafioValido = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)

// Autorizar emite el código después del "Permitir". Sólo dueña o admin conectan una IA.
func (s *Servicio) Autorizar(ctx context.Context, usuario, tienda, rol, cliente, redireccion, desafio string, acciones bool) (string, error) {
	if rol != "duenio" && rol != "admin" {
		return "", ErrSinPermiso
	}
	if tienda == "" || !desafioValido.MatchString(desafio) {
		return "", ErrInvalido
	}
	c, err := s.ClienteDe(ctx, cliente)
	if err != nil {
		return "", err
	}
	ok := false
	for _, r := range c.Redirecciones {
		ok = ok || r == redireccion
	}
	if !ok {
		return "", ErrInvalido
	}
	codigo := nuevoSecreto(32)
	err = s.Pool.EnTx(ctx, func(t pgx.Tx) error {
		if _, err := t.Exec(ctx, `DELETE FROM auth.ia_codigos WHERE vence_en < now()`); err != nil {
			return err
		}
		if _, err := t.Exec(ctx, `INSERT INTO auth.ia_codigos (hash, cliente_id, usuario_id, tienda_id, redireccion, desafio, acciones, vence_en)
			VALUES ($1, $2, $3, $4, $5, $6, $7, now() + $8::interval)`,
			hash(codigo), cliente, usuario, tienda, redireccion, desafio, acciones, vidaCodigo.String()); err != nil {
			return err
		}
		_, err := t.Exec(ctx, `UPDATE auth.ia_clientes SET usado_en = now() WHERE id = $1`, cliente)
		return err
	})
	return codigo, err
}

// Canjear cambia el código (un solo uso) por los tokens, verificando PKCE, cliente y redirección.
func (s *Servicio) Canjear(ctx context.Context, codigo, verificador, cliente, redireccion string) (Tokens, error) {
	var tk Tokens
	err := s.Pool.EnTx(ctx, func(t pgx.Tx) error {
		var cli, usuario, tienda, red, desafio string
		var acciones bool
		err := t.QueryRow(ctx, `DELETE FROM auth.ia_codigos WHERE hash = $1 AND vence_en > now()
			RETURNING cliente_id, usuario_id::text, tienda_id::text, redireccion, desafio, acciones`, hash(codigo)).
			Scan(&cli, &usuario, &tienda, &red, &desafio, &acciones)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrConcesion
		}
		if err != nil {
			return err
		}
		h := sha256.Sum256([]byte(verificador))
		esperado := base64.RawURLEncoding.EncodeToString(h[:])
		if cli != cliente || red != redireccion || len(verificador) < 43 || len(verificador) > 128 ||
			subtle.ConstantTimeCompare([]byte(esperado), []byte(desafio)) != 1 {
			return ErrConcesion
		}
		tk = Tokens{Acceso: nuevoSecreto(32), Refresco: nuevoSecreto(32), VenceEn: int(vidaAcceso.Seconds()), Acciones: acciones}
		_, err = t.Exec(ctx, `INSERT INTO auth.ia_conexiones (cliente_id, usuario_id, tienda_id, acciones, acceso_hash, acceso_vence, refresco_hash)
			VALUES ($1, $2, $3, $4, $5, now() + $6::interval, $7)`,
			cli, usuario, tienda, acciones, hash(tk.Acceso), vidaAcceso.String(), hash(tk.Refresco))
		return err
	})
	return tk, err
}

// Refrescar rota el refresco. Si llega uno que ya se usó, alguien lo copió: se corta la conexión.
func (s *Servicio) Refrescar(ctx context.Context, refresco, cliente string) (Tokens, error) {
	var tk Tokens
	err := s.Pool.EnTx(ctx, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `DELETE FROM auth.ia_conexiones WHERE refresco_previo = $1`, hash(refresco))
		if err != nil {
			return err
		}
		if tag.RowsAffected() > 0 {
			return ErrConcesion
		}
		tk = Tokens{Acceso: nuevoSecreto(32), Refresco: nuevoSecreto(32), VenceEn: int(vidaAcceso.Seconds())}
		err = t.QueryRow(ctx, `UPDATE auth.ia_conexiones SET acceso_hash = $1, acceso_vence = now() + $2::interval,
				refresco_previo = refresco_hash, refresco_hash = $3, usada_en = now()
			WHERE refresco_hash = $4 AND cliente_id = $5 RETURNING acciones`,
			hash(tk.Acceso), vidaAcceso.String(), hash(tk.Refresco), hash(refresco), cliente).Scan(&tk.Acciones)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrConcesion
		}
		return err
	})
	if errors.Is(err, ErrConcesion) {
		// El borrado de la conexión robada tiene que quedar aunque el pedido falle.
		_, _ = s.Pool.Exec(ctx, `DELETE FROM auth.ia_conexiones WHERE refresco_previo = $1`, hash(refresco))
	}
	return tk, err
}

// Revocar corta la conexión del token (de acceso o de refresco). Nunca dice si existía (RFC 7009).
func (s *Servicio) Revocar(ctx context.Context, token string) error {
	h := hash(token)
	_, err := s.Pool.Exec(ctx, `DELETE FROM auth.ia_conexiones WHERE acceso_hash = $1 OR refresco_hash = $1`, h)
	return err
}

// SesionDe resuelve un token de acceso vigente. La persona tiene que seguir en la tienda.
func (s *Servicio) SesionDe(ctx context.Context, token string) (Sesion, error) {
	var se Sesion
	var usada time.Time
	err := s.Pool.QueryRow(ctx, `
		SELECT c.id::text, u.id::text, u.nombre, u.email, c.tienda_id::text, m.rol, cl.nombre, c.acciones, c.acceso_vence, c.usada_en
		FROM auth.ia_conexiones c
		JOIN auth.ia_clientes cl ON cl.id = c.cliente_id
		JOIN auth.usuarios u ON u.id = c.usuario_id
		JOIN auth.membresias m ON m.usuario_id = c.usuario_id AND m.tienda_id = c.tienda_id
		WHERE c.acceso_hash = $1 AND c.acceso_vence > now()`, hash(token)).
		Scan(&se.Conexion, &se.UsuarioID, &se.Nombre, &se.Email, &se.TiendaID, &se.Rol, &se.Cliente, &se.Acciones, &se.VenceEn, &usada)
	if errors.Is(err, pgx.ErrNoRows) {
		return se, ErrNoExiste
	}
	if err == nil && time.Since(usada) > time.Minute {
		_, _ = s.Pool.Exec(ctx, `UPDATE auth.ia_conexiones SET usada_en = now() WHERE id = $1`, se.Conexion)
	}
	return se, err
}

// Panel: permisos y conexiones de la tienda.
type Panel struct {
	Permisos   map[string]bool `json:"permisos"`
	Conexiones []Conexion      `json:"conexiones"`
}

func (s *Servicio) Panel(ctx context.Context, tienda string) (Panel, error) {
	p := Panel{Permisos: map[string]bool{}, Conexiones: []Conexion{}}
	var crudo []byte
	if err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT ia_permisos FROM tiendas WHERE id = $1`, tienda).Scan(&crudo)
	}); err != nil {
		return p, err
	}
	guardados := map[string]bool{}
	_ = json.Unmarshal(crudo, &guardados)
	for k, def := range PermisosDeFabrica {
		v, ok := guardados[k]
		p.Permisos[k] = (ok && v) || (!ok && def) || k == "leer"
	}
	filas, err := s.Pool.Query(ctx, `SELECT c.id::text, cl.nombre, u.nombre, c.acciones, c.creada_en, c.usada_en
		FROM auth.ia_conexiones c JOIN auth.ia_clientes cl ON cl.id = c.cliente_id JOIN auth.usuarios u ON u.id = c.usuario_id
		WHERE c.tienda_id = $1 ORDER BY c.creada_en DESC`, tienda)
	if err != nil {
		return p, err
	}
	defer filas.Close()
	for filas.Next() {
		var c Conexion
		if err := filas.Scan(&c.ID, &c.Cliente, &c.Persona, &c.Acciones, &c.CreadaEn, &c.UsadaEn); err != nil {
			return p, err
		}
		p.Conexiones = append(p.Conexiones, c)
	}
	return p, filas.Err()
}

// FijarPermiso prende o apaga un permiso ("leer" no se apaga). Dueña o admin.
func (s *Servicio) FijarPermiso(ctx context.Context, tienda, rol, permiso string, valor bool) error {
	if rol != "duenio" && rol != "admin" {
		return ErrSinPermiso
	}
	if _, ok := PermisosDeFabrica[permiso]; !ok || permiso == "leer" {
		return ErrInvalido
	}
	return s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE tiendas SET ia_permisos = ia_permisos || jsonb_build_object($2::text, $3::boolean) WHERE id = $1`,
			tienda, permiso, valor)
		return err
	})
}

// Desconectar corta una conexión de la tienda (sólo de esa tienda).
func (s *Servicio) Desconectar(ctx context.Context, tienda, rol, id string) error {
	if rol != "duenio" && rol != "admin" {
		return ErrSinPermiso
	}
	tag, err := s.Pool.Exec(ctx, `DELETE FROM auth.ia_conexiones WHERE tienda_id = $1 AND id::text = $2`, tienda, id)
	if err == nil && tag.RowsAffected() == 0 {
		return ErrNoExiste
	}
	return err
}

// Purgar borra códigos vencidos y clientes que nadie usa hace 30 días (y sin conexiones).
func (s *Servicio) Purgar(ctx context.Context) error {
	_, err := s.Pool.Exec(ctx, `DELETE FROM auth.ia_codigos WHERE vence_en < now()`)
	if err != nil {
		return err
	}
	_, err = s.Pool.Exec(ctx, `DELETE FROM auth.ia_clientes cl WHERE coalesce(cl.usado_en, cl.creado_en) < now() - $1::interval
		AND NOT EXISTS (SELECT 1 FROM auth.ia_conexiones c WHERE c.cliente_id = cl.id)`, vidaCliente.String())
	return err
}

// ParaCuentas adapta SesionDe a cuentas.Servicio.SesionIA.
func (s *Servicio) ParaCuentas(ctx context.Context, token string) (cuentas.Datos, error) {
	var d cuentas.Datos
	se, err := s.SesionDe(ctx, token)
	if errors.Is(err, ErrNoExiste) {
		return d, cuentas.ErrSinSesion
	}
	if err != nil {
		return d, err
	}
	d.Usuario.ID, d.Usuario.Nombre, d.Usuario.Email = se.UsuarioID, se.Nombre, se.Email
	d.TiendaID, d.Rol, d.VenceEn = se.TiendaID, se.Rol, se.VenceEn
	d.IA, d.IAAcciones, d.IANombre = se.Conexion, se.Acciones, se.Cliente
	d.SesionID = "ia-" + se.Conexion[:8]
	return d, nil
}
