package formularios

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Rutas del dominio. `ip` resuelve la IP del visitante (la que manda el BFF, si es el BFF).
func (s *Servicio) Rutas(mux *http.ServeMux, ip func(*http.Request) string) {
	mux.HandleFunc("POST /v1/contacto", func(w http.ResponseWriter, r *http.Request) {
		var c Consulta
		if leer(w, r, &c) {
			responder(w, s.Consultar(r.Context(), c, ip(r)), http.StatusAccepted)
		}
	})
	mux.HandleFunc("POST /v1/partners/postulaciones", func(w http.ResponseWriter, r *http.Request) {
		var p Postulacion
		if leer(w, r, &p) {
			responder(w, s.Postular(r.Context(), p, ip(r)), http.StatusAccepted)
		}
	})
	mux.HandleFunc("POST /v1/novedades/suscripciones", func(w http.ResponseWriter, r *http.Request) {
		var a struct{ Email string }
		if leer(w, r, &a) {
			responder(w, s.Suscribir(r.Context(), a.Email, ip(r)), http.StatusAccepted)
		}
	})
	mux.HandleFunc("POST /v1/novedades/baja", func(w http.ResponseWriter, r *http.Request) {
		var a struct{ Token string }
		if leer(w, r, &a) {
			responder(w, s.Baja(r.Context(), a.Token), http.StatusNoContent)
		}
	})
}

func leer(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(v); err != nil {
		problema.Escribir(w, problema.Invalido("El pedido no es un JSON válido."))
		return false
	}
	return true
}

func responder(w http.ResponseWriter, err error, ok int) {
	var inv ErrInvalido
	switch {
	case err == nil:
		w.Header().Set("Cache-Control", "no-store")
		w.WriteHeader(ok)
	case errors.As(err, &inv):
		problema.Escribir(w, problema.Invalido(inv.Mensaje, problema.Campo{Ruta: inv.Campo, Mensaje: inv.Mensaje}))
	case errors.Is(err, ErrDemasiados):
		w.Header().Set("Retry-After", "3600")
		problema.Escribir(w, &problema.Problema{Estado: http.StatusTooManyRequests, Codigo: "demasiados_envios",
			Titulo: "Esperá un rato", Detalle: "Recibimos varios envíos seguidos. Probá de nuevo en una hora o escribinos a ayuda@tulocalweb.com.ar."})
	case errors.Is(err, ErrTokenInvalido):
		problema.Escribir(w, problema.NoProcesable("link_invalido", "Ese link de baja ya se usó o no existe."))
	default:
		slog.Error("formularios", "error", err)
		problema.Escribir(w, problema.Interno())
	}
}
