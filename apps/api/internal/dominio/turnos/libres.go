package turnos

import (
	"sort"
	"time"

	// La zona de Argentina viaja en el binario: el VPS puede no tener tzdata.
	_ "time/tzdata"
)

// Zona de los horarios de las tiendas (UTC−3, sin horario de verano).
var Zona = func() *time.Location {
	z, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	if err != nil {
		return time.FixedZone("ART", -3*3600)
	}
	return z
}()

// Paso de la grilla de horarios, en minutos.
const Paso = 15

// Ocupado es un turno que puede bloquear un horario (lo que la disponibilidad necesita de él).
type Ocupado struct {
	ProfesionalID string     `json:"profesionalId"`
	Inicio        time.Time  `json:"inicio"`
	Fin           time.Time  `json:"fin"`
	Estado        string     `json:"estado"`
	SenaPagada    bool       `json:"senaPagada"`
	VenceEn       *time.Time `json:"venceEn"`
}

// Bloquea: el turno ocupa su horario. Un cancelado no; un reservado sin seña pagada, sólo mientras
// no venció (por eso no hace falta un cron que los libere).
func (o Ocupado) Bloquea(ahora time.Time) bool {
	if o.Estado == "cancelado" {
		return false
	}
	if o.Estado == "reservado" && !o.SenaPagada && o.VenceEn != nil && !o.VenceEn.After(ahora) {
		return false
	}
	return true
}

// HorarioLibre es components/schemas/HorarioLibre. profesionalId es el pedido (null = "cualquiera").
type HorarioLibre struct {
	Inicio        string  `json:"inicio"`
	ProfesionalID *string `json:"profesionalId"`
}

// Consulta es todo lo que mira la disponibilidad (y la entrada de testdata/libres.json).
type Consulta struct {
	Ajustes       Ajustes       `json:"ajustes"`
	Servicio      Servicio      `json:"servicio"`
	Profesionales []Profesional `json:"profesionales"`
	Turnos        []Ocupado     `json:"turnos"`
	// nil = cualquiera.
	ProfesionalID *string   `json:"profesionalId"`
	Dia           string    `json:"dia"`
	Ahora         time.Time `json:"ahora"`
}

// minutos de "HH:MM" desde la medianoche.
func minutos(s string) (int, bool) {
	if len(s) != 5 || s[2] != ':' {
		return 0, false
	}
	h, m := int(s[0]-'0')*10+int(s[1]-'0'), int(s[3]-'0')*10+int(s[4]-'0')
	for _, c := range []byte{s[0], s[1], s[3], s[4]} {
		if c < '0' || c > '9' {
			return 0, false
		}
	}
	if h > 24 || m > 59 || (h == 24 && m != 0) {
		return 0, false
	}
	return h*60 + m, true
}

// hace: el profesional hace el servicio (sin lista, hace todos).
func (p Profesional) hace(servicioID string) bool {
	if len(p.Servicios) == 0 {
		return true
	}
	for _, s := range p.Servicios {
		if s == servicioID {
			return true
		}
	}
	return false
}

// candidatos: profesionales activos que hacen el servicio (y, si se pidió uno, sólo ese), por orden.
func (c Consulta) candidatos() []Profesional {
	var out []Profesional
	for _, p := range c.Profesionales {
		if p.Activo && p.hace(c.Servicio.ID) && (c.ProfesionalID == nil || *c.ProfesionalID == p.ID) {
			out = append(out, p)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Orden < out[j].Orden })
	return out
}

// inicios del día que caben enteros en alguna franja, sin anticipación ni turnos (ordenados).
func (c Consulta) inicios(dia time.Time) []time.Time {
	dur := c.Servicio.Duracion
	vistos := map[int]bool{}
	var ms []int
	for _, f := range c.Ajustes.Horario {
		if f.Dia != int(dia.Weekday()) {
			continue
		}
		d, ok1 := minutos(f.Desde)
		h, ok2 := minutos(f.Hasta)
		if !ok1 || !ok2 {
			continue
		}
		for m := d; m+dur <= h; m += Paso {
			if !vistos[m] {
				vistos[m] = true
				ms = append(ms, m)
			}
		}
	}
	sort.Ints(ms)
	out := make([]time.Time, 0, len(ms))
	for _, m := range ms {
		out = append(out, time.Date(dia.Year(), dia.Month(), dia.Day(), 0, m, 0, 0, Zona))
	}
	return out
}

// dia de la consulta, si está dentro de la vista (hoy y los diasVista−1 siguientes).
func (c Consulta) dia() (time.Time, bool) {
	dia, err := time.ParseInLocation("2006-01-02", c.Dia, Zona)
	if err != nil || !c.Servicio.Activo || c.Servicio.Duracion <= 0 {
		return time.Time{}, false
	}
	a := c.Ahora.In(Zona)
	hoy := time.Date(a.Year(), a.Month(), a.Day(), 0, 0, 0, 0, Zona)
	if dia.Before(hoy) || !dia.Before(hoy.AddDate(0, 0, c.Ajustes.DiasVista)) {
		return time.Time{}, false
	}
	return dia, true
}

func (c Consulta) libre(profesionalID string, t, fin time.Time) bool {
	for _, o := range c.Turnos {
		if o.ProfesionalID == profesionalID && o.Bloquea(c.Ahora) && o.Inicio.Before(fin) && o.Fin.After(t) {
			return false
		}
	}
	return true
}

// Libres devuelve los horarios del día en que se puede reservar el servicio (ESPEC, regla 1).
func Libres(c Consulta) []HorarioLibre {
	out := []HorarioLibre{}
	dia, ok := c.dia()
	if !ok {
		return out
	}
	cand := c.candidatos()
	if len(cand) == 0 {
		return out
	}
	minimo := c.Ahora.Add(time.Duration(c.Ajustes.AnticipacionMin) * time.Minute)
	dur := time.Duration(c.Servicio.Duracion) * time.Minute
	for _, t := range c.inicios(dia) {
		if t.Before(minimo) {
			continue
		}
		for _, p := range cand {
			if c.libre(p.ID, t, t.Add(dur)) {
				h := HorarioLibre{Inicio: t.Format(time.RFC3339)}
				if c.ProfesionalID != nil {
					id := *c.ProfesionalID
					h.ProfesionalID = &id
				}
				out = append(out, h)
				break
			}
		}
	}
	return out
}

// Asignar devuelve el profesional que atiende en t: el pedido, o con "cualquiera" el primero libre
// por orden. "" si t no es un horario libre.
func Asignar(c Consulta, t time.Time) string {
	dia, ok := c.dia()
	if !ok || t.Before(c.Ahora.Add(time.Duration(c.Ajustes.AnticipacionMin)*time.Minute)) {
		return ""
	}
	enGrilla := false
	for _, i := range c.inicios(dia) {
		if i.Equal(t) {
			enGrilla = true
			break
		}
	}
	if !enGrilla {
		return ""
	}
	fin := t.Add(time.Duration(c.Servicio.Duracion) * time.Minute)
	for _, p := range c.candidatos() {
		if c.libre(p.ID, t, fin) {
			return p.ID
		}
	}
	return ""
}
