package turnos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// Espera es cuánto se guarda un horario mientras se paga la seña.
const Espera = 15 * time.Minute

func tiendaDe(ctx context.Context, pool tx, slug string) (string, error) {
	if !slugValido.MatchString(slug) {
		return "", ErrNoExiste
	}
	var id *string
	if err := pool.EnTx(ctx, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT app.tienda_de_slug($1)::text`, slug).Scan(&id)
	}); err != nil {
		return "", err
	}
	if id == nil {
		return "", ErrNoExiste
	}
	return *id, nil
}

// Resumen es lo que la tienda pública muestra de los turnos (TiendaPublica.turnos). sena es el
// porcentaje que se cobra de verdad: 0 si la tienda no conectó Mercado Pago ("pagás en el local").
type Resumen struct {
	Servicios       []Servicio    `json:"servicios"`
	Profesionales   []Profesional `json:"profesionales"`
	Sena            int           `json:"sena"`
	Horario         []Franja      `json:"horario"`
	AnticipacionMin int           `json:"anticipacionMin"`
	DiasVista       int           `json:"diasVista"`
}

func conMercadoPago(ctx context.Context, t pgx.Tx, tienda string) (bool, error) {
	var mp bool
	err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM mp_conexiones WHERE tienda_id = $1 AND NOT reautorizar)`, tienda).Scan(&mp)
	return mp, err
}

// ResumenDe la tienda (por id), nil si no toma turnos. cobra dice si hay con qué cobrar la seña.
func ResumenDe(ctx context.Context, t pgx.Tx, tienda string, cobra bool) (*Resumen, error) {
	a, err := leerAjustes(ctx, t)
	if err != nil || !a.Activo {
		return nil, err
	}
	r := &Resumen{Horario: a.Horario, AnticipacionMin: a.AnticipacionMin, DiasVista: a.DiasVista}
	if cobra {
		mp, err := conMercadoPago(ctx, t, tienda)
		if err != nil {
			return nil, err
		}
		if mp {
			r.Sena = a.Sena
		}
	}
	if r.Servicios, err = leerServicios(ctx, t, true); err != nil {
		return nil, err
	}
	if r.Profesionales, err = leerProfesionales(ctx, t, true, false); err != nil {
		return nil, err
	}
	return r, nil
}

// Publico: los turnos de la tienda como los ve quien reserva. ErrApagado si no toma turnos.
func Publico(ctx context.Context, pool tx, mp Cobrador, slug string) (Resumen, error) {
	id, err := tiendaDe(ctx, pool, slug)
	if err != nil {
		return Resumen{}, err
	}
	var r *Resumen
	err = pool.EnTienda(ctx, id, func(t pgx.Tx) (err error) {
		r, err = ResumenDe(ctx, t, id, mp != nil)
		return err
	})
	if err != nil {
		return Resumen{}, err
	}
	if r == nil {
		return Resumen{}, ErrApagado
	}
	return *r, nil
}

// consulta arma la entrada de Libres para un día, con los turnos que lo tocan.
func consulta(ctx context.Context, t pgx.Tx, servicioID string, profesionalID *string, dia string, ahora time.Time, bloquear bool) (Consulta, error) {
	c := Consulta{Dia: dia, Ahora: ahora, ProfesionalID: profesionalID}
	d, err := time.ParseInLocation("2006-01-02", dia, Zona)
	if err != nil || !idValido.MatchString(servicioID) || (profesionalID != nil && !idValido.MatchString(*profesionalID)) {
		return c, ErrInvalido
	}
	if c.Ajustes, err = leerAjustes(ctx, t); err != nil {
		return c, err
	}
	if !c.Ajustes.Activo {
		return c, ErrApagado
	}
	c.Servicio, err = escanearServicio(t.QueryRow(ctx, `SELECT `+colServicio+` FROM turnos_servicios WHERE id = $1`, servicioID))
	if errors.Is(err, pgx.ErrNoRows) {
		return c, ErrInvalido
	}
	if err != nil {
		return c, err
	}
	if c.Profesionales, err = leerProfesionales(ctx, t, true, bloquear); err != nil {
		return c, err
	}
	c.Turnos, err = ocupados(ctx, t, d, d.AddDate(0, 0, 1))
	return c, err
}

func profesionalDe(s string) *string {
	if s == "" || s == "cualquiera" {
		return nil
	}
	return &s
}

// LibresDe un día para la tienda pública. profesional "" o "cualquiera" = cualquiera.
func LibresDe(ctx context.Context, pool tx, slug, servicio, profesional, dia string, ahora time.Time) ([]HorarioLibre, error) {
	id, err := tiendaDe(ctx, pool, slug)
	if err != nil {
		return nil, err
	}
	var out []HorarioLibre
	err = pool.EnTienda(ctx, id, func(t pgx.Tx) error {
		c, err := consulta(ctx, t, servicio, profesionalDe(profesional), dia, ahora, false)
		if err != nil {
			return err
		}
		out = Libres(c)
		return nil
	})
	return out, err
}

// Reserva es lo que manda la tienda (precio, duración y seña NO vienen: salen del servicio).
type Reserva struct {
	ServicioID    string  `json:"servicioId"`
	ProfesionalID *string `json:"profesionalId"`
	Inicio        string  `json:"inicio"`
	Nombre        string  `json:"nombre"`
	Telefono      string  `json:"telefono"`
	Email         string  `json:"email"`
}

// Reservado es la respuesta de reservar: el código del turno y, si hay seña, adónde pagarla.
type Reservado struct {
	Codigo string `json:"codigo"`
	URL    string `json:"url,omitempty"`
}

// Reservar un turno desde la tienda. Todo se valida de nuevo acá, y la disponibilidad se mira con
// los profesionales bloqueados (SELECT ... FOR UPDATE) dentro de la transacción del alta: dos
// personas que eligen el mismo horario a la vez no lo pueden tomar las dos.
func Reservar(ctx context.Context, pool tx, mp Cobrador, sitio, slug string, r Reserva, ahora time.Time) (Reservado, error) {
	id, err := tiendaDe(ctx, pool, slug)
	if err != nil {
		return Reservado{}, err
	}
	c := Cliente{Nombre: r.Nombre, Telefono: r.Telefono, Email: r.Email}
	ini, err := time.Parse(time.RFC3339, r.Inicio)
	if err != nil || c.validar(true) != nil {
		return Reservado{}, ErrInvalido
	}
	var turnoID, codigo string
	var sena int64
	err = pool.EnTienda(ctx, id, func(t pgx.Tx) error {
		q, err := consulta(ctx, t, r.ServicioID, profesionalDe(deref(r.ProfesionalID)), ini.In(Zona).Format("2006-01-02"), ahora, true)
		if err != nil {
			return err
		}
		prof := Asignar(q, ini)
		if prof == "" {
			return ErrOcupado
		}
		var nombreProf string
		for _, p := range q.Profesionales {
			if p.ID == prof {
				nombreProf = p.Nombre
			}
		}
		estado := "confirmado"
		var vence *time.Time
		if mp != nil && q.Ajustes.Sena > 0 {
			conMP, err := conMercadoPago(ctx, t, id)
			if err != nil {
				return err
			}
			if conMP {
				sena = SenaDe(q.Servicio.Precio, q.Ajustes.Sena)
			}
		}
		if sena > 0 {
			estado = "reservado"
			v := ahora.Add(Espera)
			vence = &v
		}
		codigo = nuevoCodigo()
		fin := ini.Add(time.Duration(q.Servicio.Duracion) * time.Minute)
		if err := t.QueryRow(ctx, `INSERT INTO turnos (tienda_id, codigo_publico, servicio_id, servicio, profesional_id,
				profesional, inicio, fin, cliente_nombre, cliente_telefono, cliente_email, precio, sena, estado, origen,
				creado_en, vence_en)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, 'tienda', $15, $16) RETURNING id::text`,
			id, codigo, q.Servicio.ID, q.Servicio.Nombre, prof, nombreProf, ini, fin, c.Nombre, c.Telefono, c.Email,
			q.Servicio.Precio, sena, estado, ahora, vence).Scan(&turnoID); err != nil {
			return err
		}
		// El aviso a la tienda (Telegram, notificaciones) sale del outbox.
		return outbox.Publicar(ctx, t, id, "turno.nuevo", map[string]any{"id": turnoID, "servicio": q.Servicio.Nombre,
			"profesional": nombreProf, "inicio": fecha(ini), "cliente": c.Nombre, "estado": estado, "sena": sena})
	})
	if err != nil {
		return Reservado{}, err
	}
	out := Reservado{Codigo: codigo}
	if sena == 0 {
		return out, nil
	}
	out.URL, err = mp.PagarTurno(ctx, id, turnoID, fmt.Sprintf("%s/t/%s/turno/%s", sitio, slug, codigo))
	if err != nil {
		// Sin preferencia no hay cómo pagar: el horario se libera ya, para poder probar de nuevo.
		// Con un contexto que no se cancela: aunque la persona cierre la página, el horario se libera.
		sigue := context.WithoutCancel(ctx)
		errLiberar := pool.EnTienda(sigue, id, func(t pgx.Tx) error {
			_, err := t.Exec(sigue, `UPDATE turnos SET estado = 'cancelado', vence_en = NULL WHERE id = $1`, turnoID)
			return err
		})
		return Reservado{}, errors.Join(ErrMercadoPago, err, errLiberar)
	}
	return out, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// TurnoPublico es lo que muestra /t/{slug}/turno/{codigo}: sin los datos de contacto.
type TurnoPublico struct {
	Codigo      string     `json:"codigo"`
	Servicio    string     `json:"servicio"`
	Profesional string     `json:"profesional"`
	Inicio      string     `json:"inicio"`
	Fin         string     `json:"fin"`
	Nombre      string     `json:"nombre"`
	Precio      int64      `json:"precio"`
	Sena        int64      `json:"sena"`
	SenaPagada  bool       `json:"senaPagada"`
	Estado      string     `json:"estado"`
	VenceEn     *time.Time `json:"venceEn"`
	Vencido     bool       `json:"vencido"`
	// Cancelable: todavía se puede cancelar desde la página (hasta anticipacionMin antes).
	Cancelable bool   `json:"cancelable"`
	Tienda     string `json:"tienda"`
	Direccion  string `json:"direccion"`
	Whatsapp   string `json:"whatsapp"`
}

func verEn(ctx context.Context, t pgx.Tx, tienda, codigo string, ahora time.Time, bloquear bool) (TurnoPublico, error) {
	q := `SELECT ` + colTurno + ` FROM turnos WHERE codigo_publico = $1`
	if bloquear {
		q += ` FOR UPDATE`
	}
	tu, err := escanearTurno(t.QueryRow(ctx, q, codigo), ahora)
	if errors.Is(err, pgx.ErrNoRows) {
		return TurnoPublico{}, ErrNoExiste
	}
	if err != nil {
		return TurnoPublico{}, err
	}
	a, err := leerAjustes(ctx, t)
	if err != nil {
		return TurnoPublico{}, err
	}
	p := TurnoPublico{Codigo: tu.Codigo, Servicio: tu.Servicio, Profesional: tu.Profesional, Inicio: tu.Inicio, Fin: tu.Fin,
		Nombre: strings.SplitN(tu.Cliente.Nombre, " ", 2)[0], Precio: tu.Precio, Sena: tu.Sena, SenaPagada: tu.SenaPagada,
		Estado: tu.Estado, VenceEn: tu.VenceEn, Vencido: tu.Vencido}
	ini, _ := time.Parse(time.RFC3339, tu.Inicio)
	p.Cancelable = (tu.Estado == "reservado" && !tu.Vencido || tu.Estado == "confirmado") &&
		!ahora.After(ini.Add(-time.Duration(a.AnticipacionMin)*time.Minute))
	var negocio []byte
	if err := t.QueryRow(ctx, `SELECT nombre, negocio FROM tiendas WHERE id = $1`, tienda).Scan(&p.Tienda, &negocio); err != nil {
		return p, err
	}
	var n struct{ Direccion, Whatsapp string }
	_ = json.Unmarshal(negocio, &n)
	p.Direccion, p.Whatsapp = n.Direccion, n.Whatsapp
	return p, nil
}

// Ver un turno por su código público (sólo dentro de su tienda).
func Ver(ctx context.Context, pool tx, slug, codigo string, ahora time.Time) (TurnoPublico, error) {
	id, err := tiendaDe(ctx, pool, slug)
	if err != nil {
		return TurnoPublico{}, err
	}
	if !codigoValid.MatchString(codigo) {
		return TurnoPublico{}, ErrNoExiste
	}
	var p TurnoPublico
	err = pool.EnTienda(ctx, id, func(t pgx.Tx) (err error) {
		p, err = verEn(ctx, t, id, codigo, ahora, false)
		return err
	})
	return p, err
}

// Cancelar un turno desde su página: hasta anticipacionMin antes. Libera el horario; si tenía seña
// pagada, la devolución la hace la tienda. Cancelar uno ya cancelado no es error.
func Cancelar(ctx context.Context, pool tx, slug, codigo string, ahora time.Time) (TurnoPublico, error) {
	id, err := tiendaDe(ctx, pool, slug)
	if err != nil {
		return TurnoPublico{}, err
	}
	if !codigoValid.MatchString(codigo) {
		return TurnoPublico{}, ErrNoExiste
	}
	var p TurnoPublico
	err = pool.EnTienda(ctx, id, func(t pgx.Tx) (err error) {
		p, err = verEn(ctx, t, id, codigo, ahora, true)
		if err != nil || p.Estado == "cancelado" {
			return err
		}
		if !p.Cancelable {
			return ErrTarde
		}
		if _, err := t.Exec(ctx, `UPDATE turnos SET estado = 'cancelado', vence_en = NULL WHERE codigo_publico = $1`, codigo); err != nil {
			return err
		}
		p.Estado, p.VenceEn, p.Cancelable, p.Vencido = "cancelado", nil, false, false
		return outbox.Publicar(ctx, t, id, "turno.cancelado", map[string]any{"codigo": codigo, "senaPagada": p.SenaPagada})
	})
	return p, err
}
