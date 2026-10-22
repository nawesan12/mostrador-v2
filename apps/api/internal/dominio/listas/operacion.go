package listas

import (
	"context"
	"fmt"
	"hash/fnv"
	"math"
	"time"

	"github.com/jackc/pgx/v5"
)

func semilla(s string) float64 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return float64(h.Sum32()) / math.MaxUint32
}

// Inventario arma el stock de cada producto con lo que se vende por día (mismos cortes que el handoff).
func Inventario(productos []Registro) []Registro {
	regs := make([]Registro, 0, len(productos))
	for _, p := range productos {
		stock := int(numero(p["stock"]))
		porDia := math.Max(0.1, math.Round(numero(p["vendidos30"])/30*10)/10)
		minimo := max(5, int(math.Round(porDia*10+semilla(p["id"].(string))*10)))
		st := "ok"
		switch {
		case stock == 0:
			st = "out"
		case float64(stock) < float64(minimo)*0.5:
			st = "crit"
		case stock < minimo:
			st = "low"
		case float64(stock) > float64(minimo)*3.2:
			st = "over"
		}
		cs, _ := p["canales"].([]string)
		enMl := false
		for _, c := range cs {
			enMl = enMl || c == "ml"
		}
		regs = append(regs, Registro{
			"id": p["id"], "nombre": p["nombre"], "foto": p["foto"], "sku": p["sku"], "stock": stock, "min": minimo,
			"porDia": porDia, "cover": int(float64(stock) / porDia), "st": st, "fast": porDia > 2, "loc": "Local",
			"proveedor": "Sin proveedor", "costo": p["costo"], "publicadoEnMl": enMl && p["publicado"] == true,
		})
	}
	return regs
}

// DefInventario: cómo filtra y ordena Inventario.
var DefInventario = Def{
	Buscar:   func(r Registro) string { return fmt.Sprint(r["nombre"], " ", r["sku"], " ", r["proveedor"]) },
	Segmento: "st", Facetas: []string{"loc", "proveedor"}, Rango: "cover", RangoMenor: true, Banderas: []string{"fast"},
	Ordenes: []Orden{
		{"cover", func(a, b Registro) bool { return numero(a["cover"]) < numero(b["cover"]) }},
		{"stock", func(a, b Registro) bool { return numero(a["stock"]) < numero(b["stock"]) }},
		{"value", func(a, b Registro) bool {
			return numero(a["stock"])*numero(a["costo"]) > numero(b["stock"])*numero(b["costo"])
		}},
	},
	Resumen: func(todos, l []Registro) map[string]float64 {
		r := map[string]float64{}
		for _, x := range l {
			r["unidades"] += numero(x["stock"])
			r["valor"] += numero(x["stock"]) * numero(x["costo"])
		}
		prov := map[any]bool{}
		for _, x := range todos {
			r["valorTotal"] += numero(x["stock"]) * numero(x["costo"])
			r["rotacion"] += math.Min(120, numero(x["cover"]))
			switch x["st"] {
			case "out":
				r["sinStock"]++
				if x["publicadoEnMl"] == true {
					r["sinStockEnMl"]++
				}
			case "crit", "low":
				r["bajo"]++
			}
			if x["st"] == "out" || x["st"] == "crit" || x["st"] == "low" {
				r["reponer"]++
			}
			if x["st"] == "out" || numero(x["cover"]) < 7 {
				r["terminanSemana"]++
				prov[x["proveedor"]] = true
			}
		}
		if len(todos) > 0 {
			r["rotacion"] = math.Round(r["rotacion"] / float64(len(todos)))
		}
		r["proveedores"] = float64(len(prov))
		return r
	},
}

// Envios arma la lista de envíos con los pedidos que se despachan.
func Envios(ctx context.Context, pool tx, tiendaID string, ahora time.Time) ([]Registro, error) {
	regs := []Registro{}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT numero, comprador->>'nombre', ciudad_envio, envio, estado, creado_en,
				(estados_en->>'En camino')::timestamptz, coalesce(seguimiento, ''), costo_envio
			FROM pedidos WHERE tienda_id = $1 AND envio <> 'retiro'
				AND estado IN ('Listo para despachar', 'En camino', 'Entregado')`, tiendaID)
		if err != nil {
			return err
		}
		for filas.Next() {
			var numero int
			var cliente, ciudad, envio, estado, seg string
			var creado time.Time
			var despachado *time.Time
			var costo int64
			if err := filas.Scan(&numero, &cliente, &ciudad, &envio, &estado, &creado, &despachado, &seg, &costo); err != nil {
				return err
			}
			viaje := 0.0
			if despachado != nil {
				viaje = ahora.Sub(*despachado).Hours() / 24
			}
			step := 0
			switch {
			case estado == "Entregado":
				step = 4
			case estado == "En camino" && viaje < 1:
				step = 1
			case estado == "En camino" && viaje < 2:
				step = 2
			case estado == "En camino":
				step = 3
			}
			st := map[int]string{0: "todo", 4: "done"}[step]
			if st == "" {
				st = "way"
				if viaje > 3 && envio != "moto" {
					st = "late"
				}
			}
			if seg == "" {
				seg = "Sin seguimiento"
			}
			dias := ahora.Sub(creado).Hours() / 24
			regs = append(regs, Registro{
				"id": fmt.Sprint(numero), "pedido": fmt.Sprintf("#%d", numero), "cliente": cliente, "ciudad": ciudad,
				"transporte": envio, "step": step, "estado": st, "dias": dias, "costo": costo, "seguimiento": seg,
				"reclamado": false, "wk": dias < 7,
			})
		}
		return filas.Err()
	})
	return regs, err
}

// DefEnvios: cómo filtra y ordena Envíos.
var DefEnvios = Def{
	Buscar: func(r Registro) string {
		return fmt.Sprint(r["cliente"], " ", r["pedido"], " ", r["ciudad"], " ", r["seguimiento"])
	},
	Segmento: "estado", Facetas: []string{"transporte", "ciudad"}, Rango: "costo", Banderas: []string{"wk"},
	Ordenes: []Orden{
		{"new", func(a, b Registro) bool { return numero(a["dias"]) < numero(b["dias"]) }},
		{"late", func(a, b Registro) bool {
			la, lb := a["estado"] == "late", b["estado"] == "late"
			if la != lb {
				return la
			}
			return numero(a["dias"]) < numero(b["dias"])
		}},
		{"cost", func(a, b Registro) bool { return numero(a["costo"]) > numero(b["costo"]) }},
	},
	Resumen: func(todos, l []Registro) map[string]float64 {
		r := map[string]float64{}
		for _, x := range l {
			r["costos"] += numero(x["costo"])
		}
		for _, x := range todos {
			if x["estado"] == "late" {
				r["demorados"]++
			}
			if x["estado"] == "todo" {
				r["porDespachar"]++
			}
		}
		return r
	},
}

// Movimiento de stock (components/schemas/MovimientoStock).
type Movimiento struct {
	Tipo     string    `json:"tipo"`
	Texto    string    `json:"texto"`
	Cantidad int       `json:"cantidad"`
	En       time.Time `json:"en"`
}

// MovimientosHoy: las ventas de hoy, producto por producto (los ajustes llegan con el depósito).
func MovimientosHoy(ctx context.Context, pool tx, tiendaID string, ahora time.Time) ([]Movimiento, error) {
	ar := time.FixedZone("ART", -3*3600)
	a := ahora.In(ar)
	desde := time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, ar)
	movs := []Movimiento{}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT numero, canal, i->>'nombre', (i->>'cantidad')::int, creado_en
			FROM pedidos, jsonb_array_elements(items) i
			WHERE tienda_id = $1 AND estado <> 'Cancelado' AND creado_en >= $2
			ORDER BY creado_en DESC LIMIT 20`, tiendaID, desde)
		if err != nil {
			return err
		}
		for filas.Next() {
			var n, cant int
			var canal, nombre string
			var en time.Time
			if err := filas.Scan(&n, &canal, &nombre, &cant, &en); err != nil {
				return err
			}
			tipo, texto := "web", fmt.Sprintf("Venta #%d · %s", n, nombre)
			switch canal {
			case "pos":
				tipo, texto = "local", "Venta en local · "+nombre
			case "ml":
				tipo, texto = "ml", fmt.Sprintf("Mercado Libre #%d · %s", n, nombre)
			}
			movs = append(movs, Movimiento{Tipo: tipo, Texto: texto, Cantidad: -cant, En: en})
		}
		return filas.Err()
	})
	return movs, err
}

// AjustarStock suma (o resta) unidades; nunca deja stock negativo.
func AjustarStock(ctx context.Context, pool tx, tiendaID string, ids []string, q int) (int, error) {
	if q == 0 || q > 100000 || q < -100000 {
		return 0, fmt.Errorf("ajuste fuera de rango")
	}
	var n int64
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `UPDATE productos SET stock = greatest(0, stock + $3) WHERE tienda_id = $1 AND id::text = ANY($2)`, tiendaID, ids, q)
		n = tag.RowsAffected()
		return err
	})
	return int(n), err
}

// FijarStock deja el stock en lo contado.
func FijarStock(ctx context.Context, pool tx, tiendaID, id string, cantidad int) (int, error) {
	if cantidad < 0 || cantidad > 1000000 {
		return 0, fmt.Errorf("cantidad fuera de rango")
	}
	var n int64
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `UPDATE productos SET stock = $3 WHERE tienda_id = $1 AND id::text = $2 AND stock <> $3`, tiendaID, id, cantidad)
		n = tag.RowsAffected()
		return err
	})
	return int(n), err
}

// Etiquetas pasa a En camino los pedidos listos (o los vuelve, para deshacer). Sin ids: todos los listos.
func Etiquetas(ctx context.Context, pool tx, tiendaID string, ids []string, poner bool) (int, error) {
	de, a := "Listo para despachar", "En camino"
	if !poner {
		de, a = a, de
	}
	var n int64
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `UPDATE pedidos SET estado = $4, estados_en = estados_en || jsonb_build_object($4::text, now())
			WHERE tienda_id = $1 AND estado = $3 AND envio <> 'retiro' AND (coalesce(cardinality($2::text[]), 0) = 0 OR numero::text = ANY($2))`,
			tiendaID, ids, de, a)
		n = tag.RowsAffected()
		return err
	})
	return int(n), err
}
