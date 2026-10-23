package mercadolibre

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Problema traduce los errores. De un rechazo de Mercado Libre se muestra su mensaje (dice qué le
// falta a la publicación), nunca el cuerpo entero.
func (s *Servicio) Problema(w http.ResponseWriter, err error) {
	var em *ErrorML
	var ev *ErrVariaciones
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, ErrApagado), errors.Is(err, ErrSinConexion):
		problema.Escribir(w, problema.Conflicto("sin_ml", "La tienda todavía no conectó Mercado Libre."))
	case errors.Is(err, ErrReautorizar):
		problema.Escribir(w, problema.Conflicto("reconectar", "Hay que volver a conectar Mercado Libre."))
	case errors.Is(err, ErrYaPublicado):
		problema.Escribir(w, problema.Conflicto("publicado", "Ese producto ya está publicado."))
	case errors.Is(err, ErrNoExiste):
		problema.Escribir(w, problema.NoEncontrado("No lo encontramos."))
	case errors.Is(err, ErrEstado):
		problema.Escribir(w, problema.Invalido("La conexión venció. Probá de nuevo."))
	case errors.Is(err, ErrInvalido):
		problema.Escribir(w, problema.Invalido("Revisá los datos (para publicar hace falta precio y stock)."))
	case errors.As(err, &ev):
		problema.Escribir(w, problema.NoProcesable("variantes_ml", ev.Mensaje))
	case errors.As(err, &em) && em.Estado < 500:
		msg := em.Mensaje
		if msg == "" {
			msg = "Mercado Libre rechazó el pedido."
		}
		problema.Escribir(w, problema.NoProcesable("mercadolibre", msg))
	case errors.As(err, &em):
		s.log().Warn("mercadolibre: la API respondió con error", "estado", em.Estado)
		problema.Escribir(w, &problema.Problema{Estado: http.StatusBadGateway, Codigo: "mercadolibre",
			Titulo: "Mercado Libre no respondió", Detalle: "Probá de nuevo en un rato."})
	default:
		s.log().Error("mercadolibre", "error", err)
		problema.Escribir(w, problema.Interno())
	}
}

func escribir(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(v)
}

func leer(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(v); err != nil {
		problema.Escribir(w, problema.Invalido("No entendimos el pedido."))
		return false
	}
	return true
}

// Rutas del panel y el webhook.
func (s *Servicio) Rutas(mux *http.ServeMux) {
	sesion := func(soloAdmin bool, h func(http.ResponseWriter, *http.Request, cuentas.Datos)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			d, ok := cuentas.SesionDe(r)
			if !ok || d.TiendaID == "" {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			if soloAdmin && d.Rol != "duenio" && d.Rol != "admin" {
				problema.Escribir(w, problema.SinPermiso("Sólo la dueña o quien administra puede hacerlo."))
				return
			}
			h(w, r, d)
		}
	}
	mux.HandleFunc("GET /v1/panel/integraciones/mercadolibre", sesion(false, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		e, err := s.EstadoDe(r.Context(), d.TiendaID)
		if err != nil {
			s.Problema(w, err)
			return
		}
		escribir(w, http.StatusOK, e)
	}))
	mux.HandleFunc("PUT /v1/panel/integraciones/mercadolibre", sesion(true, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		var b struct {
			Recargo        int  `json:"recargo"`
			PausarSinStock bool `json:"pausarSinStock"`
		}
		if leer(w, r, &b) {
			s.Problema(w, s.Ajustar(r.Context(), d.TiendaID, b.Recargo, b.PausarSinStock))
		}
	}))
	mux.HandleFunc("DELETE /v1/panel/integraciones/mercadolibre", sesion(true, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		s.Problema(w, s.Desconectar(r.Context(), d.TiendaID))
	}))
	mux.HandleFunc("POST /v1/panel/integraciones/mercadolibre/conectar", sesion(true, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		u, err := s.Conectar(r.Context(), d.TiendaID, d.Usuario.ID)
		if err != nil {
			s.Problema(w, err)
			return
		}
		escribir(w, http.StatusOK, map[string]string{"url": u})
	}))
	mux.HandleFunc("POST /v1/panel/integraciones/mercadolibre/vuelta", sesion(true, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		var b struct{ Codigo, Estado string }
		if leer(w, r, &b) {
			s.Problema(w, s.Volver(r.Context(), d.TiendaID, d.Usuario.ID, b.Codigo, b.Estado))
		}
	}))
	mux.HandleFunc("POST /v1/panel/productos/{id}/mercadolibre", sesion(false, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		p, err := s.Publicar(r.Context(), d.TiendaID, r.PathValue("id"))
		if err != nil {
			s.Problema(w, err)
			return
		}
		escribir(w, http.StatusCreated, p)
	}))
	mux.HandleFunc("DELETE /v1/panel/productos/{id}/mercadolibre", sesion(false, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		s.Problema(w, s.Despublicar(r.Context(), d.TiendaID, r.PathValue("id")))
	}))
	mux.HandleFunc("GET /v1/panel/mercadolibre/preguntas", sesion(false, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		q, err := s.Preguntas(r.Context(), d.TiendaID)
		if err != nil {
			s.Problema(w, err)
			return
		}
		escribir(w, http.StatusOK, q)
	}))
	mux.HandleFunc("POST /v1/panel/mercadolibre/preguntas/{id}/respuesta", sesion(false, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			s.Problema(w, ErrNoExiste)
			return
		}
		var b struct{ Texto string }
		if leer(w, r, &b) {
			s.Problema(w, s.Responder(r.Context(), d.TiendaID, id, b.Texto))
		}
	}))

	// Avisos: Mercado Libre espera un 200 rápido. Se encolan y el worker los procesa.
	mux.HandleFunc("POST /v1/webhooks/mercadolibre", func(w http.ResponseWriter, r *http.Request) {
		var a Aviso
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
		dec.UseNumber()
		if dec.Decode(&a) != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := s.Recibir(r.Context(), a); err != nil {
			s.log().Error("mercadolibre: aviso", "topic", a.Topic, "error", err)
			w.WriteHeader(http.StatusInternalServerError) // Mercado Libre reintenta
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}
