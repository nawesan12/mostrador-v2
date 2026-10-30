package correoargentino_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	ca "github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/correoargentino"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// miCorreoFalso responde con la forma de la API MiCorreo (docs/integraciones/correo-argentino.md):
// /token con Basic, el resto con Bearer, errores {"code", "message"}.
type miCorreoFalso struct {
	mu         sync.Mutex
	usuario    string
	clave      string
	tokens     int
	invalidar  bool // el próximo pedido con Bearer da 401 (token vencido antes de tiempo)
	tarifas    int
	agencias   int
	importados map[string]map[string]any
	ultimo     map[string]any
}

func (f *miCorreoFalso) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	errorCA := func(estado int, codigo any, msg string) {
		w.WriteHeader(estado)
		_ = json.NewEncoder(w).Encode(map[string]any{"code": codigo, "message": msg})
	}
	ruta := strings.TrimPrefix(r.URL.Path, "/micorreo/v1")
	if ruta == "/token" {
		u, c, ok := r.BasicAuth()
		if r.Method != http.MethodPost || !ok || u != f.usuario || c != f.clave {
			errorCA(401, 401, "Unauthorized")
			return
		}
		f.tokens++
		_ = json.NewEncoder(w).Encode(map[string]any{"token": "jwt-" + u, "expires": "2099-12-31 23:59:59"})
		return
	}
	if r.Header.Get("Authorization") != "Bearer jwt-"+f.usuario || f.invalidar {
		f.invalidar = false
		errorCA(401, "401", "Unauthorized")
		return
	}
	var cuerpo map[string]any
	b, _ := io.ReadAll(r.Body)
	_ = json.Unmarshal(b, &cuerpo)
	switch {
	case ruta == "/users/validate" && r.Method == http.MethodPost:
		if cuerpo["email"] != "luna@ejemplo.com" || cuerpo["password"] != "secreta" {
			errorCA(404, "404", "Usuario no valido o inexistente")
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"customerId": "0090000025", "createdAt": "2021-03-10"})
	case ruta == "/rates" && r.Method == http.MethodPost:
		f.tarifas++
		f.ultimo = cuerpo
		if cuerpo["customerId"] != "0090000025" {
			errorCA(402, "402", "Cliente FAP no identificado")
			return
		}
		todas := []map[string]any{
			{"deliveredType": "D", "productType": "CP", "productName": "Paq.ar Clásico", "price": 4980.06, "deliveryTimeMin": "2", "deliveryTimeMax": 5},
			{"deliveredType": "S", "productType": "CP", "productName": "Paq.ar Clásico", "price": 3980.5},
		}
		var rates []map[string]any
		for _, t := range todas {
			if dt, _ := cuerpo["deliveredType"].(string); dt == "" || dt == t["deliveredType"] {
				rates = append(rates, t)
			}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"customerId": "0090000025", "validTo": "2099-06-07T10:31:27.881-03:00", "rates": rates})
	case ruta == "/agencies" && r.Method == http.MethodGet:
		f.agencias++
		q := r.URL.Query()
		if q.Get("customerId") != "0090000025" {
			errorCA(402, "402", "Customer ID no valido")
			return
		}
		if q.Get("provinceCode") != "B" {
			_, _ = w.Write([]byte("[]"))
			return
		}
		ag := func(code, name, cp, status string, retiro bool) map[string]any {
			return map[string]any{"code": code, "name": name, "manager": "x", "email": "x@correoargentino.com.ar", "phone": "x",
				"services": map[string]any{"packageReception": true, "pickupAvailability": retiro},
				"location": map[string]any{"address": map[string]any{"streetName": "Vicente Lopez", "streetNumber": "448",
					"floor": nil, "apartment": nil, "locality": name, "city": name, "province": "Buenos Aires", "provinceCode": "B",
					"postalCode": cp}, "latitude": "-34.8", "longitude": "-58.4"},
				"hours": map[string]any{"sunday": nil, "monday": map[string]any{"start": "0930", "end": "1800"},
					"friday": map[string]any{"start": "1000", "end": "1800"}, "saturday": nil, "holidays": nil},
				"status": status}
		}
		_ = json.NewEncoder(w).Encode([]any{
			ag("B0500", "La Plata", "B1900ABC", "ACTIVE", true),
			ag("B0107", "Monte Grande", "B1842ZAB", "ACTIVE", true),
			ag("B0999", "Sin retiro", "B1843AAA", "ACTIVE", false),
			ag("B0998", "Cerrada", "B1842AAA", "INACTIVE", true),
		})
	case ruta == "/shipping/import" && r.Method == http.MethodPost:
		ext, _ := cuerpo["extOrderId"].(string)
		if _, ya := f.importados[ext]; ya {
			errorCA(402, "402", "La orden ya fue importada con anterioridad.")
			return
		}
		f.importados[ext] = cuerpo
		_ = json.NewEncoder(w).Encode(map[string]any{"createdAt": "2026-10-30T16:15:04.996-03:00"})
	default:
		errorCA(404, "404", "Resource not found")
	}
}

func clave(t *testing.T) *cifrado.Cifrador {
	k := make([]byte, 32)
	_, _ = rand.Read(k)
	c, err := cifrado.Nuevo(base64.StdEncoding.EncodeToString(k))
	if err != nil {
		t.Fatal(err)
	}
	return c
}

var origen = ca.Origen{CP: "1842", Calle: "Belgrano", Altura: "742", Localidad: "Monte Grande", Provincia: "b"}
var paquete = ca.Paquete{Peso: 800, Alto: 10, Ancho: 20, Largo: 30}

func TestCorreoArgentinoDePuntaAPunta(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	f := &miCorreoFalso{usuario: "tulocalweb", clave: "api-secreta", importados: map[string]map[string]any{}}
	srv := httptest.NewServer(f)
	defer srv.Close()
	s := &ca.Servicio{Pool: b.App, Cifra: clave(t), Cliente: &ca.Cliente{Base: srv.URL + "/micorreo/v1"},
		Plataforma: ca.Credenciales{Usuario: "tulocalweb", Clave: "api-secreta"}}

	// Sin conectar: el estado lo dice y no se cotiza.
	if e, err := s.EstadoDe(ctx, luna); err != nil || e.Conectada || !e.Disponible || e.PideCredenciales {
		t.Fatalf("estado inicial: %+v %v", e, err)
	}
	if _, err := s.Cotizar(ctx, luna, "1704", 1); !errors.Is(err, ca.ErrSinConexion) {
		t.Fatalf("cotizar sin conexión: %v", err)
	}

	// Contraseña de MiCorreo mal: no se guarda nada.
	alta := ca.Alta{Email: "luna@ejemplo.com", Clave: "otra", Origen: origen, Paquete: paquete}
	if _, err := s.Conectar(ctx, luna, alta); !errors.Is(err, ca.ErrCuenta) {
		t.Fatalf("contraseña incorrecta: %v", err)
	}
	// Datos que no pasan antes de ir a Correo.
	malo := alta
	malo.Paquete.Peso = 30000
	var inv ca.ErrInvalido
	if _, err := s.Conectar(ctx, luna, malo); !errors.As(err, &inv) {
		t.Fatalf("peso fuera de rango: %v", err)
	}
	alta.Clave = "secreta"
	e, err := s.Conectar(ctx, luna, alta)
	if err != nil || !e.Conectada || e.Cliente != "0090000025" || e.Propias || e.Origen.Provincia != "B" {
		t.Fatalf("conectar: %+v %v", e, err)
	}
	// La contraseña de MiCorreo no queda en la base, y sol no ve la conexión de luna (RLS).
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		var fila string
		if err := tx.QueryRow(ctx, `SELECT row_to_json(c)::text FROM ca_conexiones c`).Scan(&fila); err != nil {
			t.Fatal(err)
		}
		if strings.Contains(fila, "secreta") {
			t.Fatalf("se guardó la contraseña: %s", fila)
		}
		return nil
	})
	if e, _ := s.EstadoDe(ctx, sol); e.Conectada {
		t.Fatal("sol ve la conexión de luna")
	}

	// Cotizar: las dos modalidades, redondeadas para arriba; el peso es por unidad. Lo mismo dos veces
	// no vuelve a ir a Correo.
	c, err := s.Cotizar(ctx, luna, "1704", 3)
	if err != nil || c.Domicilio == nil || c.Domicilio.Precio != 4981 || c.Sucursal == nil || c.Sucursal.Precio != 3981 ||
		c.Domicilio.DiasMin != "2" || c.Domicilio.DiasMax != "5" {
		t.Fatalf("cotización: %+v %v", c, err)
	}
	dims := f.ultimo["dimensions"].(map[string]any)
	if f.ultimo["postalCodeOrigin"] != "1842" || f.ultimo["postalCodeDestination"] != "1704" || dims["weight"] != float64(2400) ||
		f.ultimo["deliveredType"] != nil {
		t.Fatalf("lo que se mandó a /rates: %+v", f.ultimo)
	}
	if _, err := s.Cotizar(ctx, luna, "1704", 3); err != nil || f.tarifas != 1 {
		t.Fatalf("no usó lo guardado: %d %v", f.tarifas, err)
	}
	if _, err := s.Cotizar(ctx, luna, "17", 1); !errors.As(err, &inv) {
		t.Fatalf("CP inválido: %v", err)
	}
	// Un token para todo; si Correo lo invalida antes de tiempo, se pide otro y se reintenta.
	if f.tokens != 1 {
		t.Fatalf("tokens: %d", f.tokens)
	}
	f.invalidar = true
	if _, err := s.Cotizar(ctx, luna, "5000", 1); err != nil || f.tokens != 2 {
		t.Fatalf("reintento con token nuevo: %d %v", f.tokens, err)
	}

	// Sucursales: sólo activas y que entregan, las más cercanas al CP primero, con el horario.
	ss, err := s.Sucursales(ctx, luna, "b", "1842")
	if err != nil || len(ss) != 2 || ss[0].Codigo != "B0107" || ss[1].Codigo != "B0500" || ss[0].Horario != "Lun a Vie 09:30 a 18:00" ||
		ss[0].Direccion != "Vicente Lopez 448" {
		t.Fatalf("sucursales: %+v %v", ss, err)
	}
	if _, err := s.SucursalesPublico(ctx, "luna", "C", ""); err != nil || f.agencias != 2 {
		t.Fatalf("otra provincia: %d %v", f.agencias, err)
	}
	if _, err := s.Sucursales(ctx, luna, "Ñ", ""); !errors.As(err, &inv) {
		t.Fatalf("provincia inválida: %v", err)
	}

	// La tienda online: Correo se ofrece sólo conectada y el precio sale de la API, no del navegador.
	if err := configuracion.GuardarNegocio(ctx, b.App, luna, configuracion.Negocio{Nombre: "Luna", Alias: "luna.tienda", Whatsapp: "11 5555-0000"}); err != nil {
		t.Fatal(err)
	}
	stock := 10
	p, err := listas.Guardar(ctx, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 10000, Stock: &stock})
	if err != nil {
		t.Fatal(err)
	}
	if pub, _ := vidriera.Publica(ctx, b.App, "luna"); !pub.Envios.Correo {
		t.Fatal("la tienda conectada no ofrece Correo")
	}
	if pub, _ := vidriera.Publica(ctx, b.App, "sol"); pub.Envios.Correo {
		t.Fatal("sol ofrece Correo sin conectarlo")
	}
	ahora := time.Date(2026, 10, 30, 12, 0, 0, 0, time.UTC)
	compra := vidriera.Compra{Email: "ana@ejemplo.com", Nombre: "Ana Gómez", Telefono: "11 5555-0101", CP: "1704",
		Direccion: "Av. Rivadavia 1234 3B", Envio: vidriera.CorreoDomicilio, Pago: "transferencia", Provincia: "B", Localidad: "Ramos Mejía",
		Carrito: []vidriera.LineaCarrito{{ID: p.ID, Cantidad: 3}}}
	r, err := vidriera.ComprarConCorreo(ctx, b.App, nil, s, "", "luna", compra, ahora)
	if err != nil {
		t.Fatal(err)
	}
	// Sin promociones prendidas, sin descuento por transferencia.
	if want := int64(30000 + 4981); r.Total != want {
		t.Fatalf("total con Correo a domicilio: %d, quería %d", r.Total, want)
	}
	// A sucursal: tiene que ser una que exista y entregue.
	suc := compra
	suc.Envio, suc.Sucursal, suc.Email, suc.Carrito = vidriera.CorreoSucursal, "B0999", "beto@ejemplo.com", []vidriera.LineaCarrito{{ID: p.ID, Cantidad: 1}}
	if _, err := vidriera.ComprarConCorreo(ctx, b.App, nil, s, "", "luna", suc, ahora); !errors.Is(err, vidriera.ErrInvalido) {
		t.Fatalf("sucursal sin retiro: %v", err)
	}
	suc.Sucursal = "B0107"
	r2, err := vidriera.ComprarConCorreo(ctx, b.App, nil, s, "", "luna", suc, ahora)
	if err != nil || r2.Total != 10000+3981 {
		t.Fatalf("a sucursal: %+v %v", r2, err)
	}
	// Sin Correo conectado en sol, o sin el servicio, la opción no existe.
	if _, err := vidriera.Comprar(ctx, b.App, nil, "", "luna", compra, ahora); !errors.Is(err, vidriera.ErrInvalido) {
		t.Fatalf("sin servicio: %v", err)
	}

	// Generar el envío en MiCorreo: el remitente es la tienda, la dirección se separa en calle y altura.
	if e, err := s.EnvioDe(ctx, luna, r.Numero); err != nil || !e.Conectada || e.Generado {
		t.Fatalf("antes de generar: %+v %v", e, err)
	}
	if _, err := s.Seguimiento(ctx, luna, r.Numero, "CP123456789AR"); !errors.As(err, &inv) {
		t.Fatalf("seguimiento antes de generar: %v", err)
	}
	g, err := s.Generar(ctx, luna, r.Numero)
	if err != nil || !g.Generado || g.Tipo != "D" {
		t.Fatalf("generar: %+v %v", g, err)
	}
	if len(f.importados) != 1 {
		t.Fatalf("importados: %d", len(f.importados))
	}
	var alta1 map[string]any
	for _, v := range f.importados {
		alta1 = v
	}
	envio := alta1["shipping"].(map[string]any)
	dir := envio["address"].(map[string]any)
	if alta1["customerId"] != "0090000025" || alta1["orderNumber"] != "1001" || !strings.HasPrefix(alta1["extOrderId"].(string), "TLW-") ||
		envio["deliveryType"] != "D" || envio["agency"] != nil || envio["weight"] != float64(2400) || envio["declaredValue"] != float64(r.Total) ||
		dir["streetName"] != "Av. Rivadavia" || dir["streetNumber"] != "1234" || dir["floor"] != "3B" || dir["city"] != "Ramos Mejía" ||
		dir["provinceCode"] != "B" || dir["postalCode"] != "1704" {
		t.Fatalf("alta en MiCorreo: %+v", alta1)
	}
	rec := alta1["recipient"].(map[string]any)
	rem := alta1["sender"].(map[string]any)
	if rec["email"] != "ana@ejemplo.com" || rec["name"] != "Ana Gómez" || rem["name"] != "Luna" || rem["email"] != "luna@ejemplo.com" ||
		rem["originAddress"].(map[string]any)["postalCode"] != "1842" {
		t.Fatalf("personas: %+v %+v", rec, rem)
	}
	// Un segundo toque no da de alta otro.
	if g, err := s.Generar(ctx, luna, r.Numero); err != nil || !g.Generado || len(f.importados) != 1 {
		t.Fatalf("segundo toque: %+v %v %d", g, err, len(f.importados))
	}
	// Si la respuesta se cortó y MiCorreo ya la tenía ("ya fue importada"), queda como hecha.
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		f.importados["TLW-"+strings.ReplaceAll(luna, "-", "")[:8]+"-1002"] = map[string]any{}
		return nil
	})
	if g, err := s.Generar(ctx, luna, r2.Numero); err != nil || !g.Generado || g.Tipo != "S" || g.Sucursal == nil || *g.Sucursal != "B0107" {
		t.Fatalf("recuperar la importada: %+v %v", g, err)
	}

	// El seguimiento lo carga la tienda y queda en el pedido.
	if g, err := s.Seguimiento(ctx, luna, r.Numero, " cp 1234 5678 9ar "); err != nil || g.Seguimiento == nil || *g.Seguimiento != "CP123456789AR" {
		t.Fatalf("seguimiento: %+v %v", g, err)
	}
	var seg *string
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT seguimiento FROM pedidos WHERE numero = $1`, r.Numero).Scan(&seg)
	})
	if seg == nil || *seg != "CP123456789AR" {
		t.Fatalf("seguimiento en el pedido: %v", seg)
	}
	if _, err := s.EnvioDe(ctx, sol, r.Numero); !errors.Is(err, ca.ErrNoExiste) {
		t.Fatalf("sol ve el pedido de luna: %v", err)
	}

	// Desconectar: deja de ofrecerse y no se cotiza más.
	if err := s.Desconectar(ctx, luna); err != nil {
		t.Fatal(err)
	}
	if pub, _ := vidriera.Publica(ctx, b.App, "luna"); pub.Envios.Correo {
		t.Fatal("sigue ofreciendo Correo")
	}
	if _, err := s.Cotizar(ctx, luna, "1704", 3); !errors.Is(err, ca.ErrSinConexion) {
		t.Fatalf("cotiza desconectada: %v", err)
	}
}

// Una tienda con credenciales de API propias: van cifradas y se usan en vez de las de la plataforma.
func TestCredencialesPropias(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	f := &miCorreoFalso{usuario: "propia", clave: "clave-propia", importados: map[string]map[string]any{}}
	srv := httptest.NewServer(f)
	defer srv.Close()
	// Tu Local Web sin credenciales: la tienda tiene que traer las suyas.
	s := &ca.Servicio{Pool: b.App, Cifra: clave(t), Cliente: &ca.Cliente{Base: srv.URL}}
	if e, _ := s.EstadoDe(ctx, luna); !e.PideCredenciales {
		t.Fatal("no pide credenciales")
	}
	alta := ca.Alta{Email: "luna@ejemplo.com", Clave: "secreta", Origen: origen, Paquete: paquete}
	if _, err := s.Conectar(ctx, luna, alta); !errors.Is(err, ca.ErrSinCredenciales) {
		t.Fatalf("sin credenciales: %v", err)
	}
	alta.API = &ca.Credenciales{Usuario: "propia", Clave: "mal"}
	var inv ca.ErrInvalido
	if _, err := s.Conectar(ctx, luna, alta); !errors.As(err, &inv) {
		t.Fatalf("credenciales de API incorrectas: %v", err)
	}
	alta.API.Clave = "clave-propia"
	e, err := s.Conectar(ctx, luna, alta)
	if err != nil || !e.Propias {
		t.Fatalf("conectar con propias: %+v %v", e, err)
	}
	var cred []byte
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT credencial FROM ca_conexiones`).Scan(&cred)
	})
	if len(cred) == 0 || bytes.Contains(cred, []byte("clave-propia")) {
		t.Fatalf("credencial sin cifrar: %q", cred)
	}
	if c, err := s.Cotizar(ctx, luna, "1704", 1); err != nil || c.Sucursal == nil {
		t.Fatalf("cotizar con propias: %+v %v", c, err)
	}
	// Ajustar el paquete sin volver a pedir la contraseña: lo cotizado antes no se reusa.
	p := paquete
	p.Peso = 1500
	if _, err := s.Ajustar(ctx, luna, origen, p); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Cotizar(ctx, luna, "1704", 1); err != nil || f.tarifas != 2 || f.ultimo["dimensions"].(map[string]any)["weight"] != float64(1500) {
		t.Fatalf("después de ajustar: %d %+v %v", f.tarifas, f.ultimo, err)
	}
}

// Las rutas públicas no dejan ver lo que respondió Correo.
func TestRutasPublicas(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	f := &miCorreoFalso{usuario: "tulocalweb", clave: "api-secreta", importados: map[string]map[string]any{}}
	srv := httptest.NewServer(f)
	defer srv.Close()
	s := &ca.Servicio{Pool: b.App, Cifra: clave(t), Cliente: &ca.Cliente{Base: srv.URL},
		Plataforma: ca.Credenciales{Usuario: "tulocalweb", Clave: "api-secreta"}}
	mux := http.NewServeMux()
	s.Rutas(mux, nil, nil)
	pedir := func(ruta string) (int, string) {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodGet, ruta, nil))
		return w.Code, w.Body.String()
	}
	if code, _ := pedir("/v1/publico/tiendas/luna/correo/cotizacion?cp=1704"); code != http.StatusConflict {
		t.Fatalf("sin conectar: %d", code)
	}
	if _, err := s.Conectar(ctx, luna, ca.Alta{Email: "luna@ejemplo.com", Clave: "secreta", Origen: origen, Paquete: paquete}); err != nil {
		t.Fatal(err)
	}
	code, cuerpo := pedir("/v1/publico/tiendas/luna/correo/cotizacion?cp=1704&unidades=2")
	if code != http.StatusOK || !strings.Contains(cuerpo, `"precio":4981`) {
		t.Fatalf("cotización: %d %s", code, cuerpo)
	}
	// Correo rechaza: a quien compra no le llega el mensaje de Correo.
	_ = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE ca_conexiones SET cliente = '0000000000'`)
		return err
	})
	s2 := &ca.Servicio{Pool: b.App, Cifra: s.Cifra, Cliente: s.Cliente, Plataforma: s.Plataforma}
	mux2 := http.NewServeMux()
	s2.Rutas(mux2, nil, nil)
	w := httptest.NewRecorder()
	mux2.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/publico/tiendas/luna/correo/cotizacion?cp=1704", nil))
	if w.Code != http.StatusBadGateway || strings.Contains(w.Body.String(), "FAP") {
		t.Fatalf("rechazo: %d %s", w.Code, w.Body.String())
	}
	if code, _ := pedir("/v1/publico/tiendas/no-existe/correo/sucursales?provincia=B"); code != http.StatusNotFound {
		t.Fatalf("tienda inexistente: %d", code)
	}
}
