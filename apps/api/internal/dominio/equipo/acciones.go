package equipo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/correo"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// Accion sobre una tienda (components/schemas/AccionTiendaEquipo).
type Accion struct {
	Accion       string `json:"accion"`
	Dias         int    `json:"dias,omitempty"`
	Plan         string `json:"plan,omitempty"`
	Monto        int64  `json:"monto,omitempty"`
	Cantidad     int    `json:"cantidad,omitempty"`
	Motivo       string `json:"motivo,omitempty"`
	Confirmacion string `json:"confirmacion,omitempty"`
}

// PermisoDe cada acción (la función SQL usa la misma tabla).
var PermisoDe = map[string]string{
	"extender_prueba": "tiendas.prueba", "regalar_pack": "tiendas.pack", "cambiar_plan": "tiendas.plan",
	"aplicar_credito": "tiendas.credito", "pausar": "tiendas.pausar", "reactivar": "tiendas.pausar",
	"cerrar": "tiendas.cerrar",
}

func plural(n int, uno, varios string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, uno)
	}
	return fmt.Sprintf("%d %s", n, varios)
}

// validar revisa la acción y arma los datos para la función SQL y el texto de la auditoría.
func (a Accion) validar(nombre, slug string) (map[string]any, string, error) {
	a.Motivo = strings.TrimSpace(a.Motivo)
	conMotivo := func() error {
		if utf8.RuneCountInString(a.Motivo) < 5 || utf8.RuneCountInString(a.Motivo) > 500 {
			return ErrInvalido{"motivo", "Contá el motivo (de 5 a 500 caracteres): queda en la auditoría."}
		}
		return nil
	}
	d := map[string]any{}
	if a.Motivo != "" {
		d["motivo"] = a.Motivo
	}
	switch a.Accion {
	case "extender_prueba":
		if a.Dias < 1 || a.Dias > 30 {
			return nil, "", ErrInvalido{"dias", "La prueba se extiende de 1 a 30 días."}
		}
		d["dias"] = a.Dias
		return d, "Extendió la prueba de " + nombre + " " + plural(a.Dias, "día", "días"), nil
	case "cambiar_plan":
		if _, ok := nombresPlan[a.Plan]; !ok {
			return nil, "", ErrInvalido{"plan", "El plan es gratis, emprende, pro o negocio."}
		}
		d["plan"] = a.Plan
		return d, "Cambió el plan de " + nombre + " a " + nombresPlan[a.Plan], nil
	case "aplicar_credito":
		if a.Monto < 100 || a.Monto > 150_000_000 {
			return nil, "", ErrInvalido{"monto", "El crédito va de $ 1 a $ 1.500.000."}
		}
		if err := conMotivo(); err != nil {
			return nil, "", err
		}
		d["monto"] = a.Monto
		return d, "Aplicó " + pesos(a.Monto) + " de crédito a " + nombre, nil
	case "regalar_pack":
		if a.Cantidad < 1 || a.Cantidad > 5 {
			return nil, "", ErrInvalido{"cantidad", "Se regalan de 1 a 5 packs."}
		}
		d["cantidad"] = a.Cantidad
		return d, "Regaló " + plural(a.Cantidad, "pack", "packs") + " de IA a " + nombre, nil
	case "pausar":
		if err := conMotivo(); err != nil {
			return nil, "", err
		}
		return d, "Pausó la tienda " + nombre, nil
	case "reactivar":
		return d, "Reactivó la tienda " + nombre, nil
	case "cerrar":
		if err := conMotivo(); err != nil {
			return nil, "", err
		}
		if strings.TrimSpace(strings.ToLower(a.Confirmacion)) != slug {
			return nil, "", ErrInvalido{"confirmacion", "Escribí la dirección de la tienda (" + slug + ") para confirmar."}
		}
		return d, "Cerró la tienda " + nombre, nil
	}
	return nil, "", ErrInvalido{"accion", "Esa acción no existe."}
}

// Aplicar hace la acción con equipo.accion_tienda: chequea el permiso, cambia la tienda y la anota
// en las dos auditorías (la de la tienda y la del equipo) en la misma transacción.
func (s *Servicio) Aplicar(ctx context.Context, m Miembro, tiendaID string, a Accion, cl cuentas.Cliente) error {
	permiso, ok := PermisoDe[a.Accion]
	if !ok {
		return ErrInvalido{"accion", "Esa acción no existe."}
	}
	if !Puede(m.Rol, permiso) {
		return ErrSinPermiso
	}
	return s.Staff.EnTx(ctx, func(tx pgx.Tx) error {
		var nombre, slug string
		err := tx.QueryRow(ctx, `SELECT nombre, slug FROM tiendas WHERE id::text = $1`, tiendaID).Scan(&nombre, &slug)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExiste
		}
		if err != nil {
			return err
		}
		datos, texto, err := a.validar(nombre, slug)
		if err != nil {
			return err
		}
		b, _ := json.Marshal(datos)
		var id int64
		return errSQL(tx.QueryRow(ctx, `SELECT equipo.accion_tienda($1, $2::uuid, $3, $4, $5, $6, $7)`,
			m.ID, tiendaID, a.Accion, b, texto, cl.IP, cl.Agente).Scan(&id))
	})
}

// errSQL traduce los errores de la función SQL.
func errSQL(err error) error {
	var pg *pgconn.PgError
	if !errors.As(err, &pg) {
		return err
	}
	switch pg.Code {
	case "42501":
		return ErrSinPermiso
	case "P0002":
		return ErrNoExiste
	case "55000":
		return ErrCerrada
	case "22023":
		return ErrInvalido{"accion", "Los datos de la acción no son válidos."}
	}
	return err
}

// Decidir aprueba o rechaza una postulación de partner. Al aprobar, si hay Aprobador, se da de alta
// al partner antes de confirmar: si falla, la postulación sigue nueva.
func (s *Servicio) Decidir(ctx context.Context, m Miembro, id, decision, motivo string, cl cuentas.Cliente) error {
	motivo = strings.TrimSpace(motivo)
	if decision != "aprobar" && decision != "rechazar" {
		return ErrInvalido{"decision", "La decisión es aprobar o rechazar."}
	}
	if decision == "rechazar" && (utf8.RuneCountInString(motivo) < 5 || utf8.RuneCountInString(motivo) > 500) {
		return ErrInvalido{"motivo", "Contá el motivo del rechazo (de 5 a 500 caracteres): se lo mandamos por mail."}
	}
	return s.Staff.EnTx(ctx, func(tx pgx.Tx) error {
		var p PostulacionAprobada
		var estado string
		err := tx.QueryRow(ctx, `SELECT nombre, email::text, agencia, estado FROM soporte.postulaciones_partner
			WHERE id::text = $1 FOR UPDATE`, id).Scan(&p.Nombre, &p.Email, &p.Agencia, &estado)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNoExiste
		}
		if err != nil {
			return err
		}
		if estado != "nueva" {
			return ErrResuelta
		}
		nuevo, accion, texto := "aprobada", "partner_aprobar", "Aprobó la postulación de "+p.Agencia
		asunto := "Ya sos partner de Mostrador"
		cuerpo := "¡Hola, " + p.Nombre + "!\n\nAprobamos la postulación de " + p.Agencia + " al programa de partners de Mostrador. " +
			"Ingresá con este email en " + s.Sitio + "/partners/panel para ver tu link de referido, tu cupón y los materiales."
		if decision == "rechazar" {
			nuevo, accion, texto = "rechazada", "partner_rechazar", "Rechazó la postulación de "+p.Agencia
			asunto = "Tu postulación al programa de partners"
			cuerpo = "¡Hola, " + p.Nombre + "!\n\nGracias por postularte al programa de partners de Mostrador. Por ahora no " +
				"podemos sumar a " + p.Agencia + ": " + motivo + "\n\nSi algo cambia, podés volver a escribirnos."
		}
		if _, err := tx.Exec(ctx, `UPDATE soporte.postulaciones_partner SET estado = $2, motivo = nullif($3, ''),
			resuelta_en = now(), resuelta_por = $4 WHERE id::text = $1`, id, nuevo, motivo, m.ID); err != nil {
			return err
		}
		if err := encolarCorreo(ctx, tx, p.Email, asunto, cuerpo); err != nil {
			return err
		}
		if err := anotar(ctx, tx, m, cl, accion, texto, "", map[string]string{"postulacion": id, "motivo": motivo}); err != nil {
			return err
		}
		if decision == "aprobar" && s.Aprobador != nil {
			return s.Aprobador(ctx, p)
		}
		return nil
	})
}

// Manejadores del worker: los mails de la plataforma que encola el equipo.
func Manejadores(c correo.Correo) map[string]outbox.Manejador {
	return map[string]outbox.Manejador{
		"equipo.correo": func(ctx context.Context, e outbox.Evento) error {
			var m struct{ Para, Asunto, Texto, HTML string }
			if err := json.Unmarshal(e.Datos, &m); err != nil {
				slog.Error("equipo.correo: datos ilegibles", "evento", e.ID, "error", err)
				return nil // reintentar no lo arregla
			}
			return c.Enviar(ctx, correo.Mensaje{Para: m.Para, Asunto: m.Asunto, Texto: m.Texto, HTML: m.HTML})
		},
	}
}
