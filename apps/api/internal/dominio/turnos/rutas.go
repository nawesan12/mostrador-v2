package turnos

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Config de las rutas.
type Config struct {
	// Tienda de la sesión ("" sin sesión).
	Tienda func(*http.Request) string
	// Admin: puede devolver la seña (dueña o quien administra). nil = cualquiera con sesión.
	Admin func(*http.Request) bool
	// MP cobra y devuelve la seña. nil = sin Mercado Pago (los turnos se confirman sin seña).
	MP Cobrador
	// Sitio público, para la vuelta del checkout.
	Sitio string
	Log   *slog.Logger
	Ahora func() time.Time
}

func escribir(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(v)
}

func leer(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10)).Decode(v); err != nil {
		problema.Escribir(w, problema.Invalido("No entendimos el pedido."))
		return false
	}
	return true
}

// Rutas del panel (/v1/panel/turnos, con sesión) y de la tienda pública (/v1/publico/tiendas/{slug}/turnos).
func Rutas(mux *http.ServeMux, pool tx, cfg Config) {
	ahora := cfg.Ahora
	if ahora == nil {
		ahora = time.Now
	}
	log := cfg.Log
	if log == nil {
		log = slog.Default()
	}
	errorDe := func(w http.ResponseWriter, err error) {
		switch {
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("No lo encontramos."))
		case errors.Is(err, ErrInvalido):
			problema.Escribir(w, problema.Invalido("Revisá los datos."))
		case errors.Is(err, ErrApagado):
			problema.Escribir(w, problema.NoEncontrado("Esta tienda no toma turnos online."))
		case errors.Is(err, ErrOcupado):
			problema.Escribir(w, problema.Conflicto("ocupado", "Ese horario ya no está libre. Elegí otro."))
		case errors.Is(err, ErrTarde):
			problema.Escribir(w, problema.Conflicto("tarde", "Ya no se puede cancelar online: escribile a la tienda."))
		case errors.Is(err, ErrSinSena):
			problema.Escribir(w, problema.Conflicto("sin_sena", "Este turno no tiene una seña pagada para devolver."))
		case errors.Is(err, ErrSinCobros):
			problema.Escribir(w, problema.Conflicto("sin_cobros", "La tienda todavía no conectó Mercado Pago."))
		case errors.Is(err, ErrMercadoPago):
			log.Warn("turnos: Mercado Pago no respondió", "error", err)
			problema.Escribir(w, &problema.Problema{Estado: http.StatusBadGateway, Codigo: "mercadopago",
				Titulo: "Mercado Pago no respondió", Detalle: "No pudimos armar el pago de la seña. Probá de nuevo en un rato."})
		default:
			log.Error("turnos", "error", err)
			problema.Escribir(w, problema.Interno())
		}
	}
	con := func(h func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id := ""
			if cfg.Tienda != nil {
				id = cfg.Tienda(r)
			}
			if id == "" {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			h(w, r, id)
		}
	}
	responder := func(w http.ResponseWriter, estado int, v any, err error) {
		if err != nil {
			errorDe(w, err)
			return
		}
		if v == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		escribir(w, estado, v)
	}

	// Panel: ajustes.
	mux.HandleFunc("GET /v1/panel/turnos/ajustes", con(func(w http.ResponseWriter, r *http.Request, id string) {
		a, err := LeerAjustes(r.Context(), pool, id)
		responder(w, http.StatusOK, a, err)
	}))
	mux.HandleFunc("PUT /v1/panel/turnos/ajustes", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var a Ajustes
		if leer(w, r, &a) {
			a, err := GuardarAjustes(r.Context(), pool, id, a)
			responder(w, http.StatusOK, a, err)
		}
	}))

	// Panel: servicios.
	mux.HandleFunc("GET /v1/panel/turnos/servicios", con(func(w http.ResponseWriter, r *http.Request, id string) {
		s, err := Servicios(r.Context(), pool, id)
		responder(w, http.StatusOK, s, err)
	}))
	guardarServicio := func(nuevo bool) http.HandlerFunc {
		return con(func(w http.ResponseWriter, r *http.Request, id string) {
			var b struct {
				Servicio
				Orden *int `json:"orden"`
			}
			b.Activo = true
			if !leer(w, r, &b) {
				return
			}
			estado, sid := http.StatusOK, r.PathValue("id")
			if nuevo {
				estado, sid = http.StatusCreated, ""
			}
			s, err := GuardarServicio(r.Context(), pool, id, sid, b.Servicio, b.Orden)
			responder(w, estado, s, err)
		})
	}
	mux.HandleFunc("POST /v1/panel/turnos/servicios", guardarServicio(true))
	mux.HandleFunc("PUT /v1/panel/turnos/servicios/{id}", guardarServicio(false))
	mux.HandleFunc("DELETE /v1/panel/turnos/servicios/{id}", con(func(w http.ResponseWriter, r *http.Request, id string) {
		responder(w, 0, nil, BorrarServicio(r.Context(), pool, id, r.PathValue("id")))
	}))

	// Panel: profesionales.
	mux.HandleFunc("GET /v1/panel/turnos/profesionales", con(func(w http.ResponseWriter, r *http.Request, id string) {
		p, err := Profesionales(r.Context(), pool, id)
		responder(w, http.StatusOK, p, err)
	}))
	guardarProfesional := func(nuevo bool) http.HandlerFunc {
		return con(func(w http.ResponseWriter, r *http.Request, id string) {
			var b struct {
				Profesional
				Orden *int `json:"orden"`
			}
			b.Activo = true
			if !leer(w, r, &b) {
				return
			}
			estado, pid := http.StatusOK, r.PathValue("id")
			if nuevo {
				estado, pid = http.StatusCreated, ""
			}
			p, err := GuardarProfesional(r.Context(), pool, id, pid, b.Profesional, b.Orden)
			responder(w, estado, p, err)
		})
	}
	mux.HandleFunc("POST /v1/panel/turnos/profesionales", guardarProfesional(true))
	mux.HandleFunc("PUT /v1/panel/turnos/profesionales/{id}", guardarProfesional(false))
	mux.HandleFunc("DELETE /v1/panel/turnos/profesionales/{id}", con(func(w http.ResponseWriter, r *http.Request, id string) {
		responder(w, 0, nil, BorrarProfesional(r.Context(), pool, id, r.PathValue("id")))
	}))

	// Panel: agenda y turnos.
	mux.HandleFunc("GET /v1/panel/turnos", con(func(w http.ResponseWriter, r *http.Request, id string) {
		q := r.URL.Query()
		desde, hasta := q.Get("desde"), q.Get("hasta")
		hoy := ahora().In(Zona)
		if desde == "" {
			desde = hoy.Format("2006-01-02")
		}
		if hasta == "" {
			if d, err := time.ParseInLocation("2006-01-02", desde, Zona); err == nil {
				hasta = d.AddDate(0, 0, 6).Format("2006-01-02")
			}
		}
		t, err := Agenda(r.Context(), pool, id, desde, hasta, q.Get("profesional"), ahora())
		responder(w, http.StatusOK, t, err)
	}))
	mux.HandleFunc("POST /v1/panel/turnos", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var m Manual
		if leer(w, r, &m) {
			t, err := CrearManual(r.Context(), pool, id, m, ahora())
			responder(w, http.StatusCreated, t, err)
		}
	}))
	mux.HandleFunc("PATCH /v1/panel/turnos/{id}", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var b struct {
			Estado string `json:"estado"`
		}
		if leer(w, r, &b) {
			t, err := CambiarEstado(r.Context(), pool, id, r.PathValue("id"), b.Estado, ahora())
			responder(w, http.StatusOK, t, err)
		}
	}))
	mux.HandleFunc("POST /v1/panel/turnos/{id}/devolver-sena", con(func(w http.ResponseWriter, r *http.Request, id string) {
		if cfg.Admin != nil && !cfg.Admin(r) {
			problema.Escribir(w, problema.SinPermiso("Sólo la dueña o quien administra puede devolver una seña."))
			return
		}
		t, err := DevolverSena(r.Context(), pool, cfg.MP, id, r.PathValue("id"), ahora())
		if err != nil && !errors.Is(err, ErrNoExiste) && !errors.Is(err, ErrSinSena) && !errors.Is(err, ErrSinCobros) {
			err = errors.Join(ErrMercadoPago, err)
		}
		responder(w, http.StatusOK, t, err)
	}))

	// Público (sin sesión: lo pide el BFF de /t/{slug}).
	mux.HandleFunc("GET /v1/publico/tiendas/{slug}/turnos", func(w http.ResponseWriter, r *http.Request) {
		res, err := Publico(r.Context(), pool, cfg.MP, r.PathValue("slug"))
		responder(w, http.StatusOK, res, err)
	})
	mux.HandleFunc("GET /v1/publico/tiendas/{slug}/turnos/libres", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		l, err := LibresDe(r.Context(), pool, r.PathValue("slug"), q.Get("servicio"), q.Get("profesional"), q.Get("dia"), ahora())
		responder(w, http.StatusOK, l, err)
	})
	mux.HandleFunc("POST /v1/publico/tiendas/{slug}/turnos", func(w http.ResponseWriter, r *http.Request) {
		var b Reserva
		if leer(w, r, &b) {
			res, err := Reservar(r.Context(), pool, cfg.MP, cfg.Sitio, r.PathValue("slug"), b, ahora())
			responder(w, http.StatusCreated, res, err)
		}
	})
	mux.HandleFunc("GET /v1/publico/tiendas/{slug}/turnos/{codigo}", func(w http.ResponseWriter, r *http.Request) {
		t, err := Ver(r.Context(), pool, r.PathValue("slug"), r.PathValue("codigo"), ahora())
		responder(w, http.StatusOK, t, err)
	})
	mux.HandleFunc("DELETE /v1/publico/tiendas/{slug}/turnos/{codigo}", func(w http.ResponseWriter, r *http.Request) {
		t, err := Cancelar(r.Context(), pool, r.PathValue("slug"), r.PathValue("codigo"), ahora())
		responder(w, http.StatusOK, t, err)
	})
}
