package tincho

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/cuentas"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// idPropuesta: el id de cada propuesta, que los clientes mandan como Idempotency-Key al confirmarla.
func idPropuesta() string {
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	return "p_" + hex.EncodeToString(b)
}

var idValido = regexp.MustCompile(`^[A-Za-z0-9_-]{8,64}$`)

// Confirmacion es lo que quedó de una confirmación: «tomada» (recién, la ejecuta quien pidió),
// «en_curso» (otro pedido la está haciendo o se cortó a la mitad), «hecha» (con su resultado).
type Confirmacion struct {
	Estado    string          `json:"estado"`
	Resultado json.RawMessage `json:"resultado,omitempty"`
}

// vidaEnCurso: una confirmación que quedó «en_curso» más que esto (el BFF se cayó a la mitad) ya no
// bloquea: se puede comprobar y volver a tomar.
const vidaEnCurso = 2 * time.Minute

// Tomar marca la propuesta para ejecutarla una sola vez. Si ya se hizo, devuelve el resultado.
func (s *Servicio) Tomar(ctx context.Context, tiendaID, id string) (Confirmacion, error) {
	var c Confirmacion
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		tag, err := t.Exec(ctx, `INSERT INTO tincho_confirmaciones (tienda_id, id, estado) VALUES ($1, $2, 'en_curso')
			ON CONFLICT DO NOTHING`, tiendaID, id)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 1 {
			c.Estado = "tomada"
			return nil
		}
		var estado string
		var res []byte
		var segundos int64
		if err := t.QueryRow(ctx, `SELECT estado, resultado, extract(epoch FROM now() - actualizado_en)::bigint
			FROM tincho_confirmaciones WHERE tienda_id = $1 AND id = $2 FOR UPDATE`, tiendaID, id).Scan(&estado, &res, &segundos); err != nil {
			return err
		}
		switch {
		case estado == "hecha":
			c.Estado, c.Resultado = "hecha", res
		case estado == "fallo" || time.Duration(segundos)*time.Second > vidaEnCurso:
			// Falló (no se hizo nada) o quedó colgada: se puede volver a intentar.
			if _, err := t.Exec(ctx, `UPDATE tincho_confirmaciones SET estado = 'en_curso', actualizado_en = now()
				WHERE tienda_id = $1 AND id = $2`, tiendaID, id); err != nil {
				return err
			}
			c.Estado = "tomada"
		default:
			c.Estado = "en_curso"
		}
		return nil
	})
	return c, err
}

// Terminar guarda cómo salió: «hecha» con el resultado, o «fallo» (no se hizo y se puede repetir).
func (s *Servicio) Terminar(ctx context.Context, tiendaID, id string, hecha bool, resultado json.RawMessage) error {
	estado := "fallo"
	if hecha {
		estado = "hecha"
	}
	if len(resultado) == 0 || !json.Valid(resultado) {
		resultado = json.RawMessage("{}")
	}
	return s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE tincho_confirmaciones SET estado = $3, resultado = $4, actualizado_en = now()
			WHERE tienda_id = $1 AND id = $2`, tiendaID, id, estado, []byte(resultado))
		return err
	})
}

func (s *Servicio) rutasConfirmaciones(mux *http.ServeMux, panel func(func(http.ResponseWriter, *http.Request, cuentas.Datos)) http.HandlerFunc) {
	mux.HandleFunc("POST /v1/panel/tincho/confirmaciones/{id}", panel(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		id := r.PathValue("id")
		if !idValido.MatchString(id) {
			problema.Escribir(w, problema.Invalido("Esa propuesta no es válida."))
			return
		}
		c, err := s.Tomar(r.Context(), d.TiendaID, id)
		if err != nil {
			s.escribirError(w, err)
			return
		}
		escribir(w, c)
	}))
	mux.HandleFunc("PUT /v1/panel/tincho/confirmaciones/{id}", panel(func(w http.ResponseWriter, r *http.Request, d cuentas.Datos) {
		id := r.PathValue("id")
		var b struct {
			Hecha     bool            `json:"hecha"`
			Resultado json.RawMessage `json:"resultado"`
		}
		if !idValido.MatchString(id) || !leer(w, r, &b) {
			if idValido.MatchString(id) {
				return
			}
			problema.Escribir(w, problema.Invalido("Esa propuesta no es válida."))
			return
		}
		err := s.Terminar(r.Context(), d.TiendaID, id, b.Hecha, b.Resultado)
		if errors.Is(err, context.Canceled) {
			return
		}
		if err != nil {
			s.escribirError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}))
}
