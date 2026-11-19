package tincho

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/ia"
	topes "github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/limite"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// plazoEscritura: el de la respuesta, con margen sobre tiempoPregunta.
const plazoEscritura = tiempoPregunta + 15*time.Second

func escribir(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_ = json.NewEncoder(w).Encode(v)
}

func leer(w http.ResponseWriter, r *http.Request, v any) bool {
	if json.NewDecoder(http.MaxBytesReader(w, r.Body, 32<<10)).Decode(v) != nil {
		problema.Escribir(w, problema.Invalido("El pedido no es válido."))
		return false
	}
	return true
}

func (s *Servicio) escribirError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrNoDisponible) {
		w.Header().Set("Retry-After", "30")
		problema.Escribir(w, &problema.Problema{Estado: http.StatusServiceUnavailable, Codigo: "ia_no_disponible",
			Titulo: "No disponible", Detalle: "No pude consultar a la IA, probá de nuevo."})
		return
	}
	s.log().Error("tincho", "error", err)
	problema.Escribir(w, problema.Interno())
}

// Pregunta del panel o la app: la última pregunta, la charla anterior (corta) y el formato.
type pregunta struct {
	Pregunta  string `json:"pregunta"`
	Historial []struct {
		Rol   string `json:"rol"`
		Texto string `json:"texto"`
	} `json:"historial"`
}

func (p pregunta) consulta() (Consulta, bool) {
	q := Consulta{Pregunta: strings.TrimSpace(p.Pregunta)}
	if q.Pregunta == "" || utf8.RuneCountInString(q.Pregunta) > 1000 || len(p.Historial) > 24 {
		return q, false
	}
	for _, t := range p.Historial {
		rol := "user"
		if t.Rol == "hornero" || t.Rol == "assistant" {
			rol = "assistant"
		}
		q.Historial = append(q.Historial, ia.Turno{Rol: rol, Texto: t.Texto})
	}
	return q, true
}

// Rutas: las del panel (cualquiera del equipo: roles.Panel; cada herramienta mira su permiso) y la
// de Telegram (sólo el BFF, con su clave).
func (s *Servicio) Rutas(mux *http.ServeMux, claveBFF string) {
	panel := func(h func(http.ResponseWriter, *http.Request, cuentas.Datos)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			d, ok := cuentas.SesionDe(r)
			// Una IA conectada por MCP no usa a Tincho (sería IA llamando a IA con el cupo de la tienda).
			if !ok || d.TiendaID == "" || d.IA != "" {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			// Un tope por sesión (unos 20 por minuto): un script con una sesión no vacía el cupo de la tienda.
			if d.SesionHash != "" {
				if ok, err := topes.Permitido(r.Context(), s.Pool, "tincho:sesion:"+d.SesionHash, 20, time.Minute); err == nil && !ok {
					w.Header().Set("Retry-After", "60")
					problema.Escribir(w, &problema.Problema{Estado: http.StatusTooManyRequests, Codigo: "demasiados_intentos",
						Titulo: "Esperá un momento", Detalle: "Le hiciste muchas preguntas seguidas a Tincho. Probá en un minuto."})
					return
				}
			}
			// Hasta 4 vueltas con el modelo: más que el WriteTimeout general de la API.
			_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(plazoEscritura))
			h(w, r, d)
		}
	}
	s.rutasConfirmaciones(mux, panel)
	mux.HandleFunc("POST /v1/panel/tincho/preguntar", panel(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		var p pregunta
		if !leer(w, r, &p) {
			return
		}
		q, ok := p.consulta()
		if !ok {
			problema.Escribir(w, problema.Invalido("Escribile algo a Tincho (hasta 1.000 letras)."))
			return
		}
		q.Tienda, q.Rol, q.Canal = d.TiendaID, d.Rol, "panel"
		res, err := s.Preguntar(r.Context(), q)
		if err != nil {
			s.escribirError(w, err)
			return
		}
		escribir(w, res)
	}))
	mux.HandleFunc("POST /v1/panel/tincho/mejorar", panel(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		var b struct {
			Nombre      string `json:"nombre"`
			Descripcion string `json:"descripcion"`
		}
		if !leer(w, r, &b) {
			return
		}
		b.Nombre = strings.TrimSpace(b.Nombre)
		if n := utf8.RuneCountInString(b.Nombre); n < 2 || n > 120 || utf8.RuneCountInString(b.Descripcion) > 4000 {
			problema.Escribir(w, problema.Invalido("Falta el nombre del producto."))
			return
		}
		t, err := tiendas.Obtener(r.Context(), s.Pool, d.TiendaID)
		if err != nil {
			s.escribirError(w, err)
			return
		}
		res, err := s.Mejorar(r.Context(), d.TiendaID, t.Nombre, t.Rubro, b.Nombre, b.Descripcion)
		if err != nil {
			s.escribirError(w, err)
			return
		}
		escribir(w, res)
	}))
	// El motor del BFF respondió sin IA: se cuenta para Plan y facturación.
	mux.HandleFunc("POST /v1/panel/tincho/motor", panel(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		if err := s.ContarMotor(r.Context(), d.TiendaID); err != nil {
			s.escribirError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))

	// Telegram: el webhook lo atiende el BFF (ya verificó el secreto de Telegram). Responde como la
	// dueña (quien vinculó el chat), en texto plano y sin historial.
	mux.HandleFunc("POST /v1/telegram/preguntar", func(w http.ResponseWriter, r *http.Request) {
		if !cuentas.EsBFF(r, claveBFF) {
			problema.Escribir(w, problema.SinPermiso("Sólo el BFF de Tu Local Web puede llamar a esta ruta."))
			return
		}
		_ = http.NewResponseController(w).SetWriteDeadline(time.Now().Add(plazoEscritura))
		var b struct {
			Chat     int64  `json:"chat"`
			Pregunta string `json:"pregunta"`
			Motor    bool   `json:"motor"`
		}
		if !leer(w, r, &b) {
			return
		}
		if s.RolDelChat == nil {
			problema.Escribir(w, problema.NoEncontrado("Este chat no está vinculado a ninguna tienda."))
			return
		}
		tienda, rol, err := s.RolDelChat(r.Context(), b.Chat)
		if err != nil {
			problema.Escribir(w, problema.NoEncontrado("Este chat ya no tiene acceso a la tienda."))
			return
		}
		// motor: el BFF ya respondió sin IA; sólo se cuenta.
		if b.Motor {
			if err := s.ContarMotor(r.Context(), tienda); err != nil {
				s.escribirError(w, err)
				return
			}
			w.WriteHeader(http.StatusNoContent)
			return
		}
		q, ok := pregunta{Pregunta: b.Pregunta}.consulta()
		if !ok {
			problema.Escribir(w, problema.Invalido("El mensaje está vacío o es muy largo."))
			return
		}
		// Con el rol de quien vinculó el chat (no siempre la dueña).
		q.Tienda, q.Rol, q.Canal = tienda, rol, "telegram"
		res, err := s.Preguntar(r.Context(), q)
		if err != nil {
			s.escribirError(w, err)
			return
		}
		escribir(w, res)
	})
}
