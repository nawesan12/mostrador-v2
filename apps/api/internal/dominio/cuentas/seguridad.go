package cuentas

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/netip"
	"strings"
	"time"
	// La zona de Argentina viaja en el binario: el VPS puede no tener tzdata.
	_ "time/tzdata"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// Inactividad: con «Cerrar sesiones inactivas» prendido, una sesión que no se usó en este tiempo
// deja de valer.
const Inactividad = 7 * 24 * time.Hour

// SeguridadDef son las opciones de seguridad de la persona (auth.usuarios.seguridad) y su valor si
// nunca las tocó: el aviso por mail de cada ingreso, prendido; cerrar las inactivas, apagado.
var SeguridadDef = map[string]bool{"avisoIngreso": true, "inactivas": false}

var zonaAR = func() *time.Location {
	z, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	if err != nil {
		return time.FixedZone("ART", -3*60*60)
	}
	return z
}()

// Aparato es lo que se sabe del dispositivo por el agente (User-Agent).
type Aparato struct {
	// Tipo: compu, celu o tablet (el ícono).
	Tipo      string
	Sistema   string
	Navegador string
}

// Nombre: «Chrome en Mac», «App de Tu Local Web en iPhone».
func (a Aparato) Nombre() string {
	switch {
	case a.Navegador != "" && a.Sistema != "":
		return a.Navegador + " en " + a.Sistema
	case a.Navegador != "":
		return a.Navegador
	case a.Sistema != "":
		return a.Sistema
	}
	return "Dispositivo sin identificar"
}

// Dispositivo lee el agente. Lo que no se reconoce queda vacío (no se adivina).
func Dispositivo(agente string) Aparato {
	a := strings.ToLower(agente)
	has := func(s ...string) bool {
		for _, x := range s {
			if strings.Contains(a, x) {
				return true
			}
		}
		return false
	}
	ap := Aparato{Tipo: "compu"}
	switch {
	case has("ipad"):
		ap.Tipo, ap.Sistema = "tablet", "iPad"
	case has("iphone", "ipod"):
		ap.Tipo, ap.Sistema = "celu", "iPhone"
	case has("android"):
		ap.Tipo, ap.Sistema = "tablet", "Android"
		if has("mobile") {
			ap.Tipo = "celu"
		}
	case has("cros"):
		ap.Sistema = "ChromeOS"
	case has("macintosh", "mac os"):
		ap.Sistema = "Mac"
	case has("windows"):
		ap.Sistema = "Windows"
	case has("linux"):
		ap.Sistema = "Linux"
	}
	switch {
	// Las apps nativas: la de iPhone sale con el agente de CFNetwork; la de Android, con el de
	// OkHttp o Dalvik.
	case has("cfnetwork", "tulocalweb"):
		ap.Navegador = "App de Tu Local Web"
		if ap.Sistema == "" || ap.Sistema == "Mac" {
			ap.Tipo, ap.Sistema = "celu", "iPhone"
		}
	case has("okhttp", "dalvik"):
		ap.Navegador = "App de Tu Local Web"
		if ap.Sistema == "" || ap.Sistema == "Linux" {
			ap.Tipo, ap.Sistema = "celu", "Android"
		}
	case has("edg/", "edga/", "edgios/"):
		ap.Navegador = "Edge"
	case has("opr/", "opera"):
		ap.Navegador = "Opera"
	case has("samsungbrowser"):
		ap.Navegador = "Samsung Internet"
	case has("crios", "chrome"):
		ap.Navegador = "Chrome"
	case has("fxios", "firefox"):
		ap.Navegador = "Firefox"
	case has("safari"):
		ap.Navegador = "Safari"
	}
	return ap
}

// IPCorta: la IPv4 entera; de la IPv6, los primeros cuatro grupos (lo demás no le dice nada a una
// persona). Vacía si no es una IP.
func IPCorta(ip string) string {
	a, err := netip.ParseAddr(strings.TrimSpace(ip))
	if err != nil {
		return ""
	}
	a = a.Unmap()
	if a.Is4() {
		return a.String()
	}
	g := strings.Split(a.StringExpanded(), ":")
	for i := range g[:4] {
		g[i] = strings.TrimLeft(g[i], "0")
		if g[i] == "" {
			g[i] = "0"
		}
	}
	return strings.Join(g[:4], ":") + ":…"
}

// HoraAR: «30/10/2026 a las 14:05».
func HoraAR(t time.Time) string {
	return t.In(zonaAR).Format("02/01/2006 a las 15:04")
}

// ingreso son los datos del mail «Entraron a tu cuenta» (evento cuenta.ingreso del outbox).
type ingreso struct {
	Para      string `json:"para"`
	Nombre    string `json:"nombre"`
	Aparato   string `json:"aparato"`
	Ubicacion string `json:"ubicacion"`
	IP        string `json:"ip"`
	Hora      string `json:"hora"`
}

// avisarIngreso encola el mail de la sesión nueva, en la misma transacción, si la persona no lo
// apagó. El evento va sin tienda: es de la persona, no de una tienda (migración 00051). Con el mismo
// interruptor sale el push «Entraron a tu cuenta» a sus iPhone (cuenta.ingreso_push, §40), menos al
// de la sesión nueva (sesion es el hash de su token).
func (s *Servicio) avisarIngreso(ctx context.Context, tx pgx.Tx, usuario, sesion string, cl Cliente) error {
	var d ingreso
	var avisar bool
	if err := tx.QueryRow(ctx, `SELECT email, nombre, coalesce((seguridad->>'avisoIngreso')::boolean, true)
		FROM auth.usuarios WHERE id = $1`, usuario).Scan(&d.Para, &d.Nombre, &avisar); err != nil {
		return err
	}
	if !avisar {
		return nil
	}
	d.Aparato = Dispositivo(cl.Agente).Nombre()
	if s.Geo != nil {
		d.Ubicacion = s.Geo.Ubicar(cl.IP)
	}
	d.IP = IPCorta(cl.IP)
	d.Hora = HoraAR(time.Now())
	b, err := json.Marshal(d)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO outbox (tienda_id, tipo, datos) VALUES (NULL, 'cuenta.ingreso', $1)`, b); err != nil {
		return err
	}
	p, err := json.Marshal(IngresoPush{Usuario: usuario, Sesion: sesion, Aparato: d.Aparato})
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `INSERT INTO outbox (tienda_id, tipo, datos) VALUES (NULL, 'cuenta.ingreso_push', $1)`, p)
	return err
}

// IngresoPush son los datos de cuenta.ingreso_push: a quién, qué sesión no avisar (la nueva) y desde
// qué aparato se entró. Sin IP ni lugar: se ve en la pantalla bloqueada.
type IngresoPush struct {
	Usuario string `json:"usuario"`
	Sesion  string `json:"sesion"`
	Aparato string `json:"aparato"`
}

// Manejadores del worker: el mail de cada ingreso y el de las invitaciones al equipo. sitio arma el link a Configuración › Seguridad.
func Manejadores(c correo.Correo, sitio string) map[string]outbox.Manejador {
	return map[string]outbox.Manejador{
		"cuenta.ingreso": func(ctx context.Context, e outbox.Evento) error {
			var d ingreso
			if err := json.Unmarshal(e.Datos, &d); err != nil || d.Para == "" {
				slog.Error("cuenta.ingreso: datos ilegibles", "evento", e.ID, "error", err)
				return nil // reintentar no lo arregla
			}
			return c.Enviar(ctx, mailIngreso(d, sitio+"/panel/configuracion?tab=seguridad"))
		},
		// La invitación al equipo de una tienda (configuracion.Invitar y Reenviar). El link viene armado.
		"cuenta.invitacion": func(ctx context.Context, e outbox.Evento) error {
			var d MailInvitacion
			if err := json.Unmarshal(e.Datos, &d); err != nil || d.Para == "" || d.Link == "" {
				slog.Error("cuenta.invitacion: datos ilegibles", "evento", e.ID, "error", err)
				return nil
			}
			return c.Enviar(ctx, mailInvitacion(d))
		},
	}
}
