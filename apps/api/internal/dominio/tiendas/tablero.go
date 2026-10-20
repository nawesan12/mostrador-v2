package tiendas

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Argentina no tiene horario de verano: UTC−3 fijo.
var horaArgentina = time.FixedZone("ART", -3*60*60)

// MetaMaxima: el tope de la meta del mes (100 mil millones de pesos).
const MetaMaxima = 100_000_000_000

// ErrMeta: la meta está fuera de rango.
var ErrMeta = errors.New("meta fuera de rango")

type comparado struct {
	Actual   float64 `json:"actual"`
	Anterior float64 `json:"anterior"`
}

// Tablero es el Resumen del panel (components/schemas/Tablero). Los números de ventas, pedidos,
// caja, envíos, bandeja y reseñas se llenan a medida que llegan esos dominios; hasta entonces van
// en cero y vacíos, nunca inventados. La meta del mes ya es real.
type Tablero struct {
	Periodo string    `json:"periodo"`
	Desde   time.Time `json:"desde"`
	Hasta   time.Time `json:"hasta"`
	Hola    *string   `json:"hola"`
	Ventas  comparado `json:"ventas"`
	Serie   struct {
		Actual   []float64 `json:"actual"`
		Anterior []float64 `json:"anterior"`
	} `json:"serie"`
	Kpis struct {
		Pedidos        comparado `json:"pedidos"`
		Ticket         comparado `json:"ticket"`
		Conversion     comparado `json:"conversion"`
		ClientesNuevos comparado `json:"clientesNuevos"`
	} `json:"kpis"`
	Canales []any `json:"canales"`
	Meta    struct {
		Objetivo   int64 `json:"objetivo"`
		Actual     int64 `json:"actual"`
		Proyeccion int64 `json:"proyeccion"`
	} `json:"meta"`
	Avisos struct {
		CertificadoDias *int `json:"certificadoDias"`
	} `json:"avisos"`
	UltimosPedidos []any `json:"ultimosPedidos"`
	ParaHacer      struct {
		Preparar            int   `json:"preparar"`
		Chats               int   `json:"chats"`
		ChatMasViejoMin     int   `json:"chatMasViejoMin"`
		Transferencias      int   `json:"transferencias"`
		TransferenciasMonto int64 `json:"transferenciasMonto"`
		SinStock            int   `json:"sinStock"`
		SinStockEnMl        int   `json:"sinStockEnMl"`
	} `json:"paraHacer"`
	EnVivo struct {
		Personas      int   `json:"personas"`
		Paginas       []any `json:"paginas"`
		Carritos      int   `json:"carritos"`
		CarritosMonto int64 `json:"carritosMonto"`
	} `json:"enVivo"`
	Caja struct {
		Abierta *time.Time `json:"abierta"`
		Total   int64      `json:"total"`
		Medios  []any      `json:"medios"`
	} `json:"caja"`
	Liquidaciones []any `json:"liquidaciones"`
	Hornero       struct {
		ChatsHoy      int      `json:"chatsHoy"`
		Ideas         int      `json:"ideas"`
		Carritos      int      `json:"carritos"`
		CarritosMonto int64    `json:"carritosMonto"`
		CostoMensaje  int      `json:"costoMensaje"`
		Recordados    bool     `json:"recordados"`
		Audiencia     int      `json:"audiencia"`
		IdeasHechas   []string `json:"ideasHechas"`
	} `json:"hornero"`
	Bandeja       []any `json:"bandeja"`
	Integraciones []any `json:"integraciones"`
	MasVendidos   []any `json:"masVendidos"`
	Resenas       struct {
		Promedio *float64 `json:"promedio"`
		Nuevas   []any    `json:"nuevas"`
	} `json:"resenas"`
	Envios struct {
		Paquetes int   `json:"paquetes"`
		Retiros  []any `json:"retiros"`
	} `json:"envios"`
}

// Ventana devuelve desde y hasta del período (hoy, 7d o 30d) en hora de Argentina.
func Ventana(periodo string, ahora time.Time) (time.Time, time.Time, error) {
	a := ahora.In(horaArgentina)
	switch periodo {
	case "hoy":
		desde := time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, horaArgentina)
		hasta := desde.Add(2 * time.Hour)
		if ahora.After(hasta) {
			hasta = ahora
		}
		return desde, hasta, nil
	case "7d":
		return ahora.Add(-7 * 24 * time.Hour), ahora, nil
	case "30d":
		return ahora.Add(-30 * 24 * time.Hour), ahora, nil
	}
	return time.Time{}, time.Time{}, errors.New("período inválido")
}

// ObtenerTablero arma el Resumen de la tienda para el período.
func ObtenerTablero(ctx context.Context, pool interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}, tiendaID, periodo string, ahora time.Time) (Tablero, error) {
	var t Tablero
	desde, hasta, err := Ventana(periodo, ahora)
	if err != nil {
		return t, err
	}
	t.Periodo, t.Desde, t.Hasta = periodo, desde, hasta
	t.Serie.Actual, t.Serie.Anterior = make([]float64, 14), make([]float64, 14)
	t.Canales, t.UltimosPedidos, t.Liquidaciones, t.Bandeja = []any{}, []any{}, []any{}, []any{}
	t.Integraciones, t.MasVendidos, t.EnVivo.Paginas, t.Caja.Medios = []any{}, []any{}, []any{}, []any{}
	t.Resenas.Nuevas, t.Envios.Retiros = []any{}, []any{}
	t.Hornero.CostoMensaje, t.Hornero.IdeasHechas = 120, []string{}
	err = pool.EnTienda(ctx, tiendaID, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT meta_mensual FROM tiendas WHERE id = $1`, tiendaID).Scan(&t.Meta.Objetivo)
	})
	return t, err
}

// GuardarMeta guarda la meta de ventas del mes (0 la saca).
func GuardarMeta(ctx context.Context, pool interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}, tiendaID string, objetivo int64) error {
	if objetivo < 0 || objetivo > MetaMaxima {
		return ErrMeta
	}
	return pool.EnTienda(ctx, tiendaID, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE tiendas SET meta_mensual = $2 WHERE id = $1`, tiendaID, objetivo)
		return err
	})
}

func rutasTablero(mux *http.ServeMux, pool interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}, tienda func(*http.Request) string) {
	mux.HandleFunc("GET /v1/panel/tablero", func(w http.ResponseWriter, r *http.Request) {
		id := tienda(r)
		if id == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		periodo := r.URL.Query().Get("periodo")
		if periodo == "" {
			periodo = "7d"
		}
		t, err := ObtenerTablero(r.Context(), pool, id, periodo, time.Now())
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			problema.Escribir(w, problema.SinSesion())
			return
		case err != nil && t.Periodo == "":
			problema.Escribir(w, problema.Invalido("El período es hoy, 7d o 30d."))
			return
		case err != nil:
			problema.Escribir(w, problema.Interno())
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		escribir(w, t)
	})
	mux.HandleFunc("PUT /v1/panel/meta", func(w http.ResponseWriter, r *http.Request) {
		id := tienda(r)
		if id == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		var cuerpo struct {
			Objetivo *int64 `json:"objetivo"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&cuerpo); err != nil || cuerpo.Objetivo == nil {
			problema.Escribir(w, problema.Invalido("Mandá la meta en pesos.", problema.Campo{Ruta: "objetivo", Mensaje: "Escribí la meta en pesos."}))
			return
		}
		switch err := GuardarMeta(r.Context(), pool, id, *cuerpo.Objetivo); {
		case err == nil:
			w.WriteHeader(http.StatusNoContent)
		case errors.Is(err, ErrMeta):
			problema.Escribir(w, problema.Invalido("La meta tiene que estar entre 0 y 100.000 millones.", problema.Campo{Ruta: "objetivo", Mensaje: "Fuera de rango."}))
		default:
			problema.Escribir(w, problema.Interno())
		}
	})
	// Sin ARCA conectado todavía no hay aviso del certificado: ocultarlo no tiene nada que hacer.
	mux.HandleFunc("POST /v1/panel/avisos/{aviso}/ocultar", func(w http.ResponseWriter, r *http.Request) {
		if tienda(r) == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		if r.PathValue("aviso") != "certificado" {
			problema.Escribir(w, problema.Invalido("No hay un aviso con ese nombre."))
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	// Lo que Hornero sabe de la tienda: se llena con cada dominio (pedidos, productos, clientes).
	mux.HandleFunc("GET /v1/panel/hornero/contexto", func(w http.ResponseWriter, r *http.Request) {
		if tienda(r) == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		escribir(w, map[string]any{
			"ventasPorDia": []any{}, "productos": []any{}, "canales30d": []any{},
			"clientes": map[string]int{"total": 0, "nuevos30d": 0, "conWhatsApp": 0, "sinVolver": 0},
		})
	})
	// Las ideas salen de carritos, stock y campañas: sin esos dominios todavía no hay ninguna.
	mux.HandleFunc("POST /v1/panel/hornero/ideas/{idea}", func(w http.ResponseWriter, r *http.Request) {
		if tienda(r) == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		problema.Escribir(w, problema.NoEncontrado("Esa idea no está disponible para tu tienda."))
	})
	// Sin carritos todavía (llegan con su dominio): no hay a quién recordarle nada.
	mux.HandleFunc("POST /v1/panel/carritos/recordar", func(w http.ResponseWriter, r *http.Request) {
		if tienda(r) == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		escribir(w, map[string]int{"enviados": 0})
	})
}
