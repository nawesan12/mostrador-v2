package ia_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/ia"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/servidor"
)

func pedir(h http.Handler, metodo, ruta, cuerpo, token string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	return rec
}

func json1[T any](t *testing.T, rec *httptest.ResponseRecorder, estado int) T {
	t.Helper()
	var v T
	if rec.Code != estado {
		t.Fatalf("esperaba %d, vino %d: %s", estado, rec.Code, rec.Body)
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &v)
	return v
}

func registrar(t *testing.T, h http.Handler, email, tienda string) string {
	t.Helper()
	s := json1[cuentas.Sesion](t, pedir(h, "POST", "/v1/auth/registro",
		`{"nombre":"Lucía","tienda":"`+tienda+`","email":"`+email+`","clave":"Abcdefghijk1"}`, ""), 201)
	return s.Token
}

const verificador = "un-verificador-largo-de-prueba-con-mas-de-cuarenta-y-tres-caracteres"

func desafio() string {
	h := sha256.Sum256([]byte(verificador))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func TestIAConectada(t *testing.T) {
	b := pruebadb.Nueva(t)
	h := servidor.Nuevo(servidor.Dependencias{App: b.App, IA: &ia.Servicio{Pool: b.App},
		Cuentas: &cuentas.Servicio{Pool: b.App, Correo: &correo.Memoria{}, Sitio: "https://tulocalweb.com.ar"}})
	dueña := registrar(t, h, "lucia@laesquina.com.ar", "La Esquina")
	otra := registrar(t, h, "ana@otra.com.ar", "Otra Tienda")

	const vuelta = "https://claude.ai/api/mcp/auth_callback"
	cli := json1[ia.Cliente](t, pedir(h, "POST", "/v1/ia/clientes", `{"nombre":"Claude","redirecciones":["`+vuelta+`"]}`, ""), 201)
	if rec := pedir(h, "POST", "/v1/ia/clientes", `{"nombre":"X","redirecciones":["http://malo.com/cb"]}`, ""); rec.Code != 400 {
		t.Fatalf("redirección http a otro host: %d", rec.Code)
	}

	conectar := func(t *testing.T, acciones bool) ia.Tokens {
		t.Helper()
		cuerpo := `{"cliente":"` + cli.ID + `","redireccion":"` + vuelta + `","desafio":"` + desafio() + `","acciones":` + map[bool]string{true: "true", false: "false"}[acciones] + `}`
		c := json1[map[string]string](t, pedir(h, "POST", "/v1/ia/autorizar", cuerpo, dueña), 200)
		canje := `{"tipo":"authorization_code","codigo":"` + c["codigo"] + `","verificador":"` + verificador + `","cliente":"` + cli.ID + `","redireccion":"` + vuelta + `"}`
		tk := json1[ia.Tokens](t, pedir(h, "POST", "/v1/ia/token", canje, ""), 200)
		if rec := pedir(h, "POST", "/v1/ia/token", canje, ""); rec.Code != 400 || !strings.Contains(rec.Body.String(), "invalid_grant") {
			t.Fatalf("el código se usó dos veces: %d", rec.Code)
		}
		return tk
	}

	// Dos pedidos a la vez con el mismo código o el mismo refresco: pasa exactamente uno.
	aLaVez := func(cuerpo string) (ok int, primero ia.Tokens) {
		var mu sync.Mutex
		var wg sync.WaitGroup
		largada := make(chan struct{})
		for range 2 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-largada
				rec := pedir(h, "POST", "/v1/ia/token", cuerpo, "")
				mu.Lock()
				defer mu.Unlock()
				if rec.Code == 200 {
					ok++
					_ = json.Unmarshal(rec.Body.Bytes(), &primero)
				}
			}()
		}
		close(largada)
		wg.Wait()
		return ok, primero
	}
	t.Run("un solo uso aunque lleguen dos pedidos juntos", func(t *testing.T) {
		c := json1[map[string]string](t, pedir(h, "POST", "/v1/ia/autorizar",
			`{"cliente":"`+cli.ID+`","redireccion":"`+vuelta+`","desafio":"`+desafio()+`","acciones":false}`, dueña), 200)
		n, tk := aLaVez(`{"tipo":"authorization_code","codigo":"` + c["codigo"] + `","verificador":"` + verificador + `","cliente":"` + cli.ID + `","redireccion":"` + vuelta + `"}`)
		if n != 1 {
			t.Fatalf("canjes del mismo código que pasaron: %d", n)
		}
		if n, _ := aLaVez(`{"tipo":"refresh_token","refresco":"` + tk.Refresco + `","cliente":"` + cli.ID + `"}`); n != 1 {
			t.Fatalf("refrescos con el mismo token que pasaron: %d", n)
		}
	})

	t.Run("PKCE: con otro verificador no hay token", func(t *testing.T) {
		c := json1[map[string]string](t, pedir(h, "POST", "/v1/ia/autorizar",
			`{"cliente":"`+cli.ID+`","redireccion":"`+vuelta+`","desafio":"`+desafio()+`","acciones":false}`, dueña), 200)
		rec := pedir(h, "POST", "/v1/ia/token", `{"tipo":"authorization_code","codigo":"`+c["codigo"]+`","verificador":"otro-verificador-otro-verificador-otro-verificador","cliente":"`+cli.ID+`","redireccion":"`+vuelta+`"}`, "")
		if rec.Code != 400 {
			t.Fatalf("verificador incorrecto: %d", rec.Code)
		}
	})

	t.Run("lectura: ve pedidos pero no escribe ni entra a la cuenta", func(t *testing.T) {
		tk := conectar(t, false)
		if rec := pedir(h, "GET", "/v1/panel/pedidos", "", tk.Acceso); rec.Code != 200 {
			t.Fatalf("leer pedidos: %d %s", rec.Code, rec.Body)
		}
		for _, c := range []struct{ metodo, ruta string }{
			{"GET", "/v1/auth/sesion"}, {"GET", "/v1/panel/tienda"}, {"POST", "/v1/panel/cupones"},
			{"PUT", "/v1/panel/productos/x"}, {"GET", "/v1/panel/ia"}, {"POST", "/v1/ia/autorizar"},
		} {
			if rec := pedir(h, c.metodo, c.ruta, `{}`, tk.Acceso); rec.Code != 403 && rec.Code != 401 {
				t.Errorf("%s %s con token de lectura: %d", c.metodo, c.ruta, rec.Code)
			}
		}
		yo := json1[map[string]any](t, pedir(h, "GET", "/v1/ia/yo", "", tk.Acceso), 200)
		if yo["acciones"] != false {
			t.Fatalf("yo: %v", yo)
		}
	})

	t.Run("acciones: respetan los permisos del panel", func(t *testing.T) {
		tk := conectar(t, true)
		// "plata" viene apagado: crear cupones no pasa la guardia.
		if rec := pedir(h, "POST", "/v1/panel/cupones", `{}`, tk.Acceso); rec.Code != 403 {
			t.Fatalf("cupón sin permiso: %d", rec.Code)
		}
		if rec := pedir(h, "PUT", "/v1/panel/ia/permisos/plata", `{"valor":true}`, dueña); rec.Code != 204 {
			t.Fatalf("prender plata: %d %s", rec.Code, rec.Body)
		}
		if rec := pedir(h, "POST", "/v1/panel/cupones", `{}`, tk.Acceso); rec.Code == 403 {
			t.Fatalf("cupón con permiso: %d %s", rec.Code, rec.Body)
		}
		if rec := pedir(h, "PUT", "/v1/panel/ia/permisos/leer", `{"valor":false}`, dueña); rec.Code != 400 {
			t.Fatalf("apagar leer: %d", rec.Code)
		}
	})

	t.Run("refresco rota y el reuso corta la conexión", func(t *testing.T) {
		tk := conectar(t, false)
		ref := func(r string) *httptest.ResponseRecorder {
			return pedir(h, "POST", "/v1/ia/token", `{"tipo":"refresh_token","refresco":"`+r+`","cliente":"`+cli.ID+`"}`, "")
		}
		nuevo := json1[ia.Tokens](t, ref(tk.Refresco), 200)
		if rec := pedir(h, "GET", "/v1/panel/pedidos", "", tk.Acceso); rec.Code != 401 {
			t.Fatalf("el acceso viejo sigue: %d", rec.Code)
		}
		if rec := ref(tk.Refresco); rec.Code != 400 {
			t.Fatalf("reuso del refresco: %d", rec.Code)
		}
		if rec := pedir(h, "GET", "/v1/panel/pedidos", "", nuevo.Acceso); rec.Code != 401 {
			t.Fatalf("después del reuso la conexión tiene que estar cortada: %d", rec.Code)
		}
	})

	t.Run("el panel lista, otra tienda no desconecta, revocar corta", func(t *testing.T) {
		tk := conectar(t, false)
		p := json1[ia.Panel](t, pedir(h, "GET", "/v1/panel/ia", "", dueña), 200)
		if len(p.Conexiones) == 0 || p.Conexiones[0].Cliente != "Claude" || !p.Permisos["leer"] {
			t.Fatalf("panel: %+v", p)
		}
		if rec := pedir(h, "DELETE", "/v1/panel/ia/conexiones/"+p.Conexiones[0].ID, "", otra); rec.Code != 404 {
			t.Fatalf("otra tienda desconectó: %d", rec.Code)
		}
		if p2 := json1[ia.Panel](t, pedir(h, "GET", "/v1/panel/ia", "", otra), 200); len(p2.Conexiones) != 0 {
			t.Fatalf("otra tienda ve conexiones ajenas: %+v", p2.Conexiones)
		}
		if rec := pedir(h, "POST", "/v1/ia/revocar", `{"token":"`+tk.Acceso+`"}`, ""); rec.Code != 204 {
			t.Fatalf("revocar: %d", rec.Code)
		}
		if rec := pedir(h, "GET", "/v1/panel/pedidos", "", tk.Acceso); rec.Code != 401 {
			t.Fatalf("token revocado: %d", rec.Code)
		}
	})
}
