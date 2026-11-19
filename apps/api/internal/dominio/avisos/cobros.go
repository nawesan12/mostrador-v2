package avisos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net/url"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cobros"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/roles"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// cobrosCambio manda a la dueña (al email de su cuenta: el del negocio lo pudo cambiar quien hizo
// el cambio) el aviso de que cambió adónde llega la plata, con el botón para volverlo atrás y cerrar
// todas las sesiones. Sale siempre, también si el cambio lo hizo ella.
func (s *Servicio) cobrosCambio(ctx context.Context, ev outbox.Evento) error {
	var d cobros.AvisoCambio
	if json.Unmarshal(ev.Datos, &d) != nil || d.Token == "" || s.Correo == nil {
		return nil
	}
	var tienda, para string
	err := s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT t.nombre, u.email::text FROM tiendas t
			JOIN auth.membresias m ON m.tienda_id = t.id AND m.rol = $2
			JOIN auth.usuarios u ON u.id = m.usuario_id WHERE t.id = $1 LIMIT 1`, ev.TiendaID, roles.Duenia).Scan(&tienda, &para)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if !emailValido(para) {
		return nil
	}
	m := mailCobrosCambio(s.Sitio, tienda, d)
	m.Para = para
	return s.Correo.Enviar(ctx, m)
}

func mailCobrosCambio(sitio, tienda string, d cobros.AvisoCambio) correo.Mensaje {
	quien := d.Quien
	if quien == "" {
		quien = "Alguien de tu equipo"
	}
	var titulo string
	var filas [][3]string // dato, antes, ahora
	if d.Tipo == cobros.CambioMercadoPago {
		titulo = "Cambió la cuenta de Mercado Pago de " + tienda
		filas = append(filas, [3]string{"Cuenta", vacio(d.Antes["cuenta"]), vacio(d.Despues["cuenta"])})
	} else {
		titulo = "Cambiaron los datos para transferencias de " + tienda
		for _, k := range [][2]string{{"alias", "Alias"}, {"cbu", "CBU o CVU"}, {"titular", "Titular"}} {
			if d.Antes[k[0]] != d.Despues[k[0]] {
				filas = append(filas, [3]string{k[1], vacio(d.Antes[k[0]]), vacio(d.Despues[k[0]])})
			}
		}
	}
	link := strings.TrimRight(sitio, "/") + "/revertir-cobro?token=" + url.QueryEscape(d.Token)
	intro := quien + " lo cambió recién. Desde ahora, la plata de tus ventas va adonde dice «Ahora»."
	cierre := "Si no fuiste vos ni nadie de tu equipo, tocá el botón: lo volvemos atrás y cerramos todas las sesiones. El link vale 7 días."
	var txt strings.Builder
	txt.WriteString(titulo + "\n\n" + intro + "\n\n")
	for _, f := range filas {
		fmt.Fprintf(&txt, "%s: %s → %s\n", f[0], f[1], f[2])
	}
	txt.WriteString("\n" + cierre + "\nNo fui yo: " + link + "\n\nTu Local Web")

	var b strings.Builder
	fmt.Fprintf(&b, `<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:Helvetica,Arial,sans-serif;color:#2A140C"><div style="max-width:560px;margin:0 auto;padding:32px 24px"><div style="font-size:20px;font-weight:700">Tu Local Web</div><h1 style="margin:24px 0 0;font-size:26px;line-height:1.15;font-weight:600">%s</h1>`,
		html.EscapeString(titulo))
	fmt.Fprintf(&b, `<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">%s</p>`, sinViuda(intro))
	b.WriteString(`<table style="margin:18px 0 0;border-collapse:collapse;font-size:14.5px;width:100%"><tr style="color:#7A6252"><td style="padding:6px 0"></td><td style="padding:6px 8px">Antes</td><td style="padding:6px 0">Ahora</td></tr>`)
	for _, f := range filas {
		fmt.Fprintf(&b, `<tr style="border-top:1px solid #EADFD3"><td style="padding:8px 0;color:#7A6252">%s</td><td style="padding:8px">%s</td><td style="padding:8px 0;font-weight:600">%s</td></tr>`,
			html.EscapeString(f[0]), html.EscapeString(f[1]), html.EscapeString(f[2]))
	}
	b.WriteString(`</table>`)
	fmt.Fprintf(&b, `<p style="margin:20px 0 0;font-size:14.5px;line-height:1.5;color:#5C4536">%s</p>`, sinViuda(cierre))
	fmt.Fprintf(&b, `<p style="margin:18px 0 0"><a href="%s" style="display:inline-block;padding:12px 18px;border-radius:999px;background:#B0441D;color:#fff;text-decoration:none;font-weight:600">No fui yo: volver atrás</a></p>`, html.EscapeString(link))
	b.WriteString(`<p style="margin:28px 0 0;font-size:12px;line-height:1.5;color:#9C8676">Tu Local Web. Este aviso sale siempre que cambia adónde llega tu plata.</p></div></body></html>`)
	return correo.Mensaje{Asunto: titulo, Texto: txt.String(), HTML: b.String()}
}

func vacio(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}
