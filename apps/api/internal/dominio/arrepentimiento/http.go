package arrepentimiento

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Rutas del dominio. `ip` resuelve la IP del visitante (la que manda el BFF, si es el BFF).
func (s *Servicio) Rutas(mux *http.ServeMux, ip func(*http.Request) string) {
	mux.HandleFunc("POST /v1/arrepentimientos", func(w http.ResponseWriter, r *http.Request) {
		var p Pedido
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&p); err != nil {
			problema.Escribir(w, problema.Invalido("El pedido no es un JSON válido."))
			return
		}
		codigo, err := s.Pedir(r.Context(), p, ip(r))
		var inv ErrInvalido
		switch {
		case err == nil, errors.Is(err, ErrMailPendiente):
			if err != nil {
				slog.Warn("arrepentimiento: no salió el mail", "codigo", codigo, "error", err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Cache-Control", "no-store")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]string{"codigo": codigo})
		case errors.As(err, &inv):
			problema.Escribir(w, problema.Invalido(inv.Mensaje, problema.Campo{Ruta: inv.Campo, Mensaje: inv.Mensaje}))
		case errors.Is(err, ErrDemasiadosPedidos):
			w.Header().Set("Retry-After", "3600")
			problema.Escribir(w, &problema.Problema{Estado: http.StatusTooManyRequests, Codigo: "demasiados_pedidos",
				Titulo: "Esperá un rato", Detalle: "Recibimos varios pedidos seguidos. Probá de nuevo en una hora o escribinos a legales@mostrador.com.ar."})
		default:
			slog.Error("arrepentimiento", "error", err)
			problema.Escribir(w, problema.Interno())
		}
	})
}
