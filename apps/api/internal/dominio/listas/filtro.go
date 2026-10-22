// Package listas: el motor de listas del panel (components/schemas/ListaGenerica). Cada lista arma
// sus registros y el filtro genérico los recorta, ordena y cuenta: cada segmento y cada faceta se
// cuentan sin su propio filtro, igual que la demo (apps/web/src/lib/server/sandbox/listas.ts).
package listas

import (
	"fmt"
	"slices"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	"unicode"
)

// Registro es una fila de cualquier lista (se manda tal cual como JSON).
type Registro = map[string]any

// Filtros son los de la URL del panel.
type Filtros struct {
	Q        string
	Seg      []string
	Facetas  map[string][]string
	Rango    float64
	Banderas []string
	Orden    string
	Desde    int
	Cuantos  int
}

// Def dice cómo filtra y ordena una lista.
type Def struct {
	Buscar     func(Registro) string
	Segmento   string
	Facetas    []string
	Rango      string
	RangoMenor bool
	Banderas   []string
	// Ordenes en orden de preferencia: el primero es el de por defecto.
	Ordenes []Orden
	Resumen func(todos, filtrados []Registro) map[string]float64
}

// Orden es un criterio de orden con su clave de la URL.
type Orden struct {
	Clave string
	Menor func(a, b Registro) bool
}

// Lista es la respuesta (components/schemas/ListaGenerica).
type Lista struct {
	Total     int                       `json:"total"`
	Filas     []Registro                `json:"filas"`
	Segmentos map[string]int            `json:"segmentos"`
	Facetas   map[string]map[string]int `json:"facetas"`
	Resumen   map[string]float64        `json:"resumen"`
}

func lista(v string) []string {
	if v == "" {
		return nil
	}
	return strings.Split(v, ",")
}

// FiltrosDe lee los filtros de la query.
func FiltrosDe(q map[string][]string) Filtros {
	g := func(k string) string {
		if v := q[k]; len(v) > 0 {
			return v[0]
		}
		return ""
	}
	f := Filtros{Q: g("q"), Seg: lista(g("seg")), Banderas: lista(g("banderas")), Orden: g("orden"), Facetas: map[string][]string{}}
	if len(f.Q) > 100 {
		f.Q = f.Q[:100]
	}
	for _, kv := range lista(g("facetas")) {
		if k, v, ok := strings.Cut(kv, ":"); ok && k != "" {
			f.Facetas[k] = append(f.Facetas[k], v)
		}
	}
	f.Rango, _ = strconv.ParseFloat(g("rango"), 64)
	f.Desde, _ = strconv.Atoi(g("desde"))
	f.Cuantos, _ = strconv.Atoi(g("cuantos"))
	return f
}

var sinTildes = transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)

func normal(s string) string {
	t, _, err := transform.String(sinTildes, s)
	if err != nil {
		t = s
	}
	return strings.ToLower(t)
}

func numero(v any) float64 {
	switch x := v.(type) {
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case float64:
		return x
	}
	return 0
}

func valores(r Registro, k string) []string {
	switch v := r[k].(type) {
	case []string:
		return v
	case nil:
		return []string{""}
	default:
		return []string{fmt.Sprint(v)}
	}
}

// Aplicar filtra, ordena, cuenta y recorta la ventana.
func Aplicar(registros []Registro, d Def, f Filtros) Lista {
	q := normal(strings.TrimSpace(f.Q))
	base := make([]Registro, 0, len(registros))
	for _, r := range registros {
		if q != "" && !strings.Contains(normal(d.Buscar(r)), q) {
			continue
		}
		if f.Rango > 0 {
			v := numero(r[d.Rango])
			if (d.RangoMenor && v >= f.Rango) || (!d.RangoMenor && v < f.Rango) {
				continue
			}
		}
		ok := true
		for _, b := range f.Banderas {
			if slices.Contains(d.Banderas, b) && r[b] != true {
				ok = false
			}
		}
		if ok {
			base = append(base, r)
		}
	}
	elegidos := func(k string) []string {
		if k == d.Segmento {
			return f.Seg
		}
		return f.Facetas[k]
	}
	pasa := func(r Registro, sin string) bool {
		for _, k := range append([]string{d.Segmento}, d.Facetas...) {
			e := elegidos(k)
			if k == sin || len(e) == 0 {
				continue
			}
			if !slices.ContainsFunc(valores(r, k), func(v string) bool { return slices.Contains(e, v) }) {
				return false
			}
		}
		return true
	}
	contar := func(k string) map[string]int {
		m := map[string]int{}
		for _, r := range base {
			if pasa(r, k) {
				for _, v := range valores(r, k) {
					m[v]++
				}
			}
		}
		return m
	}
	filtrados := []Registro{}
	for _, r := range base {
		if pasa(r, "") {
			filtrados = append(filtrados, r)
		}
	}
	orden := d.Ordenes[0]
	for _, o := range d.Ordenes {
		if o.Clave == f.Orden {
			orden = o
		}
	}
	sort.SliceStable(filtrados, func(i, j int) bool { return orden.Menor(filtrados[i], filtrados[j]) })
	if f.Cuantos <= 0 || f.Cuantos > 200 {
		f.Cuantos = 60
	}
	desde := min(max(0, f.Desde), len(filtrados))
	hasta := min(len(filtrados), desde+f.Cuantos)
	l := Lista{Total: len(filtrados), Filas: filtrados[desde:hasta], Segmentos: contar(d.Segmento), Facetas: map[string]map[string]int{}}
	for _, k := range d.Facetas {
		l.Facetas[k] = contar(k)
	}
	l.Resumen = d.Resumen(registros, filtrados)
	return l
}
