// Package auditoria anota quién hizo qué, en la misma transacción que el cambio. Es la base de la
// pantalla Actividad y del Deshacer: guarda el antes y el después.
package auditoria

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
)

// Actor es quién hizo el cambio.
type Actor struct {
	Tipo string // persona | hornero | sistema | staff | integracion
	ID   string
	// IP y Agente del pedido, si vino de una persona.
	IP, Agente string
}

// Registro es una entrada de la auditoría.
type Registro struct {
	Accion, Entidad, EntidadID string
	Antes, Despues             any
}

// Anotar inserta el registro en la transacción (que ya tiene la tienda fijada). Devuelve el id,
// que es lo que usa el Deshacer.
func Anotar(ctx context.Context, tx pgx.Tx, tiendaID string, a Actor, r Registro) (int64, error) {
	antes, err := aJSON(r.Antes)
	if err != nil {
		return 0, err
	}
	despues, err := aJSON(r.Despues)
	if err != nil {
		return 0, err
	}
	var id int64
	err = tx.QueryRow(ctx, `
		INSERT INTO auditoria (tienda_id, actor_tipo, actor_id, accion, entidad, entidad_id, antes, despues, ip, agente)
		VALUES ($1, $2, nullif($3, ''), $4, $5, nullif($6, ''), $7, $8, nullif($9, '')::inet, nullif($10, ''))
		RETURNING id`,
		tiendaID, a.Tipo, a.ID, r.Accion, r.Entidad, r.EntidadID, antes, despues, a.IP, a.Agente).Scan(&id)
	return id, err
}

func aJSON(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}
