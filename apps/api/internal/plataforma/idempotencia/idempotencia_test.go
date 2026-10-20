package idempotencia_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/idempotencia"
)

func TestIdempotencia(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")

	var veces atomic.Int32
	var falla atomic.Bool
	h := idempotencia.Middleware(b.App, func(r *http.Request) string { return r.Header.Get("X-Tienda") })(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			n := veces.Add(1)
			if falla.Load() {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"pedido":` + string(rune('0'+n)) + `}`))
		}))
	pedir := func(tienda, clave, cuerpo string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/v1/pedidos", strings.NewReader(cuerpo))
		r.Header.Set("X-Tienda", tienda)
		if clave != "" {
			r.Header.Set(idempotencia.Encabezado, clave)
		}
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		return rec
	}

	primera := pedir(luna, "clave-0001", `{"x":1}`)
	if primera.Code != 201 || primera.Body.String() != `{"pedido":1}` {
		t.Fatalf("primera: %d %s", primera.Code, primera.Body)
	}

	t.Run("la repetición devuelve lo guardado sin ejecutar", func(t *testing.T) {
		otra := pedir(luna, "clave-0001", `{"x":1}`)
		if otra.Code != 201 || otra.Body.String() != `{"pedido":1}` || otra.Header().Get(idempotencia.Repetida) != "true" {
			t.Fatalf("repetida: %d %s %v", otra.Code, otra.Body, otra.Header())
		}
		if otra.Header().Get("Content-Type") != "application/json" {
			t.Fatalf("Content-Type perdido: %q", otra.Header().Get("Content-Type"))
		}
		if veces.Load() != 1 {
			t.Fatalf("el handler corrió %d veces", veces.Load())
		}
	})

	t.Run("misma clave con otro cuerpo es un error", func(t *testing.T) {
		if rec := pedir(luna, "clave-0001", `{"x":2}`); rec.Code != 422 {
			t.Fatalf("estado %d", rec.Code)
		}
	})

	t.Run("la misma clave en otra tienda es otro pedido", func(t *testing.T) {
		if rec := pedir(sol, "clave-0001", `{"x":1}`); rec.Code != 201 || rec.Header().Get(idempotencia.Repetida) != "" {
			t.Fatalf("otra tienda: %d %v", rec.Code, rec.Header())
		}
	})

	t.Run("sin clave no hay idempotencia", func(t *testing.T) {
		antes := veces.Load()
		pedir(luna, "", `{}`)
		pedir(luna, "", `{}`)
		if veces.Load() != antes+2 {
			t.Fatal("sin clave tendría que ejecutar siempre")
		}
	})

	t.Run("un 5xx libera la clave para reintentar", func(t *testing.T) {
		falla.Store(true)
		if rec := pedir(luna, "clave-0002", `{}`); rec.Code != 500 {
			t.Fatalf("estado %d", rec.Code)
		}
		falla.Store(false)
		if rec := pedir(luna, "clave-0002", `{}`); rec.Code != 201 || rec.Header().Get(idempotencia.Repetida) != "" {
			t.Fatalf("reintento: %d %v", rec.Code, rec.Header())
		}
	})

	t.Run("clave muy corta", func(t *testing.T) {
		if rec := pedir(luna, "corta", `{}`); rec.Code != 400 {
			t.Fatalf("estado %d", rec.Code)
		}
	})
}

func TestPedidoIgualEnCurso(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	adentro := make(chan struct{})
	soltar := make(chan struct{})
	h := idempotencia.Middleware(b.App, func(*http.Request) string { return luna })(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			close(adentro)
			<-soltar
			w.WriteHeader(http.StatusCreated)
		}))
	pedir := func() *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/v1/cobros", strings.NewReader(`{}`))
		r.Header.Set(idempotencia.Encabezado, "clave-cobro-1")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		return rec
	}
	primera := make(chan *httptest.ResponseRecorder)
	go func() { primera <- pedir() }()
	<-adentro
	if rec := pedir(); rec.Code != http.StatusConflict {
		t.Fatalf("mientras corre la primera: %d, quería 409", rec.Code)
	}
	close(soltar)
	if rec := <-primera; rec.Code != http.StatusCreated {
		t.Fatalf("primera: %d", rec.Code)
	}
}
