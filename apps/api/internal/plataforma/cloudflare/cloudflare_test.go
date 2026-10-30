package cloudflare

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// Un Cloudflare de mentira con zonas, registros y routes en memoria.
type falso struct {
	mu        sync.Mutex
	zonas     map[string]Zona
	registros []registro
	routes    []route
	pedidos   []string
}

func (f *falso) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pedidos = append(f.pedidos, r.Method+" "+r.URL.Path)
	ok := func(v any) {
		b, _ := json.Marshal(v)
		_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "result": json.RawMessage(b)})
	}
	if r.Header.Get("Authorization") != "Bearer tok" {
		w.WriteHeader(403)
		_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "errors": []map[string]any{{"code": 10000, "message": "auth"}}})
		return
	}
	p := r.URL.Path
	switch {
	case r.Method == "GET" && p == "/zones":
		if z, ok2 := f.zonas[r.URL.Query().Get("name")]; ok2 && r.URL.Query().Get("account.id") == "cta" {
			ok([]Zona{z})
		} else {
			ok([]Zona{})
		}
	case r.Method == "POST" && p == "/zones":
		var b struct{ Name string }
		_ = json.NewDecoder(r.Body).Decode(&b)
		if _, ya := f.zonas[b.Name]; ya {
			_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "errors": []map[string]any{{"code": 1061, "message": "ya existe"}}})
			return
		}
		z := Zona{ID: "z-" + b.Name, Nombre: b.Name, Estado: "pending", Nameservers: []string{"ana.ns.cloudflare.com", "beto.ns.cloudflare.com"}}
		f.zonas[b.Name] = z
		ok(z)
	case strings.HasSuffix(p, "/dns_records") && r.Method == "GET":
		var out []registro
		for _, x := range f.registros {
			if x.Nombre == r.URL.Query().Get("name") {
				out = append(out, x)
			}
		}
		ok(out)
	case strings.HasSuffix(p, "/dns_records") && r.Method == "POST":
		var x registro
		_ = json.NewDecoder(r.Body).Decode(&x)
		x.ID = "r" + x.Nombre
		f.registros = append(f.registros, x)
		ok(x)
	case strings.Contains(p, "/dns_records/") && r.Method == "DELETE":
		id := p[strings.LastIndex(p, "/")+1:]
		for i, x := range f.registros {
			if x.ID == id {
				f.registros = append(f.registros[:i], f.registros[i+1:]...)
				break
			}
		}
		ok(map[string]string{"id": id})
	case strings.HasSuffix(p, "/workers/routes") && r.Method == "GET":
		ok(f.routes)
	case strings.HasSuffix(p, "/workers/routes") && r.Method == "POST":
		var x route
		_ = json.NewDecoder(r.Body).Decode(&x)
		x.ID = "w" + strings.NewReplacer("/", "_", "*", "x").Replace(x.Patron)
		f.routes = append(f.routes, x)
		ok(x)
	case strings.Contains(p, "/workers/routes/") && r.Method == "DELETE":
		id := p[strings.LastIndex(p, "/")+1:]
		for i, x := range f.routes {
			if x.ID == id {
				f.routes = append(f.routes[:i], f.routes[i+1:]...)
				break
			}
		}
		ok(map[string]string{"id": id})
	default:
		w.WriteHeader(404)
		_ = json.NewEncoder(w).Encode(map[string]any{"success": false, "errors": []map[string]any{{"code": 7003, "message": "no"}}})
	}
}

func TestZonaRegistrosYRoutes(t *testing.T) {
	ctx := context.Background()
	f := &falso{zonas: map[string]Zona{"tengolibre.com.ar": {ID: "z1", Nombre: "tengolibre.com.ar", Estado: "active"}}}
	srv := httptest.NewServer(f)
	defer srv.Close()
	c := Nuevo("tok", "cta")
	c.Base = srv.URL

	z, err := c.CrearZona(ctx, "lunamates.com.ar")
	if err != nil || z.Estado != "pending" || len(z.Nameservers) != 2 {
		t.Fatalf("zona nueva: %+v %v", z, err)
	}
	// Ya agregada: la misma, sin error.
	if z2, err := c.CrearZona(ctx, "lunamates.com.ar"); err != nil || z2.ID != z.ID {
		t.Fatalf("repetida: %+v %v", z2, err)
	}
	// Un subdominio cae en la zona del padre.
	if z, err := c.ZonaDe(ctx, "prueba.tengolibre.com.ar"); err != nil || z == nil || z.ID != "z1" {
		t.Fatalf("zona del padre: %+v %v", z, err)
	}
	if z, err := c.ZonaDe(ctx, "otro.com.ar"); err != nil || z != nil {
		t.Fatalf("sin zona: %+v %v", z, err)
	}

	for range 2 { // dos veces: no duplica nada
		if err := c.Apuntar(ctx, "z1", "prueba.tengolibre.com.ar", "mostrador"); err != nil {
			t.Fatal(err)
		}
	}
	if len(f.registros) != 1 || f.registros[0].Valor != "100::" || !f.registros[0].Proxied || len(f.routes) != 1 ||
		f.routes[0].Patron != "prueba.tengolibre.com.ar/*" || f.routes[0].Script != "mostrador" {
		t.Fatalf("apuntar: %+v %+v", f.registros, f.routes)
	}
	if err := c.Desapuntar(ctx, "z1", "prueba.tengolibre.com.ar"); err != nil || len(f.registros) != 0 || len(f.routes) != 0 {
		t.Fatalf("desapuntar: %v %+v %+v", err, f.registros, f.routes)
	}

	c.Token = "otro"
	if _, err := c.ZonaDe(ctx, "x.com.ar"); err == nil {
		t.Fatal("sin token válido tiene que fallar")
	}
}
