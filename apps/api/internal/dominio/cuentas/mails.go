package cuentas

import (
	"fmt"
	"html"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
)

// mailRecuperar: el link para crear una contraseña nueva. Colores y tono de la marca; el handoff no
// diseña mails transaccionales (sí las firmas de Papelería, que se suman cuando estén).
func mailRecuperar(para, nombre, link string) correo.Mensaje {
	n := html.EscapeString(nombre)
	l := html.EscapeString(link)
	return correo.Mensaje{
		Para:   para,
		Asunto: "Creá tu nueva contraseña de Mostrador",
		Texto: fmt.Sprintf("Hola, %s.\n\nPara crear una contraseña nueva entrá a este link (vence en 30 minutos):\n%s\n\n"+
			"Si no fuiste vos, ignorá este mail: tu contraseña sigue siendo la misma.\n\nMostrador", nombre, link),
		HTML: fmt.Sprintf(`<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:'Instrument Sans',Helvetica,Arial,sans-serif;color:#2A140C">
<div style="max-width:560px;margin:0 auto;padding:32px 24px">
<div style="font-size:22px;font-weight:700;letter-spacing:-.03em">mostrador</div>
<h1 style="margin:28px 0 0;font-size:28px;line-height:1.1;letter-spacing:-.03em;font-weight:600">Creá tu nueva contraseña</h1>
<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">Hola, %s. Tocá el botón para elegir una contraseña nueva. El link vence en 30 minutos.</p>
<p style="margin:24px 0"><a href="%s" style="display:inline-block;padding:14px 22px;border-radius:999px;background:#D2582F;color:#fff;font-weight:600;text-decoration:none">Crear contraseña nueva</a></p>
<p style="margin:0;font-size:13.5px;line-height:1.5;color:#7A6252">Si no fuiste vos, ignorá este mail: tu contraseña sigue siendo la misma.</p>
</div></body></html>`, n, l),
	}
}
