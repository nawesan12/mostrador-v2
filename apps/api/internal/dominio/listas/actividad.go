package listas

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
)

// Lo que anotan los productos en la Actividad (docs/decisiones.md §39).

// antesProducto: lo que había antes de editar, para anotar el antes → después.
type antesProducto struct {
	nombre        string
	precio, costo int64
	stock         int
	publicado     bool
}

type publicacion struct {
	ID, Nombre string
	Antes      bool
}

type cambioPrecio struct {
	ID, Nombre     string
	Antes, Despues int64
}

// registrarEdicion anota lo que importa de una edición: precio, costo y stock con su valor anterior,
// si se volvió a publicar y, si no cambió nada de eso, que se editó.
func registrarEdicion(ctx context.Context, t pgx.Tx, tiendaID, id string, a antesProducto, e Editable, conStock bool) error {
	nombre := strings.TrimSpace(e.Nombre)
	reg := func(accion, texto string, antes, despues any) auditoria.Registro {
		return auditoria.Registro{Accion: accion, Entidad: "producto", EntidadID: id, Texto: texto, Antes: antes, Despues: despues}
	}
	var rs []auditoria.Registro
	if nombre != a.nombre {
		rs = append(rs, reg("producto.editar", fmt.Sprintf("le cambió el nombre a %s: ahora es %s", a.nombre, nombre),
			map[string]string{"nombre": a.nombre}, map[string]string{"nombre": nombre}))
	}
	if !e.SinPrecios && e.Precio != a.precio {
		rs = append(rs, reg("producto.precio", fmt.Sprintf("cambió el precio de %s de %s a %s", nombre,
			auditoria.Pesos(a.precio), auditoria.Pesos(e.Precio)), map[string]int64{"precio": a.precio}, map[string]int64{"precio": e.Precio}))
	}
	if !e.SinPrecios && e.Costo != nil && *e.Costo != a.costo {
		rs = append(rs, reg("producto.costo", fmt.Sprintf("cambió el costo de %s de %s a %s", nombre,
			auditoria.Pesos(a.costo), auditoria.Pesos(*e.Costo)), map[string]int64{"costo": a.costo}, map[string]int64{"costo": *e.Costo}))
	}
	if conStock && e.Stock != nil && *e.Stock != a.stock {
		rs = append(rs, reg("producto.stock", fmt.Sprintf("cambió el stock de %s de %d a %d", nombre, a.stock, *e.Stock),
			map[string]int{"stock": a.stock}, map[string]int{"stock": *e.Stock}))
	}
	if !a.publicado {
		rs = append(rs, reg("producto.publicar", "publicó "+nombre, map[string]bool{"publicado": false}, map[string]bool{"publicado": true}))
	}
	if len(rs) == 0 {
		rs = append(rs, reg("producto.editar", "editó "+nombre, nil, nil))
	}
	return auditoria.Registrar(ctx, t, tiendaID, rs...)
}

// stockTocado es una fila de stock ajustada: el producto (y la variante) con el antes y el después.
type stockTocado struct {
	ID, Nombre     string
	Antes, Despues int
}

// registrarStock anota cada ajuste de stock con su valor anterior.
func registrarStock(ctx context.Context, t pgx.Tx, tiendaID string, l []stockTocado) error {
	rs := make([]auditoria.Registro, 0, len(l))
	for _, x := range l {
		if x.Antes == x.Despues {
			continue
		}
		rs = append(rs, auditoria.Registro{Accion: "producto.stock", Entidad: "producto", EntidadID: x.ID,
			Texto: fmt.Sprintf("cambió el stock de %s de %d a %d", x.Nombre, x.Antes, x.Despues),
			Antes: map[string]int{"stock": x.Antes}, Despues: map[string]int{"stock": x.Despues}})
	}
	return auditoria.Registrar(ctx, t, tiendaID, rs...)
}
