// Comando api: el servidor HTTP de Mostrador para el VPS.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/arrepentimiento"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/config"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/servidor"
)

func main() {
	if err := correr(); err != nil {
		slog.Error("la api no arrancó", "error", err)
		os.Exit(1)
	}
}

func correr() error {
	cfg, err := config.Leer()
	if err != nil {
		return err
	}
	if err := cfg.Exigir("BaseApp"); err != nil {
		return err
	}
	// Fuera de desarrollo los mails salen por Resend y hace falta la clave del BFF: si faltan, la API
	// no arranca (mejor que arrancar y no mandar los links de recuperación).
	if !cfg.Desarrollo() {
		if err := cfg.Exigir("ResendClave", "ClaveBFF"); err != nil {
			return err
		}
	}
	var mails correo.Correo = &correo.SMTP{Direccion: cfg.SMTP, De: cfg.CorreoDe}
	if cfg.ResendClave != "" {
		mails = &correo.Resend{Clave: cfg.ResendClave, De: cfg.CorreoDe}
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := db.Abrir(ctx, cfg.BaseApp)
	if err != nil {
		return err
	}
	defer app.Close()

	srv := &http.Server{
		Addr: cfg.Direccion,
		Handler: servidor.Nuevo(servidor.Dependencias{
			App:             app,
			Cuentas:         &cuentas.Servicio{Pool: app, Correo: mails, Sitio: cfg.Sitio},
			Arrepentimiento: &arrepentimiento.Servicio{Pool: app, Correo: mails},
			ClaveBFF:        cfg.ClaveBFF,
		}),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       90 * time.Second,
	}
	errs := make(chan error, 1)
	go func() {
		slog.Info("api escuchando", "direccion", cfg.Direccion, "entorno", cfg.Entorno)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errs <- err
		}
	}()
	select {
	case err := <-errs:
		return err
	case <-ctx.Done():
	}
	apagar, cancelar := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelar()
	return srv.Shutdown(apagar)
}
