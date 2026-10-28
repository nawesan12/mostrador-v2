// Package ventas: bandeja de mensajes y carritos abandonados (design/PanelVentas.dc.html). Las
// conversaciones llegan con WhatsApp Cloud, Instagram, Telegram y Mercado Libre, y los carritos con
// el checkout de la tienda: hasta entonces las listas van vacías, nunca inventadas.
package ventas

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}

// Automaticos son los recordatorios automáticos de carritos.
type Automaticos struct {
	Hora  bool `json:"hora"`
	Dia   bool `json:"dia"`
	Email bool `json:"email"`
}

// LeerAutomaticos y GuardarAutomaticos: la configuración de la tienda.
func LeerAutomaticos(ctx context.Context, pool tx, tiendaID string) (Automaticos, error) {
	var a Automaticos
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT carritos_automaticos FROM tiendas WHERE id = $1`, tiendaID).Scan(&a)
	})
	return a, err
}

func GuardarAutomaticos(ctx context.Context, pool tx, tiendaID string, a Automaticos) error {
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE tiendas SET carritos_automaticos = $2 WHERE id = $1`, tiendaID, a)
		return err
	})
}

func escribir(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(v)
}

// Rutas registra la bandeja y los carritos.
// PreguntaPanel es una pregunta de la pantalla Reseñas y preguntas.
type PreguntaPanel struct {
	ID        string    `json:"id"`
	Producto  string    `json:"producto"`
	Foto      *string   `json:"foto"`
	Pregunta  string    `json:"pregunta"`
	Creada    time.Time `json:"creada"`
	Sugerida  string    `json:"sugerida"`
	Respuesta *string   `json:"respuesta"`
}

// Preguntas de los canales (por ahora, Mercado Libre). nil: no hay canal conectado.
type Preguntas interface {
	PreguntasPanel(ctx context.Context, tiendaID string) ([]PreguntaPanel, error)
	ResponderPanel(ctx context.Context, tiendaID, id, texto string) error
}

func Rutas(mux *http.ServeMux, pool tx, tienda func(*http.Request) string, preguntas Preguntas) {
	con := func(h func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id := tienda(r)
			if id == "" {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			h(w, r, id)
		}
	}
	rutasTarjeta(mux, pool, con)
	mux.HandleFunc("GET /v1/panel/bandeja", con(func(w http.ResponseWriter, _ *http.Request, _ string) {
		escribir(w, map[string]any{"conversaciones": []any{}, "atendidasHoy": 0})
	}))
	sinConversacion := con(func(w http.ResponseWriter, _ *http.Request, _ string) {
		problema.Escribir(w, problema.NoEncontrado("Esa conversación no existe."))
	})
	mux.HandleFunc("GET /v1/panel/bandeja/{id}", sinConversacion)
	mux.HandleFunc("POST /v1/panel/bandeja/{id}", sinConversacion)
	mux.HandleFunc("GET /v1/panel/carritos", con(func(w http.ResponseWriter, r *http.Request, id string) {
		a, err := LeerAutomaticos(r.Context(), pool, id)
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		escribir(w, map[string]any{"carritos": []any{}, "abiertos": 0, "monto": 0, "recuperado": 0, "tasa": 0,
			"costoMensaje": 120, "automaticos": a})
	}))
	mux.HandleFunc("PUT /v1/panel/carritos", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var a Automaticos
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&a); err != nil {
			problema.Escribir(w, problema.Invalido("Mandá los recordatorios."))
			return
		}
		if err := GuardarAutomaticos(r.Context(), pool, id, a); err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	// Las devoluciones están en el paquete devoluciones (botón de arrepentimiento de la tienda).
	mux.HandleFunc("GET /v1/panel/resenas", con(func(w http.ResponseWriter, r *http.Request, id string) {
		rs, ps, err := ResenasPanel(r.Context(), pool, id)
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		if preguntas != nil {
			ml, err := preguntas.PreguntasPanel(r.Context(), id)
			if err != nil {
				problema.Escribir(w, problema.Interno())
				return
			}
			ps = append(ps, ml...)
		}
		prom, total, por, rec := Resumen(rs)
		escribir(w, map[string]any{"preguntas": ps, "resenas": rs, "promedio": prom, "total": total,
			"porEstrellas": por, "recomendaria": rec})
	}))
	mux.HandleFunc("POST /v1/panel/resenas/{id}", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var b struct{ Accion, Respuesta string }
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&b) != nil {
			problema.Escribir(w, problema.Invalido("Pedido inválido."))
			return
		}
		if err := AccionResena(r.Context(), pool, id, r.PathValue("id"), b.Accion, b.Respuesta); err != nil {
			problema.Escribir(w, problema.NoEncontrado("Esa reseña no existe."))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	mux.HandleFunc("POST /v1/panel/preguntas/{id}", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var b struct{ Respuesta string }
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&b) != nil {
			problema.Escribir(w, problema.NoEncontrado("Esa pregunta no existe."))
			return
		}
		// Las de la tienda online empiezan con "t-"; las demás son de Mercado Libre.
		if strings.HasPrefix(r.PathValue("id"), "t-") {
			if err := ResponderPregunta(r.Context(), pool, id, r.PathValue("id"), b.Respuesta); err != nil {
				problema.Escribir(w, problema.NoEncontrado("No se pudo responder esa pregunta."))
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if preguntas == nil {
			problema.Escribir(w, problema.NoEncontrado("Esa pregunta no existe."))
			return
		}
		if err := preguntas.ResponderPanel(r.Context(), id, r.PathValue("id"), b.Respuesta); err != nil {
			problema.Escribir(w, problema.NoEncontrado("No se pudo responder esa pregunta."))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	mux.HandleFunc("POST /v1/panel/carritos/{id}/recordar", con(func(w http.ResponseWriter, _ *http.Request, _ string) {
		problema.Escribir(w, problema.NoEncontrado("Ese carrito no existe."))
	}))
}
