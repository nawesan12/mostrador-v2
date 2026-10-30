package listas

import (
	"context"
	"errors"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"
)

var esID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// ErrEnMercadoLibre: el producto tiene una publicación en Mercado Libre; borrarlo la dejaría viva.
var ErrEnMercadoLibre = errors.New("el producto está publicado en Mercado Libre")

// Borrar saca el producto del catálogo.
func Borrar(ctx context.Context, pool tx, tiendaID, id string) error {
	if !esID.MatchString(id) {
		return ErrNoExiste
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var enMl bool
		if err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM ml_publicaciones WHERE tienda_id = $1 AND producto_id = $2)`,
			tiendaID, id).Scan(&enMl); err != nil {
			return err
		}
		if enMl {
			return ErrEnMercadoLibre
		}
		r, err := t.Exec(ctx, `DELETE FROM productos WHERE tienda_id = $1 AND id = $2`, tiendaID, id)
		if err != nil {
			return err
		}
		if r.RowsAffected() == 0 {
			return ErrNoExiste
		}
		return nil
	})
}

// Duplicar crea una copia pausada, con las mismas fotos, opciones y precios, sin stock ni SKU (son
// de cada producto).
func Duplicar(ctx context.Context, pool tx, tiendaID, id string, ahora time.Time) (Producto, error) {
	p, err := Obtener(ctx, pool, tiendaID, id, ahora)
	if err != nil {
		return Producto{}, err
	}
	nombre := []rune(p.Nombre + " (copia)")
	if len(nombre) > 120 {
		nombre = nombre[:120]
	}
	cero, vacio := 0, ""
	e := Editable{Nombre: string(nombre), Descripcion: &p.Descripcion, Categoria: p.Categoria, Precio: p.Precio,
		Costo: &p.Costo, SKU: &vacio, Stock: &cero, Foto: p.Foto, Canales: map[string]bool{"web": true}, Fotos: p.Fotos,
		Opciones: p.Opciones}
	for _, v := range p.Variantes {
		e.Variantes = append(e.Variantes, VarianteEditable{Valores: v.Valores, SKU: &vacio, Precio: v.Precio, Stock: &cero,
			Foto: v.Foto, Activa: &v.Activa})
	}
	c, err := Guardar(ctx, pool, tiendaID, "", e)
	if err != nil {
		return Producto{}, err
	}
	err = pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE productos SET publicado = false WHERE tienda_id = $1 AND id = $2`, tiendaID, c.ID)
		return err
	})
	c.Publicado = false
	return c, err
}
