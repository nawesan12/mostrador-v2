package pedidos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
)

// Un pedido de la tienda online con Mercado Pago que todavía no se pagó no es una venta, como en
// Shopify (docs/decisiones.md §63): se ve «Esperando pago», no avisa ni cuenta en las ventas hasta
// que el pago (o la seña) se acredita, y a las 24 h se cancela solo y devuelve el stock y el cupón.

// EsperaPago: el mismo criterio que app.esperando_pago en la base (00085). Las consultas de ventas,
// clientes y métricas dejan afuera lo que está «Sin pagar» con Mercado Pago en la tienda online.
func EsperaPago(canal, pago, estadoPago string) bool {
	switch estadoPago {
	case "Aprobado", "Acreditada", "Cobrado", "Seña":
		return false
	}
	return canal == "web" && pago == "mp"
}

// metricasCupon: lo que el pedido había sumado a lo vendido y lo descontado del cupón (sólo si ya era
// una venta: uno que esperaba el pago de Mercado Pago no sumó nada).
type metricasCupon struct {
	confirmado       bool
	vendio, desconto int64
}

// cantidadDevuelta: lo que bajó el stock al vender (Descontado, si fue menos que lo pedido).
func cantidadDevuelta(it Item) int {
	if it.Descontado != nil {
		return *it.Descontado
	}
	return it.Cantidad
}

// devolver vuelve a sumar el stock de los ítems y libera el uso del cupón (si tenía), en la
// transacción del pedido que se cancela. Si ya era una venta, también resta del cupón lo vendido y
// lo descontado.
func devolver(ctx context.Context, t pgx.Tx, tiendaID string, items []byte, cupon *string, m metricasCupon) error {
	var ls []Item
	if err := json.Unmarshal(items, &ls); err != nil {
		return err
	}
	for _, it := range ls {
		n := cantidadDevuelta(it)
		if it.ProductoID == nil || n <= 0 {
			continue
		}
		var err error
		if it.VarianteID != nil {
			_, err = t.Exec(ctx, `UPDATE variantes SET stock = stock + $4
				WHERE tienda_id = $1 AND producto_id::text = $2 AND id::text = $3`, tiendaID, *it.ProductoID, *it.VarianteID, n)
		} else {
			_, err = t.Exec(ctx, `UPDATE productos SET stock = stock + $3
				WHERE tienda_id = $1 AND id::text = $2 AND opciones = '[]'::jsonb`, tiendaID, *it.ProductoID, n)
		}
		if err != nil {
			return err
		}
	}
	if cupon == nil {
		return nil
	}
	if !m.confirmado {
		m.vendio, m.desconto = 0, 0
	}
	_, err := t.Exec(ctx, `UPDATE cupones SET usos = greatest(usos - 1, 0), vendio = greatest(vendio - $3, 0),
		desconto = greatest(desconto - $4, 0) WHERE tienda_id = $1 AND codigo = $2`, tiendaID, *cupon, m.vendio, m.desconto)
	return err
}

// retomar es lo contrario de devolver (deshacer una cancelación): vuelve a bajar el stock, sólo si
// alcanza para cada línea (ErrSinStock si no), y vuelve a contar el uso del cupón.
func retomar(ctx context.Context, t pgx.Tx, tiendaID string, items []byte, cupon *string, m metricasCupon) error {
	var ls []Item
	if err := json.Unmarshal(items, &ls); err != nil {
		return err
	}
	for _, it := range ls {
		it.Cantidad = cantidadDevuelta(it)
		if it.ProductoID == nil || it.Cantidad <= 0 {
			continue
		}
		var tag pgconn.CommandTag
		var err error
		if it.VarianteID != nil {
			tag, err = t.Exec(ctx, `UPDATE variantes SET stock = stock - $4
				WHERE tienda_id = $1 AND producto_id::text = $2 AND id::text = $3 AND (stock IS NULL OR stock >= $4)`, tiendaID, *it.ProductoID, *it.VarianteID, it.Cantidad)
		} else {
			tag, err = t.Exec(ctx, `UPDATE productos SET stock = stock - $3
				WHERE tienda_id = $1 AND id::text = $2 AND opciones = '[]'::jsonb AND (stock IS NULL OR stock >= $3)`, tiendaID, *it.ProductoID, it.Cantidad)
		}
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			// El producto ya no está (borrado) o no alcanza: si existe con menos stock, no se puede.
			var hay bool
			if err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM productos WHERE tienda_id = $1 AND id::text = $2)`, tiendaID, *it.ProductoID).Scan(&hay); err != nil {
				return err
			}
			if hay {
				return ErrSinStock
			}
		}
	}
	if cupon == nil {
		return nil
	}
	if !m.confirmado {
		m.vendio, m.desconto = 0, 0
	}
	_, err := t.Exec(ctx, `UPDATE cupones SET usos = usos + 1, vendio = vendio + $3, desconto = desconto + $4
		WHERE tienda_id = $1 AND codigo = $2`, tiendaID, *cupon, m.vendio, m.desconto)
	return err
}

// CancelarSinPago cancela un pedido de la tienda online que espera el pago de Mercado Pago hace más
// de 24 h: el stock vuelve y el cupón se libera. No avisa a nadie (no fue una venta); queda en la
// Actividad. Si mientras tanto se pagó, no hace nada y devuelve false.
func CancelarSinPago(ctx context.Context, pool tx, tiendaID string, numero int) (bool, error) {
	var hecho bool
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var items []byte
		var cupon *string
		err := t.QueryRow(ctx, `UPDATE pedidos SET estado = 'Cancelado',
				estados_en = estados_en || jsonb_build_object('Cancelado', now())
			WHERE tienda_id = $1 AND numero = $2 AND estado = 'Sin pagar' AND canal = 'web' AND pago = 'mp'
				AND sena IS NULL AND codigo_publico IS NOT NULL AND estado_pago = 'Pendiente' AND creado_en < now() - interval '24 hours'
				-- Con un pago en curso (efectivo, en revisión) se espera: ver la migración 00087.
				AND NOT EXISTS (SELECT 1 FROM mp_cobros c WHERE c.tienda_id = $1 AND c.tipo = 'pedido' AND c.objetivo = $3
					AND c.estado IN ('pending', 'in_process', 'authorized') AND c.actualizado_en > now() - interval '7 days')
			RETURNING items, cupon`, tiendaID, numero, strconv.Itoa(numero)).Scan(&items, &cupon)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := devolver(ctx, t, tiendaID, items, cupon, metricasCupon{}); err != nil {
			return err
		}
		hecho = true
		_, err = auditoria.Anotar(ctx, t, tiendaID, auditoria.Actor{Tipo: "sistema"}, auditoria.Registro{
			Accion: "pedido.vencido", Entidad: "pedido", EntidadID: fmt.Sprint(numero),
			Texto: fmt.Sprintf("canceló el pedido #%d: no se pagó con Mercado Pago en 24 horas (el stock volvió)", numero),
			Antes: map[string]string{"estado": "Sin pagar"}, Despues: map[string]string{"estado": "Cancelado"}})
		return err
	})
	return hecho, err
}

type txs interface {
	tx
	EnTx(context.Context, func(pgx.Tx) error) error
}

// VencerSinPago cancela, de a tandas, los pedidos que esperan el pago de Mercado Pago hace más de
// 24 h. Lo corre el worker cada minuto; el UPDATE condicionado evita que dos workers cancelen dos veces.
func VencerSinPago(ctx context.Context, pool txs) (int, error) {
	type par struct {
		tienda string
		numero int
	}
	var pares []par
	if err := pool.EnTx(ctx, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT tienda_id::text, numero FROM app.mp_sin_pagar_vencidos(200)`)
		if err != nil {
			return err
		}
		defer filas.Close()
		for filas.Next() {
			var p par
			if err := filas.Scan(&p.tienda, &p.numero); err != nil {
				return err
			}
			pares = append(pares, p)
		}
		return filas.Err()
	}); err != nil {
		return 0, err
	}
	n := 0
	for _, p := range pares {
		hecho, err := CancelarSinPago(ctx, pool, p.tienda, p.numero)
		if err != nil {
			return n, err
		}
		if hecho {
			n++
		}
	}
	return n, nil
}

// Pendiente es un pedido de la tienda online que espera el pago de Mercado Pago.
type Pendiente struct {
	Numero int
	Codigo string
	Total  int64
	items  []Item
	envio  string
	cupon  string
	entr   *Entrega
}

// EsperandoPagoDe: el pedido de Mercado Pago de ese email que espera el pago, armado en la última
// media hora (nil si no hay). Sin seña y sin un pago en curso (efectivo, en revisión).
func EsperandoPagoDe(ctx context.Context, pool tx, tiendaID, email string) (*Pendiente, error) {
	var p Pendiente
	var items, entrega []byte
	var cupon *string
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT numero, codigo_publico, total, items, envio, cupon, entrega FROM pedidos
			WHERE tienda_id = $1 AND canal = 'web' AND pago = 'mp' AND estado = 'Sin pagar' AND estado_pago = 'Pendiente'
				AND sena IS NULL AND codigo_publico IS NOT NULL AND lower(comprador->>'email') = lower($2)
				AND creado_en > now() - interval '30 minutes'
				AND NOT EXISTS (SELECT 1 FROM mp_cobros c WHERE c.tienda_id = $1 AND c.tipo = 'pedido' AND c.objetivo = numero::text
					AND c.estado IN ('pending', 'in_process', 'authorized', 'approved'))
			ORDER BY creado_en DESC LIMIT 1`, tiendaID, email).Scan(&p.Numero, &p.Codigo, &p.Total, &items, &p.envio, &cupon, &entrega)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	_ = json.Unmarshal(items, &p.items)
	if len(entrega) > 0 && string(entrega) != "null" {
		var e Entrega
		if json.Unmarshal(entrega, &e) == nil {
			p.entr = &e
		}
	}
	if cupon != nil {
		p.cupon = *cupon
	}
	return &p, nil
}

// Igual: el pedido nuevo es el mismo que el que espera el pago (los mismos productos, la misma
// entrega y el mismo cupón).
func (p *Pendiente) Igual(n PedidoNuevo) bool {
	if p.envio != n.Envio || !strings.EqualFold(p.cupon, n.Cupon) || len(p.items) != len(n.Items) {
		return false
	}
	clave := func(prod, vari *string, cant int) string {
		a, b := "", ""
		if prod != nil {
			a = *prod
		}
		if vari != nil {
			b = *vari
		}
		return a + "|" + b + "|" + strconv.Itoa(cant)
	}
	cuenta := map[string]int{}
	for _, it := range p.items {
		cuenta[clave(it.ProductoID, it.VarianteID, it.Cantidad)]++
	}
	for _, l := range n.Items {
		pid := l.ProductoID
		var vid *string
		if l.VarianteID != "" {
			v := l.VarianteID
			vid = &v
		}
		k := clave(&pid, vid, l.Cantidad)
		if cuenta[k] == 0 {
			return false
		}
		cuenta[k]--
	}
	a, b := Entrega{}, Entrega{}
	if p.entr != nil {
		a = *p.entr
	}
	if n.Entrega != nil {
		b = *n.Entrega
	}
	return a.Direccion == b.Direccion && a.CP == b.CP && a.Sucursal == b.Sucursal
}

// CancelarReemplazado cancela el pedido que esperaba el pago porque quien compraba armó otro: el
// stock y el cupón vuelven (no fue una venta). No avisa a nadie; queda en la Actividad.
func CancelarReemplazado(ctx context.Context, pool tx, tiendaID string, numero int) (bool, error) {
	var hecho bool
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var items []byte
		var cupon *string
		err := t.QueryRow(ctx, `UPDATE pedidos SET estado = 'Cancelado', estados_en = estados_en || jsonb_build_object('Cancelado', now())
			WHERE tienda_id = $1 AND numero = $2 AND estado = 'Sin pagar' AND canal = 'web' AND pago = 'mp'
				AND sena IS NULL AND estado_pago = 'Pendiente'
			RETURNING items, cupon`, tiendaID, numero).Scan(&items, &cupon)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if err := devolver(ctx, t, tiendaID, items, cupon, metricasCupon{}); err != nil {
			return err
		}
		hecho = true
		_, err = auditoria.Anotar(ctx, t, tiendaID, auditoria.Actor{Tipo: "sistema"}, auditoria.Registro{
			Accion: "pedido.reemplazado", Entidad: "pedido", EntidadID: fmt.Sprint(numero),
			Texto: fmt.Sprintf("canceló el pedido #%d: quien compraba armó otro antes de pagarlo (el stock volvió)", numero),
			Antes: map[string]string{"estado": "Sin pagar"}, Despues: map[string]string{"estado": "Cancelado"}})
		return err
	})
	return hecho, err
}
