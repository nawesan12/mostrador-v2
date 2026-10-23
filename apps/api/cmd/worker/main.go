// Comando worker: consume el outbox (eventos de dominio) con reintentos. Corre aparte de la API.
package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/avisos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/facturacion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/ia"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadolibre"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadopago"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/config"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
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
	ml := mercadolibre.DesdeConfig(app, cfg.OAuth["mercadolibre"][0], cfg.OAuth["mercadolibre"][1], cfg.Sitio,
		cfg.ClaveCifrado, slog.Default())
	// Mails a quien compra (confirmación, pago, en camino, retiro): el mismo correo que la API, y
	// como en la API, sin Resend fuera de desarrollo no arranca (no se manda a Mailpit en producción).
	if !cfg.Desarrollo() {
		if err := cfg.Exigir("ResendClave"); err != nil {
			return err
		}
	}
	var mails correo.Correo = &correo.SMTP{Direccion: cfg.SMTP, De: cfg.CorreoDe}
	if cfg.ResendClave != "" {
		mails = &correo.Resend{Clave: cfg.ResendClave, De: cfg.CorreoDe}
	}
	av := &avisos.Servicio{Pool: app, Correo: mails, Sitio: cfg.Sitio}
	for _, ms := range []map[string]outbox.Manejador{mp.Manejadores(), ml.Manejadores(), av.Manejadores()} {
		for tipo, m := range ms {
			w.Manejadores[tipo] = m
		}
	}
	// Limpieza de las IA conectadas: códigos vencidos y clientes que nadie usa hace 30 días.
	go func() {
		iaS := &ia.Servicio{Pool: app}
		t := time.NewTicker(6 * time.Hour)
		defer t.Stop()
		for {
			if err := iaS.Purgar(ctx); err != nil && ctx.Err() == nil {
				slog.Error("ia: purgar", "error", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
		}
	}()
	// Facturación: el drenador pide los CAE de las ventas cobradas (cada 30 s, con espera creciente
	// por comprobante). Corre en el servidor propio.
	go func() {
		f := &facturacion.Servicio{Pool: app}
		if c, err := cifrado.Nuevo(cfg.ClaveCifrado); err == nil && cfg.ClaveCifrado != "" {
			f.Cifra = c
		}
		t := time.NewTicker(30 * time.Second)
		defer t.Stop()
		for {
			if err := f.Drenar(ctx, time.Now()); err != nil && ctx.Err() == nil {
				slog.Error("facturacion: drenar", "error", err)
			}
			select {
			case <-ctx.Done():
				return
			case <-t.C:
			}
		}
	}()
	slog.Info("worker consumiendo el outbox", "entorno", cfg.Entorno)
	return w.Correr(ctx)
}
