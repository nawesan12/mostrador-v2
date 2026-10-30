package cuentas

import (
	"fmt"
	"html"
	"strings"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
)

// mailRecuperar: el link para crear una contraseña nueva. Colores y tono de la marca; el handoff no
// diseña mails transaccionales (sí las firmas de Papelería, que se suman cuando estén).
func mailRecuperar(para, nombre, link string) correo.Mensaje {
	n := html.EscapeString(nombre)
	l := html.EscapeString(link)
	return correo.Mensaje{
		Para:   para,
		Asunto: "Creá tu nueva contraseña de Tu Local Web",
		Texto: fmt.Sprintf("Hola, %s.\n\nPara crear una contraseña nueva entrá a este link (vence en 30 minutos):\n%s\n\n"+
			"Si no fuiste vos, ignorá este mail: tu contraseña sigue siendo la misma.\n\nTu Local Web", nombre, link),
		HTML: fmt.Sprintf(`<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:'Instrument Sans',Helvetica,Arial,sans-serif;color:#2A140C">
<div style="max-width:560px;margin:0 auto;padding:32px 24px">
<div style="font-size:22px;font-weight:700;letter-spacing:-.03em">Tu Local Web</div>
<h1 style="margin:28px 0 0;font-size:28px;line-height:1.1;letter-spacing:-.03em;font-weight:600">Creá tu nueva contraseña</h1>
<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">Hola, %s. Tocá el botón para elegir una contraseña nueva. El link vence en 30 minutos.</p>
<p style="margin:24px 0"><a href="%s" style="display:inline-block;padding:14px 22px;border-radius:999px;background:#D2582F;color:#fff;font-weight:600;text-decoration:none">Crear contraseña nueva</a></p>
<p style="margin:0;font-size:13.5px;line-height:1.5;color:#7A6252">Si no fuiste vos, ignorá este mail: tu contraseña sigue siendo la misma.</p>
</div></body></html>`, n, l),
	}
}

// mailBienvenida sale al crear la cuenta (con email y contraseña o con un proveedor). Misma plantilla
// y remitente que el de la contraseña.
func mailBienvenida(para, nombre, tienda, link string) correo.Mensaje {
	n := html.EscapeString(nombre)
	t := html.EscapeString(tienda)
	l := html.EscapeString(link)
	return correo.Mensaje{
		Para:   para,
		Asunto: "Bienvenida a Tu Local Web",
		Texto: fmt.Sprintf("Hola, %s.\n\nYa creaste %s en Tu Local Web. Tenés 14 días gratis para probar todo, sin tarjeta.\n\n"+
			"Seguí armando tu tienda desde acá:\n%s\n\nSi no fuiste vos, respondé este mail y lo revisamos.\n\nTu Local Web", nombre, tienda, link),
		HTML: fmt.Sprintf(`<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:'Instrument Sans',Helvetica,Arial,sans-serif;color:#2A140C">
<div style="max-width:560px;margin:0 auto;padding:32px 24px">
<div style="font-size:22px;font-weight:700;letter-spacing:-.03em">Tu Local Web</div>
<h1 style="margin:28px 0 0;font-size:28px;line-height:1.1;letter-spacing:-.03em;font-weight:600">¡Te damos la bienvenida!</h1>
<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">Hola, %s. Ya creaste <b>%s</b> en Tu Local Web. Tenés 14 días gratis para probar todo, sin&nbsp;tarjeta.</p>
<p style="margin:24px 0"><a href="%s" style="display:inline-block;padding:14px 22px;border-radius:999px;background:#D2582F;color:#fff;font-weight:600;text-decoration:none">Seguir armando mi tienda</a></p>
<p style="margin:0;font-size:13.5px;line-height:1.5;color:#7A6252">Si no fuiste vos, respondé este mail y lo&nbsp;revisamos.</p>
</div></body></html>`, n, t, l),
	}
}

// mailIngreso: «Entraron a tu cuenta», cada vez que se abre una sesión nueva (no en el registro).
// Misma plantilla que los otros mails de cuentas. Lo que no se sabe (lugar, IP) no se muestra.
func mailIngreso(d ingreso, link string) correo.Mensaje {
	filas := [][2]string{{"Dispositivo", d.Aparato}}
	if d.Ubicacion != "" {
		filas = append(filas, [2]string{"Lugar aproximado", d.Ubicacion})
	}
	if d.IP != "" {
		filas = append(filas, [2]string{"IP", d.IP})
	}
	filas = append(filas, [2]string{"Hora", d.Hora + " (hora de Argentina)"})
	var txt, tabla strings.Builder
	for _, f := range filas {
		fmt.Fprintf(&txt, "%s: %s\n", f[0], f[1])
		fmt.Fprintf(&tabla, `<tr><td style="padding:6px 16px 6px 0;color:#7A6252;white-space:nowrap;vertical-align:top">%s</td><td style="padding:6px 0;font-weight:600">%s</td></tr>`,
			html.EscapeString(f[0]), html.EscapeString(f[1]))
	}
	n := html.EscapeString(d.Nombre)
	l := html.EscapeString(link)
	return correo.Mensaje{
		Para:   d.Para,
		Asunto: "Entraron a tu cuenta de Tu Local Web",
		Texto: fmt.Sprintf("Hola, %s.\n\nAlguien entró a tu cuenta de Tu Local Web:\n\n%s\n"+
			"Si fuiste vos, no tenés que hacer nada.\n\nSi no fuiste vos, cerrá esa sesión desde Configuración › Seguridad y cambiá tu contraseña:\n%s\n\nTu Local Web",
			d.Nombre, txt.String(), link),
		HTML: fmt.Sprintf(`<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:'Instrument Sans',Helvetica,Arial,sans-serif;color:#2A140C">
<div style="max-width:560px;margin:0 auto;padding:32px 24px">
<div style="font-size:22px;font-weight:700;letter-spacing:-.03em">Tu Local Web</div>
<h1 style="margin:28px 0 0;font-size:28px;line-height:1.1;letter-spacing:-.03em;font-weight:600">Entraron a tu&nbsp;cuenta</h1>
<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">Hola, %s. Alguien entró a tu cuenta de Tu Local Web. Si fuiste vos, no tenés que hacer&nbsp;nada.</p>
<table style="margin:20px 0 0;border-collapse:collapse;font-size:14.5px;line-height:1.4">%s</table>
<p style="margin:24px 0"><a href="%s" style="display:inline-block;padding:14px 22px;border-radius:999px;background:#D2582F;color:#fff;font-weight:600;text-decoration:none">Revisar mis sesiones</a></p>
<p style="margin:0;font-size:13.5px;line-height:1.5;color:#7A6252">Si no fuiste vos, cerrá esa sesión desde Configuración › Seguridad y cambiá tu&nbsp;contraseña.</p>
</div></body></html>`, n, tabla.String(), l),
	}
}
