package tiemporeal

import (
	"context"
	"errors"
	"log/slog"
	"math/rand/v2"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Escucha mantiene una conexión dedicada a Postgres con LISTEN (fuera del pool: una conexión del
// pool vuelve a otros usos y perdería la suscripción). Si se corta, reconecta con espera creciente
// y, al volver, avisa a todos que resincronicen: lo que pasó mientras tanto no llegó.
type Escucha struct {
	URL string
	Hub *Hub
	Log *slog.Logger
	// Latido: cada cuánto, sin avisos, se comprueba que la conexión siga viva.
	Latido time.Duration
	// EsperaMin y EsperaMax: la espera entre reintentos (se duplica, con algo de azar).
	EsperaMin, EsperaMax time.Duration

	conectada atomic.Bool
}

// Conectada dice si la escucha está andando ahora.
func (e *Escucha) Conectada() bool { return e.conectada.Load() }

// Correr escucha hasta que se cancele ctx.
func (e *Escucha) Correr(ctx context.Context) {
	if e.Log == nil {
		e.Log = slog.Default()
	}
	if e.Latido <= 0 {
		e.Latido = 30 * time.Second
	}
	if e.EsperaMin <= 0 {
		e.EsperaMin = 500 * time.Millisecond
	}
	if e.EsperaMax <= 0 {
		e.EsperaMax = 30 * time.Second
	}
	primera := true
	espera := e.EsperaMin
	for ctx.Err() == nil {
		err := e.escuchar(ctx, func() {
			if !primera {
				e.Log.Info("tiempo real: la escucha volvió, se pide resincronizar")
				e.Hub.Resincronizar()
			}
			primera = false
			espera = e.EsperaMin
		})
		e.conectada.Store(false)
		if ctx.Err() != nil {
			return
		}
		e.Log.Warn("tiempo real: se cortó la escucha de Postgres", "error", err, "reintento", espera)
		// Con azar, para que varias instancias no reintenten todas a la vez.
		d := espera/2 + time.Duration(rand.Int64N(int64(espera)))
		select {
		case <-ctx.Done():
			return
		case <-time.After(d):
		}
		espera = min(espera*2, e.EsperaMax)
	}
}

func (e *Escucha) escuchar(ctx context.Context, conectada func()) error {
	cctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	conn, err := pgx.Connect(cctx, sinPool(e.URL))
	cancel()
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close(context.WithoutCancel(ctx)) }()
	if _, err := conn.Exec(ctx, "LISTEN "+pgx.Identifier{Canal}.Sanitize()); err != nil {
		return err
	}
	e.conectada.Store(true)
	conectada()
	for {
		wctx, cancel := context.WithTimeout(ctx, e.Latido)
		n, err := conn.WaitForNotification(wctx)
		cancel()
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if errors.Is(err, context.DeadlineExceeded) || pgconn.Timeout(err) {
				// Un rato sin avisos: si la conexión murió sin que nos enteremos, el ping lo dice.
				pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
				err = conn.Ping(pctx)
				cancel()
				if err != nil {
					return err
				}
				continue
			}
			return err
		}
		if ev, ok := Leer(n.Payload); ok {
			e.Hub.Publicar(ev)
		}
	}
}

// sinPool saca de la URL los parámetros del pool (pool_max_conns…), que una conexión suelta
// mandaría a Postgres como si fueran opciones y fallaría.
func sinPool(u string) string {
	p, err := url.Parse(u)
	if err != nil || p.Scheme == "" {
		return u
	}
	q := p.Query()
	for k := range q {
		if strings.HasPrefix(k, "pool_") {
			q.Del(k)
		}
	}
	p.RawQuery = q.Encode()
	return p.String()
}
