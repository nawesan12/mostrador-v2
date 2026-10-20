// Package correo manda mails transaccionales (recuperar contraseña, avisos). Tres implementaciones:
// Resend en producción, SMTP contra Mailpit en desarrollo y Memoria en los tests.
package correo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
	"strings"
	"sync"
	"time"
)

// Mensaje es un mail. HTML y Texto van los dos (los clientes que no muestran HTML usan el texto).
type Mensaje struct {
	Para, Asunto, HTML, Texto string
}

// Correo manda un mensaje.
type Correo interface {
	Enviar(ctx context.Context, m Mensaje) error
}

// Resend usa la API HTTP de Resend.
type Resend struct {
	Clave, De string
	HTTP      *http.Client
	URL       string // por defecto https://api.resend.com/emails
}

func (r *Resend) Enviar(ctx context.Context, m Mensaje) error {
	url := r.URL
	if url == "" {
		url = "https://api.resend.com/emails"
	}
	cuerpo, _ := json.Marshal(map[string]any{"from": r.De, "to": []string{m.Para}, "subject": m.Asunto, "html": m.HTML, "text": m.Texto})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(cuerpo))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+r.Clave)
	req.Header.Set("Content-Type", "application/json")
	c := r.HTTP
	if c == nil {
		c = &http.Client{Timeout: 10 * time.Second}
	}
	res, err := c.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		var b bytes.Buffer
		_, _ = b.ReadFrom(res.Body)
		return fmt.Errorf("resend respondió %d: %s", res.StatusCode, strings.TrimSpace(b.String()))
	}
	return nil
}

// SMTP sin autenticación: para Mailpit en desarrollo (infra/compose.dev.yml).
type SMTP struct{ Direccion, De string }

func (s *SMTP) Enviar(_ context.Context, m Mensaje) error {
	frontera := "mostrador-frontera"
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\n", s.De, m.Para, m.Asunto)
	fmt.Fprintf(&b, "Content-Type: multipart/alternative; boundary=%s\r\n\r\n", frontera)
	fmt.Fprintf(&b, "--%s\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s\r\n", frontera, m.Texto)
	fmt.Fprintf(&b, "--%s\r\nContent-Type: text/html; charset=utf-8\r\n\r\n%s\r\n--%s--\r\n", frontera, m.HTML, frontera)
	return smtp.SendMail(s.Direccion, nil, direccionDe(s.De), []string{m.Para}, []byte(b.String()))
}

func direccionDe(de string) string {
	if i := strings.LastIndex(de, "<"); i >= 0 {
		return strings.TrimSuffix(de[i+1:], ">")
	}
	return de
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

// Ultimo devuelve el último mensaje mandado a esa dirección.
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
