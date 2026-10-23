// Comando worker: consume el outbox (eventos de dominio) con reintentos. Corre aparte de la API.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadopago"
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
	if err := cfg.Exigir("BaseWorker", "BaseApp"); err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pool, err := db.Abrir(ctx, cfg.BaseWorker)
	if err != nil {
		return err
	}
	defer pool.Close()
	// Los manejadores usan el rol de la app (con la RLS de cada tienda); el worker sólo toma eventos.
	app, err := db.Abrir(ctx, cfg.BaseApp)
	if err != nil {
		return err
	}
	defer app.Close()
	mp := mercadopago.DesdeConfig(app, cfg.OAuth["mercadopago"][0], cfg.OAuth["mercadopago"][1], cfg.MPToken,
		cfg.MPSecretoWebhook, cfg.Sitio, cfg.ClaveCifrado, slog.Default())
	w := &outbox.Worker{
		Pool: pool,
		// Los dominios registran acá sus manejadores a medida que se suman.
		Manejadores: map[string]outbox.Manejador{},
	}
	for tipo, m := range mp.Manejadores() {
		w.Manejadores[tipo] = m
	}
	slog.Info("worker consumiendo el outbox", "entorno", cfg.Entorno)
	return w.Correr(ctx)
}
