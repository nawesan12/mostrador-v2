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

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/servidor"
)

func main() {
	dir := os.Getenv("MOSTRADOR_DIRECCION")
	if dir == "" {
		dir = ":4000"
	}
	srv := &http.Server{
		Addr:              dir,
		Handler:           servidor.Nuevo(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       90 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		slog.Info("api escuchando", "direccion", dir)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("el servidor se cayó", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	apagar, cancelar := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancelar()
	if err := srv.Shutdown(apagar); err != nil {
		slog.Error("no se pudo apagar a tiempo", "error", err)
	}
}
