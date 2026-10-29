package importacion

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

func escribir(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(v)
}

func escribirError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrSinPermiso):
		problema.Escribir(w, problema.SinPermiso("Sólo la dueña o quien administra puede importar una tienda."))
	case errors.Is(err, ErrInvalido):
		problema.Escribir(w, problema.Invalido("Revisá la plataforma y la dirección de tu tienda."))
	case errors.Is(err, ErrSinFuente):
		problema.Escribir(w, problema.Conflicto("sin_fuente", "Todavía no podemos importar desde esa plataforma."))
	case errors.Is(err, ErrActiva):
		problema.Escribir(w, problema.Conflicto("importacion_activa", "Ya hay una importación en curso: esperá que termine o cancelala."))
	case errors.Is(err, ErrTerminada):
		problema.Escribir(w, problema.Conflicto("terminada", "Esa importación ya terminó."))
	case errors.Is(err, ErrSinCifrado):
		problema.Escribir(w, problema.Conflicto("sin_cifrado", "No podemos guardar la clave de tu tienda en este momento. Probá más tarde."))
	case errors.Is(err, ErrTiendanubeApagado):
		problema.Escribir(w, problema.Conflicto("sin_tiendanube", "La conexión con Tiendanube todavía no está disponible."))
	case errors.Is(err, ErrEstado):
		problema.Escribir(w, problema.Invalido("La conexión con Tiendanube venció. Probá de nuevo."))
	case errors.Is(err, ErrCanje):
		slog.Warn("importacion: canje de Tiendanube", "error", err)
		problema.Escribir(w, &problema.Problema{Estado: http.StatusBadGateway, Codigo: "tiendanube",
			Titulo: "Tiendanube no respondió", Detalle: "No pudimos conectar tu tienda. Probá de nuevo."})
	case errors.Is(err, ErrNoExiste):
		problema.Escribir(w, problema.NoEncontrado("Esa importación no existe."))
	default:
		slog.Error("importacion", "error", err)
		problema.Escribir(w, problema.Interno())
	}
}

// Rutas del panel (con la sesión de la tienda). Crear y cancelar, sólo dueña o admin.
func (s *Servicio) Rutas(mux *http.ServeMux) {
	panel := func(h func(http.ResponseWriter, *http.Request, cuentas.Datos)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			d, ok := cuentas.SesionDe(r)
			if !ok || d.TiendaID == "" || d.IA != "" {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			h(w, r, d)
		}
	}
	mux.HandleFunc("POST /v1/panel/importaciones", panel(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		var n Nueva
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&n) != nil {
			escribirError(w, ErrInvalido)
			return
		}
		imp, err := s.Crear(r.Context(), d.TiendaID, d.Rol, n)
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusCreated, imp)
	}))
	mux.HandleFunc("GET /v1/panel/importaciones/actual", panel(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		imp, err := s.Actual(r.Context(), d.TiendaID)
		if err != nil {
			escribirError(w, err)
			return
		}
		if imp == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		escribir(w, http.StatusOK, imp)
	}))
	// Tiendanube: si se puede conectar, adónde mandar a la dueña y la vuelta del OAuth (la web pasa el
	// código y el state; el token se canjea y se guarda acá, cifrado, y nunca vuelve).
	mux.HandleFunc("GET /v1/panel/importaciones/tiendanube", panel(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		escribir(w, http.StatusOK, map[string]bool{"disponible": s.TiendanubeDisponible()})
	}))
	mux.HandleFunc("POST /v1/panel/importaciones/tiendanube/conectar", panel(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		u, err := s.ConectarTiendanube(r.Context(), d.TiendaID, d.Rol, d.Usuario.ID)
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusOK, map[string]string{"url": u})
	}))
	mux.HandleFunc("POST /v1/panel/importaciones/tiendanube/vuelta", panel(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		var b struct{ Codigo, Estado string }
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&b) != nil {
			escribirError(w, ErrInvalido)
			return
		}
		imp, err := s.VolverTiendanube(r.Context(), d.TiendaID, d.Rol, d.Usuario.ID, b.Codigo, b.Estado)
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusCreated, imp)
	}))
	mux.HandleFunc("POST /v1/panel/importaciones/{id}/cancelar", panel(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		imp, err := s.Cancelar(r.Context(), d.TiendaID, d.Rol, r.PathValue("id"))
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusOK, imp)
	}))
}
