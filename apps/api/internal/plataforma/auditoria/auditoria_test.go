package auditoria_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/auditoria"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

func TestAnotarGuardaAntesYDespues(t *testing.T) {
	b := pruebadb.Nueva(t)
	ctx := context.Background()
	tienda := pruebadb.CrearTienda(t, b.App, "luna")
	var id int64
	err := b.App.EnTienda(ctx, tienda, func(tx pgx.Tx) error {
		var err error
		id, err = auditoria.Anotar(ctx, tx, tienda,
			auditoria.Actor{Tipo: "persona", ID: "u1", IP: "190.2.3.4", Agente: "Chrome · macOS"},
			auditoria.Registro{Accion: "producto.precio", Entidad: "producto", EntidadID: "p1",
				Antes: map[string]int{"precio": 1000}, Despues: map[string]int{"precio": 1200}})
		return err
	})
	if err != nil || id == 0 {
		t.Fatalf("id=%d err=%v", id, err)
	}
	var antes, despues, ip string
	err = b.App.EnTienda(ctx, tienda, func(tx pgx.Tx) error {
		return tx.QueryRow(ctx, "SELECT antes->>'precio', despues->>'precio', host(ip) FROM auditoria WHERE id = $1", id).Scan(&antes, &despues, &ip)
	})
	if err != nil || antes != "1000" || despues != "1200" || ip != "190.2.3.4" {
		t.Fatalf("antes=%s despues=%s ip=%s err=%v", antes, despues, ip, err)
	}
	// La app no puede reescribir la historia: sólo marcar deshecho.
	err = b.App.EnTienda(ctx, tienda, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, "UPDATE auditoria SET accion = 'otra' WHERE id = $1", id)
		return err
	})
	if err == nil {
		t.Fatal("la app pudo reescribir una entrada de auditoría")
	}
}
