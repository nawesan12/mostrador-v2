package equipo

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

func token(r *http.Request) string {
	if t, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
		return strings.TrimSpace(t)
	}
	return ""
}

// SesionDe la respuesta de /v1/equipo/sesion (components/schemas/SesionEquipo).
type SesionDe struct {
	Miembro  Miembro  `json:"miembro"`
	Rol      string   `json:"rol"`
	Permisos []string `json:"permisos"`
	VenceEn  string   `json:"venceEn"`
}

// SoloLectura frena las escrituras de una mirada del equipo al panel de una tienda.
func SoloLectura(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if d, ok := cuentas.SesionDe(r); ok && d.Mirada && r.Method != http.MethodGet && r.Method != http.MethodHead {
			problema.Escribir(w, problema.SinPermiso("Estás viendo la tienda en modo lectura: no se puede cambiar nada."))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Rutas del backoffice (/v1/equipo/*). Cada una pide la sesión del equipo y su permiso.
func (s *Servicio) Rutas(mux *http.ServeMux) {
	cliente := func(r *http.Request) cuentas.Cliente { return cuentas.ClienteDe(r, s.ClaveBFF) }
	// con exige la sesión del equipo (nunca la de una tienda) y el permiso ("" = cualquiera del equipo).
	con := func(permiso string, fn func(http.ResponseWriter, *http.Request, Miembro)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			m, err := s.Sesion(r.Context(), token(r))
			if errors.Is(err, ErrNoExiste) {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			if err != nil {
				escribirError(w, err)
				return
			}
			if permiso != "" && !Puede(m.Rol, permiso) {
				escribirError(w, ErrSinPermiso)
				return
			}
			w.Header().Set("Cache-Control", "no-store")
			fn(w, r, m)
		}
	}
	responder := func(w http.ResponseWriter, estado int, v any, err error) {
		if err != nil {
			escribirError(w, err)
			return
		}
		if v == nil {
			w.WriteHeader(estado)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(estado)
		_ = json.NewEncoder(w).Encode(v)
	}

	mux.HandleFunc("POST /v1/equipo/ingresar", func(w http.ResponseWriter, r *http.Request) {
		var a struct{ Email, Clave string }
		if leer(w, r, &a) {
			ses, err := s.Ingresar(r.Context(), a.Email, a.Clave, cliente(r))
			responder(w, http.StatusOK, ses, err)
		}
	})
	mux.HandleFunc("POST /v1/equipo/salir", func(w http.ResponseWriter, r *http.Request) {
		if t := token(r); t != "" {
			if err := s.Salir(r.Context(), t); err != nil {
				escribirError(w, err)
				return
			}
		}
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("GET /v1/equipo/sesion", con("", func(w http.ResponseWriter, r *http.Request, m Miembro) {
		responder(w, http.StatusOK, SesionDe{Miembro: m, Rol: m.Rol, Permisos: Permisos[m.Rol],
			VenceEn: m.VenceEn.Format("2006-01-02T15:04:05Z07:00")}, nil)
	}))
	mux.HandleFunc("GET /v1/equipo/inicio", con("tiendas.ver", func(w http.ResponseWriter, r *http.Request, _ Miembro) {
		v, err := s.Inicio(r.Context())
		responder(w, http.StatusOK, v, err)
	}))
	mux.HandleFunc("GET /v1/equipo/tiendas", con("tiendas.ver", func(w http.ResponseWriter, r *http.Request, _ Miembro) {
		v, err := s.Tiendas(r.Context(), listas.FiltrosDe(r.URL.Query()))
		responder(w, http.StatusOK, v, err)
	}))
	mux.HandleFunc("GET /v1/equipo/tiendas/{id}", con("tiendas.ver", func(w http.ResponseWriter, r *http.Request, m Miembro) {
		v, err := s.Ficha(r.Context(), m, r.PathValue("id"), cliente(r))
		responder(w, http.StatusOK, v, err)
	}))
	mux.HandleFunc("POST /v1/equipo/tiendas/{id}/acciones", con("tiendas.ver", func(w http.ResponseWriter, r *http.Request, m Miembro) {
		var a Accion
		if leer(w, r, &a) {
			responder(w, http.StatusNoContent, nil, s.Aplicar(r.Context(), m, r.PathValue("id"), a, cliente(r)))
		}
	}))
	mux.HandleFunc("POST /v1/equipo/tiendas/{id}/notas", con("tiendas.notas", func(w http.ResponseWriter, r *http.Request, m Miembro) {
		var a struct{ Texto string }
		if leer(w, r, &a) {
			v, err := s.Notar(r.Context(), m, r.PathValue("id"), strings.TrimSpace(a.Texto), cliente(r))
			responder(w, http.StatusCreated, v, err)
		}
	}))
	mux.HandleFunc("POST /v1/equipo/tiendas/{id}/mirar", con("tiendas.mirar", func(w http.ResponseWriter, r *http.Request, m Miembro) {
		v, err := s.Mirar(r.Context(), m, r.PathValue("id"), cliente(r))
		responder(w, http.StatusOK, v, err)
	}))
	mux.HandleFunc("GET /v1/equipo/consultas", con("consultas.ver", func(w http.ResponseWriter, r *http.Request, _ Miembro) {
		v, err := s.Consultas(r.Context())
		responder(w, http.StatusOK, v, err)
	}))
	mux.HandleFunc("GET /v1/equipo/consultas/{tipo}/{id}", con("consultas.ver", func(w http.ResponseWriter, r *http.Request, m Miembro) {
		v, err := s.Consulta(r.Context(), m, r.PathValue("tipo"), r.PathValue("id"), cliente(r))
		responder(w, http.StatusOK, v, err)
	}))
	mux.HandleFunc("POST /v1/equipo/consultas/{tipo}/{id}", con("consultas.responder", func(w http.ResponseWriter, r *http.Request, m Miembro) {
		var a map[string]json.RawMessage
		if !leer(w, r, &a) {
			return
		}
		var c Cambio
		if e, ok := a["estado"]; ok {
			var v string
			if json.Unmarshal(e, &v) != nil {
				escribirError(w, ErrInvalido{"estado", "El estado es nueva, en_curso o resuelta."})
				return
			}
			c.Estado = &v
		}
		if v, ok := a["asignadoA"]; ok {
			c.Asignar = true
			_ = json.Unmarshal(v, &c.AsignadoA)
		}
		responder(w, http.StatusNoContent, nil, s.Actualizar(r.Context(), m, r.PathValue("tipo"), r.PathValue("id"), c, cliente(r)))
	}))
	mux.HandleFunc("POST /v1/equipo/consultas/{tipo}/{id}/responder", con("consultas.responder", func(w http.ResponseWriter, r *http.Request, m Miembro) {
		var a struct{ Asunto, Texto string }
		if leer(w, r, &a) {
			v, err := s.Responder(r.Context(), m, r.PathValue("tipo"), r.PathValue("id"), a.Asunto, a.Texto, cliente(r))
			responder(w, http.StatusCreated, v, err)
		}
	}))
	mux.HandleFunc("GET /v1/equipo/partners", con("partners.ver", func(w http.ResponseWriter, r *http.Request, _ Miembro) {
		v, err := s.Partners(r.Context())
		responder(w, http.StatusOK, v, err)
	}))
	mux.HandleFunc("POST /v1/equipo/partners/{id}", con("partners.decidir", func(w http.ResponseWriter, r *http.Request, m Miembro) {
		var a struct{ Decision, Motivo string }
		if leer(w, r, &a) {
			responder(w, http.StatusNoContent, nil, s.Decidir(r.Context(), m, r.PathValue("id"), a.Decision, a.Motivo, cliente(r)))
		}
	}))
	mux.HandleFunc("GET /v1/equipo/facturacion", con("facturacion.ver", func(w http.ResponseWriter, r *http.Request, _ Miembro) {
		v, err := s.Facturacion(r.Context())
		responder(w, http.StatusOK, v, err)
	}))
	mux.HandleFunc("GET /v1/equipo/auditoria", con("auditoria.ver", func(w http.ResponseWriter, r *http.Request, _ Miembro) {
		q := r.URL.Query()
		v, err := s.Auditoria(r.Context(), q.Get("miembro"), q.Get("accion"), q.Get("tienda"))
		responder(w, http.StatusOK, v, err)
	}))
}

func leer(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(v); err != nil {
		problema.Escribir(w, problema.Invalido("El pedido no es un JSON válido."))
		return false
	}
	return true
}

func escribirError(w http.ResponseWriter, err error) {
	var inv ErrInvalido
	switch {
	case errors.As(err, &inv):
		problema.Escribir(w, problema.Invalido(inv.Mensaje, problema.Campo{Ruta: inv.Campo, Mensaje: inv.Mensaje}))
	case errors.Is(err, ErrCredenciales):
		p := problema.SinSesion()
		p.Codigo, p.Titulo, p.Detalle = "credenciales_invalidas", "No pudimos ingresar", "El email o la contraseña no son correctos."
		problema.Escribir(w, p)
	case errors.Is(err, ErrIntentos):
		w.Header().Set("Retry-After", "900")
		problema.Escribir(w, &problema.Problema{Estado: http.StatusTooManyRequests, Codigo: "demasiados_intentos",
			Titulo: "Esperá un rato", Detalle: "Hubo demasiados intentos. Probá de nuevo en 15 minutos."})
	case errors.Is(err, ErrSinPermiso):
		problema.Escribir(w, problema.SinPermiso("Tu rol en el equipo no puede hacer esto."))
	case errors.Is(err, ErrNoExiste):
		problema.Escribir(w, problema.NoEncontrado("No existe o ya no está."))
	case errors.Is(err, ErrResuelta):
		problema.Escribir(w, problema.Conflicto("ya_resuelta", "Alguien del equipo ya la resolvió."))
	case errors.Is(err, ErrCerrada):
		problema.Escribir(w, problema.Conflicto("tienda_cerrada", "La tienda está cerrada: no se le pueden hacer cambios."))
	default:
		slog.Error("equipo", "error", err)
		problema.Escribir(w, problema.Interno())
	}
}
