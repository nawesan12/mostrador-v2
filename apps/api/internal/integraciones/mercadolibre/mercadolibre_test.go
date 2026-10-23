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
	// sinVariaciones: la categoría sugerida no tiene atributos que admitan variaciones.
	sinVariaciones bool
	mu             sync.Mutex
	pedidos        []string
	cuerpos        map[string]map[string]any
	ordenes        map[string]map[string]any
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
	case strings.HasPrefix(r.URL.Path, "/categories/") && strings.HasSuffix(r.URL.Path, "/attributes"):
		if f.sinVariaciones {
			enviar([]map[string]any{{"id": "BRAND", "name": "Marca", "tags": map[string]any{}}})
			return
		}
		enviar([]map[string]any{
			{"id": "BRAND", "name": "Marca", "tags": map[string]any{}},
			{"id": "SIZE", "name": "Talle", "tags": map[string]any{"allow_variations": true}},
			{"id": "COLOR", "name": "Color", "tags": map[string]any{"allow_variations": true},
				"values": []map[string]any{{"id": "52049", "name": "Negro"}}},
		})
	case r.URL.Path == "/items" && r.Method == http.MethodPost:
		res := map[string]any{"id": "MLA900", "permalink": "https://articulo.mercadolibre.com.ar/MLA-900", "status": "active"}
		if vs, ok := c["variations"].([]any); ok {
			out := []map[string]any{}
			for i := range vs {
				out = append(out, map[string]any{"id": 7000 + i})
			}
			res["variations"] = out
		}
		enviar(res)
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

func TestMercadoLibreConVariantes(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
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
	u, _ := s.Conectar(ctx, luna, ana)
	q, _ := url.Parse(u)
	if err := s.Volver(ctx, luna, ana, "c", q.Query().Get("state")); err != nil {
		t.Fatal(err)
	}
	n := func(v int) *int { return &v }
	opciones := []listas.Opcion{{Nombre: "Talle", Valores: []string{"M", "L"}}, {Nombre: "Color", Valores: []string{"Negro"}},
		{Nombre: "Corte", Valores: []string{"Recto"}}}
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Remera lisa", Precio: 20000,
		Fotos: []string{"https://f/a.jpg", "https://f/b.jpg"}, Opciones: opciones,
		Variantes: []listas.VarianteEditable{
			{Valores: map[string]string{"Talle": "M", "Color": "Negro", "Corte": "Recto"}, Stock: n(3), Foto: ptr("https://f/b.jpg"), SKU: ptr("REM-M")},
			{Valores: map[string]string{"Talle": "L", "Color": "Negro", "Corte": "Recto"}, Stock: n(2)},
		}})
	if err != nil {
		t.Fatal(err)
	}
	l := p.Variantes[1]

	// Precios distintos entre variantes: Mercado Libre no lo admite, no se publica.
	if _, err := listas.Guardar(ctx, b.App, luna, p.ID, listas.Editable{Nombre: p.Nombre, Precio: 20000, Fotos: p.Fotos,
		Variantes: []listas.VarianteEditable{{Valores: l.Valores, Precio: ptr(int64(25000))}}}); err != nil {
		t.Fatal(err)
	}
	var ev *mercadolibre.ErrVariaciones
	if _, err := s.Publicar(ctx, luna, p.ID); !errors.As(err, &ev) || !strings.Contains(ev.Mensaje, "mismo precio") {
		t.Fatalf("precios distintos: %v", err)
	}
	_, _ = listas.Guardar(ctx, b.App, luna, p.ID, listas.Editable{Nombre: p.Nombre, Precio: 20000, Fotos: p.Fotos,
		Variantes: []listas.VarianteEditable{{Valores: l.Valores}}})

	// Una categoría sin atributos de variación: aviso claro, nada publicado.
	f.sinVariaciones = true
	if _, err := s.Publicar(ctx, luna, p.ID); !errors.As(err, &ev) || !strings.Contains(ev.Mensaje, "no admite variantes") {
		t.Fatalf("categoría sin variaciones: %v", err)
	}
	if f.cuantos("POST /items") != 0 {
		t.Fatal("publicó igual")
	}
	f.sinVariaciones = false

	// Publicar: una variación por variante, Talle y Color mapeados a SIZE y COLOR (con value_id),
	// Corte como atributo personalizado, stock y foto por variación, galería completa.
	if _, err := s.Publicar(ctx, luna, p.ID); err != nil {
		t.Fatal(err)
	}
	item := f.cuerpo("POST /items")
	vars := item["variations"].([]any)
	if len(vars) != 2 || item["available_quantity"] != nil || len(item["pictures"].([]any)) != 2 || item["price"].(float64) != 22000 {
		t.Fatalf("item: %v", item)
	}
	v0 := vars[0].(map[string]any)
	combo := v0["attribute_combinations"].([]any)
	talle, color, corte := combo[0].(map[string]any), combo[1].(map[string]any), combo[2].(map[string]any)
	if talle["id"] != "SIZE" || talle["value_name"] != "M" || color["id"] != "COLOR" || color["value_id"] != "52049" ||
		corte["name"] != "Corte" || corte["id"] != nil || v0["available_quantity"].(float64) != 3 ||
		v0["seller_custom_field"] != "REM-M" || v0["picture_ids"].([]any)[0] != "https://f/b.jpg" {
		t.Fatalf("variación: %v", v0)
	}

	// Una venta de la variación de L entra como esa variante, con su texto, y descuenta su stock.
	f.ordenes["3000"] = map[string]any{"id": 3000, "status": "paid",
		"order_items": []map[string]any{{"item": map[string]any{"id": "MLA900", "variation_id": 7001}, "quantity": 1, "unit_price": 22000}},
		"buyer":       map[string]any{"nickname": "COMPRADOR2"}}
	cuerpo := `{"resource":"/orders/3000","user_id":555,"topic":"orders_v2","application_id":810}`
	mux := http.NewServeMux()
	s.Rutas(mux)
	mux.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/v1/webhooks/mercadolibre", strings.NewReader(cuerpo)))
	for range 3 {
		_, _ = w.Pasada(ctx)
	}
	var texto string
	var total int64
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT items->0->>'variante', total FROM pedidos WHERE canal = 'ml'`).Scan(&texto, &total)
	})
	if texto != "Talle L · Color Negro · Corte Recto" || total != 22000 {
		t.Fatalf("pedido de ML: %q %d", texto, total)
	}
	p, _ = listas.Obtener(ctx, b.App, luna, p.ID, time.Now())
	if p.Variantes[1].Stock != 1 || p.Stock != 4 {
		t.Fatalf("stock: %+v", p.Variantes)
	}

	// El cambio de stock se sincroniza por variación.
	_, _ = w.Pasada(ctx)
	put := f.cuerpo("PUT /items/MLA900")
	pv, _ := put["variations"].([]any)
	if len(pv) != 2 || pv[0].(map[string]any)["id"].(float64) != 7000 || pv[1].(map[string]any)["available_quantity"].(float64) != 1 || put["available_quantity"] != nil {
		t.Fatalf("sincronizar: %v", put)
	}
}

func ptr[T any](v T) *T { return &v }
