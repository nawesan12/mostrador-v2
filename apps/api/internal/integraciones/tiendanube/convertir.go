package tiendanube

import (
	"encoding/json"
	"html"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
)

// StockIlimitado es lo que se carga cuando Tiendanube no lleva el stock (stock_management false o
// stock null). El mismo valor que usa el importador de Empretienda.
const StockIlimitado = 999

// Los topes del editor de productos (ProductoEditable en el contrato).
const (
	maxNombre      = 120
	maxDescripcion = 4000
	maxCategoria   = 60
	maxFotos       = 8
	maxOpciones    = 3
	maxValores     = 20
	maxTexto       = 20
	maxVariantes   = 100
	maxSKU         = 40
)

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
		Src      string `json:"src"`
		Position int    `json:"position"`
	} `json:"images"`
	Variants []struct {
		Price            *string `json:"price"`
		PromotionalPrice *string `json:"promotional_price"`
		StockManagement  bool    `json:"stock_management"`
		Stock            *int    `json:"stock"`
		SKU              *string `json:"sku"`
		Values           []Texto `json:"values"`
	} `json:"variants"`
}

// Importado es un producto listo para guardar, con lo que hace falta para la migración.
type Importado struct {
	Externo  string          `json:"externo"`
	URLVieja string          `json:"urlVieja"`
	Producto listas.Editable `json:"producto"`
	// Recortes cuenta lo que no entró en los topes de Mostrador (fotos, opciones, valores,
	// variantes, textos largos), para avisarle a la tienda.
	Recortes int `json:"recortes"`
}

// Producto convierte un producto de Tiendanube. ok es false si no se puede importar (sin nombre o
// sin variantes: la API siempre manda al menos una).
func Producto(p ProductoTN, idioma string) (Importado, bool) {
	nombre := espacios(p.Name.En(idioma))
	if len([]rune(nombre)) < 2 || len(p.Variants) == 0 {
		return Importado{}, false
	}
	recortes := 0
	cortar := func(s string, n int) string {
		r := []rune(s)
		if len(r) <= n {
			return s
		}
		recortes++
		return strings.TrimSpace(string(r[:n]))
	}

	e := listas.Editable{
		Nombre:  cortar(nombre, maxNombre),
		Canales: map[string]bool{"web": true},
	}
	if d := TextoDeHTML(p.Description.En(idioma)); d != "" {
		d = cortar(d, maxDescripcion)
		e.Descripcion = &d
	}
	// La categoría más específica: una que no sea madre de otra de las que trae el producto.
	madres := map[int64]bool{}
	for _, c := range p.Categories {
		if c.Parent != nil {
			madres[*c.Parent] = true
		}
	}
	for _, c := range p.Categories {
		if !madres[c.ID] {
			if n := espacios(c.Name.En(idioma)); n != "" {
				n = cortar(n, maxCategoria)
				e.Categoria = &n
			}
			break
		}
	}

	// Fotos en el orden de la tienda (position), sin repetidas. Las tiendas viejas las tienen en http
	// y el CDN de Tiendanube también las sirve en https: el editor sólo acepta https.
	imgs := append(p.Images[:0:0], p.Images...)
	sort.SliceStable(imgs, func(i, j int) bool { return imgs[i].Position < imgs[j].Position })
	vistas := map[string]bool{}
	for _, im := range imgs {
		src := strings.TrimSpace(im.Src)
		if strings.HasPrefix(src, "http://") {
			src = "https://" + strings.TrimPrefix(src, "http://")
		}
		if !strings.HasPrefix(src, "https://") || vistas[src] {
			continue
		}
		vistas[src] = true
		if len(e.Fotos) == maxFotos {
			recortes++
			continue
		}
		e.Fotos = append(e.Fotos, src)
	}
	if len(e.Fotos) > 0 {
		e.Foto = &e.Fotos[0]
	}

	// Precio y stock de cada variante: el de oferta si es menor; sin stock_management, ilimitado.
	type variante struct {
		valores []string
		precio  int64
		stock   int
		sku     string
	}
	vars := make([]variante, 0, len(p.Variants))
	for _, v := range p.Variants {
		precio := pesos(v.Price)
		if oferta := pesos(v.PromotionalPrice); oferta > 0 && (precio == 0 || oferta < precio) {
			precio = oferta
		}
		stock := StockIlimitado
		if v.StockManagement && v.Stock != nil {
			stock = max(0, *v.Stock)
		}
		sku := ""
		if v.SKU != nil {
			sku = cortar(strings.TrimSpace(*v.SKU), maxSKU)
		}
		valores := make([]string, 0, len(v.Values))
		for _, val := range v.Values {
			valores = append(valores, espacios(val.En(idioma)))
		}
		vars = append(vars, variante{valores, precio, stock, sku})
	}

	// El precio del producto es el más bajo; las variantes que cuestan otra cosa llevan el suyo.
	e.Precio = vars[0].precio
	for _, v := range vars {
		e.Precio = min(e.Precio, v.precio)
	}

	// Sin atributos (o con una sola variante sin valores) es un producto simple.
	nOpc := min(len(p.Attributes), maxOpciones)
	if len(p.Attributes) > maxOpciones {
		recortes++
	}
	if nOpc == 0 || len(vars[0].valores) == 0 {
		stock, sku := vars[0].stock, vars[0].sku
		e.Stock = &stock
		if sku != "" {
			e.SKU = &sku
		}
		return Importado{Externo: strconv.FormatInt(p.ID, 10), URLVieja: urlVieja(p, idioma), Producto: e, Recortes: recortes}, true
	}

	opciones := make([]listas.Opcion, nOpc)
	for i := range nOpc {
		n := espacios(p.Attributes[i].En(idioma))
		if n == "" {
			n = "Opción " + strconv.Itoa(i+1)
		}
		opciones[i].Nombre = cortar(n, maxTexto)
	}
	// Los valores, en el orden en que aparecen. Las variantes con un valor que no entra se descartan.
	usadas := map[string]bool{}
	for _, v := range vars {
		if len(v.valores) < nOpc {
			recortes++
			continue
		}
		if len(e.Variantes) == maxVariantes {
			recortes++
			continue
		}
		valores := map[string]string{}
		fuera := false
		for i := range nOpc {
			val := v.valores[i]
			if val == "" {
				fuera = true
				break
			}
			val = cortar(val, maxTexto)
			if ya, ok := igual(opciones[i].Valores, val); ok {
				val = ya
			} else if len(opciones[i].Valores) == maxValores {
				fuera = true
				break
			} else {
				opciones[i].Valores = append(opciones[i].Valores, val)
			}
			valores[opciones[i].Nombre] = val
		}
		clave := claveDe(valores, opciones)
		if fuera || usadas[clave] {
			recortes++
			continue
		}
		usadas[clave] = true
		ve := listas.VarianteEditable{Valores: valores, Stock: ptr(v.stock)}
		if v.sku != "" {
			ve.SKU = ptr(v.sku)
		}
		if v.precio != e.Precio {
			ve.Precio = ptr(v.precio)
		}
		e.Variantes = append(e.Variantes, ve)
	}
	e.Opciones = opciones
	return Importado{Externo: strconv.FormatInt(p.ID, 10), URLVieja: urlVieja(p, idioma), Producto: e, Recortes: recortes}, true
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

var (
	reSalto    = regexp.MustCompile(`(?i)<br\s*/?>|</p>|</li>|</h[1-6]>`)
	reEtiqueta = regexp.MustCompile(`<[^>]+>`)
	reEspacios = regexp.MustCompile(`[ \t\r\f\v]+`)
	reLineas   = regexp.MustCompile(`\n{3,}`)
)

// TextoDeHTML pasa la descripción en HTML a texto, con los párrafos como renglones.
func TextoDeHTML(s string) string {
	s = reSalto.ReplaceAllString(s, "\n")
	s = reEtiqueta.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, " ", " ")
	lineas := strings.Split(s, "\n")
	for i, l := range lineas {
		lineas[i] = strings.TrimSpace(reEspacios.ReplaceAllString(l, " "))
	}
	return strings.TrimSpace(reLineas.ReplaceAllString(strings.Join(lineas, "\n"), "\n\n"))
}

func espacios(s string) string { return strings.Join(strings.Fields(s), " ") }

// igual busca el valor sin mirar mayúsculas y devuelve cómo quedó escrito la primera vez.
func igual(xs []string, x string) (string, bool) {
	for _, v := range xs {
		if strings.EqualFold(v, x) {
			return v, true
		}
	}
	return "", false
}

func claveDe(valores map[string]string, opciones []listas.Opcion) string {
	var b strings.Builder
	for _, o := range opciones {
		b.WriteString(strings.ToLower(valores[o.Nombre]))
		b.WriteByte(0)
	}
	return b.String()
}

func ptr[T any](v T) *T { return &v }
