package ia

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/limite"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

func escribir(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(v)
}

func leer(w http.ResponseWriter, r *http.Request, v any) bool {
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(v) != nil {
		problema.Escribir(w, problema.Invalido("El pedido no es válido."))
		return false
	}
	return true
}

func escribirError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalido):
		problema.Escribir(w, problema.Invalido("El pedido no es válido."))
	case errors.Is(err, ErrNoExiste):
		problema.Escribir(w, problema.NoEncontrado("No existe."))
	case errors.Is(err, ErrSinPermiso):
		problema.Escribir(w, problema.SinPermiso("Sólo quien es dueña o admin de la tienda puede conectar una IA."))
	case errors.Is(err, ErrConcesion):
		problema.Escribir(w, &problema.Problema{Estado: http.StatusBadRequest, Codigo: "invalid_grant",
			Titulo: "Concesión inválida", Detalle: "El código o el token venció, ya se usó o es de otro cliente."})
	default:
		slog.Error("ia", "error", err)
		problema.Escribir(w, problema.Interno())
	}
}

// Rutas: las de OAuth las llama el BFF (que habla con las IA); las del panel, con la sesión de la tienda.
func (s *Servicio) Rutas(mux *http.ServeMux, ip func(*http.Request) string) {
	mux.HandleFunc("POST /v1/ia/clientes", func(w http.ResponseWriter, r *http.Request) {
		ok, err := limite.Permitido(r.Context(), s.Pool, "ia:registro:"+ip(r), 20, time.Hour)
		if err != nil {
			// Una caída de la base no es un límite: se loguea y es un 500.
			escribirError(w, err)
			return
		}
		if !ok {
			w.Header().Set("Retry-After", "3600")
			problema.Escribir(w, &problema.Problema{Estado: http.StatusTooManyRequests, Codigo: "demasiados",
				Titulo: "Demasiados registros", Detalle: "Probá de nuevo en una hora."})
			return
		}
		var a struct {
			Nombre        string   `json:"nombre"`
			Redirecciones []string `json:"redirecciones"`
		}
		if !leer(w, r, &a) {
			return
		}
		c, err := s.RegistrarCliente(r.Context(), a.Nombre, a.Redirecciones)
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusCreated, c)
	})
	mux.HandleFunc("GET /v1/ia/clientes/{id}", func(w http.ResponseWriter, r *http.Request) {
		c, err := s.ClienteDe(r.Context(), r.PathValue("id"))
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusOK, c)
	})
	mux.HandleFunc("POST /v1/ia/autorizar", func(w http.ResponseWriter, r *http.Request) {
		d, ok := cuentas.SesionDe(r)
		if !ok || d.IA != "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		var a struct {
			Cliente     string `json:"cliente"`
			Redireccion string `json:"redireccion"`
			Desafio     string `json:"desafio"`
			Acciones    bool   `json:"acciones"`
		}
		if !leer(w, r, &a) {
			return
		}
		codigo, err := s.Autorizar(r.Context(), d.Usuario.ID, d.TiendaID, d.Rol, a.Cliente, a.Redireccion, a.Desafio, a.Acciones)
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusOK, map[string]string{"codigo": codigo})
	})
	mux.HandleFunc("POST /v1/ia/token", func(w http.ResponseWriter, r *http.Request) {
		var a struct {
			Tipo        string `json:"tipo"`
			Codigo      string `json:"codigo"`
			Verificador string `json:"verificador"`
			Cliente     string `json:"cliente"`
			Redireccion string `json:"redireccion"`
			Refresco    string `json:"refresco"`
		}
		if !leer(w, r, &a) {
			return
		}
		var tk Tokens
		var err error
		switch a.Tipo {
		case "authorization_code":
			tk, err = s.Canjear(r.Context(), a.Codigo, a.Verificador, a.Cliente, a.Redireccion)
		case "refresh_token":
			tk, err = s.Refrescar(r.Context(), a.Refresco, a.Cliente)
		default:
			err = ErrInvalido
		}
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusOK, tk)
	})
	mux.HandleFunc("POST /v1/ia/revocar", func(w http.ResponseWriter, r *http.Request) {
		var a struct {
			Token string `json:"token"`
		}
		if !leer(w, r, &a) {
			return
		}
		if err := s.Revocar(r.Context(), a.Token); err != nil {
			escribirError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	// Lo que sabe el MCP de su propia conexión (para decir qué puede hacer antes de intentarlo).
	mux.HandleFunc("GET /v1/ia/yo", func(w http.ResponseWriter, r *http.Request) {
		d, ok := cuentas.SesionDe(r)
		if !ok || d.IA == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		p, err := s.Panel(r.Context(), d.TiendaID)
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusOK, map[string]any{"conexion": d.IA, "acciones": d.IAAcciones, "permisos": p.Permisos})
	})

	panel := func(f func(w http.ResponseWriter, r *http.Request, d cuentas.Datos)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			d, ok := cuentas.SesionDe(r)
			if !ok || d.TiendaID == "" || d.IA != "" {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			f(w, r, d)
		}
	}
	mux.HandleFunc("GET /v1/panel/ia", panel(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		p, err := s.Panel(r.Context(), d.TiendaID)
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusOK, p)
	}))
	mux.HandleFunc("PUT /v1/panel/ia/permisos/{permiso}", panel(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		var a struct {
			Valor bool `json:"valor"`
		}
		if !leer(w, r, &a) {
			return
		}
		if err := s.FijarPermiso(r.Context(), d.TiendaID, d.Rol, r.PathValue("permiso"), a.Valor); err != nil {
			escribirError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	mux.HandleFunc("DELETE /v1/panel/ia/conexiones/{id}", panel(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		if err := s.Desconectar(r.Context(), d.TiendaID, d.Rol, r.PathValue("id")); err != nil {
			escribirError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
}

// Lo que puede pedir una IA: las rutas que usan las herramientas del MCP (apps/web
// lib/server/mcp/herramientas.ts por medio de DatosApi). Todo lo demás (cuenta, usuarios, plan,
// sesiones, integraciones de cobro…) queda afuera aunque la persona sea dueña.
var (
	lecturas = regexp.MustCompile(`^/v1/(ia/yo|panel/(tablero|meta|informes|cupones|pedidos(/[0-9]+)?|productos(/[^/]+)?|listas/(productos|clientes)|integraciones/(mercadopago|mercadolibre)))$`)
	acciones = map[string]*regexp.Regexp{
		http.MethodPost: regexp.MustCompile(`^/v1/panel/(pedidos/[0-9]+/estado|cupones|links)$`),
		http.MethodPut:  regexp.MustCompile(`^/v1/panel/productos/[^/]+$`),
	}
)

// permisoPanel: qué interruptor del panel necesita una acción ("" si alcanza con el alcance).
func permisoPanel(metodo, ruta string) string {
	switch {
	case metodo == http.MethodPut && strings.HasPrefix(ruta, "/v1/panel/productos/"):
		return "escribir"
	case metodo == http.MethodPost && ruta == "/v1/panel/cupones":
		return "plata"
	}
	return ""
}

// Guardia limita lo que hace un token de IA: rutas permitidas, lectura vs acciones y 60 pedidos
// por minuto por conexión. A las sesiones normales no las toca.
func (s *Servicio) Guardia(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		d, ok := cuentas.SesionDe(r)
		if !ok || d.IA == "" {
			next.ServeHTTP(w, r)
			return
		}
		ruta := strings.TrimSuffix(r.URL.Path, "/")
		permitido := false
		switch r.Method {
		case http.MethodGet:
			permitido = lecturas.MatchString(ruta)
		default:
			if re, hay := acciones[r.Method]; hay {
				permitido = d.IAAcciones && re.MatchString(ruta)
			}
		}
		if !permitido {
			problema.Escribir(w, problema.SinPermiso("Esta conexión de IA no puede hacer eso."))
			return
		}
		// Además del alcance, los interruptores del panel: editar productos y tocar plata.
		if permiso := permisoPanel(r.Method, ruta); permiso != "" {
			p, err := s.Panel(r.Context(), d.TiendaID)
			if err != nil {
				escribirError(w, err)
				return
			}
			if !p.Permisos[permiso] {
				problema.Escribir(w, problema.SinPermiso("En Mostrador › Integraciones está apagado este permiso para las IA."))
				return
			}
		}
		ok, err := limite.Permitido(r.Context(), s.Pool, "ia:conexion:"+d.IA, 60, time.Minute)
		if err != nil {
			escribirError(w, err)
			return
		}
		if !ok {
			w.Header().Set("Retry-After", "60")
			problema.Escribir(w, &problema.Problema{Estado: http.StatusTooManyRequests, Codigo: "demasiados",
				Titulo: "Demasiados pedidos", Detalle: "La IA hizo más de 60 pedidos en un minuto. Esperá un poco."})
			return
		}
		next.ServeHTTP(w, r)
	})
}
