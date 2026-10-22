package listas

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

func escribir(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(v)
}

// sinInternos saca los campos que no van al cliente ("_ficha").
func sinInternos(l Lista) Lista {
	for i, r := range l.Filas {
		c := Registro{}
		for k, v := range r {
			if k != "_ficha" {
				c[k] = v
			}
		}
		l.Filas[i] = c
	}
	return l
}

// Rutas registra el motor de listas y los productos.
func Rutas(mux *http.ServeMux, pool tx, tienda func(*http.Request) string) {
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
	mux.HandleFunc("GET /v1/panel/listas/{lista}", con(func(w http.ResponseWriter, r *http.Request, id string) {
		f := FiltrosDe(r.URL.Query())
		var regs []Registro
		var def Def
		var err error
		switch r.PathValue("lista") {
		case "clientes":
			regs, err = Clientes(r.Context(), pool, id, time.Now())
			def = DefClientes
		case "productos":
			regs, err = Productos(r.Context(), pool, id, time.Now())
			def = DefProductos
		case "inventario", "facturacion", "envios":
			// Llegan con su dominio (inventario por ubicación, ARCA, transportes).
			escribir(w, Lista{Filas: []Registro{}, Segmentos: map[string]int{}, Facetas: map[string]map[string]int{}, Resumen: map[string]float64{}})
			return
		default:
			problema.Escribir(w, problema.NoEncontrado("Esa lista no existe."))
			return
		}
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		escribir(w, sinInternos(Aplicar(regs, def, f)))
	}))
	mux.HandleFunc("POST /v1/panel/listas/{lista}/acciones", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var c struct {
			IDs    []string        `json:"ids"`
			Accion string          `json:"accion"`
			Valor  json.RawMessage `json:"valor"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&c); err != nil || len(c.IDs) > 500 {
			problema.Escribir(w, problema.Invalido("Elegí filas y una acción."))
			return
		}
		n, err := 0, error(nil)
		if r.PathValue("lista") == "productos" {
			switch c.Accion {
			case "publicar":
				var on bool
				_ = json.Unmarshal(c.Valor, &on)
				n, err = Publicar(r.Context(), pool, id, c.IDs, on)
			case "pausar":
				n, err = Publicar(r.Context(), pool, id, c.IDs, false)
			case "precio":
				var pct float64
				_ = json.Unmarshal(c.Valor, &pct)
				n, err = Precio(r.Context(), pool, id, c.IDs, pct)
			}
		}
		// Cupones, etiquetas y mensajes llegan con Cupones, Clientes y WhatsApp: hasta entonces, 0.
		if err != nil {
			problema.Escribir(w, problema.Invalido(err.Error()))
			return
		}
		escribir(w, map[string]int{"hechos": n})
	}))
	guardar := func(w http.ResponseWriter, r *http.Request, tiendaID, id string, estado int) {
		var e Editable
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(&e); err != nil {
			problema.Escribir(w, problema.Invalido("Mandá el producto."))
			return
		}
		p, err := Guardar(r.Context(), pool, tiendaID, id, e)
		switch {
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("Ese producto no existe."))
		case err != nil:
			problema.Escribir(w, problema.Invalido(err.Error()))
		default:
			w.WriteHeader(estado)
			escribir(w, p)
		}
	}
	mux.HandleFunc("POST /v1/panel/productos", con(func(w http.ResponseWriter, r *http.Request, id string) {
		guardar(w, r, id, "", http.StatusCreated)
	}))
	mux.HandleFunc("PUT /v1/panel/productos/{id}", con(func(w http.ResponseWriter, r *http.Request, id string) {
		guardar(w, r, id, r.PathValue("id"), http.StatusOK)
	}))
	mux.HandleFunc("GET /v1/panel/productos/{id}", con(func(w http.ResponseWriter, r *http.Request, id string) {
		p, err := Obtener(r.Context(), pool, id, r.PathValue("id"), time.Now())
		switch {
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("Ese producto no existe."))
		case err != nil:
			problema.Escribir(w, problema.Interno())
		default:
			escribir(w, p)
		}
	}))
}
