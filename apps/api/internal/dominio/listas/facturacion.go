package listas

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

// Facturas son las acciones de la lista de Facturación (las implementa el paquete facturacion).
type Facturas interface {
	EmitirManual(ctx context.Context, tienda string, valor json.RawMessage, ahora time.Time) (int, error)
	Reintentar(ctx context.Context, tienda string, ids []string, ahora time.Time) (int, error)
	Anular(ctx context.Context, tienda string, ids []string, ahora time.Time) (int, error)
}

// Comprobantes de la tienda para la lista (lib/panel/listas/facturacion.ts: Comprobante).
func Comprobantes(ctx context.Context, pool tx, tiendaID string, ahora time.Time) ([]Registro, error) {
	regs := []Registro{}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT c.id::text, c.tipo, c.punto_venta, c.numero, c.estado, coalesce(c.motivo, ''), c.receptor->>'nombre',
				c.total, c.pedido_numero, coalesce(p.canal, 'web'), c.creado_en,
				EXISTS (SELECT 1 FROM comprobantes n WHERE n.asociado = c.id AND n.estado = 'emitido')
			FROM comprobantes c LEFT JOIN pedidos p ON p.tienda_id = c.tienda_id AND p.numero = c.pedido_numero
			WHERE c.tienda_id = $1 ORDER BY c.creado_en DESC LIMIT 5000`, tiendaID)
		if err != nil {
			return err
		}
		defer filas.Close()
		for filas.Next() {
			var id, estado, motivo, cliente, canal string
			var tipo int
			var pv, num, pedido *int
			var total int64
			var creado time.Time
			var anulada bool
			if err := filas.Scan(&id, &tipo, &pv, &num, &estado, &motivo, &cliente, &total, &pedido, &canal, &creado, &anulada); err != nil {
				return err
			}
			letra := map[int]string{1: "A", 3: "A", 6: "B", 8: "B", 11: "C", 13: "C"}[tipo]
			numero := "—"
			if pv != nil && num != nil {
				numero = fmt.Sprintf("%s %04d-%08d", letra, *pv, *num)
			}
			if tipo == 3 || tipo == 8 || tipo == 13 {
				numero = "NC " + numero
			}
			est := map[string]string{"emitido": "ok", "pendiente": "pend", "rechazado": "rej", "no_facturado": "nofac"}[estado]
			if anulada {
				est = "void"
			}
			dias := int(ahora.Sub(creado).Hours() / 24)
			per := "y"
			switch {
			case creado.Year() == ahora.Year() && creado.Month() == ahora.Month():
				per = "m"
			case dias < 92:
				per = "q"
			}
			r := Registro{"id": id, "numero": numero, "tipo": letra, "cliente": cliente, "dias": dias, "estado": est,
				"motivo": motivo, "total": total, "canal": canal, "enviado": false, "pedido": pedido, "per": per, "nomail": true}
			regs = append(regs, r)
		}
		return filas.Err()
	})
	return regs, err
}

// DefFacturacion es DEF_FACTURACION de la demo.
var DefFacturacion = Def{
	Buscar: func(r Registro) string {
		p := ""
		if n, ok := r["pedido"].(*int); ok && n != nil {
			p = fmt.Sprint(*n)
		}
		return fmt.Sprint(r["cliente"], " ", r["numero"], " ", p)
	},
	Segmento: "estado", Facetas: []string{"tipo", "canal", "per"}, Rango: "total", Banderas: []string{"nomail"},
	Ordenes: []Orden{
		{"new", func(a, b Registro) bool { return numero(a["dias"]) < numero(b["dias"]) }},
		{"total", func(a, b Registro) bool { return numero(a["total"]) > numero(b["total"]) }},
	},
	Resumen: func(todos, l []Registro) map[string]float64 {
		var fact, ult30, res float64
		for _, r := range l {
			if r["estado"] == "ok" {
				fact += numero(r["total"])
			}
		}
		for _, r := range todos {
			if r["estado"] == "ok" && numero(r["dias"]) < 30 {
				ult30 += numero(r["total"])
			}
			if r["estado"] == "pend" || r["estado"] == "rej" {
				res++
			}
		}
		return map[string]float64{"facturado": fact, "ultimos30": ult30, "porResolver": res}
	},
}
