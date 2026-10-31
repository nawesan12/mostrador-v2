package push

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/webpush"
)

// Push web del panel instalado como app (docs/decisiones.md §41): la misma lógica que el iPhone,
// con la suscripción de Web Push del navegador en vez del token de APNs. Se activa desde
// Notificaciones › «Push en este celular»; en el iPhone sólo anda con el panel instalado (iOS 16.4+).

// SuscripcionWeb es lo que manda el panel al activar los avisos (components/schemas/SuscripcionWeb).
type SuscripcionWeb struct {
	Endpoint string `json:"endpoint"`
	P256dh   string `json:"p256dh"`
	Auth     string `json:"auth"`
	// Aparato: cómo se ve en la lista («iPhone · Safari»).
	Aparato string `json:"aparato"`
}

// RegistrarWeb ata la suscripción a la sesión (sesion es el hash del token) y a la persona. Si el
// endpoint ya estaba con otra sesión, se mueve a esta.
func RegistrarWeb(ctx context.Context, p Pool, usuario, sesion string, s SuscripcionWeb) error {
	n, err := webpush.Normalizar(webpush.Suscripcion{Endpoint: s.Endpoint, P256dh: s.P256dh, Auth: s.Auth})
	if err != nil {
		return ErrInvalido
	}
	return p.EnUsuario(ctx, usuario, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `SELECT app.registrar_suscripcion_web($1, $2, $3, $4, $5)`,
			sesion, n.Endpoint, n.P256dh, n.Auth, limpiar(s.Aparato, 60))
		return err
	})
}

// BajaWeb borra las suscripciones de esta sesión.
func BajaWeb(ctx context.Context, p Pool, usuario, sesion string) error {
	return p.EnUsuario(ctx, usuario, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `DELETE FROM auth.suscripciones_web WHERE sesion_hash = $1`, sesion)
		return err
	})
}

// suscriptas cuenta las suscripciones de esta sesión.
func suscriptas(ctx context.Context, p Pool, usuario, sesion string) (int, error) {
	var n int
	err := p.EnUsuario(ctx, usuario, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT count(*) FROM auth.suscripciones_web WHERE sesion_hash = $1`, sesion).Scan(&n)
	})
	return n, err
}

// Prueba son los datos de cuenta.push_prueba: a qué sesión de qué persona mandar el aviso de prueba.
type Prueba struct {
	Usuario string `json:"usuario"`
	Sesion  string `json:"sesion"`
}

// EstadoWeb: GET /v1/panel/push-web. Clave vacía: el push web está apagado (sin VAPID).
type EstadoWeb struct {
	Clave     string `json:"clave"`
	Suscripta bool   `json:"suscripta"`
}

// RutasWeb: GET, POST y DELETE /v1/panel/push-web y POST /v1/panel/push-web/prueba (roles.Panel).
// clave es la pública VAPID (vacía si está apagado). Como los iPhone, sólo con la sesión de una
// persona: ni una IA conectada ni el soporte mirando la tienda.
func RutasWeb(mux *http.ServeMux, p Pool, clave string) {
	con := func(h func(http.ResponseWriter, *http.Request, cuentas.Datos)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			d, ok := cuentas.SesionDe(r)
			if !ok {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			if d.SesionHash == "" || d.IA != "" || d.Mirada {
				problema.Escribir(w, problema.SinPermiso("Los avisos al celular son de la sesión de una persona."))
				return
			}
			h(w, r, d)
		}
	}
	mux.HandleFunc("GET /v1/panel/push-web", con(func(w http.ResponseWriter, r *http.Request, s cuentas.Datos) {
		n, err := suscriptas(r.Context(), p, s.Usuario.ID, s.SesionHash)
		if err != nil {
			slog.Error("push web: estado", "error", err)
			problema.Escribir(w, problema.Interno())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(EstadoWeb{Clave: clave, Suscripta: n > 0})
	}))
	mux.HandleFunc("POST /v1/panel/push-web", con(func(w http.ResponseWriter, r *http.Request, s cuentas.Datos) {
		if clave == "" {
			problema.Escribir(w, problema.Conflicto("push_apagado", "Los avisos al celular todavía no están disponibles."))
			return
		}
		var d SuscripcionWeb
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<10)).Decode(&d); err != nil {
			problema.Escribir(w, problema.Invalido("Revisá los datos."))
			return
		}
		err := RegistrarWeb(r.Context(), p, s.Usuario.ID, s.SesionHash, d)
		switch {
		case errors.Is(err, ErrInvalido):
			problema.Escribir(w, problema.Invalido("La suscripción del navegador no es válida.",
				problema.Campo{Ruta: "endpoint", Mensaje: "Tiene que ser de un servicio de push conocido, con sus claves."}))
		case err != nil:
			slog.Error("push web: registrar", "error", err)
			problema.Escribir(w, problema.Interno())
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	mux.HandleFunc("DELETE /v1/panel/push-web", con(func(w http.ResponseWriter, r *http.Request, s cuentas.Datos) {
		if err := BajaWeb(r.Context(), p, s.Usuario.ID, s.SesionHash); err != nil {
			slog.Error("push web: baja", "error", err)
			problema.Escribir(w, problema.Interno())
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	// El aviso de prueba lo manda el worker (cuenta.push_prueba), como todos: acá sólo se encola.
	mux.HandleFunc("POST /v1/panel/push-web/prueba", con(func(w http.ResponseWriter, r *http.Request, s cuentas.Datos) {
		ctx := r.Context()
		if clave == "" {
			problema.Escribir(w, problema.Conflicto("push_apagado", "Los avisos al celular todavía no están disponibles."))
			return
		}
		n, err := suscriptas(ctx, p, s.Usuario.ID, s.SesionHash)
		if err == nil && n == 0 {
			problema.Escribir(w, problema.Conflicto("sin_suscripcion", "Primero activá los avisos en este celular."))
			return
		}
		if err == nil {
			b, _ := json.Marshal(Prueba{Usuario: s.Usuario.ID, Sesion: s.SesionHash})
			err = p.EnUsuario(ctx, s.Usuario.ID, func(t pgx.Tx) error {
				_, err := t.Exec(ctx, `INSERT INTO outbox (tienda_id, tipo, datos) VALUES (NULL, 'cuenta.push_prueba', $1)`, b)
				return err
			})
		}
		if err != nil {
			slog.Error("push web: prueba", "error", err)
			problema.Escribir(w, problema.Interno())
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}))
}

// URLPanel traduce la ruta de la app a la página del panel que abre el aviso.
func URLPanel(ruta string) string {
	switch {
	case strings.HasPrefix(ruta, "pedido/"):
		return "/panel/pedidos/" + strings.TrimPrefix(ruta, "pedido/")
	case strings.HasPrefix(ruta, "producto/"):
		return "/panel/productos/" + strings.TrimPrefix(ruta, "producto/")
	case strings.HasPrefix(ruta, "preguntas/"):
		return "/panel/resenas-y-preguntas"
	case ruta == "cuenta/seguridad":
		return "/panel/configuracion?tab=seguridad"
	case ruta == "notificaciones":
		return "/panel/notificaciones"
	}
	return "/panel"
}

// PayloadWeb arma lo que lee el service worker del panel (static/sw-panel.js): título, cuerpo, la
// página que abre, la tienda (para cambiar a esa antes de abrirla) y la etiqueta (dos avisos con la
// misma se ven como uno).
func PayloadWeb(a Aviso, etiqueta string) []byte {
	b, _ := json.Marshal(struct {
		Title  string `json:"title"`
		Body   string `json:"body"`
		URL    string `json:"url"`
		Tienda string `json:"tienda,omitempty"`
		Tag    string `json:"tag,omitempty"`
	}{a.Titulo, a.Cuerpo, URLPanel(a.Ruta), a.Tienda, etiqueta})
	return b
}

// AvisoPrueba: lo que llega al tocar «Mandame un aviso de prueba».
func AvisoPrueba() Aviso {
	return Aviso{Titulo: "Aviso de prueba", Cuerpo: "Así te van a llegar las ventas nuevas a este celular.",
		Hilo: "cuenta", Ruta: "notificaciones"}
}
