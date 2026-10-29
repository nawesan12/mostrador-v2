package turnos_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/turnos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/limite"
)

// Lunes 2 de noviembre de 2026, 8 de la mañana en Argentina.
var lunes8 = time.Date(2026, 11, 2, 8, 0, 0, 0, turnos.Zona)

func a(h, m int) time.Time { return time.Date(2026, 11, 2, h, m, 0, 0, turnos.Zona) }

type barberia struct {
	id          string
	corte       turnos.Servicio
	color       turnos.Servicio
	ana, beto   turnos.Profesional
	horarioLun9 []turnos.Franja
}

// armar una barbería: lunes a sábado de 9 a 13 y de 14 a 19; Ana hace todo, Beto sólo cortes.
func armar(t *testing.T, app *db.Pool, slug string) barberia {
	t.Helper()
	ctx := context.Background()
	b := barberia{id: pruebadb.CrearTienda(t, app, slug)}
	var h []turnos.Franja
	for d := 1; d <= 6; d++ {
		h = append(h, turnos.Franja{Dia: d, Desde: "09:00", Hasta: "13:00"}, turnos.Franja{Dia: d, Desde: "14:00", Hasta: "19:00"})
	}
	if _, err := turnos.GuardarAjustes(ctx, app, b.id, turnos.Ajustes{Activo: true, Sena: 30, AnticipacionMin: 60, DiasVista: 14, Horario: h}); err != nil {
		t.Fatal(err)
	}
	var err error
	if b.corte, err = turnos.GuardarServicio(ctx, app, b.id, "", turnos.Servicio{Nombre: "Corte clásico", Precio: 12000, Duracion: 30, Activo: true}, nil); err != nil {
		t.Fatal(err)
	}
	if b.color, err = turnos.GuardarServicio(ctx, app, b.id, "", turnos.Servicio{Nombre: "Color", Precio: 30000, Duracion: 90, Activo: true}, nil); err != nil {
		t.Fatal(err)
	}
	if b.ana, err = turnos.GuardarProfesional(ctx, app, b.id, "", turnos.Profesional{Nombre: "Ana", Rol: "Colorista", Activo: true}, nil); err != nil {
		t.Fatal(err)
	}
	if b.beto, err = turnos.GuardarProfesional(ctx, app, b.id, "", turnos.Profesional{Nombre: "Beto", Rol: "Barbero",
		Servicios: []string{b.corte.ID}, Activo: true}, nil); err != nil {
		t.Fatal(err)
	}
	return b
}

func reserva(servicio string, profesional *string, inicio time.Time) turnos.Reserva {
	return turnos.Reserva{ServicioID: servicio, ProfesionalID: profesional, Inicio: inicio.Format(time.RFC3339),
		Nombre: "Lucía Pérez", Telefono: "11 5555-0101", Email: "lucia@ejemplo.com"}
}

func agenda(t *testing.T, app *db.Pool, tienda string) []turnos.Turno {
	t.Helper()
	l, err := turnos.Agenda(context.Background(), app, tienda, "2026-11-02", "2026-11-08", "", lunes8)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

func TestPanel(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")

	// Sin fila: los de por defecto, apagado.
	aj, err := turnos.LeerAjustes(ctx, b.App, luna)
	if err != nil || aj.Activo || aj.Sena != 30 || aj.AnticipacionMin != 60 || aj.DiasVista != 14 || aj.Horario == nil {
		t.Fatalf("por defecto: %+v %v", aj, err)
	}
	for _, mal := range []turnos.Ajustes{
		{Sena: 120, DiasVista: 14},
		{DiasVista: 0},
		{DiasVista: 14, Horario: []turnos.Franja{{Dia: 7, Desde: "09:00", Hasta: "10:00"}}},
		{DiasVista: 14, Horario: []turnos.Franja{{Dia: 1, Desde: "10:00", Hasta: "09:00"}}},
		{DiasVista: 14, Horario: []turnos.Franja{{Dia: 1, Desde: "09:10", Hasta: "10:00"}}},
		{DiasVista: 14, Horario: []turnos.Franja{{Dia: 1, Desde: "9:00", Hasta: "10:00"}}},
	} {
		if _, err := turnos.GuardarAjustes(ctx, b.App, luna, mal); !errors.Is(err, turnos.ErrInvalido) {
			t.Errorf("aceptó %+v: %v", mal, err)
		}
	}
	nuevo := turnos.Ajustes{Activo: true, Sena: 50, AnticipacionMin: 120, DiasVista: 7, Horario: []turnos.Franja{{Dia: 2, Desde: "10:00", Hasta: "18:00"}}}
	if _, err := turnos.GuardarAjustes(ctx, b.App, luna, nuevo); err != nil {
		t.Fatal(err)
	}
	nuevo.Sena = 40
	if _, err := turnos.GuardarAjustes(ctx, b.App, luna, nuevo); err != nil {
		t.Fatal(err)
	}
	if aj, _ := turnos.LeerAjustes(ctx, b.App, luna); aj.Sena != 40 || len(aj.Horario) != 1 || aj.Horario[0].Desde != "10:00" {
		t.Fatalf("ajustes guardados: %+v", aj)
	}

	// Servicios: validación, orden al final, editar, borrar.
	if _, err := turnos.GuardarServicio(ctx, b.App, luna, "", turnos.Servicio{Nombre: "Corte", Precio: 1000, Duracion: 20}, nil); !errors.Is(err, turnos.ErrInvalido) {
		t.Fatal("aceptó una duración que no es múltiplo de 15")
	}
	if _, err := turnos.GuardarServicio(ctx, b.App, luna, "", turnos.Servicio{Nombre: "Corte", Precio: 0, Duracion: 30}, nil); !errors.Is(err, turnos.ErrInvalido) {
		t.Fatal("aceptó precio 0")
	}
	promo := "Promo"
	s1, err := turnos.GuardarServicio(ctx, b.App, luna, "", turnos.Servicio{Nombre: " Corte ", Precio: 12000, Duracion: 30, Etiqueta: &promo, Activo: true}, nil)
	if err != nil || s1.Nombre != "Corte" || s1.Orden != 0 || s1.Etiqueta == nil || *s1.Etiqueta != "Promo" {
		t.Fatalf("servicio: %+v %v", s1, err)
	}
	s2, _ := turnos.GuardarServicio(ctx, b.App, luna, "", turnos.Servicio{Nombre: "Barba", Precio: 8000, Duracion: 15, Activo: true}, nil)
	if s2.Orden != 1 {
		t.Fatalf("el segundo servicio no fue al final: %d", s2.Orden)
	}
	s2.Precio, s2.Activo = 9000, false
	if s2, err = turnos.GuardarServicio(ctx, b.App, luna, s2.ID, s2, nil); err != nil || s2.Precio != 9000 || s2.Activo || s2.Orden != 1 {
		t.Fatalf("editar: %+v %v", s2, err)
	}
	if _, err := turnos.GuardarServicio(ctx, b.App, luna, "00000000-0000-0000-0000-000000000000", s2, nil); !errors.Is(err, turnos.ErrNoExiste) {
		t.Fatal("editó uno que no existe")
	}

	// Profesionales: los servicios que no son de la tienda se descartan.
	sol := armar(t, b.App, "sol")
	p, err := turnos.GuardarProfesional(ctx, b.App, luna, "", turnos.Profesional{Nombre: "Caro", Servicios: []string{s1.ID, sol.corte.ID}, Activo: true}, nil)
	if err != nil || len(p.Servicios) != 1 || p.Servicios[0] != s1.ID {
		t.Fatalf("profesional: %+v %v", p, err)
	}
	p.Rol = "Barbera"
	if p, err = turnos.GuardarProfesional(ctx, b.App, luna, p.ID, p, nil); err != nil || p.Rol != "Barbera" {
		t.Fatalf("editar profesional: %+v %v", p, err)
	}
	if l, _ := turnos.Profesionales(ctx, b.App, luna); len(l) != 1 {
		t.Fatalf("luna ve %d profesionales", len(l))
	}
	// Otra tienda no edita ni borra lo ajeno.
	if _, err := turnos.GuardarServicio(ctx, b.App, luna, sol.corte.ID, s1, nil); !errors.Is(err, turnos.ErrNoExiste) {
		t.Fatal("luna editó un servicio de sol")
	}
	if err := turnos.BorrarProfesional(ctx, b.App, luna, sol.ana.ID); !errors.Is(err, turnos.ErrNoExiste) {
		t.Fatal("luna borró una profesional de sol")
	}
	if err := turnos.BorrarServicio(ctx, b.App, luna, s2.ID); err != nil {
		t.Fatal(err)
	}
	if l, _ := turnos.Servicios(ctx, b.App, luna); len(l) != 1 {
		t.Fatalf("quedan %d servicios", len(l))
	}

	// Turno a mano: confirmado, sin seña; superponer pide forzar.
	m := turnos.Manual{ServicioID: sol.corte.ID, ProfesionalID: sol.beto.ID, Inicio: a(20, 0).Format(time.RFC3339), Nombre: "Juan"}
	t1, err := turnos.CrearManual(ctx, b.App, sol.id, m, lunes8)
	if err != nil || t1.Estado != "confirmado" || t1.Origen != "panel" || t1.Sena != 0 || t1.Precio != 12000 || t1.Fin != "2026-11-02T20:30:00-03:00" {
		t.Fatalf("manual: %+v %v", t1, err)
	}
	m.Inicio = a(20, 15).Format(time.RFC3339)
	if _, err := turnos.CrearManual(ctx, b.App, sol.id, m, lunes8); !errors.Is(err, turnos.ErrOcupado) {
		t.Fatalf("superpuso sin forzar: %v", err)
	}
	m.Forzar = true
	if _, err := turnos.CrearManual(ctx, b.App, sol.id, m, lunes8); err != nil {
		t.Fatal(err)
	}
	m.ServicioID = s1.ID // de otra tienda
	if _, err := turnos.CrearManual(ctx, b.App, sol.id, m, lunes8); !errors.Is(err, turnos.ErrInvalido) {
		t.Fatal("dio un turno con un servicio de otra tienda")
	}

	// Estados: atendido, ausente, cancelar y deshacer.
	for _, e := range []string{"atendido", "ausente", "cancelado", "confirmado"} {
		if t1, err = turnos.CambiarEstado(ctx, b.App, sol.id, t1.ID, e, lunes8); err != nil || t1.Estado != e {
			t.Fatalf("estado %s: %+v %v", e, t1, err)
		}
	}
	if _, err := turnos.CambiarEstado(ctx, b.App, sol.id, t1.ID, "perdido", lunes8); !errors.Is(err, turnos.ErrInvalido) {
		t.Fatal("aceptó un estado inventado")
	}
	if _, err := turnos.CambiarEstado(ctx, b.App, luna, t1.ID, "cancelado", lunes8); !errors.Is(err, turnos.ErrNoExiste) {
		t.Fatal("luna canceló un turno de sol")
	}
	if l := agenda(t, b.App, sol.id); len(l) != 2 {
		t.Fatalf("agenda de sol: %d", len(l))
	}
	if l := agenda(t, b.App, luna); len(l) != 0 {
		t.Fatalf("luna ve %d turnos de sol", len(l))
	}
	if l, _ := turnos.Agenda(ctx, b.App, sol.id, "2026-11-02", "2026-11-02", sol.ana.ID, lunes8); len(l) != 0 {
		t.Fatalf("filtro por profesional: %d", len(l))
	}
	if _, err := turnos.Agenda(ctx, b.App, sol.id, "2026-11-02", "2027-11-02", "", lunes8); !errors.Is(err, turnos.ErrInvalido) {
		t.Fatal("aceptó un rango de un año")
	}

	// Devolver la seña de un turno sin seña: no.
	if _, err := turnos.DevolverSena(ctx, b.App, &cobradorFalso{}, sol.id, t1.ID, lunes8); !errors.Is(err, turnos.ErrSinSena) {
		t.Fatalf("devolvió una seña que no existe: %v", err)
	}
}

func TestReservaPublica(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := armar(t, b.App, "luna")
	sol := armar(t, b.App, "sol")

	// Lo público: servicios y profesionales activos; sin Mercado Pago, la seña es 0.
	r, err := turnos.Publico(ctx, b.App, nil, "luna")
	if err != nil || len(r.Servicios) != 2 || len(r.Profesionales) != 2 || r.Sena != 0 || len(r.Horario) != 12 {
		t.Fatalf("público: %+v %v", r, err)
	}
	if vt, err := vidriera.Publica(ctx, b.App, "luna"); err != nil || vt.Turnos == nil || len(vt.Turnos.Servicios) != 2 {
		t.Fatalf("la tienda pública no trae los turnos: %+v %v", vt.Turnos, err)
	}

	// Libres: la anticipación (60 min) deja afuera las 9; Beto no hace color.
	l, err := turnos.LibresDe(ctx, b.App, "luna", luna.corte.ID, luna.beto.ID, "2026-11-02", lunes8)
	if err != nil || len(l) == 0 || l[0].Inicio != "2026-11-02T09:00:00-03:00" {
		t.Fatalf("libres: %+v %v", l, err)
	}
	if l, _ := turnos.LibresDe(ctx, b.App, "luna", luna.color.ID, luna.beto.ID, "2026-11-02", lunes8); len(l) != 0 {
		t.Fatalf("Beto no hace color y tiene %d horarios", len(l))
	}
	if _, err := turnos.LibresDe(ctx, b.App, "luna", "no-es-un-id", "", "2026-11-02", lunes8); !errors.Is(err, turnos.ErrInvalido) {
		t.Fatal("aceptó un servicio inválido")
	}

	// Reservar: el precio sale del servidor, sin Mercado Pago queda confirmado y sin seña.
	res, err := turnos.Reservar(ctx, b.App, nil, "https://tulocalweb.com.ar", "luna", reserva(luna.corte.ID, &luna.beto.ID, a(10, 0)), lunes8)
	if err != nil || len(res.Codigo) < 16 || res.URL != "" {
		t.Fatalf("reservar: %+v %v", res, err)
	}
	v, err := turnos.Ver(ctx, b.App, "luna", res.Codigo, lunes8)
	if err != nil || v.Estado != "confirmado" || v.Precio != 12000 || v.Sena != 0 || v.Profesional != "Beto" || v.Nombre != "Lucía" ||
		v.Inicio != "2026-11-02T10:00:00-03:00" || v.Fin != "2026-11-02T10:30:00-03:00" || !v.Cancelable {
		t.Fatalf("ver: %+v %v", v, err)
	}
	// El mismo horario ya no está, y no se puede reservar dos veces.
	if l, _ := turnos.LibresDe(ctx, b.App, "luna", luna.corte.ID, luna.beto.ID, "2026-11-02", lunes8); contiene(l, "2026-11-02T10:00:00-03:00") {
		t.Fatal("el horario reservado sigue libre")
	}
	if _, err := turnos.Reservar(ctx, b.App, nil, "", "luna", reserva(luna.corte.ID, &luna.beto.ID, a(10, 15)), lunes8); !errors.Is(err, turnos.ErrOcupado) {
		t.Fatalf("superpuso: %v", err)
	}
	// Con "cualquiera", a las 10 atiende Ana (Beto está ocupado).
	res2, err := turnos.Reservar(ctx, b.App, nil, "", "luna", reserva(luna.corte.ID, nil, a(10, 0)), lunes8)
	if err != nil {
		t.Fatal(err)
	}
	if v, _ := turnos.Ver(ctx, b.App, "luna", res2.Codigo, lunes8); v.Profesional != "Ana" {
		t.Fatalf("cualquiera asignó a %q", v.Profesional)
	}
	// Fuera de la grilla, del horario, de la anticipación o con datos flojos: no.
	for nombre, rv := range map[string]turnos.Reserva{
		"fuera de la grilla":   reserva(luna.corte.ID, &luna.ana.ID, a(11, 10)),
		"en el corte del día":  reserva(luna.corte.ID, &luna.ana.ID, a(13, 0)),
		"antes de anticipar":   reserva(luna.corte.ID, &luna.ana.ID, a(8, 30)),
		"un domingo":           reserva(luna.corte.ID, &luna.ana.ID, time.Date(2026, 11, 8, 10, 0, 0, 0, turnos.Zona)),
		"después de la vista":  reserva(luna.corte.ID, &luna.ana.ID, time.Date(2026, 11, 17, 10, 0, 0, 0, turnos.Zona)),
		"color no entra al 13": reserva(luna.color.ID, &luna.ana.ID, a(12, 0)),
	} {
		if _, err := turnos.Reservar(ctx, b.App, nil, "", "luna", rv, lunes8); !errors.Is(err, turnos.ErrOcupado) {
			t.Errorf("%s: %v", nombre, err)
		}
	}
	sinMail := reserva(luna.corte.ID, &luna.ana.ID, a(11, 0))
	sinMail.Email = "lucia"
	if _, err := turnos.Reservar(ctx, b.App, nil, "", "luna", sinMail, lunes8); !errors.Is(err, turnos.ErrInvalido) {
		t.Errorf("sin email válido: %v", err)
	}
	// Un servicio o profesional de otra tienda no sirve.
	if _, err := turnos.Reservar(ctx, b.App, nil, "", "luna", reserva(sol.corte.ID, nil, a(11, 0)), lunes8); !errors.Is(err, turnos.ErrInvalido) {
		t.Errorf("servicio de otra tienda: %v", err)
	}
	if _, err := turnos.Reservar(ctx, b.App, nil, "", "luna", reserva(luna.corte.ID, &sol.ana.ID, a(11, 0)), lunes8); !errors.Is(err, turnos.ErrOcupado) {
		t.Errorf("profesional de otra tienda: %v", err)
	}

	// Aislamiento: otra tienda no ve ni cancela el turno.
	if _, err := turnos.Ver(ctx, b.App, "sol", res.Codigo, lunes8); !errors.Is(err, turnos.ErrNoExiste) {
		t.Fatal("sol vio el turno de luna")
	}
	if _, err := turnos.Cancelar(ctx, b.App, "sol", res.Codigo, lunes8); !errors.Is(err, turnos.ErrNoExiste) {
		t.Fatal("sol canceló el turno de luna")
	}
	if _, err := turnos.Ver(ctx, b.App, "luna", "codigo-que-no-existe-0000", lunes8); !errors.Is(err, turnos.ErrNoExiste) {
		t.Fatal("vio un código inventado")
	}

	// Cancelar por código: hasta la anticipación; libera el horario; dos veces no es error.
	if _, err := turnos.Cancelar(ctx, b.App, "luna", res.Codigo, a(9, 15)); !errors.Is(err, turnos.ErrTarde) {
		t.Fatalf("canceló a 45 minutos: %v", err)
	}
	c, err := turnos.Cancelar(ctx, b.App, "luna", res.Codigo, lunes8)
	if err != nil || c.Estado != "cancelado" || c.Cancelable {
		t.Fatalf("cancelar: %+v %v", c, err)
	}
	if _, err := turnos.Cancelar(ctx, b.App, "luna", res.Codigo, lunes8); err != nil {
		t.Fatalf("cancelar dos veces: %v", err)
	}
	if l, _ := turnos.LibresDe(ctx, b.App, "luna", luna.corte.ID, luna.beto.ID, "2026-11-02", lunes8); !contiene(l, "2026-11-02T10:00:00-03:00") {
		t.Fatal("cancelar no liberó el horario")
	}
	// El aviso a la tienda sale por el outbox.
	var avisos int
	_ = b.App.EnTienda(ctx, luna.id, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT count(*) FROM outbox WHERE tipo = 'turno.nuevo'`).Scan(&avisos)
	})
	if avisos != 2 {
		t.Fatalf("avisos de turno nuevo: %d", avisos)
	}

	// Apagado: la tienda no toma turnos y la tienda pública lo dice con null.
	aj, _ := turnos.LeerAjustes(ctx, b.App, sol.id)
	aj.Activo = false
	_, _ = turnos.GuardarAjustes(ctx, b.App, sol.id, aj)
	if _, err := turnos.Publico(ctx, b.App, nil, "sol"); !errors.Is(err, turnos.ErrApagado) {
		t.Fatal("una tienda apagada muestra turnos")
	}
	if _, err := turnos.Reservar(ctx, b.App, nil, "", "sol", reserva(sol.corte.ID, nil, a(11, 0)), lunes8); !errors.Is(err, turnos.ErrApagado) {
		t.Fatal("una tienda apagada tomó un turno")
	}
	if vt, _ := vidriera.Publica(ctx, b.App, "sol"); vt.Turnos != nil {
		t.Fatal("la tienda pública trae turnos estando apagado")
	}
}

func contiene(l []turnos.HorarioLibre, inicio string) bool {
	for _, h := range l {
		if h.Inicio == inicio {
			return true
		}
	}
	return false
}

// Dos (o diez) personas eligen el mismo horario a la vez: lo toma una sola.
func TestReservaConcurrente(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := armar(t, b.App, "luna")

	correr := func(n int, profesional *string, inicio time.Time) (ganaron, ocupado int) {
		var wg sync.WaitGroup
		var mu sync.Mutex
		largada := make(chan struct{})
		for i := 0; i < n; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-largada
				_, err := turnos.Reservar(ctx, b.App, nil, "", "luna", reserva(luna.corte.ID, profesional, inicio), lunes8)
				mu.Lock()
				defer mu.Unlock()
				switch {
				case err == nil:
					ganaron++
				case errors.Is(err, turnos.ErrOcupado):
					ocupado++
				default:
					t.Errorf("error inesperado: %v", err)
				}
			}()
		}
		close(largada)
		wg.Wait()
		return
	}
	if g, o := correr(10, &luna.beto.ID, a(15, 0)); g != 1 || o != 9 {
		t.Fatalf("mismo profesional: ganaron %d, ocupado %d", g, o)
	}
	// Con "cualquiera" y dos profesionales libres, ganan exactamente dos.
	if g, o := correr(8, nil, a(16, 0)); g != 2 || o != 6 {
		t.Fatalf("cualquiera: ganaron %d, ocupado %d", g, o)
	}
	var superpuestos int
	_ = b.App.EnTienda(ctx, luna.id, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `SELECT count(*) FROM turnos x JOIN turnos y ON x.id < y.id AND x.profesional_id = y.profesional_id
			AND x.inicio < y.fin AND x.fin > y.inicio WHERE x.estado <> 'cancelado' AND y.estado <> 'cancelado'`).Scan(&superpuestos)
	})
	if superpuestos != 0 {
		t.Fatalf("%d pares de turnos superpuestos", superpuestos)
	}
}

// cobradorFalso hace de Mercado Pago para la seña (el de verdad se prueba en el paquete mercadopago).
type cobradorFalso struct {
	mu        sync.Mutex
	falla     bool
	pagados   []string
	devueltos []string
}

func (c *cobradorFalso) PagarTurno(_ context.Context, _, turnoID, vuelta string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.falla {
		return "", errors.New("mercado pago caído")
	}
	c.pagados = append(c.pagados, turnoID)
	return "https://www.mercadopago.com.ar/checkout/v1/redirect?pref_id=x&vuelta=" + vuelta, nil
}

func (c *cobradorFalso) DevolverSena(_ context.Context, _, turnoID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.devueltos = append(c.devueltos, turnoID)
	return nil
}

func conectarMP(t *testing.T, app *db.Pool, tienda string) {
	t.Helper()
	ctx := context.Background()
	if err := app.EnTienda(ctx, tienda, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO mp_conexiones (tienda_id, mp_usuario, acceso, renovacion, vence_en)
			VALUES ($1, 1000 + (random() * 1e6)::int, '\x00', '\x00', now() + interval '1 day')`, tienda)
		return err
	}); err != nil {
		t.Fatal(err)
	}
}

func TestReservaConSena(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := armar(t, b.App, "luna")
	conectarMP(t, b.App, luna.id)
	mp := &cobradorFalso{}

	if r, _ := turnos.Publico(ctx, b.App, mp, "luna"); r.Sena != 30 {
		t.Fatalf("con Mercado Pago la seña es %d", r.Sena)
	}
	res, err := turnos.Reservar(ctx, b.App, mp, "https://tulocalweb.com.ar", "luna", reserva(luna.color.ID, &luna.ana.ID, a(10, 0)), lunes8)
	if err != nil || !strings.Contains(res.URL, "vuelta=https://tulocalweb.com.ar/t/luna/turno/"+res.Codigo) || len(mp.pagados) != 1 {
		t.Fatalf("reservar con seña: %+v %v", res, err)
	}
	v, _ := turnos.Ver(ctx, b.App, "luna", res.Codigo, lunes8)
	if v.Estado != "reservado" || v.Sena != 9000 || v.SenaPagada || v.VenceEn == nil || !v.VenceEn.Equal(lunes8.Add(turnos.Espera)) {
		t.Fatalf("turno esperando la seña: %+v", v)
	}
	// Mientras espera, ocupa el horario; vencido, se libera solo (sin cron).
	if _, err := turnos.Reservar(ctx, b.App, mp, "", "luna", reserva(luna.corte.ID, &luna.ana.ID, a(10, 30)), lunes8.Add(10*time.Minute)); !errors.Is(err, turnos.ErrOcupado) {
		t.Fatalf("tomó un horario que espera la seña: %v", err)
	}
	despues := lunes8.Add(16 * time.Minute)
	if v, _ := turnos.Ver(ctx, b.App, "luna", res.Codigo, despues); !v.Vencido || v.Cancelable {
		t.Fatalf("vencido: %+v", v)
	}
	res2, err := turnos.Reservar(ctx, b.App, mp, "", "luna", reserva(luna.corte.ID, &luna.ana.ID, a(10, 30)), despues)
	if err != nil {
		t.Fatalf("el horario vencido no se liberó: %v", err)
	}

	// La seña pagada (lo que hace el webhook): confirmado, y devolverla pasa por Mercado Pago.
	var id2 string
	_ = b.App.EnTienda(ctx, luna.id, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, `UPDATE turnos SET sena_pagada = true, estado = 'confirmado', vence_en = NULL
			WHERE codigo_publico = $1 RETURNING id::text`, res2.Codigo).Scan(&id2)
	})
	if _, err := turnos.DevolverSena(ctx, b.App, nil, luna.id, id2, despues); !errors.Is(err, turnos.ErrSinCobros) {
		t.Fatalf("devolver sin Mercado Pago: %v", err)
	}
	if _, err := turnos.DevolverSena(ctx, b.App, mp, luna.id, id2, despues); err != nil || len(mp.devueltos) != 1 || mp.devueltos[0] != id2 {
		t.Fatalf("devolver: %v %v", err, mp.devueltos)
	}

	// Si Mercado Pago no responde, el turno no queda tomando el horario.
	mp.falla = true
	if _, err := turnos.Reservar(ctx, b.App, mp, "", "luna", reserva(luna.corte.ID, &luna.beto.ID, a(11, 0)), lunes8); !errors.Is(err, turnos.ErrMercadoPago) {
		t.Fatalf("con Mercado Pago caído: %v", err)
	}
	mp.falla = false
	if _, err := turnos.Reservar(ctx, b.App, mp, "", "luna", reserva(luna.corte.ID, &luna.beto.ID, a(11, 0)), lunes8); err != nil {
		t.Fatalf("el horario quedó tomado después de la falla: %v", err)
	}

	// Seña 0: confirmado directo aunque haya Mercado Pago.
	aj, _ := turnos.LeerAjustes(ctx, b.App, luna.id)
	aj.Sena = 0
	_, _ = turnos.GuardarAjustes(ctx, b.App, luna.id, aj)
	res3, err := turnos.Reservar(ctx, b.App, mp, "", "luna", reserva(luna.corte.ID, &luna.beto.ID, a(12, 0)), lunes8)
	if v, _ := turnos.Ver(ctx, b.App, "luna", res3.Codigo, lunes8); err != nil || res3.URL != "" || v.Estado != "confirmado" || v.Sena != 0 {
		t.Fatalf("seña 0: %+v %+v %v", res3, v, err)
	}
}

func TestRutas(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := armar(t, b.App, "luna")
	mux := http.NewServeMux()
	turnos.Rutas(mux, b.App, turnos.Config{
		Tienda: func(r *http.Request) string {
			if r.Header.Get("X-Sesion") == "ok" {
				return luna.id
			}
			return ""
		},
		Admin: func(r *http.Request) bool { return r.Header.Get("X-Admin") == "si" },
		Ahora: func() time.Time { return lunes8 },
	})
	pedir := func(metodo, ruta, cuerpo string, sesion bool) *httptest.ResponseRecorder {
		r := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
		if sesion {
			r.Header.Set("X-Sesion", "ok")
		}
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w
	}
	if w := pedir("GET", "/v1/panel/turnos/servicios", "", false); w.Code != http.StatusUnauthorized {
		t.Fatalf("sin sesión: %d", w.Code)
	}
	w := pedir("POST", "/v1/panel/turnos/servicios", `{"nombre":"Barba","precio":8000,"duracion":15}`, true)
	var s turnos.Servicio
	if w.Code != http.StatusCreated || json.Unmarshal(w.Body.Bytes(), &s) != nil || !s.Activo || s.Orden != 2 {
		t.Fatalf("crear servicio: %d %s", w.Code, w.Body)
	}
	if w := pedir("DELETE", "/v1/panel/turnos/servicios/"+s.ID, "", true); w.Code != http.StatusNoContent {
		t.Fatalf("borrar: %d", w.Code)
	}
	if w := pedir("PUT", "/v1/panel/turnos/ajustes", `{"activo":true,"sena":200,"diasVista":14}`, true); w.Code != http.StatusBadRequest &&
		w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("ajustes inválidos: %d", w.Code)
	}
	if w := pedir("GET", "/v1/publico/tiendas/luna/turnos/libres?servicio="+luna.corte.ID+"&profesional=cualquiera&dia=2026-11-02", "", false); w.Code != http.StatusOK ||
		!strings.HasPrefix(w.Body.String(), `[{"inicio":"2026-11-02T09:00:00-03:00","profesionalId":null}`) {
		t.Fatalf("libres: %d %s", w.Code, w.Body)
	}
	w = pedir("POST", "/v1/publico/tiendas/luna/turnos", `{"servicioId":"`+luna.corte.ID+`","profesionalId":null,"inicio":"2026-11-02T09:00:00-03:00","nombre":"Lucía","telefono":"1155550101","email":"lucia@ejemplo.com"}`, false)
	var res turnos.Reservado
	if w.Code != http.StatusCreated || json.Unmarshal(w.Body.Bytes(), &res) != nil || res.Codigo == "" {
		t.Fatalf("reservar: %d %s", w.Code, w.Body)
	}
	if w := pedir("POST", "/v1/publico/tiendas/luna/turnos", `{"servicioId":"`+luna.corte.ID+`","profesionalId":"`+luna.ana.ID+`","inicio":"2026-11-02T09:00:00-03:00","nombre":"Otra","telefono":"1155550101","email":"otra@ejemplo.com"}`, false); w.Code != http.StatusConflict {
		t.Fatalf("horario tomado: %d %s", w.Code, w.Body)
	}
	if w := pedir("GET", "/v1/publico/tiendas/luna/turnos/"+res.Codigo, "", false); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"estado":"confirmado"`) {
		t.Fatalf("ver: %d %s", w.Code, w.Body)
	}
	l := agenda(t, b.App, luna.id)
	if w := pedir("PATCH", "/v1/panel/turnos/"+l[0].ID, `{"estado":"confirmado"}`, true); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"estado":"confirmado"`) {
		t.Fatalf("estado: %d %s", w.Code, w.Body)
	}
	if w := pedir("POST", "/v1/panel/turnos/"+l[0].ID+"/devolver-sena", "", true); w.Code != http.StatusForbidden {
		t.Fatalf("devolver sin ser admin: %d", w.Code)
	}
	if w := pedir("GET", "/v1/panel/turnos?desde=2026-11-02", "", true); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"senaPagada":false`) {
		t.Fatalf("agenda: %d %s", w.Code, w.Body)
	}
	if w := pedir("DELETE", "/v1/publico/tiendas/luna/turnos/"+res.Codigo, "", false); w.Code != http.StatusOK {
		t.Fatalf("cancelar: %d %s", w.Code, w.Body)
	}
	if w := pedir("GET", "/v1/publico/tiendas/no-existe/turnos", "", false); w.Code != http.StatusNotFound {
		t.Fatalf("tienda inexistente: %d", w.Code)
	}
}

// Tope de reservas públicas: 3 por teléfono y 10 por IP en una hora (el horario ocupado no cuenta
// distinto: cada intento suma).
func TestTopeDeReservas(t *testing.T) {
	b := pruebadb.Nueva(t)
	luna := armar(t, b.App, "luna")
	mux := http.NewServeMux()
	turnos.Rutas(mux, b.App, turnos.Config{
		Ahora: func() time.Time { return lunes8 },
		IP:    func(r *http.Request) string { return r.Header.Get("X-IP") },
		Permitido: func(ctx context.Context, clave string, max int, ventana time.Duration) (bool, error) {
			return limite.Permitido(ctx, b.App, clave, max, ventana)
		},
	})
	reservar := func(hora, tel, ip string) int {
		r := httptest.NewRequest("POST", "/v1/publico/tiendas/luna/turnos", strings.NewReader(`{"servicioId":"`+luna.corte.ID+
			`","profesionalId":null,"inicio":"2026-11-02T`+hora+`:00-03:00","nombre":"Lucía","telefono":"`+tel+`","email":"lucia@ejemplo.com"}`))
		r.Header.Set("X-IP", ip)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)
		return w.Code
	}
	for _, h := range []string{"09:00", "09:30", "10:00"} {
		if c := reservar(h, "1155550101", "190.0.0.1"); c != http.StatusCreated {
			t.Fatalf("reserva %s: %d", h, c)
		}
	}
	if c := reservar("10:30", "1155550101", "190.0.0.2"); c != http.StatusTooManyRequests {
		t.Fatalf("cuarta con el mismo teléfono: %d", c)
	}
	for i := range 7 {
		reservar("11:00", "115555021"+strconv.Itoa(i), "190.0.0.1")
	}
	if c := reservar("12:00", "1155559999", "190.0.0.1"); c != http.StatusTooManyRequests {
		t.Fatalf("once desde la misma IP: %d", c)
	}
	if c := reservar("12:00", "1155559999", "190.0.0.3"); c != http.StatusCreated {
		t.Fatalf("otra IP y otro teléfono: %d", c)
	}
}
