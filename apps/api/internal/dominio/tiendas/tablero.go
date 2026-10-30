package tiendas

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/ventas"
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

// Tablero es el Resumen del panel (components/schemas/Tablero). Ventas, pedidos, canales, meta,
// para hacer y más vendidos salen de los pedidos (tablero_ventas.go); lo que todavía no se mide
// (visitas, caja, bandeja) va en cero y vacío, nunca inventado.
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
		if err := tx.QueryRow(ctx, `SELECT meta_mensual FROM tiendas WHERE id = $1`, tiendaID).Scan(&t.Meta.Objetivo); err != nil {
			return err
		}
		if err := llenarVentas(ctx, tx, tiendaID, &t, desde, hasta, ahora); err != nil {
			return err
		}
		// Carritos abandonados de la semana para recordar (Hornero los ofrece en el tablero).
		return tx.QueryRow(ctx, `SELECT count(*), coalesce(sum(total), 0), bool_or(recordado_en IS NOT NULL) IS TRUE
			FROM carritos WHERE tienda_id = $1 AND recuperado_en IS NULL AND actualizado > now() - interval '7 days'`, tiendaID).
			Scan(&t.Hornero.Carritos, &t.Hornero.CarritosMonto, &t.Hornero.Recordados)
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

// ContextoHornero es lo que Hornero sabe de la tienda además del tablero (el panel y Telegram usan
// el mismo): las ventas de los últimos 14 días, el catálogo con lo vendido, los clientes y los
// canales del mes. Si algo falla, va vacío: Tincho dice que no sabe, no inventa.
func ContextoHornero(ctx context.Context, pool interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}, tiendaID string) map[string]any {
	vacio := map[string]any{
		"ventasPorDia": []any{}, "productos": []any{}, "canales30d": []any{},
		"clientes": map[string]int{"total": 0, "nuevos30d": 0, "conWhatsApp": 0, "sinVolver": 0},
	}
	c, err := contextoHornero(ctx, pool, tiendaID, time.Now())
	if err != nil {
		return vacio
	}
	return c
}

func contextoHornero(ctx context.Context, pool interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}, tiendaID string, ahora time.Time) (map[string]any, error) {
	dias, productos, canales := []any{}, []any{}, []any{}
	clientes := map[string]int{}
	err := pool.EnTienda(ctx, tiendaID, func(tx pgx.Tx) error {
		a := ahora.In(horaArgentina)
		hoy := time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, horaArgentina)
		desde := hoy.AddDate(0, 0, -13)
		por := map[string][2]int64{}
		filas, err := tx.Query(ctx, `SELECT to_char(creado_en AT TIME ZONE 'America/Argentina/Buenos_Aires', 'YYYY-MM-DD'),
				sum(total), count(*)
			FROM pedidos WHERE tienda_id = $1 AND creado_en >= $2 AND `+vendido+` GROUP BY 1`, tiendaID, desde)
		if err != nil {
			return err
		}
		for filas.Next() {
			var f string
			var v, n int64
			if err := filas.Scan(&f, &v, &n); err != nil {
				filas.Close()
				return err
			}
			por[f] = [2]int64{v, n}
		}
		filas.Close()
		for d := desde; !d.After(hoy); d = d.AddDate(0, 0, 1) {
			f := d.Format("2006-01-02")
			dias = append(dias, map[string]any{"fecha": f, "ventas": por[f][0], "pedidos": por[f][1]})
		}

		hace30 := ahora.Add(-30 * 24 * time.Hour)
		filas, err = tx.Query(ctx, `WITH v AS (
				SELECT i->>'nombre' AS nombre, sum((i->>'cantidad')::int) AS n
				FROM pedidos, jsonb_array_elements(items) i
				WHERE tienda_id = $1 AND creado_en >= $2 AND `+vendido+` GROUP BY 1)
			SELECT p.nombre, p.categoria, p.precio, p.stock, coalesce(v.n, 0), coalesce((p.canales->>'ml')::boolean, false)
			FROM productos p LEFT JOIN v ON v.nombre = p.nombre
			WHERE p.tienda_id = $1 AND p.publicado ORDER BY coalesce(v.n, 0) DESC, p.nombre LIMIT 200`, tiendaID, hace30)
		if err != nil {
			return err
		}
		for filas.Next() {
			var nombre string
			var categoria *string
			var precio int64
			var stock, vendidos int
			var enMl bool
			if err := filas.Scan(&nombre, &categoria, &precio, &stock, &vendidos, &enMl); err != nil {
				filas.Close()
				return err
			}
			productos = append(productos, map[string]any{"nombre": nombre, "categoria": categoria, "precio": precio,
				"stock": stock, "vendidos30d": vendidos, "enMl": enMl})
		}
		filas.Close()

		var total, nuevos, conWa, sinVolver int
		if err := tx.QueryRow(ctx, `WITH c AS (
				SELECT lower(coalesce(comprador->>'email', comprador->>'nombre')) AS quien, min(creado_en) AS primero,
					max(creado_en) AS ultimo, count(*) AS veces, bool_or(comprador->>'telefono' IS NOT NULL) AS tel
				FROM pedidos WHERE tienda_id = $1 AND `+vendido+` GROUP BY 1)
			SELECT count(*), count(*) FILTER (WHERE primero >= $2), count(*) FILTER (WHERE tel),
				count(*) FILTER (WHERE veces > 1 AND ultimo < $2) FROM c`, tiendaID, hace30).
			Scan(&total, &nuevos, &conWa, &sinVolver); err != nil {
			return err
		}
		clientes = map[string]int{"total": total, "nuevos30d": nuevos, "conWhatsApp": conWa, "sinVolver": sinVolver}

		filas, err = tx.Query(ctx, `SELECT canal, sum(total) FROM pedidos
			WHERE tienda_id = $1 AND creado_en >= $2 AND `+vendido+` GROUP BY 1 ORDER BY 2 DESC`, tiendaID, hace30)
		if err != nil {
			return err
		}
		defer filas.Close()
		for filas.Next() {
			var canal string
			var v int64
			if err := filas.Scan(&canal, &v); err != nil {
				return err
			}
			canales = append(canales, map[string]any{"canal": canal, "ventas": v})
		}
		return filas.Err()
	})
	return map[string]any{"ventasPorDia": dias, "productos": productos, "clientes": clientes, "canales30d": canales}, err
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
		escribir(w, ContextoHornero(r.Context(), pool, tienda(r)))
	})
	// Las ideas salen de carritos, stock y campañas: sin esos dominios todavía no hay ninguna.
	mux.HandleFunc("POST /v1/panel/hornero/ideas/{idea}", func(w http.ResponseWriter, r *http.Request) {
		if tienda(r) == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		problema.Escribir(w, problema.NoEncontrado("Esa idea no está disponible para tu tienda."))
	})
	// Recordar a todos los carritos abandonados de la semana que no se recordaron hoy.
	mux.HandleFunc("POST /v1/panel/carritos/recordar", func(w http.ResponseWriter, r *http.Request) {
		id := tienda(r)
		if id == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		n, err := ventas.RecordarTodos(r.Context(), pool, id, time.Now())
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		escribir(w, map[string]int{"enviados": n})
	})
}
