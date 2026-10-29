package configuracion

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
)

var roles = []string{"duenio", "admin", "ventas", "deposito"}

// Permiso: los mismos que el panel (lib/panel/permisos.ts).
type permiso struct {
	ID, Titulo, Detalle string
	Def                 [4]bool
}

var permisos = []permiso{
	{"ventas", "Ver ventas y pedidos", "Lista de pedidos y clientes", [4]bool{true, true, true, true}},
	{"productos", "Editar productos y precios", "Incluye stock e importaciones", [4]bool{true, true, false, true}},
	{"whatsapp", "Responder WhatsApp", "Y aprobar respuestas de Tincho", [4]bool{true, true, true, false}},
	{"despachar", "Preparar y despachar", "Etiquetas y cambios de estado", [4]bool{true, true, true, true}},
	{"finanzas", "Ver finanzas", "Saldo, liquidaciones y Caja", [4]bool{true, true, false, false}},
	{"facturas", "Emitir facturas", "Factura A, B y C ante ARCA", [4]bool{true, true, true, false}},
	{"diseno", "Cambiar diseño de la tienda", "Plantilla, colores y secciones", [4]bool{true, true, false, false}},
	{"cuenta", "Plan, usuarios y datos", "Solo el dueño puede cambiarlo", [4]bool{true, false, false, false}},
}

func fijo(p, rol string) bool { return rol == "duenio" || p == "cuenta" }

func mes(t time.Time) string { return t.In(time.FixedZone("ART", -3*3600)).Format("2006-01") }

// Permisos devuelve la matriz.
func Permisos(ctx context.Context, pool tx, tiendaID string) ([]map[string]any, error) {
	var cambios map[string]bool
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT permisos FROM tiendas WHERE id = $1`, tiendaID).Scan(&cambios)
	})
	out := []map[string]any{}
	for _, p := range permisos {
		r := map[string]bool{}
		for i, rol := range roles {
			r[rol] = p.Def[i]
			if v, ok := cambios[p.ID+":"+rol]; ok && !fijo(p.ID, rol) {
				r[rol] = v
			}
		}
		out = append(out, map[string]any{"id": p.ID, "titulo": p.Titulo, "detalle": p.Detalle, "roles": r})
	}
	return out, err
}

// FijarPermiso cambia una celda de la matriz.
func FijarPermiso(ctx context.Context, pool tx, tiendaID, p, rol string, activo bool) error {
	if !slices.ContainsFunc(permisos, func(x permiso) bool { return x.ID == p }) || !slices.Contains(roles, rol) {
		return ErrNoExiste
	}
	if fijo(p, rol) {
		return ErrNoSePuede
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE tiendas SET permisos = permisos || jsonb_build_object($2::text, $3::boolean) WHERE id = $1`,
			tiendaID, p+":"+rol, activo)
		return err
	})
}

// CambiarRol: a nadie se le da ni se le saca el rol de dueño, y uno no se cambia a sí mismo.
func CambiarRol(ctx context.Context, pool tx, tiendaID string, q Quien, id, rol string) error {
	if !slices.Contains([]string{"admin", "ventas", "deposito"}, rol) {
		return ErrInvalido
	}
	if id == q.UsuarioID {
		return ErrNoSePuede
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
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

// Quitar saca a alguien del equipo (y le cierra las sesiones) o retira la invitación.
func Quitar(ctx context.Context, pool tx, tiendaID string, q Quien, id string) error {
	if id == q.UsuarioID {
		return ErrNoSePuede
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
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
	mux.HandleFunc("GET /v1/panel/permisos", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		v, err := Permisos(r.Context(), pool, id)
		escribir(w, map[string]any{"permisos": v}, err)
	}))
	mux.HandleFunc("PUT /v1/panel/permisos/{permiso}/{rol}", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		var b struct {
			Activo bool `json:"activo"`
		}
		if leer(w, r, &b) {
			errorDe(w, FijarPermiso(r.Context(), pool, id, r.PathValue("permiso"), r.PathValue("rol"), b.Activo))
		}
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
