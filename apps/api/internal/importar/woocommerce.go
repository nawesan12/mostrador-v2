package importar

import (
	"slices"
)

// WooCommerce publica su catálogo en la Store API (/wp-json/wc/store/v1/products?per_page=100&page=N),
// sin clave. Los precios vienen en unidades menores (currency_minor_unit). Las variaciones de los
// productos variables, con su precio y stock, se piden aparte con type=variation&parent=<ids>.
//
// Stock: low_stock_remaining si viene (sólo cuando la tienda lleva la cuenta y queda poco); si no,
// en stock → 1 y se avisa que es estimado; sin stock → 0.

type preciosWoo struct {
	Price             string `json:"price"`
	RegularPrice      string `json:"regular_price"`
	CurrencyMinorUnit int    `json:"currency_minor_unit"`
}

// atributoWoo es Color, Talle…; sólo los que tienen has_variations arman variantes.
type atributoWoo struct {
	Name          string `json:"name"`
	HasVariations bool   `json:"has_variations"`
	Terms         []struct {
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"terms"`
}

// ProductoWoo es un producto (o una variación) de la Store API (sólo lo que se usa).
type ProductoWoo struct {
	ID               int64       `json:"id"`
	Name             string      `json:"name"`
	Type             string      `json:"type"`
	SKU              string      `json:"sku"`
	Description      string      `json:"description"`
	ShortDescription string      `json:"short_description"`
	Prices           *preciosWoo `json:"prices"`
	Images           []struct {
		Src string `json:"src"`
	} `json:"images"`
	Categories []struct {
		Name string `json:"name"`
	} `json:"categories"`
	Attributes []atributoWoo `json:"attributes"`
	Variations []struct {
		ID         int64 `json:"id"`
		Attributes []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"attributes"`
	} `json:"variations"`
	IsInStock         bool `json:"is_in_stock"`
	LowStockRemaining *int `json:"low_stock_remaining"`
}

func precioWoo(p *preciosWoo) (int64, bool) {
	if p == nil {
		return 0, false
	}
	return Pesos(p.Price, p.CurrencyMinorUnit)
}

// stockWoo: el stock y si es estimado.
func stockWoo(p *ProductoWoo) (int, bool) {
	switch {
	case p.LowStockRemaining != nil:
		return max(0, *p.LowStockRemaining), false
	case p.IsInStock:
		return 1, true
	}
	return 0, false
}

// DeWoo convierte un producto de la Store API. variaciones son las que se pidieron aparte: las que
// falten quedan con el precio del producto y el stock que diga el producto.
func DeWoo(p ProductoWoo, variaciones []ProductoWoo) (Importado, bool) {
	atributos := []atributoWoo{}
	for _, a := range p.Attributes {
		if a.HasVariations && len(a.Terms) > 0 {
			atributos = append(atributos, a)
		}
	}
	variable := p.Type == "variable" && len(atributos) > 0
	stock, estimado := stockWoo(&p)
	precio, hayPrecio := precioWoo(p.Prices)
	c := Crudo{
		Nombre:      TextoRico(p.Name),
		Descripcion: TextoRico(p.Description),
		SKU:         p.SKU,
		Precio:      precio,
		Stock:       stock,
		Fotos:       []string{},
	}
	if p.Description == "" {
		c.Descripcion = TextoRico(p.ShortDescription)
	}
	if len(p.Categories) > 0 && p.Categories[0].Name != "" {
		c.Categoria = TextoRico(p.Categories[0].Name)
	}
	for _, i := range p.Images {
		c.Fotos = append(c.Fotos, i.Src)
	}
	if variable {
		for _, a := range atributos {
			o := Opcion{Nombre: TextoRico(a.Name)}
			for _, t := range a.Terms {
				o.Valores = append(o.Valores, TextoRico(t.Name))
			}
			c.Opciones = append(c.Opciones, o)
		}
		porID := map[int64]*ProductoWoo{}
		for i := range variaciones {
			porID[variaciones[i].ID] = &variaciones[i]
		}
		for _, v := range p.Variations {
			// Cada atributo: el valor (slug o nombre del término); vacío es "cualquiera".
			posibles := make([][]string, len(atributos))
			for i, a := range atributos {
				dado := ""
				for _, x := range v.Attributes {
					if Normal(x.Name) == Normal(a.Name) {
						dado = x.Value
						break
					}
				}
				if dado == "" {
					posibles[i] = c.Opciones[i].Valores
					continue
				}
				posibles[i] = []string{dado}
				for _, t := range a.Terms {
					if Normal(t.Slug) == Normal(dado) || Normal(t.Name) == Normal(dado) {
						posibles[i] = []string{TextoRico(t.Name)}
						break
					}
				}
			}
			d := porID[v.ID]
			s, e, sku, foto := stock, estimado, "", ""
			pv := precio
			if !hayPrecio {
				pv = 0
			}
			if d != nil {
				s, e = stockWoo(d)
				sku = d.SKU
				if n, ok := precioWoo(d.Prices); ok {
					pv = n
				}
				if len(d.Images) > 0 {
					foto = d.Images[0].Src
				}
			}
			c.Estimado = c.Estimado || e
			// La foto de la variación se suma a la galería si todavía no está.
			if foto != "" && !slices.Contains(c.Fotos, foto) {
				c.Fotos = append(c.Fotos, foto)
			}
			combos := [][]string{{}}
			for _, vals := range posibles {
				sig := make([][]string, 0, len(combos)*len(vals))
				for _, co := range combos {
					for _, x := range vals {
						sig = append(sig, append(slices.Clone(co), x))
					}
				}
				combos = sig
			}
			for _, valores := range combos {
				c.Variantes = append(c.Variantes, VarianteCruda{Valores: valores, SKU: sku, Precio: pv, Stock: s, Foto: foto})
			}
		}
	}
	c.Estimado = c.Estimado || estimado
	return Ajustar(c)
}

// PaginaWoo convierte una página de productos con las variaciones de sus variables.
func PaginaWoo(productos, variaciones []ProductoWoo) (items []Importado, leidos, omitidos int) {
	for _, p := range productos {
		if i, ok := DeWoo(p, variaciones); ok {
			items = append(items, i)
		}
	}
	return items, len(productos), len(productos) - len(items)
}

// ResultadoWoo resume una página de productos con sus variaciones.
func ResultadoWoo(productos, variaciones []ProductoWoo) Resultado {
	items, _, omitidos := PaginaWoo(productos, variaciones)
	return Resumir(items, omitidos)
}
