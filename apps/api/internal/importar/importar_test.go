package importar_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/importar"
)

// Los casos son los mismos que corre la web (apps/web/src/lib/server/importar/*.spec.ts).
var casos = filepath.Join("..", "..", "..", "..", "packages", "contrato", "casos", "importar")

func leer(t *testing.T, archivo string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(casos, archivo))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// igual compara como JSON: mismo contenido, sin importar el orden de las claves.
func igual(t *testing.T, got importar.Resultado, archivo string) {
	t.Helper()
	b, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	var g, q any
	if err := json.Unmarshal(b, &g); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(leer(t, archivo), &q); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(g, q) {
		e, _ := json.MarshalIndent(g, "", "  ")
		t.Fatalf("no coincide con %s; salió:\n%s", archivo, e)
	}
}

func TestShopifyIgualQueLaWeb(t *testing.T) {
	r, err := importar.ResultadoShopify(leer(t, "shopify.json"))
	if err != nil {
		t.Fatal(err)
	}
	igual(t, r, "shopify.esperado.json")
}

func TestWooCommerceIgualQueLaWeb(t *testing.T) {
	var d struct {
		Productos   []importar.ProductoWoo `json:"productos"`
		Variaciones []importar.ProductoWoo `json:"variaciones"`
	}
	if err := json.Unmarshal(leer(t, "woocommerce.json"), &d); err != nil {
		t.Fatal(err)
	}
	igual(t, importar.ResultadoWoo(d.Productos, d.Variaciones), "woocommerce.esperado.json")
}

func TestPesos(t *testing.T) {
	for _, c := range []struct {
		v    string
		dec  int
		want int64
		ok   bool
	}{
		{"50.00", 0, 50, true}, {"19.99", 0, 20, true}, {"12499.50", 0, 12500, true},
		{"1250049", 2, 12500, true}, {"1250050", 2, 12501, true}, {"", 0, 0, false}, {"-5", 0, 0, false},
	} {
		if n, ok := importar.Pesos(c.v, c.dec); n != c.want || ok != c.ok {
			t.Errorf("Pesos(%q, %d) = %d, %v; quería %d, %v", c.v, c.dec, n, ok, c.want, c.ok)
		}
	}
}

func TestTextoRico(t *testing.T) {
	got := importar.TextoRico("<p>Hola\n   mundo</p><ul><li>Uno</li><li>Dos&#8217;s</li></ul>")
	if got != "Hola mundo\nUno\nDos’s" {
		t.Fatalf("%q", got)
	}
	if got := importar.TextoRico("<p> </p>"); got != "" {
		t.Fatalf("%q", got)
	}
}

// El mismo caso que comun.spec.ts: más de 100 combinaciones, opciones de más y valores repetidos.
func TestAjustarCombinaciones(t *testing.T) {
	vs := func(n int) []string {
		out := []string{}
		for i := range n {
			out = append(out, "v"+strconv.Itoa(i))
		}
		return out
	}
	i, ok := importar.Ajustar(importar.Crudo{
		Nombre: "Producto", Precio: 100, Stock: 2,
		Opciones: []importar.Opcion{
			{Nombre: "Talle", Valores: vs(12)}, {Nombre: "Color", Valores: vs(10)},
			{Nombre: "Color", Valores: []string{"x"}}, {Nombre: "Otra", Valores: []string{"y"}},
			{Nombre: "Cuarta", Valores: []string{"z"}},
		},
		Variantes: []importar.VarianteCruda{
			{Valores: []string{"v0", "v0", "x", "y", "z"}, SKU: "A", Precio: 90, Stock: 3},
			{Valores: []string{"v11", "v0", "x", "y", "z"}, SKU: "B", Precio: 80, Stock: 1},
			{Valores: []string{"v1", "V0", "x", "y", "z"}, SKU: "C", Precio: 95, Stock: 1},
		},
	})
	if !ok || !i.Recortado {
		t.Fatal("tenía que recortarse")
	}
	p := i.Producto
	if len(p.Opciones) != 3 || len(p.Opciones[0].Valores) != 10 || len(p.Opciones[1].Valores) != 10 || p.Opciones[2].Nombre != "Otra" {
		t.Fatalf("opciones: %+v", p.Opciones)
	}
	if len(p.Variantes) != 2 || p.Variantes[0].SKU != "A" || p.Variantes[1].SKU != "C" {
		t.Fatalf("variantes: %+v", p.Variantes)
	}
	if p.Precio != 90 || p.Variantes[0].Precio != nil || *p.Variantes[1].Precio != 95 {
		t.Fatalf("precios: %d %+v", p.Precio, p.Variantes)
	}
}

func TestEmpretiendaIgualQueLaWeb(t *testing.T) {
	var d struct {
		Fichas []importar.FichaET `json:"fichas"`
	}
	if err := json.Unmarshal(leer(t, "empretienda.json"), &d); err != nil {
		t.Fatal(err)
	}
	r := importar.ResultadoEmpretienda(d.Fichas)
	// ACTUALIZAR=1 reescribe lo esperado (después hay que revisarlo y correr la web).
	if os.Getenv("ACTUALIZAR") == "1" {
		b, _ := json.MarshalIndent(r, "", "\t")
		if err := os.WriteFile(filepath.Join(casos, "empretienda.esperado.json"), append(b, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	igual(t, r, "empretienda.esperado.json")
}

func TestEmpretiendaHTML(t *testing.T) {
	html := "<script>\nvar p_nombre = \"Mate\";\nvar stock = [{\"s_cantidad\":2,\"s_precio\":5000,\"valoratributo\":[]}];\nvar cdn_products = \"https://cdn/\";\nvar imagenes = [{\"idImagenes\":1,\"i_link\":\"a.jpg\"}];\n</script><div class=\"product-vip__description\"><p>Calabaza</p></div>"
	f, ok := importar.FichaDeHTML(html, "https://t.empretienda.com.ar/mates/mate")
	if !ok || f.Nombre != "Mate" || len(f.Stock) != 1 || len(f.Imagenes) != 1 || f.Descripcion != "<p>Calabaza</p>" {
		t.Fatalf("ficha = %+v, %v", f, ok)
	}
	if _, ok := importar.FichaDeHTML("<html>no es una ficha</html>", "https://x"); ok {
		t.Fatal("una página sin variables no es una ficha")
	}
	xml := `<urlset><url><loc>https://t.empretienda.com.ar/mates/mate</loc><priority>0.8</priority></url><url><loc>https://t.empretienda.com.ar/mates</loc><priority>0.5</priority></url><url><loc>https://otra.com/x</loc><priority>0.8</priority></url></urlset>`
	if got := importar.FichasDelSitemap(xml, "https://t.empretienda.com.ar"); len(got) != 1 || got[0] != "https://t.empretienda.com.ar/mates/mate" {
		t.Fatalf("fichas = %v", got)
	}
}
