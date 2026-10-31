package tiendas

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
)

// vendido: lo que cuenta como venta. Un pedido sin pagar todavía no lo es; uno cancelado ya no.
const vendido = `estado NOT IN ('Sin pagar', 'Cancelado')`

// llenarVentas pone en el tablero lo que sale de los pedidos y del catálogo: ventas, serie,
// indicadores, canales, meta del mes, últimos pedidos, lo que hay para hacer y los más vendidos.
func llenarVentas(ctx context.Context, tx pgx.Tx, tiendaID string, t *Tablero, desde, hasta, ahora time.Time) error {
	largo := hasta.Sub(desde)
	antes := desde.Add(-largo)

	// Ventas, pedidos y clientes nuevos del período y del anterior.
	periodo := func(d, h time.Time) (ventas float64, pedidos float64, nuevos float64, err error) {
		err = tx.QueryRow(ctx, `SELECT coalesce(sum(total), 0), count(*),
				count(DISTINCT lower(comprador->>'email')) FILTER (WHERE NOT EXISTS (
					SELECT 1 FROM pedidos q WHERE q.tienda_id = p.tienda_id AND q.creado_en < $2
						AND lower(q.comprador->>'email') = lower(p.comprador->>'email')))
			FROM pedidos p WHERE tienda_id = $1 AND creado_en >= $2 AND creado_en < $3 AND `+vendido,
			tiendaID, d, h).Scan(&ventas, &pedidos, &nuevos)
		return
	}
	va, pa, na, err := periodo(desde, hasta)
	if err != nil {
		return err
	}
	vb, pb, nb, err := periodo(antes, desde)
	if err != nil {
		return err
	}
	*t.Ventas = comparado{va, vb}
	t.Kpis.Pedidos = comparado{pa, pb}
	t.Kpis.ClientesNuevos = comparado{na, nb}
	ticket := func(v, p float64) float64 {
		if p == 0 {
			return 0
		}
		return float64(int64(v / p))
	}
	*t.Kpis.Ticket = comparado{ticket(va, pa), ticket(vb, pb)}
	// La conversión necesita las visitas: la llena llenarVisitas (tablero.go), con días enteros.

	// Serie: ventas acumuladas en 14 tramos, de este período y del anterior.
	filas, err := tx.Query(ctx, `SELECT creado_en, total FROM pedidos
		WHERE tienda_id = $1 AND creado_en >= $2 AND creado_en < $3 AND `+vendido, tiendaID, antes, hasta)
	if err != nil {
		return err
	}
	tramo := largo / 14
	for filas.Next() {
		var en time.Time
		var total float64
		if err := filas.Scan(&en, &total); err != nil {
			filas.Close()
			return err
		}
		serie, base := t.Serie.Actual, desde
		if en.Before(desde) {
			serie, base = t.Serie.Anterior, antes
		}
		i := 0
		if tramo > 0 {
			i = min(int(en.Sub(base)/tramo), 13)
		}
		for j := i; j < 14; j++ {
			serie[j] += total
		}
	}
	filas.Close()
	if err := filas.Err(); err != nil {
		return err
	}

	// Canales: la parte de cada uno en las ventas del período, de mayor a menor.
	if va > 0 {
		filas, err := tx.Query(ctx, `SELECT canal, sum(total) FROM pedidos
			WHERE tienda_id = $1 AND creado_en >= $2 AND creado_en < $3 AND `+vendido+`
			GROUP BY canal ORDER BY 2 DESC`, tiendaID, desde, hasta)
		if err != nil {
			return err
		}
		for filas.Next() {
			var canal string
			var monto float64
			if err := filas.Scan(&canal, &monto); err != nil {
				filas.Close()
				return err
			}
			t.Canales = append(t.Canales, map[string]any{"canal": canal, "parte": monto / va})
		}
		filas.Close()
		if err := filas.Err(); err != nil {
			return err
		}
	}

	// Meta: lo vendido en el mes y, al ritmo de los días que van, cómo cierra.
	a := ahora.In(horaArgentina)
	inicioMes := time.Date(a.Year(), a.Month(), 1, 0, 0, 0, 0, horaArgentina)
	if err := tx.QueryRow(ctx, `SELECT coalesce(sum(total), 0) FROM pedidos
		WHERE tienda_id = $1 AND creado_en >= $2 AND `+vendido, tiendaID, inicioMes).Scan(&t.Meta.Actual); err != nil {
		return err
	}
	diasMes := inicioMes.AddDate(0, 1, -1).Day()
	t.Meta.Proyeccion = t.Meta.Actual * int64(diasMes) / int64(a.Day())

	// Los últimos pedidos, de cualquier estado.
	filas, err = tx.Query(ctx, `SELECT numero, comprador->>'nombre', comprador->>'foto', canal, estado, total, creado_en
		FROM pedidos WHERE tienda_id = $1 ORDER BY creado_en DESC LIMIT 5`, tiendaID)
	if err != nil {
		return err
	}
	for filas.Next() {
		var numero int
		var cliente, canal, estado string
		var foto *string
		var total int64
		var creado time.Time
		if err := filas.Scan(&numero, &cliente, &foto, &canal, &estado, &total, &creado); err != nil {
			filas.Close()
			return err
		}
		t.UltimosPedidos = append(t.UltimosPedidos, map[string]any{"numero": numero, "cliente": cliente, "foto": foto,
			"canal": canal, "estado": estado, "total": total, "creado": creado})
	}
	filas.Close()
	if err := filas.Err(); err != nil {
		return err
	}

	// Para hacer: pedidos por preparar, transferencias por confirmar y productos sin stock.
	if err := tx.QueryRow(ctx, `SELECT
			count(*) FILTER (WHERE estado = 'Para preparar'),
			count(*) FILTER (WHERE estado = 'Sin pagar' AND pago = 'transf'),
			coalesce(sum(total) FILTER (WHERE estado = 'Sin pagar' AND pago = 'transf'), 0),
			count(*) FILTER (WHERE estado = 'Listo para despachar' AND envio <> 'retiro')
		FROM pedidos WHERE tienda_id = $1`, tiendaID).Scan(&t.ParaHacer.Preparar, &t.ParaHacer.Transferencias,
		t.ParaHacer.TransferenciasMonto, &t.Envios.Paquetes); err != nil {
		return err
	}
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM productos WHERE tienda_id = $1 AND publicado AND stock <= 0`,
		tiendaID).Scan(&t.ParaHacer.SinStock); err != nil {
		return err
	}

	// Los 4 más vendidos de los últimos 7 días (por nombre: el pedido guarda una copia del producto).
	filas, err = tx.Query(ctx, `WITH v AS (
			SELECT i->>'nombre' AS nombre, (array_agg(i->>'foto') FILTER (WHERE i->>'foto' IS NOT NULL))[1] AS foto,
				sum((i->>'cantidad')::int) AS unidades
			FROM pedidos, jsonb_array_elements(items) i
			WHERE tienda_id = $1 AND creado_en >= $2 AND `+vendido+`
			GROUP BY 1 ORDER BY 3 DESC, 1 LIMIT 4)
		SELECT coalesce((SELECT pr.id::text FROM productos pr WHERE pr.tienda_id = $1 AND pr.nombre = v.nombre LIMIT 1), v.nombre),
			v.nombre, v.foto, v.unidades
		FROM v ORDER BY v.unidades DESC, v.nombre`, tiendaID, ahora.Add(-7*24*time.Hour))
	if err != nil {
		return err
	}
	defer filas.Close()
	for filas.Next() {
		var id, nombre string
		var foto *string
		var unidades int
		if err := filas.Scan(&id, &nombre, &foto, &unidades); err != nil {
			return err
		}
		t.MasVendidos = append(t.MasVendidos, map[string]any{"productoId": id, "nombre": nombre, "foto": foto,
			"unidades": unidades})
	}
	return filas.Err()
}
