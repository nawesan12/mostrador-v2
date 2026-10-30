package mercadopago

import (
	"context"
	"errors"
	"testing"
)

// Un id que no es numérico no llega a la cola (ni a la base).
func TestAvisarVueltaRechazaIDs(t *testing.T) {
	s := &Servicio{}
	for _, id := range []string{"", "abc", "12;DROP", "123456789012345678901"} {
		if err := s.AvisarVuelta(context.Background(), "tienda", id); !errors.Is(err, ErrInvalido) {
			t.Errorf("id %q: %v", id, err)
		}
	}
}
