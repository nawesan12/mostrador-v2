// Package tincho: el asistente de Tu Local Web con Claude Haiku (docs-ia/asistente-ia.md). El motor
// de reglas del BFF contesta primero; lo que no resuelve llega acá. Tincho habla sólo con la dueña y
// su equipo (nunca con clientes de la tienda), lee con herramientas que reusan las consultas del
// panel y, para hacer algo (cupón, estado, stock, link de pago), devuelve una propuesta que la
// persona confirma en la web o la app: nunca ejecuta.
package tincho

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/abono"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/planes"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/roles"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/ia"
)

// Lo que prometen los planes (dominio/planes y apps/web/src/lib/panel/planes.ts): respuestas con IA
// por mes. Cada pack extra (pago, $ 4.900) suma 2.000.
var PorPlan = func() map[string]int {
	m := map[string]int{}
	for _, p := range planes.Todos {
		m[p.ID] = p.Tincho
	}
	return m
}()

const PorPack = planes.PackTincho

// Presupuesto de la plataforma por defecto: la parte de Tu Local Web de los USD 100 (40).
const PresupuestoUSD = 40

// Tiempo total de una pregunta (hasta 4 llamadas de 25 s, pero no más que esto).
const tiempoPregunta = 60 * time.Second

var ErrNoDisponible = ia.ErrNoDisponible

// Servicio de Tincho. Modelo nil: sin clave, sólo responde el motor (ErrNoDisponible).
type Servicio struct {
	Pool   *db.Pool
	Modelo ia.Modelo
	Cache  *ia.Cache
	// Presupuesto del mes en micro-dólares (0: sin tope global).
	Presupuesto int64
	Ahora       func() time.Time
	// RolDelChat resuelve el chat de Telegram vinculado (integraciones/telegram): la tienda y el rol de
	// quien lo vinculó (si ya no está en el equipo, da error y el vínculo se borra).
	RolDelChat func(ctx context.Context, chat int64) (string, string, error)
	Log        *slog.Logger

	mu    sync.Mutex
	datos map[string]datosGuardados
}

func (s *Servicio) ahora() time.Time {
	if s.Ahora != nil {
		return s.Ahora()
	}
	return time.Now()
}

func (s *Servicio) log() *slog.Logger {
	if s.Log != nil {
		return s.Log
	}
	return slog.Default()
}

// Consulta es una pregunta de alguien del equipo.
type Consulta struct {
	Tienda, Rol string
	// "panel" (web y apps) o "telegram".
	Canal     string
	Pregunta  string
	Historial []ia.Turno
}

// Propuesta es una acción que pidió Tincho y que la persona tiene que confirmar. Herramienta y
// Argumentos son los de las herramientas del MCP del BFF, que es quien la ejecuta al confirmar.
type Propuesta struct {
	// ID: la clave para confirmarla una sola vez (Idempotency-Key en /panel/hornero/confirmar).
	ID          string         `json:"id"`
	Herramienta string         `json:"herramienta"`
	Argumentos  map[string]any `json:"argumentos"`
	Titulo      string         `json:"titulo"`
	Detalle     string         `json:"detalle,omitempty"`
}

// UsoPlan: cuántas respuestas con IA lleva la tienda en el mes y cuántas tiene.
type UsoPlan struct {
	Usadas int `json:"usadas"`
	Tope   int `json:"tope"`
}

// Respuesta de Tincho.
type Respuesta struct {
	Texto      string      `json:"texto"`
	Propuestas []Propuesta `json:"propuestas"`
	// ia, cache o tope.
	Fuente string `json:"fuente"`
	// Por qué no respondió la IA: plan, dia o plataforma (vacío si respondió).
	Tope string   `json:"tope,omitempty"`
	Uso  *UsoPlan `json:"uso,omitempty"`
}

// MensajeDeTope: lo que se le dice a la persona cuando la IA no responde por un tope. El motor
// sigue contestando lo de todos los días.
func MensajeDeTope(c ia.Corte, tope int) string {
	switch c {
	case ia.CortePlan:
		return fmt.Sprintf("Ya usaste las %s respuestas con IA de tu plan este mes. Lo de todos los días (ventas, pedidos, stock, clientes) te lo sigo respondiendo; para preguntas abiertas, sumá un pack o cambiá de plan en Plan y facturación.", miles(tope))
	case ia.CorteDia:
		return "Por hoy ya usaste las respuestas con IA del día. Mañana sigo; mientras, te respondo lo de todos los días: ventas, pedidos, stock y clientes."
	case ia.CorteOcupado:
		return "Estoy terminando de responderte otra pregunta. Probá de nuevo en un momento."
	case ia.CortePlataforma:
		return "Las respuestas con IA están en pausa por este mes. Lo de todos los días (ventas, pedidos, stock, clientes) te lo sigo respondiendo con tus datos."
	}
	return ""
}

func miles(n int) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "." + s[i:]
	}
	return s
}

// TopeDelMes: el del plan vigente (en la prueba gratis, Gratis usa Crece) más los packs del mes.
func TopeDelMes(plan string, enPrueba bool, packs int) int {
	if plan == "gratis" && enPrueba {
		plan = planes.DePrueba
	}
	n, ok := PorPlan[plan]
	if !ok {
		n = PorPlan["gratis"]
	}
	return n + packs*PorPack
}

// TopeDelDia frena abusos: un décimo del mes, nunca menos de 25.
func TopeDelDia(mes int) int { return max(25, mes/10) }

// Topes de la tienda según su plan.
func (s *Servicio) Topes(ctx context.Context, tiendaID string) (ia.Topes, error) {
	ahora := s.ahora()
	var t ia.Topes
	err := s.Pool.EnTienda(ctx, tiendaID, func(tx pgx.Tx) error {
		var err error
		t, err = TopesEnTx(ctx, tx, tiendaID, ahora)
		return err
	})
	t.Presupuesto = s.Presupuesto
	return t, err
}

// TopesEnTx lee el plan y los packs dentro de una transacción de la tienda (también la usa la
// pantalla del plan).
func TopesEnTx(ctx context.Context, tx pgx.Tx, tiendaID string, ahora time.Time) (ia.Topes, error) {
	e, err := abono.Leer(ctx, tx, tiendaID)
	if err != nil {
		return ia.Topes{}, err
	}
	var packs int
	if err := tx.QueryRow(ctx, `SELECT coalesce((packs->>$2)::int, 0) FROM tiendas WHERE id = $1`, tiendaID, ia.Mes(ahora)).Scan(&packs); err != nil {
		return ia.Topes{}, err
	}
	prueba := ahora.Before(e.FinPrueba)
	mes := TopeDelMes(e.Plan, prueba, packs)
	dia := TopeDelDia(mes)
	// En la prueba sin pagar se usa Pro, pero con un tope diario más bajo: la experiencia es la de Pro
	// y unas decenas de cuentas de prueba no se comen el presupuesto de la plataforma.
	if e.Plan == "gratis" && prueba {
		dia = min(dia, TopeDiaPrueba)
	}
	return ia.Topes{Mes: mes, Dia: dia}, nil
}

// TopeDiaPrueba: respuestas con IA por día durante la prueba gratis.
const TopeDiaPrueba = 30

// revisar aplica los topes y reserva el lugar (hay que liberarlo con s.liberar). Devuelve la
// respuesta de tope (si corta) y el uso del mes.
func (s *Servicio) revisar(ctx context.Context, tiendaID string) (*Respuesta, UsoPlan, error) {
	topes, err := s.Topes(ctx, tiendaID)
	if err != nil {
		return nil, UsoPlan{}, err
	}
	corte, c, err := ia.Reservar(ctx, s.Pool, tiendaID, topes, s.ahora())
	if err != nil {
		return nil, UsoPlan{}, err
	}
	uso := UsoPlan{Usadas: c.Mes, Tope: topes.Mes}
	if corte != ia.SinCorte {
		return &Respuesta{Texto: MensajeDeTope(corte, topes.Mes), Propuestas: []Propuesta{}, Fuente: "tope",
			Tope: string(corte), Uso: &uso}, uso, nil
	}
	return nil, uso, nil
}

// liberar devuelve el lugar que reservó revisar.
func (s *Servicio) liberar(ctx context.Context, tiendaID string) {
	if err := ia.Liberar(context.WithoutCancel(ctx), s.Pool, tiendaID); err != nil {
		s.log().Error("tincho: liberar el cupo en vuelo", "error", err)
	}
}

func rutaDe(canal string) string {
	if canal == "telegram" {
		return ia.RutaTelegram
	}
	return ia.RutaChat
}

// Preguntar responde una pregunta que el motor no resolvió. ErrNoDisponible: la IA no está o no
// respondió (quien llama contesta con el motor).
func (s *Servicio) Preguntar(ctx context.Context, q Consulta) (Respuesta, error) {
	q.Pregunta = strings.TrimSpace(q.Pregunta)
	corte, uso, err := s.revisar(ctx, q.Tienda)
	if err != nil {
		return Respuesta{}, err
	}
	if corte != nil {
		return *corte, nil
	}
	defer s.liberar(ctx, q.Tienda)
	montos := roles.Puede(q.Rol, roles.Finanzas)
	hist, _ := json.Marshal(ia.Recortar(q.Historial))
	clave := ia.Clave(q.Tienda, q.Rol, q.Canal, string(hist), ia.Normalizar(q.Pregunta))
	if s.Cache != nil {
		if t, ok := s.Cache.Leer(clave); ok {
			return Respuesta{Texto: t, Propuestas: []Propuesta{}, Fuente: "cache", Uso: &uso}, nil
		}
	}
	if s.Modelo == nil {
		return Respuesta{}, ErrNoDisponible
	}
	datos, err := s.datosDe(ctx, q.Tienda, montos)
	if err != nil {
		return Respuesta{}, err
	}
	pregunta := q.Pregunta
	if q.Canal == "telegram" {
		pregunta = "[telegram] " + pregunta
	}
	params := ia.Armar(ia.Pedido{Sistema: Sistema, Herramientas: Herramientas(), Datos: datos,
		Historial: q.Historial, Pregunta: pregunta, MaxTokens: ia.TokensChat})

	h := &herramientas{s: s, tienda: q.Tienda, rol: q.Rol, ahora: s.ahora()}
	ctx, cancel := context.WithTimeout(ctx, tiempoPregunta)
	defer cancel()
	r, err := ia.Conversar(ctx, s.Modelo, params, h.ejecutar)
	if r.Uso != (ia.Uso{}) {
		// Lo que se gastó se cuenta aunque haya fallado una vuelta posterior.
		if e := ia.Registrar(context.WithoutCancel(ctx), s.Pool, q.Tienda, rutaDe(q.Canal), r.Uso, s.ahora()); e != nil {
			s.log().Error("tincho: registrar uso", "error", e)
		}
		s.log().Info("tincho", "tienda", q.Tienda, "vueltas", r.Vueltas, "herramientas", r.Herramientas,
			"entrada", r.Uso.Entrada, "cacheLeido", r.Uso.CacheLeido, "cacheEscrito", r.Uso.CacheEscrito, "salida", r.Uso.Salida)
	}
	if err != nil {
		if !errors.Is(err, context.Canceled) {
			s.log().Warn("tincho: la IA no respondió", "error", err)
		}
		return Respuesta{}, ErrNoDisponible
	}
	uso.Usadas++
	res := Respuesta{Texto: r.Texto, Propuestas: h.propuestas, Fuente: "ia", Uso: &uso}
	if res.Propuestas == nil {
		res.Propuestas = []Propuesta{}
	}
	// Sólo se guardan las respuestas sin acciones (y no los rechazos ni los cortes).
	if s.Cache != nil && len(h.propuestas) == 0 && !r.Rechazo && r.Texto != ia.TextoSinVueltas {
		s.Cache.Guardar(clave, r.Texto)
	}
	return res, nil
}

// SistemaMejorar: «Mejorar con Tincho» (fijo, para que el caché lo comparta toda la plataforma).
const SistemaMejorar = `Escribís descripciones de productos para tiendas online argentinas de Tu Local Web.
Español rioplatense, claro y vendedor sin exagerar. Entre 2 y 4 oraciones, sin títulos, sin listas, sin emojis.
No inventes medidas, materiales, envíos, garantías ni datos que no estén en lo que te pasan.
Respondé sólo con la descripción nueva, sin comillas ni comentarios.`

// Mejorar reescribe la descripción de un producto (cuenta en el mismo cupo que la charla).
func (s *Servicio) Mejorar(ctx context.Context, tiendaID, tienda, rubro, nombre, descripcion string) (Respuesta, error) {
	corte, uso, err := s.revisar(ctx, tiendaID)
	if err != nil {
		return Respuesta{}, err
	}
	if corte != nil {
		return *corte, nil
	}
	defer s.liberar(ctx, tiendaID)
	if s.Modelo == nil {
		return Respuesta{}, ErrNoDisponible
	}
	if strings.TrimSpace(descripcion) == "" {
		descripcion = "(vacía)"
	}
	params := ia.Armar(ia.Pedido{Sistema: SistemaMejorar,
		Datos:    fmt.Sprintf("Tienda: %s (%s).", tienda, rubro),
		Pregunta: fmt.Sprintf("Producto: %s\nDescripción actual: %s", nombre, descripcion), MaxTokens: ia.TokensTextos})
	ctx, cancel := context.WithTimeout(ctx, tiempoPregunta)
	defer cancel()
	r, err := ia.Conversar(ctx, s.Modelo, params, nil)
	if r.Uso != (ia.Uso{}) {
		if e := ia.Registrar(context.WithoutCancel(ctx), s.Pool, tiendaID, ia.RutaMejorar, r.Uso, s.ahora()); e != nil {
			s.log().Error("tincho: registrar uso", "error", e)
		}
	}
	if err != nil || r.Rechazo || r.Texto == ia.TextoVacio {
		return Respuesta{}, ErrNoDisponible
	}
	uso.Usadas++
	return Respuesta{Texto: strings.Trim(r.Texto, "\"« »\n"), Propuestas: []Propuesta{}, Fuente: "ia", Uso: &uso}, nil
}

// ContarMotor suma una respuesta del motor (sin IA): es lo que muestra «Respuestas automáticas sin
// IA» en Plan y facturación.
func (s *Servicio) ContarMotor(ctx context.Context, tiendaID string) error {
	return ia.Registrar(ctx, s.Pool, tiendaID, ia.RutaMotor, ia.Uso{}, s.ahora())
}
