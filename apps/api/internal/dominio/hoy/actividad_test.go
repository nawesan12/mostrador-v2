package hoy_test

import (
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cupones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/facturacion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/hoy"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mitienda"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/precios"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/correoargentino"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadolibre"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadopago"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// Cada acción que importa anota quién la hizo, qué y sobre qué, en la misma transacción
// (docs/decisiones.md §38). Sin autor en el contexto (el worker, la tienda online), no se anota.
func TestCadaAccionQuedaEnLaActividad(t *testing.T) {
	b := pruebadb.Nueva(t)
	cs := &cuentas.Servicio{Pool: b.App, Correo: &correo.Memoria{}, Sitio: "https://tulocalweb.test"}
	ses, err := cs.Registrar(ctx, cuentas.Alta{Nombre: "Valeria", Tienda: "Luna", Email: "vale@luna.com", Clave: "clave-larga-1"},
		cuentas.Cliente{IP: "190.2.3.4"})
	if err != nil {
		t.Fatal(err)
	}
	id := ses.TiendaID
	var yo string
	_ = b.App.QueryRow(ctx, `SELECT id::text FROM auth.usuarios WHERE email = 'vale@luna.com'`).Scan(&yo)
	leo := auditoria.Actor{Tipo: "persona", ID: "8f0c1d3e-0000-4000-8000-000000000001", Nombre: "Leo", Rol: "ventas", IP: "190.2.3.5", Agente: "Chrome · Android"}
	vale := auditoria.Actor{Tipo: "persona", ID: yo, Nombre: "Valeria", Rol: "duenio"}
	comoLeo := auditoria.ConActor(ctx, leo)
	comoVale := auditoria.ConActor(ctx, vale)

	// Sin autor (lo que corre el worker o una importación): nada.
	stock := 10
	if _, err := listas.Guardar(ctx, b.App, id, "", listas.Editable{Nombre: "Yerba", Precio: 5000, Stock: &stock}); err != nil {
		t.Fatal(err)
	}

	costo := int64(6000)
	p, err := listas.Guardar(comoVale, b.App, id, "", listas.Editable{Nombre: "Mate imperial", Precio: 12000, Costo: &costo, Stock: &stock})
	must(t, err)
	e := listas.Editable{Nombre: "Mate imperial", Precio: 13500, Costo: &costo, Stock: &stock}
	_, err = listas.Guardar(comoLeo, b.App, id, p.ID, e)
	must(t, err)
	nuevoStock, nuevoCosto := 7, int64(6500)
	e.Stock, e.Costo = &nuevoStock, &nuevoCosto
	_, err = listas.Guardar(comoVale, b.App, id, p.ID, e)
	must(t, err)
	_, err = listas.Publicar(comoVale, b.App, id, []string{p.ID}, false)
	must(t, err)
	_, err = listas.Publicar(comoVale, b.App, id, []string{p.ID}, true)
	must(t, err)
	_, err = listas.AjustarStock(comoLeo, b.App, id, []string{p.ID}, 3)
	must(t, err)
	_, err = listas.Precio(comoVale, b.App, id, []string{p.ID}, 10)
	must(t, err)
	copia, err := listas.Duplicar(comoVale, b.App, id, p.ID, time.Now())
	must(t, err)
	must(t, listas.Borrar(comoVale, b.App, id, copia.ID))

	// Pedidos y punto de venta.
	cliente := "Ana"
	no := false
	manual := pedidos.PedidoNuevo{Pago: "transf", Canal: "wa", Envio: "retiro", Cliente: &cliente, Pagado: &no,
		Items: []pedidos.Linea{{ProductoID: p.ID, Cantidad: 1}}}
	numero, _, err := pedidos.CrearDesdeProductos(comoLeo, b.App, id, manual)
	must(t, err)
	must(t, pedidos.CambiarEstado(comoLeo, b.App, id, numero, "Para preparar"))
	must(t, pedidos.CambiarEstado(comoLeo, b.App, id, numero, "Sin pagar"))
	must(t, pedidos.CambiarEstado(comoVale, b.App, id, numero, "Cancelado"))
	pos := pedidos.PedidoNuevo{Pago: "efectivo", Canal: "pos", Envio: "retiro", Items: []pedidos.Linea{{ProductoID: p.ID, Cantidad: 2}}}
	_, _, err = pedidos.CrearDesdeProductos(comoLeo, b.App, id, pos)
	must(t, err)

	// Cupones, promociones y precios.
	c, err := cupones.Crear(comoVale, b.App, id, cupones.Nuevo{Codigo: "FINDE10", Tipo: "pct", Valor: 10, Alcance: "todo", Quien: "todos",
		PorCliente: "libre", Vigencia: "sin", Canales: []string{"web"}}, time.Now())
	must(t, err)
	_, err = cupones.Accion(comoVale, b.App, id, c.ID, "pausar", 0, time.Now())
	must(t, err)
	must(t, cupones.FijarAutomatica(comoVale, b.App, id, "transf", true))
	_, _, err = precios.Aplicar(comoVale, b.App, id, 5, "no", false, time.Now())
	must(t, err)

	// Configuración: cobros (el CBU nunca entero), envíos, dominio, plan, exportaciones, sesiones y Mi tienda.
	cbu := "2850590940090418135201"
	must(t, configuracion.GuardarNegocio(comoVale, b.App, id, configuracion.Negocio{Nombre: "Luna Deco", Alias: "luna.deco.mp", CBU: cbu, Titular: "Valeria Gómez"}))
	must(t, configuracion.GuardarEntregas(comoVale, b.App, id, configuracion.Entregas{Retiro: true, RetiroDireccion: "Av. Corrientes 1234",
		Moto: true, MotoCosto: 3500, MotoZona: "Palermo y Belgrano"}))
	must(t, configuracion.ConectarDominio(comoVale, b.App, id, "lunadeco.com.ar"))
	must(t, configuracion.QuitarDominio(comoVale, b.App, id, "lunadeco.com.ar"))
	must(t, configuracion.CambiarPlan(comoVale, b.App, id, "emprende", "mes"))
	must(t, configuracion.RegistrarExportacion(comoVale, b.App, id, "clientes", "csv", time.Now()))
	q := configuracion.Quien{UsuarioID: yo, SesionID: "otra", Rol: "duenio"}
	_, err = configuracion.CerrarOtras(comoVale, b.App, id, q)
	must(t, err)
	must(t, mitienda.Publicar(comoVale, b.App, id))

	// Equipo: invitar, cambiar el rol de la invitación y cancelarla.
	inv, err := configuracion.Invitar(comoVale, b.App, cs.Sitio, id, q, "leo@luna.com", "ventas")
	must(t, err)
	must(t, configuracion.CambiarRol(comoVale, b.App, id, q, inv.ID, "local"))
	must(t, configuracion.Quitar(comoVale, b.App, id, q, inv.ID))

	// Integraciones: desconectar Mercado Pago, Mercado Libre y Correo, y los ajustes de ARCA.
	must(t, b.App.EnTienda(ctx, id, func(tx pgx.Tx) error {
		for _, s := range []string{
			`INSERT INTO mp_conexiones (tienda_id, mp_usuario, acceso, renovacion, vence_en) VALUES ($1, 11, '\x00', '\x00', now())`,
			`INSERT INTO ml_conexiones (tienda_id, ml_usuario, acceso, renovacion, vence_en) VALUES ($1, 22, '\x00', '\x00', now())`,
			`INSERT INTO ca_conexiones (tienda_id, cliente, email, origen, paquete) VALUES ($1, 'C1', 'vale@luna.com', '{}', '{}')`,
			`INSERT INTO arca_config (tienda_id, cuit, razon_social, condicion, punto_venta, ambiente, clave)
				VALUES ($1, '20123456786', 'Luna', 'monotributo', 1, 'homologacion', '\x00')`,
		} {
			if _, err := tx.Exec(ctx, s, id); err != nil {
				return err
			}
		}
		return nil
	}))
	must(t, (&mercadopago.Servicio{Pool: b.App}).Desconectar(comoVale, id))
	must(t, (&mercadolibre.Servicio{Pool: b.App}).Desconectar(comoVale, id))
	must(t, (&correoargentino.Servicio{Pool: b.App}).Desconectar(comoVale, id))
	pv := 3
	must(t, (&facturacion.Servicio{Pool: b.App}).Ajustar(comoVale, id, "duenio", facturacion.Ajustes{PuntoVenta: &pv}))

	// Todo junto: la frase, el autor y el tipo.
	var todo []hoy.Evento
	f := hoy.FiltroActividad{}
	for {
		pag, err := hoy.Actividad(ctx, b.App, id, yo, f)
		must(t, err)
		todo = append(todo, pag.Items...)
		if pag.Siguiente == nil {
			break
		}
		f.Cursor = *pag.Siguiente
	}
	quiere := []struct{ quien, categoria, texto string }{
		{"Vos", "producto", "creó el producto Mate imperial a $ 12.000"},
		{"Leo", "producto", "cambió el precio de Mate imperial de $ 12.000 a $ 13.500"},
		{"Vos", "producto", "cambió el costo de Mate imperial de $ 6.000 a $ 6.500"},
		{"Vos", "producto", "cambió el stock de Mate imperial de 10 a 7"},
		{"Vos", "producto", "pausó Mate imperial"},
		{"Vos", "producto", "publicó Mate imperial"},
		{"Leo", "producto", "cambió el stock de Mate imperial de 7 a 10"},
		{"Vos", "producto", "cambió el precio de Mate imperial de $ 13.500 a $ 14.850"},
		{"Vos", "producto", "duplicó Mate imperial (la copia quedó pausada)"},
		{"Vos", "producto", "borró el producto Mate imperial (copia)"},
		{"Leo", "pedido", "cargó a mano el pedido #1001 de Ana por $ 14.850"},
		{"Leo", "pedido", "marcó como pagado el pedido #1001"},
		{"Leo", "pedido", "deshizo el pago del pedido #1001 (volvió a sin pagar)"},
		{"Vos", "pedido", "canceló el pedido #1001"},
		{"Leo", "venta", "cobró $ 29.700 en el punto de venta (venta #1002, efectivo)"},
		{"Vos", "promocion", "creó el cupón FINDE10 (10 % off en toda la tienda)"},
		{"Vos", "promocion", "pausó el cupón FINDE10"},
		{"Vos", "promocion", "prendió la promoción «Descuento por transferencia»"},
		{"Vos", "producto", "subió un 5 % los precios de todo el catálogo (2 productos)"},
		{"Vos", "negocio", "cambió el nombre de la tienda de «Luna» a «Luna Deco»"},
		{"Vos", "cobros", "cargó el alias para cobrar por transferencia: luna.deco.mp"},
		{"Vos", "cobros", "cargó el CBU para cobrar por transferencia: uno terminado en 5201"},
		{"Vos", "cobros", "cargó el titular para cobrar por transferencia: Valeria Gómez"},
		{"Vos", "envios", "cambió la dirección o el horario de retiro"},
		{"Vos", "envios", "prendió el envío en moto a $ 3.500"},
		{"Vos", "dominio", "conectó el dominio lunadeco.com.ar"},
		{"Vos", "dominio", "quitó el dominio lunadeco.com.ar"},
		{"Vos", "plan", "cambió el plan de Gratis a Emprende"},
		{"Vos", "datos", "bajó los clientes en CSV"},
		{"Vos", "sesion", "cerró sus otras sesiones abiertas (1)"},
		{"Vos", "mitienda", "publicó el diseño de la tienda (Mi tienda)"},
		{"Vos", "personas", "invitó a leo@luna.com al equipo como Ventas"},
		{"Vos", "personas", "le cambió el rol a la invitación de leo@luna.com de Ventas a Local"},
		{"Vos", "personas", "canceló la invitación de leo@luna.com (Local)"},
		{"Vos", "integracion", "desconectó Mercado Pago"},
		{"Vos", "integracion", "desconectó Mercado Libre"},
		{"Vos", "integracion", "desconectó Correo Argentino"},
		{"Vos", "integracion", "cambió los ajustes de ARCA: punto de venta 3"},
	}
	hay := map[string]hoy.Evento{}
	for _, x := range todo {
		hay[x.Texto] = x
		if strings.Contains(x.Texto, cbu) {
			t.Errorf("el CBU quedó entero: %q", x.Texto)
		}
		if strings.Contains(x.Texto, "Yerba") {
			t.Errorf("sin autor no se anota: %q", x.Texto)
		}
	}
	for _, w := range quiere {
		x, ok := hay[w.texto]
		if !ok {
			t.Errorf("falta %q", w.texto)
			continue
		}
		if x.Quien != w.quien || x.Categoria != w.categoria || x.Tipo != "equipo" {
			t.Errorf("%q: quien=%q categoria=%q tipo=%q", w.texto, x.Quien, x.Categoria, x.Tipo)
		}
	}
	if t.Failed() {
		for _, x := range todo {
			t.Logf("hay: %s %s [%s]", x.Quien, x.Texto, x.Categoria)
		}
	}
	if x := hay["cambió el precio de Mate imperial de $ 12.000 a $ 13.500"]; x.Rol == nil || *x.Rol != "Ventas" || x.Donde != "Chrome · Android" {
		t.Errorf("rol y desde dónde de Leo: %+v", x)
	}

	t.Run("filtros por persona, tipo y fecha", func(t *testing.T) {
		pag, err := hoy.Actividad(ctx, b.App, id, yo, hoy.FiltroActividad{Persona: leo.ID})
		must(t, err)
		for _, x := range pag.Items {
			if x.Quien != "Leo" {
				t.Fatalf("filtro por Leo trajo: %+v", x)
			}
		}
		if len(pag.Items) != 6 {
			t.Fatalf("de Leo hay %d", len(pag.Items))
		}
		pag, err = hoy.Actividad(ctx, b.App, id, yo, hoy.FiltroActividad{Categoria: "venta"})
		must(t, err)
		if len(pag.Items) != 1 || pag.Items[0].Categoria != "venta" {
			t.Fatalf("filtro por punto de venta: %+v", pag.Items)
		}
		hoyAR := time.Now().In(time.FixedZone("ART", -3*3600)).Format("2006-01-02")
		manana := time.Now().In(time.FixedZone("ART", -3*3600)).AddDate(0, 0, 1).Format("2006-01-02")
		if pag, _ := hoy.Actividad(ctx, b.App, id, yo, hoy.FiltroActividad{Desde: hoyAR, Hasta: hoyAR}); len(pag.Items) != min(len(todo), hoy.TamPagina) && time.Now().In(time.FixedZone("ART", -3*3600)).Hour() > 0 {
			t.Fatalf("hoy: %d", len(pag.Items))
		}
		if pag, _ := hoy.Actividad(ctx, b.App, id, yo, hoy.FiltroActividad{Desde: manana}); len(pag.Items) != 0 {
			t.Fatalf("mañana: %d", len(pag.Items))
		}
		if len(pag.Personas) != 2 {
			t.Fatalf("personas para el filtro: %+v", pag.Personas)
		}
		for _, mal := range []hoy.FiltroActividad{{Persona: "' OR 1=1"}, {Categoria: "nada"}, {Desde: "ayer"}, {Cursor: "%%"}} {
			if _, err := hoy.Actividad(ctx, b.App, id, yo, mal); err != hoy.ErrFiltro {
				t.Errorf("%+v: %v", mal, err)
			}
		}
	})

	t.Run("CSV con los filtros, sin fórmulas, y la descarga queda anotada", func(t *testing.T) {
		filas, err := hoy.ExportarActividad(comoVale, b.App, id, hoy.FiltroActividad{Persona: leo.ID})
		must(t, err)
		if len(filas) != 7 || filas[0][2] != "Quién" || filas[1][2] != "Leo" || filas[1][3] != "Ventas" {
			t.Fatalf("csv: %v", filas)
		}
		pag, _ := hoy.Actividad(ctx, b.App, id, yo, hoy.FiltroActividad{Categoria: "datos"})
		if len(pag.Items) != 2 || pag.Items[0].Texto != "bajó la actividad en CSV (6 registros)" {
			t.Fatalf("descarga: %+v", pag.Items)
		}
	})
}

// Una IA conectada queda con su nombre y quién la conectó.
func TestIAConectadaEnLaActividad(t *testing.T) {
	b := pruebadb.Nueva(t)
	id := pruebadb.CrearTienda(t, b.App, "sol")
	ia := auditoria.ConActor(ctx, auditoria.Actor{Tipo: "ia", ID: "8f0c1d3e-0000-4000-8000-000000000002", Nombre: "Carla", Rol: "admin", IA: "Claude"})
	stock := 3
	_, err := listas.Guardar(ia, b.App, id, "", listas.Editable{Nombre: "Termo", Precio: 40000, Stock: &stock})
	must(t, err)
	pag, err := hoy.Actividad(ctx, b.App, id, "", hoy.FiltroActividad{})
	must(t, err)
	if len(pag.Items) != 1 || pag.Items[0].Tipo != "ia" || pag.Items[0].Quien != "Claude (conectada por Carla)" ||
		pag.Items[0].Donde != "IA conectada" || *pag.Items[0].Rol != "Administración" {
		t.Fatalf("ia: %+v", pag.Items)
	}
}

// Una tienda no ve la actividad de otra (RLS), y lo de hace más de un año se purga.
func TestActividadAisladaYPurga(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	sol := pruebadb.CrearTienda(t, b.App, "sol")
	a := auditoria.ConActor(ctx, auditoria.Actor{Tipo: "persona", ID: "8f0c1d3e-0000-4000-8000-000000000003", Nombre: "Ana", Rol: "duenio"})
	stock := 1
	_, err := listas.Guardar(a, b.App, luna, "", listas.Editable{Nombre: "Mate", Precio: 1000, Stock: &stock})
	must(t, err)
	if pag, _ := hoy.Actividad(ctx, b.App, sol, "", hoy.FiltroActividad{}); len(pag.Items) != 0 || len(pag.Personas) != 0 {
		t.Fatalf("sol ve lo de luna: %+v", pag)
	}
	// Paginada: 60 más (en un solo INSERT) son dos páginas, sin repetir ni saltear.
	rs := make([]auditoria.Registro, 60)
	for i := range rs {
		rs[i] = auditoria.Registro{Accion: "producto.stock", Entidad: "producto", Texto: "ajuste " + string(rune('A'+i%26)) + string(rune('a'+i/26))}
	}
	must(t, b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error { return auditoria.Registrar(a, tx, luna, rs...) }))
	p1, err := hoy.Actividad(ctx, b.App, luna, "", hoy.FiltroActividad{})
	must(t, err)
	if len(p1.Items) != hoy.TamPagina || p1.Siguiente == nil {
		t.Fatalf("página 1: %d %v", len(p1.Items), p1.Siguiente)
	}
	p2, err := hoy.Actividad(ctx, b.App, luna, "", hoy.FiltroActividad{Cursor: *p1.Siguiente})
	must(t, err)
	vistos := map[string]bool{}
	for _, x := range append(p1.Items, p2.Items...) {
		vistos[x.ID] = true
	}
	if len(p2.Items) != 11 || p2.Siguiente != nil || len(vistos) != 61 {
		t.Fatalf("página 2: %d %v (distintos %d)", len(p2.Items), p2.Siguiente, len(vistos))
	}
	// Desde sol no se puede anotar en luna.
	err = b.App.EnTienda(ctx, sol, func(tx pgx.Tx) error {
		return auditoria.Registrar(a, tx, luna, auditoria.Registro{Accion: "producto.crear", Entidad: "producto", Texto: "intruso"})
	})
	if err == nil {
		t.Fatal("se pudo anotar en otra tienda")
	}
	// Uno viejo (hace 400 días) y uno reciente: la purga sólo borra el viejo.
	must(t, b.App.EnTienda(ctx, sol, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO auditoria (tienda_id, actor_tipo, accion, entidad, texto, creado_en)
			VALUES ($1, 'persona', 'producto.crear', 'producto', 'viejo', now() - interval '400 days')`, sol)
		return err
	}))
	n, err := auditoria.Purgar(ctx, b.App)
	must(t, err)
	if n != 1 {
		t.Fatalf("purgó %d", n)
	}
	if pag, _ := hoy.Actividad(ctx, b.App, luna, "", hoy.FiltroActividad{Categoria: "producto", Desde: "2020-01-01"}); pag.Siguiente == nil {
		t.Fatalf("la purga borró lo reciente: %+v", pag.Items)
	}
	// La app no puede borrar a mano (sólo la purga).
	err = b.App.EnTienda(ctx, luna, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `DELETE FROM auditoria`)
		return err
	})
	if err == nil {
		t.Fatal("la app pudo borrar la actividad")
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
