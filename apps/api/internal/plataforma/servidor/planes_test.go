package servidor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// Los topes de cada plan con la API entera (19/11/2026): se aplican en el servidor con 402
// tope_plan y el plan que lo habilita, y bajar de plan no borra nada (lo que sobra queda oculto en
// la tienda y en sólo lectura).
func TestTopesDePlan(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	cs := &cuentas.Servicio{Pool: b.App, Correo: &correo.Memoria{}, Sitio: "https://tulocalweb.test"}
	h := Nuevo(Dependencias{App: b.App, Cuentas: cs, Sitio: "https://tulocalweb.test", Log: silencio()})
	cl := cuentas.Cliente{IP: "190.210.32.1"}
	ana, err := cs.Registrar(ctx, cuentas.Alta{Nombre: "Ana", Tienda: "Sol", Email: "ana@sol.com", Clave: "clave-larga-1"}, cl)
	if err != nil {
		t.Fatal(err)
	}
	admin, err := pgx.Connect(ctx, b.AdminURL)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)
	var anaID, slug string
	_ = admin.QueryRow(ctx, `SELECT id::text FROM auth.usuarios WHERE email = 'ana@sol.com'`).Scan(&anaID)
	_ = admin.QueryRow(ctx, `SELECT slug FROM tiendas WHERE id = $1`, ana.TiendaID).Scan(&slug)
	// plan fija el plan; sin prueba, la prueba gratis ya terminó.
	plan := func(p string, prueba bool) {
		t.Helper()
		creada := "now() - interval '30 days'"
		if prueba {
			creada = "now()"
		}
		if _, err := admin.Exec(ctx, `UPDATE tiendas SET plan = $2, creada_en = `+creada+` WHERE id = $1`, ana.TiendaID, p); err != nil {
			t.Fatal(err)
		}
	}
	tokens := map[string]string{"ana": ana.Token}
	pedir := func(quien, metodo, ruta, cuerpo string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
		r.Header.Set("Authorization", "Bearer "+tokens[quien])
		r.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		return rec
	}
	type prob struct {
		Code, Plan, Funcion, Detail string
	}
	tope := func(t *testing.T, rec *httptest.ResponseRecorder, que, sube string) {
		t.Helper()
		var p prob
		_ = json.Unmarshal(rec.Body.Bytes(), &p)
		if rec.Code != http.StatusPaymentRequired || p.Code != "tope_plan" || p.Funcion != que || p.Plan != sube || p.Detail == "" {
			t.Fatalf("esperaba tope %s → %s, vino %d %s", que, sube, rec.Code, rec.Body)
		}
	}
	crear := func(n int, fotos int) *httptest.ResponseRecorder {
		f := make([]string, fotos)
		for i := range f {
			f[i] = fmt.Sprintf(`"/f/fotos/%s/%08d-0000-0000-0000-000000000000.webp"`, ana.TiendaID, i)
		}
		return pedir("ana", "POST", "/v1/panel/productos",
			fmt.Sprintf(`{"nombre":"Producto %d","precio":1000,"stock":1,"canales":{"web":true},"fotos":[%s]}`, n, strings.Join(f, ",")))
	}
	cuantos := func() int {
		var n int
		_ = admin.QueryRow(ctx, `SELECT count(*) FROM productos WHERE tienda_id = $1`, ana.TiendaID).Scan(&n)
		return n
	}
	visibles := func() int {

		tv, err := vidriera.Publica(ctx, b.App, slug)
		if err != nil {
			t.Fatal(err)
		}
		return len(tv.Productos)
	}
	invitar := func(email, rol string) error {
		_, err := configuracion.Invitar(ctx, b.App, cs.Sitio, ana.TiendaID, configuracion.Quien{UsuarioID: anaID, Rol: "duenio"}, email, rol)
		return err
	}
	entrar := func(email string) string {
		var raw []byte
		if err := b.Worker.QueryRow(ctx, `SELECT datos FROM outbox WHERE tipo = 'cuenta.invitacion' AND datos->>'para' = $1
			ORDER BY id DESC LIMIT 1`, email).Scan(&raw); err != nil {
			t.Fatal(err)
		}
		var m cuentas.MailInvitacion
		_ = json.Unmarshal(raw, &m)
		u, _ := url.Parse(m.Link)
		ses, err := cs.RegistrarConInvitacion(ctx, u.Query().Get("token"), "Persona", "clave-larga-2", cl)
		if err != nil {
			t.Fatal(err)
		}
		return ses.Token
	}

	t.Run("gratis: 20 productos, el 21 pide Emprende", func(t *testing.T) {
		plan("gratis", false)
		for i := 1; i <= 20; i++ {
			if rec := crear(i, 1); rec.Code != http.StatusCreated {
				t.Fatalf("producto %d: %d %s", i, rec.Code, rec.Body)
			}
		}
		tope(t, crear(21, 1), "productos", "emprende")
		if cuantos() != 20 {
			t.Fatalf("hay %d", cuantos())
		}
	})

	t.Run("gratis: 3 fotos por producto", func(t *testing.T) {
		var p struct{ ID string }
		_ = admin.QueryRow(ctx, `SELECT id::text FROM productos WHERE tienda_id = $1 ORDER BY creado_en LIMIT 1`, ana.TiendaID).Scan(&p.ID)
		f := func(n int) string {
			s := make([]string, n)
			for i := range s {
				s[i] = fmt.Sprintf(`"/f/fotos/%s/%08d-0000-0000-0000-000000000001.webp"`, ana.TiendaID, i)
			}
			return strings.Join(s, ",")
		}
		if rec := pedir("ana", "PUT", "/v1/panel/productos/"+p.ID, `{"nombre":"Producto 1","precio":1000,"canales":{"web":true},"fotos":[`+f(3)+`]}`); rec.Code != http.StatusOK {
			t.Fatalf("3 fotos: %d %s", rec.Code, rec.Body)
		}
		tope(t, pedir("ana", "PUT", "/v1/panel/productos/"+p.ID, `{"nombre":"Producto 1","precio":1000,"canales":{"web":true},"fotos":[`+f(4)+`]}`), "fotos", "emprende")
	})

	t.Run("gratis: funciones de otros planes", func(t *testing.T) {
		tope(t, pedir("ana", "POST", "/v1/panel/dominios", `{"dominio":"sol.com.ar"}`), "dominio", "emprende")
		tope(t, pedir("ana", "POST", "/v1/panel/cupones", `{}`), "cupones", "emprende")
		tope(t, pedir("ana", "GET", "/v1/panel/informes", ""), "newsletter", "pro")
		tope(t, pedir("ana", "PUT", "/v1/panel/configuracion/mayorista", `{}`), "mayorista", "negocio")
		tope(t, pedir("ana", "POST", "/v1/panel/integraciones/mercadolibre/conectar", `{}`), "mercadolibre", "pro")
		var p struct{ ID string }
		_ = admin.QueryRow(ctx, `SELECT id::text FROM productos WHERE tienda_id = $1 ORDER BY creado_en LIMIT 1`, ana.TiendaID).Scan(&p.ID)
		tope(t, pedir("ana", "POST", "/v1/panel/pedidos", `{"items":[{"productoId":"`+p.ID+`","cantidad":1}],"pago":"efectivo","canal":"pos"}`), "local", "pro")
		// Lo que no es de un plan sigue andando.
		if rec := pedir("ana", "GET", "/v1/panel/listas/productos", ""); rec.Code != http.StatusOK {
			t.Fatalf("lista: %d", rec.Code)
		}
	})

	t.Run("gratis: 1 persona; emprende: 2, sin roles; crece: con roles", func(t *testing.T) {
		if err := invitar("beto@sol.com", "admin"); err == nil || !strings.Contains(err.Error(), "personas") {
			t.Fatalf("gratis invitó: %v", err)
		}
		plan("emprende", false)
		if err := invitar("beto@sol.com", "ventas"); err == nil || !strings.Contains(err.Error(), "roles") {
			t.Fatalf("emprende con rol ventas: %v", err)
		}
		if err := invitar("beto@sol.com", "admin"); err != nil {
			t.Fatal(err)
		}
		if err := invitar("caro@sol.com", "admin"); err == nil || !strings.Contains(err.Error(), "personas") {
			t.Fatalf("emprende invitó a la tercera: %v", err)
		}
		tokens["beto"] = entrar("beto@sol.com")
		plan("pro", false)
		if err := invitar("caro@sol.com", "ventas"); err != nil {
			t.Fatal(err)
		}
		tokens["caro"] = entrar("caro@sol.com")
	})

	t.Run("la prueba gratis rige como Crece", func(t *testing.T) {
		plan("gratis", true)
		if rec := crear(21, 8); rec.Code != http.StatusCreated {
			t.Fatalf("en la prueba: %d %s", rec.Code, rec.Body)
		}
		if rec := pedir("ana", "GET", "/v1/panel/informes", ""); rec.Code != http.StatusOK {
			t.Fatalf("informes en la prueba: %d %s", rec.Code, rec.Body)
		}
	})

	t.Run("bajar a Gratis no borra nada: lo que sobra queda oculto y en sólo lectura", func(t *testing.T) {
		plan("negocio", false)
		for i := 22; i <= 25; i++ {
			if rec := crear(i, 1); rec.Code != http.StatusCreated {
				t.Fatalf("producto %d: %d %s", i, rec.Code, rec.Body)
			}
		}
		if visibles() != 25 {
			t.Fatalf("negocio ve %d", visibles())
		}
		plan("gratis", false)
		if cuantos() != 25 {
			t.Fatalf("se borraron productos: %d", cuantos())
		}
		if v := visibles(); v != 20 {
			t.Fatalf("la tienda muestra %d", v)
		}
		// La foto de más (el 21 tiene 8) queda guardada pero la tienda muestra 3.

		tv, _ := vidriera.Publica(ctx, b.App, slug)
		for _, p := range tv.Productos {
			if len(p.Fotos) > 3 {
				t.Fatalf("%s muestra %d fotos", p.Nombre, len(p.Fotos))
			}
		}
		var fotos []string
		if err := admin.QueryRow(ctx, `SELECT fotos FROM productos WHERE tienda_id = $1 AND nombre = 'Producto 21'`, ana.TiendaID).Scan(&fotos); err != nil {
			t.Fatal(err)
		}
		if len(fotos) != 8 {
			t.Fatalf("se perdieron fotos: %d", len(fotos))
		}
		// En el panel están todos, con los de más marcados.
		var l struct {
			Filas []map[string]any `json:"filas"`
		}
		_ = json.Unmarshal(pedir("ana", "GET", "/v1/panel/listas/productos?cuantos=200", "").Body.Bytes(), &l)
		fuera := []string{}
		for _, f := range l.Filas {
			if f["fueraDePlan"] == true {
				fuera = append(fuera, f["id"].(string))
			}
		}
		if len(l.Filas) != 25 || len(fuera) != 5 {
			t.Fatalf("panel: %d filas, %d fuera", len(l.Filas), len(fuera))
		}
		// Uno de más no se edita; se puede borrar y libera lugar.
		rec := pedir("ana", "PUT", "/v1/panel/productos/"+fuera[0], `{"nombre":"Cambio","precio":1,"canales":{"web":true}}`)
		tope(t, rec, "productos", "emprende")
		var p prob
		_ = json.Unmarshal(rec.Body.Bytes(), &p)
		if !strings.Contains(p.Detail, "quedó fuera") {
			t.Fatalf("detalle: %s", p.Detail)
		}
		if rec := pedir("ana", "DELETE", "/v1/panel/productos/"+fuera[0], ""); rec.Code != http.StatusNoContent {
			t.Fatalf("borrar: %d %s", rec.Code, rec.Body)
		}
		// El equipo de más (Beto y Caro, con la dueña son 3) queda en sólo lectura: mira pero no cambia.
		if rec := pedir("beto", "GET", "/v1/panel/listas/productos", ""); rec.Code != http.StatusOK {
			t.Fatalf("beto no ve: %d", rec.Code)
		}
		tope(t, pedir("beto", "POST", "/v1/panel/productos", `{"nombre":"Otro","precio":1,"canales":{}}`), "personas", "emprende")
		// La dueña sigue pudiendo todo lo de su plan.
		if rec := pedir("ana", "PUT", "/v1/panel/productos/"+fuera[1], `{"nombre":"Cambio","precio":1,"canales":{"web":true}}`); rec.Code != http.StatusPaymentRequired {
			t.Fatalf("otro de más: %d", rec.Code)
		}
		// Vuelve a subir: todo está y se ve.
		plan("negocio", false)
		if v := visibles(); v != 24 {
			t.Fatalf("de vuelta en negocio muestra %d", v)
		}
		if rec := pedir("beto", "POST", "/v1/panel/productos", `{"nombre":"Otro","precio":1,"canales":{}}`); rec.Code != http.StatusCreated {
			t.Fatalf("beto de vuelta: %d %s", rec.Code, rec.Body)
		}
	})

	t.Run("la pantalla del plan trae el catálogo y lo que quedó fuera", func(t *testing.T) {
		plan("gratis", false)
		var v struct {
			Vigente string
			Planes  []struct {
				ID        string
				Productos int
				Mensual   int64
			}
			Fuera map[string]int
			Pack  map[string]int
		}
		rec := pedir("ana", "GET", "/v1/panel/cuenta/plan", "")
		if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
			t.Fatal(err)
		}
		if v.Vigente != "gratis" || len(v.Planes) != 4 || v.Planes[2].Mensual != 34_900 || v.Fuera["productos"] != 5 ||
			v.Fuera["personas"] != 2 || v.Pack["precio"] != 4_900 {
			t.Fatalf("plan: %s", rec.Body)
		}
	})
}
