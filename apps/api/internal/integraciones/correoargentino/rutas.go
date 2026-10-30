package correoargentino

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/limite"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Problema traduce los errores para el panel. Los mensajes de Correo (peso, sucursal, provincia)
// sólo se le muestran a la tienda, nunca a quien compra (ver problemaPublico).
func (s *Servicio) Problema(w http.ResponseWriter, err error) {
	var inv ErrInvalido
	var eca *ErrorCA
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.As(err, &inv):
		problema.Escribir(w, problema.Invalido(inv.Motivo))
	case errors.Is(err, ErrApagado):
		problema.Escribir(w, problema.Conflicto("sin_envios", "Correo Argentino todavía no está disponible en Tu Local Web."))
	case errors.Is(err, ErrSinCredenciales):
		problema.Escribir(w, problema.Conflicto("sin_credenciales", "Faltan las credenciales de la API de MiCorreo."))
	case errors.Is(err, ErrSinConexion):
		problema.Escribir(w, problema.Conflicto("sin_conexion", "La tienda todavía no conectó Correo Argentino."))
	case errors.Is(err, ErrCuenta):
		problema.Escribir(w, problema.Invalido("MiCorreo no reconoce ese email y contraseña. Revisalos en micorreo.correoargentino.com.ar."))
	case errors.Is(err, ErrNoExiste):
		problema.Escribir(w, problema.NoEncontrado("Ese pedido no existe."))
	case errors.Is(err, ErrNoEsCorreo):
		problema.Escribir(w, problema.Conflicto("no_es_correo", "Este pedido no va por Correo Argentino."))
	case errors.As(err, &eca) && eca.Estado == http.StatusPaymentRequired && eca.Mensaje != "":
		// 402: los datos llegaron bien pero Correo no los aceptó (peso, sucursal, provincia…).
		problema.Escribir(w, problema.Conflicto("correo_rechazo", "Correo Argentino respondió: "+eca.Mensaje))
	case errors.As(err, &eca):
		s.log().Warn("correoargentino: la API respondió con error", "estado", eca.Estado, "mensaje", eca.Mensaje)
		problema.Escribir(w, &problema.Problema{Estado: http.StatusBadGateway, Codigo: "correo",
			Titulo: "Correo Argentino no respondió", Detalle: "Probá de nuevo en un rato."})
	default:
		s.log().Error("correoargentino", "error", err)
		problema.Escribir(w, problema.Interno())
	}
}

// problemaPublico: a quien compra sólo le llega que no se pudo cotizar.
func (s *Servicio) problemaPublico(w http.ResponseWriter, err error) {
	var inv ErrInvalido
	switch {
	case errors.As(err, &inv):
		problema.Escribir(w, problema.Invalido(inv.Motivo))
	case errors.Is(err, ErrNoExiste):
		problema.Escribir(w, problema.NoEncontrado("Esta tienda no existe o no está publicada."))
	case errors.Is(err, ErrSinConexion), errors.Is(err, ErrSinCredenciales), errors.Is(err, ErrApagado):
		problema.Escribir(w, problema.Conflicto("sin_correo", "Esta tienda no envía por Correo Argentino."))
	default:
		var eca *ErrorCA
		if errors.As(err, &eca) {
			s.log().Warn("correoargentino: no se pudo cotizar", "estado", eca.Estado, "mensaje", eca.Mensaje)
		} else {
			s.log().Error("correoargentino: cotizar", "error", err)
		}
		problema.Escribir(w, &problema.Problema{Estado: http.StatusBadGateway, Codigo: "correo",
			Titulo: "Sin cotización", Detalle: "No pudimos cotizar con Correo Argentino. Probá de nuevo en un rato."})
	}
}

func escribir(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(v)
}

func leer(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(v); err != nil {
		problema.Escribir(w, problema.Invalido("No entendimos el pedido."))
		return false
	}
	return true
}

// Rutas del panel y las públicas del checkout. ip: la del visitante (para el tope de cotizaciones).
func (s *Servicio) Rutas(mux *http.ServeMux, consultor limite.Consultor, ip func(*http.Request) string) {
	sesion := func(soloAdmin bool, h func(http.ResponseWriter, *http.Request, cuentas.Datos)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			d, ok := cuentas.SesionDe(r)
			if !ok || d.TiendaID == "" {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			if soloAdmin && d.Rol != "duenio" && d.Rol != "admin" {
				problema.Escribir(w, problema.SinPermiso("Sólo la dueña o quien administra puede hacerlo."))
				return
			}
			h(w, r, d)
		}
	}
	numero := func(w http.ResponseWriter, r *http.Request) (int, bool) {
		n, err := strconv.Atoi(r.PathValue("numero"))
		if err != nil || n < 1 {
			s.Problema(w, ErrNoExiste)
			return 0, false
		}
		return n, true
	}

	mux.HandleFunc("GET /v1/panel/integraciones/correo", sesion(false, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		e, err := s.EstadoDe(r.Context(), d.TiendaID)
		if err != nil {
			s.Problema(w, err)
			return
		}
		escribir(w, http.StatusOK, e)
	}))
	mux.HandleFunc("POST /v1/panel/integraciones/correo", sesion(true, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		var a Alta
		if !leer(w, r, &a) {
			return
		}
		e, err := s.Conectar(r.Context(), d.TiendaID, a)
		if err != nil {
			s.Problema(w, err)
			return
		}
		escribir(w, http.StatusOK, e)
	}))
	mux.HandleFunc("PUT /v1/panel/integraciones/correo", sesion(true, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		var b struct {
			Origen  Origen  `json:"origen"`
			Paquete Paquete `json:"paquete"`
		}
		if !leer(w, r, &b) {
			return
		}
		e, err := s.Ajustar(r.Context(), d.TiendaID, b.Origen, b.Paquete)
		if err != nil {
			s.Problema(w, err)
			return
		}
		escribir(w, http.StatusOK, e)
	}))
	mux.HandleFunc("DELETE /v1/panel/integraciones/correo", sesion(true, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		s.Problema(w, s.Desconectar(r.Context(), d.TiendaID))
	}))
	mux.HandleFunc("GET /v1/panel/pedidos/{numero}/correo", sesion(false, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		n, ok := numero(w, r)
		if !ok {
			return
		}
		e, err := s.EnvioDe(r.Context(), d.TiendaID, n)
		if err != nil {
			s.Problema(w, err)
			return
		}
		escribir(w, http.StatusOK, e)
	}))
	mux.HandleFunc("POST /v1/panel/pedidos/{numero}/correo", sesion(false, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		n, ok := numero(w, r)
		if !ok {
			return
		}
		e, err := s.Generar(r.Context(), d.TiendaID, n)
		if err != nil {
			s.Problema(w, err)
			return
		}
		escribir(w, http.StatusOK, e)
	}))
	mux.HandleFunc("PUT /v1/panel/pedidos/{numero}/correo/seguimiento", sesion(false, func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		n, ok := numero(w, r)
		if !ok {
			return
		}
		var b struct {
			Seguimiento string `json:"seguimiento"`
		}
		if !leer(w, r, &b) {
			return
		}
		e, err := s.Seguimiento(r.Context(), d.TiendaID, n, b.Seguimiento)
		if err != nil {
			s.Problema(w, err)
			return
		}
		escribir(w, http.StatusOK, e)
	}))

	// Públicas (las pide el BFF del checkout). Cada una puede pegarle a Correo: tope por IP.
	tope := func(w http.ResponseWriter, r *http.Request) bool {
		if consultor == nil || ip == nil {
			return true
		}
		if ok, err := limite.Permitido(r.Context(), consultor, "tienda:correo:"+ip(r), 120, time.Hour); err == nil && !ok {
			w.WriteHeader(http.StatusTooManyRequests)
			return false
		}
		return true
	}
	mux.HandleFunc("GET /v1/publico/tiendas/{slug}/correo/cotizacion", func(w http.ResponseWriter, r *http.Request) {
		if !tope(w, r) {
			return
		}
		u, _ := strconv.Atoi(r.URL.Query().Get("unidades"))
		if u < 1 || u > 99*50 {
			u = 1
		}
		c, err := s.CotizarPublico(r.Context(), r.PathValue("slug"), r.URL.Query().Get("cp"), u)
		if err != nil {
			s.problemaPublico(w, err)
			return
		}
		escribir(w, http.StatusOK, c)
	})
	mux.HandleFunc("GET /v1/publico/tiendas/{slug}/correo/sucursales", func(w http.ResponseWriter, r *http.Request) {
		if !tope(w, r) {
			return
		}
		l, err := s.SucursalesPublico(r.Context(), r.PathValue("slug"), r.URL.Query().Get("provincia"), r.URL.Query().Get("cp"))
		if err != nil {
			s.problemaPublico(w, err)
			return
		}
		escribir(w, http.StatusOK, l)
	})
}
