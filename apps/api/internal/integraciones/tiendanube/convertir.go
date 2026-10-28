package tiendanube

import (
	"encoding/json"
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/importar"
)

// StockIlimitado es lo que se carga cuando Tiendanube no lleva el stock (stock_management false o
// stock null). El mismo valor que usa el importador de Empretienda.
const StockIlimitado = 999

// Texto es un campo por idioma ({"es": "...", "pt": "..."}). Tiendanube los manda así aunque la tienda
// tenga un solo idioma.
type Texto map[string]string

// En devuelve el texto en el idioma pedido; si no está, en español; si no, el primero que haya (en
// orden alfabético, para que sea estable).
func (t Texto) En(idioma string) string {
	if v, ok := t[idioma]; ok && strings.TrimSpace(v) != "" {
		return v
	}
	if v, ok := t["es"]; ok && strings.TrimSpace(v) != "" {
		return v
	}
	claves := make([]string, 0, len(t))
	for k, v := range t {
		if strings.TrimSpace(v) != "" {
			claves = append(claves, k)
		}
	}
	sort.Strings(claves)
	if len(claves) == 0 {
		return ""
	}
	return t[claves[0]]
}

// ProductoTN es lo que se usa del producto de la API (GET /products).
type ProductoTN struct {
	ID          int64   `json:"id"`
	Name        Texto   `json:"name"`
	Description Texto   `json:"description"`
	Handle      Texto   `json:"handle"`
	Attributes  []Texto `json:"attributes"`
	Categories  []struct {
		ID            int64           `json:"id"`
		Name          Texto           `json:"name"`
		Parent        *int64          `json:"parent"`
		Subcategories json.RawMessage `json:"subcategories"`
	} `json:"categories"`
	Images []struct {
		ID       int64  `json:"id"`
		Src      string `json:"src"`
		Position int    `json:"position"`
	} `json:"images"`
	Variants []struct {
		ImageID          *int64  `json:"image_id"`
		Price            *string `json:"price"`
		PromotionalPrice *string `json:"promotional_price"`
		StockManagement  bool    `json:"stock_management"`
		Stock            *int    `json:"stock"`
		SKU              *string `json:"sku"`
		Values           []Texto `json:"values"`
	} `json:"variants"`
}

// Importado es un producto listo para guardar, con lo que hace falta para la migración: el id en
// Tiendanube y el path de su ficha vieja (para la redirección 301).
type Importado struct {
	Externo   string            `json:"externo"`
	URLVieja  string            `json:"urlVieja"`
	Producto  importar.Producto `json:"producto"`
	Recortado bool              `json:"recortado"`
}

// Producto convierte un producto de Tiendanube. ok es false si no se puede importar (sin nombre o
// sin variantes: la API siempre manda al menos una). Los topes los pone importar.Ajustar, igual que
// para las demás plataformas.
func Producto(p ProductoTN, idioma string) (Importado, bool) {
	if len(p.Variants) == 0 {
		return Importado{}, false
	}
	c := importar.Crudo{
		Nombre:      p.Name.En(idioma),
		Descripcion: importar.TextoRico(p.Description.En(idioma)),
	}
	// La categoría más específica: una que no sea madre de otra de las que trae el producto.
	madres := map[int64]bool{}
	for _, cat := range p.Categories {
		if cat.Parent != nil {
			madres[*cat.Parent] = true
		}
	}
	for _, cat := range p.Categories {
		if !madres[cat.ID] {
			c.Categoria = strings.TrimSpace(cat.Name.En(idioma))
			break
		}
	}
	// Fotos en el orden de la tienda. Las tiendas viejas las tienen en http y el CDN de Tiendanube
	// también las sirve en https (el editor sólo acepta https).
	imgs := append(p.Images[:0:0], p.Images...)
	sort.SliceStable(imgs, func(i, j int) bool { return imgs[i].Position < imgs[j].Position })
	fotoDe := map[int64]string{}
	for _, im := range imgs {
		src := strings.TrimSpace(im.Src)
		if strings.HasPrefix(src, "http://") {
			src = "https://" + strings.TrimPrefix(src, "http://")
		}
		c.Fotos = append(c.Fotos, src)
		fotoDe[im.ID] = src
	}

	// Precio y stock de cada variante: el de oferta si es menor; sin stock_management, ilimitado.
	vars := make([]importar.VarianteCruda, 0, len(p.Variants))
	for _, v := range p.Variants {
		precio := pesos(v.Price)
		if oferta := pesos(v.PromotionalPrice); oferta > 0 && (precio == 0 || oferta < precio) {
			precio = oferta
		}
		stock := StockIlimitado
		if v.StockManagement && v.Stock != nil {
			stock = max(0, *v.Stock)
		}
		vc := importar.VarianteCruda{Precio: precio, Stock: stock}
		if v.SKU != nil {
			vc.SKU = strings.TrimSpace(*v.SKU)
		}
		if v.ImageID != nil {
			vc.Foto = fotoDe[*v.ImageID]
		}
		for _, val := range v.Values {
			vc.Valores = append(vc.Valores, val.En(idioma))
		}
		vars = append(vars, vc)
	}
	c.Precio, c.Stock, c.SKU = vars[0].Precio, vars[0].Stock, vars[0].SKU
	for _, v := range vars {
		c.Precio = min(c.Precio, v.Precio)
	}

	// Sin atributos (o con una sola variante sin valores) es un producto simple.
	if len(p.Attributes) > 0 && len(vars[0].Valores) > 0 {
		for i, a := range p.Attributes {
			o := importar.Opcion{Nombre: strings.TrimSpace(a.En(idioma))}
			if o.Nombre == "" {
				o.Nombre = "Opción " + strconv.Itoa(i+1)
			}
			// Los valores, en el orden en que aparecen en las variantes.
			vistos := map[string]bool{}
			for _, v := range vars {
				if i < len(v.Valores) && v.Valores[i] != "" && !vistos[strings.ToLower(v.Valores[i])] {
					vistos[strings.ToLower(v.Valores[i])] = true
					o.Valores = append(o.Valores, v.Valores[i])
				}
			}
			c.Opciones = append(c.Opciones, o)
		}
		c.Variantes = vars
	}
	i, ok := importar.Ajustar(c)
	if !ok {
		return Importado{}, false
	}
	return Importado{Externo: strconv.FormatInt(p.ID, 10), URLVieja: urlVieja(p, idioma), Producto: i.Producto, Recortado: i.Recortado}, true
}

// urlVieja es el path de la ficha en la tienda de Tiendanube (/productos/<handle>/).
func urlVieja(p ProductoTN, idioma string) string {
	h := strings.Trim(p.Handle.En(idioma), "/ ")
	if h == "" {
		return ""
	}
	return "/productos/" + h + "/"
}

// pesos pasa "18900.50" a pesos enteros, redondeando. Vacío o inválido es 0.
func pesos(s *string) int64 {
	if s == nil {
		return 0
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(*s), 64)
	if err != nil || f < 0 || math.IsNaN(f) || math.IsInf(f, 0) {
		return 0
	}
	return int64(math.Round(f))
}

func espacios(s string) string { return strings.Join(strings.Fields(s), " ") }
