package visitas

import (
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Rutas: POST /v1/interno/visitas, sólo para la web (clave del BFF). Sin clave configurada, no se
// monta nada: nadie de afuera puede sumar visitas.
func Rutas(mux *http.ServeMux, pool tx, claveBFF string, log *slog.Logger) {
	if claveBFF == "" {
		return
	}
	mux.HandleFunc("POST /v1/interno/visitas", func(w http.ResponseWriter, r *http.Request) {
		if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-Mostrador-BFF")), []byte(claveBFF)) != 1 {
			problema.Escribir(w, problema.NoEncontrado("No existe."))
			return
		}
		var b struct {
			Visitas []Vista `json:"visitas"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 512<<10)).Decode(&b) != nil {
			problema.Escribir(w, problema.Invalido("Mandá las visitas."))
			return
		}
		if _, err := Registrar(r.Context(), pool, b.Visitas, time.Now()); err != nil {
			if log != nil {
				log.Warn("visitas: registrar", "error", err)
			}
			problema.Escribir(w, problema.Interno())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
