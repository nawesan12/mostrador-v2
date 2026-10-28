// Package servidor arma el http.Handler de la API: rutas de salud y, más adelante, los dominios.
package servidor

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/arrepentimiento"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/configuracion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/crecer"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cupones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/devoluciones"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/equipo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/facturacion"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/finanzas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/formularios"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/hoy"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/ia"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/links"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/mitienda"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/partners"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/precios"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tiendas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/turnos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/ventas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/vidriera"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadolibre"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadopago"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/telegram"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/limite"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Dependencias del servidor. App puede ser nil en tests que no tocan la base.
type Dependencias struct {
	App *db.Pool
	Log *slog.Logger
	// Cuentas y sesiones. Sin esto, el panel responde 401.
	Cuentas  *cuentas.Servicio
	ClaveBFF string
	// Botón de arrepentimiento (público, sin sesión).
	Arrepentimiento *arrepentimiento.Servicio
	// Formularios públicos del sitio: contacto, partners y novedades.
	Formularios *formularios.Servicio
	// Sitio público (para armar la dirección de las tiendas).
	Sitio string
	// Mercado Pago: conexión de las tiendas, cobros, abono y webhook. Sin credenciales, responde 409.
	MercadoPago *mercadopago.Servicio
	// Mercado Libre como canal: conexión, publicaciones, pedidos y preguntas.
	MercadoLibre *mercadolibre.Servicio
	// IA conectadas por MCP (OAuth 2.1). Sin esto, los tokens de IA no entran.
	IA *ia.Servicio
	// Facturación con ARCA (comprobantes, certificado, factura manual).
	Facturacion *facturacion.Servicio
	// Correo para los avisos de la tienda (arrepentimiento). nil: no se manda.
	Correo correo.Correo
	// Hornero en Telegram: vínculo del chat (panel) y lo que pide el bot por el BFF.
	Telegram *telegram.Servicio
	// Backoffice del equipo (/v1/equipo/*) y "ver la tienda" de sólo lectura. nil: no se monta.
	Equipo *equipo.Servicio
}

// Nuevo devuelve el handler raíz de la API.
func Nuevo(d Dependencias) http.Handler {
	if d.Log == nil {
		d.Log = slog.Default()
	}
	mux := http.NewServeMux()
	// Vivo: el proceso responde. No toca la base, para que un corte de Postgres no reinicie la API.
	mux.HandleFunc("GET /salud/vivo", func(w http.ResponseWriter, _ *http.Request) {
		escribirJSON(w, http.StatusOK, map[string]string{"estado": "vivo"})
	})
	// Listo: puede atender pedidos (la base responde). El balanceador saca la instancia si no.
	mux.HandleFunc("GET /salud/listo", func(w http.ResponseWriter, r *http.Request) {
		if d.App == nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := d.App.Ping(ctx); err != nil {
			d.Log.Warn("salud: la base no responde", "error", err)
			w.Header().Set("Retry-After", "5")
			problema.Escribir(w, &problema.Problema{Estado: http.StatusServiceUnavailable, Codigo: "sin_base",
				Titulo: "No disponible", Detalle: "La base de datos no responde."})
			return
		}
		escribirJSON(w, http.StatusOK, map[string]string{"estado": "listo"})
	})
	var h http.Handler = mux
	if d.App != nil {
		// La tienda del pedido sale siempre de la sesión: nunca de algo que diga el cliente.
		tienda := func(*http.Request) string { return "" }
		if d.Cuentas != nil {
			d.Cuentas.Rutas(mux, d.ClaveBFF)
			tienda = cuentas.TiendaDe
			var conGuardia http.Handler = mux
			if d.IA != nil {
				d.Cuentas.SesionIA = d.IA.ParaCuentas
				d.IA.Rutas(mux, func(r *http.Request) string { return cuentas.ClienteDe(r, d.ClaveBFF).IP })
				conGuardia = d.IA.Guardia(mux)
			}
			if d.Equipo != nil {
				d.Cuentas.SesionEquipo = d.Equipo.ParaCuentas
				conGuardia = equipo.SoloLectura(conGuardia)
			}
			h = d.Cuentas.Autenticar(conGuardia)
		}
		tiendas.Rutas(mux, d.App, tienda)
		pedidos.Rutas(mux, d.App, tienda)
		var fact listas.Facturas
		if d.Facturacion != nil {
			fact = facturacion.ParaListas{S: d.Facturacion}
			d.Facturacion.Rutas(mux)
		}
		listas.Rutas(mux, d.App, tienda, fact)
		var preguntas ventas.Preguntas
		if d.MercadoLibre != nil {
			preguntas = d.MercadoLibre
		}
		ventas.Rutas(mux, d.App, tienda, preguntas)
		var pagar links.Pagar
		if d.MercadoPago != nil {
			pagar = d.MercadoPago.PagarLink
		}
		links.Rutas(mux, d.App, tienda, pagar, d.MercadoPago.Problema)
		precios.Rutas(mux, d.App, tienda)
		cupones.Rutas(mux, d.App, tienda)
		crecer.Rutas(mux, d.App, tienda)
		hoy.Rutas(mux, d.App, tienda)
		finanzas.Rutas(mux, d.App, tienda)
		mitienda.Rutas(mux, d.App, tienda, d.Sitio)
		partners.Rutas(mux, d.App, d.ClaveBFF, nil)
		configuracion.Rutas(mux, d.App, d.Sitio, d.MercadoPago != nil && d.MercadoPago.Config.TokenPlataforma != "")
		if d.MercadoPago != nil {
			d.MercadoPago.Rutas(mux)
		}
		if d.MercadoLibre != nil {
			d.MercadoLibre.Rutas(mux)
		}
		var pagador vidriera.Pagador
		if d.MercadoPago != nil {
			pagador = d.MercadoPago
		}
		var cobrador turnos.Cobrador
		if d.MercadoPago != nil {
			cobrador = d.MercadoPago
		}
		turnos.Rutas(mux, d.App, turnos.Config{Tienda: tienda, MP: cobrador, Sitio: d.Sitio, Log: d.Log,
			IP: func(r *http.Request) string { return cuentas.ClienteDe(r, d.ClaveBFF).IP },
			Permitido: func(ctx context.Context, clave string, max int, ventana time.Duration) (bool, error) {
				return limite.Permitido(ctx, d.App, clave, max, ventana)
			},
			Admin: func(r *http.Request) bool {
				s, ok := cuentas.SesionDe(r)
				return ok && (s.Rol == "duenio" || s.Rol == "admin")
			}})
		if d.Telegram != nil {
			d.Telegram.Rutas(mux, d.ClaveBFF)
		}
		ip := func(r *http.Request) string { return cuentas.ClienteDe(r, d.ClaveBFF).IP }
		dev := &devoluciones.Servicio{Pool: d.App, Correo: d.Correo}
		dev.Rutas(mux, tienda)
		vidriera.Rutas(mux, d.App, pagador, d.Sitio, dev, ip)
		if d.Arrepentimiento != nil {
			d.Arrepentimiento.Rutas(mux, ip)
		}
		if d.Formularios != nil {
			d.Formularios.Rutas(mux, ip)
		}
		if d.Equipo != nil {
			d.Equipo.Rutas(mux)
		}
	}
	return recuperar(d.Log, h)
}

// recuperar convierte un panic en un 500 con problem+json y lo registra, sin tirar el proceso.
func recuperar(log *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				if v == http.ErrAbortHandler {
					panic(v)
				}
				log.Error("panic en un handler", "ruta", r.URL.Path, "valor", v)
				problema.Escribir(w, problema.Interno())
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func escribirJSON(w http.ResponseWriter, estado int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(estado)
	_ = json.NewEncoder(w).Encode(v)
}
