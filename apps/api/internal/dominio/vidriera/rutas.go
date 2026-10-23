package vidriera

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/ventas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/limite"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

func escribir(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(v)
}

// Rutas públicas de la tienda online (sin sesión: las pide el BFF de /t/{slug}).
func Rutas(mux *http.ServeMux, pool tx, mp Pagador, sitio string, ip func(*http.Request) string) {
	error404 := func(w http.ResponseWriter, err error) {
		switch {
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("Esta tienda no existe o no está publicada."))
		case errors.Is(err, ErrInvalido):
			problema.Escribir(w, problema.Invalido("Revisá los datos del pedido."))
		case errors.Is(err, ErrSinStock):
			problema.Escribir(w, problema.Conflicto("sin_stock", "Algún producto se quedó sin stock."))
		default:
			problema.Escribir(w, problema.Interno())
		}
	}
	mux.HandleFunc("GET /v1/publico/tiendas/{slug}", func(w http.ResponseWriter, r *http.Request) {
		t, err := Publica(r.Context(), pool, r.PathValue("slug"))
		if err != nil {
			error404(w, err)
			return
		}
		escribir(w, http.StatusOK, t)
	})
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/pedidos", func(w http.ResponseWriter, r *http.Request) {
		var c Compra
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10)).Decode(&c) != nil {
			error404(w, ErrInvalido)
			return
		}
		res, err := Comprar(r.Context(), pool, mp, sitio, r.PathValue("slug"), c, time.Now())
		if err != nil && res.Numero == 0 {
			error404(w, err)
			return
		}
		// Si el pedido quedó pero Mercado Pago no respondió, igual se confirma: se puede pagar después.
		escribir(w, http.StatusCreated, res)
	})
	// Reseñas y preguntas de quien compra (con tope por IP: no se puede inundar una tienda).
	tope := func(w http.ResponseWriter, r *http.Request, que string) bool {
		c, ok := pool.(limite.Consultor)
		if !ok {
			return true
		}
		if ok, err := limite.Permitido(r.Context(), c, "tienda:"+que+":"+ip(r), 10, time.Hour); err == nil && !ok {
			problema.Escribir(w, &problema.Problema{Estado: http.StatusTooManyRequests, Codigo: "demasiados",
				Titulo: "Demasiados envíos", Detalle: "Probá de nuevo en un rato."})
			return false
		}
		return true
	}
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/resenas", func(w http.ResponseWriter, r *http.Request) {
		if !tope(w, r, "resena") {
			return
		}
		var n ventas.NuevaResena
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&n) != nil {
			error404(w, ErrInvalido)
			return
		}
		id, err := tiendaDe(r.Context(), pool, r.PathValue("slug"))
		if err == nil {
			err = ventas.DejarResena(r.Context(), pool, id, n)
		}
		switch {
		case errors.Is(err, ventas.ErrYaReseno):
			problema.Escribir(w, problema.Conflicto("ya_reseno", "Ya dejaste tu reseña de esa compra. ¡Gracias!"))
		case errors.Is(err, ventas.ErrResenaInvalida):
			error404(w, ErrInvalido)
		case err != nil:
			error404(w, err)
		default:
			w.WriteHeader(http.StatusCreated)
		}
	})
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/preguntas", func(w http.ResponseWriter, r *http.Request) {
		if !tope(w, r, "pregunta") {
			return
		}
		var q struct{ ProductoID, Nombre, Email, Pregunta string }
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&q) != nil {
			error404(w, ErrInvalido)
			return
		}
		id, err := tiendaDe(r.Context(), pool, r.PathValue("slug"))
		if err == nil {
			err = ventas.Preguntar(r.Context(), pool, id, q.ProductoID, q.Nombre, q.Email, q.Pregunta)
		}
		if errors.Is(err, ventas.ErrResenaInvalida) {
			error404(w, ErrInvalido)
			return
		}
		if err != nil {
			error404(w, err)
			return
		}
		w.WriteHeader(http.StatusCreated)
	})
	mux.HandleFunc("GET /v1/publico/tiendas/{slug}/pedidos/{codigo}", func(w http.ResponseWriter, r *http.Request) {
		c, err := VerPedido(r.Context(), pool, r.PathValue("slug"), r.PathValue("codigo"))
		if err != nil {
			error404(w, err)
			return
		}
		escribir(w, http.StatusOK, c)
	})
}
