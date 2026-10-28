package importar

import "github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"

// Editable pasa el producto importado a lo que guarda el editor (listas.Editable). Ya viene dentro
// de los topes (Ajustar): acá sólo cambia la forma. Con opciones, el stock del producto lo calcula la
// base (la suma de las variantes) y no se manda.
func (p Producto) Editable() listas.Editable {
	e := listas.Editable{
		Nombre:    p.Nombre,
		Categoria: p.Categoria,
		Precio:    p.Precio,
		Canales:   p.Canales,
		Fotos:     p.Fotos,
		Opciones:  make([]listas.Opcion, 0, len(p.Opciones)),
	}
	if e.Canales == nil {
		e.Canales = map[string]bool{"web": true}
	}
	if e.Fotos == nil {
		e.Fotos = []string{}
	}
	if p.Descripcion != "" {
		e.Descripcion = &p.Descripcion
	}
	if p.SKU != "" {
		e.SKU = &p.SKU
	}
	if len(p.Fotos) > 0 {
		e.Foto = &p.Fotos[0]
	}
	for _, o := range p.Opciones {
		e.Opciones = append(e.Opciones, listas.Opcion{Nombre: o.Nombre, Valores: o.Valores})
	}
	if len(p.Opciones) == 0 {
		e.Stock = &p.Stock
		return e
	}
	for _, v := range p.Variantes {
		ve := listas.VarianteEditable{Valores: v.Valores, Precio: v.Precio, Stock: &v.Stock, Foto: v.Foto}
		if v.SKU != "" {
			ve.SKU = &v.SKU
		}
		e.Variantes = append(e.Variantes, ve)
	}
	return e
}
