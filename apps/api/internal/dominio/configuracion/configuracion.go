// Package configuracion: la pantalla Configuración (design/PanelOperacion.dc.html). Negocio,
// equipo (membresías e invitaciones), seguridad y avisos, sesiones de la persona y plan.
package configuracion

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}

var (
	ErrNoExiste    = errors.New("no existe")
	ErrRepetido    = errors.New("ya está en el equipo")
	ErrNoSePuede   = errors.New("no se puede")
	ErrInvalido    = errors.New("inválido")
	seguridadDef   = map[string]bool{"dobleFactor": true, "otroPais": true, "alertas": true, "inactivas": false}
	avisosDef      = map[string]bool{"venta": true, "chats": true, "stock": true, "diario": true, "ideas": true, "novedades": false}
	planes         = []string{"gratis", "emprende", "pro", "negocio"}
	contactoValido = regexp.MustCompile(`^([^\s@]+@[^\s@]+\.[^\s@]+|\+?[\d\s-]{8,20})$`)
)

// Negocio es components/schemas/Negocio.
type Negocio struct {
	Nombre      string `json:"nombre"`
	RazonSocial string `json:"razonSocial"`
	Cuit        string `json:"cuit"`
	Condicion   string `json:"condicion"`
	Direccion   string `json:"direccion"`
	Whatsapp    string `json:"whatsapp"`
	Email       string `json:"email"`
	Horario     string `json:"horario"`
}

// Miembro es components/schemas/Miembro.
type Miembro struct {
	ID          string     `json:"id"`
	Nombre      string     `json:"nombre"`
	Contacto    string     `json:"contacto"`
	Rol         string     `json:"rol"`
	Foto        *string    `json:"foto"`
	Ultimo      *time.Time `json:"ultimo"`
	Pendiente   bool       `json:"pendiente"`
	Vos         bool       `json:"vos"`
	DobleFactor bool       `json:"dobleFactor"`
}

type Sesion struct {
	ID     string    `json:"id"`
	Tipo   string    `json:"tipo"`
	Nombre string    `json:"nombre"`
	Donde  string    `json:"donde"`
	Cuando time.Time `json:"cuando"`
	Actual bool      `json:"actual"`
}

// Quien es la persona que mira (para "vos" y sus sesiones).
type Quien struct{ UsuarioID, SesionID string }

func conDefecto(def, v map[string]bool) map[string]bool {
	out := map[string]bool{}
	for k, d := range def {
		out[k] = d
		if x, ok := v[k]; ok {
			out[k] = x
		}
	}
	return out
}

// dispositivo: un nombre legible a partir del navegador.
func dispositivo(agente string) (string, string) {
	a := strings.ToLower(agente)
	tipo := "compu"
	if strings.Contains(a, "iphone") || strings.Contains(a, "android") || strings.Contains(a, "mobile") {
		tipo = "celu"
	}
	nav := "Navegador"
	switch {
	case strings.Contains(a, "edg/"):
		nav = "Edge"
	case strings.Contains(a, "chrome"):
		nav = "Chrome"
	case strings.Contains(a, "firefox"):
		nav = "Firefox"
	case strings.Contains(a, "safari"):
		nav = "Safari"
	}
	so := ""
	switch {
	case strings.Contains(a, "iphone"):
		so = " en iPhone"
	case strings.Contains(a, "android"):
		so = " en Android"
	case strings.Contains(a, "mac os"):
		so = " en Mac"
	case strings.Contains(a, "windows"):
		so = " en Windows"
	}
	return tipo, nav + so
}

// Leer arma toda la configuración.
func Leer(ctx context.Context, pool tx, tiendaID string, q Quien, ahora time.Time) (map[string]any, error) {
	var res map[string]any
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var nombre, plan, ciclo string
		var negocio Negocio
		var prefs struct {
			Seguridad map[string]bool `json:"seguridad"`
			Avisos    map[string]bool `json:"avisos"`
		}
		var creada time.Time
		if err := t.QueryRow(ctx, `SELECT nombre, negocio, preferencias, plan, plan_ciclo, creada_en FROM tiendas WHERE id = $1`, tiendaID).
			Scan(&nombre, &negocio, &prefs, &plan, &ciclo, &creada); err != nil {
			return err
		}
		negocio.Nombre = nombre
		equipo := []Miembro{}
		filas, err := t.Query(ctx, `SELECT u.id::text, u.nombre, u.email, m.rol,
				(SELECT max(visto_en) FROM auth.sesiones s WHERE s.usuario_id = u.id)
			FROM auth.membresias m JOIN auth.usuarios u ON u.id = m.usuario_id
			WHERE m.tienda_id = $1 ORDER BY m.creada_en`, tiendaID)
		if err != nil {
			return err
		}
		for filas.Next() {
			var m Miembro
			if err := filas.Scan(&m.ID, &m.Nombre, &m.Contacto, &m.Rol, &m.Ultimo); err != nil {
				return err
			}
			m.Vos = m.ID == q.UsuarioID
			equipo = append(equipo, m)
		}
		if err := filas.Err(); err != nil {
			return err
		}
		inv, err := t.Query(ctx, `SELECT id::text, contacto, rol FROM invitaciones WHERE tienda_id = $1 ORDER BY creada_en`, tiendaID)
		if err != nil {
			return err
		}
		for inv.Next() {
			var m Miembro
			if err := inv.Scan(&m.ID, &m.Contacto, &m.Rol); err != nil {
				return err
			}
			m.Nombre, m.Pendiente = m.Contacto, true
			equipo = append(equipo, m)
		}
		if err := inv.Err(); err != nil {
			return err
		}
		sesiones := []Sesion{}
		ss, err := t.Query(ctx, `SELECT left(token_hash, 16), coalesce(agente, ''), visto_en FROM auth.sesiones
			WHERE usuario_id::text = $1 AND vence_en > now() ORDER BY visto_en DESC LIMIT 20`, q.UsuarioID)
		if err != nil {
			return err
		}
		for ss.Next() {
			var s Sesion
			var agente string
			if err := ss.Scan(&s.ID, &agente, &s.Cuando); err != nil {
				return err
			}
			s.Tipo, s.Nombre = dispositivo(agente)
			s.Donde, s.Actual = "Argentina", s.ID == q.SesionID
			sesiones = append(sesiones, s)
		}
		if err := ss.Err(); err != nil {
			return err
		}
		prueba := creada.Add(14 * 24 * time.Hour)
		res = map[string]any{
			"negocio":   negocio,
			"equipo":    equipo,
			"seguridad": conDefecto(seguridadDef, prefs.Seguridad),
			"sesiones":  sesiones,
			"avisos":    conDefecto(avisosDef, prefs.Avisos),
			"plan": map[string]any{"actual": plan, "ciclo": ciclo, "enPrueba": ahora.Before(prueba),
				"diasPrueba": max(0, int(prueba.Sub(ahora).Hours()/24))},
		}
		return ss.Err()
	})
	return res, err
}

// GuardarNegocio guarda los datos (el nombre va en la columna de la tienda).
func GuardarNegocio(ctx context.Context, pool tx, tiendaID string, n Negocio) error {
	n.Nombre = strings.TrimSpace(n.Nombre)
	if n.Nombre == "" || len(n.Nombre) > 80 {
		return ErrInvalido
	}
	resto := n
	resto.Nombre = ""
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE tiendas SET nombre = $2, negocio = $3 WHERE id = $1`, tiendaID, n.Nombre, resto)
		return err
	})
}

// FijarOpcion prende o apaga una opción de seguridad o un aviso.
func FijarOpcion(ctx context.Context, pool tx, tiendaID, grupo, clave string, activo bool) error {
	def := map[string]map[string]bool{"seguridad": seguridadDef, "avisos": avisosDef}[grupo]
	if _, ok := def[clave]; !ok {
		return ErrNoExiste
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE tiendas SET preferencias = jsonb_set(preferencias, ARRAY[$2::text],
			coalesce(preferencias->$2, '{}') || jsonb_build_object($3::text, $4::boolean)) WHERE id = $1`, tiendaID, grupo, clave, activo)
		return err
	})
}

// CerrarSesion cierra otra sesión de la misma persona.
func CerrarSesion(ctx context.Context, pool tx, tiendaID string, q Quien, id string) error {
	if id == q.SesionID {
		return ErrNoSePuede
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `DELETE FROM auth.sesiones WHERE usuario_id::text = $1 AND left(token_hash, 16) = $2`, q.UsuarioID, id)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrNoExiste
		}
		return err
	})
}

// Invitar suma una invitación pendiente. Mandarla por email o WhatsApp llega con esos servicios.
func Invitar(ctx context.Context, pool tx, tiendaID, contacto, rol string) (Miembro, error) {
	contacto = strings.TrimSpace(contacto)
	m := Miembro{Nombre: contacto, Contacto: contacto, Rol: rol, Pendiente: true}
	if !contactoValido.MatchString(contacto) || !slices.Contains([]string{"admin", "ventas", "deposito"}, rol) {
		return m, ErrInvalido
	}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var ya bool
		if err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM auth.membresias mm JOIN auth.usuarios u ON u.id = mm.usuario_id
			WHERE mm.tienda_id = $1 AND lower(u.email) = lower($2))`, tiendaID, contacto).Scan(&ya); err != nil {
			return err
		}
		if ya {
			return ErrRepetido
		}
		err := t.QueryRow(ctx, `INSERT INTO invitaciones (tienda_id, contacto, rol) VALUES ($1, $2, $3)
			ON CONFLICT (tienda_id, contacto) DO NOTHING RETURNING id::text`, tiendaID, contacto, rol).Scan(&m.ID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRepetido
		}
		return err
	})
	return m, err
}

// CambiarPlan guarda el plan elegido. El cobro llega con Mercado Pago Suscripciones.
func CambiarPlan(ctx context.Context, pool tx, tiendaID, plan, ciclo string) error {
	if !slices.Contains(planes, plan) || (ciclo != "" && ciclo != "mes" && ciclo != "anio") {
		return ErrInvalido
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE tiendas SET plan = $2, plan_ciclo = coalesce(nullif($3, ''), plan_ciclo) WHERE id = $1`, tiendaID, plan, ciclo)
		return err
	})
}

// Rutas del panel.
func Rutas(mux *http.ServeMux, pool tx, sitio string) {
	con := func(h func(http.ResponseWriter, *http.Request, string, Quien)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			d, ok := cuentas.SesionDe(r)
			if !ok || d.TiendaID == "" {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			h(w, r, d.TiendaID, Quien{UsuarioID: d.Usuario.ID, SesionID: d.SesionID})
		}
	}
	errorDe := func(w http.ResponseWriter, err error) {
		switch {
		case err == nil:
			w.WriteHeader(http.StatusNoContent)
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("No existe."))
		case errors.Is(err, ErrRepetido):
			problema.Escribir(w, problema.Conflicto("repetido", "Esa persona ya está en el equipo."))
		case errors.Is(err, ErrNoSePuede):
			problema.Escribir(w, problema.Conflicto("actual", "Para cerrar esta, salí de la cuenta."))
		case errors.Is(err, ErrInvalido):
			problema.Escribir(w, problema.Invalido("Revisá los datos."))
		default:
			problema.Escribir(w, problema.Interno())
		}
	}
	leer := func(w http.ResponseWriter, r *http.Request, v any) bool {
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(v); err != nil {
			problema.Escribir(w, problema.Invalido("Revisá los datos."))
			return false
		}
		return true
	}
	rutasCuenta(mux, pool, con, errorDe, leer)
	rutasDatos(mux, pool, sitio, con, errorDe, leer)
	mux.HandleFunc("GET /v1/panel/configuracion", con(func(w http.ResponseWriter, r *http.Request, id string, q Quien) {
		v, err := Leer(r.Context(), pool, id, q, time.Now())
		if err != nil {
			errorDe(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}))
	mux.HandleFunc("PUT /v1/panel/configuracion/negocio", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		var n Negocio
		if leer(w, r, &n) {
			errorDe(w, GuardarNegocio(r.Context(), pool, id, n))
		}
	}))
	mux.HandleFunc("PUT /v1/panel/configuracion/{grupo}/{clave}", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		var b struct {
			Activo bool `json:"activo"`
		}
		if leer(w, r, &b) {
			errorDe(w, FijarOpcion(r.Context(), pool, id, r.PathValue("grupo"), r.PathValue("clave"), b.Activo))
		}
	}))
	mux.HandleFunc("DELETE /v1/panel/sesiones/{id}", con(func(w http.ResponseWriter, r *http.Request, id string, q Quien) {
		errorDe(w, CerrarSesion(r.Context(), pool, id, q, r.PathValue("id")))
	}))
	mux.HandleFunc("POST /v1/panel/equipo", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		var b struct{ Contacto, Rol string }
		if !leer(w, r, &b) {
			return
		}
		m, err := Invitar(r.Context(), pool, id, b.Contacto, b.Rol)
		if err != nil {
			errorDe(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(m)
	}))
	mux.HandleFunc("PUT /v1/panel/plan", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		var b struct{ Plan, Ciclo string }
		if leer(w, r, &b) {
			errorDe(w, CambiarPlan(r.Context(), pool, id, b.Plan, b.Ciclo))
		}
	}))
}
