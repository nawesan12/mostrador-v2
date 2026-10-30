package avisos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// venta es lo que la dueña necesita saber de una venta nueva de la tienda online.
type venta struct {
	numero                      int
	tienda                      string
	para                        string // el email del negocio o el de la cuenta dueña
	pago, estadoPago, envio     string
	comprador                   pedidos.Comprador
	nota                        string
	entrega                     pedidos.Entrega
	items                       []item
	subtotal, envioCosto, total int64
	descuentos                  []pedidos.Descuento
}

// ventaNueva manda «Nueva venta #N» a la dueña (evento "tienda.venta", que encola el trigger
// app.avisar_venta_tienda: por transferencia al crearse, por Mercado Pago al acreditarse). Sale una
// sola vez (pedidos.avisos->'tienda_venta') y no sale si la tienda apagó Venta nueva › Email.
func (s *Servicio) ventaNueva(ctx context.Context, ev outbox.Evento) error {
	var d struct {
		Numero int `json:"numero"`
	}
	if json.Unmarshal(ev.Datos, &d) != nil || d.Numero <= 0 || s.Correo == nil {
		return nil
	}
	var v venta
	var ya, prendido bool
	var estado string
	err := s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		var items, comprador, entrega []byte
		var nota *string
		err := t.QueryRow(ctx, `SELECT p.numero, p.comprador, p.pago, p.estado_pago, p.estado, p.envio, p.entrega, p.nota,
				p.items, p.subtotal, p.costo_envio, p.total, p.descuentos, p.avisos ? 'tienda_venta', t.nombre,
				coalesce(t.preferencias->'canales'->'matriz'->'venta:email', 'true') <> 'false',
				coalesce(nullif(t.negocio->>'email', ''), (SELECT u.email::text FROM auth.membresias m
					JOIN auth.usuarios u ON u.id = m.usuario_id WHERE m.tienda_id = t.id AND m.rol = 'duenio'
					ORDER BY m.creada_en LIMIT 1), '')
			FROM pedidos p JOIN tiendas t ON t.id = p.tienda_id WHERE p.numero = $1`, d.Numero).
			Scan(&v.numero, &comprador, &v.pago, &v.estadoPago, &estado, &v.envio, &entrega, &nota,
				&items, &v.subtotal, &v.envioCosto, &v.total, &v.descuentos, &ya, &v.tienda, &prendido, &v.para)
		if err != nil {
			return err
		}
		_ = json.Unmarshal(comprador, &v.comprador)
		_ = json.Unmarshal(items, &v.items)
		if len(entrega) > 0 {
			_ = json.Unmarshal(entrega, &v.entrega)
		}
		if nota != nil {
			v.nota = *nota
		}
		v.descuentos = pedidos.DescuentosCompletos(v.descuentos, v.subtotal+v.envioCosto-v.total)
		return nil
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if ya || !prendido || estado == "Cancelado" || !emailValido(v.para) {
		return nil
	}
	if err := s.Correo.Enviar(ctx, mailVenta(s.Sitio, v)); err != nil {
		return err // el outbox reintenta con espera creciente
	}
	return s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE pedidos SET avisos = avisos || jsonb_build_object('tienda_venta', $2::text) WHERE numero = $1`,
			d.Numero, time.Now().UTC().Format(time.RFC3339))
		return err
	})
}

// emailValido: una sola dirección, sin saltos ni separadores (va a un encabezado).
func emailValido(e string) bool {
	return e != "" && !strings.ContainsAny(e, "\r\n<>,; ") && strings.Contains(e, "@")
}

// sinViuda escapa el texto y une las dos últimas palabras: ninguna queda sola en el último renglón.
func sinViuda(s string) string {
	e := html.EscapeString(s)
	if i := strings.LastIndex(e, " "); i > 0 {
		e = e[:i] + "&nbsp;" + e[i+1:]
	}
	return e
}

// entregaDe dice cómo recibe el pedido quien compró, con lo que quedó guardado.
func entregaDe(v venta) string {
	e := v.entrega
	lugar := strings.TrimSpace(e.Direccion)
	if l := strings.TrimSpace(e.Localidad); l != "" && lugar != "" {
		lugar += ", " + l
	}
	if cp := strings.TrimSpace(e.CP); cp != "" && lugar != "" {
		lugar += " (CP " + cp + ")"
	}
	switch v.envio {
	case "retiro":
		return "Retira en el local"
	case "moto":
		if lugar != "" {
			return "Envío en moto a " + lugar
		}
		return "Envío en moto"
	case "correo":
		if e.Tipo == "S" {
			if e.SucursalNombre != "" {
				return "Correo Argentino, retira en la sucursal " + e.SucursalNombre
			}
			return "Correo Argentino, retira en sucursal"
		}
		if lugar != "" {
			return "Correo Argentino a domicilio: " + lugar
		}
		return "Correo Argentino a domicilio"
	}
	return v.envio
}

// mailVenta arma «Nueva venta #N». La plantilla es la de los mails de pedidos; la respuesta le
// llega a quien compró.
func mailVenta(sitio string, v venta) correo.Mensaje {
	nombre := strings.TrimSpace(v.comprador.Nombre)
	if nombre == "" {
		nombre = "Alguien"
	}
	asunto := fmt.Sprintf("Nueva venta #%d · %s", v.numero, pesos(v.total))
	titulo := fmt.Sprintf("Nueva venta #%d", v.numero)
	var parrafos []string
	var medio string
	switch {
	case v.pago == "transf" && v.estadoPago == "Pendiente":
		medio = "Transferencia · pendiente de comprobante"
		parrafos = append(parrafos,
			fmt.Sprintf("%s compró en %s y eligió pagar por transferencia: todavía no está acreditada.", nombre, v.tienda),
			"Cuando veas la plata en tu cuenta, marcá el pedido como pagado. Quien compró puede mandarte el comprobante respondiendo el mail del pedido.")
	case v.pago == "transf":
		medio = "Transferencia · acreditada"
		parrafos = append(parrafos, fmt.Sprintf("%s compró en %s y la transferencia ya está acreditada.", nombre, v.tienda))
	case v.pago == "mp":
		medio = "Mercado Pago · acreditado"
		parrafos = append(parrafos, fmt.Sprintf("%s compró en %s y el pago por Mercado Pago ya se acreditó: podés empezar a prepararlo.", nombre, v.tienda))
	default:
		medio = v.pago
		parrafos = append(parrafos, fmt.Sprintf("%s compró en %s.", nombre, v.tienda))
	}

	var det []string
	for _, it := range v.items {
		n := it.Nombre
		if it.Variante != "" {
			n += " (" + it.Variante + ")"
		}
		det = append(det, fmt.Sprintf("%d × %s — %s", it.Cantidad, n, pesos(it.Precio*int64(it.Cantidad))))
	}
	if len(v.descuentos) > 0 || v.envioCosto > 0 {
		det = append(det, "Subtotal — "+pesos(v.subtotal))
	}
	if v.envioCosto > 0 {
		det = append(det, "Envío — "+pesos(v.envioCosto))
	}
	for _, d := range v.descuentos {
		det = append(det, d.Concepto+" — −"+pesos(d.Monto))
	}
	det = append(det, "Total — "+pesos(v.total))

	filas := [][2]string{{"Pago", medio}, {"Entrega", entregaDe(v)}, {"Nombre", nombre}}
	if v.comprador.Email != nil && *v.comprador.Email != "" {
		filas = append(filas, [2]string{"Email", *v.comprador.Email})
	}
	if v.comprador.Telefono != nil && *v.comprador.Telefono != "" {
		filas = append(filas, [2]string{"Teléfono", *v.comprador.Telefono})
	}
	// La nota de la tienda online trae el DNI y adónde se entrega; lo de la entrega ya está arriba.
	var nota []string
	for _, x := range strings.Split(v.nota, " · ") {
		if x = strings.TrimSpace(x); x != "" && !strings.HasPrefix(x, "Entrega: ") {
			nota = append(nota, x)
		}
	}
	if len(nota) > 0 {
		filas = append(filas, [2]string{"Nota", strings.Join(nota, " · ")})
	}

	base := strings.TrimRight(sitio, "/")
	link := fmt.Sprintf("%s/panel/pedidos/%d", base, v.numero)
	ajustes := base + "/panel/notificaciones"

	var txt strings.Builder
	txt.WriteString(titulo + "\n\n" + strings.Join(parrafos, "\n\n") + "\n\n" + strings.Join(det, "\n") + "\n\n")
	for _, f := range filas {
		fmt.Fprintf(&txt, "%s: %s\n", f[0], f[1])
	}
	fmt.Fprintf(&txt, "\nVer el pedido: %s\n\nTe llega porque tenés prendido «Venta nueva» por email. Lo cambiás en %s\n\nTu Local Web", link, ajustes)

	var b strings.Builder
	fmt.Fprintf(&b, `<!doctype html><html lang="es-AR"><body style="margin:0;background:#FBF7F2;font-family:Helvetica,Arial,sans-serif;color:#2A140C"><div style="max-width:560px;margin:0 auto;padding:32px 24px"><div style="font-size:20px;font-weight:700">%s</div><h1 style="margin:24px 0 0;font-size:26px;line-height:1.15;font-weight:600">%s</h1>`,
		html.EscapeString(v.tienda), html.EscapeString(titulo))
	for _, x := range parrafos {
		fmt.Fprintf(&b, `<p style="margin:12px 0 0;font-size:15.5px;line-height:1.5;color:#5C4536">%s</p>`, sinViuda(x))
	}
	b.WriteString(`<div style="margin:18px 0 0;padding:14px 16px;border-radius:14px;background:#fff;font-size:14.5px;line-height:1.7">`)
	for _, x := range det {
		fmt.Fprintf(&b, `<div>%s</div>`, html.EscapeString(x))
	}
	b.WriteString(`</div><table style="margin:18px 0 0;border-collapse:collapse;font-size:14.5px;line-height:1.4">`)
	for _, f := range filas {
		fmt.Fprintf(&b, `<tr><td style="padding:6px 16px 6px 0;color:#7A6252;white-space:nowrap;vertical-align:top">%s</td><td style="padding:6px 0;font-weight:600">%s</td></tr>`,
			html.EscapeString(f[0]), sinViuda(f[1]))
	}
	b.WriteString(`</table>`)
	fmt.Fprintf(&b, `<p style="margin:22px 0 0"><a href="%s" style="display:inline-block;padding:12px 18px;border-radius:999px;background:#2A140C;color:#fff;text-decoration:none;font-weight:600">Ver el pedido</a></p>`, html.EscapeString(link))
	fmt.Fprintf(&b, `<p style="margin:28px 0 0;font-size:12px;line-height:1.5;color:#9C8676">Te llega porque tenés prendido «Venta nueva» por email. Lo cambiás en <a href="%s" style="color:#9C8676">Notificaciones</a>. Tu Local Web.</p></div></body></html>`, html.EscapeString(ajustes))

	responder := ""
	if v.comprador.Email != nil && emailValido(*v.comprador.Email) {
		responder = *v.comprador.Email
	}
	return correo.Mensaje{Para: v.para, Asunto: asunto, Texto: txt.String(), HTML: b.String(), ResponderA: responder}
}
