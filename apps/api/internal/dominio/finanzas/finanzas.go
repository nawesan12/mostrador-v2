// Package finanzas: la pantalla Finanzas (design/PanelOperacion.dc.html). Lo disponible, las
// liquidaciones y el retiro dependen de Mercado Pago y los demás medios; hasta conectarlos van en
// cero y el retiro responde 409. Ventas, margen y transferencias conciliadas salen de los datos.
package finanzas

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}

var (
	argentina = time.FixedZone("ART", -3*3600)
	meses     = []string{"enero", "febrero", "marzo", "abril", "mayo", "junio", "julio", "agosto", "septiembre", "octubre", "noviembre", "diciembre"}
)

type Movimiento struct {
	ID      string `json:"id"`
	Tipo    string `json:"tipo"`
	Titulo  string `json:"titulo"`
	Detalle string `json:"detalle"`
	Monto   int64  `json:"monto"`
}

type Dia struct {
	Dia   int   `json:"dia"`
	Entra int64 `json:"entra"`
	Sale  int64 `json:"sale"`
}

// Leer arma la pantalla.
func Leer(ctx context.Context, pool tx, tiendaID string, ahora time.Time) (map[string]any, error) {
	var ventas int64
	var margen *int
	movs := []Movimiento{}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if err := t.QueryRow(ctx, `SELECT coalesce(sum(total), 0) FROM pedidos WHERE tienda_id = $1
			AND estado NOT IN ('Cancelado', 'Sin pagar') AND creado_en > $2`, tiendaID, ahora.Add(-30*24*time.Hour)).Scan(&ventas); err != nil {
			return err
		}
		// El margen sale de los costos cargados en los productos.
		if err := t.QueryRow(ctx, `SELECT round(avg((precio - costo) * 100.0 / precio))::int FROM productos
			WHERE tienda_id = $1 AND costo > 0 AND precio > 0`, tiendaID).Scan(&margen); err != nil {
			return err
		}
		filas, err := t.Query(ctx, `SELECT numero, comprador->>'nombre', total FROM pedidos WHERE tienda_id = $1 AND pago = 'transf'
			AND estado NOT IN ('Cancelado', 'Sin pagar') ORDER BY creado_en DESC LIMIT 10`, tiendaID)
		if err != nil {
			return err
		}
		for filas.Next() {
			var n int
			var quien string
			var total int64
			if err := filas.Scan(&n, &quien, &total); err != nil {
				return err
			}
			movs = append(movs, Movimiento{ID: fmt.Sprintf("transf-%d", n), Tipo: "transf", Titulo: "Transferencia de " + quien,
				Detalle: fmt.Sprintf("conciliada con #%d", n), Monto: total})
		}
		return filas.Err()
	})
	if err != nil {
		return nil, err
	}
	var ganancia int64
	if margen != nil {
		ganancia = ventas * int64(*margen) / 100
	}
	flujo := make([]Dia, 14)
	for i := range flujo {
		flujo[i] = Dia{Dia: ahora.In(argentina).AddDate(0, 0, i).Day()}
	}
	return map[string]any{
		"disponible": 0, "cuenta": nil, "banco": nil, "caja": nil,
		"kpis": map[string]any{
			"aLiquidar": 0, "aLiquidarDetalle": "Conectá tus medios de pago",
			"gastos": 0, "gastosDetalle": "Se suman con envíos y campañas",
			"ganancia": ganancia, "margen": margen,
			"mes": meses[ahora.In(argentina).Month()-1],
		},
		"flujo":       flujo,
		"movimientos": movs,
		"conciliado":  true,
	}, nil
}

// Rutas del panel.
func Rutas(mux *http.ServeMux, pool tx, tienda func(*http.Request) string) {
	mux.HandleFunc("GET /v1/panel/finanzas", func(w http.ResponseWriter, r *http.Request) {
		id := tienda(r)
		if id == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		v, err := Leer(r.Context(), pool, id, time.Now())
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	})
	mux.HandleFunc("POST /v1/panel/finanzas/retiro", func(w http.ResponseWriter, r *http.Request) {
		if tienda(r) == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		problema.Escribir(w, problema.Conflicto("sin_cobros", "Conectá Mercado Pago para transferir."))
	})
}
