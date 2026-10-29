package importar

import (
	"encoding/json"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

// Empretienda no tiene API: se lee la tienda pública como la ve un comprador. El sitemap lista las
// fichas (las URL con prioridad 0.8) y cada ficha trae su producto en variables de JavaScript
// (stock, imagenes, p_nombre, categorias_flatten, cdn_products) y la descripción en el HTML. Cada
// stock es una combinación: valoratributo dice el atributo (at_nombre) y su valor, y s_imagen la foto.
// Gemelo en la web: apps/web/src/lib/server/empretienda.ts.

// StockIlimitadoET es lo que se carga cuando Empretienda dice "sin límite".
const StockIlimitadoET = 999

// StockET es una combinación de la ficha.
type StockET struct {
	Cantidad      int   `json:"s_cantidad"`
	Ilimitado     int   `json:"s_ilimitado"`
	Precio        int64 `json:"s_precio"`
	Oferta        int   `json:"s_oferta"`
	PrecioOferta  int64 `json:"s_precio_oferta"`
	Imagen        int64 `json:"s_imagen"`
	ValorAtributo []struct {
		Atributo string `json:"at_nombre"`
		Valor    struct {
			Valor string `json:"vat_valor"`
		} `json:"valor"`
	} `json:"valoratributo"`
}

// FichaET es lo que se usa de una ficha (las variables de la página y su descripción).
type FichaET struct {
	URL      string `json:"url"`
	Nombre   string `json:"p_nombre"`
	CDN      string `json:"cdn_products"`
	Imagenes []struct {
		ID   int64  `json:"idImagenes"`
		Link string `json:"i_link"`
	} `json:"imagenes"`
	Categorias []struct {
		Nombre string `json:"c_nombre"`
		Link   string `json:"c_link_full"`
	} `json:"categorias_flatten"`
	Stock       []StockET `json:"stock"`
	Descripcion string    `json:"descripcion"`
}

// variableET lee `var nombre = <json>;` del HTML.
func variableET(html, nombre string, v any) bool {
	i := strings.Index(html, "var "+nombre+" = ")
	if i < 0 {
		return false
	}
	desde := i + len(nombre) + 7
	fin := strings.Index(html[desde:], ";\n")
	crudo := html[desde:]
	if fin >= 0 {
		crudo = crudo[:fin]
	}
	return json.Unmarshal([]byte(strings.TrimSuffix(strings.TrimSpace(crudo), ";")), v) == nil
}

var reDescripcionET = regexp.MustCompile(`class="product-vip__description[^"]*"[^>]*>([\s\S]*?)</div>`)

// FichaDeHTML saca la ficha de la página. ok es false si la página no es una ficha.
func FichaDeHTML(html, u string) (FichaET, bool) {
	f := FichaET{URL: u}
	if !variableET(html, "p_nombre", &f.Nombre) || !variableET(html, "stock", &f.Stock) || len(f.Stock) == 0 {
		return FichaET{}, false
	}
	variableET(html, "cdn_products", &f.CDN)
	variableET(html, "imagenes", &f.Imagenes)
	variableET(html, "categorias_flatten", &f.Categorias)
	if m := reDescripcionET.FindStringSubmatch(html); m != nil {
		f.Descripcion = strings.TrimSpace(m[1])
	}
	return f, true
}

var (
	reURLET       = regexp.MustCompile(`<url>([\s\S]*?)</url>`)
	reLocET       = regexp.MustCompile(`<loc>([^<]+)</loc>`)
	rePrioridadET = regexp.MustCompile(`<priority>([^<]+)</priority>`)
)

// FichasDelSitemap: en Empretienda las fichas son las URL con prioridad 0.8, y sólo las del origen.
func FichasDelSitemap(xml, origen string) []string {
	o, err := url.Parse(origen)
	if err != nil {
		return nil
	}
	var fichas []string
	for _, b := range reURLET.FindAllStringSubmatch(xml, -1) {
		loc, pr := reLocET.FindStringSubmatch(b[1]), rePrioridadET.FindStringSubmatch(b[1])
		if loc == nil || pr == nil || strings.TrimSpace(pr[1]) != "0.8" {
			continue
		}
		u, err := url.Parse(strings.TrimSpace(loc[1]))
		if err != nil || u.Hostname() != o.Hostname() || slices.Contains(fichas, u.String()) {
			continue
		}
		fichas = append(fichas, u.String())
	}
	return fichas
}

// DeEmpretienda pasa una ficha a un producto, con sus variantes reales (antes quedaban anotadas en la
// descripción). El precio de cada combinación es el de oferta si la tiene.
func DeEmpretienda(f FichaET) (Importado, bool) {
	camino := ""
	if u, err := url.Parse(f.URL); err == nil {
		camino = u.Path
	}
	c := Crudo{Nombre: f.Nombre, Descripcion: TextoRico(f.Descripcion)}
	// La categoría es la más larga cuyo link es prefijo del path de la ficha.
	largo := -1
	for _, cat := range f.Categorias {
		if strings.HasPrefix(camino, cat.Link+"/") && len(cat.Link) > largo {
			largo, c.Categoria = len(cat.Link), strings.TrimSpace(cat.Nombre)
		}
	}
	fotoDe := map[int64]string{}
	for _, im := range f.Imagenes {
		if im.Link != "" {
			fotoDe[im.ID] = f.CDN + im.Link
			c.Fotos = append(c.Fotos, f.CDN+im.Link)
		}
	}
	vars := make([]VarianteCruda, 0, len(f.Stock))
	for _, s := range f.Stock {
		v := VarianteCruda{Precio: s.Precio, Stock: max(0, s.Cantidad), Foto: fotoDe[s.Imagen]}
		if s.Oferta != 0 && s.PrecioOferta > 0 {
			v.Precio = s.PrecioOferta
		}
		if s.Ilimitado != 0 {
			v.Stock = StockIlimitadoET
		}
		for _, a := range s.ValorAtributo {
			v.Valores = append(v.Valores, a.Valor.Valor)
		}
		vars = append(vars, v)
	}
	c.Precio, c.Stock = vars[0].Precio, vars[0].Stock
	for _, v := range vars {
		c.Precio = min(c.Precio, v.Precio)
	}
	// Las opciones salen de la primera combinación (todas traen los mismos atributos, en orden).
	if attrs := f.Stock[0].ValorAtributo; len(attrs) > 0 {
		for i, a := range attrs {
			o := Opcion{Nombre: strings.TrimSpace(a.Atributo)}
			if o.Nombre == "" {
				o.Nombre = "Opción " + strconv.Itoa(i+1)
			}
			vistos := map[string]bool{}
			for _, v := range vars {
				if i < len(v.Valores) && strings.TrimSpace(v.Valores[i]) != "" && !vistos[Normal(v.Valores[i])] {
					vistos[Normal(v.Valores[i])] = true
					o.Valores = append(o.Valores, strings.TrimSpace(v.Valores[i]))
				}
			}
			c.Opciones = append(c.Opciones, o)
		}
		c.Variantes = vars
	}
	i, ok := Ajustar(c)
	i.Externo, i.URL = camino, camino
	return i, ok
}

// ResultadoEmpretienda resume las fichas de una página (lo que comparan los casos con la web).
func ResultadoEmpretienda(fichas []FichaET) Resultado {
	var items []Importado
	omitidos := 0
	for _, f := range fichas {
		if i, ok := DeEmpretienda(f); ok {
			items = append(items, i)
		} else {
			omitidos++
		}
	}
	return Resumir(items, omitidos)
}
