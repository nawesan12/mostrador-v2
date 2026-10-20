// Package servidor arma el http.Handler de la API: rutas de salud y, más adelante, los dominios.
package servidor

import (
	"encoding/json"
	"net/http"
)

// Nuevo devuelve el handler raíz de la API.
func Nuevo() http.Handler {
	mux := http.NewServeMux()
	// Vivo: el proceso responde. No toca la base, para que un corte de Postgres no reinicie la API.
	mux.HandleFunc("GET /salud/vivo", func(w http.ResponseWriter, _ *http.Request) {
		escribirJSON(w, http.StatusOK, map[string]string{"estado": "vivo"})
	})
	return mux
}

func escribirJSON(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(v)
}
