package tiendanube

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// conServidor apunta el paquete a un servidor falso mientras dura el test.
func conServidor(t *testing.T, h http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(h)
	viejos := [3]string{base, autorizar, urlToken}
	base, urlToken = srv.URL+"/"+version, srv.URL+"/token"
	t.Cleanup(func() {
		srv.Close()
		base, autorizar, urlToken = viejos[0], viejos[1], viejos[2]
	})
	return srv
}

func TestCanjear(t *testing.T) {
	conServidor(t, func(w http.ResponseWriter, r *http.Request) {
		var b map[string]string
		_ = json.NewDecoder(r.Body).Decode(&b)
		if r.URL.Path != "/token" || b["code"] != "abc" || b["grant_type"] != "authorization_code" || b["client_secret"] != "s" {
			http.Error(w, `{"error":"invalid"}`, http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(`{"access_token":"tok","token_type":"bearer","scope":"read_products,read_orders","user_id":789}`))
	})
	tok, err := App{ID: "49256", Secreto: "s"}.Canjear(context.Background(), "abc")
	if err != nil || tok.Acceso != "tok" || tok.TiendaID != "789" || len(tok.Alcance) != 2 {
		t.Fatalf("Canjear = %+v, %v", tok, err)
	}
	if _, err := (App{ID: "49256", Secreto: "s"}).Canjear(context.Background(), "otro"); err == nil {
		t.Fatal("un código inválido tendría que fallar")
	}
	if _, err := (App{}).Canjear(context.Background(), "abc"); !errors.Is(err, ErrApagado) {
		t.Fatalf("sin credenciales: %v", err)
	}
}

func TestURLAutorizar(t *testing.T) {
	u, err := App{ID: "49256", Secreto: "s"}.URLAutorizar("a b")
	if err != nil || u != "https://www.tiendanube.com/apps/49256/authorize?state=a+b" {
		t.Fatalf("URLAutorizar = %q, %v", u, err)
	}
}

func TestLeerPaginaYReintento(t *testing.T) {
	pedidos := 0
	var srv *httptest.Server
	srv = conServidor(t, func(w http.ResponseWriter, r *http.Request) {
		pedidos++
		if r.Header.Get("Authorization") != "Bearer tok" || !strings.HasPrefix(r.Header.Get("User-Agent"), "Tu Local Web") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if pedidos == 1 {
			w.Header().Set("X-Rate-Limit-Reset", "1500")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("X-Total-Count", "201")
		w.Header().Set("X-Main-Language", "es")
		w.Header().Set("Link", `<`+srv.URL+`/`+version+`/789/products?page=2&per_page=200>; rel="next", <https://otro.com/x>; rel="last"`)
		_, _ = w.Write([]byte(`[{"id":1},{"id":2}]`))
	})
	var esperas []time.Duration
	c := API{Token: "tok", TiendaID: "789", Dormir: func(_ context.Context, d time.Duration) error {
		esperas = append(esperas, d)
		return nil
	}}
	p, err := c.Leer(context.Background(), c.Primera("products", nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Registros) != 2 || p.Total != 201 || p.Idioma != "es" || !strings.HasSuffix(p.Siguiente, "/789/products?page=2&per_page=200") {
		t.Fatalf("página = %+v", p)
	}
	if len(esperas) != 1 || esperas[0] != 1500*time.Millisecond {
		t.Fatalf("esperas = %v", esperas)
	}
	if _, err := c.Leer(context.Background(), "https://otro.com/789/products"); err == nil {
		t.Fatal("no tiene que seguir URLs fuera de la API")
	}
	if _, err := (API{Token: "mal", TiendaID: "789"}).Leer(context.Background(), c.Primera("products", nil)); !errors.Is(err, ErrReautorizar) {
		t.Fatalf("token inválido: %v", err)
	}
}
