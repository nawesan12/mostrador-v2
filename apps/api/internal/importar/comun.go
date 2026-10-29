// Package importar lee catálogos de otras plataformas (Shopify, WooCommerce) y los convierte en
// productos de Tu Local Web, ajustados a sus topes. Sólo los parsers: la cola que lee las tiendas y los
// endpoints vienen después.
//
// Es el gemelo de apps/web/src/lib/server/importar/: los dos pasan los mismos casos de
// packages/contrato/casos/importar/. Si cambia una regla, cambia en los dos lados.
package importar

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Topes de ProductoEditable (los mismos que valida listas).
const (
	topeNombre       = 120
	topeDescripcion  = 4000
	topeCategoria    = 60
	topeSKU          = 40
	topeFotos        = 8
	topeLargoFoto    = 600
	topeOpciones     = 3
	topeNombreOpcion = 20
	topeValores      = 20
	topeValor        = 20
	topeVariantes    = 100
	topeStock        = 1_000_000
)

// Opcion es "Talle" con sus valores.
type Opcion struct {
	Nombre  string   `json:"nombre"`
	Valores []string `json:"valores"`
}

// Variante es una combinación (components/schemas/VarianteEditable). Precio nil: el del producto.
type Variante struct {
	Valores map[string]string `json:"valores"`
	SKU     string            `json:"sku"`
	Precio  *int64            `json:"precio"`
	Stock   int               `json:"stock"`
	Foto    *string           `json:"foto"`
}

// Producto es lo que se guarda (components/schemas/ProductoEditable), con todos los campos.
type Producto struct {
	Nombre      string          `json:"nombre"`
	Descripcion string          `json:"descripcion"`
	Categoria   *string         `json:"categoria"`
	Precio      int64           `json:"precio"`
	SKU         string          `json:"sku"`
	Stock       int             `json:"stock"`
	Fotos       []string        `json:"fotos"`
	Opciones    []Opcion        `json:"opciones"`
	Variantes   []Variante      `json:"variantes"`
	Canales     map[string]bool `json:"canales"`
}

// VarianteCruda va por posición: Valores[i] es el valor de Crudo.Opciones[i].
type VarianteCruda struct {
	Valores []string
	SKU     string
	Precio  int64
	Stock   int
	Foto    string // "" sin foto
}

// Crudo es un producto como lo leyó el importador, antes de los topes.
type Crudo struct {
	Nombre      string
	Descripcion string
	Categoria   string // "" sin categoría
	Precio      int64  // en pesos enteros
	SKU         string
	Stock       int
	// Estimado: la plataforma no publica la cantidad; el stock es 1 (hay) o 0 (no hay).
	Estimado  bool
	Fotos     []string
	Opciones  []Opcion
	Variantes []VarianteCruda
}

// Importado es un producto listo para guardar, con lo que hay que avisar.
type Importado struct {
	Producto  Producto
	Recortado bool
	Estimado  bool
	// Externo es el id en la plataforma vieja y URL el path de su ficha allá (para la redirección
	// 301): los pone cada plataforma, no Ajustar.
	Externo, URL string
}

// Resultado es lo de una página (y lo que comparan los casos compartidos con la web).
type Resultado struct {
	Productos []Producto `json:"productos"`
	// StockEstimado: algún producto tiene el stock estimado; la interfaz pide revisarlo.
	StockEstimado bool `json:"stockEstimado"`
	// Recortes: cuántos productos se recortaron para entrar en los topes.
	Recortes int `json:"recortes"`
	// Omitidos: cuántos se descartaron (sin nombre).
	Omitidos int `json:"omitidos"`
}

// Resumir junta los importados de una página.
func Resumir(items []Importado, omitidos int) Resultado {
	r := Resultado{Productos: make([]Producto, 0, len(items)), Omitidos: omitidos}
	for _, i := range items {
		r.Productos = append(r.Productos, i.Producto)
		r.StockEstimado = r.StockEstimado || i.Estimado
		if i.Recortado {
			r.Recortes++
		}
	}
	return r
}

// ---------------------------------------------------------------- texto

// esEspacioJS es lo que saca String.prototype.trim() en JavaScript.
func esEspacioJS(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', ' ', 0xa0, 0x1680, 0x2028, 0x2029, 0x202f, 0x205f, 0x3000, 0xfeff:
		return true
	}
	return r >= 0x2000 && r <= 0x200a
}

func trimJS(s string) string { return strings.TrimFunc(s, esEspacioJS) }

var (
	reBr        = regexp.MustCompile(`(?i)<br\s*/?>`)
	reFinP      = regexp.MustCompile(`(?i)</p>`)
	reEtiqueta  = regexp.MustCompile(`<[^>]+>`)
	reApostrofe = regexp.MustCompile(`&#0?39;`)
	reSaltos    = regexp.MustCompile(`\n{3,}`)
	reEspacios  = regexp.MustCompile(`[ \t\n\r\f]+`)
	reBloque    = regexp.MustCompile(`(?i)</(li|div|h[1-6]|tr|ul|ol)>`)
	reNumerica  = regexp.MustCompile(`&#(\d{1,7});`)
	reHex       = regexp.MustCompile(`&#[xX]([0-9a-fA-F]{1,6});`)
	reNombrada  = regexp.MustCompile(`&([a-zA-Z]+);`)
	rePuntas    = regexp.MustCompile(`^\n+|\n+$`)
)

// Texto pasa HTML a texto: saltos en <br> y </p>, sin etiquetas y con las entidades de siempre
// (como texto() de la web).
func Texto(html string) string {
	t := reBr.ReplaceAllString(html, "\n")
	t = reFinP.ReplaceAllString(t, "\n")
	t = reEtiqueta.ReplaceAllString(t, "")
	for _, par := range [][2]string{{"&nbsp;", " "}, {"&amp;", "&"}, {"&lt;", "<"}, {"&gt;", ">"}, {"&quot;", `"`}} {
		t = strings.ReplaceAll(t, par[0], par[1])
	}
	t = reApostrofe.ReplaceAllString(t, "'")
	t = reSaltos.ReplaceAllString(t, "\n\n")
	return trimJS(t)
}

// Entidades con nombre que aparecen en los catálogos (el resto queda como vino).
var entidades = map[string]string{
	"ndash": "–", "mdash": "—", "lsquo": "‘", "rsquo": "’", "ldquo": "“", "rdquo": "”",
	"laquo": "«", "raquo": "»", "hellip": "…", "deg": "°", "ordm": "º", "ordf": "ª",
	"iexcl": "¡", "iquest": "¿", "aacute": "á", "eacute": "é", "iacute": "í", "oacute": "ó",
	"uacute": "ú", "Aacute": "Á", "Eacute": "É", "Iacute": "Í", "Oacute": "Ó", "Uacute": "Ú",
	"ntilde": "ñ", "Ntilde": "Ñ", "uuml": "ü", "Uuml": "Ü", "times": "×", "trade": "™",
	"reg": "®", "copy": "©",
}

// caracter: un código de entidad a su carácter ("" si no es válido).
func caracter(n int64) string {
	if n <= 0 || n > 0x10ffff || (n >= 0xd800 && n <= 0xdfff) {
		return ""
	}
	if n == 160 {
		return " "
	}
	return string(rune(n))
}

// TextoRico pasa el HTML de un catálogo a texto: los espacios del código no cuentan, los bloques
// cortan renglón y se entienden las entidades numéricas. Cada renglón sin espacios en las puntas.
func TextoRico(html string) string {
	t := reEspacios.ReplaceAllString(html, " ")
	t = Texto(reBloque.ReplaceAllString(t, "\n"))
	t = reNumerica.ReplaceAllStringFunc(t, func(e string) string {
		n, _ := strconv.ParseInt(reNumerica.FindStringSubmatch(e)[1], 10, 64)
		if c := caracter(n); c != "" {
			return c
		}
		return e
	})
	t = reHex.ReplaceAllStringFunc(t, func(e string) string {
		n, _ := strconv.ParseInt(reHex.FindStringSubmatch(e)[1], 16, 64)
		if c := caracter(n); c != "" {
			return c
		}
		return e
	})
	t = reNombrada.ReplaceAllStringFunc(t, func(e string) string {
		if c, ok := entidades[e[1:len(e)-1]]; ok {
			return c
		}
		return e
	})
	renglones := strings.Split(t, "\n")
	for i, r := range renglones {
		renglones[i] = strings.Trim(r, " \t")
	}
	t = reSaltos.ReplaceAllString(strings.Join(renglones, "\n"), "\n\n")
	return rePuntas.ReplaceAllString(t, "")
}

// ---------------------------------------------------------------- precios

var rePrecio = regexp.MustCompile(`^(\d+)(?:\.(\d+))?$`)

// Pesos pasa un precio en texto ("50.00", o "1250050" con 2 decimales) a pesos enteros, redondeando
// la mitad para arriba. ok es false si no es un número.
func Pesos(valor string, decimales int) (int64, bool) {
	m := rePrecio.FindStringSubmatch(trimJS(valor))
	if m == nil {
		return 0, false
	}
	digitos := strings.TrimLeft(m[1]+m[2], "0")
	if digitos == "" {
		digitos = "0"
	}
	d := decimales + len(m[2])
	if len(digitos) > 15 || d > 15 {
		return 0, false
	}
	n, err := strconv.ParseInt(digitos, 10, 64)
	if err != nil {
		return 0, false
	}
	div := int64(1)
	for range d {
		div *= 10
	}
	q := n / div
	if (n-q*div)*2 >= div {
		q++
	}
	return q, true
}

// ---------------------------------------------------------------- topes

var sinTildes = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

// Normal: sin mayúsculas ni tildes, como compara Tu Local Web opciones y valores repetidos.
func Normal(s string) string {
	t, _, err := transform.String(sinTildes, s)
	if err != nil {
		t = s
	}
	return strings.ToLower(t)
}

var reColapsar = regexp.MustCompile(`[ \t\n\r\f]+`)

func colapsar(s string) string { return strings.Trim(reColapsar.ReplaceAllString(s, " "), " ") }

func largo(s string) int { return utf8.RuneCountInString(s) }

func cortar(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		r = r[:n]
	}
	return strings.TrimRight(string(r), " ")
}

func entre(n, lo, hi int) int { return max(lo, min(hi, n)) }

// Ajustar deja un producto dentro de los topes de Tu Local Web. Lo que no entra se recorta y se marca:
// nombre y textos se cortan; fotos (sólo https), opciones y valores se limitan; si las
// combinaciones pasan de 100, se sacan valores de la opción más larga. Las variantes que quedan sin
// valor se descartan. ok es false si el nombre no llega a 2 letras.
func Ajustar(c Crudo) (Importado, bool) {
	rec := false
	corta := func(s string, n int) string {
		if largo(s) <= n {
			return s
		}
		rec = true
		return cortar(s, n)
	}
	nombreLimpio := colapsar(c.Nombre)
	if largo(nombreLimpio) < 2 {
		return Importado{}, false
	}
	p := Producto{
		Nombre:      corta(nombreLimpio, topeNombre),
		Descripcion: corta(c.Descripcion, topeDescripcion),
		Canales:     map[string]bool{"web": true},
		Fotos:       []string{},
		Opciones:    []Opcion{},
		Variantes:   []Variante{},
	}
	if cat := colapsar(c.Categoria); cat != "" {
		cat = corta(cat, topeCategoria)
		p.Categoria = &cat
	}
	sku := corta(trimJS(c.SKU), topeSKU)

	for _, f := range c.Fotos {
		f = trimJS(f)
		switch {
		case !strings.HasPrefix(f, "https://") || len(f) > topeLargoFoto:
			rec = true
		case slices.Contains(p.Fotos, f):
		case len(p.Fotos) == topeFotos:
			rec = true
		default:
			p.Fotos = append(p.Fotos, f)
		}
	}

	// Cada opción de la plataforma tiene su mapa valor (sin mayúsculas ni tildes) → valor final;
	// nil: la opción se descartó.
	mapas := make([]map[string]string, 0, len(c.Opciones))
	for _, o := range c.Opciones {
		n := corta(colapsar(o.Nombre), topeNombreOpcion)
		repetida := slices.ContainsFunc(p.Opciones, func(x Opcion) bool { return Normal(x.Nombre) == Normal(n) })
		if n == "" || len(p.Opciones) == topeOpciones || repetida {
			rec = true
			mapas = append(mapas, nil)
			continue
		}
		valores := []string{}
		mapa := map[string]string{}
		for _, crudo := range o.Valores {
			v := corta(colapsar(crudo), topeValor)
			if v == "" {
				continue
			}
			i := slices.IndexFunc(valores, func(x string) bool { return Normal(x) == Normal(v) })
			switch {
			case i >= 0:
				mapa[Normal(v)] = valores[i]
			case len(valores) == topeValores:
				rec = true
			default:
				valores = append(valores, v)
				mapa[Normal(v)] = v
			}
		}
		if len(valores) == 0 {
			rec = true
			mapas = append(mapas, nil)
			continue
		}
		p.Opciones = append(p.Opciones, Opcion{Nombre: n, Valores: valores})
		mapas = append(mapas, mapa)
	}
	vivos := make([]map[string]string, 0, len(p.Opciones))
	for _, m := range mapas {
		if m != nil {
			vivos = append(vivos, m)
		}
	}
	// Hasta 100 combinaciones: se saca el último valor de la opción con más (la última si empatan).
	combinaciones := func() int {
		t := 1
		for _, o := range p.Opciones {
			t *= len(o.Valores)
		}
		return t
	}
	for len(p.Opciones) > 0 && combinaciones() > topeVariantes {
		j := 0
		for i, o := range p.Opciones {
			if len(o.Valores) >= len(p.Opciones[j].Valores) {
				j = i
			}
		}
		vs := p.Opciones[j].Valores
		fuera := vs[len(vs)-1]
		p.Opciones[j].Valores = vs[:len(vs)-1]
		for k, v := range vivos[j] {
			if v == fuera {
				delete(vivos[j], k)
			}
		}
		rec = true
	}

	type variante struct {
		Variante
		precio int64
	}
	variantes := []variante{}
	vistas := map[string]bool{}
	if len(p.Opciones) > 0 {
		for _, v := range c.Variantes {
			valores := map[string]string{}
			ok := true
			k := 0
			for i, mapa := range mapas {
				if mapa == nil {
					continue
				}
				dado := ""
				if i < len(v.Valores) {
					dado = v.Valores[i]
				}
				if final, hay := mapa[Normal(cortar(colapsar(dado), topeValor))]; hay {
					valores[p.Opciones[k].Nombre] = final
				} else {
					ok = false
				}
				k++
			}
			partes := make([]string, len(p.Opciones))
			for i, o := range p.Opciones {
				partes[i] = valores[o.Nombre]
			}
			clave := strings.Join(partes, "\x00")
			if !ok || vistas[clave] {
				rec = true
				continue
			}
			vistas[clave] = true
			var foto *string
			if v.Foto != "" && slices.Contains(p.Fotos, v.Foto) {
				f := v.Foto
				foto = &f
			}
			variantes = append(variantes, variante{
				Variante: Variante{
					Valores: valores,
					SKU:     corta(trimJS(v.SKU), topeSKU),
					Stock:   entre(v.Stock, 0, topeStock),
					Foto:    foto,
				},
				precio: v.Precio,
			})
		}
	}

	if len(variantes) > 0 {
		p.Precio = variantes[0].precio
		for _, v := range variantes {
			p.Precio = min(p.Precio, v.precio)
			p.Stock += v.Stock
		}
		for _, v := range variantes {
			if v.precio != p.Precio {
				pr := v.precio
				v.Precio = &pr
			}
			p.Variantes = append(p.Variantes, v.Variante)
		}
	} else {
		p.Precio = c.Precio
		p.SKU = sku
		p.Stock = entre(c.Stock, 0, topeStock)
	}
	return Importado{Producto: p, Recortado: rec, Estimado: c.Estimado}, true
}
