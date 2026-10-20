// Package idempotencia hace que repetir una escritura con la misma Idempotency-Key no la aplique
// dos veces: el segundo pedido recibe la respuesta guardada del primero. Es lo que permite
// reintentar un cobro o un alta sin miedo cuando se corta la conexión.
package idempotencia

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/db"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// Encabezado del pedido y de la respuesta repetida.
const (
	Encabezado = "Idempotency-Key"
	Repetida   = "Idempotent-Replayed"
	maxCuerpo  = 1 << 20
)

// Middleware envuelve las escrituras (POST, PUT, PATCH, DELETE) que traen la clave. tienda dice de
// qué tienda es el pedido (sale de la sesión); sin tienda, el pedido pasa sin idempotencia.
func Middleware(pool *db.Pool, tienda func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clave := r.Header.Get(Encabezado)
			tiendaID := tienda(r)
			if clave == "" || tiendaID == "" || !escritura(r.Method) {
				next.ServeHTTP(w, r)
				return
			}
			if len(clave) < 8 || len(clave) > 255 {
				problema.Escribir(w, problema.Invalido("La Idempotency-Key tiene que tener entre 8 y 255 caracteres."))
				return
			}
			cuerpo, err := io.ReadAll(io.LimitReader(r.Body, maxCuerpo+1))
			if err != nil || len(cuerpo) > maxCuerpo {
				problema.Escribir(w, problema.Invalido("El pedido es demasiado grande."))
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(cuerpo))
			suma := sha256.Sum256(append([]byte(r.Method+" "+r.URL.Path+"\n"), cuerpo...))
			huella := hex.EncodeToString(suma[:])

			previa, err := reservar(r.Context(), pool, tiendaID, clave, r.Method, r.URL.Path, huella)
			switch {
			case err != nil:
				problema.Escribir(w, problema.Interno())
				return
			case previa == nil:
				// Primera vez: se ejecuta y se guarda lo que respondió.
			case previa.huella != huella:
				problema.Escribir(w, problema.NoProcesable("clave_reutilizada",
					"Esa Idempotency-Key ya se usó con otro pedido. Generá una nueva."))
				return
			case previa.estado == 0:
				problema.Escribir(w, problema.Conflicto("en_curso",
					"Ese mismo pedido todavía se está procesando. Esperá unos segundos."))
				return
			default:
				if previa.tipo != "" {
					w.Header().Set("Content-Type", previa.tipo)
				}
				w.Header().Set(Repetida, "true")
				w.WriteHeader(previa.estado)
				_, _ = w.Write(previa.cuerpo)
				return
			}

			grab := &grabadora{ResponseWriter: w, estado: http.StatusOK}
			next.ServeHTTP(grab, r)
			ctx := context.WithoutCancel(r.Context())
			if grab.estado >= 500 {
				// Falló de nuestro lado: se libera la clave para que se pueda reintentar.
				_ = pool.EnTienda(ctx, tiendaID, func(tx pgx.Tx) error {
					_, err := tx.Exec(ctx, `DELETE FROM idempotencia WHERE tienda_id = $1 AND clave = $2`, tiendaID, clave)
					return err
				})
				return
			}
			_ = pool.EnTienda(ctx, tiendaID, func(tx pgx.Tx) error {
				_, err := tx.Exec(ctx, `UPDATE idempotencia SET estado = $3, respuesta = $4, tipo_contenido = nullif($5, '')
					WHERE tienda_id = $1 AND clave = $2`, tiendaID, clave, grab.estado, grab.buf.Bytes(), grab.Header().Get("Content-Type"))
				return err
			})
		})
	}
}

type previa struct {
	huella string
	estado int
	cuerpo []byte
	tipo   string
}

// reservar inserta la clave (en su propia transacción, así otro pedido igual la ve "en curso").
// Devuelve nil si la reservó; si ya existía, lo que hay guardado.
func reservar(ctx context.Context, pool *db.Pool, tiendaID, clave, metodo, ruta, huella string) (*previa, error) {
	var p *previa
	err := pool.EnTienda(ctx, tiendaID, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `INSERT INTO idempotencia (tienda_id, clave, metodo, ruta, huella)
			VALUES ($1, $2, $3, $4, $5) ON CONFLICT DO NOTHING`, tiendaID, clave, metodo, ruta, huella)
		if err != nil || tag.RowsAffected() == 1 {
			return err
		}
		var estado *int
		var tipo *string
		x := &previa{}
		err = tx.QueryRow(ctx, `SELECT huella, estado, respuesta, tipo_contenido FROM idempotencia WHERE tienda_id = $1 AND clave = $2`,
			tiendaID, clave).Scan(&x.huella, &estado, &x.cuerpo, &tipo)
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("la clave desapareció entre el insert y la lectura")
		}
		if estado != nil {
			x.estado = *estado
		}
		if tipo != nil {
			x.tipo = *tipo
		}
		p = x
		return err
	})
	return p, err
}

func escritura(m string) bool {
	return m == http.MethodPost || m == http.MethodPut || m == http.MethodPatch || m == http.MethodDelete
}

// grabadora copia la respuesta mientras se manda.
type grabadora struct {
	http.ResponseWriter
	estado int
	buf    bytes.Buffer
}

func (g *grabadora) WriteHeader(c int) { g.estado = c; g.ResponseWriter.WriteHeader(c) }
func (g *grabadora) Write(b []byte) (int, error) {
	g.buf.Write(b)
	return g.ResponseWriter.Write(b)
}
