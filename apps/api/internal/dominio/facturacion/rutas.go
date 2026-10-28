package facturacion

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/arca"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

func escribir(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(v)
}

// Error lleva un error del dominio a problem+json (también lo usa listas para las acciones).
func Error(w http.ResponseWriter, err error) {
	var r arca.Rechazado
	var e arca.ErrARCA
	switch {
	case errors.Is(err, ErrInvalido):
		problema.Escribir(w, problema.Invalido("Revisá los datos."))
	case errors.Is(err, ErrSinPermiso):
		problema.Escribir(w, problema.SinPermiso("Sólo quien es dueña o admin maneja la conexión con ARCA."))
	case errors.Is(err, ErrSinARCA):
		problema.Escribir(w, problema.Conflicto("sin_arca", "Primero conectá ARCA con tu certificado."))
	case errors.Is(err, arca.ErrCertificadoVencido):
		problema.Escribir(w, problema.Invalido("Ese certificado ya venció: generá uno nuevo en ARCA con el pedido que bajaste de acá."))
	case errors.Is(err, arca.ErrCertificado):
		problema.Escribir(w, problema.Invalido("No pudimos leer el certificado: pegá el archivo .crt completo que te dio ARCA, desde -----BEGIN CERTIFICATE-----."))
	case errors.Is(err, ErrCertificado):
		problema.Escribir(w, problema.Invalido("Ese certificado no es de esta conexión: tiene que ser el que ARCA generó con el pedido que bajaste de acá, del mismo CUIT."))
	case errors.As(err, &r):
		problema.Escribir(w, problema.NoProcesable("arca_"+r.Codigo, r.Mensaje))
	case errors.As(err, &e):
		problema.Escribir(w, &problema.Problema{Estado: http.StatusBadGateway, Codigo: "arca", Titulo: "ARCA", Detalle: e.Mensaje})
	case errors.Is(err, pgx.ErrNoRows):
		problema.Escribir(w, problema.NoEncontrado("No existe."))
	default:
		slog.Error("facturacion", "error", err)
		problema.Escribir(w, problema.Interno())
	}
}

func (s *Servicio) Rutas(mux *http.ServeMux) {
	con := func(f func(w http.ResponseWriter, r *http.Request, d cuentas.Datos)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			d, ok := cuentas.SesionDe(r)
			if !ok || d.TiendaID == "" {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			f(w, r, d)
		}
	}
	leer := func(w http.ResponseWriter, r *http.Request, v any) bool {
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10)).Decode(v) != nil {
			problema.Escribir(w, problema.Invalido("El pedido no es válido."))
			return false
		}
		return true
	}
	mux.HandleFunc("GET /v1/panel/arca", con(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		c, err := s.Leer(r.Context(), d.TiendaID)
		if err != nil {
			Error(w, err)
			return
		}
		escribir(w, http.StatusOK, c)
	}))
	// Genera la clave y devuelve el pedido de certificado (CSR) para subir a ARCA.
	mux.HandleFunc("POST /v1/panel/arca/pedido", con(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		var a Alta
		if !leer(w, r, &a) {
			return
		}
		csr, err := s.GenerarPedido(r.Context(), d.TiendaID, d.Rol, a)
		if err != nil {
			Error(w, err)
			return
		}
		escribir(w, http.StatusOK, map[string]string{"pedido": string(csr)})
	}))
	mux.HandleFunc("PUT /v1/panel/arca/certificado", con(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		var a struct {
			Certificado string `json:"certificado"`
		}
		if !leer(w, r, &a) {
			return
		}
		c, err := s.GuardarCertificado(r.Context(), d.TiendaID, d.Rol, a.Certificado, time.Now())
		if err != nil {
			Error(w, err)
			return
		}
		escribir(w, http.StatusOK, c)
	}))
	mux.HandleFunc("PUT /v1/panel/arca/ajustes", con(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		var a Ajustes
		if !leer(w, r, &a) {
			return
		}
		if err := s.Ajustar(r.Context(), d.TiendaID, d.Rol, a); err != nil {
			Error(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	mux.HandleFunc("POST /v1/panel/arca/probar", con(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		n, err := s.Probar(r.Context(), d.TiendaID, time.Now())
		if err != nil {
			Error(w, err)
			return
		}
		escribir(w, http.StatusOK, map[string]any{"ok": true, "ultimo": n, "detalle": fmt.Sprintf("Conectada. Último comprobante autorizado: %d.", n)})
	}))
}
