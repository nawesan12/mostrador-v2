package mercadopago

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// El abono de Tu Local Web: suscripción sin plan, en estado pending (la guía de Subscriptions). El
// precio sale de este catálogo del servidor; el navegador sólo elige plan y ciclo. Tu Local Web nunca ve
// la tarjeta: la persona la carga en Mercado Pago (init_point).

// Precios por mes, en pesos (lib/panel/planes.ts). El anual es 25 % menos, cobrado una vez por año.
var precios = map[string]int64{"emprende": 19_900, "pro": 49_900, "negocio": 149_900}

// DiasDePrueba de la prueba gratis.
const DiasDePrueba = 14

var nombres = map[string]string{"emprende": "Emprende", "pro": "Pro", "negocio": "Negocio"}

// PrecioDe un plan en un ciclo: lo que se cobra cada vez.
func PrecioDe(plan, ciclo string) (int64, bool) {
	p, ok := precios[plan]
	if !ok || (ciclo != "mes" && ciclo != "anio") {
		return 0, false
	}
	if ciclo == "anio" {
		return p * 12 * 3 / 4, true
	}
	return p, true
}

type preapproval struct {
	ID                string `json:"id"`
	Status            string `json:"status"`
	InitPoint         string `json:"init_point"`
	ExternalReference string `json:"external_reference"`
}

// Suscribir crea la suscripción pendiente y devuelve el init_point. El plan cambia recién cuando
// Mercado Pago avisa que quedó autorizada.
func (s *Servicio) Suscribir(ctx context.Context, tiendaID, usuarioID, plan, ciclo string) (string, error) {
	if s.Config.TokenPlataforma == "" {
		return "", ErrApagado
	}
	importe, ok := PrecioDe(plan, ciclo)
	if !ok {
		return "", ErrPlanInvalido
	}
	var email string
	var creada time.Time
	if err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if err := t.QueryRow(ctx, `SELECT creada_en FROM tiendas WHERE id = $1`, tiendaID).Scan(&creada); err != nil {
			return err
		}
		return t.QueryRow(ctx, `SELECT email FROM auth.usuarios WHERE id = $1`, usuarioID).Scan(&email)
	}); err != nil {
		return "", err
	}
	// Lo que queda de la prueba gratis (14 días desde el alta) no se cobra.
	diasPrueba := DiasDePrueba - int(s.ahora().Sub(creada).Hours()/24)
	frecuencia := 1
	if ciclo == "anio" {
		frecuencia = 12
	}
	recurrencia := map[string]any{
		"frequency": frecuencia, "frequency_type": "months", "transaction_amount": importe, "currency_id": "ARS",
	}
	// Lo que queda de la prueba gratis no se cobra.
	if diasPrueba > 0 {
		recurrencia["free_trial"] = map[string]any{"frequency": diasPrueba, "frequency_type": "days"}
	}
	nombreCiclo := "mensual"
	if ciclo == "anio" {
		nombreCiclo = "anual"
	}
	var p preapproval
	if err := s.cliente().JSON(ctx, http.MethodPost, "/preapproval", s.Config.TokenPlataforma, "", map[string]any{
		"reason":             "Tu Local Web " + nombres[plan] + " (" + nombreCiclo + ")",
		"external_reference": tiendaID,
		"payer_email":        sinEtiqueta(email),
		"auto_recurring":     recurrencia,
		"back_url":           s.Config.Sitio + "/panel/plan-y-facturacion?suscripcion=vuelta",
		"status":             "pending",
	}, &p); err != nil {
		return "", err
	}
	if p.ID == "" || p.InitPoint == "" {
		return "", errors.New("mercadopago: la suscripción no trajo init_point")
	}
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE tiendas SET plan_pendiente = $2, ciclo_pendiente = $3, suscripcion_estado = 'pending',
			suscripcion_mp = CASE WHEN suscripcion_estado = 'authorized' THEN suscripcion_mp ELSE $4 END
			WHERE id = $1`, tiendaID, plan, ciclo, p.ID)
		return err
	})
	return p.InitPoint, err
}

// sinEtiqueta saca la parte "+algo" del mail ("ana+tienda@x.com" → "ana@x.com"): Mercado Pago
// rechaza esos mails con "User bad request", y es la misma casilla.
func sinEtiqueta(email string) string {
	local, dominio, ok := strings.Cut(email, "@")
	if !ok {
		return email
	}
	if base, _, ok := strings.Cut(local, "+"); ok && base != "" {
		return base + "@" + dominio
	}
	return email
}

var esUUID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// ConciliarSuscripcion trae la suscripción (con el token de la plataforma) y la aplica a su tienda.
func (s *Servicio) ConciliarSuscripcion(ctx context.Context, id string) error {
	if s.Config.TokenPlataforma == "" {
		return ErrApagado
	}
	var p preapproval
	if err := s.cliente().JSON(ctx, http.MethodGet, "/preapproval/"+url.PathEscape(id), s.Config.TokenPlataforma, "", nil, &p); err != nil {
		return err
	}
	if !esUUID.MatchString(p.ExternalReference) {
		return nil // no la creó Tu Local Web
	}
	tiendaID := p.ExternalReference
	var otras []string
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var actual *string
		var estado *string
		var pendiente, cicloPend *string
		err := t.QueryRow(ctx, `SELECT suscripcion_mp, suscripcion_estado, plan_pendiente, ciclo_pendiente FROM tiendas WHERE id = $1 FOR UPDATE`,
			tiendaID).Scan(&actual, &estado, &pendiente, &cicloPend)
		if err != nil {
			return err
		}
		switch p.Status {
		case "authorized":
			if pendiente == nil {
				return nil
			}
			if actual != nil && *actual != p.ID {
				otras = append(otras, *actual)
			}
			_, err = t.Exec(ctx, `UPDATE tiendas SET plan = $2, plan_ciclo = $3, suscripcion_mp = $4, suscripcion_estado = 'authorized',
				plan_pendiente = NULL, ciclo_pendiente = NULL, cancelada_en = NULL WHERE id = $1`,
				tiendaID, *pendiente, *cicloPend, p.ID)
			if err != nil {
				return err
			}
			// La campanita de la dueña (§45).
			return outbox.Publicar(ctx, t, tiendaID, "plan.activado", map[string]string{"plan": *pendiente})
		case "cancelled", "paused":
			// Sólo cuenta si es la vigente: una pendiente vieja que se cancela no cambia nada.
			if actual == nil || *actual != p.ID {
				return nil
			}
			plan := "plan"
			if p.Status == "cancelled" {
				plan = "'gratis'"
			}
			_, err = t.Exec(ctx, `UPDATE tiendas SET suscripcion_estado = $2, plan = `+plan+` WHERE id = $1`, tiendaID, p.Status)
			if err != nil || p.Status != "paused" || (estado != nil && *estado == "paused") {
				return err
			}
			// Mercado Pago pausó el cobro del abono vigente: la campanita de la dueña (§45).
			return outbox.Publicar(ctx, t, tiendaID, "plan.pausado", map[string]string{})
		}
		return nil
	})
	if err != nil {
		return err
	}
	// La anterior queda cancelada para no cobrar dos abonos.
	for _, o := range otras {
		if err := s.cambiarSuscripcion(ctx, o, "cancelled"); err != nil {
			s.log().Error("mercadopago: no se pudo cancelar la suscripción anterior", "tienda", tiendaID, "error", err)
		}
	}
	return nil
}

func (s *Servicio) cambiarSuscripcion(ctx context.Context, id, estado string) error {
	return s.cliente().JSON(ctx, http.MethodPut, "/preapproval/"+url.PathEscape(id), s.Config.TokenPlataforma, "",
		map[string]string{"status": estado}, nil)
}

// CancelarSuscripcion: la cancelación del plan desde el panel (el plan sigue hasta que Mercado Pago
// confirme con su aviso).
func (s *Servicio) CancelarSuscripcion(ctx context.Context, tiendaID string) error {
	if s.Config.TokenPlataforma == "" {
		return nil
	}
	var id *string
	if err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT suscripcion_mp FROM tiendas WHERE id = $1`, tiendaID).Scan(&id)
	}); err != nil || id == nil {
		return err
	}
	return s.cambiarSuscripcion(ctx, *id, "cancelled")
}
