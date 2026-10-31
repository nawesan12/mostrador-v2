package hoy

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/roles"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
)

// Actividad (docs/decisiones.md §39): lo que anotó cada sección en la auditoría, con quién lo hizo.
// La ven la dueña y administración (roles.Tienda); se guarda un año (app.purgar_auditoria).

// Evento es components/schemas/Evento.
type Evento struct {
	ID    string `json:"id"`
	Tipo  string `json:"tipo"` // equipo | ia | hornero | soporte | sistema: quién
	Quien string `json:"quien"`
	// Rol de la persona en ese momento (Dueña, Administración, Ventas, Local).
	Rol       *string   `json:"rol"`
	Categoria string    `json:"categoria"` // de qué: producto, pedido, venta…
	Foto      *string   `json:"foto"`
	Texto     string    `json:"texto"`
	Cuando    time.Time `json:"cuando"`
	Donde     string    `json:"donde"`
	Deshacer  bool      `json:"deshacer"`
	Deshecho  bool      `json:"deshecho"`
}

// Persona es una opción del filtro por persona.
type Persona struct {
	ID     string `json:"id"`
	Nombre string `json:"nombre"`
}

// PaginaActividad es lo que devuelve GET /v1/panel/actividad.
type PaginaActividad struct {
	Items     []Evento  `json:"items"`
	Siguiente *string   `json:"siguiente"`
	Personas  []Persona `json:"personas"`
}

// FiltroActividad: persona (id de usuario, o "soporte"), categoría, desde y hasta (YYYY-MM-DD, hora
// de Argentina, inclusive) y el cursor de la página.
type FiltroActividad struct {
	Persona, Categoria, Desde, Hasta, Cursor string
}

// Categorias de la Actividad, con su nombre (el filtro «Tipo» y el CSV).
var Categorias = map[string]string{
	"producto":    "Productos",
	"pedido":      "Pedidos",
	"venta":       "Punto de venta",
	"promocion":   "Cupones y promociones",
	"cobros":      "Cobros",
	"negocio":     "Datos del negocio",
	"envios":      "Envíos y retiro",
	"dominio":     "Dominio",
	"integracion": "Integraciones",
	"personas":    "Equipo",
	"plan":        "Plan",
	"datos":       "Exportaciones",
	"sesion":      "Sesiones",
	"mitienda":    "Mi tienda",
	"soporte":     "Soporte de Tu Local Web",
}

// TamPagina es cuántos eventos trae cada página; TopeExportar, cuántos baja el CSV como mucho.
const (
	TamPagina    = 50
	TopeExportar = 20000
)

var (
	ErrFiltro = errors.New("filtro inválido")
	esDia     = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	esUUID    = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
)

// Lo que hace el equipo de soporte (equipo.accion_tienda anota "equipo.<acción>" sin frase).
var textoSoporte = map[string]string{
	"equipo.extender_prueba": "extendió la prueba gratis",
	"equipo.regalar_pack":    "sumó un pack de regalo",
	"equipo.cambiar_plan":    "cambió el plan",
	"equipo.aplicar_credito": "aplicó un crédito a la cuenta",
	"equipo.pausar":          "pausó la tienda",
	"equipo.reactivar":       "reactivó la tienda",
	"equipo.cerrar":          "cerró la tienda",
}

// consulta arma el WHERE de los filtros. Devuelve la condición y los argumentos ($1 es la tienda).
func consulta(tiendaID string, f FiltroActividad) (string, []any, error) {
	cond := []string{"tienda_id = $1"}
	args := []any{tiendaID}
	arg := func(v any) string { args = append(args, v); return fmt.Sprintf("$%d", len(args)) }
	switch {
	case f.Persona == "":
	case f.Persona == "soporte":
		cond = append(cond, "actor_tipo = 'staff'")
	case esUUID.MatchString(f.Persona):
		cond = append(cond, "actor_id = "+arg(f.Persona)+" AND actor_tipo IN ('persona', 'ia')")
	default:
		return "", nil, ErrFiltro
	}
	switch {
	case f.Categoria == "":
	case f.Categoria == "soporte":
		cond = append(cond, "actor_tipo = 'staff'")
	case Categorias[f.Categoria] != "":
		cond = append(cond, "actor_tipo <> 'staff' AND accion LIKE "+arg(f.Categoria+".%"))
	default:
		return "", nil, ErrFiltro
	}
	for _, d := range []struct {
		v, op string
		mas   int
	}{{f.Desde, ">=", 0}, {f.Hasta, "<", 1}} {
		if d.v == "" {
			continue
		}
		t, err := time.Parse("2006-01-02", d.v)
		if !esDia.MatchString(d.v) || err != nil {
			return "", nil, ErrFiltro
		}
		cond = append(cond, "creado_en "+d.op+" "+arg(time.Date(t.Year(), t.Month(), t.Day()+d.mas, 0, 0, 0, 0, argentina)))
	}
	if f.Cursor != "" {
		crudo, err := base64.RawURLEncoding.DecodeString(f.Cursor)
		micro, id, ok := strings.Cut(string(crudo), "-")
		m, e1 := strconv.ParseInt(micro, 10, 64)
		n, e2 := strconv.ParseInt(id, 10, 64)
		if err != nil || !ok || e1 != nil || e2 != nil {
			return "", nil, ErrFiltro
		}
		cond = append(cond, "(creado_en, id) < ("+arg(time.UnixMicro(m))+", "+arg(n)+")")
	}
	return strings.Join(cond, " AND "), args, nil
}

type fila struct {
	id                                                    int64
	actor, actorID, nombre, rol, ia, accion, entidad, eid string
	texto, agente                                         string
	cuando                                                time.Time
	deshecho                                              bool
}

func leerFilas(ctx context.Context, t pgx.Tx, donde string, args []any, limite int) ([]fila, error) {
	filas, err := t.Query(ctx, `SELECT id, actor_tipo, coalesce(actor_id, ''), coalesce(actor_nombre, ''), coalesce(actor_rol, ''),
			coalesce(actor_ia, ''), accion, entidad, coalesce(entidad_id, ''), coalesce(texto, ''), coalesce(agente, ''),
			creado_en, deshecho_en IS NOT NULL
		FROM auditoria WHERE `+donde+fmt.Sprintf(` ORDER BY creado_en DESC, id DESC LIMIT %d`, limite), args...)
	if err != nil {
		return nil, err
	}
	var l []fila
	for filas.Next() {
		var x fila
		if err := filas.Scan(&x.id, &x.actor, &x.actorID, &x.nombre, &x.rol, &x.ia, &x.accion, &x.entidad, &x.eid,
			&x.texto, &x.agente, &x.cuando, &x.deshecho); err != nil {
			return nil, err
		}
		l = append(l, x)
	}
	return l, filas.Err()
}

// evento arma el evento legible. yo es el usuario que mira: lo suyo dice «Vos» ("" en el CSV).
func evento(x fila, yo string) Evento {
	e := Evento{ID: fmt.Sprint(x.id), Cuando: x.cuando, Deshecho: x.deshecho, Donde: "Panel"}
	nombre := x.nombre
	if nombre == "" {
		nombre = "Alguien del equipo"
	}
	if r, ok := roles.Nombre[x.rol]; ok {
		e.Rol = &r
	}
	switch x.actor {
	case "persona":
		e.Tipo, e.Quien = "equipo", nombre
		if yo != "" && x.actorID == yo {
			e.Quien = "Vos"
		}
	case "ia":
		ia := x.ia
		if ia == "" {
			ia = "Una IA"
		}
		de := nombre
		if yo != "" && x.actorID == yo {
			de = "vos"
		}
		e.Tipo, e.Quien, e.Donde = "ia", ia+" (conectada por "+de+")", "IA conectada"
	case "hornero":
		e.Tipo, e.Quien = "hornero", "Tincho"
	case "staff":
		e.Tipo, e.Quien, e.Donde = "soporte", "Soporte de Tu Local Web", "Soporte"
	default:
		e.Tipo, e.Quien = "sistema", "Sistema"
	}
	e.Categoria, _, _ = strings.Cut(x.accion, ".")
	if x.actor == "staff" {
		e.Categoria = "soporte"
	}
	if Categorias[e.Categoria] == "" {
		e.Categoria = "otro"
	}
	switch {
	case x.texto != "":
		e.Texto = x.texto
	case textoSoporte[x.accion] != "":
		e.Texto = textoSoporte[x.accion]
	default:
		e.Texto = strings.TrimSpace(x.accion + " " + x.entidad + " " + x.eid)
	}
	if x.agente != "" && x.actor != "ia" && x.actor != "staff" {
		e.Donde = x.agente
	}
	if e.Deshecho {
		e.Texto += " — deshecho"
	}
	return e
}

// Actividad devuelve una página de la Actividad con los filtros, y las personas para el filtro.
func Actividad(ctx context.Context, pool tx, tiendaID, yo string, f FiltroActividad) (PaginaActividad, error) {
	p := PaginaActividad{Items: []Evento{}, Personas: []Persona{}}
	donde, args, err := consulta(tiendaID, f)
	if err != nil {
		return p, err
	}
	err = pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		l, err := leerFilas(ctx, t, donde, args, TamPagina+1)
		if err != nil {
			return err
		}
		if len(l) > TamPagina {
			u := l[TamPagina-1]
			s := base64.RawURLEncoding.EncodeToString(fmt.Appendf(nil, "%d-%d", u.cuando.UnixMicro(), u.id))
			p.Siguiente, l = &s, l[:TamPagina]
		}
		for _, x := range l {
			p.Items = append(p.Items, evento(x, yo))
		}
		// Quiénes aparecen (con el nombre más reciente), para el filtro.
		filas, err := t.Query(ctx, `SELECT id, nombre FROM (
				SELECT DISTINCT ON (actor_id) actor_id AS id, coalesce(actor_nombre, '') AS nombre
				FROM auditoria WHERE tienda_id = $1 AND actor_tipo IN ('persona', 'ia') AND actor_id IS NOT NULL
				ORDER BY actor_id, creado_en DESC) x
			UNION ALL
			(SELECT 'soporte', 'Soporte de Tu Local Web' FROM auditoria WHERE tienda_id = $1 AND actor_tipo = 'staff' LIMIT 1)`, tiendaID)
		if err != nil {
			return err
		}
		for filas.Next() {
			var x Persona
			if err := filas.Scan(&x.ID, &x.Nombre); err != nil {
				return err
			}
			if x.Nombre == "" {
				x.Nombre = "Alguien del equipo"
			}
			if yo != "" && x.ID == yo {
				x.Nombre += " (vos)"
			}
			p.Personas = append(p.Personas, x)
		}
		return filas.Err()
	})
	return p, err
}

// celda evita que una planilla tome un texto como fórmula.
func celda(s string) string {
	if s != "" && strings.ContainsRune("=+-@\t\r", rune(s[0])) {
		return "'" + s
	}
	return s
}

// ExportarActividad arma las filas del CSV (con los mismos filtros, hasta TopeExportar) y anota la
// descarga en la Actividad.
func ExportarActividad(ctx context.Context, pool tx, tiendaID string, f FiltroActividad) ([][]string, error) {
	f.Cursor = ""
	donde, args, err := consulta(tiendaID, f)
	if err != nil {
		return nil, err
	}
	out := [][]string{{"Fecha", "Hora", "Quién", "Rol", "Tipo", "Qué hizo", "Desde"}}
	err = pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		l, err := leerFilas(ctx, t, donde, args, TopeExportar)
		if err != nil {
			return err
		}
		for _, x := range l {
			e := evento(x, "")
			rol := ""
			if e.Rol != nil {
				rol = *e.Rol
			}
			c := e.Cuando.In(argentina)
			tipo := Categorias[e.Categoria]
			if tipo == "" {
				tipo = "Otro"
			}
			out = append(out, []string{c.Format("2006-01-02"), c.Format("15:04"), celda(e.Quien), rol, tipo, celda(e.Texto), celda(e.Donde)})
		}
		return auditoria.Registrar(ctx, t, tiendaID, auditoria.Registro{Accion: "datos.exportar", Entidad: "actividad",
			Texto: fmt.Sprintf("bajó la actividad en CSV (%d %s)", len(l), plural2(len(l), "registro", "registros"))})
	})
	return out, err
}

func plural2(n int, s, p string) string {
	if n == 1 {
		return s
	}
	return p
}
