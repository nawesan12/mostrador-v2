// Package finanzas: la pantalla Finanzas (design/PanelOperacion.dc.html). Lo disponible, las
// liquidaciones y el retiro dependen de Mercado Pago y los demás medios; hasta conectarlos van en
// cero y el retiro responde 409. Ventas y margen salen de los pedidos y de los costos cargados: si a
// algo vendido le falta el costo, la ganancia no se estima (no se inventa un margen). Las
// transferencias marcadas como cobradas a mano no se dan por conciliadas.
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
	var ventas, vendido, ganado int64
	var sinCosto int
	movs := []Movimiento{}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if err := t.QueryRow(ctx, `SELECT coalesce(sum(total), 0) FROM pedidos WHERE tienda_id = $1
			AND estado NOT IN ('Cancelado', 'Sin pagar') AND creado_en > $2`, tiendaID, ahora.Add(-30*24*time.Hour)).Scan(&ventas); err != nil {
			return err
		}
		// La ganancia, de lo vendido de verdad: precio cobrado menos el costo cargado del producto.
		// Los ítems guardan el nombre; los que no tienen costo se cuentan aparte.
		if err := t.QueryRow(ctx, `SELECT coalesce(sum((i->>'precio')::bigint * (i->>'cantidad')::int), 0),
				coalesce(sum(((i->>'precio')::bigint - pr.costo) * (i->>'cantidad')::int) FILTER (WHERE pr.costo > 0), 0),
				count(DISTINCT i->>'nombre') FILTER (WHERE coalesce(pr.costo, 0) <= 0)
			FROM pedidos p CROSS JOIN LATERAL jsonb_array_elements(p.items) i
			LEFT JOIN LATERAL (SELECT costo FROM productos WHERE tienda_id = p.tienda_id AND nombre = i->>'nombre'
				ORDER BY costo DESC LIMIT 1) pr ON true
			WHERE p.tienda_id = $1 AND p.estado NOT IN ('Cancelado', 'Sin pagar') AND p.creado_en > $2`,
			tiendaID, ahora.Add(-30*24*time.Hour)).Scan(&vendido, &ganado, &sinCosto); err != nil {
			return err
		}
		// Lo que entró: las últimas ventas cobradas, con su medio (antes sólo las transferencias, así
		// que una tienda que cobraba en efectivo o con Mercado Pago veía «Movimientos» vacío).
		filas, err := t.Query(ctx, `SELECT numero, coalesce(comprador->>'nombre', ''), total, pago FROM pedidos WHERE tienda_id = $1
			AND estado NOT IN ('Cancelado', 'Sin pagar') ORDER BY creado_en DESC LIMIT 10`, tiendaID)
		if err != nil {
			return err
		}
		for filas.Next() {
			var n int
			var quien, pago string
			var total int64
			if err := filas.Scan(&n, &quien, &total, &pago); err != nil {
				return err
			}
			titulo := fmt.Sprintf("Pedido #%d", n)
			if quien != "" {
				titulo += " de " + quien
			}
			tipo, medio := pago, nombreMedio(pago)
			if pago == "tarjeta" {
				tipo = "visa"
			}
			detalle := medio
			if pago == "transf" {
				detalle += " · marcada como cobrada"
			}
			movs = append(movs, Movimiento{ID: fmt.Sprintf("pedido-%d", n), Tipo: tipo, Titulo: titulo, Detalle: detalle, Monto: total})
		}
		return filas.Err()
	})
	if err != nil {
		return nil, err
	}
	var ganancia *int64
	var margen *int
	detalle := "Sin ventas en 30 días"
	switch {
	case sinCosto > 0:
		detalle = fmt.Sprintf("Falta el costo de %d %s", sinCosto, plural(sinCosto, "producto vendido", "productos vendidos"))
	case vendido > 0:
		m := int(ganado * 100 / vendido)
		ganancia, margen = &ganado, &m
		detalle = ""
	}
	flujo := make([]Dia, 14)
	for i := range flujo {
		flujo[i] = Dia{Dia: ahora.In(argentina).AddDate(0, 0, i).Day()}
	}
	return map[string]any{
		"disponible": nil, "cuenta": nil, "banco": nil, "caja": nil,
		"kpis": map[string]any{
			// Lo disponible, lo que falta liquidar y los gastos todavía no se traen de ningún lado: el
			// panel los muestra como «sin dato», no como un cero.
			"aLiquidar": nil, "aLiquidarDetalle": "Lo ves en tu cuenta de Mercado Pago",
			"gastos": nil, "gastosDetalle": "Todavía no se registran gastos",
			"ganancia": ganancia, "margen": margen, "gananciaDetalle": detalle,
			"mes": meses[ahora.In(argentina).Month()-1],
		},
		"flujo": flujo,
		// Lo vendido y cobrado en 30 días: lo que sí se sabe mientras no se trae lo de Mercado Pago.
		"vendido30":   ventas,
		"movimientos": movs,
		// Hasta conectar los medios de pago no hay conciliación automática.
		"conciliado": false,
	}, nil
}

// nombreMedio: cómo se lee el medio de pago de un pedido.
func nombreMedio(pago string) string {
	switch pago {
	case "mp":
		return "Mercado Pago"
	case "transf":
		return "Transferencia"
	case "efectivo":
		return "Efectivo"
	case "tarjeta":
		return "Tarjeta"
	case "modo":
		return "MODO"
	}
	return "Otro medio"
}

func plural(n int, uno, varios string) string {
	if n == 1 {
		return uno
	}
	return varios
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
