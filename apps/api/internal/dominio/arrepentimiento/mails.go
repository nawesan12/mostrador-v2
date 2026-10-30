package arrepentimiento

import (
	"fmt"
	"html"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
)

// mailRecibido confirma el pedido con su código, con el mismo tono que la pantalla de éxito.
func mailRecibido(para, nombre, codigo, plan string) correo.Mensaje {
	n, c, p := html.EscapeString(nombre), html.EscapeString(codigo), html.EscapeString(plan)
	return correo.Mensaje{
		Para:   para,
		Asunto: "Recibimos tu pedido de arrepentimiento · " + codigo,
		Texto: fmt.Sprintf("Hola, %s.\n\nRecibimos tu pedido de revocación del plan %s.\nTu código de trámite es %s.\n\n"+
			"Cancelamos el débito automático y te devolvemos el total en hasta 10 días hábiles, por el mismo medio de pago.\n\n"+
			"Si no fuiste vos, respondé este mail.\n\nTu Local Web", nombre, plan, codigo),
		HTML: fmt.Sprintf(`<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:'Instrument Sans',Helvetica,Arial,sans-serif;color:#2A140C">
<div style="max-width:560px;margin:0 auto;padding:32px 24px">
<div style="font-size:22px;font-weight:700;letter-spacing:-.03em">Tu Local Web</div>
<h1 style="margin:28px 0 0;font-size:28px;line-height:1.1;letter-spacing:-.03em;font-weight:600">Recibimos tu pedido</h1>
<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">Hola, %s. Registramos la revocación del plan <b style="color:#2A140C">%s</b>.</p>
<p style="margin:18px 0 0;padding:14px 16px;border-radius:14px;background:#fff;font-size:15px">Código de trámite: <b>%s</b></p>
<p style="margin:18px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">Cancelamos el débito automático y te devolvemos el total en hasta 10 días hábiles, por el mismo medio de pago.</p>
<p style="margin:18px 0 0;font-size:13.5px;line-height:1.5;color:#7A6252">Si no fuiste vos, respondé este mail.</p>
</div></body></html>`, n, p, c),
	}
}

// mailEquipo avisa al equipo que hay una devolución para procesar (plazo legal: 10 días hábiles).
func mailEquipo(para, codigo string, p Pedido) correo.Mensaje {
	return correo.Mensaje{
		Para:   para,
		Asunto: "Arrepentimiento " + codigo + ": " + p.Nombre,
		Texto: fmt.Sprintf("Pedido de revocación %s.\n\nEmail: %s\nNombre: %s\nDocumento: %s\nPlan: %s\nMotivo: %s\n\n"+
			"Cancelar el débito y devolver el total en hasta 10 días hábiles.", codigo, p.Email, p.Nombre, p.Documento, p.Plan, p.Motivo),
		HTML: fmt.Sprintf(`<!doctype html><html lang="es-AR"><body style="font-family:Helvetica,Arial,sans-serif;color:#2A140C">
<h1 style="font-size:22px">Pedido de revocación %s</h1>
<p>Email: %s<br>Nombre: %s<br>Documento: %s<br>Plan: %s<br>Motivo: %s</p>
<p>Cancelar el débito y devolver el total en hasta 10 días hábiles.</p></body></html>`,
			html.EscapeString(codigo), html.EscapeString(p.Email), html.EscapeString(p.Nombre), html.EscapeString(p.Documento),
			html.EscapeString(p.Plan), html.EscapeString(p.Motivo)),
	}
}
