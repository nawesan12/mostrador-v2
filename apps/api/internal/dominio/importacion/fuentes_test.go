package importacion

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/importar"
)

const casos = "../../../../../packages/contrato/casos/importar/"

func leerCaso(t *testing.T, nombre string) []byte {
	t.Helper()
	b, err := os.ReadFile(casos + nombre)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// Shopify: la página 1 trae los productos del caso; la 2, vacía. El servidor falso es http: las
// fuentes reciben el cliente (en producción, importar.Publico).
func TestFuenteShopify(t *testing.T) {
	caso := leerCaso(t, "shopify.json")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/products.json" || r.URL.Query().Get("limit") != "250" {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(caso)
	}))
	defer srv.Close()
	abrir := Fuentes(Dependencias{Publico: srv.Client()})["shopify"]
	f, err := abrir(context.Background(), Origen{Plataforma: "shopify", URL: srv.URL})
	if err != nil {
		t.Fatal(err)
	}
	p, err := f.Pagina(context.Background(), "productos", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Productos) == 0 || p.Siguiente != nil {
		t.Fatalf("página = %d productos, siguiente %s", len(p.Productos), p.Siguiente)
	}
	for _, x := range p.Productos {
		if x.Externo == "" || x.URL == "" || x.Producto.Nombre == "" {
			t.Fatalf("producto sin id, URL o nombre: %+v", x)
		}
	}
	// Las etapas que Shopify no tiene sin cuenta terminan enseguida.
	if c, err := f.Pagina(context.Background(), "clientes", nil); err != nil || len(c.Clientes) != 0 || c.Siguiente != nil {
		t.Fatalf("clientes = %+v, %v", c, err)
	}
}

func TestFuenteShopifyNoExiste(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()
	f, _ := Fuentes(Dependencias{Publico: srv.Client()})["shopify"](context.Background(), Origen{URL: srv.URL})
	if _, err := f.Pagina(context.Background(), "productos", nil); !errors.Is(err, ErrDefinitivo) {
		t.Fatalf("una tienda que no existe tiene que cortar la importación: %v", err)
	}
}

func TestFuenteWoo(t *testing.T) {
	var caso struct {
		Productos   json.RawMessage `json:"productos"`
		Variaciones json.RawMessage `json:"variaciones"`
	}
	if err := json.Unmarshal(leerCaso(t, "woocommerce.json"), &caso); err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		switch {
		case r.URL.Path != apiWoo:
			http.NotFound(w, r)
		case q.Get("type") == "variation":
			_, _ = w.Write(caso.Variaciones)
		case q.Get("page") == "1":
			w.Header().Set("X-WP-TotalPages", "1")
			_, _ = w.Write(caso.Productos)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	defer srv.Close()
	f, _ := Fuentes(Dependencias{Publico: srv.Client()})["woocommerce"](context.Background(), Origen{URL: srv.URL})
	p, err := f.Pagina(context.Background(), "productos", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Productos) == 0 || p.Siguiente != nil {
		t.Fatalf("página = %d productos, siguiente %s", len(p.Productos), p.Siguiente)
	}
	conOpciones := false
	for _, x := range p.Productos {
		conOpciones = conOpciones || len(x.Producto.Opciones) > 0
	}
	if !conOpciones {
		t.Fatal("el producto variable tendría que venir con sus opciones")
	}
}

func TestFuenteTiendanubeSinCredencial(t *testing.T) {
	if _, err := Fuentes(Dependencias{})["tiendanube"](context.Background(), Origen{}); !errors.Is(err, ErrDefinitivo) {
		t.Fatalf("sin credencial: %v", err)
	}
}

func TestFuenteMercadoLibreSinServicio(t *testing.T) {
	if _, err := Fuentes(Dependencias{})["mercadolibre"](context.Background(), Origen{}); !errors.Is(err, ErrDefinitivo) {
		t.Fatalf("sin Mercado Libre: %v", err)
	}
}

func TestOrigenPublico(t *testing.T) {
	for entrada, quiero := range map[string]string{
		"mitienda.com.ar":            "https://mitienda.com.ar",
		"http://Tienda.com/algo?x=1": "https://tienda.com",
		"localhost":                  "",
		"https://127.0.0.1":          "",
		"https://tienda.com:8080":    "",
		"intranet":                   "",
	} {
		o, err := importar.Origen(entrada)
		if (quiero == "" && err == nil) || o != quiero {
			t.Errorf("Origen(%q) = %q, %v; quiero %q", entrada, o, err, quiero)
		}
	}
}
