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
	// Carrito abandonado: el checkout lo guarda cuando quien compra deja su email (tope por IP).
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/carritos", func(w http.ResponseWriter, r *http.Request) {
		if c, ok := pool.(limite.Consultor); ok && ip != nil {
			if ok, err := limite.Permitido(r.Context(), c, "tienda:carrito:"+ip(r), 60, time.Hour); err == nil && !ok {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
		}
		var b struct {
			Email   string                `json:"email"`
			Nombre  string                `json:"nombre"`
			Carrito []ventas.LineaCarrito `json:"carrito"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&b) != nil {
			error404(w, ErrInvalido)
			return
		}
		id, err := tiendaDe(r.Context(), pool, r.PathValue("slug"))
		if err == nil {
			err = ventas.GuardarCarrito(r.Context(), pool, id, b.Email, b.Nombre, b.Carrito)
		}
		if errors.Is(err, ventas.ErrCarritoInvalido) {
			error404(w, ErrInvalido)
			return
		}
		if err != nil {
			error404(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	// El link del recordatorio: las líneas para volver a armar el carrito.
	mux.HandleFunc("GET /v1/publico/tiendas/{slug}/carritos/{token}", func(w http.ResponseWriter, r *http.Request) {
		id, err := tiendaDe(r.Context(), pool, r.PathValue("slug"))
		var ls []ventas.LineaCarrito
		if err == nil {
			ls, err = ventas.CarritoPorToken(r.Context(), pool, id, r.PathValue("token"))
		}
		if errors.Is(err, ventas.ErrNoExisteCarrito) {
			error404(w, ErrNoExiste)
			return
		}
		if err != nil {
			error404(w, err)
			return
		}
		escribir(w, http.StatusOK, map[string]any{"carrito": ls})
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
