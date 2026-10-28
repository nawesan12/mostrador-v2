package avisos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// preguntaRespondida le avisa a quien preguntó en la ficha de un producto que la tienda le respondió
// (evento "pregunta.respondida" que encola ventas.ResponderPregunta la primera vez que se responde).
func (s *Servicio) preguntaRespondida(ctx context.Context, ev outbox.Evento) error {
	var d struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(ev.Datos, &d); err != nil || s.Correo == nil {
		return nil
	}
	var nombre, email, pregunta, respuesta, producto, productoID, tienda, slug string
	err := s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT q.nombre, coalesce(q.email, ''), q.pregunta, coalesce(q.respuesta, ''), p.nombre, p.id::text,
				ti.nombre, ti.slug
			FROM preguntas q JOIN productos p ON p.tienda_id = q.tienda_id AND p.id = q.producto_id
			JOIN tiendas ti ON ti.id = q.tienda_id WHERE q.id::text = $1`, d.ID).
			Scan(&nombre, &email, &pregunta, &respuesta, &producto, &productoID, &tienda, &slug)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if email == "" || respuesta == "" {
		return nil
	}
	return s.Correo.Enviar(ctx, mailRespuesta(s.Sitio, nombre, email, pregunta, respuesta, producto, productoID, tienda, slug))
}

func mailRespuesta(sitio, nombre, email, pregunta, respuesta, producto, productoID, tienda, slug string) correo.Mensaje {
	nombre = strings.Split(strings.TrimSpace(nombre)+" ", " ")[0]
	link := ""
	if sitio != "" {
		link = fmt.Sprintf("%s/t/%s/producto/%s", strings.TrimRight(sitio, "/"), slug, productoID)
	}
	asunto := fmt.Sprintf("%s respondió tu pregunta sobre %s", tienda, producto)
	texto := fmt.Sprintf("Hola, %s.\n\nPreguntaste: %s\n\n%s te respondió: %s", nombre, pregunta, tienda, respuesta)
	if link != "" {
		texto += "\n\nVer el producto: " + link
	}
	var b strings.Builder
	fmt.Fprintf(&b, `<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:Helvetica,Arial,sans-serif;color:#2A140C"><div style="max-width:560px;margin:0 auto;padding:32px 24px"><div style="font-size:20px;font-weight:700">%s</div><h1 style="margin:24px 0 0;font-size:26px;line-height:1.15;font-weight:600">Te respondieron</h1>`,
		html.EscapeString(tienda))
	fmt.Fprintf(&b, `<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">Hola, %s. Sobre <b>%s</b> preguntaste:</p><p style="margin:10px 0 0;padding:12px 14px;border-radius:12px;background:#fff;font-size:15px">%s</p>`,
		html.EscapeString(nombre), html.EscapeString(producto), html.EscapeString(pregunta))
	fmt.Fprintf(&b, `<p style="margin:16px 0 0;font-size:15.5px;color:#5C4536">%s te respondió:</p><p style="margin:10px 0 0;padding:12px 14px;border-radius:12px;background:#fff;font-size:15px">%s</p>`,
		html.EscapeString(tienda), html.EscapeString(respuesta))
	if link != "" {
		fmt.Fprintf(&b, `<p style="margin:22px 0 0"><a href="%s" style="display:inline-block;padding:12px 18px;border-radius:999px;background:#2A140C;color:#fff;text-decoration:none;font-weight:600">Ver el producto</a></p>`, html.EscapeString(link))
	}
	b.WriteString(`<p style="margin:28px 0 0;font-size:12px;color:#9C8676">Tienda creada con Mostrador.</p></div></body></html>`)
	return correo.Mensaje{Para: email, Asunto: asunto, Texto: texto, HTML: b.String()}
}
