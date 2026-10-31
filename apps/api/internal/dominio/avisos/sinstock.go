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

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/roles"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// sinStock manda «Te quedaste sin stock» (evento "tienda.sin_stock", que encola Go al descontar,
// uno por producto por día). Va a la dueña (al email del negocio o, si no hay, al de su cuenta) y a
// quien del equipo tiene Productos (al de su cuenta). No sale si la tienda apagó Sin stock › Email,
// ni si al mandarlo el producto ya no está en 0 o se pausó (docs/decisiones.md §41).
func (s *Servicio) sinStock(ctx context.Context, ev outbox.Evento) error {
	var d pedidos.SinStock
	if json.Unmarshal(ev.Datos, &d) != nil || d.Producto == "" || s.Correo == nil {
		return nil
	}
	var a pedidos.Agotado
	var tienda string
	var prendido bool
	var para []string
	err := s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		var err error
		if a, err = pedidos.LeerSinStock(ctx, t, d); err != nil {
			return err
		}
		var negocio string
		if err := t.QueryRow(ctx, `SELECT nombre, coalesce(negocio->>'email', ''),
				coalesce(preferencias->'canales'->'matriz'->'sinstock:email', 'true') <> 'false'
			FROM tiendas WHERE id = $1`, ev.TiendaID).Scan(&tienda, &negocio, &prendido); err != nil {
			return err
		}
		filas, err := t.Query(ctx, `SELECT m.rol, u.email::text FROM auth.membresias m
			JOIN auth.usuarios u ON u.id = m.usuario_id WHERE m.tienda_id = $1 ORDER BY m.creada_en`, ev.TiendaID)
		if err != nil {
			return err
		}
		defer filas.Close()
		visto := map[string]bool{}
		for filas.Next() {
			var rol, email string
			if err := filas.Scan(&rol, &email); err != nil {
				return err
			}
			if !roles.Puede(rol, roles.Productos) {
				continue
			}
			if rol == roles.Duenia && strings.TrimSpace(negocio) != "" {
				email = strings.TrimSpace(negocio)
			}
			if k := strings.ToLower(email); emailValido(email) && !visto[k] {
				visto[k] = true
				para = append(para, email)
			}
		}
		return filas.Err()
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if !prendido || !a.Sigue {
		return nil
	}
	// Un mail por persona. Si alguno no sale y ninguno salió, el outbox reintenta; si ya salió alguno,
	// reintentar se lo mandaría dos veces a quien ya lo tiene: se anota y sigue.
	var fallo error
	enviados := 0
	for _, p := range para {
		m := mailSinStock(s.Sitio, tienda, d.Producto, a)
		m.Para = p
		if err := s.Correo.Enviar(ctx, m); err != nil {
			fallo = err
			continue
		}
		enviados++
	}
	if enviados == 0 {
		return fallo
	}
	return nil
}

// mailSinStock arma «Te quedaste sin stock», con la plantilla de los mails de la tienda y el botón
// para reponer en el panel.
func mailSinStock(sitio, tienda, producto string, a pedidos.Agotado) correo.Mensaje {
	nombre := a.Nombre
	if a.Variante != "" {
		nombre += " (" + a.Variante + ")"
	}
	asunto := "Te quedaste sin stock de " + nombre
	titulo := "Te quedaste sin stock"
	parrafos := []string{
		fmt.Sprintf("Se vendió la última unidad de %s.", nombre),
		"En la tienda online figura «Sin stock» y nadie lo puede comprar hasta que repongas.",
	}
	if a.Variante != "" {
		parrafos[1] = "En la tienda online esa variante figura «Sin stock» y nadie la puede comprar hasta que repongas."
	}
	base := strings.TrimRight(sitio, "/")
	link := base + "/panel/productos?editar=" + url.QueryEscape(producto)
	ajustes := base + "/panel/notificaciones"

	txt := titulo + "\n\n" + strings.Join(parrafos, "\n\n") + "\n\nReponer stock: " + link +
		"\n\nTe llega porque tenés prendido «Sin stock» por email. Lo cambiás en " + ajustes + "\n\nTu Local Web"

	var b strings.Builder
	fmt.Fprintf(&b, `<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:Helvetica,Arial,sans-serif;color:#2A140C"><div style="max-width:560px;margin:0 auto;padding:32px 24px"><div style="font-size:20px;font-weight:700">%s</div><h1 style="margin:24px 0 0;font-size:26px;line-height:1.15;font-weight:600">%s</h1>`,
		html.EscapeString(tienda), html.EscapeString(titulo))
	for _, x := range parrafos {
		fmt.Fprintf(&b, `<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">%s</p>`, sinViuda(x))
	}
	fmt.Fprintf(&b, `<p style="margin:22px 0 0"><a href="%s" style="display:inline-block;padding:12px 18px;border-radius:999px;background:#2A140C;color:#fff;text-decoration:none;font-weight:600">Reponer stock</a></p>`, html.EscapeString(link))
	fmt.Fprintf(&b, `<p style="margin:28px 0 0;font-size:12px;line-height:1.5;color:#9C8676">Te llega porque tenés prendido «Sin stock» por email. Lo cambiás en <a href="%s" style="color:#9C8676">Notificaciones</a>. Tu Local Web.</p></div></body></html>`, html.EscapeString(ajustes))

	return correo.Mensaje{Asunto: asunto, Texto: txt, HTML: b.String()}
}
