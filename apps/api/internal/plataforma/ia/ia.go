// Package ia es la pieza común de los asistentes con Claude (docs-ia/asistente-ia.md): el cliente
// detrás de una interfaz (las pruebas usan uno falso, sin red), el armado del prompt con caché, el
// bucle de herramientas con tope de vueltas, la medición en la base, los topes y el caché de
// respuestas. Lo propio de cada asistente (system, herramientas, datos) vive en su dominio.
//
// No confundir con dominio/ia, que es el OAuth de las IA conectadas por MCP.
package ia

import (
	"context"
	"errors"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Un solo modelo para todo, con esfuerzo bajo fijo: cambiarlo invalida el caché.
const (
	ModeloHaiku = "claude-haiku-5-5"
	Esfuerzo    = anthropic.OutputConfigEffortLow
	// Timeout de cada llamada (el SDK reintenta una vez ante 429/5xx/conexión).
	TimeoutLlamada = 25 * time.Second
	// Vueltas de herramientas por pregunta, como máximo.
	MaxVueltas = 4
)

// max_tokens por ruta (chat corto, textos más largos).
const (
	TokensChat   = 800
	TokensTextos = 1200
)

// ErrNoDisponible: la IA no respondió (timeout, error de red o de la API). Quien llama contesta
// con el motor lo que pueda.
var ErrNoDisponible = errors.New("ia: no disponible")

// Modelo es lo único que se le pide a la API de Anthropic: crear un mensaje.
type Modelo interface {
	Crear(ctx context.Context, p anthropic.MessageNewParams) (*anthropic.Message, error)
}

type cliente struct{ c anthropic.Client }

// Nuevo arma el cliente real con la clave (ANTHROPIC_API_KEY). Sin clave devuelve nil: el
// asistente queda sólo con el motor.
func Nuevo(clave string) Modelo {
	if clave == "" {
		return nil
	}
	return cliente{anthropic.NewClient(option.WithAPIKey(clave), option.WithMaxRetries(1),
		option.WithRequestTimeout(TimeoutLlamada))}
}

func (c cliente) Crear(ctx context.Context, p anthropic.MessageNewParams) (*anthropic.Message, error) {
	return c.c.Messages.New(ctx, p)
}

// Uso son los tokens de una respuesta (o la suma de las vueltas de una pregunta).
type Uso struct {
	Entrada      int64 `json:"entrada"`
	CacheLeido   int64 `json:"cacheLeido"`
	CacheEscrito int64 `json:"cacheEscrito"`
	Salida       int64 `json:"salida"`
}

func (u *Uso) Sumar(x anthropic.Usage) {
	u.Entrada += x.InputTokens
	u.CacheLeido += x.CacheReadInputTokens
	u.CacheEscrito += x.CacheCreationInputTokens
	u.Salida += x.OutputTokens
}

// Precios de Haiku 5.5 en milésimas de micro-dólar por token: USD 0,10 / 0,50 por millón de
// entrada / salida; el caché leído sale 0,1× y el escrito (5 min) 1,25×.
const (
	precioEntrada = 100
	precioLeido   = 10
	precioEscrito = 125
	precioSalida  = 500
)

// Costo en micro-dólares, redondeado para arriba.
func (u Uso) Costo() int64 {
	m := u.Entrada*precioEntrada + u.CacheLeido*precioLeido + u.CacheEscrito*precioEscrito + u.Salida*precioSalida
	return (m + 999) / 1000
}
