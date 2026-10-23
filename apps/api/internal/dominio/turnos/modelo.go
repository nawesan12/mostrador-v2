// Package turnos: turnos con seña para las tiendas de servicios (peluquería, barbería y cualquier
// rubro que lo prenda). La tienda publica servicios y profesionales, quien compra reserva desde la
// tienda (con seña por Mercado Pago si la tienda lo conectó) y la tienda maneja la agenda en el panel.
package turnos

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
)

var (
	ErrNoExiste    = errors.New("turnos: no existe")
	ErrInvalido    = errors.New("turnos: datos inválidos")
	ErrApagado     = errors.New("turnos: la tienda no toma turnos")
	ErrOcupado     = errors.New("turnos: el horario ya no está libre")
	ErrTarde       = errors.New("turnos: ya no se puede cancelar")
	ErrSinSena     = errors.New("turnos: el turno no tiene seña pagada")
	ErrSinCobros   = errors.New("turnos: la tienda no conectó Mercado Pago")
	ErrMercadoPago = errors.New("turnos: Mercado Pago no respondió")
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
	EnTx(context.Context, func(pgx.Tx) error) error
}

// Cobrador cobra y devuelve la seña con la cuenta de Mercado Pago de la tienda (paquete mercadopago).
type Cobrador interface {
	PagarTurno(ctx context.Context, tiendaID, turnoID, vuelta string) (string, error)
	DevolverSena(ctx context.Context, tiendaID, turnoID string) error
}

// Franja es components/schemas/Franja: dia 0 = domingo, hora de Argentina.
type Franja struct {
	Dia   int    `json:"dia"`
	Desde string `json:"desde"`
	Hasta string `json:"hasta"`
}

// Ajustes es components/schemas/AjustesTurnos.
type Ajustes struct {
	Activo          bool     `json:"activo"`
	Sena            int      `json:"sena"`
	AnticipacionMin int      `json:"anticipacionMin"`
	DiasVista       int      `json:"diasVista"`
	Horario         []Franja `json:"horario"`
}

// PorDefecto: sin fila, la tienda no toma turnos.
func PorDefecto() Ajustes {
	return Ajustes{Sena: 30, AnticipacionMin: 60, DiasVista: 14, Horario: []Franja{}}
}

// Servicio es components/schemas/Servicio. precio en pesos enteros, duracion en minutos.
type Servicio struct {
	ID          string  `json:"id"`
	Nombre      string  `json:"nombre"`
	Descripcion string  `json:"descripcion"`
	Precio      int64   `json:"precio"`
	Duracion    int     `json:"duracion"`
	Foto        *string `json:"foto"`
	Etiqueta    *string `json:"etiqueta"`
	Activo      bool    `json:"activo"`
	Orden       int     `json:"orden"`
}

// Profesional es components/schemas/Profesional. servicios vacío = hace todos.
type Profesional struct {
	ID        string   `json:"id"`
	Nombre    string   `json:"nombre"`
	Rol       string   `json:"rol"`
	Foto      *string  `json:"foto"`
	Servicios []string `json:"servicios"`
	Activo    bool     `json:"activo"`
	Orden     int      `json:"orden"`
}

type Cliente struct {
	Nombre   string `json:"nombre"`
	Telefono string `json:"telefono"`
	Email    string `json:"email"`
}

// Turno es components/schemas/Turno (para el panel).
type Turno struct {
	ID            string     `json:"id"`
	Codigo        string     `json:"codigo"`
	ServicioID    string     `json:"servicioId"`
	Servicio      string     `json:"servicio"`
	ProfesionalID string     `json:"profesionalId"`
	Profesional   string     `json:"profesional"`
	Inicio        string     `json:"inicio"`
	Fin           string     `json:"fin"`
	Cliente       Cliente    `json:"cliente"`
	Precio        int64      `json:"precio"`
	Sena          int64      `json:"sena"`
	Estado        string     `json:"estado"`
	SenaPagada    bool       `json:"senaPagada"`
	SenaDevuelta  bool       `json:"senaDevuelta"`
	Origen        string     `json:"origen"`
	Nota          *string    `json:"nota"`
	CreadoEn      time.Time  `json:"creadoEn"`
	VenceEn       *time.Time `json:"venceEn"`
	// Vencido: reservado, sin seña y pasado venceEn (ya no ocupa el horario).
	Vencido bool `json:"vencido"`
}

var (
	idValido    = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	emailValido = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	codigoValid = regexp.MustCompile(`^[A-Za-z0-9_-]{16,64}$`)
	slugValido  = regexp.MustCompile(`^[a-z0-9-]{2,60}$`)
	estados     = map[string]bool{"reservado": true, "confirmado": true, "cancelado": true, "atendido": true, "ausente": true}
)

func largo(s string) int { return utf8.RuneCountInString(s) }

func nuevoCodigo() string {
	b := make([]byte, 18)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// fecha en ISO con el offset de Argentina (−03:00).
func fecha(t time.Time) string { return t.In(Zona).Format(time.RFC3339) }

// SenaDe: el monto de la seña (porcentaje del precio, redondeado al peso).
func SenaDe(precio int64, porcentaje int) int64 {
	return (precio*int64(porcentaje) + 50) / 100
}

func vacioANil(s *string) *string {
	if s == nil {
		return nil
	}
	v := strings.TrimSpace(*s)
	if v == "" {
		return nil
	}
	return &v
}

func (a *Ajustes) validar() error {
	if a.Sena < 0 || a.Sena > 100 || a.AnticipacionMin < 0 || a.AnticipacionMin > 10080 || a.DiasVista < 1 || a.DiasVista > 90 ||
		len(a.Horario) > 50 {
		return ErrInvalido
	}
	if a.Horario == nil {
		a.Horario = []Franja{}
	}
	for _, f := range a.Horario {
		d, ok1 := minutos(f.Desde)
		h, ok2 := minutos(f.Hasta)
		if f.Dia < 0 || f.Dia > 6 || !ok1 || !ok2 || d >= h || d%Paso != 0 || h%Paso != 0 {
			return ErrInvalido
		}
	}
	return nil
}

func (s *Servicio) validar() error {
	s.Nombre, s.Descripcion = strings.TrimSpace(s.Nombre), strings.TrimSpace(s.Descripcion)
	s.Foto, s.Etiqueta = vacioANil(s.Foto), vacioANil(s.Etiqueta)
	if largo(s.Nombre) < 2 || largo(s.Nombre) > 60 || largo(s.Descripcion) > 200 || s.Precio <= 0 || s.Precio > 1e10 ||
		s.Duracion < 15 || s.Duracion > 480 || s.Duracion%Paso != 0 || (s.Foto != nil && len(*s.Foto) > 2048) ||
		(s.Etiqueta != nil && largo(*s.Etiqueta) > 30) || s.Orden < 0 || s.Orden > 10000 {
		return ErrInvalido
	}
	return nil
}

func (p *Profesional) validar() error {
	p.Nombre, p.Rol = strings.TrimSpace(p.Nombre), strings.TrimSpace(p.Rol)
	p.Foto = vacioANil(p.Foto)
	if largo(p.Nombre) < 2 || largo(p.Nombre) > 40 || largo(p.Rol) > 60 || (p.Foto != nil && len(*p.Foto) > 2048) ||
		len(p.Servicios) > 100 || p.Orden < 0 || p.Orden > 10000 {
		return ErrInvalido
	}
	if p.Servicios == nil {
		p.Servicios = []string{}
	}
	for _, s := range p.Servicios {
		if !idValido.MatchString(s) {
			return ErrInvalido
		}
	}
	return nil
}

func (c *Cliente) validar(estricto bool) error {
	c.Nombre, c.Telefono, c.Email = strings.TrimSpace(c.Nombre), strings.TrimSpace(c.Telefono), strings.TrimSpace(c.Email)
	if largo(c.Nombre) < 2 || largo(c.Nombre) > 80 || largo(c.Telefono) > 30 || len(c.Email) > 120 {
		return ErrInvalido
	}
	if estricto && (largo(c.Telefono) < 6 || !emailValido.MatchString(c.Email)) {
		return ErrInvalido
	}
	if !estricto && c.Email != "" && !emailValido.MatchString(c.Email) {
		return ErrInvalido
	}
	return nil
}

// Lecturas compartidas por el panel y lo público.

func leerAjustes(ctx context.Context, t pgx.Tx) (Ajustes, error) {
	a := PorDefecto()
	err := t.QueryRow(ctx, `SELECT activo, sena, anticipacion_min, dias_vista, horario FROM turnos_ajustes`).
		Scan(&a.Activo, &a.Sena, &a.AnticipacionMin, &a.DiasVista, &a.Horario)
	if errors.Is(err, pgx.ErrNoRows) {
		return a, nil
	}
	if a.Horario == nil {
		a.Horario = []Franja{}
	}
	return a, err
}

const colServicio = `id::text, nombre, descripcion, precio, duracion, foto, etiqueta, activo, orden`

func escanearServicio(f pgx.Row) (Servicio, error) {
	var s Servicio
	err := f.Scan(&s.ID, &s.Nombre, &s.Descripcion, &s.Precio, &s.Duracion, &s.Foto, &s.Etiqueta, &s.Activo, &s.Orden)
	return s, err
}

func leerServicios(ctx context.Context, t pgx.Tx, soloActivos bool) ([]Servicio, error) {
	filas, err := t.Query(ctx, `SELECT `+colServicio+` FROM turnos_servicios WHERE activo OR NOT $1 ORDER BY orden, creado_en`, soloActivos)
	if err != nil {
		return nil, err
	}
	out := []Servicio{}
	for filas.Next() {
		s, err := escanearServicio(filas)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, filas.Err()
}

const colProfesional = `id::text, nombre, rol, foto, servicios::text[], activo, orden`

func escanearProfesional(f pgx.Row) (Profesional, error) {
	var p Profesional
	err := f.Scan(&p.ID, &p.Nombre, &p.Rol, &p.Foto, &p.Servicios, &p.Activo, &p.Orden)
	if p.Servicios == nil {
		p.Servicios = []string{}
	}
	return p, err
}

func leerProfesionales(ctx context.Context, t pgx.Tx, soloActivos bool, bloquear bool) ([]Profesional, error) {
	q := `SELECT ` + colProfesional + ` FROM turnos_profesionales WHERE activo OR NOT $1 ORDER BY orden, creado_en`
	if bloquear {
		// Siempre en el mismo orden (por id) para que dos reservas no se traben entre sí.
		q = `SELECT ` + colProfesional + ` FROM turnos_profesionales WHERE activo OR NOT $1 ORDER BY id FOR UPDATE`
	}
	filas, err := t.Query(ctx, q, soloActivos)
	if err != nil {
		return nil, err
	}
	out := []Profesional{}
	for filas.Next() {
		p, err := escanearProfesional(filas)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, filas.Err()
}

// ocupados: los turnos no cancelados de los profesionales que tocan el rango [desde, hasta).
func ocupados(ctx context.Context, t pgx.Tx, desde, hasta time.Time) ([]Ocupado, error) {
	filas, err := t.Query(ctx, `SELECT profesional_id::text, inicio, fin, estado, sena_pagada, vence_en FROM turnos
		WHERE estado <> 'cancelado' AND inicio < $2 AND fin > $1`, desde, hasta)
	if err != nil {
		return nil, err
	}
	out := []Ocupado{}
	for filas.Next() {
		var o Ocupado
		if err := filas.Scan(&o.ProfesionalID, &o.Inicio, &o.Fin, &o.Estado, &o.SenaPagada, &o.VenceEn); err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	return out, filas.Err()
}

const colTurno = `id::text, codigo_publico, servicio_id::text, servicio, profesional_id::text, profesional, inicio, fin,
	cliente_nombre, cliente_telefono, cliente_email, precio, sena, estado, sena_pagada, sena_devuelta, origen, nota,
	creado_en, vence_en`

func escanearTurno(f pgx.Row, ahora time.Time) (Turno, error) {
	var t Turno
	var ini, fin time.Time
	err := f.Scan(&t.ID, &t.Codigo, &t.ServicioID, &t.Servicio, &t.ProfesionalID, &t.Profesional, &ini, &fin,
		&t.Cliente.Nombre, &t.Cliente.Telefono, &t.Cliente.Email, &t.Precio, &t.Sena, &t.Estado, &t.SenaPagada,
		&t.SenaDevuelta, &t.Origen, &t.Nota, &t.CreadoEn, &t.VenceEn)
	t.Inicio, t.Fin = fecha(ini), fecha(fin)
	o := Ocupado{Estado: t.Estado, SenaPagada: t.SenaPagada, VenceEn: t.VenceEn}
	t.Vencido = t.Estado == "reservado" && !o.Bloquea(ahora)
	return t, err
}
