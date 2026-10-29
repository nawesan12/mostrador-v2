package mercadolibre

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/importar"
)

// Importar las publicaciones que la tienda ya tiene en Mercado Libre (la migración). Es al revés de
// publicar.go: lee de Mercado Libre y no escribe nada allá. Las publicaciones se listan con la
// búsqueda "scan" (sirve para más de 1000 y no se puede mezclar con offset) y se piden de a 20 con el
// multiget de /items.

// porLote es cuántas publicaciones se piden juntas en /items?ids= (el tope del multiget).
const porLote = 20

// ItemML es lo que se usa de una publicación.
type ItemML struct {
	ID                string  `json:"id"`
	Title             string  `json:"title"`
	Price             float64 `json:"price"`
	AvailableQuantity int     `json:"available_quantity"`
	Status            string  `json:"status"`
	Permalink         string  `json:"permalink"`
	SellerCustomField *string `json:"seller_custom_field"`
	Pictures          []struct {
		ID        string `json:"id"`
		SecureURL string `json:"secure_url"`
	} `json:"pictures"`
	Attributes []struct {
		ID        string  `json:"id"`
		ValueName *string `json:"value_name"`
	} `json:"attributes"`
	Variations []struct {
		Price                 float64  `json:"price"`
		AvailableQuantity     int      `json:"available_quantity"`
		SellerCustomField     *string  `json:"seller_custom_field"`
		PictureIDs            []string `json:"picture_ids"`
		AttributeCombinations []struct {
			Name      string  `json:"name"`
			ValueName *string `json:"value_name"`
		} `json:"attribute_combinations"`
	} `json:"variations"`
	// Descripcion no viene en /items: sale de /items/{id}/description (plain_text).
	Descripcion string `json:"-"`
	// Categoria es el nombre de la categoría de Mercado Libre (la hoja), si se pudo leer.
	Categoria string `json:"-"`
}

// PaginaML es una vuelta de la búsqueda: las publicaciones y el scroll para la siguiente ("" si no hay
// más).
type PaginaML struct {
	Items     []ItemML
	Siguiente string
	Total     int
}

// LeerPublicaciones trae una página de publicaciones activas y pausadas de la tienda conectada, con
// su descripción. scroll es "" la primera vez y después lo que devolvió la página anterior.
func (s *Servicio) LeerPublicaciones(ctx context.Context, tiendaID, scroll string) (PaginaML, error) {
	token, usuario, err := s.tokenDe(ctx, tiendaID)
	if err != nil {
		return PaginaML{}, err
	}
	q := url.Values{"search_type": {"scan"}, "limit": {"100"}}
	if scroll != "" {
		q.Set("scroll_id", scroll)
	}
	var busqueda struct {
		Results  []string `json:"results"`
		ScrollID string   `json:"scroll_id"`
		Paging   struct {
			Total int `json:"total"`
		} `json:"paging"`
	}
	if err := s.pedir(ctx, http.MethodGet, fmt.Sprintf("/users/%d/items/search?%s", usuario, q.Encode()), token, nil, &busqueda); err != nil {
		return PaginaML{}, err
	}
	p := PaginaML{Total: busqueda.Paging.Total}
	// La búsqueda termina cuando devuelve una página vacía; con resultados, siempre hay scroll.
	if len(busqueda.Results) > 0 {
		p.Siguiente = busqueda.ScrollID
	}
	campos := "id,title,price,available_quantity,status,permalink,seller_custom_field,pictures,attributes,variations,category_id"
	categorias := map[string]string{}
	for i := 0; i < len(busqueda.Results); i += porLote {
		ids := busqueda.Results[i:min(i+porLote, len(busqueda.Results))]
		var lote []struct {
			Code int `json:"code"`
			Body struct {
				ItemML
				CategoryID string `json:"category_id"`
			} `json:"body"`
		}
		ruta := "/items?ids=" + url.QueryEscape(strings.Join(ids, ",")) + "&attributes=" + campos
		if err := s.pedir(ctx, http.MethodGet, ruta, token, nil, &lote); err != nil {
			return PaginaML{}, err
		}
		for _, r := range lote {
			if r.Code != http.StatusOK || r.Body.ID == "" {
				continue
			}
			// Las cerradas o dadas de baja no se traen: no tienen nada que vender.
			if r.Body.Status != "active" && r.Body.Status != "paused" {
				continue
			}
			it := r.Body.ItemML
			var d struct {
				PlainText string `json:"plain_text"`
			}
			// Sin descripción (404) se importa igual.
			if err := s.pedir(ctx, http.MethodGet, "/items/"+url.PathEscape(it.ID)+"/description", token, nil, &d); err == nil {
				it.Descripcion = strings.TrimSpace(d.PlainText)
			}
			if c := r.Body.CategoryID; c != "" {
				if _, ok := categorias[c]; !ok {
					var cat struct {
						Name string `json:"name"`
					}
					_ = s.pedir(ctx, http.MethodGet, "/categories/"+url.PathEscape(c), "", nil, &cat)
					categorias[c] = cat.Name
				}
				it.Categoria = categorias[c]
			}
			p.Items = append(p.Items, it)
		}
	}
	return p, nil
}

// Crudo pasa una publicación a lo que guarda Tu Local Web, antes de los topes (importar.Ajustar). Las
// variaciones son las combinaciones de atributos (Color, Talle) con su precio, stock y foto.
func (it ItemML) Crudo() importar.Crudo {
	fotos := map[string]string{}
	c := importar.Crudo{
		Nombre:      it.Title,
		Descripcion: it.Descripcion,
		Categoria:   it.Categoria,
		Precio:      int64(math.Round(it.Price)),
		Stock:       max(0, it.AvailableQuantity),
		SKU:         skuDe(it.SellerCustomField, it.Attributes),
	}
	for _, p := range it.Pictures {
		if p.SecureURL != "" {
			c.Fotos = append(c.Fotos, p.SecureURL)
			fotos[p.ID] = p.SecureURL
		}
	}
	if len(it.Variations) == 0 {
		return c
	}
	// Las opciones salen de la primera variación (todas traen los mismos atributos, en el mismo orden).
	for _, a := range it.Variations[0].AttributeCombinations {
		c.Opciones = append(c.Opciones, importar.Opcion{Nombre: a.Name})
	}
	for _, v := range it.Variations {
		vc := importar.VarianteCruda{
			Precio: int64(math.Round(v.Price)),
			Stock:  max(0, v.AvailableQuantity),
		}
		if v.SellerCustomField != nil {
			vc.SKU = strings.TrimSpace(*v.SellerCustomField)
		}
		if len(v.PictureIDs) > 0 {
			vc.Foto = fotos[v.PictureIDs[0]]
		}
		for i, o := range c.Opciones {
			val := ""
			for _, a := range v.AttributeCombinations {
				if a.Name == o.Nombre && a.ValueName != nil {
					val = strings.TrimSpace(*a.ValueName)
				}
			}
			vc.Valores = append(vc.Valores, val)
			if val != "" && !slices.Contains(c.Opciones[i].Valores, val) {
				c.Opciones[i].Valores = append(c.Opciones[i].Valores, val)
			}
		}
		c.Variantes = append(c.Variantes, vc)
	}
	c.Precio = c.Variantes[0].Precio
	for _, v := range c.Variantes {
		c.Precio = min(c.Precio, v.Precio)
	}
	return c
}

// skuDe: el SKU propio de la publicación (seller_custom_field) o el atributo SELLER_SKU.
func skuDe(campo *string, atributos []struct {
	ID        string  `json:"id"`
	ValueName *string `json:"value_name"`
}) string {
	if campo != nil && strings.TrimSpace(*campo) != "" {
		return strings.TrimSpace(*campo)
	}
	for _, a := range atributos {
		if a.ID == "SELLER_SKU" && a.ValueName != nil {
			return strings.TrimSpace(*a.ValueName)
		}
	}
	return ""
}
