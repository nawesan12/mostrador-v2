package push

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/roles"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/apns"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// Enviador es lo que hace falta de APNs (*apns.Cliente lo cumple; en los tests, uno falso).
type Enviador interface {
	Enviar(context.Context, apns.Envio) error
}

// Servicio manda los avisos del outbox. Sin APNs (nil), los manejadores no hacen nada.
type Servicio struct {
	Pool Pool
	APNs Enviador
	Log  *slog.Logger
}

// Nuevo arma el servicio. Con cli nil (sin MOSTRADOR_APNS_*) el envío queda apagado y lo dice el log
// una sola vez, al arrancar.
func Nuevo(p Pool, cli *apns.Cliente, log *slog.Logger) *Servicio {
	if log == nil {
		log = slog.Default()
	}
	s := &Servicio{Pool: p, Log: log}
	if cli == nil {
		log.Info("push: APNs sin configurar (MOSTRADOR_APNS_*), no se mandan avisos al iPhone")
		return s
	}
	s.APNs = cli
	return s
}

func (s *Servicio) activo() bool {
	if s == nil || s.APNs == nil {
		return false
	}
	if c, ok := s.APNs.(*apns.Cliente); ok && c == nil {
		return false
	}
	return true
}

// Manejadores del outbox.
func (s *Servicio) Manejadores() map[string]outbox.Manejador {
	return map[string]outbox.Manejador{
		"push.venta":          s.venta,
		"push.pregunta_ml":    s.preguntaML,
		"cuenta.ingreso_push": s.ingreso,
		"push.sin_stock":      s.sinStock,
	}
}

// AvisoSinStock: «Te quedaste sin stock» con el producto (y la variante), y la ruta para reponer.
func AvisoSinStock(tienda, producto string, a pedidos.Agotado) Aviso {
	nombre := a.Nombre
	if a.Variante != "" {
		nombre += " (" + a.Variante + ")"
	}
	return Aviso{Titulo: "Te quedaste sin stock", Cuerpo: "Se vendió la última unidad de " + nombre + ". Tocá para reponer.",
		Hilo: HiloTienda(tienda), Ruta: "producto/" + producto, Tienda: tienda}
}

// sinStock: push.sin_stock, que encola Go al descontar (uno por producto por día, junto con el mail
// tienda.sin_stock). A quien tiene Productos; no sale si la tienda apagó Sin stock › Celular ni si
// el producto ya no está en 0 o se pausó.
func (s *Servicio) sinStock(ctx context.Context, ev outbox.Evento) error {
	if !s.activo() {
		return nil
	}
	var d pedidos.SinStock
	if json.Unmarshal(ev.Datos, &d) != nil || d.Producto == "" {
		return nil
	}
	var a pedidos.Agotado
	var prendido bool
	err := s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		var err error
		if a, err = pedidos.LeerSinStock(ctx, t, d); err != nil {
			return err
		}
		return t.QueryRow(ctx, `SELECT coalesce(preferencias->'canales'->'matriz'->'sinstock:celular', 'true') <> 'false'
			FROM tiendas WHERE id = $1`, ev.TiendaID).Scan(&prendido)
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
	dest, err := s.deTienda(ctx, ev.TiendaID)
	if err != nil {
		return err
	}
	p := Payload(AvisoSinStock(ev.TiendaID, d.Producto, a))
	var envios []apns.Envio
	for _, x := range dest {
		if roles.Puede(x.rol, roles.Productos) {
			envios = append(envios, apns.Envio{Token: x.token, Entorno: x.entorno, Payload: p,
				Colapso: "sin-stock-" + d.Producto, Vence: time.Now().Add(24 * time.Hour)})
		}
	}
	return s.mandar(ctx, envios)
}

// destino es un iPhone con el rol de su persona en la tienda del aviso.
type destino struct {
	usuario, rol, token, entorno string
}

// deTienda: los iPhone del equipo de la tienda con sesión vigente.
func (s *Servicio) deTienda(ctx context.Context, tienda string) ([]destino, error) {
	var out []destino
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT usuario_id::text, rol, token, entorno FROM app.dispositivos_de_tienda()`)
		if err != nil {
			return err
		}
		defer filas.Close()
		for filas.Next() {
			var d destino
			if err := filas.Scan(&d.usuario, &d.rol, &d.token, &d.entorno); err != nil {
				return err
			}
			out = append(out, d)
		}
		return filas.Err()
	})
	return out, err
}

// mandar entrega un aviso a cada destino. Un token que APNs da por muerto se borra; si algún envío
// falló de una forma que vale reintentar (APNs caído, 429), devuelve error para que el outbox
// reintente (el apns-collapse-id evita que se vea dos veces). Lo demás se anota y sigue.
func (s *Servicio) mandar(ctx context.Context, envios []apns.Envio) error {
	var reintentar error
	for _, e := range envios {
		err := s.APNs.Enviar(ctx, e)
		var ea *apns.Error
		switch {
		case err == nil:
		case errors.Is(err, apns.ErrTokenInvalido):
			if err := s.Pool.EnTx(ctx, func(t pgx.Tx) error {
				_, err := t.Exec(ctx, `SELECT app.olvidar_dispositivo($1)`, e.Token)
				return err
			}); err != nil {
				return err
			}
		case errors.As(err, &ea) && !ea.Reintentable():
			s.Log.Warn("push: APNs rechazó el aviso", "estado", ea.Estado, "motivo", ea.Motivo)
		default:
			reintentar = err
		}
	}
	return reintentar
}

// Venta es lo que dice el push de una venta, según quién lo lee.
type Venta struct {
	Numero           int
	Total            int64
	Productos        int
	Pago, EstadoPago string
}

// AvisoVenta arma «Nueva venta #N» para un rol: el monto sólo con Finanzas (§37); nada de quien
// compró (se ve en la pantalla bloqueada).
func AvisoVenta(tienda, rol string, v Venta) Aviso {
	cuerpo := ""
	if roles.Puede(rol, roles.Finanzas) {
		cuerpo = pesos(v.Total) + " · "
	}
	switch v.Productos {
	case 0:
	case 1:
		cuerpo += "1 producto · "
	default:
		cuerpo += strconv.Itoa(v.Productos) + " productos · "
	}
	switch {
	case v.Pago == "transf" && v.EstadoPago == "Pendiente":
		cuerpo += "Transferencia por confirmar"
	case v.Pago == "transf":
		cuerpo += "Transferencia acreditada"
	case v.Pago == "mp":
		cuerpo += "Mercado Pago acreditado"
	default:
		cuerpo += "Entró por la tienda online"
	}
	return Aviso{Titulo: fmt.Sprintf("Nueva venta #%d", v.Numero), Cuerpo: cuerpo, Hilo: HiloTienda(tienda),
		Ruta: fmt.Sprintf("pedido/%d", v.Numero), Tienda: tienda}
}

// venta: push.venta, que encola el trigger app.avisar_venta_tienda junto con el mail. Sale una sola
// vez (pedidos.avisos->'push_venta'), a quien tiene Pedidos, y no sale si la tienda apagó
// Notificaciones › Venta nueva › Celular.
func (s *Servicio) venta(ctx context.Context, ev outbox.Evento) error {
	if !s.activo() {
		return nil
	}
	var d struct {
		Numero int `json:"numero"`
	}
	if json.Unmarshal(ev.Datos, &d) != nil || d.Numero <= 0 {
		return nil
	}
	var v Venta
	var ya, prendido bool
	var estado string
	err := s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT p.numero, p.total, coalesce((SELECT sum((i->>'cantidad')::int) FROM jsonb_array_elements(p.items) i), 0)::int,
				p.pago, p.estado_pago, p.estado, p.avisos ? 'push_venta',
				coalesce(t.preferencias->'canales'->'matriz'->'venta:celular', 'true') <> 'false'
			FROM pedidos p JOIN tiendas t ON t.id = p.tienda_id WHERE p.numero = $1`, d.Numero).
			Scan(&v.Numero, &v.Total, &v.Productos, &v.Pago, &v.EstadoPago, &estado, &ya, &prendido)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if ya || !prendido || estado == "Cancelado" {
		return nil
	}
	dest, err := s.deTienda(ctx, ev.TiendaID)
	if err != nil {
		return err
	}
	var envios []apns.Envio
	for _, x := range dest {
		if !roles.Puede(x.rol, roles.Pedidos) {
			continue
		}
		envios = append(envios, apns.Envio{Token: x.token, Entorno: x.entorno, Payload: Payload(AvisoVenta(ev.TiendaID, x.rol, v)),
			Colapso: fmt.Sprintf("venta-%d", v.Numero), Vence: time.Now().Add(24 * time.Hour)})
	}
	if len(envios) == 0 {
		return nil
	}
	if err := s.mandar(ctx, envios); err != nil {
		return err
	}
	return s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE pedidos SET avisos = avisos || jsonb_build_object('push_venta', $2::text) WHERE numero = $1`,
			d.Numero, time.Now().UTC().Format(time.RFC3339))
		return err
	})
}

// AvisoPreguntaML: el nombre del producto si se sabe; la pregunta no (puede traer datos de quien
// pregunta y se ve en la pantalla bloqueada).
func AvisoPreguntaML(tienda string, id int64, producto string) Aviso {
	cuerpo := "Respondela rápido: ayuda a vender."
	if producto != "" {
		cuerpo = "Te preguntaron por " + producto + "."
	}
	return Aviso{Titulo: "Pregunta nueva en Mercado Libre", Cuerpo: cuerpo, Hilo: HiloTienda(tienda),
		Ruta: fmt.Sprintf("preguntas/ml/%d", id), Tienda: tienda}
}

// preguntaML: push.pregunta_ml, que encola mercadolibre al guardar una pregunta nueva sin
// responder. A quien tiene Clientes (las preguntas son de ahí).
func (s *Servicio) preguntaML(ctx context.Context, ev outbox.Evento) error {
	if !s.activo() {
		return nil
	}
	var d struct {
		ID int64 `json:"id"`
	}
	if json.Unmarshal(ev.Datos, &d) != nil || d.ID <= 0 {
		return nil
	}
	var producto, estado string
	err := s.Pool.EnTienda(ctx, ev.TiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT coalesce(p.nombre, ''), q.estado FROM ml_preguntas q
			LEFT JOIN ml_publicaciones m ON m.tienda_id = q.tienda_id AND m.item_id = q.item_id
			LEFT JOIN productos p ON p.tienda_id = m.tienda_id AND p.id = m.producto_id
			WHERE q.pregunta_id = $1`, d.ID).Scan(&producto, &estado)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if estado != "UNANSWERED" {
		return nil
	}
	dest, err := s.deTienda(ctx, ev.TiendaID)
	if err != nil {
		return err
	}
	var envios []apns.Envio
	for _, x := range dest {
		if roles.Puede(x.rol, roles.Clientes) {
			envios = append(envios, apns.Envio{Token: x.token, Entorno: x.entorno, Payload: Payload(AvisoPreguntaML(ev.TiendaID, d.ID, producto)),
				Colapso: fmt.Sprintf("pregunta-ml-%d", d.ID), Vence: time.Now().Add(24 * time.Hour)})
		}
	}
	return s.mandar(ctx, envios)
}

// AvisoIngreso: «Entraron a tu cuenta», con el aparato y sin IP ni lugar.
func AvisoIngreso(aparato string) Aviso {
	cuerpo := "Alguien entró a tu cuenta de Tu Local Web. Si no fuiste vos, cerrá esa sesión."
	if aparato != "" {
		cuerpo = "Entraron desde " + aparato + ". Si no fuiste vos, cerrá esa sesión."
	}
	return Aviso{Titulo: "Entraron a tu cuenta", Cuerpo: cuerpo, Hilo: "cuenta", Ruta: "cuenta/seguridad"}
}

// ingreso: cuenta.ingreso_push, sólo a los iPhone de la persona, menos el de la sesión nueva.
func (s *Servicio) ingreso(ctx context.Context, ev outbox.Evento) error {
	if !s.activo() {
		return nil
	}
	var d cuentas.IngresoPush
	if json.Unmarshal(ev.Datos, &d) != nil || d.Usuario == "" {
		return nil
	}
	var envios []apns.Envio
	err := s.Pool.EnUsuario(ctx, d.Usuario, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT d.token, d.entorno FROM auth.dispositivos d
			JOIN auth.sesiones s ON s.token_hash = d.sesion_hash
			JOIN auth.usuarios u ON u.id = d.usuario_id
			WHERE d.sesion_hash <> $1 AND s.vence_en > now()
				AND NOT (coalesce((u.seguridad->>'inactivas')::boolean, false) AND s.visto_en < now() - $2::interval)`,
			d.Sesion, cuentas.Inactividad.String())
		if err != nil {
			return err
		}
		defer filas.Close()
		p := Payload(AvisoIngreso(d.Aparato))
		for filas.Next() {
			e := apns.Envio{Payload: p, Colapso: fmt.Sprintf("ingreso-%d", ev.ID), Vence: time.Now().Add(6 * time.Hour)}
			if err := filas.Scan(&e.Token, &e.Entorno); err != nil {
				return err
			}
			envios = append(envios, e)
		}
		return filas.Err()
	})
	if err != nil {
		return err
	}
	return s.mandar(ctx, envios)
}
