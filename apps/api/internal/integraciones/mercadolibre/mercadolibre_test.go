package mercadolibre_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadolibre"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

type mlFalso struct {
	mu      sync.Mutex
	pedidos []string
	cuerpos map[string]map[string]any
	ordenes map[string]map[string]any
}

func (f *mlFalso) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pedidos = append(f.pedidos, r.Method+" "+r.URL.Path)
	b, _ := io.ReadAll(r.Body)
	var c map[string]any
	_ = json.Unmarshal(b, &c)
	f.cuerpos[r.Method+" "+r.URL.Path] = c
	enviar := func(v any) { _ = json.NewEncoder(w).Encode(v) }
	switch {
	case r.URL.Path == "/oauth/token":
		v, _ := url.ParseQuery(string(b))
		if v.Get("grant_type") == "authorization_code" && len(v.Get("code_verifier")) < 43 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		enviar(map[string]any{"access_token": "APP_USR-ml-vendedora", "refresh_token": "TG-renovar", "expires_in": 21600, "user_id": 555})
	case r.URL.Path == "/users/me":
		enviar(map[string]any{"id": 555, "nickname": "LUNA_TIENDA"})
	case r.URL.Path == "/sites/MLA/domain_discovery/search":
		enviar([]map[string]any{{"category_id": "MLA1234", "attributes": []map[string]any{{"id": "BRAND", "value_name": "Stanley"}}}})
	case r.URL.Path == "/items" && r.Method == http.MethodPost:
		enviar(map[string]any{"id": "MLA900", "permalink": "https://articulo.mercadolibre.com.ar/MLA-900", "status": "active"})
	case strings.HasPrefix(r.URL.Path, "/items/"):
		if r.Method == http.MethodGet {
			enviar(map[string]any{"id": "MLA900", "status": "paused"})
			return
		}
		enviar(map[string]any{})
	case strings.HasPrefix(r.URL.Path, "/orders/"):
		enviar(f.ordenes[strings.TrimPrefix(r.URL.Path, "/orders/")])
	case strings.HasPrefix(r.URL.Path, "/questions/"):
		enviar(map[string]any{"id": 77, "text": "¿Tenés en negro?", "status": "UNANSWERED", "item_id": "MLA900", "date_created": "2026-10-23T10:00:00.000-04:00"})
	case r.URL.Path == "/answers":
		enviar(map[string]any{"id": 77})
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (f *mlFalso) cuantos(pref string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, p := range f.pedidos {
		if strings.HasPrefix(p, pref) {
			n++
		}
	}
	return n
}

func (f *mlFalso) cuerpo(k string) map[string]any {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.cuerpos[k]
}

func TestMercadoLibreDePuntaAPunta(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	f := &mlFalso{cuerpos: map[string]map[string]any{}, ordenes: map[string]map[string]any{}}
	srv := httptest.NewServer(f)
	defer srv.Close()
	k := make([]byte, 32)
	_, _ = rand.Read(k)
	cif, _ := cifrado.Nuevo(base64.StdEncoding.EncodeToString(k))
	s := &mercadolibre.Servicio{Pool: b.App, Cifra: cif, API: srv.URL, Auth: srv.URL,
		Config: mercadolibre.Config{AppID: "810", Secreto: "sec", Sitio: "https://mostrador.iamnaw.dev"}}
	w := &outbox.Worker{Pool: b.Worker, Manejadores: s.Manejadores()}
	ana := "00000000-0000-0000-0000-0000000000aa"

	// Conectar con PKCE; el state es de un solo uso y de la tienda.
	u, err := s.Conectar(ctx, luna, ana)
	if err != nil {
		t.Fatal(err)
	}
	q, _ := url.Parse(u)
	if q.Query().Get("code_challenge_method") != "S256" || !strings.HasSuffix(q.Query().Get("redirect_uri"), "/integraciones/mercadolibre/callback") {
		t.Fatalf("url: %s", u)
	}
	if err := s.Volver(ctx, sol, ana, "c", q.Query().Get("state")); !errors.Is(err, mercadolibre.ErrEstado) {
		t.Fatal("otra tienda usó el state")
	}
	if err := s.Volver(ctx, luna, ana, "c", q.Query().Get("state")); err != nil {
		t.Fatal(err)
	}
	if e, _ := s.EstadoDe(ctx, luna); !e.Conectada || e.Apodo != "LUNA_TIENDA" {
		t.Fatalf("estado: %+v", e)
	}

	// Publicar: categoría sugerida, precio con el 10 % de recargo, stock y descripción.
	stock := 5
	desc := "Acero inoxidable"
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Termo Stanley 1 L", Descripcion: &desc, Precio: 40000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	pub, err := s.Publicar(ctx, luna, p.ID)
	if err != nil || pub.ItemID != "MLA900" {
		t.Fatalf("publicar: %+v %v", pub, err)
	}
	item := f.cuerpo("POST /items")
	if item["category_id"] != "MLA1234" || item["price"].(float64) != 44000 || item["available_quantity"].(float64) != 5 {
		t.Fatalf("item: %v", item)
	}
	if f.cuantos("POST /items/MLA900/description") != 1 {
		t.Fatal("no cargó la descripción")
	}
	if _, err := s.Publicar(ctx, luna, p.ID); !errors.Is(err, mercadolibre.ErrYaPublicado) {
		t.Fatal("publicó dos veces")
	}

	// Un cambio de stock en Mostrador se encola solo (trigger) y el worker lo manda.
	cero := 0
	if _, err := listas.Guardar(ctx, b.App, luna, p.ID, listas.Editable{Nombre: p.Nombre, Precio: 40000, Stock: &cero}); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Pasada(ctx); err != nil {
		t.Fatal(err)
	}
	if put := f.cuerpo("PUT /items/MLA900"); put == nil || put["status"] != "paused" {
		t.Fatalf("sin stock no se pausó: %v", put)
	}
	stock = 3
	_, _ = listas.Guardar(ctx, b.App, luna, p.ID, listas.Editable{Nombre: p.Nombre, Precio: 40000, Stock: &stock})
	_, _ = w.Pasada(ctx)
	if put := f.cuerpo("PUT /items/MLA900"); put["status"] != "active" || put["available_quantity"].(float64) != 3 {
		t.Fatalf("con stock no se reactivó: %v", put)
	}

	// Una venta: dos avisos de la misma orden crean un solo pedido, con el precio de Mercado Libre.
	f.ordenes["2000"] = map[string]any{"id": 2000, "status": "paid",
		"order_items": []map[string]any{{"item": map[string]any{"id": "MLA900"}, "quantity": 1, "unit_price": 44000}},
		"buyer":       map[string]any{"nickname": "COMPRADOR1", "first_name": "Lucía", "last_name": "Pérez"}}
	aviso := func(topic, recurso, app string) int {
		cuerpo := `{"resource":"` + recurso + `","user_id":555,"topic":"` + topic + `","application_id":` + app + `}`
		r := httptest.NewRequest(http.MethodPost, "/v1/webhooks/mercadolibre", strings.NewReader(cuerpo))
		rec := httptest.NewRecorder()
		mux := http.NewServeMux()
		s.Rutas(mux)
		mux.ServeHTTP(rec, r)
		return rec.Code
	}
	if c := aviso("orders_v2", "/orders/2000", "810"); c != 200 {
		t.Fatalf("aviso: %d", c)
	}
	aviso("orders_v2", "/orders/2000", "810")
	aviso("orders_v2", "/orders/2000", "999") // otra aplicación: se ignora
	for range 3 {
		_, _ = w.Pasada(ctx)
	}
	var n int
	var total int64
	var canal, envio string
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT count(*), max(total), max(canal), max(envio) FROM pedidos WHERE canal = 'ml'`).Scan(&n, &total, &canal, &envio)
	})
	if n != 1 || total != 44000 || envio != "mercadoenvios" {
		t.Fatalf("pedidos de ML: %d total %d envío %s", n, total, envio)
	}
	var quedan int
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT stock FROM productos WHERE id::text = $1`, p.ID).Scan(&quedan)
	})
	if quedan != 2 {
		t.Fatalf("stock después de la venta: %d", quedan)
	}

	// Preguntas: llegan por aviso y se responden.
	aviso("questions", "/questions/77", "810")
	_, _ = w.Pasada(ctx)
	ps, _ := s.Preguntas(ctx, luna)
	if len(ps) != 1 || ps[0].Producto != "Termo Stanley 1 L" || ps[0].Estado != "UNANSWERED" {
		t.Fatalf("preguntas: %+v", ps)
	}
	if err := s.Responder(ctx, luna, 77, "¡Sí! Tenemos en negro y en verde."); err != nil {
		t.Fatal(err)
	}
	if e, _ := s.EstadoDe(ctx, luna); e.Preguntas != 0 {
		t.Fatal("la pregunta sigue sin responder")
	}
	// La otra tienda no ve nada.
	if ps, _ := s.Preguntas(ctx, sol); len(ps) != 0 {
		t.Fatal("una tienda vio las preguntas de otra")
	}

	// Un recurso que no es de la lista no se procesa.
	if c := aviso("orders_v2", "/users/1", "810"); c != 200 {
		t.Fatal(c)
	}
	_ = time.Second
}
