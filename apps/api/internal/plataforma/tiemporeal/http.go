package tiemporeal

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/coder/websocket"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Quien es la sesión de una conexión del panel, como la resuelve el paquete cuentas.
type Quien struct {
	Usuario string
	// Sesion: el id público de la sesión (prefijo del hash, el de Configuración › Seguridad).
	Sesion  string
	Tienda  string
	Rol     string
	VenceEn time.Time
	// IA: el token es de una IA conectada (MCP). No usa el tiempo real.
	IA bool
}

// ErrNoExiste: la sesión no vale, o la tienda o el pedido no existen.
var ErrNoExiste = errors.New("no existe")

// Config del endpoint.
type Config struct {
	Hub *Hub
	// Autenticar resuelve un token de sesión (ErrNoExiste si no vale).
	Autenticar func(ctx context.Context, token string) (Quien, error)
	// Publico resuelve la tienda de un slug y, si viene, el número del pedido con ese código
	// público (ErrNoExiste si no existen).
	Publico func(ctx context.Context, slug, codigo string) (tienda, pedido string, err error)
	// SlugDeDominio: la tienda de un dominio propio verificado (para aceptar su Origin).
	SlugDeDominio func(ctx context.Context, dominio string) (string, error)
	// Origenes: de dónde se acepta el panel ("https://tulocalweb.com.ar"; admite comodines de
	// path.Match, como "http://localhost:*" en desarrollo).
	Origenes []string
	// Desarrollo: acepta también las cookies sin el prefijo __Host- (http://localhost).
	Desarrollo bool
	// IP del navegador (detrás de Cloudflare y Caddy).
	IP  func(*http.Request) string
	Log *slog.Logger
	// Ping: cada cuánto se manda un ping (Cloudflare corta a los 100 s de silencio).
	Ping time.Duration
	// Revalidar: cada cuánto se vuelve a mirar la sesión (inactivas, quitada del equipo).
	Revalidar time.Duration
}

// Cookies del panel que valen para conectarse, en orden (la «mirada» del soporte primero: si
// está, el panel la usa a ella). Iguales a las de apps/web/src/lib/server/auth/sesion.ts y
// staff/index.ts.
var (
	cookiesProduccion = []string{"__Host-mostrador_mirada", "__Host-tlw_sesion", "__Host-mostrador_sesion"}
	cookiesDesarrollo = []string{"mostrador_mirada", "tlw_sesion", "mostrador_sesion"}
)

const (
	// Lo que puede mandar el navegador: nada en realidad (sólo pongs y el cierre).
	limiteEntrante  = 512
	esperaEscritura = 10 * time.Second
)

// Rutas monta GET /v1/tiempo-real (panel, con sesión) y GET /v1/tiempo-real/publico (tienda
// online, sin sesión).
func Rutas(mux *http.ServeMux, c Config) {
	if c.Log == nil {
		c.Log = slog.Default()
	}
	if c.Ping <= 0 {
		c.Ping = 25 * time.Second
	}
	if c.Revalidar <= 0 {
		c.Revalidar = 5 * time.Minute
	}
	if c.IP == nil {
		c.IP = func(r *http.Request) string { return r.RemoteAddr }
	}
	mux.HandleFunc("GET /v1/tiempo-real", c.panel)
	mux.HandleFunc("GET /v1/tiempo-real/publico", c.publico)
}

func (c Config) origenPermitido(origen string) bool {
	u, err := url.Parse(origen)
	if err != nil || u.Host == "" {
		return false
	}
	o := strings.ToLower(u.Scheme + "://" + u.Host)
	for _, p := range c.Origenes {
		if ok, _ := path.Match(strings.ToLower(p), o); ok {
			return true
		}
	}
	return false
}

// token: el Bearer (la app de iPhone) o la cookie del panel (el navegador).
func (c Config) token(r *http.Request) (token string, deCookie bool) {
	if t, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok && strings.TrimSpace(t) != "" {
		return strings.TrimSpace(t), false
	}
	nombres := cookiesProduccion
	if c.Desarrollo {
		nombres = append(append([]string{}, cookiesDesarrollo...), cookiesProduccion...)
	}
	for _, n := range nombres {
		if k, err := r.Cookie(n); err == nil && k.Value != "" {
			return k.Value, true
		}
	}
	return "", true
}

// aceptar hace el upgrade. Los plazos del http.Server (lectura y escritura) se sacan: esta
// conexión vive mucho más y tiene sus propios (ping y escritura con tope).
func aceptar(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	rc := http.NewResponseController(w)
	_ = rc.SetReadDeadline(time.Time{})
	_ = rc.SetWriteDeadline(time.Time{})
	// El Origin ya se controló antes (más estricto que el de la biblioteca).
	ws, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return nil, err
	}
	ws.SetReadLimit(limiteEntrante)
	return ws, nil
}

func (c Config) panel(w http.ResponseWriter, r *http.Request) {
	token, deCookie := c.token(r)
	origen := r.Header.Get("Origin")
	// Con la cookie, el Origin es obligatorio y tiene que ser del panel: si no, cualquier página
	// podría abrir la conexión con la cookie de quien la visita (CSWSH). Con Bearer no hay nada
	// ambiental que robar, pero si el navegador manda Origin también se controla.
	if (deCookie && !c.origenPermitido(origen)) || (!deCookie && origen != "" && !c.origenPermitido(origen)) {
		problema.Escribir(w, problema.SinPermiso("Origen no permitido."))
		return
	}
	ws, err := aceptar(w, r)
	if err != nil {
		return
	}
	defer ws.CloseNow()
	if token == "" {
		_ = ws.Close(CierreSinSesion, "sin sesión")
		return
	}
	q, err := c.Autenticar(r.Context(), token)
	switch {
	case errors.Is(err, ErrNoExiste) || (err == nil && q.Tienda == ""):
		_ = ws.Close(CierreSinSesion, "sin sesión")
		return
	case err != nil:
		c.Log.Error("tiempo real: resolver la sesión", "error", err)
		_ = ws.Close(websocket.StatusInternalError, "probá de nuevo")
		return
	case q.IA:
		_ = ws.Close(CierreSinPermiso, "sólo para personas")
		return
	}
	cli := &Cliente{Tienda: q.Tienda, Usuario: q.Usuario, Sesion: q.Sesion, Rol: q.Rol, IP: c.IP(r)}
	if err := c.Hub.Registrar(cli); err != nil {
		_ = ws.Close(CierreLimite, "demasiadas conexiones")
		return
	}
	defer c.Hub.Quitar(cli)
	c.servir(r.Context(), ws, cli, func(ctx context.Context) *Corte {
		q2, err := c.Autenticar(ctx, token)
		switch {
		case errors.Is(err, ErrNoExiste) || (err == nil && q2.Tienda == ""):
			return &Corte{CierreSinSesion, "sesión cerrada"}
		case err != nil:
			return nil // la base no responde: se sigue y se vuelve a mirar después
		case q2.Tienda != q.Tienda || q2.Rol != q.Rol:
			return &Corte{CierreResincronizar, "cambió la sesión"}
		}
		return nil
	}, q.VenceEn)
}

var (
	slugValido   = regexp.MustCompile(`^[a-z0-9-]{2,60}$`)
	codigoValido = regexp.MustCompile(`^[A-Za-z0-9_-]{16,64}$`)
)

func (c Config) publico(w http.ResponseWriter, r *http.Request) {
	slug, codigo := r.URL.Query().Get("tienda"), r.URL.Query().Get("pedido")
	origen := r.Header.Get("Origin")
	if origen != "" && !c.origenPermitido(origen) {
		// El dominio propio de esta misma tienda también vale.
		ok := false
		if u, err := url.Parse(origen); err == nil && c.SlugDeDominio != nil && slugValido.MatchString(slug) {
			if s, err := c.SlugDeDominio(r.Context(), u.Hostname()); err == nil && s == slug {
				ok = true
			}
		}
		if !ok {
			problema.Escribir(w, problema.SinPermiso("Origen no permitido."))
			return
		}
	}
	ws, err := aceptar(w, r)
	if err != nil {
		return
	}
	defer ws.CloseNow()
	if !slugValido.MatchString(slug) || (codigo != "" && !codigoValido.MatchString(codigo)) {
		_ = ws.Close(websocket.StatusPolicyViolation, "no existe")
		return
	}
	tienda, pedido, err := c.Publico(r.Context(), slug, codigo)
	if errors.Is(err, ErrNoExiste) {
		_ = ws.Close(websocket.StatusPolicyViolation, "no existe")
		return
	}
	if err != nil {
		c.Log.Error("tiempo real: resolver la tienda", "error", err)
		_ = ws.Close(websocket.StatusInternalError, "probá de nuevo")
		return
	}
	cli := &Cliente{Tienda: tienda, Pedido: pedido, Publico: true, IP: c.IP(r)}
	if err := c.Hub.Registrar(cli); err != nil {
		_ = ws.Close(CierreLimite, "demasiadas conexiones")
		return
	}
	defer c.Hub.Quitar(cli)
	c.servir(r.Context(), ws, cli, nil, time.Time{})
}

// servir atiende la conexión hasta que se cierre: manda lo que reparte el Hub, hace ping, corta
// cuando el Hub lo pide, al vencer la sesión o si al revalidarla ya no vale.
func (c Config) servir(ctx context.Context, ws *websocket.Conn, cli *Cliente, revalidar func(context.Context) *Corte, vence time.Time) {
	// El navegador no manda datos: CloseRead lee sólo control (pong, cierre) y corta si llega algo.
	ctx = ws.CloseRead(ctx)
	escribir := func(m Mensaje) error {
		b, _ := json.Marshal(m)
		wctx, cancel := context.WithTimeout(ctx, esperaEscritura)
		defer cancel()
		return ws.Write(wctx, websocket.MessageText, b)
	}
	if escribir(Mensaje{Tipo: TipoHola}) != nil {
		return
	}
	ping := time.NewTicker(c.Ping)
	defer ping.Stop()
	var revisar <-chan time.Time
	if revalidar != nil {
		t := time.NewTicker(c.Revalidar)
		defer t.Stop()
		revisar = t.C
	}
	var vencio <-chan time.Time
	if !vence.IsZero() {
		t := time.NewTimer(time.Until(vence))
		defer t.Stop()
		vencio = t.C
	}
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-cli.Envios():
			if escribir(m) != nil {
				return
			}
		case k := <-cli.Cortes():
			_ = ws.Close(k.Codigo, k.Motivo)
			return
		case <-ping.C:
			pctx, cancel := context.WithTimeout(ctx, min(esperaEscritura, c.Ping))
			err := ws.Ping(pctx)
			cancel()
			// Además un latido de texto: el navegador no ve los pings del protocolo y así se da
			// cuenta solo de una conexión muerta (la compu se durmió, cambió la red).
			if err != nil || escribir(Mensaje{Tipo: TipoLatido}) != nil {
				return
			}
		case <-revisar:
			rctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			k := revalidar(rctx)
			cancel()
			if k != nil {
				_ = ws.Close(k.Codigo, k.Motivo)
				return
			}
		case <-vencio:
			_ = ws.Close(CierreSinSesion, "la sesión venció")
			return
		}
	}
}
