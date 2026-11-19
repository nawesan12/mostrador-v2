package tincho_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tincho"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/ia"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/servidor"
)

const claveBFF = "clave-del-bff-de-prueba"

// falso es el modelo sin red: responde lo armado y guarda lo que recibió.
type falso struct {
	mu         sync.Mutex
	respuestas []string
	recibidos  []anthropic.MessageNewParams
	err        error
}

func (f *falso) Crear(_ context.Context, p anthropic.MessageNewParams) (*anthropic.Message, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	p.Messages = append([]anthropic.MessageParam(nil), p.Messages...)
	f.recibidos = append(f.recibidos, p)
	if f.err != nil {
		return nil, f.err
	}
	if len(f.respuestas) == 0 {
		return nil, errors.New("falso: sin respuestas")
	}
	var m anthropic.Message
	if err := json.Unmarshal([]byte(f.respuestas[0]), &m); err != nil {
		return nil, err
	}
	f.respuestas = f.respuestas[1:]
	return &m, nil
}

func texto(t string) string {
	b, _ := json.Marshal(t)
	return `{"id":"m","type":"message","role":"assistant","model":"claude-haiku-5-5","stop_reason":"end_turn",
		"content":[{"type":"text","text":` + string(b) + `}],"usage":{"input_tokens":120,"output_tokens":30,"cache_read_input_tokens":2000}}`
}

func usar(id, nombre, entrada string) string {
	return `{"id":"m","type":"message","role":"assistant","model":"claude-haiku-5-5","stop_reason":"tool_use",
		"content":[{"type":"tool_use","id":"` + id + `","name":"` + nombre + `","input":` + entrada + `}],
		"usage":{"input_tokens":100,"output_tokens":15,"cache_creation_input_tokens":2000}}`
}

type api struct {
	t *testing.T
	h http.Handler
}

func (a api) pedir(metodo, ruta, cuerpo, token string, bff bool) *httptest.ResponseRecorder {
	r := httptest.NewRequest(metodo, ruta, strings.NewReader(cuerpo))
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	if bff {
		r.Header.Set("X-Mostrador-BFF", claveBFF)
	}
	rec := httptest.NewRecorder()
	a.h.ServeHTTP(rec, r)
	return rec
}

func leer[T any](t *testing.T, rec *httptest.ResponseRecorder, estado int) T {
	t.Helper()
	var v T
	if rec.Code != estado {
		t.Fatalf("esperaba %d, vino %d: %s", estado, rec.Code, rec.Body)
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &v)
	return v
}

func nuevo(t *testing.T, m *falso) (api, *pruebadb.Bases, *tincho.Servicio) {
	b := pruebadb.Nueva(t)
	ahora := time.Date(2026, 11, 18, 15, 7, 0, 0, time.UTC)
	tin := &tincho.Servicio{Pool: b.App, Modelo: m, Cache: ia.NuevoCache(), Presupuesto: 40_000_000,
		Ahora: func() time.Time { return ahora }}
	h := servidor.Nuevo(servidor.Dependencias{App: b.App, ClaveBFF: claveBFF,
		Cuentas: &cuentas.Servicio{Pool: b.App, Correo: &correo.Memoria{}, Sitio: "https://tulocalweb.com.ar"},
		Tincho:  tin})
	return api{t, h}, b, tin
}

func (a api) registrar(email, tienda string) cuentas.Sesion {
	a.t.Helper()
	return leer[cuentas.Sesion](a.t, a.pedir("POST", "/v1/auth/registro",
		`{"nombre":"Lucía","tienda":"`+tienda+`","email":"`+email+`","clave":"Abcdefghijk1"}`, "", false), 201)
}

func TestPreguntarConHerramientasPropuestaYCache(t *testing.T) {
	m := &falso{respuestas: []string{
		usar("t1", "listar_productos", `{"limite":3}`),
		usar("t2", "crear_cupon", `{"codigo":"VERANO15","tipo":"porcentaje","valor":15}`),
		texto("Te dejé lista la propuesta del cupón **VERANO15**: confirmala con el botón."),
		texto("Tu ticket promedio viene bien."),
	}}
	a, _, _ := nuevo(t, m)
	s := a.registrar("lucia@laesquina.com.ar", "La Esquina")

	r := leer[tincho.Respuesta](t, a.pedir("POST", "/v1/panel/tincho/preguntar",
		`{"pregunta":"Armame un cupón del 15 para lo que menos sale","historial":[{"rol":"hornero","texto":"¡Hola!"}]}`, s.Token, false), 200)
	if r.Fuente != "ia" || len(r.Propuestas) != 1 || r.Propuestas[0].Herramienta != "crear_cupon" ||
		r.Propuestas[0].Argumentos["codigo"] != "VERANO15" || r.Uso == nil || r.Uso.Usadas != 1 || r.Uso.Tope != 2000 {
		t.Fatalf("respuesta: %+v", r)
	}
	if len(m.recibidos) != 3 {
		t.Fatalf("llamadas al modelo: %d", len(m.recibidos))
	}
	// El resultado de listar_productos fue JSON compacto, sin errores.
	res := m.recibidos[1].Messages[len(m.recibidos[1].Messages)-1].Content[0].OfToolResult
	if res == nil || res.IsError.Value || !strings.Contains(res.Content[0].OfText.Text, `"productos"`) {
		t.Fatalf("resultado de la herramienta: %+v", res)
	}
	// Los datos del negocio van primero, con el nombre de la tienda.
	datos := m.recibidos[0].Messages[0].Content[0].OfText
	if !strings.Contains(datos.Text, "La Esquina") || !strings.HasPrefix(datos.Text, "DATOS DEL NEGOCIO (al 18/11/2026 12:00") {
		t.Fatalf("datos: %s", datos.Text)
	}

	// Una pregunta sin acciones se guarda: repetirla (con otras tildes) no llama al modelo.
	_ = leer[tincho.Respuesta](t, a.pedir("POST", "/v1/panel/tincho/preguntar", `{"pregunta":"¿Cómo viene el ticket?"}`, s.Token, false), 200)
	r = leer[tincho.Respuesta](t, a.pedir("POST", "/v1/panel/tincho/preguntar", `{"pregunta":"como viene el ticket"}`, s.Token, false), 200)
	if r.Fuente != "cache" || r.Texto != "Tu ticket promedio viene bien." || len(m.recibidos) != 4 {
		t.Fatalf("caché: %+v (%d llamadas)", r, len(m.recibidos))
	}
	// La segunda pregunta leyó el mismo prefijo (los mismos datos de la franja).
	if m.recibidos[3].Messages[0].Content[0].OfText.Text != datos.Text {
		t.Fatal("los datos cambiaron dentro de la misma franja")
	}

	// La pantalla del plan muestra lo medido: 2 respuestas con IA (la del caché no cuenta).
	plan := leer[struct {
		Uso struct{ Hornero, SinIa int }
	}](t, a.pedir("GET", "/v1/panel/cuenta/plan", "", s.Token, false), 200)
	if plan.Uso.Hornero != 2 {
		t.Fatalf("uso del plan: %+v", plan.Uso)
	}
	if rec := a.pedir("POST", "/v1/panel/tincho/motor", "", s.Token, false); rec.Code != 204 {
		t.Fatalf("motor: %d", rec.Code)
	}
	plan = leer[struct {
		Uso struct{ Hornero, SinIa int }
	}](t, a.pedir("GET", "/v1/panel/cuenta/plan", "", s.Token, false), 200)
	if plan.Uso.SinIa != 1 || plan.Uso.Hornero != 2 {
		t.Fatalf("uso del plan con el motor: %+v", plan.Uso)
	}
}

func TestTopeDelPlanYFallas(t *testing.T) {
	m := &falso{}
	a, b, tin := nuevo(t, m)
	s := a.registrar("ana@otra.com.ar", "Otra")

	// La IA falla: 503 con el código que el BFF reconoce (y el motor contesta lo que pueda).
	m.err = errors.New("529 overloaded")
	rec := a.pedir("POST", "/v1/panel/tincho/preguntar", `{"pregunta":"¿qué hago con el stock?"}`, s.Token, false)
	if rec.Code != 503 || !strings.Contains(rec.Body.String(), "ia_no_disponible") {
		t.Fatalf("falla: %d %s", rec.Code, rec.Body)
	}

	// Llegó al tope del plan (3.000 en la prueba): la IA no se llama y el mensaje lo dice.
	tienda := s.TiendaID
	// La falla no contó como respuesta.
	if _, c, _ := ia.Revisar(context.Background(), b.App, tienda, ia.Topes{Mes: 10, Dia: 10}, tin.Ahora()); c.Mes != 0 {
		t.Fatalf("la falla contó: %+v", c)
	}
	if err := ia.Registrar(context.Background(), b.App, tienda, ia.RutaChat, ia.Uso{}, tin.Ahora()); err != nil {
		t.Fatal(err)
	}
	if err := b.App.EnTienda(context.Background(), tienda, func(tx pgx.Tx) error {
		_, err := tx.Exec(context.Background(), `UPDATE ia_uso SET pedidos = 2000, dia = dia - 1 WHERE tienda_id = $1`, tienda)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	m.err, m.recibidos = nil, nil
	r := leer[tincho.Respuesta](t, a.pedir("POST", "/v1/panel/tincho/preguntar", `{"pregunta":"¿qué hago con el stock?"}`, s.Token, false), 200)
	if r.Fuente != "tope" || r.Tope != "plan" || !strings.Contains(r.Texto, "2.000 respuestas") || len(m.recibidos) != 0 {
		t.Fatalf("tope: %+v", r)
	}
	r = leer[tincho.Respuesta](t, a.pedir("POST", "/v1/panel/tincho/mejorar", `{"nombre":"Mate imperial","descripcion":""}`, s.Token, false), 200)
	if r.Tope != "plan" {
		t.Fatalf("mejorar con tope: %+v", r)
	}

	// Telegram: sólo el BFF.
	if rec := a.pedir("POST", "/v1/telegram/preguntar", `{"chat":1,"pregunta":"hola"}`, s.Token, false); rec.Code != 403 {
		t.Fatalf("telegram sin clave: %d", rec.Code)
	}
	// Sin sesión no hay Tincho.
	if rec := a.pedir("POST", "/v1/panel/tincho/preguntar", `{"pregunta":"hola"}`, "", false); rec.Code != 401 {
		t.Fatalf("sin sesión: %d", rec.Code)
	}
}
