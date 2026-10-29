package telegram

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

func escribir(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(v)
}

func leer(w http.ResponseWriter, r *http.Request, v any) bool {
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(v) != nil {
		problema.Escribir(w, problema.Invalido("El pedido no es válido."))
		return false
	}
	return true
}

func escribirError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrCodigo):
		problema.Escribir(w, problema.Conflicto("codigo_invalido", "Ese link venció o ya se usó. Generá uno nuevo desde Integraciones › Telegram."))
	case errors.Is(err, ErrChatOcupado):
		problema.Escribir(w, problema.Conflicto("chat_ocupado", "Este chat ya está vinculado a otra tienda."))
	case errors.Is(err, ErrSinVinculo):
		problema.Escribir(w, problema.NoEncontrado("Este chat no está vinculado a ninguna tienda."))
	case errors.Is(err, ErrSinPermiso):
		problema.Escribir(w, problema.SinPermiso("Sólo la dueña o quien administra puede vincular Telegram."))
	case errors.Is(err, ErrInvalido):
		problema.Escribir(w, problema.Invalido("El pedido no es válido."))
	case errors.Is(err, ErrDemasiados):
		w.Header().Set("Retry-After", "3600")
		problema.Escribir(w, &problema.Problema{Estado: http.StatusTooManyRequests, Codigo: "demasiados",
			Titulo: "Demasiados intentos", Detalle: "Probá de nuevo en una hora."})
	default:
		slog.Error("telegram", "error", err)
		problema.Escribir(w, problema.Interno())
	}
}

// Rutas: las del panel con la sesión de la tienda; las del bot (/v1/telegram/*) sólo con la clave
// del BFF, que es quien recibe el webhook de Telegram ya verificado con su secreto.
func (s *Servicio) Rutas(mux *http.ServeMux, claveBFF string) {
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
	mux.HandleFunc("GET /v1/panel/telegram", panel(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		e, err := s.Estado(r.Context(), d.TiendaID)
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusOK, e)
	}))
	mux.HandleFunc("POST /v1/panel/telegram/codigo", panel(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		c, err := s.Generar(r.Context(), d.TiendaID, d.Rol)
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusCreated, c)
	}))
	// Devuelve el chat que estaba para que el BFF le avise (el navegador no lo ve).
	mux.HandleFunc("DELETE /v1/panel/telegram", panel(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		chat, err := s.Desvincular(r.Context(), d.TiendaID, d.Rol)
		if err != nil {
			escribirError(w, err)
			return
		}
		if chat == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		escribir(w, http.StatusOK, map[string]int64{"chat": chat})
	}))

	bff := func(h func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if !cuentas.EsBFF(r, claveBFF) {
				problema.Escribir(w, problema.SinPermiso("Sólo el BFF de Tu Local Web puede llamar a esta ruta."))
				return
			}
			h(w, r)
		}
	}
	mux.HandleFunc("POST /v1/telegram/vincular", bff(func(w http.ResponseWriter, r *http.Request) {
		var a struct {
			Codigo string `json:"codigo"`
			Chat   int64  `json:"chat"`
			Nombre string `json:"nombre"`
		}
		if !leer(w, r, &a) {
			return
		}
		v, err := s.Vincular(r.Context(), a.Codigo, a.Chat, a.Nombre)
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusOK, v)
	}))
	mux.HandleFunc("POST /v1/telegram/desvincular", bff(func(w http.ResponseWriter, r *http.Request) {
		var a struct {
			Chat int64 `json:"chat"`
		}
		if !leer(w, r, &a) {
			return
		}
		if err := s.DesvincularChat(r.Context(), a.Chat); err != nil {
			escribirError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	// POST: suma el mensaje al tope del día de la tienda.
	mux.HandleFunc("POST /v1/telegram/contexto", bff(func(w http.ResponseWriter, r *http.Request) {
		var a struct {
			Chat int64 `json:"chat"`
		}
		if !leer(w, r, &a) {
			return
		}
		c, err := s.ContextoDe(r.Context(), a.Chat, time.Now())
		if err != nil {
			escribirError(w, err)
			return
		}
		escribir(w, http.StatusOK, c)
	}))
}
