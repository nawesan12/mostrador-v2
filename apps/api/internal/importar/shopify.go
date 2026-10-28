package importar

import (
	"encoding/json"
	"slices"
)

// Shopify publica su catálogo sin clave en /products.json?limit=250&page=N. No publica cuánto stock
// hay, sólo si la variante está disponible: disponible → 1, no disponible → 0, y el resultado avisa
// que el stock es estimado. La única opción "Title" con "Default Title" es un producto sin variantes.

type varianteShopify struct {
	ID        int64   `json:"id"`
	Option1   *string `json:"option1"`
	Option2   *string `json:"option2"`
	Option3   *string `json:"option3"`
	SKU       *string `json:"sku"`
	Price     *string `json:"price"`
	Available *bool   `json:"available"`
}

type opcionShopify struct {
	Name     string   `json:"name"`
	Position int      `json:"position"`
	Values   []string `json:"values"`
}

// ProductoShopify es un producto de /products.json (sólo lo que se usa).
type ProductoShopify struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Handle      string            `json:"handle"`
	BodyHTML    *string           `json:"body_html"`
	ProductType *string           `json:"product_type"`
	Options     []opcionShopify   `json:"options"`
	Variants    []varianteShopify `json:"variants"`
	Images      []struct {
		Src        string  `json:"src"`
		VariantIDs []int64 `json:"variant_ids"`
	} `json:"images"`
}

func texto0(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func (p ProductoShopify) sinVariantes() bool {
	if len(p.Options) == 0 {
		return true
	}
	if len(p.Options) != 1 || p.Options[0].Name != "Title" {
		return false
	}
	for _, v := range p.Options[0].Values {
		if v != "Default Title" {
			return false
		}
	}
	return true
}

func stockShopify(v *varianteShopify) int {
	if v != nil && (v.Available == nil || *v.Available) {
		return 1
	}
	return 0
}

func precioShopify(v *varianteShopify) int64 {
	if v == nil || v.Price == nil {
		return 0
	}
	n, _ := Pesos(*v.Price, 0)
	return n
}

// DeShopify convierte un producto de /products.json (ok false si no tiene nombre).
func DeShopify(p ProductoShopify) (Importado, bool) {
	c := Crudo{
		Nombre:      p.Title,
		Descripcion: TextoRico(texto0(p.BodyHTML)),
		Categoria:   texto0(p.ProductType),
		Estimado:    true,
		Fotos:       make([]string, 0, len(p.Images)),
	}
	for _, i := range p.Images {
		c.Fotos = append(c.Fotos, i.Src)
	}
	var primera *varianteShopify
	if len(p.Variants) > 0 {
		primera = &p.Variants[0]
		c.SKU = texto0(primera.SKU)
	}
	c.Precio = precioShopify(primera)
	c.Stock = stockShopify(primera)
	if !p.sinVariantes() {
		opciones := slices.Clone(p.Options)
		slices.SortStableFunc(opciones, func(a, b opcionShopify) int {
			return a.Position - b.Position
		})
		for _, o := range opciones {
			c.Opciones = append(c.Opciones, Opcion{Nombre: o.Name, Valores: o.Values})
		}
		for i := range p.Variants {
			v := &p.Variants[i]
			valores := []string{texto0(v.Option1), texto0(v.Option2), texto0(v.Option3)}
			if len(c.Opciones) < len(valores) {
				valores = valores[:len(c.Opciones)]
			}
			foto := ""
			for _, img := range p.Images {
				if slices.Contains(img.VariantIDs, v.ID) {
					foto = img.Src
					break
				}
			}
			c.Variantes = append(c.Variantes, VarianteCruda{
				Valores: valores,
				SKU:     texto0(v.SKU),
				Precio:  precioShopify(v),
				Stock:   stockShopify(v),
				Foto:    foto,
			})
		}
	}
	return Ajustar(c)
}

// PaginaShopify convierte una página de /products.json: los importados, cuántos productos traía y
// cuántos se descartaron.
func PaginaShopify(cuerpo []byte) (items []Importado, leidos, omitidos int, err error) {
	var d struct {
		Products []ProductoShopify `json:"products"`
	}
	if err := json.Unmarshal(cuerpo, &d); err != nil {
		return nil, 0, 0, err
	}
	for _, p := range d.Products {
		if i, ok := DeShopify(p); ok {
			items = append(items, i)
		}
	}
	return items, len(d.Products), len(d.Products) - len(items), nil
}

// ResultadoShopify resume una página de /products.json.
func ResultadoShopify(cuerpo []byte) (Resultado, error) {
	items, _, omitidos, err := PaginaShopify(cuerpo)
	if err != nil {
		return Resultado{}, err
	}
	return Resumir(items, omitidos), nil
}
