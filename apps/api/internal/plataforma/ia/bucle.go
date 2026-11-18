package ia

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

// Ejecutor corre una herramienta y devuelve su resultado (JSON compacto) o un error para el
// modelo (esError). Las que «escriben» no escriben: devuelven una propuesta para confirmar.
type Ejecutor func(ctx context.Context, nombre string, entrada json.RawMessage) (resultado string, esError bool)

// Resultado de una pregunta.
type Resultado struct {
	Texto   string
	Uso     Uso
	Vueltas int
	// Rechazo: el modelo no quiso responder (stop_reason refusal).
	Rechazo bool
	// Herramientas que pidió, en orden (para el registro y las pruebas).
	Herramientas []string
}

// Textos fijos de los casos raros.
const (
	TextoRechazo    = "No puedo ayudarte con eso."
	TextoSinVueltas = "Me faltó juntar algunos datos para responderte. ¿Me lo preguntás de forma más puntual?"
	TextoVacio      = "No me salió la respuesta. ¿Me lo preguntás de otra forma?"
)

// Conversar corre el bucle: llama al modelo, ejecuta las herramientas que pide y le devuelve los
// resultados, hasta que responde o se cumplen MaxVueltas llamadas. Nunca se llama con una
// transacción abierta: las herramientas abren y cierran la suya.
func Conversar(ctx context.Context, m Modelo, p anthropic.MessageNewParams, ejecutar Ejecutor) (Resultado, error) {
	var r Resultado
	if m == nil {
		return r, ErrNoDisponible
	}
	for r.Vueltas < MaxVueltas {
		r.Vueltas++
		resp, err := m.Crear(ctx, p)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return r, err
			}
			return r, errors.Join(ErrNoDisponible, err)
		}
		r.Uso.Sumar(resp.Usage)
		if resp.StopReason == anthropic.StopReasonRefusal {
			r.Rechazo, r.Texto = true, TextoRechazo
			return r, nil
		}
		var texto strings.Builder
		resultados := []anthropic.ContentBlockParamUnion{}
		for _, b := range resp.Content {
			switch v := b.AsAny().(type) {
			case anthropic.TextBlock:
				texto.WriteString(v.Text)
			case anthropic.ToolUseBlock:
				r.Herramientas = append(r.Herramientas, v.Name)
				if resp.StopReason != anthropic.StopReasonToolUse {
					continue
				}
				res, esError := ejecutar(ctx, v.Name, json.RawMessage(v.JSON.Input.Raw()))
				resultados = append(resultados, anthropic.NewToolResultBlock(v.ID, res, esError))
			}
		}
		if t := strings.TrimSpace(texto.String()); resp.StopReason != anthropic.StopReasonToolUse || len(resultados) == 0 {
			if t == "" {
				t = TextoVacio
			}
			r.Texto = t
			return r, nil
		}
		// La respuesta va al historial antes de los resultados (con sus bloques tal cual).
		p.Messages = append(p.Messages, resp.ToParam(), anthropic.NewUserMessage(resultados...))
	}
	r.Texto = TextoSinVueltas
	return r, nil
}
