// Package equipo: el backoffice del equipo de Tu Local Web (/staff). Las personas del equipo tienen
// cuentas y sesiones propias (esquema equipo), separadas de las de las tiendas. La API lee todas las
// tiendas con el rol mostrador_staff (sólo lectura, por la política staff_lee de cada tabla) y lo que
// cambia en una tienda pasa por equipo.accion_tienda, que chequea el permiso y audita en la misma
// transacción. Todo acceso a una tienda o a una persona queda en equipo.auditoria.
package equipo

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/mercadopago"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
)

// Roles del equipo.
const (
	Soporte        = "soporte"
	Administracion = "administracion"
	Duenio         = "duenio"
)

// Permisos de cada rol: Soporte lee tiendas y responde; Administración además ve la facturación,
// decide partners, cambia planes, aplica crédito y pausa; Dueño, todo (incluido cerrar una tienda).
// Es la misma matriz que la tabla equipo.permisos (la que usa la función SQL) y que
// apps/web/src/lib/staff/permisos.ts: los tests comparan las tres contra testdata/permisos.json.
var Permisos = map[string][]string{
	Soporte: {"tiendas.ver", "tiendas.mirar", "tiendas.notas", "tiendas.prueba", "tiendas.pack",
		"consultas.ver", "consultas.responder"},
	Administracion: {"tiendas.ver", "tiendas.mirar", "tiendas.notas", "tiendas.prueba", "tiendas.pack",
		"tiendas.plan", "tiendas.credito", "tiendas.pausar", "consultas.ver", "consultas.responder",
		"partners.ver", "partners.decidir", "facturacion.ver", "auditoria.ver"},
	Duenio: {"tiendas.ver", "tiendas.mirar", "tiendas.notas", "tiendas.prueba", "tiendas.pack",
		"tiendas.plan", "tiendas.credito", "tiendas.pausar", "tiendas.cerrar", "consultas.ver",
		"consultas.responder", "partners.ver", "partners.decidir", "facturacion.ver", "auditoria.ver"},
}

// Puede dice si el rol tiene el permiso.
func Puede(rol, permiso string) bool { return slices.Contains(Permisos[rol], permiso) }

// PostulacionAprobada es lo que el equipo aprobó: con esto se da de alta al partner.
type PostulacionAprobada struct {
	Email, Nombre, Agencia string
}

// Aprobador da de alta al partner de una postulación aprobada. Lo implementa el paquete partners
// (rama pr-74): partners.Aprobar(ctx, app, partners.Aprobacion{Email, Nombre: Agencia}). Si falla
// (por ejemplo, partners.ErrSinCuenta: el email todavía no tiene cuenta en Tu Local Web), la
// postulación sigue nueva y el equipo ve el error.
type Aprobador func(ctx context.Context, p PostulacionAprobada) error

// Servicio del backoffice.
type Servicio struct {
	// Staff es el pool con el rol mostrador_staff: lee todo, escribe sólo lo del equipo.
	Staff *db.Pool
	// App sólo se usa para los límites de intentos (auth.intentos), como el resto de los ingresos.
	App *db.Pool
	// Sitio público (links de los mails).
	Sitio string
	// ClaveBFF: para creerle la IP del visitante al BFF.
	ClaveBFF string
	// Aprobador de partners. nil: aprobar sólo cambia el estado y avisa por mail.
	Aprobador Aprobador
	// Ahora (tests).
	Ahora func() time.Time
}

func (s *Servicio) ahora() time.Time {
	if s.Ahora != nil {
		return s.Ahora()
	}
	return time.Now()
}

// Miembro es la persona del equipo de la sesión.
type Miembro struct {
	ID      string    `json:"id"`
	Nombre  string    `json:"nombre"`
	Email   string    `json:"email"`
	Rol     string    `json:"-"`
	VenceEn time.Time `json:"-"`
}

// Errores de dominio.
var (
	ErrCredenciales = errors.New("email o contraseña incorrectos")
	ErrIntentos     = errors.New("demasiados intentos")
	ErrNoExiste     = errors.New("no existe")
	ErrSinPermiso   = errors.New("sin permiso")
	ErrResuelta     = errors.New("ya está resuelta")
	ErrCerrada      = errors.New("la tienda está cerrada")
)

// ErrInvalido es un dato mal cargado.
type ErrInvalido struct{ Campo, Mensaje string }

func (e ErrInvalido) Error() string { return e.Campo + ": " + e.Mensaje }

// Hora de Argentina (UTC−3 fijo).
var ar = time.FixedZone("AR", -3*3600)

const diasPrueba = 14

func nuevoToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func hashToken(t string) string {
	h := sha256.Sum256([]byte(t))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// ---- Tiendas ----

// tienda es lo que el backoffice necesita de cada tienda para listas, inicio y facturación.
type tienda struct {
	ID, Nombre, Slug, Plan, Ciclo, Rubro, Plantilla, Partner string
	Creada, PruebaHasta                                      time.Time
	Cancelada, Cerrada, Pausada                              *time.Time
	Suscripcion, BajaMotivo                                  *string
	Credito                                                  int64
	Titular, Email                                           string
}

const sqlTiendas = `
	SELECT t.id::text, t.nombre, t.slug, t.plan, t.plan_ciclo, t.rubro, t.plantilla,
		coalesce(lower(t.origen->>'ref'), ''), t.creada_en,
		coalesce(t.prueba_hasta, t.creada_en + interval '14 days'), t.cancelada_en, t.cerrada_en, t.pausada_en,
		t.suscripcion_estado, t.baja_motivo, t.credito, coalesce(u.nombre, ''), coalesce(u.email::text, '')
	FROM tiendas t
	LEFT JOIN LATERAL (
		SELECT u.nombre, u.email FROM auth.membresias m JOIN auth.usuarios u ON u.id = m.usuario_id
		WHERE m.tienda_id = t.id AND m.rol = 'duenio' ORDER BY m.creada_en LIMIT 1
	) u ON true`

func leerTiendas(ctx context.Context, q interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, donde string, args ...any) ([]tienda, error) {
	filas, err := q.Query(ctx, sqlTiendas+" "+donde, args...)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(filas, func(f pgx.CollectableRow) (tienda, error) {
		var t tienda
		err := f.Scan(&t.ID, &t.Nombre, &t.Slug, &t.Plan, &t.Ciclo, &t.Rubro, &t.Plantilla, &t.Partner, &t.Creada,
			&t.PruebaHasta, &t.Cancelada, &t.Cerrada, &t.Pausada, &t.Suscripcion, &t.BajaMotivo, &t.Credito,
			&t.Titular, &t.Email)
		return t, err
	})
}

// Estado de una tienda (la demo usa la misma regla en lib/staff/estado.ts).
func (t tienda) estado(ahora time.Time) string {
	switch {
	case t.Cerrada != nil:
		return "cerrada"
	case t.Pausada != nil:
		return "pausada"
	case t.Cancelada != nil:
		return "cancelada"
	case t.Plan != "gratis" && t.Suscripcion != nil && *t.Suscripcion == "paused":
		return "cobro_rechazado"
	case ahora.Before(t.PruebaHasta):
		return "prueba"
	}
	return "activa"
}

// Mensual en centavos del plan (el anual, mensualizado con el 25 % menos). Gratis: 0.
func Mensual(plan, ciclo string) int64 {
	v, ok := mercadopago.PrecioDe(plan, ciclo)
	if !ok {
		return 0
	}
	if ciclo == "anio" {
		v /= 12
	}
	return v * 100
}

// Cobro en centavos: lo que se cobra cada vez (el anual, por año).
func Cobro(plan, ciclo string) int64 {
	v, _ := mercadopago.PrecioDe(plan, ciclo)
	return v * 100
}

// mrr que aporta la tienda: sólo las activas con plan pago.
func (t tienda) mrr(ahora time.Time) int64 {
	if t.estado(ahora) != "activa" {
		return 0
	}
	return Mensual(t.Plan, t.Ciclo)
}

// proximoCobro: desde el fin de la prueba, de a un mes (o un año) hasta pasar hoy.
func (t tienda) proximoCobro(ahora time.Time) *time.Time {
	if t.Plan == "gratis" || t.Cerrada != nil || t.Cancelada != nil {
		return nil
	}
	p := t.PruebaHasta
	for !p.After(ahora) {
		if t.Ciclo == "anio" {
			p = p.AddDate(1, 0, 0)
		} else {
			p = p.AddDate(0, 1, 0)
		}
	}
	return &p
}

// diasHabiles suma n días hábiles (lunes a viernes) a la fecha en hora de Argentina. Los feriados no
// se descuentan: el plazo queda del lado seguro.
func diasHabiles(desde time.Time, n int) time.Time {
	d := desde.In(ar)
	d = time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, ar)
	for n > 0 {
		d = d.AddDate(0, 0, 1)
		if d.Weekday() != time.Saturday && d.Weekday() != time.Sunday {
			n--
		}
	}
	return d
}

func fecha(t time.Time) string { return t.In(ar).Format(time.DateOnly) }

// pesos formatea centavos como "$ 12.345" (espacio duro, como en el front).
func pesos(c int64) string {
	n := fmt.Sprint(c / 100)
	var b strings.Builder
	for i, r := range n {
		if i > 0 && (len(n)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(r)
	}
	return "$ " + b.String()
}

var nombresPlan = map[string]string{"gratis": "Gratis", "emprende": "Emprende", "pro": "Pro", "negocio": "Negocio"}

// anotar deja un registro en la auditoría del equipo, en la transacción del acceso.
func anotar(ctx context.Context, tx pgx.Tx, m Miembro, cl cuentas.Cliente, accion, texto, tiendaID string, detalle any) error {
	_, err := tx.Exec(ctx, `INSERT INTO equipo.auditoria (miembro_id, accion, texto, tienda_id, detalle, ip, agente)
		VALUES ($1, $2, left($3, 300), nullif($4, '')::uuid, $5, nullif($6, '')::inet, nullif($7, ''))`,
		m.ID, accion, texto, tiendaID, detalle, cl.IP, cl.Agente)
	return err
}
