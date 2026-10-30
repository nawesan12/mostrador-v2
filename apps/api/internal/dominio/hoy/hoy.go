// Package hoy: Actividad y Para hacer hoy (design/PanelCrecer.dc.html). La actividad es la
// auditoría de la tienda; las tareas salen de pedidos y productos. Chats, carritos y el certificado
// de ARCA se suman cuando esas integraciones existan.
package hoy

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}

var (
	ErrNoExiste = errors.New("no existe")
	argentina   = time.FixedZone("ART", -3*3600)
	tareas      = []string{"preparar", "chats", "carritos", "transferencias", "certificado", "reponer"}
)

// Evento es components/schemas/Evento.
type Evento struct {
	ID       string    `json:"id"`
	Tipo     string    `json:"tipo"`
	Quien    string    `json:"quien"`
	Foto     *string   `json:"foto"`
	Texto    string    `json:"texto"`
	Cuando   time.Time `json:"cuando"`
	Donde    string    `json:"donde"`
	Deshacer bool      `json:"deshacer"`
	Deshecho bool      `json:"deshecho"`
}

// Actividad: los últimos 200 registros de la auditoría.
func Actividad(ctx context.Context, pool tx, tiendaID string) ([]Evento, error) {
	lista := []Evento{}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		filas, err := t.Query(ctx, `SELECT id, actor_tipo, coalesce(actor_id, ''), accion, entidad, coalesce(entidad_id, ''),
				coalesce(agente, ''), creado_en, deshecho_en IS NOT NULL
			FROM auditoria WHERE tienda_id = $1 ORDER BY creado_en DESC, id DESC LIMIT 200`, tiendaID)
		if err != nil {
			return err
		}
		for filas.Next() {
			var id int64
			var actor, actorID, accion, entidad, entidadID, agente string
			var e Evento
			if err := filas.Scan(&id, &actor, &actorID, &accion, &entidad, &entidadID, &agente, &e.Cuando, &e.Deshecho); err != nil {
				return err
			}
			e.ID = fmt.Sprint(id)
			switch actor {
			case "hornero":
				e.Tipo, e.Quien = "hornero", "Tincho"
			case "persona", "staff":
				e.Tipo, e.Quien = "equipo", "Vos"
			default:
				e.Tipo, e.Quien = "sistema", "Sistema"
			}
			e.Texto = strings.TrimSpace(accion + " " + entidad + " " + entidadID)
			e.Donde = "Panel"
			if agente != "" {
				e.Donde = agente
			}
			if e.Deshecho {
				e.Texto += " — deshecho"
			}
			lista = append(lista, e)
		}
		return filas.Err()
	})
	return lista, err
}

type Tarea struct {
	ID         string `json:"id"`
	Tipo       string `json:"tipo"`
	Titulo     string `json:"titulo"`
	Detalle    string `json:"detalle"`
	Valor      *int64 `json:"valor"`
	Aproximado bool   `json:"aproximado"`
	Accion     string `json:"accion"`
	Hecha      bool   `json:"hecha"`
}

type Grupo struct {
	Titulo string  `json:"titulo"`
	Tareas []Tarea `json:"tareas"`
}

func dia(ahora time.Time) string { return ahora.In(argentina).Format("2006-01-02") }

func plural(n int64, s, p string) string {
	if n == 1 {
		return "1 " + s
	}
	return fmt.Sprintf("%d %s", n, p)
}

// ParaHacer: las tareas de hoy.
func ParaHacer(ctx context.Context, pool tx, tiendaID string, ahora time.Time) ([]Grupo, error) {
	grupos := []Grupo{}
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var preparar, prepararMonto, transf, transfMonto, reponer int64
		var nums []int32
		if err := t.QueryRow(ctx, `SELECT
				count(*) FILTER (WHERE estado = 'Para preparar'), coalesce(sum(total) FILTER (WHERE estado = 'Para preparar'), 0),
				count(*) FILTER (WHERE estado = 'Sin pagar' AND pago = 'transf'), coalesce(sum(total) FILTER (WHERE estado = 'Sin pagar' AND pago = 'transf'), 0),
				coalesce((array_agg(numero ORDER BY numero DESC) FILTER (WHERE estado = 'Sin pagar' AND pago = 'transf'))[1:3], '{}')
			FROM pedidos WHERE tienda_id = $1`, tiendaID).Scan(&preparar, &prepararMonto, &transf, &transfMonto, &nums); err != nil {
			return err
		}
		if err := t.QueryRow(ctx, `SELECT count(*) FROM productos WHERE tienda_id = $1 AND publicado AND stock < 5`, tiendaID).Scan(&reponer); err != nil {
			return err
		}
		filas, err := t.Query(ctx, `SELECT tarea FROM tareas_hechas WHERE tienda_id = $1 AND dia = $2`, tiendaID, dia(ahora))
		if err != nil {
			return err
		}
		hechas, err := pgx.CollectRows(filas, pgx.RowTo[string])
		if err != nil {
			return err
		}
		con := func(x Tarea) Tarea { x.Hecha = slices.Contains(hechas, x.ID); return x }
		urgente, semana, hoy := []Tarea{}, []Tarea{}, []Tarea{}
		if preparar > 0 {
			urgente = append(urgente, con(Tarea{ID: "preparar", Tipo: "preparar", Titulo: "Despachar " + plural(preparar, "pedido", "pedidos"),
				Detalle: "Pedidos pagados que esperan su paquete", Valor: &prepararMonto, Accion: "Preparar"}))
		}
		if transf > 0 {
			ids := make([]string, len(nums))
			for i, n := range nums {
				ids[i] = fmt.Sprintf("#%d", n)
			}
			detalle := "Pedido " + ids[0]
			if len(ids) > 1 {
				detalle = "Pedidos " + strings.Join(ids[:len(ids)-1], ", ") + " y " + ids[len(ids)-1]
			}
			hoy = append(hoy, con(Tarea{ID: "transferencias", Tipo: "transferencias", Titulo: "Conciliar " + plural(transf, "transferencia", "transferencias"),
				Detalle: detalle, Valor: &transfMonto, Accion: "Conciliar"}))
		}
		if reponer > 0 {
			semana = append(semana, con(Tarea{ID: "reponer", Tipo: "reponer", Titulo: "Reponer " + plural(reponer, "producto", "productos"),
				Detalle: "Quedan menos de 5 unidades", Accion: "Revisar"}))
		}
		for _, g := range []Grupo{{"Urgente", urgente}, {"Hoy", hoy}, {"Esta semana", semana}} {
			if len(g.Tareas) > 0 {
				grupos = append(grupos, g)
			}
		}
		return nil
	})
	return grupos, err
}

// Marcar marca o desmarca una tarea de hoy.
func Marcar(ctx context.Context, pool tx, tiendaID, id string, hecha bool, ahora time.Time) error {
	if !slices.Contains(tareas, id) {
		return ErrNoExiste
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if hecha {
			_, err := t.Exec(ctx, `INSERT INTO tareas_hechas (tienda_id, dia, tarea) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`, tiendaID, dia(ahora), id)
			return err
		}
		_, err := t.Exec(ctx, `DELETE FROM tareas_hechas WHERE tienda_id = $1 AND dia = $2 AND tarea = $3`, tiendaID, dia(ahora), id)
		return err
	})
}

// Rutas del panel.
func Rutas(mux *http.ServeMux, pool tx, tienda func(*http.Request) string) {
	con := func(h func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			id := tienda(r)
			if id == "" {
				problema.Escribir(w, problema.SinSesion())
				return
			}
			h(w, r, id)
		}
	}
	escribir := func(w http.ResponseWriter, v any, err error) {
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}
	mux.HandleFunc("GET /v1/panel/actividad", con(func(w http.ResponseWriter, r *http.Request, id string) {
		l, err := Actividad(r.Context(), pool, id)
		escribir(w, map[string]any{"items": l}, err)
	}))
	// El deshacer genérico llega con cada dominio anotando su antes y después.
	mux.HandleFunc("POST /v1/panel/actividad/{id}/deshacer", con(func(w http.ResponseWriter, _ *http.Request, _ string) {
		problema.Escribir(w, problema.Conflicto("no_se_puede", "Ese cambio no se puede deshacer."))
	}))
	mux.HandleFunc("GET /v1/panel/para-hacer", con(func(w http.ResponseWriter, r *http.Request, id string) {
		g, err := ParaHacer(r.Context(), pool, id, time.Now())
		escribir(w, map[string]any{"grupos": g}, err)
	}))
	mux.HandleFunc("PUT /v1/panel/para-hacer/{id}", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var b struct {
			Hecha bool `json:"hecha"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&b); err != nil {
			problema.Escribir(w, problema.Invalido("Falta si está hecha."))
			return
		}
		switch err := Marcar(r.Context(), pool, id, r.PathValue("id"), b.Hecha, time.Now()); {
		case errors.Is(err, ErrNoExiste):
			problema.Escribir(w, problema.NoEncontrado("Esa tarea no existe."))
		case err != nil:
			problema.Escribir(w, problema.Interno())
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}))
}
