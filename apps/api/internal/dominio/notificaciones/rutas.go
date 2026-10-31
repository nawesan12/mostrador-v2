package notificaciones

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Rutas de la campanita. Son roles.Panel (cualquiera del equipo) y cada persona ve sólo las suyas en
// la tienda de su sesión. Una IA conectada o alguien del soporte mirando no tienen notificaciones
// propias: la lista les llega vacía y no pueden marcar.
func Rutas(mux *http.ServeMux, p Pool) {
	type quien struct{ tienda, usuario string }
	// persona: la sesión de una persona (no una IA ni una mirada). ok=false si no hay sesión.
	persona := func(w http.ResponseWriter, r *http.Request) (q quien, propia, ok bool) {
		d, ok := cuentas.SesionDe(r)
		if !ok || d.TiendaID == "" {
			problema.Escribir(w, problema.SinSesion())
			return q, false, false
		}
		return quien{d.TiendaID, d.Usuario.ID}, d.IA == "" && !d.Mirada && d.Usuario.ID != "", true
	}
	escribir := func(w http.ResponseWriter, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(v)
	}
	mux.HandleFunc("GET /v1/panel/notificaciones", func(w http.ResponseWriter, r *http.Request) {
		q, propia, ok := persona(w, r)
		if !ok {
			return
		}
		if !propia {
			escribir(w, Pagina{Notificaciones: []Notificacion{}})
			return
		}
		limite, _ := strconv.Atoi(r.URL.Query().Get("limite"))
		pg, err := Listar(r.Context(), p, q.tienda, q.usuario, r.URL.Query().Get("cursor"), limite)
		if errors.Is(err, ErrCursor) {
			problema.Escribir(w, problema.Invalido("El cursor no es válido.", problema.Campo{Ruta: "cursor", Mensaje: "Usá el que vino en «siguiente»."}))
			return
		}
		if err != nil {
			slog.Error("notificaciones: listar", "error", err)
			problema.Escribir(w, problema.Interno())
			return
		}
		escribir(w, pg)
	})
	mux.HandleFunc("GET /v1/panel/notificaciones/sin-leer", func(w http.ResponseWriter, r *http.Request) {
		q, propia, ok := persona(w, r)
		if !ok {
			return
		}
		n := 0
		if propia {
			var err error
			if n, err = SinLeer(r.Context(), p, q.tienda, q.usuario); err != nil {
				slog.Error("notificaciones: contar", "error", err)
				problema.Escribir(w, problema.Interno())
				return
			}
		}
		escribir(w, map[string]int{"sinLeer": n})
	})
	mux.HandleFunc("POST /v1/panel/notificaciones/{id}/leida", func(w http.ResponseWriter, r *http.Request) {
		q, propia, ok := persona(w, r)
		if !ok {
			return
		}
		if !propia {
			problema.Escribir(w, problema.SinPermiso("Las notificaciones son de cada persona."))
			return
		}
		err := MarcarLeida(r.Context(), p, q.tienda, q.usuario, r.PathValue("id"))
		switch {
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("Esa notificación ya no está."))
		case err != nil:
			slog.Error("notificaciones: marcar", "error", err)
			problema.Escribir(w, problema.Interno())
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	})
	mux.HandleFunc("POST /v1/panel/notificaciones/leidas", func(w http.ResponseWriter, r *http.Request) {
		q, propia, ok := persona(w, r)
		if !ok {
			return
		}
		if propia {
			if _, err := MarcarTodas(r.Context(), p, q.tienda, q.usuario); err != nil {
				slog.Error("notificaciones: marcar todas", "error", err)
				problema.Escribir(w, problema.Interno())
				return
			}
		}
		w.WriteHeader(http.StatusNoContent)
	})
}
