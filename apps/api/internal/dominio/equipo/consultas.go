package equipo

import (
	"context"
	"errors"
	"html"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
)

// Consulta (components/schemas/ConsultaEquipo). Junta las del formulario de contacto y los pedidos
// de arrepentimiento (Res. 424/2020) con un mismo circuito: nueva, en curso, resuelta.
type Consulta struct {
	ID          string    `json:"id"`
	Tipo        string    `json:"tipo"`
	Estado      string    `json:"estado"`
	Nombre      string    `json:"nombre"`
	Email       string    `json:"email"`
	Asunto      string    `json:"asunto"`
	Mensaje     string    `json:"mensaje"`
	CreadaEn    time.Time `json:"creadaEn"`
	AsignadoA   *string   `json:"asignadoA"`
	Codigo      string    `json:"codigo,omitempty"`
	Plazo       string    `json:"plazo,omitempty"`
	Plan        string    `json:"plan,omitempty"`
	Whatsapp    *string   `json:"whatsapp,omitempty"`
	Web         *string   `json:"web,omitempty"`
	Facturacion *string   `json:"facturacion,omitempty"`
	Origen      *string   `json:"origen,omitempty"`
}

// ConsultaDetalle suma el documento (arrepentimiento) y las respuestas.
type ConsultaDetalle struct {
	Consulta
	Documento  string      `json:"documento,omitempty"`
	Respuestas []Respuesta `json:"respuestas"`
}

// Respuesta por mail del equipo.
type Respuesta struct {
	ID       string    `json:"id"`
	Autor    string    `json:"autor"`
	Asunto   string    `json:"asunto"`
	Texto    string    `json:"texto"`
	CreadaEn time.Time `json:"creadaEn"`
}

// Bandeja (components/schemas/BandejaEquipo).
type Bandeja struct {
	Consultas []Consulta     `json:"consultas"`
	Miembros  []MiembroBreve `json:"miembros"`
}

// Estados de cada tabla ↔ el circuito del backoffice.
var (
	deContacto = map[string]string{"nueva": "nueva", "contactada": "en_curso", "cerrada": "resuelta"}
	aContacto  = map[string]string{"nueva": "nueva", "en_curso": "contactada", "resuelta": "cerrada"}
	deArr      = map[string]string{"recibido": "nueva", "en_curso": "en_curso", "procesado": "resuelta", "rechazado": "resuelta"}
	aArr       = map[string]string{"nueva": "recibido", "en_curso": "en_curso", "resuelta": "procesado"}
)

const (
	sqlContacto = `SELECT id::text, nombre, email::text, whatsapp, web, facturacion, origen, coalesce(mensaje, ''), estado,
		creada_en, asignado_a::text FROM soporte.consultas`
	sqlArr = `SELECT codigo, nombre, email::text, documento, plan, coalesce(motivo, ''), estado, creado_en, asignado_a::text
		FROM soporte.arrepentimientos`
)

func escanearContacto(f pgx.CollectableRow) (Consulta, error) {
	c := Consulta{Tipo: "contacto", Asunto: "Consulta de ventas"}
	var estado string
	err := f.Scan(&c.ID, &c.Nombre, &c.Email, &c.Whatsapp, &c.Web, &c.Facturacion, &c.Origen, &c.Mensaje, &estado, &c.CreadaEn, &c.AsignadoA)
	c.Estado = deContacto[estado]
	return c, err
}

func escanearArr(f pgx.CollectableRow) (ConsultaDetalle, error) {
	c := ConsultaDetalle{Consulta: Consulta{Tipo: "arrepentimiento"}}
	var estado string
	err := f.Scan(&c.ID, &c.Nombre, &c.Email, &c.Documento, &c.Plan, &c.Mensaje, &estado, &c.CreadaEn, &c.AsignadoA)
	c.Estado, c.Codigo, c.Asunto = deArr[estado], c.ID, "Arrepentimiento "+c.ID
	c.Plazo = fecha(diasHabiles(c.CreadaEn, 10))
	return c, err
}

// Consultas: la bandeja entera (las más nuevas primero) y la gente del equipo para asignar.
func (s *Servicio) Consultas(ctx context.Context) (Bandeja, error) {
	b := Bandeja{Consultas: []Consulta{}}
	err := s.Staff.EnTx(ctx, func(tx pgx.Tx) error {
		filas, err := tx.Query(ctx, sqlContacto+` ORDER BY creada_en DESC LIMIT 500`)
		if err != nil {
			return err
		}
		cs, err := pgx.CollectRows(filas, escanearContacto)
		if err != nil {
			return err
		}
		filas, err = tx.Query(ctx, sqlArr+` ORDER BY creado_en DESC LIMIT 500`)
		if err != nil {
			return err
		}
		as, err := pgx.CollectRows(filas, escanearArr)
		if err != nil {
			return err
		}
		for _, a := range as {
			cs = append(cs, a.Consulta) // sin el documento: sólo en el detalle (queda auditado)
		}
		sortConsultas(cs)
		b.Consultas = cs
		b.Miembros, err = miembros(ctx, tx, false)
		return err
	})
	return b, err
}

func sortConsultas(cs []Consulta) {
	for i := 1; i < len(cs); i++ {
		for j := i; j > 0 && cs[j].CreadaEn.After(cs[j-1].CreadaEn); j-- {
			cs[j], cs[j-1] = cs[j-1], cs[j]
		}
	}
}

func leerConsulta(ctx context.Context, tx pgx.Tx, tipo, id string, bloquear bool) (ConsultaDetalle, error) {
	sufijo := ""
	if bloquear {
		sufijo = " FOR UPDATE"
	}
	var c ConsultaDetalle
	var filas pgx.Rows
	var err error
	switch tipo {
	case "contacto":
		filas, err = tx.Query(ctx, sqlContacto+` WHERE id::text = $1`+sufijo, id)
		if err != nil {
			return c, err
		}
		var x Consulta
		x, err = pgx.CollectExactlyOneRow(filas, escanearContacto)
		c.Consulta = x
	case "arrepentimiento":
		filas, err = tx.Query(ctx, sqlArr+` WHERE codigo = $1`+sufijo, id)
		if err != nil {
			return c, err
		}
		c, err = pgx.CollectExactlyOneRow(filas, escanearArr)
	default:
		return c, ErrNoExiste
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return c, ErrNoExiste
	}
	return c, err
}

// Consulta con sus respuestas; ver los datos de una persona queda en la auditoría.
func (s *Servicio) Consulta(ctx context.Context, m Miembro, tipo, id string, cl cuentas.Cliente) (ConsultaDetalle, error) {
	var c ConsultaDetalle
	err := s.Staff.EnTx(ctx, func(tx pgx.Tx) error {
		var err error
		if c, err = leerConsulta(ctx, tx, tipo, id, false); err != nil {
			return err
		}
		filas, err := tx.Query(ctx, `SELECT r.id::text, m.nombre, r.asunto, r.texto, r.creada_en FROM equipo.respuestas r
			JOIN equipo.miembros m ON m.id = r.miembro_id WHERE r.tipo = $1 AND r.ref = $2 ORDER BY r.creada_en`, tipo, c.ID)
		if err != nil {
			return err
		}
		if c.Respuestas, err = pgx.CollectRows(filas, pgx.RowToStructByPos[Respuesta]); err != nil {
			return err
		}
		return anotar(ctx, tx, m, cl, "ver_consulta", "Abrió la consulta de "+c.Nombre+" ("+c.Asunto+")", "",
			map[string]string{"tipo": tipo, "id": c.ID})
	})
	return c, err
}

// Cambio de una consulta: estado, persona asignada o las dos. Asignar "" desasigna.
type Cambio struct {
	Estado    *string
	AsignadoA *string
	Asignar   bool
}

// Actualizar cambia el estado o la asignación de una consulta.
func (s *Servicio) Actualizar(ctx context.Context, m Miembro, tipo, id string, c Cambio, cl cuentas.Cliente) error {
	tabla, clave, mapa := "soporte.consultas", "id::text", aContacto
	if tipo == "arrepentimiento" {
		tabla, clave, mapa = "soporte.arrepentimientos", "codigo", aArr
	}
	if c.Estado != nil {
		if _, ok := mapa[*c.Estado]; !ok {
			return ErrInvalido{"estado", "El estado es nueva, en_curso o resuelta."}
		}
	}
	return s.Staff.EnTx(ctx, func(tx pgx.Tx) error {
		actual, err := leerConsulta(ctx, tx, tipo, id, true)
		if err != nil {
			return err
		}
		if c.Estado != nil && *c.Estado != actual.Estado {
			if _, err := tx.Exec(ctx, `UPDATE `+tabla+` SET estado = $2 WHERE `+clave+` = $1`, actual.ID, mapa[*c.Estado]); err != nil {
				return err
			}
			if err := anotar(ctx, tx, m, cl, "consulta_estado", "Pasó la consulta de "+actual.Nombre+" a "+
				map[string]string{"nueva": "nueva", "en_curso": "en curso", "resuelta": "resuelta"}[*c.Estado], "",
				map[string]string{"tipo": tipo, "id": actual.ID}); err != nil {
				return err
			}
		}
		if c.Asignar {
			var quien *string
			texto := "Dejó sin asignar la consulta de " + actual.Nombre
			if c.AsignadoA != nil && *c.AsignadoA != "" {
				var nombre string
				err := tx.QueryRow(ctx, `SELECT id::text, nombre FROM equipo.miembros WHERE id::text = $1 AND activo`, *c.AsignadoA).Scan(&c.AsignadoA, &nombre)
				if errors.Is(err, pgx.ErrNoRows) {
					return ErrInvalido{"asignadoA", "Esa persona no es del equipo."}
				}
				if err != nil {
					return err
				}
				quien, texto = c.AsignadoA, "Asignó la consulta de "+actual.Nombre+" a "+nombre
			}
			if _, err := tx.Exec(ctx, `UPDATE `+tabla+` SET asignado_a = $2::uuid WHERE `+clave+` = $1`, actual.ID, quien); err != nil {
				return err
			}
			return anotar(ctx, tx, m, cl, "consulta_asignar", texto, "", map[string]string{"tipo": tipo, "id": actual.ID})
		}
		return nil
	})
}

// Responder manda la respuesta por mail (outbox → equipo.correo) y deja la consulta en curso.
func (s *Servicio) Responder(ctx context.Context, m Miembro, tipo, id, asunto, texto string, cl cuentas.Cliente) (Respuesta, error) {
	asunto, texto = strings.TrimSpace(asunto), strings.TrimSpace(texto)
	r := Respuesta{Autor: m.Nombre, Asunto: asunto, Texto: texto}
	if l := len([]rune(asunto)); l < 3 || l > 140 {
		return r, ErrInvalido{"asunto", "Poné un asunto de 3 a 140 caracteres."}
	}
	if l := len([]rune(texto)); l < 3 || l > 4000 {
		return r, ErrInvalido{"texto", "Escribí la respuesta (hasta 4000 caracteres)."}
	}
	err := s.Staff.EnTx(ctx, func(tx pgx.Tx) error {
		c, err := leerConsulta(ctx, tx, tipo, id, true)
		if err != nil {
			return err
		}
		if err := tx.QueryRow(ctx, `INSERT INTO equipo.respuestas (tipo, ref, miembro_id, para, asunto, texto)
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING id::text, creada_en`, tipo, c.ID, m.ID, c.Email, asunto, texto).
			Scan(&r.ID, &r.CreadaEn); err != nil {
			return err
		}
		if err := encolarCorreo(ctx, tx, c.Email, asunto, texto); err != nil {
			return err
		}
		if c.Estado == "nueva" {
			tabla, clave, valor := "soporte.consultas", "id::text", "contactada"
			if tipo == "arrepentimiento" {
				tabla, clave, valor = "soporte.arrepentimientos", "codigo", "en_curso"
			}
			if _, err := tx.Exec(ctx, `UPDATE `+tabla+` SET estado = $2 WHERE `+clave+` = $1`, c.ID, valor); err != nil {
				return err
			}
		}
		return anotar(ctx, tx, m, cl, "responder", "Le respondió por mail a "+c.Nombre, "",
			map[string]string{"tipo": tipo, "id": c.ID, "asunto": asunto})
	})
	return r, err
}

// encolarCorreo deja el mail en el outbox (sin tienda: es de la plataforma). Lo manda el worker.
func encolarCorreo(ctx context.Context, tx pgx.Tx, para, asunto, texto string) error {
	cuerpo := strings.ReplaceAll(html.EscapeString(texto), "\n", "<br>")
	h := `<!doctype html><html><body style="margin:0;padding:24px;background:#FBF7F2;font-family:system-ui,sans-serif;color:#2A140C">` +
		`<p style="font-size:15.5px;line-height:1.55;margin:0">` + cuerpo + `</p>` +
		`<p style="margin:22px 0 0;font-size:13px;color:#7E6656">El equipo de Mostrador</p></body></html>`
	_, err := tx.Exec(ctx, `INSERT INTO outbox (tienda_id, tipo, datos) VALUES (NULL, 'equipo.correo', $1)`,
		map[string]string{"para": para, "asunto": asunto, "texto": texto + "\n\n— El equipo de Mostrador", "html": h})
	return err
}
