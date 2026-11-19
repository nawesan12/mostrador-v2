package planes

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Rutas: las rutas del panel que necesitan una función del plan, con el mismo patrón con que se
// registran. Sólo las que crean o prenden algo: mirar, apagar, desconectar, borrar y los ajustes de
// un canal ya conectado (el recargo de Mercado Libre, pausar sin stock) siguen andando
// después de bajar de plan (nada se pierde y se puede liberar). Los topes con cuenta (productos,
// personas, fotos) y los que dependen del cuerpo (venta en el local, roles) los aplica cada dominio.
var Rutas = map[string]Funcion{
	"POST /v1/panel/dominios":                              Dominio,
	"PUT /v1/panel/dominios/{dominio}":                     Dominio,
	"PUT /v1/panel/arca/ajustes":                           ARCA,
	"PUT /v1/panel/arca/certificado":                       ARCA,
	"POST /v1/panel/arca/pedido":                           ARCA,
	"POST /v1/panel/arca/probar":                           ARCA,
	"POST /v1/panel/integraciones/correo":                  Correo,
	"PUT /v1/panel/integraciones/correo":                   Correo,
	"POST /v1/panel/pedidos/{numero}/correo":               Correo,
	"POST /v1/panel/cupones":                               Cupones,
	"PUT /v1/panel/carritos":                               Cupones,
	"PUT /v1/panel/carritos/recordatorios":                 Cupones,
	"POST /v1/panel/carritos/recordar":                     Cupones,
	"POST /v1/panel/carritos/{id}/recordar":                Cupones,
	"POST /v1/panel/pedidos/{numero}/qr":                   Local,
	"POST /v1/panel/integraciones/mercadopago/caja":        Local,
	"POST /v1/panel/integraciones/mercadolibre/conectar":   MercadoLibre,
	"POST /v1/panel/integraciones/mercadolibre/vuelta":     MercadoLibre,
	"POST /v1/panel/productos/{id}/mercadolibre":           MercadoLibre,
	"POST /v1/panel/mercadolibre/preguntas/{id}/respuesta": MercadoLibre,
	"POST /v1/panel/newsletter/envios":                     Newsletter,
	"GET /v1/panel/informes":                               Newsletter,
	"PUT /v1/panel/configuracion/mayorista":                Mayorista,
}

type pool interface {
	EnTienda(ctx context.Context, tiendaID string, fn func(pgx.Tx) error) error
}

type marca Funcion

func (marca) ServeHTTP(http.ResponseWriter, *http.Request) {}

// Sesion: la tienda, la persona y su rol (ok false sin sesión o sin tienda: sigue de largo).
type Sesion func(*http.Request) (tiendaID, usuarioID, rol string, ok bool)

// Guardia aplica el plan a cada pedido de /v1/panel/: la función que pide la ruta y, para lo que
// cambia algo, que la persona no haya quedado fuera del equipo del plan. Sólo consulta la base
// cuando hace falta (la dueña en una ruta sin función no cuesta nada).
func Guardia(db pool, sesion Sesion, next http.Handler) http.Handler {
	mux := http.NewServeMux()
	for patron, f := range Rutas {
		mux.Handle(patron, marca(f))
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/panel/") {
			next.ServeHTTP(w, r)
			return
		}
		tienda, usuario, rol, ok := sesion(r)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}
		var f Funcion
		if h, _ := mux.Handler(r); h != nil {
			if m, ok := h.(marca); ok {
				f = Funcion(m)
			}
		}
		cambia := r.Method != http.MethodGet && r.Method != http.MethodHead
		// Lo que cualquiera hace sobre sí misma (sus sesiones, su push) no se frena.
		persona := cambia && rol != "duenio" && usuario != "" && !dePersona(r.URL.Path)
		if f == "" && !persona {
			next.ServeHTTP(w, r)
			return
		}
		var tope error
		err := db.EnTienda(r.Context(), tienda, func(t pgx.Tx) error {
			p, err := Vigente(r.Context(), t, tienda, time.Now())
			if err != nil {
				return err
			}
			if f != "" {
				if tope = ExigirFuncion(p, f); tope != nil {
					return nil
				}
			}
			if persona {
				fuera, err := PersonaFuera(r.Context(), t, tienda, usuario, p)
				if err != nil {
					return err
				}
				if fuera {
					tope = personaFuera(p)
				}
			}
			return nil
		})
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		if p := problema.De(tope); p != nil {
			problema.Escribir(w, p)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// dePersona: rutas de la propia persona (sesiones, dispositivos, notificaciones, lo que ve).
func dePersona(ruta string) bool {
	for _, p := range []string{"/v1/panel/sesiones", "/v1/panel/dispositivos", "/v1/panel/push-web", "/v1/panel/notificaciones",
		"/v1/panel/avisos/", "/v1/panel/configuracion/seguridad/", "/v1/panel/para-hacer"} {
		if strings.HasPrefix(ruta, p) {
			return true
		}
	}
	return false
}
