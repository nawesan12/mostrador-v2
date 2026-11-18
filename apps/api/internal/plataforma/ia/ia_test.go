package ia

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

// Falso responde con mensajes armados en JSON (como los manda la API) y guarda lo que recibió.
type Falso struct {
	Respuestas []string
	Recibidos  []anthropic.MessageNewParams
	Err        error
}

func (f *Falso) Crear(_ context.Context, p anthropic.MessageNewParams) (*anthropic.Message, error) {
	// Copia de los mensajes: el bucle sigue agregando al mismo slice.
	p.Messages = append([]anthropic.MessageParam(nil), p.Messages...)
	f.Recibidos = append(f.Recibidos, p)
	if f.Err != nil {
		return nil, f.Err
	}
	if len(f.Respuestas) == 0 {
		return nil, errors.New("falso: sin respuestas")
	}
	var m anthropic.Message
	if err := json.Unmarshal([]byte(f.Respuestas[0]), &m); err != nil {
		return nil, err
	}
	f.Respuestas = f.Respuestas[1:]
	return &m, nil
}

func texto(t string) string {
	return `{"id":"m","type":"message","role":"assistant","model":"claude-haiku-5-5","stop_reason":"end_turn",
		"content":[{"type":"text","text":` + jsonStr(t) + `}],
		"usage":{"input_tokens":100,"output_tokens":20,"cache_read_input_tokens":900,"cache_creation_input_tokens":0}}`
}

func herramienta(id, nombre, entrada string) string {
	return `{"id":"m","type":"message","role":"assistant","model":"claude-haiku-5-5","stop_reason":"tool_use",
		"content":[{"type":"text","text":"Miro."},{"type":"tool_use","id":"` + id + `","name":"` + nombre + `","input":` + entrada + `}],
		"usage":{"input_tokens":50,"output_tokens":10,"cache_read_input_tokens":0,"cache_creation_input_tokens":1000}}`
}

func jsonStr(s string) string { b, _ := json.Marshal(s); return string(b) }

func pedido(pregunta string, historial ...Turno) Pedido {
	return Pedido{
		Sistema: "Sos Tincho.",
		Herramientas: []anthropic.ToolParam{
			{Name: "ver_pedido", InputSchema: anthropic.ToolInputSchemaParam{Properties: map[string]any{"numero": map[string]any{"type": "integer"}}}},
			{Name: "listar_productos", InputSchema: anthropic.ToolInputSchemaParam{Properties: map[string]any{}}},
		},
		Datos:     `DATOS DEL NEGOCIO {"tienda":"La Esquina"}`,
		Historial: historial,
		Pregunta:  pregunta,
		MaxTokens: TokensChat,
	}
}

// crudo devuelve el JSON del pedido como lo manda el SDK, por campo.
func crudo(t *testing.T, p anthropic.MessageNewParams) map[string]json.RawMessage {
	t.Helper()
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]json.RawMessage
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

// prefijo: lo que la API lee del caché hasta el segundo breakpoint (herramientas, system y el bloque
// de datos del primer mensaje), en bytes.
func prefijo(t *testing.T, p anthropic.MessageNewParams) []byte {
	m := crudo(t, p)
	var msgs []struct {
		Content []json.RawMessage `json:"content"`
	}
	if err := json.Unmarshal(m["messages"], &msgs); err != nil {
		t.Fatal(err)
	}
	return bytes.Join([][]byte{m["model"], m["output_config"], m["tool_choice"], m["tools"], m["system"], msgs[0].Content[0]}, []byte("|"))
}

func TestPrefijoIdenticoEntrePreguntas(t *testing.T) {
	a := Armar(pedido("¿Cuánto vendí ayer?"))
	b := Armar(pedido("¿Qué cupón me conviene?", Turno{"assistant", "¡Hola!"}, Turno{"user", "hola"}, Turno{"assistant", "¿En qué te ayudo?"}))
	pa, pb := prefijo(t, a), prefijo(t, b)
	if !bytes.Equal(pa, pb) {
		t.Fatalf("el prefijo cambia entre preguntas:\n%s\n%s", pa, pb)
	}
	// Las herramientas van por nombre, el system y los datos con su breakpoint, y el automático arriba.
	ca := crudo(t, a)
	if !strings.Contains(string(ca["tools"]), `"listar_productos"`) || strings.Index(string(ca["tools"]), "listar_productos") > strings.Index(string(ca["tools"]), "ver_pedido") {
		t.Fatalf("herramientas sin ordenar: %s", ca["tools"])
	}
	if !strings.Contains(string(ca["system"]), `"cache_control":{"type":"ephemeral"}`) {
		t.Fatalf("system sin breakpoint: %s", ca["system"])
	}
	if !strings.Contains(string(pa), `DATOS DEL NEGOCIO`) || strings.Count(string(prefijo(t, a)), `"cache_control"`) != 2 {
		t.Fatalf("datos sin breakpoint: %s", pa)
	}
	if string(ca["cache_control"]) != `{"type":"ephemeral"}` || !strings.Contains(string(ca["output_config"]), `"low"`) {
		t.Fatalf("caché automático o esfuerzo: %s %s", ca["cache_control"], ca["output_config"])
	}
	if string(ca["model"]) != `"claude-haiku-5-5"` || string(ca["tool_choice"]) != `{"type":"auto"}` {
		t.Fatalf("modelo o tool_choice: %s %s", ca["model"], ca["tool_choice"])
	}
	// Sin el saludo de Tincho al principio: el primer mensaje es de la persona, con los datos.
	if b.Messages[0].Role != anthropic.MessageParamRoleUser || len(b.Messages) != 3 {
		t.Fatalf("mensajes: %+v", b.Messages)
	}
}

func TestHistorialRecortado(t *testing.T) {
	h := []Turno{}
	for i := 0; i < 12; i++ {
		h = append(h, Turno{"user", strings.Repeat("a", 900)}, Turno{"assistant", "ok"})
	}
	r := Recortar(h)
	if len(r) != MaxTurnos {
		t.Fatalf("turnos: %d", len(r))
	}
	for _, x := range r {
		if len([]rune(x.Texto)) > MaxTextoHistorial+1 {
			t.Fatalf("sin recortar: %d", len(x.Texto))
		}
	}
}

func TestBucleConHerramientas(t *testing.T) {
	f := &Falso{Respuestas: []string{
		herramienta("t1", "ver_pedido", `{"numero":12}`),
		texto("El pedido **#12** está para preparar."),
	}}
	var pedidas []string
	r, err := Conversar(context.Background(), f, Armar(pedido("¿cómo está el 12?")), func(_ context.Context, n string, e json.RawMessage) (string, bool) {
		pedidas = append(pedidas, n+string(e))
		return `{"numero":12,"estado":"Para preparar"}`, false
	})
	if err != nil {
		t.Fatal(err)
	}
	if r.Texto != "El pedido **#12** está para preparar." || r.Vueltas != 2 || len(pedidas) != 1 || pedidas[0] != `ver_pedido{"numero":12}` {
		t.Fatalf("resultado: %+v %v", r, pedidas)
	}
	if r.Uso != (Uso{Entrada: 150, CacheLeido: 900, CacheEscrito: 1000, Salida: 30}) {
		t.Fatalf("uso: %+v", r.Uso)
	}
	// La segunda llamada lleva la respuesta del modelo y el resultado de la herramienta.
	seg := f.Recibidos[1].Messages
	if len(seg) != 3 || seg[1].Role != anthropic.MessageParamRoleAssistant || seg[2].Content[0].OfToolResult == nil ||
		seg[2].Content[0].OfToolResult.ToolUseID != "t1" {
		t.Fatalf("segunda vuelta: %+v", seg)
	}
	// El prefijo con caché es el mismo en las dos vueltas.
	if !bytes.Equal(prefijo(t, f.Recibidos[0]), prefijo(t, f.Recibidos[1])) {
		t.Fatal("el prefijo cambió entre vueltas")
	}
}

func TestBucleTopeDeVueltas(t *testing.T) {
	f := &Falso{}
	for i := 0; i < 10; i++ {
		f.Respuestas = append(f.Respuestas, herramienta("t", "listar_productos", `{}`))
	}
	r, err := Conversar(context.Background(), f, Armar(pedido("todo")), func(context.Context, string, json.RawMessage) (string, bool) {
		return `[]`, false
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Recibidos) != MaxVueltas || r.Texto != TextoSinVueltas {
		t.Fatalf("llamadas %d, texto %q", len(f.Recibidos), r.Texto)
	}
}

func TestBucleRechazoYError(t *testing.T) {
	f := &Falso{Respuestas: []string{`{"id":"m","type":"message","role":"assistant","model":"claude-haiku-5-5","stop_reason":"refusal","content":[],"usage":{"input_tokens":10,"output_tokens":0}}`}}
	r, err := Conversar(context.Background(), f, Armar(pedido("algo")), nil)
	if err != nil || !r.Rechazo || r.Texto != TextoRechazo {
		t.Fatalf("rechazo: %+v %v", r, err)
	}
	_, err = Conversar(context.Background(), &Falso{Err: errors.New("503")}, Armar(pedido("algo")), nil)
	if !errors.Is(err, ErrNoDisponible) {
		t.Fatalf("error: %v", err)
	}
	if _, err := Conversar(context.Background(), nil, Armar(pedido("algo")), nil); !errors.Is(err, ErrNoDisponible) {
		t.Fatalf("sin modelo: %v", err)
	}
}

func TestCosto(t *testing.T) {
	// 1M de entrada = USD 0,10; 1M de salida = USD 0,50; 1M leído del caché = USD 0,01.
	if c := (Uso{Entrada: 1_000_000}).Costo(); c != 100_000 {
		t.Fatalf("entrada: %d", c)
	}
	if c := (Uso{Salida: 1_000_000, CacheLeido: 1_000_000}).Costo(); c != 510_000 {
		t.Fatalf("salida + leído: %d", c)
	}
	if c := (Uso{Entrada: 1}).Costo(); c != 1 {
		t.Fatalf("redondea para arriba: %d", c)
	}
}

func TestDecidir(t *testing.T) {
	topes := Topes{Mes: 100, Dia: 25, Presupuesto: 40_000_000}
	casos := []struct {
		c       Consumo
		gastado int64
		corte   Corte
	}{
		{Consumo{Mes: 10, Hoy: 3}, 0, SinCorte},
		{Consumo{Mes: 100, Hoy: 3}, 0, CortePlan},
		{Consumo{Mes: 50, Hoy: 25}, 0, CorteDia},
		{Consumo{Mes: 1, Hoy: 1}, 40_000_000, CortePlataforma},
	}
	for _, c := range casos {
		if got := Decidir(c.c, c.gastado, topes); got != c.corte {
			t.Errorf("%+v %d: %q, quería %q", c.c, c.gastado, got, c.corte)
		}
	}
	if Decidir(Consumo{}, 1<<40, Topes{Mes: 1, Dia: 1}) != SinCorte {
		t.Error("sin presupuesto no corta por plataforma")
	}
}

func TestCacheDeRespuestas(t *testing.T) {
	ahora := time.Date(2026, 11, 18, 10, 0, 0, 0, time.UTC)
	c := NuevoCache()
	c.Ahora = func() time.Time { return ahora }
	k := Clave("tienda", Normalizar("¿Cuánto vendí AYER?"))
	if k != Clave("tienda", Normalizar("cuanto vendi ayer")) {
		t.Fatal("la normalización no iguala")
	}
	c.Guardar(k, "Ayer vendiste $ 10.000.")
	if v, ok := c.Leer(k); !ok || v != "Ayer vendiste $ 10.000." {
		t.Fatal("no lo guardó")
	}
	ahora = ahora.Add(VidaRespuesta)
	if _, ok := c.Leer(k); ok {
		t.Fatal("no venció a los 10 minutos")
	}
}
