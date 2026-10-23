package vidriera

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/devoluciones"
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
func Rutas(mux *http.ServeMux, pool tx, mp Pagador, sitio string, dev *devoluciones.Servicio, ip func(*http.Request) string) {
	error404 := func(w http.ResponseWriter, err error) {
		var ec ErrCupon
		switch {
		case errors.As(err, &ec):
			problema.Escribir(w, problema.Conflicto("cupon", ec.Motivo))
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
	// Vista previa del cupón en el carrito y el checkout (al comprar se vuelve a evaluar).
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/cupon", func(w http.ResponseWriter, r *http.Request) {
		var c Compra
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10)).Decode(&c) != nil || len(c.Carrito) > 50 {
			error404(w, ErrInvalido)
			return
		}
		t, err := Publica(r.Context(), pool, r.PathValue("slug"))
		if err != nil {
			error404(w, err)
			return
		}
		id, _ := tiendaDe(r.Context(), pool, r.PathValue("slug"))
		porID := map[string]ProductoPublico{}
		for _, p := range t.Productos {
			porID[p.ID] = p
		}
		res, err := EvaluarCupon(r.Context(), pool, id, strings.ToUpper(strings.TrimSpace(c.Cupon)), c.Email, lineasDe(c, porID), time.Now())
		if err != nil {
			error404(w, err)
			return
		}
		escribir(w, http.StatusOK, res)
	})
	// Botón de arrepentimiento de la tienda (Ley 24.240): devuelve el código de trámite.
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/arrepentimiento", func(w http.ResponseWriter, r *http.Request) {
		if c, ok := pool.(limite.Consultor); ok {
			if ok, err := limite.Permitido(r.Context(), c, "tienda:arrepentimiento:"+ip(r), 10, time.Hour); err == nil && !ok {
				problema.Escribir(w, &problema.Problema{Estado: http.StatusTooManyRequests, Codigo: "demasiados",
					Titulo: "Demasiados pedidos", Detalle: "Probá de nuevo en un rato."})
				return
			}
		}
		var p devoluciones.Pedido
		if dev == nil || json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&p) != nil {
			error404(w, ErrInvalido)
			return
		}
		t, err := Publica(r.Context(), pool, r.PathValue("slug"))
		if err != nil {
			error404(w, err)
			return
		}
		id, _ := tiendaDe(r.Context(), pool, r.PathValue("slug"))
		codigo, err := dev.Arrepentirse(r.Context(), id, t.Nombre, p, time.Now())
		switch {
		case errors.Is(err, devoluciones.ErrInvalido):
			problema.Escribir(w, problema.Invalido("Revisá tu nombre y tu email."))
		case errors.Is(err, devoluciones.ErrNoEncontrado):
			problema.Escribir(w, problema.NoEncontrado("No encontramos ese pedido. Revisá el número y que el email sea el de la compra."))
		case errors.Is(err, devoluciones.ErrFueraDePlazo):
			problema.Escribir(w, problema.Conflicto("fuera_de_plazo", "Pasaron más de 10 días desde que recibiste la compra. Escribile a la tienda para ver qué se puede hacer."))
		case err != nil:
			error404(w, err)
		default:
			escribir(w, http.StatusCreated, map[string]string{"codigo": codigo})
		}
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
