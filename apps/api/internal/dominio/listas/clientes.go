package listas

import (
	"context"
	"hash/fnv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}

// Clientes arma la lista de clientes a partir de los pedidos (agrupados por email, o por nombre si
// no hay). Cuando llegue la ficha de cliente propia, se suma acá.
func Clientes(ctx context.Context, pool tx, tiendaID string, ahora time.Time) ([]Registro, error) {
	regs := []Registro{}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `
			SELECT coalesce(comprador->>'email', comprador->>'nombre') AS clave,
				(array_agg(comprador->>'nombre' ORDER BY creado_en DESC))[1],
				(array_agg(comprador->>'foto' ORDER BY creado_en DESC))[1],
				(array_agg(comprador->>'telefono' ORDER BY creado_en DESC))[1],
				(array_agg(ciudad_envio ORDER BY creado_en DESC))[1],
				(array_agg(canal ORDER BY creado_en))[1],
				count(*), max(creado_en), sum(total), bool_or(mayorista),
				array_agg(creado_en)
			FROM pedidos WHERE tienda_id = $1 AND estado <> 'Cancelado'
			GROUP BY 1`, tiendaID)
		if err != nil {
			return err
		}
		for filas.Next() {
			var clave, nombre, ciudad, canal string
			var foto, tel *string
			var pedidos int
			var ultima time.Time
			var gastado int64
			var mayorista bool
			var fechas []time.Time
			if err := filas.Scan(&clave, &nombre, &foto, &tel, &ciudad, &canal, &pedidos, &ultima, &gastado, &mayorista, &fechas); err != nil {
				return err
			}
			dias := ahora.Sub(ultima).Hours() / 24
			meses := make([]int, 12)
			for _, f := range fechas {
				k := 11 - int(ahora.Sub(f).Hours()/24/30)
				if k >= 0 && k < 12 && meses[k] < 3 {
					meses[k]++
				}
			}
			h := fnv.New32a()
			_, _ = h.Write([]byte(clave))
			regs = append(regs, Registro{
				"id": clave, "nombre": nombre, "foto": foto, "ini": int(h.Sum32() % 5), "ciudad": ciudad,
				"canal": canal, "telefono": tel, "pedidos": pedidos, "ultima": dias, "gastado": gastado,
				"seg": segmento(pedidos, dias), "meses": meses, "cuit": mayorista, "wa": tel != nil && *tel != "",
				"cupon": false, "etiquetas": []string{},
			})
		}
		return filas.Err()
	})
	return regs, err
}

// segmento: los mismos cortes que el handoff (PanelListas).
func segmento(pedidos int, dias float64) string {
	switch {
	case pedidos >= 12 && dias < 60:
		return "vip"
	case pedidos >= 4 && dias < 90:
		return "freq"
	case pedidos <= 1 && dias < 30:
		return "new"
	case dias > 180:
		return "sleep"
	case dias > 75:
		return "risk"
	}
	return "freq"
}

// DefClientes: cómo filtra y ordena Clientes.
var DefClientes = Def{
	Buscar: func(r Registro) string {
		t, _ := r["telefono"].(*string)
		tel := ""
		if t != nil {
			tel = *t
		}
		return strings.Join([]string{r["nombre"].(string), r["ciudad"].(string), tel}, " ")
	},
	Segmento: "seg", Facetas: []string{"canal", "ciudad"}, Rango: "gastado", Banderas: []string{"wa", "cuit"},
	Ordenes: []Orden{
		{"ltv", func(a, b Registro) bool { return numero(a["gastado"]) > numero(b["gastado"]) }},
		{"last", func(a, b Registro) bool { return numero(a["ultima"]) < numero(b["ultima"]) }},
		{"orders", func(a, b Registro) bool { return numero(a["pedidos"]) > numero(b["pedidos"]) }},
		{"az", func(a, b Registro) bool { return normal(a["nombre"].(string)) < normal(b["nombre"].(string)) }},
	},
	Resumen: func(todos, l []Registro) map[string]float64 {
		r := map[string]float64{}
		for _, x := range l {
			r["gastado"] += numero(x["gastado"])
		}
		for _, x := range todos {
			if x["seg"] == "risk" {
				r["enRiesgo"]++
			}
		}
		return r
	},
}
