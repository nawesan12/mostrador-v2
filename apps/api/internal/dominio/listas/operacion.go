package listas

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
)

// Inventario arma el stock de cada producto con lo que se vende por día (mismos cortes que el handoff).
func Inventario(productos []Registro) []Registro {
	regs := make([]Registro, 0, len(productos))
	for _, p := range productos {
		stock := int(numero(p["stock"]))
		porDia := math.Max(0.1, math.Round(numero(p["vendidos30"])/30*10)/10)
		// El mínimo es lo que se vende en 10 días (al menos 5): sale de las ventas, no de un número al azar.
		minimo := max(5, int(math.Round(porDia*10)))
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
		// Al abrir la fila, el stock por variante: el id "producto:variante" es el que aceptan las acciones.
		vs, _ := p["variantes"].([]Registro)
		porVariante := make([]Registro, 0, len(vs))
		for _, v := range vs {
			porVariante = append(porVariante, Registro{"id": fmt.Sprint(p["id"], ":", v["id"]), "varianteId": v["id"],
				"texto": v["texto"], "sku": v["sku"], "stock": v["stock"], "activa": v["activa"]})
		}
		regs = append(regs, Registro{
			"id": p["id"], "nombre": p["nombre"], "foto": p["foto"], "sku": p["sku"], "stock": stock, "min": minimo, "variantes": porVariante,
			"porDia": porDia, "cover": int(float64(stock) / porDia), "vendidos30": numero(p["vendidos30"]), "st": st, "fast": porDia > 2, "loc": "Local",
			"proveedor": "Sin proveedor", "costo": p["costo"], "publicadoEnMl": enMl && p["publicado"] == true,
		})
	}
	return regs
}

// DefInventario: cómo filtra y ordena Inventario.
var DefInventario = Def{
	Buscar: func(r Registro) string {
		return fmt.Sprint(r["nombre"], " ", r["sku"], " ", r["proveedor"]) + skusDe(r)
	},
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
		conVentas := 0
		for _, x := range todos {
			r["valorTotal"] += numero(x["stock"]) * numero(x["costo"])
			// La rotación sólo promedia lo que tiene ventas para proyectar (3 o más en 30 días, como el panel).
			if numero(x["vendidos30"]) >= 3 {
				r["rotacion"] += math.Min(120, numero(x["cover"]))
				conVentas++
			}
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
		if conVentas > 0 {
			r["rotacion"] = math.Round(r["rotacion"] / float64(conVentas))
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
		filas, err := t.Query(ctx, `SELECT numero, canal, i->>'nombre', i->>'variante', (i->>'cantidad')::int, creado_en
			FROM pedidos, jsonb_array_elements(items) i
			WHERE tienda_id = $1 AND estado <> 'Cancelado' AND creado_en >= $2
			ORDER BY creado_en DESC LIMIT 20`, tiendaID, desde)
		if err != nil {
			return err
		}
		for filas.Next() {
			var n, cant int
			var canal, nombre string
			var variante *string
			var en time.Time
			if err := filas.Scan(&n, &canal, &nombre, &variante, &cant, &en); err != nil {
				return err
			}
			if variante != nil && *variante != "" {
				nombre += " · " + *variante
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

// AjustarStock suma (o resta) unidades; nunca deja stock negativo. Los ids son de producto (sin
// opciones) o "producto:variante"; un producto con opciones no se ajusta entero (se ajusta cada variante).
func AjustarStock(ctx context.Context, pool tx, tiendaID string, ids []string, q int) (int, error) {
	if q == 0 || q > 100000 || q < -100000 {
		return 0, fmt.Errorf("ajuste fuera de rango")
	}
	prods, vars := separarIDs(ids)
	var n int64
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `UPDATE productos p SET stock = greatest(0, p.stock + $3)
			FROM productos a WHERE p.tienda_id = $1 AND p.id::text = ANY($2) AND p.opciones = '[]'::jsonb
				AND a.tienda_id = p.tienda_id AND a.id = p.id
			RETURNING p.id::text, p.nombre, a.stock, p.stock`, tiendaID, prods, q)
		if err != nil {
			return err
		}
		tocados, err := pgx.CollectRows(filas, pgx.RowToStructByPos[stockTocado])
		if err != nil {
			return err
		}
		n = int64(len(tocados))
		if len(vars) > 0 {
			filas, err = t.Query(ctx, `UPDATE variantes v SET stock = greatest(0, v.stock + $3)
				FROM variantes a, productos p WHERE v.tienda_id = $1 AND v.producto_id::text || ':' || v.id::text = ANY($2)
					AND a.tienda_id = v.tienda_id AND a.id = v.id AND p.tienda_id = v.tienda_id AND p.id = v.producto_id
				RETURNING v.producto_id::text, 'una variante de ' || p.nombre, a.stock, v.stock`, tiendaID, vars, q)
			if err != nil {
				return err
			}
			deVariantes, err := pgx.CollectRows(filas, pgx.RowToStructByPos[stockTocado])
			if err != nil {
				return err
			}
			n += int64(len(deVariantes))
			tocados = append(tocados, deVariantes...)
		}
		return registrarStock(ctx, t, tiendaID, tocados)
	})
	return int(n), err
}

// FijarStock deja el stock en lo contado (id de producto sin opciones o "producto:variante").
func FijarStock(ctx context.Context, pool tx, tiendaID, id string, cantidad int) (int, error) {
	if cantidad < 0 || cantidad > 1000000 {
		return 0, fmt.Errorf("cantidad fuera de rango")
	}
	consulta := `UPDATE productos p SET stock = $3 FROM productos a
		WHERE p.tienda_id = $1 AND p.id::text = $2 AND p.stock <> $3 AND p.opciones = '[]'::jsonb AND a.tienda_id = p.tienda_id AND a.id = p.id
		RETURNING p.id::text, p.nombre, a.stock, p.stock`
	if strings.Contains(id, ":") {
		consulta = `UPDATE variantes v SET stock = $3 FROM variantes a, productos p
			WHERE v.tienda_id = $1 AND v.producto_id::text || ':' || v.id::text = $2 AND v.stock <> $3
				AND a.tienda_id = v.tienda_id AND a.id = v.id AND p.tienda_id = v.tienda_id AND p.id = v.producto_id
			RETURNING v.producto_id::text, 'una variante de ' || p.nombre, a.stock, v.stock`
	}
	var n int64
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, consulta, tiendaID, id, cantidad)
		if err != nil {
			return err
		}
		tocados, err := pgx.CollectRows(filas, pgx.RowToStructByPos[stockTocado])
		if err != nil {
			return err
		}
		n = int64(len(tocados))
		return registrarStock(ctx, t, tiendaID, tocados)
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
		filas, err := t.Query(ctx, `UPDATE pedidos SET estado = $4, estados_en = estados_en || jsonb_build_object($4::text, now())
			WHERE tienda_id = $1 AND estado = $3 AND envio <> 'retiro' AND (coalesce(cardinality($2::text[]), 0) = 0 OR numero::text = ANY($2))
			RETURNING numero`, tiendaID, ids, de, a)
		if err != nil {
			return err
		}
		numeros, err := pgx.CollectRows(filas, pgx.RowTo[int])
		if err != nil {
			return err
		}
		n = int64(len(numeros))
		rs := make([]auditoria.Registro, 0, len(numeros))
		for _, num := range numeros {
			rs = append(rs, auditoria.Registro{Accion: "pedido.estado", Entidad: "pedido", EntidadID: fmt.Sprint(num),
				Texto: fmt.Sprintf("pasó el pedido #%d de %s a %s", num, strings.ToLower(de), strings.ToLower(a)),
				Antes: map[string]string{"estado": de}, Despues: map[string]string{"estado": a}})
		}
		return auditoria.Registrar(ctx, t, tiendaID, rs...)
	})
	return int(n), err
}
