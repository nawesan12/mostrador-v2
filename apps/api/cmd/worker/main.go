// Comando worker: consume el outbox (eventos de dominio) con reintentos. Corre aparte de la API.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/config"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

func main() {
	if err := correr(); err != nil {
		slog.Error("el worker no arrancó", "error", err)
		os.Exit(1)
	}
}

func correr() error {
	cfg, err := config.Leer()
	if err != nil {
		return err
	}
	if err := cfg.Exigir("BaseWorker"); err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pool, err := db.Abrir(ctx, cfg.BaseWorker)
	if err != nil {
		return err
	}
	defer pool.Close()
	w := &outbox.Worker{
		Pool: pool,
		// Los dominios registran acá sus manejadores a medida que se suman.
		Manejadores: map[string]outbox.Manejador{},
	}
	slog.Info("worker consumiendo el outbox", "entorno", cfg.Entorno)
	return w.Correr(ctx)
}
