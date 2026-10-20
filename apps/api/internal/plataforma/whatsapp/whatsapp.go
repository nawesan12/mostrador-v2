// Package whatsapp manda mensajes de la plataforma por WhatsApp (links de ingreso y de contraseña
// nueva). Con credenciales usa la API de WhatsApp Cloud (Meta) y una plantilla aprobada; en
// desarrollo y en tests, una memoria o el log.
package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// Mensaje: el link y para qué es (elige la plantilla). El botón de una plantilla de WhatsApp tiene
// una URL fija con una parte variable al final: la plantilla se da de alta con
// "https://mostrador.iamnaw.dev/ingresar?{{1}}" y acá va Sufijo ("m=enlace&token=…"). Link es la
// dirección completa (log y tests).
type Mensaje struct {
	Para   string // E.164 sin "+"
	Tipo   string // "ingreso" | "restablecer"
	Link   string
	Sufijo string
}

// Mensajero manda un mensaje.
type Mensajero interface {
	Enviar(ctx context.Context, m Mensaje) error
}

// Cloud es WhatsApp Cloud API. Cada tipo tiene su plantilla (de utilidad, con un botón de link cuyo
// sufijo es el token): {Tipo: nombre de la plantilla}.
type Cloud struct {
	Token, NumeroID string
	Plantillas      map[string]string
	Base            string // por defecto https://graph.facebook.com/v21.0
	HTTP            *http.Client
}

func (c *Cloud) Enviar(ctx context.Context, m Mensaje) error {
	plantilla, ok := c.Plantillas[m.Tipo]
	if !ok {
		return fmt.Errorf("whatsapp: sin plantilla para %q", m.Tipo)
	}
	base := c.Base
	if base == "" {
		base = "https://graph.facebook.com/v21.0"
	}
	cuerpo, _ := json.Marshal(map[string]any{
		"messaging_product": "whatsapp",
		"to":                m.Para,
		"type":              "template",
		"template": map[string]any{
			"name":     plantilla,
			"language": map[string]string{"code": "es_AR"},
			"components": []any{map[string]any{
				"type": "button", "sub_type": "url", "index": "0",
				"parameters": []any{map[string]string{"type": "text", "text": m.Sufijo}},
			}},
		},
	})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/"+c.NumeroID+"/messages", bytes.NewReader(cuerpo))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.Token)
	req.Header.Set("Content-Type", "application/json")
	cli := c.HTTP
	if cli == nil {
		cli = &http.Client{Timeout: 10 * time.Second}
	}
	res, err := cli.Do(req)
	if err != nil {
		return fmt.Errorf("whatsapp: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(res.Body, 512))
		return fmt.Errorf("whatsapp: Meta respondió %d: %s", res.StatusCode, b)
	}
	return nil
}

// Log escribe el mensaje en el log (desarrollo, sin credenciales).
type Log struct{}

func (Log) Enviar(_ context.Context, m Mensaje) error {
	slog.Info("whatsapp (sin credenciales, no se manda)", "para", m.Para, "tipo", m.Tipo, "link", m.Link)
	return nil
}

// Memoria guarda los mensajes (tests).
type Memoria struct {
	mu       sync.Mutex
	Enviados []Mensaje
}

func (m *Memoria) Enviar(_ context.Context, msj Mensaje) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Enviados = append(m.Enviados, msj)
	return nil
}

// Ultimo devuelve el último mensaje a ese número.
func (m *Memoria) Ultimo(para string) (Mensaje, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := len(m.Enviados) - 1; i >= 0; i-- {
		if m.Enviados[i].Para == para {
			return m.Enviados[i], true
		}
	}
	return Mensaje{}, false
}

// Normalizar lleva un número argentino a E.164 sin "+" para WhatsApp (549 + característica +
// número). Acepta "11 5555-0000", "+54 9 11 5555 0000", "0351 15 555 0000"… Devuelve "" si no sirve.
func Normalizar(s string) string {
	d := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			d = append(d, s[i])
		}
	}
	n := string(d)
	switch {
	case len(n) == 13 && n[:3] == "549":
		return n
	case len(n) == 12 && n[:2] == "54":
		return "549" + n[2:]
	}
	if len(n) > 0 && n[0] == '0' {
		n = n[1:]
	}
	// "0351 15 555 0000": característica + 15 + número; el 15 se saca.
	if len(n) == 12 {
		for _, k := range []int{2, 3, 4} {
			if n[k:k+2] == "15" {
				n = n[:k] + n[k+2:]
				break
			}
		}
	}
	if len(n) == 10 {
		return "549" + n
	}
	return ""
}
