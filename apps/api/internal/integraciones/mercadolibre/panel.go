package mercadolibre

import (
	"context"
	"strconv"
	"strings"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/ventas"
)

// PreguntasPanel: las preguntas de Mercado Libre para Reseñas y preguntas (id "ml:<pregunta>").
func (s *Servicio) PreguntasPanel(ctx context.Context, tiendaID string) ([]ventas.PreguntaPanel, error) {
	ps, err := s.Preguntas(ctx, tiendaID)
	out := make([]ventas.PreguntaPanel, 0, len(ps))
	for _, p := range ps {
		producto := p.Producto
		if producto == "" {
			producto = "Publicación " + p.ItemID
		}
		out = append(out, ventas.PreguntaPanel{ID: "ml:" + strconv.FormatInt(p.ID, 10), Producto: producto,
			Pregunta: p.Texto, Creada: p.Creada, Respuesta: p.Respuesta})
	}
	return out, err
}

// ResponderPanel responde desde Reseñas y preguntas.
func (s *Servicio) ResponderPanel(ctx context.Context, tiendaID, id, texto string) error {
	n, err := strconv.ParseInt(strings.TrimPrefix(id, "ml:"), 10, 64)
	if err != nil || !strings.HasPrefix(id, "ml:") {
		return ErrNoExiste
	}
	return s.Responder(ctx, tiendaID, n, texto)
}
