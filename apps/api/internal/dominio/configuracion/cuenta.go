package configuracion

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/roles"
)

func mes(t time.Time) string { return t.In(time.FixedZone("ART", -3*3600)).Format("2006-01") }

// esDuenia: ErrDuenia si id es la dueña de la tienda.
func esDuenia(ctx context.Context, t pgx.Tx, tiendaID, id string) error {
	var es bool
	if err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM auth.membresias WHERE tienda_id = $1 AND usuario_id::text = $2 AND rol = 'duenio')`,
		tiendaID, id).Scan(&es); err != nil {
		return err
	}
	if es {
		return ErrDuenia
	}
	return nil
}

// CambiarRol: a nadie se le da ni se le saca el rol de dueña, y una no se cambia a sí misma. Vale
// para alguien del equipo o para una invitación pendiente.
func CambiarRol(ctx context.Context, pool tx, tiendaID string, q Quien, id, rol string) error {
	if !roles.Invitable(rol) {
		return ErrInvalido
	}
	if id == q.UsuarioID {
		return ErrVos
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if err := esDuenia(ctx, t, tiendaID, id); err != nil {
			return err
		}
		tag, err := t.Exec(ctx, `UPDATE auth.membresias SET rol = $3 WHERE tienda_id = $1 AND usuario_id::text = $2 AND rol <> 'duenio'`,
			tiendaID, id, rol)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 1 {
			return nil
		}
		tag, err = t.Exec(ctx, `UPDATE invitaciones SET rol = $3 WHERE tienda_id = $1 AND id::text = $2`, tiendaID, id, rol)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrNoExiste
		}
		return err
	})
}

// Quitar saca a alguien del equipo (y le cierra las sesiones que tenía abiertas en esta tienda) o
// cancela la invitación (su link deja de valer). A la dueña no se la quita.
func Quitar(ctx context.Context, pool tx, tiendaID string, q Quien, id string) error {
	if id == q.UsuarioID {
		return ErrVos
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		if err := esDuenia(ctx, t, tiendaID, id); err != nil {
			return err
		}
		tag, err := t.Exec(ctx, `DELETE FROM auth.membresias WHERE tienda_id = $1 AND usuario_id::text = $2 AND rol <> 'duenio'`, tiendaID, id)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 1 {
			_, err = t.Exec(ctx, `DELETE FROM auth.sesiones WHERE usuario_id::text = $1 AND tienda_id = $2`, id, tiendaID)
			return err
		}
		tag, err = t.Exec(ctx, `DELETE FROM invitaciones WHERE tienda_id = $1 AND id::text = $2`, tiendaID, id)
		if err == nil && tag.RowsAffected() == 0 {
			return ErrNoExiste
		}
		return err
	})
}

// Plan arma la pantalla Plan y facturación.
func Plan(ctx context.Context, pool tx, tiendaID string, ahora time.Time) (map[string]any, error) {
	var res map[string]any
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		var plan, ciclo string
		var creada time.Time
		var cancelada *time.Time
		var packs map[string]int
		var negocio Negocio
		var productos, personas int
		if err := t.QueryRow(ctx, `SELECT plan, plan_ciclo, creada_en, cancelada_en, packs, negocio,
				(SELECT count(*) FROM productos WHERE tienda_id = $1),
				(SELECT count(*) FROM auth.membresias WHERE tienda_id = $1) + (SELECT count(*) FROM invitaciones WHERE tienda_id = $1)
			FROM tiendas WHERE id = $1`, tiendaID).Scan(&plan, &ciclo, &creada, &cancelada, &packs, &negocio, &productos, &personas); err != nil {
			return err
		}
		prueba := creada.Add(14 * 24 * time.Hour)
		proximo := prueba
		for !proximo.After(ahora) {
			proximo = proximo.AddDate(0, 1, 0)
		}
		var canc any
		if cancelada != nil {
			canc = map[string]any{"hasta": proximo}
		}
		tipo := "Factura B"
		if negocio.Cuit != "" {
			tipo = "Factura A"
		}
		res = map[string]any{
			"actual": plan, "ciclo": ciclo, "enPrueba": ahora.Before(prueba), "diasPrueba": max(0, int(prueba.Sub(ahora).Hours()/24)),
			"proximoCobro": proximo,
			// Las respuestas de Hornero se cuentan cuando Hornero corra en la API.
			"uso":         map[string]any{"productos": productos, "personas": personas, "hornero": 0, "packs": packs[mes(ahora)], "sinIa": 0},
			"medio":       nil,
			"facturacion": map[string]any{"razonSocial": negocio.RazonSocial, "cuit": negocio.Cuit, "tipo": tipo},
			"facturas":    []any{},
			"cancelado":   canc,
		}
		return nil
	})
	return res, err
}

func rutasCuenta(mux *http.ServeMux, pool tx, con func(func(http.ResponseWriter, *http.Request, string, Quien)) http.HandlerFunc,
	errorDe func(http.ResponseWriter, error), leer func(http.ResponseWriter, *http.Request, any) bool) {
	escribir := func(w http.ResponseWriter, v any, err error) {
		if err != nil {
			errorDe(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}
	mux.HandleFunc("GET /v1/panel/cuenta/plan", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		v, err := Plan(r.Context(), pool, id, time.Now())
		escribir(w, v, err)
	}))
	mux.HandleFunc("POST /v1/panel/cuenta/pack", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		errorDe(w, pool.EnTienda(r.Context(), id, func(t pgx.Tx) error {
			_, err := t.Exec(r.Context(), `UPDATE tiendas SET packs = packs || jsonb_build_object($2::text, coalesce((packs->>$2)::int, 0) + 1) WHERE id = $1`, id, mes(time.Now()))
			return err
		}))
	}))
	mux.HandleFunc("DELETE /v1/panel/cuenta/pack", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		errorDe(w, pool.EnTienda(r.Context(), id, func(t pgx.Tx) error {
			tag, err := t.Exec(r.Context(), `UPDATE tiendas SET packs = packs || jsonb_build_object($2::text, (packs->>$2)::int - 1)
				WHERE id = $1 AND coalesce((packs->>$2)::int, 0) > 0`, id, mes(time.Now()))
			if err == nil && tag.RowsAffected() == 0 {
				return ErrNoExiste
			}
			return err
		}))
	}))
	mux.HandleFunc("POST /v1/panel/cuenta/cancelacion", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		var b struct {
			Motivo int    `json:"motivo"`
			Accion string `json:"accion"`
		}
		if !leer(w, r, &b) {
			return
		}
		if b.Accion == "cancelar" {
			if err := pool.EnTienda(r.Context(), id, func(t pgx.Tx) error {
				_, err := t.Exec(r.Context(), `UPDATE tiendas SET cancelada_en = now() WHERE id = $1`, id)
				return err
			}); err != nil {
				errorDe(w, err)
				return
			}
		}
		v, err := Plan(r.Context(), pool, id, time.Now())
		if err != nil {
			errorDe(w, err)
			return
		}
		escribir(w, map[string]any{"hasta": v["proximoCobro"]}, nil)
	}))
	mux.HandleFunc("DELETE /v1/panel/cuenta/cancelacion", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		errorDe(w, pool.EnTienda(r.Context(), id, func(t pgx.Tx) error {
			_, err := t.Exec(r.Context(), `UPDATE tiendas SET cancelada_en = NULL WHERE id = $1`, id)
			return err
		}))
	}))
	mux.HandleFunc("PUT /v1/panel/equipo/{id}", con(func(w http.ResponseWriter, r *http.Request, id string, q Quien) {
		var b struct{ Rol string }
		if leer(w, r, &b) {
			errorDe(w, CambiarRol(r.Context(), pool, id, q, r.PathValue("id"), b.Rol))
		}
	}))
	mux.HandleFunc("DELETE /v1/panel/equipo/{id}", con(func(w http.ResponseWriter, r *http.Request, id string, q Quien) {
		errorDe(w, Quitar(r.Context(), pool, id, q, r.PathValue("id")))
	}))
}
