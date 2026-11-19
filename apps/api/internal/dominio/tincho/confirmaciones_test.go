package tincho_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tincho"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db/pruebadb"
)

// Confirmar una propuesta dos veces (se cortó la red, otro dispositivo) no la hace dos veces.
func TestConfirmarUnaSolaVez(t *testing.T) {
	ctx := context.Background()
	b := pruebadb.Nueva(t)
	luna := pruebadb.CrearTienda(t, b.App, "luna")
	s := &tincho.Servicio{Pool: b.App}
	c, err := s.Tomar(ctx, luna, "p_00000001")
	if err != nil || c.Estado != "tomada" {
		t.Fatalf("primera: %+v %v", c, err)
	}
	if c, _ := s.Tomar(ctx, luna, "p_00000001"); c.Estado != "en_curso" {
		t.Fatalf("mientras se hace: %+v", c)
	}
	if err := s.Terminar(ctx, luna, "p_00000001", true, json.RawMessage(`{"url":"https://x"}`)); err != nil {
		t.Fatal(err)
	}
	c, _ = s.Tomar(ctx, luna, "p_00000001")
	if c.Estado != "hecha" || string(c.Resultado) != `{"url": "https://x"}` {
		t.Fatalf("repetida: %+v %s", c, c.Resultado)
	}
	// Si falló, se puede volver a tomar.
	_, _ = s.Tomar(ctx, luna, "p_00000002")
	_ = s.Terminar(ctx, luna, "p_00000002", false, nil)
	if c, _ := s.Tomar(ctx, luna, "p_00000002"); c.Estado != "tomada" {
		t.Fatalf("después de fallar: %+v", c)
	}
}
