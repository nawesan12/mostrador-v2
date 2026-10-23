package turnos

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

// LeerAjustes de la tienda (sin fila, los de por defecto, apagado).
func LeerAjustes(ctx context.Context, pool tx, tienda string) (Ajustes, error) {
	var a Ajustes
	err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) (err error) {
		a, err = leerAjustes(ctx, t)
		return err
	})
	return a, err
}

// GuardarAjustes reemplaza los ajustes de la tienda.
func GuardarAjustes(ctx context.Context, pool tx, tienda string, a Ajustes) (Ajustes, error) {
	if err := a.validar(); err != nil {
		return Ajustes{}, err
	}
	err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `INSERT INTO turnos_ajustes (tienda_id, activo, sena, anticipacion_min, dias_vista, horario)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (tienda_id) DO UPDATE SET activo = excluded.activo, sena = excluded.sena,
				anticipacion_min = excluded.anticipacion_min, dias_vista = excluded.dias_vista, horario = excluded.horario,
				actualizado_en = now()`, tienda, a.Activo, a.Sena, a.AnticipacionMin, a.DiasVista, a.Horario)
		return err
	})
	return a, err
}

// Servicios de la tienda, también los pausados.
func Servicios(ctx context.Context, pool tx, tienda string) ([]Servicio, error) {
	var out []Servicio
	err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) (err error) {
		out, err = leerServicios(ctx, t, false)
		return err
	})
	return out, err
}

// GuardarServicio crea (id vacío) o reemplaza un servicio. Sin orden, va al final.
func GuardarServicio(ctx context.Context, pool tx, tienda, id string, s Servicio, orden *int) (Servicio, error) {
	if orden != nil {
		s.Orden = *orden
	}
	if err := s.validar(); err != nil {
		return Servicio{}, err
	}
	if id != "" && !idValido.MatchString(id) {
		return Servicio{}, ErrNoExiste
	}
	var out Servicio
	err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) (err error) {
		if id == "" {
			out, err = escanearServicio(t.QueryRow(ctx, `INSERT INTO turnos_servicios
					(tienda_id, nombre, descripcion, precio, duracion, foto, etiqueta, activo, orden)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8,
					CASE WHEN $9 THEN $10 ELSE (SELECT coalesce(max(orden) + 1, 0) FROM turnos_servicios) END)
				RETURNING `+colServicio, tienda, s.Nombre, s.Descripcion, s.Precio, s.Duracion, s.Foto, s.Etiqueta, s.Activo,
				orden != nil, s.Orden))
			return err
		}
		out, err = escanearServicio(t.QueryRow(ctx, `UPDATE turnos_servicios SET nombre = $2, descripcion = $3, precio = $4,
				duracion = $5, foto = $6, etiqueta = $7, activo = $8, orden = CASE WHEN $9 THEN $10 ELSE orden END
			WHERE id = $1 RETURNING `+colServicio, id, s.Nombre, s.Descripcion, s.Precio, s.Duracion, s.Foto, s.Etiqueta,
			s.Activo, orden != nil, s.Orden))
		return err
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return out, ErrNoExiste
	}
	return out, err
}

// BorrarServicio: los turnos ya dados conservan el nombre copiado.
func BorrarServicio(ctx context.Context, pool tx, tienda, id string) error {
	return borrar(ctx, pool, tienda, "turnos_servicios", id)
}

// BorrarProfesional: sus turnos ya dados quedan en la agenda.
func BorrarProfesional(ctx context.Context, pool tx, tienda, id string) error {
	return borrar(ctx, pool, tienda, "turnos_profesionales", id)
}

func borrar(ctx context.Context, pool tx, tienda, tabla, id string) error {
	if !idValido.MatchString(id) {
		return ErrNoExiste
	}
	return pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		r, err := t.Exec(ctx, `DELETE FROM `+tabla+` WHERE id = $1`, id)
		if err == nil && r.RowsAffected() == 0 {
			return ErrNoExiste
		}
		return err
	})
}

// Profesionales de la tienda, también los pausados.
func Profesionales(ctx context.Context, pool tx, tienda string) ([]Profesional, error) {
	var out []Profesional
	err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) (err error) {
		out, err = leerProfesionales(ctx, t, false, false)
		return err
	})
	return out, err
}

// GuardarProfesional crea (id vacío) o reemplaza un profesional. Los servicios que no son de la
// tienda se descartan.
func GuardarProfesional(ctx context.Context, pool tx, tienda, id string, p Profesional, orden *int) (Profesional, error) {
	if orden != nil {
		p.Orden = *orden
	}
	if err := p.validar(); err != nil {
		return Profesional{}, err
	}
	if id != "" && !idValido.MatchString(id) {
		return Profesional{}, ErrNoExiste
	}
	var out Profesional
	err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) (err error) {
		var servicios []string
		if err := t.QueryRow(ctx, `SELECT coalesce(array_agg(id::text ORDER BY orden), '{}') FROM turnos_servicios
			WHERE id::text = ANY($1)`, p.Servicios).Scan(&servicios); err != nil {
			return err
		}
		if id == "" {
			out, err = escanearProfesional(t.QueryRow(ctx, `INSERT INTO turnos_profesionales
					(tienda_id, nombre, rol, foto, servicios, activo, orden)
				VALUES ($1, $2, $3, $4, $5::text[]::uuid[], $6,
					CASE WHEN $7 THEN $8 ELSE (SELECT coalesce(max(orden) + 1, 0) FROM turnos_profesionales) END)
				RETURNING `+colProfesional, tienda, p.Nombre, p.Rol, p.Foto, servicios, p.Activo, orden != nil, p.Orden))
			return err
		}
		out, err = escanearProfesional(t.QueryRow(ctx, `UPDATE turnos_profesionales SET nombre = $2, rol = $3, foto = $4,
				servicios = $5::text[]::uuid[], activo = $6, orden = CASE WHEN $7 THEN $8 ELSE orden END
			WHERE id = $1 RETURNING `+colProfesional, id, p.Nombre, p.Rol, p.Foto, servicios, p.Activo, orden != nil, p.Orden))
		return err
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return out, ErrNoExiste
	}
	return out, err
}

// Agenda: los turnos entre dos días de Argentina (inclusive), opcionalmente de un profesional.
func Agenda(ctx context.Context, pool tx, tienda, desde, hasta, profesional string, ahora time.Time) ([]Turno, error) {
	d, err1 := time.ParseInLocation("2006-01-02", desde, Zona)
	h, err2 := time.ParseInLocation("2006-01-02", hasta, Zona)
	if err1 != nil || err2 != nil || h.Before(d) || h.Sub(d) > 62*24*time.Hour || (profesional != "" && !idValido.MatchString(profesional)) {
		return nil, ErrInvalido
	}
	out := []Turno{}
	err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT `+colTurno+` FROM turnos
			WHERE inicio >= $1 AND inicio < $2 AND ($3 = '' OR profesional_id::text = $3) ORDER BY inicio, profesional`,
			d, h.AddDate(0, 0, 1), profesional)
		if err != nil {
			return err
		}
		for filas.Next() {
			tu, err := escanearTurno(filas, ahora)
			if err != nil {
				return err
			}
			out = append(out, tu)
		}
		return filas.Err()
	})
	return out, err
}

// Manual es lo que manda el panel para dar un turno a mano (sin seña). forzar acepta superponerlo.
type Manual struct {
	ServicioID    string  `json:"servicioId"`
	ProfesionalID string  `json:"profesionalId"`
	Inicio        string  `json:"inicio"`
	Nombre        string  `json:"nombre"`
	Telefono      string  `json:"telefono"`
	Email         string  `json:"email"`
	Nota          *string `json:"nota"`
	Forzar        bool    `json:"forzar"`
}

// CrearManual da un turno desde el panel: confirmado, sin seña, fuera de la grilla si hace falta.
// Superponerse con otro turno del profesional es ErrOcupado salvo que se pida forzar.
func CrearManual(ctx context.Context, pool tx, tienda string, m Manual, ahora time.Time) (Turno, error) {
	c := Cliente{Nombre: m.Nombre, Telefono: m.Telefono, Email: m.Email}
	ini, err := time.Parse(time.RFC3339, m.Inicio)
	m.Nota = vacioANil(m.Nota)
	if err != nil || c.validar(false) != nil || !idValido.MatchString(m.ServicioID) || !idValido.MatchString(m.ProfesionalID) ||
		(m.Nota != nil && largo(*m.Nota) > 500) {
		return Turno{}, ErrInvalido
	}
	var out Turno
	err = pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		s, err := escanearServicio(t.QueryRow(ctx, `SELECT `+colServicio+` FROM turnos_servicios WHERE id = $1`, m.ServicioID))
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalido
		}
		if err != nil {
			return err
		}
		var profesional string
		err = t.QueryRow(ctx, `SELECT nombre FROM turnos_profesionales WHERE id = $1 FOR UPDATE`, m.ProfesionalID).Scan(&profesional)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvalido
		}
		if err != nil {
			return err
		}
		fin := ini.Add(time.Duration(s.Duracion) * time.Minute)
		if !m.Forzar {
			os, err := ocupados(ctx, t, ini, fin)
			if err != nil {
				return err
			}
			for _, o := range os {
				if o.ProfesionalID == m.ProfesionalID && o.Bloquea(ahora) {
					return ErrOcupado
				}
			}
		}
		out, err = escanearTurno(t.QueryRow(ctx, `INSERT INTO turnos (tienda_id, codigo_publico, servicio_id, servicio,
				profesional_id, profesional, inicio, fin, cliente_nombre, cliente_telefono, cliente_email, precio, sena,
				estado, origen, nota)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, 0, 'confirmado', 'panel', $13)
			RETURNING `+colTurno, tienda, nuevoCodigo(), s.ID, s.Nombre, m.ProfesionalID, profesional, ini, fin,
			c.Nombre, c.Telefono, c.Email, s.Precio, m.Nota), ahora)
		return err
	})
	return out, err
}

// CambiarEstado de un turno desde el panel (confirmar, cancelar, atendido, ausente; y deshacer).
// Cancelar no devuelve la seña: eso es DevolverSena.
func CambiarEstado(ctx context.Context, pool tx, tienda, id, estado string, ahora time.Time) (Turno, error) {
	if !idValido.MatchString(id) {
		return Turno{}, ErrNoExiste
	}
	if !estados[estado] {
		return Turno{}, ErrInvalido
	}
	var out Turno
	err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) (err error) {
		// Fuera de 'reservado' no hay nada que esperar: venceEn se limpia.
		out, err = escanearTurno(t.QueryRow(ctx, `UPDATE turnos SET estado = $2,
				vence_en = CASE WHEN $2 = 'reservado' THEN vence_en ELSE NULL END
			WHERE id = $1 RETURNING `+colTurno, id, estado), ahora)
		return err
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return out, ErrNoExiste
	}
	return out, err
}

// DevolverSena de un turno con seña pagada, con el cobrador de Mercado Pago.
func DevolverSena(ctx context.Context, pool tx, mp Cobrador, tienda, id string, ahora time.Time) (Turno, error) {
	if !idValido.MatchString(id) {
		return Turno{}, ErrNoExiste
	}
	var tu Turno
	err := pool.EnTienda(ctx, tienda, func(t pgx.Tx) (err error) {
		tu, err = escanearTurno(t.QueryRow(ctx, `SELECT `+colTurno+` FROM turnos WHERE id = $1`, id), ahora)
		return err
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return tu, ErrNoExiste
	}
	if err != nil {
		return tu, err
	}
	if !tu.SenaPagada || tu.SenaDevuelta {
		return tu, ErrSinSena
	}
	if mp == nil {
		return tu, ErrSinCobros
	}
	if err := mp.DevolverSena(ctx, tienda, id); err != nil {
		return tu, errors.Join(ErrDevolucion, err)
	}
	err = pool.EnTienda(ctx, tienda, func(t pgx.Tx) (err error) {
		tu, err = escanearTurno(t.QueryRow(ctx, `SELECT `+colTurno+` FROM turnos WHERE id = $1`, id), ahora)
		return err
	})
	return tu, err
}
