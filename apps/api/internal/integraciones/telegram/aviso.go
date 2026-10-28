package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// Bot habla con la Bot API de Telegram. Sin token no manda nada.
type Bot struct {
	Token string
	// Base de la API (en los tests, un servidor falso). Vacía: https://api.telegram.org.
	Base string
	HTTP *http.Client
}

// ErrorTelegram es una respuesta no exitosa de la Bot API (nunca lleva el token).
type ErrorTelegram struct{ Estado int }

func (e *ErrorTelegram) Error() string { return fmt.Sprintf("telegram: HTTP %d", e.Estado) }

// Enviar manda un mensaje de texto plano al chat.
func (b *Bot) Enviar(ctx context.Context, chat int64, texto string) error {
	if b == nil || b.Token == "" {
		return nil
	}
	if r := []rune(texto); len(r) > 4000 {
		texto = string(r[:4000])
	}
	base := b.Base
	if base == "" {
		base = "https://api.telegram.org"
	}
	cuerpo, _ := json.Marshal(map[string]any{"chat_id": chat, "text": texto})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(base, "/")+"/bot"+b.Token+"/sendMessage", bytes.NewReader(cuerpo))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	cli := b.HTTP
	if cli == nil {
		cli = &http.Client{Timeout: 10 * time.Second}
	}
	res, err := cli.Do(req)
	if err != nil {
		// El error de red puede traer la URL con el token: no se propaga.
		return errors.New("telegram: no se pudo conectar con la Bot API")
	}
	defer res.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 64<<10))
	if res.StatusCode >= 300 {
		return &ErrorTelegram{Estado: res.StatusCode}
	}
	return nil
}

// Manejadores del outbox.
func (s *Servicio) Manejadores() map[string]outbox.Manejador {
	return map[string]outbox.Manejador{"telegram.venta": s.avisarVenta}
}

var canales = map[string]string{
	"pos": "en el local", "wa": "por WhatsApp", "web": "en la tienda", "ig": "por Instagram",
	"ml": "en Mercado Libre", "tg": "por Telegram",
}

func pesos(n int64) string {
	s := fmt.Sprint(n)
	var b strings.Builder
	for i, r := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(r)
	}
	return "$ " + b.String()
}

// TextoVenta es el aviso de la demo: "🛒 Nueva venta en el local: pedido #12 por $ 44.600."
func TextoVenta(canal string, numero int, total int64) string {
	donde := ""
	if c, ok := canales[canal]; ok {
		donde = " " + c
	}
	return fmt.Sprintf("🛒 Nueva venta%s: pedido #%d por %s.", donde, numero, pesos(total))
}

// avisarVenta manda el aviso al chat vinculado una sola vez (pedidos.avisos.telegram). Sin token o
// sin chat, no hace nada. Un 4xx de Telegram (el bot bloqueado, el chat borrado) no se reintenta.
func (s *Servicio) avisarVenta(ctx context.Context, ev outbox.Evento) error {
	if s.Bot == nil || s.Bot.Token == "" {
		return nil
	}
	var d struct {
		Numero int `json:"numero"`
	}
	if json.Unmarshal(ev.Datos, &d) != nil || d.Numero <= 0 {
		return nil
	}
	var chat, total int64
	var canal string
	var ya bool
	err := s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT v.chat, p.canal, p.total, p.avisos ? 'telegram'
			FROM pedidos p JOIN telegram_vinculos v ON v.tienda_id = p.tienda_id WHERE p.numero = $1`, d.Numero).
			Scan(&chat, &canal, &total, &ya)
	})
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && ya) {
		return nil
	}
	if err != nil {
		return err
	}
	if err := s.Bot.Enviar(ctx, chat, TextoVenta(canal, d.Numero, total)); err != nil {
		var et *ErrorTelegram
		if errors.As(err, &et) && et.Estado >= 400 && et.Estado < 500 && et.Estado != http.StatusTooManyRequests {
			slog.Warn("telegram: aviso de venta rechazado", "tienda", ev.TiendaID, "estado", et.Estado)
			return nil
		}
		return err // el outbox reintenta con espera creciente
	}
	return s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE pedidos SET avisos = avisos || jsonb_build_object('telegram', $2::text) WHERE numero = $1`,
			d.Numero, time.Now().UTC().Format(time.RFC3339))
		return err
	})
}
