package servidor

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// Los permisos por rol con la API entera (sesiones de verdad, las rutas reales): lo sensible
// responde 403 a quien no le toca, aunque la web no lo muestre (docs/decisiones.md §37).
func TestPermisosPorRol(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	cs := &cuentas.Servicio{Pool: b.App, Correo: &correo.Memoria{}, Sitio: "https://tulocalweb.test"}
	h := Nuevo(Dependencias{App: b.App, Cuentas: cs, Sitio: "https://tulocalweb.test", Log: silencio()})
	cl := cuentas.Cliente{IP: "190.210.32.1"}
	ana, err := cs.Registrar(ctx, cuentas.Alta{Nombre: "Ana", Tienda: "Sol", Email: "ana@sol.com", Clave: "clave-larga-1"}, cl)
	if err != nil {
		t.Fatal(err)
	}
	var anaID string
	_ = b.App.QueryRow(ctx, `SELECT id::text FROM auth.usuarios WHERE email = 'ana@sol.com'`).Scan(&anaID)
	tokens := map[string]string{"duenio": ana.Token}
	for _, rol := range []string{"admin", "ventas", "local"} {
		email := rol + "@sol.com"
		if _, err := configuracion.Invitar(ctx, b.App, cs.Sitio, ana.TiendaID, configuracion.Quien{UsuarioID: anaID, Rol: "duenio"}, email, rol); err != nil {
			t.Fatal(err)
		}
		var raw []byte
		if err := b.Worker.QueryRow(ctx, `SELECT datos FROM outbox WHERE tipo = 'cuenta.invitacion' AND datos->>'para' = $1`, email).Scan(&raw); err != nil {
			t.Fatal(err)
		}
		var m cuentas.MailInvitacion
		_ = json.Unmarshal(raw, &m)
		u, _ := url.Parse(m.Link)
		ses, err := cs.RegistrarConInvitacion(ctx, u.Query().Get("token"), "Persona "+rol, "clave-larga-2", cl)
		if err != nil {
			t.Fatal(err)
		}
		tokens[rol] = ses.Token
	}
	pedir := func(rol, metodo, ruta, cuerpo string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
		r.Header.Set("Authorization", "Bearer "+tokens[rol])
		r.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, r)
		return rec
	}
	// Un producto para tocar.
	rec := pedir("duenio", "POST", "/v1/panel/productos", `{"nombre":"Mate","precio":1000000,"costo":400000,"stock":3,"canales":{}}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("crear producto: %d %s", rec.Code, rec.Body)
	}
	var prod struct{ ID string }
	_ = json.Unmarshal(rec.Body.Bytes(), &prod)

	casos := []struct {
		metodo, ruta, cuerpo string
		pueden               []string
		// soloNegar: no se prueba con quienes pueden (cambiaría algo que otros casos usan).
		soloNegar bool
	}{
		{"GET", "/v1/panel/pedidos", "", []string{"duenio", "admin", "ventas", "local"}, false},
		{"GET", "/v1/panel/listas/productos", "", []string{"duenio", "admin", "ventas", "local"}, false},
		{"GET", "/v1/panel/listas/clientes", "", []string{"duenio", "admin", "ventas"}, false},
		{"POST", "/v1/panel/productos", `{"nombre":"Bombilla","precio":5000,"canales":{}}`, []string{"duenio", "admin"}, false},
		{"DELETE", "/v1/panel/productos/" + prod.ID, "", []string{"duenio", "admin"}, true},
		{"POST", "/v1/panel/listas/productos/acciones", `{"ids":["` + prod.ID + `"],"accion":"precio","valor":10}`, []string{"duenio", "admin"}, true},
		{"POST", "/v1/panel/listas-precio/aumento", `{}`, []string{"duenio", "admin"}, true},
		// Finanzas.
		{"GET", "/v1/panel/finanzas", "", []string{"duenio", "admin"}, false},
		{"GET", "/v1/panel/informes", "", []string{"duenio", "admin"}, false},
		{"POST", "/v1/panel/finanzas/retiro", `{}`, []string{"duenio", "admin"}, true},
		{"GET", "/v1/panel/listas/facturacion", "", []string{"duenio", "admin"}, false},
		{"GET", "/v1/panel/arca", "", []string{"duenio", "admin"}, true},
		{"GET", "/v1/panel/integraciones/mercadopago", "", []string{"duenio", "admin"}, true},
		// Configuración de cobros (alias y CBU) y de la tienda.
		{"PUT", "/v1/panel/configuracion/negocio", `{"nombre":"Sol","alias":"sol.deco.mp"}`, []string{"duenio", "admin"}, false},
		{"PUT", "/v1/panel/configuracion/entregas", `{}`, []string{"duenio", "admin"}, true},
		{"PUT", "/v1/panel/mi-tienda", `{}`, []string{"duenio", "admin"}, true},
		// Exportaciones y datos.
		{"GET", "/v1/panel/datos", "", []string{"duenio", "admin"}, false},
		{"POST", "/v1/panel/datos/exportaciones", `{"tipo":"clientes"}`, []string{"duenio", "admin"}, true},
		// Plan y equipo: sólo la dueña.
		{"GET", "/v1/panel/cuenta/plan", "", []string{"duenio"}, false},
		{"PUT", "/v1/panel/plan", `{"plan":"emprende"}`, []string{"duenio"}, false},
		{"POST", "/v1/panel/cuenta/cancelacion", `{"motivo":0,"accion":"oferta"}`, []string{"duenio"}, true},
		{"DELETE", "/v1/panel/cuenta", `{"confirmacion":"sol"}`, []string{"duenio"}, true},
		{"POST", "/v1/panel/equipo", `{"contacto":"nueva@sol.com","rol":"ventas"}`, []string{"duenio"}, false},
		{"DELETE", "/v1/panel/equipo/" + anaID, "", []string{"duenio"}, true},
		// Lo de cualquiera: sus sesiones y su seguridad.
		{"GET", "/v1/panel/configuracion", "", []string{"duenio", "admin", "ventas", "local"}, false},
		{"PUT", "/v1/panel/configuracion/seguridad/avisoIngreso", `{"activo":true}`, []string{"duenio", "admin", "ventas", "local"}, false},
		{"PUT", "/v1/panel/configuracion/avisos/venta", `{"activo":true}`, []string{"duenio", "admin"}, false},
		// Actividad: quién hizo qué, sólo la dueña y administración (docs/decisiones.md §38).
		{"GET", "/v1/panel/actividad", "", []string{"duenio", "admin"}, false},
		{"GET", "/v1/panel/actividad/exportar", "", []string{"duenio", "admin"}, false},
	}
	for _, c := range casos {
		for _, rol := range []string{"duenio", "admin", "ventas", "local"} {
			puede := false
			for _, p := range c.pueden {
				puede = puede || p == rol
			}
			if puede && c.soloNegar {
				continue
			}
			rec := pedir(rol, c.metodo, c.ruta, c.cuerpo)
			if puede == (rec.Code == http.StatusForbidden) || rec.Code == http.StatusUnauthorized {
				t.Errorf("%s %s como %s: %d %s", c.metodo, c.ruta, rol, rec.Code, strings.TrimSpace(rec.Body.String()))
			}
		}
	}

	t.Run("ventas ve los productos sin costo y no cambia el precio al editar", func(t *testing.T) {
		rec := pedir("ventas", "GET", "/v1/panel/listas/productos", "")
		if strings.Contains(rec.Body.String(), `"costo"`) || strings.Contains(rec.Body.String(), `"margen"`) {
			t.Fatalf("ve costos: %s", rec.Body)
		}
		if rec := pedir("duenio", "GET", "/v1/panel/listas/productos", ""); !strings.Contains(rec.Body.String(), `"costo":400000`) {
			t.Fatalf("la dueña no ve el costo: %s", rec.Body)
		}
		rec = pedir("ventas", "PUT", "/v1/panel/productos/"+prod.ID, `{"nombre":"Mate de calabaza","precio":1,"costo":1,"stock":5,"canales":{}}`)
		if rec.Code != http.StatusOK {
			t.Fatalf("editar: %d %s", rec.Code, rec.Body)
		}
		var p struct {
			Nombre        string
			Precio, Costo int64
			Stock         int
		}
		_ = json.Unmarshal(pedir("duenio", "GET", "/v1/panel/productos/"+prod.ID, "").Body.Bytes(), &p)
		if p.Nombre != "Mate de calabaza" || p.Stock != 5 || p.Precio != 1000000 || p.Costo != 400000 {
			t.Fatalf("después de editar como ventas: %+v", p)
		}
		var v struct{ Costo int64 }
		_ = json.Unmarshal(pedir("local", "GET", "/v1/panel/productos/"+prod.ID, "").Body.Bytes(), &v)
		if v.Costo != 0 {
			t.Fatalf("local ve el costo: %+v", v)
		}
	})

	t.Run("la actividad dice quién fue, con su rol, desde la sesión", func(t *testing.T) {
		rec := pedir("admin", "GET", "/v1/panel/actividad?tipo=producto", "")
		if rec.Code != http.StatusOK {
			t.Fatalf("actividad: %d %s", rec.Code, rec.Body)
		}
		var pag struct {
			Items []struct {
				Quien, Texto string
				Rol          *string
			}
			Personas []struct{ ID, Nombre string }
		}
		_ = json.Unmarshal(rec.Body.Bytes(), &pag)
		var ventas, duenia bool
		for _, x := range pag.Items {
			ventas = ventas || (x.Quien == "Persona ventas" && x.Rol != nil && *x.Rol == "Ventas" &&
				x.Texto == "le cambió el nombre a Mate: ahora es Mate de calabaza")
			duenia = duenia || (x.Quien == "Ana" && x.Texto == "creó el producto Mate a $ 1.000.000")
		}
		if !ventas || !duenia {
			t.Fatalf("actividad: %s", rec.Body)
		}
		// Lo que ventas no pudo tocar (el precio) no aparece como cambiado.
		if strings.Contains(rec.Body.String(), "cambió el precio de Mate de calabaza") {
			t.Fatalf("anotó un precio que no cambió: %s", rec.Body)
		}
		if rec := pedir("duenio", "GET", "/v1/panel/actividad?persona=nada", ""); rec.Code != http.StatusBadRequest {
			t.Fatalf("filtro inválido: %d", rec.Code)
		}
		if rec := pedir("duenio", "GET", "/v1/panel/actividad", ""); !strings.Contains(rec.Body.String(), `"quien":"Vos"`) {
			t.Fatalf("lo propio dice Vos: %s", rec.Body)
		}
	})

	t.Run("quien no es dueña no recibe el CBU ni el equipo", func(t *testing.T) {
		rec := pedir("local", "GET", "/v1/panel/configuracion", "")
		if strings.Contains(rec.Body.String(), "sol.deco.mp") || strings.Contains(rec.Body.String(), "ana@sol.com") {
			t.Fatalf("local ve: %s", rec.Body)
		}
		if rec := pedir("duenio", "GET", "/v1/panel/configuracion", ""); !strings.Contains(rec.Body.String(), "sol.deco.mp") {
			t.Fatalf("la dueña no ve su alias: %s", rec.Body)
		}
	})

	t.Run("acceso: rol y permisos de la sesión", func(t *testing.T) {
		var a cuentas.Acceso
		_ = json.Unmarshal(pedir("local", "GET", "/v1/auth/tiendas", "").Body.Bytes(), &a)
		if a.Rol != "local" || len(a.Permisos) != 2 || len(a.Tiendas) != 1 || a.Tiendas[0].Nombre != "Sol" {
			t.Fatalf("acceso: %+v", a)
		}
	})
}
