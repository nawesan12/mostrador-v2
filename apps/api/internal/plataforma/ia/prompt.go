package ia

import (
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/anthropics/anthropic-sdk-go"
)

// Historial: últimos 8 turnos, cada uno recortado a 600 caracteres.
const (
	MaxTurnos         = 8
	MaxTextoHistorial = 600
)

// Turno es un mensaje anterior de la charla.
type Turno struct {
	// "user" o "assistant".
	Rol   string
	Texto string
}

// Pedido es lo que arma cada asistente. El orden del caché es herramientas → system → mensajes:
// las herramientas van siempre ordenadas por nombre, Sistema es fijo por app (sin fechas ni datos)
// y Datos (el negocio, armado de forma determinística) es el primer bloque del primer mensaje.
type Pedido struct {
	Sistema      string
	Herramientas []anthropic.ToolParam
	Datos        string
	Historial    []Turno
	Pregunta     string
	MaxTokens    int64
}

// Armar devuelve los parámetros con tres puntos de caché: el último bloque del system (cubre
// herramientas y system), el bloque de datos del negocio y el automático para la charla.
func Armar(p Pedido) anthropic.MessageNewParams {
	herr := slices.Clone(p.Herramientas)
	slices.SortFunc(herr, func(a, b anthropic.ToolParam) int { return strings.Compare(a.Name, b.Name) })
	tools := make([]anthropic.ToolUnionParam, len(herr))
	for i := range herr {
		tools[i] = anthropic.ToolUnionParam{OfTool: &herr[i]}
	}

	datos := anthropic.TextBlockParam{Text: p.Datos, CacheControl: anthropic.NewCacheControlEphemeralParam()}
	msgs := []anthropic.MessageParam{}
	agregar := func(rol, texto string) {
		bloque := anthropic.ContentBlockParamUnion{OfText: &anthropic.TextBlockParam{Text: texto}}
		if n := len(msgs); n > 0 && string(msgs[n-1].Role) == rol {
			msgs[n-1].Content = append(msgs[n-1].Content, bloque)
			return
		}
		if len(msgs) == 0 {
			// El primer mensaje es de la persona y empieza con los datos del negocio.
			msgs = append(msgs, anthropic.NewUserMessage(anthropic.ContentBlockParamUnion{OfText: &datos}, bloque))
			return
		}
		if rol == "assistant" {
			msgs = append(msgs, anthropic.NewAssistantMessage(bloque))
		} else {
			msgs = append(msgs, anthropic.NewUserMessage(bloque))
		}
	}
	for _, t := range Recortar(p.Historial) {
		if len(msgs) == 0 && t.Rol != "user" {
			// Un saludo de Tincho al principio no hace falta: el primer mensaje es de la persona.
			continue
		}
		agregar(t.Rol, t.Texto)
	}
	agregar("user", p.Pregunta)

	params := anthropic.MessageNewParams{
		Model:        ModeloHaiku,
		MaxTokens:    p.MaxTokens,
		System:       []anthropic.TextBlockParam{{Text: p.Sistema, CacheControl: anthropic.NewCacheControlEphemeralParam()}},
		Messages:     msgs,
		OutputConfig: anthropic.OutputConfigParam{Effort: Esfuerzo},
		// Automático: la charla y las vueltas de herramientas leen lo de la vuelta anterior.
		CacheControl: anthropic.NewCacheControlEphemeralParam(),
	}
	if len(tools) > 0 {
		params.Tools = tools
		params.ToolChoice = anthropic.ToolChoiceUnionParam{OfAuto: &anthropic.ToolChoiceAutoParam{}}
	}
	return params
}

// Recortar deja los últimos turnos con texto, cada uno hasta MaxTextoHistorial caracteres.
func Recortar(h []Turno) []Turno {
	out := []Turno{}
	for _, t := range h {
		t.Texto = strings.TrimSpace(t.Texto)
		if t.Texto == "" || (t.Rol != "user" && t.Rol != "assistant") {
			continue
		}
		if utf8.RuneCountInString(t.Texto) > MaxTextoHistorial {
			t.Texto = string([]rune(t.Texto)[:MaxTextoHistorial]) + "…"
		}
		out = append(out, t)
	}
	if len(out) > MaxTurnos {
		out = out[len(out)-MaxTurnos:]
	}
	return out
}
