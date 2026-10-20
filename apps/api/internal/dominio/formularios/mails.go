package formularios

import (
	"fmt"
	"html"
	"strings"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
)

// Cómo se leen las opciones (los chips del handoff).
var (
	nombreFacturacion = map[string]string{"empiezo": "Recién empieza", "hasta-5": "Hasta $ 5 M", "5-a-20": "$ 5 a 20 M",
		"20-a-100": "$ 20 a 100 M", "mas-de-100": "Más de $ 100 M"}
	nombreOrigen = map[string]string{"tiendanube": "Tiendanube", "empretienda": "Empretienda", "mercadoshops": "Mercado Shops",
		"shopify": "Shopify", "woocommerce": "WooCommerce", "ninguna": "Ninguna"}
	nombreEspecialidad = map[string]string{"diseno": "Diseño", "desarrollo": "Desarrollo", "marketing": "Marketing",
		"contabilidad": "Contabilidad", "fotografia": "Fotografía"}
)

// plantilla: el mismo mail simple de marca que los de cuentas (el handoff no diseña mails).
func plantilla(titulo string, parrafos ...string) string {
	var b strings.Builder
	b.WriteString(`<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:'Instrument Sans',Helvetica,Arial,sans-serif;color:#2A140C">
<div style="max-width:560px;margin:0 auto;padding:32px 24px">
<div style="font-size:22px;font-weight:700;letter-spacing:-.03em">mostrador</div>
<h1 style="margin:28px 0 0;font-size:28px;line-height:1.1;letter-spacing:-.03em;font-weight:600">`)
	b.WriteString(html.EscapeString(titulo))
	b.WriteString(`</h1>`)
	for _, p := range parrafos {
		b.WriteString(`<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">` + p + `</p>`)
	}
	b.WriteString(`</div></body></html>`)
	return b.String()
}

func filas(pares ...[2]string) (string, string) {
	var t, h strings.Builder
	h.WriteString(`<table style="border-collapse:collapse;font-size:14.5px">`)
	for _, p := range pares {
		if p[1] == "" {
			continue
		}
		fmt.Fprintf(&t, "%s: %s\n", p[0], p[1])
		fmt.Fprintf(&h, `<tr><td style="padding:4px 12px 4px 0;color:#7A6252">%s</td><td style="padding:4px 0">%s</td></tr>`,
			html.EscapeString(p[0]), html.EscapeString(p[1]))
	}
	h.WriteString(`</table>`)
	return t.String(), h.String()
}

func mailConsultaRecibida(c Consulta) correo.Mensaje {
	n := html.EscapeString(c.Nombre)
	return correo.Mensaje{
		Para:   c.Email,
		Asunto: "Recibimos tu consulta",
		Texto: fmt.Sprintf("Hola, %s.\n\nGracias por escribirnos. Alguien del equipo de ventas te contacta en menos de 2 horas hábiles"+
			" (por WhatsApp si nos dejaste el número).\n\nMientras tanto, podés crear tu tienda gratis.\n\nMostrador", c.Nombre),
		HTML: plantilla("Recibimos tu consulta",
			fmt.Sprintf("Hola, %s. Gracias por escribirnos.", n),
			"Alguien del equipo de ventas te contacta en menos de 2 horas hábiles (por WhatsApp si nos dejaste el número). Te atiende una persona, no un call center.",
			"Mientras tanto, podés crear tu tienda gratis."),
	}
}

func mailConsultaEquipo(para string, c Consulta) correo.Mensaje {
	t, h := filas([2]string{"Nombre", c.Nombre}, [2]string{"Email", c.Email}, [2]string{"WhatsApp", c.Whatsapp},
		[2]string{"Web", c.Web}, [2]string{"Facturación", nombreFacturacion[c.Facturacion]},
		[2]string{"Viene de", nombreOrigen[c.Origen]}, [2]string{"Mensaje", c.Mensaje})
	return correo.Mensaje{
		Para:   para,
		Asunto: "Consulta de ventas: " + c.Nombre,
		Texto:  "Llegó una consulta desde /contacto.\n\n" + t,
		HTML:   plantilla("Consulta de ventas", "Llegó desde /contacto. Respondé en menos de 2 horas hábiles.", h),
	}
}

func mailPostulacionRecibida(p Postulacion) correo.Mensaje {
	n := html.EscapeString(p.Nombre)
	return correo.Mensaje{
		Para:   p.Email,
		Asunto: "Te postulaste al programa de partners",
		Texto: fmt.Sprintf("Hola, %s.\n\nRecibimos la postulación de %s al programa de partners de Mostrador."+
			" El acceso al panel de partners te llega a este email.\n\nMostrador", p.Nombre, p.Agencia),
		HTML: plantilla("¡Bienvenido al programa!",
			fmt.Sprintf("Hola, %s. Recibimos la postulación de <b style=\"color:#2A140C\">%s</b> al programa de partners de Mostrador.", n, html.EscapeString(p.Agencia)),
			"El acceso al panel de partners te llega a este email. Revisá también la carpeta de spam."),
	}
}

func mailPostulacionEquipo(para string, p Postulacion) correo.Mensaje {
	esp := make([]string, 0, len(p.Especialidades))
	for _, e := range p.Especialidades {
		esp = append(esp, nombreEspecialidad[e])
	}
	t, h := filas([2]string{"Nombre", p.Nombre}, [2]string{"Email", p.Email}, [2]string{"Agencia", p.Agencia},
		[2]string{"Web", p.Web}, [2]string{"Hace", strings.Join(esp, ", ")})
	return correo.Mensaje{
		Para:   para,
		Asunto: "Postulación de partner: " + p.Agencia,
		Texto:  "Llegó una postulación desde /partners.\n\n" + t,
		HTML:   plantilla("Postulación de partner", "Llegó desde /partners.", h),
	}
}

func mailSuscripcion(para, linkBaja string) correo.Mensaje {
	l := html.EscapeString(linkBaja)
	return correo.Mensaje{
		Para:   para,
		Asunto: "Te sumaste a las novedades de Mostrador",
		Texto: "Listo: una vez por mes te mandamos un resumen de lo nuevo en Mostrador. Nada más.\n\n" +
			"Si no fuiste vos o no lo querés más, date de baja acá: " + linkBaja + "\n\nMostrador",
		HTML: plantilla("Te sumaste a las novedades",
			"Una vez por mes te mandamos un resumen de lo nuevo en Mostrador. Nada más.",
			fmt.Sprintf(`<span style="font-size:13.5px;color:#7A6252">Si no fuiste vos o no lo querés más, <a href="%s" style="color:#C24E25">date de baja</a>.</span>`, l)),
	}
}
