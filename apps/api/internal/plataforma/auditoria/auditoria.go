// Package auditoria anota quién hizo qué, en la misma transacción que el cambio. Es la base de la
// pantalla Actividad (docs/decisiones.md §39): guarda quién (persona del equipo o IA conectada, con
// su nombre y su rol de ese momento), qué (la acción y una frase en castellano), sobre qué y cuándo,
// con el antes y el después.
//
// Cada dominio llama a Registrar donde hace la acción, con la transacción que ya tiene la tienda
// fijada. El autor sale del contexto del pedido (ConActor lo deja ahí, después de resolver la
// sesión): lo que corre sin autor (el worker, los webhooks, la importación) no se anota.
package auditoria

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// Actor es quién hizo el cambio.
type Actor struct {
	Tipo string // persona | ia | hornero | sistema | staff | integracion
	// ID es el usuario (la persona, o quien conectó la IA).
	ID string
	// Nombre y Rol de la persona en ese momento: si después la quitan del equipo, el registro sigue
	// diciendo quién fue.
	Nombre, Rol string
	// IA: el nombre de la IA conectada (Claude, ChatGPT…), si el cambio lo hizo una IA.
	IA string
	// IP y Agente del pedido.
	IP, Agente string
}

// Registro es una entrada de la auditoría.
type Registro struct {
	// Accion es "<tipo>.<qué>": producto.precio, pedido.estado, personas.invitar… El tipo es el
	// filtro de Actividad.
	Accion, Entidad, EntidadID string
	// Texto es la frase sin el sujeto: «cambió el precio de Mate imperial de $ 12.000 a $ 13.500».
	// Nunca lleva datos sensibles enteros (claves, tokens, CBU).
	Texto          string
	Antes, Despues any
}

type claveActor struct{}
type claveSilencio struct{}

// ConActor deja el autor en el contexto del pedido.
func ConActor(ctx context.Context, a Actor) context.Context {
	return context.WithValue(ctx, claveActor{}, a)
}

// ActorDe devuelve el autor del contexto.
func ActorDe(ctx context.Context) (Actor, bool) {
	a, ok := ctx.Value(claveActor{}).(Actor)
	return a, ok && a.Tipo != ""
}

// Silenciar apaga Registrar en lo que corre con ese contexto: lo usa una acción que reusa otra y
// anota ella misma lo que hizo (duplicar un producto no es «creó el producto»).
func Silenciar(ctx context.Context) context.Context {
	return context.WithValue(ctx, claveSilencio{}, true)
}

// Anotar inserta el registro en la transacción (que ya tiene la tienda fijada) con un autor
// explícito. Devuelve el id.
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
		INSERT INTO auditoria (tienda_id, actor_tipo, actor_id, actor_nombre, actor_rol, actor_ia, accion, entidad,
			entidad_id, texto, antes, despues, ip, agente)
		VALUES ($1, $2, nullif($3, ''), nullif($4, ''), nullif($5, ''), nullif($6, ''), $7, $8, nullif($9, ''),
			nullif($10, ''), $11, $12, nullif($13, '')::inet, nullif($14, ''))
		RETURNING id`,
		tiendaID, a.Tipo, a.ID, corto(a.Nombre, 120), a.Rol, corto(a.IA, 80), r.Accion, r.Entidad, r.EntidadID,
		corto(r.Texto, 500), antes, despues, ipValida(a.IP), corto(a.Agente, 200)).Scan(&id)
	return id, err
}

// Registrar anota los registros con el autor del contexto. Sin autor (o silenciado), no hace nada.
// Corre en la misma transacción que el cambio: si no se puede anotar, el cambio tampoco queda.
func Registrar(ctx context.Context, tx pgx.Tx, tiendaID string, rs ...Registro) error {
	a, ok := ActorDe(ctx)
	if !ok || ctx.Value(claveSilencio{}) != nil || len(rs) == 0 {
		return nil
	}
	if len(rs) == 1 {
		_, err := Anotar(ctx, tx, tiendaID, a, rs[0])
		return err
	}
	// Varios (una acción sobre muchas filas): un solo INSERT.
	type fila struct {
		Accion    string          `json:"accion"`
		Entidad   string          `json:"entidad"`
		EntidadID string          `json:"entidad_id"`
		Texto     string          `json:"texto"`
		Antes     json.RawMessage `json:"antes"`
		Despues   json.RawMessage `json:"despues"`
	}
	filas := make([]fila, 0, len(rs))
	for _, r := range rs {
		antes, err := aJSON(r.Antes)
		if err != nil {
			return err
		}
		despues, err := aJSON(r.Despues)
		if err != nil {
			return err
		}
		filas = append(filas, fila{r.Accion, r.Entidad, r.EntidadID, corto(r.Texto, 500), antes, despues})
	}
	crudo, err := json.Marshal(filas)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO auditoria (tienda_id, actor_tipo, actor_id, actor_nombre, actor_rol, actor_ia, accion, entidad,
			entidad_id, texto, antes, despues, ip, agente)
		SELECT $1, $2, nullif($3, ''), nullif($4, ''), nullif($5, ''), nullif($6, ''), f.accion, f.entidad,
			nullif(f.entidad_id, ''), nullif(f.texto, ''), f.antes, f.despues, nullif($7, '')::inet, nullif($8, '')
		FROM jsonb_to_recordset($9::jsonb) AS f(accion text, entidad text, entidad_id text, texto text, antes jsonb, despues jsonb)`,
		tiendaID, a.Tipo, a.ID, corto(a.Nombre, 120), a.Rol, corto(a.IA, 80), ipValida(a.IP), corto(a.Agente, 200), crudo)
	return err
}

func aJSON(v any) ([]byte, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}

func corto(s string, n int) string {
	r := []rune(strings.TrimSpace(s))
	if len(r) > n {
		return string(r[:n-1]) + "…"
	}
	return string(r)
}

// ipValida descarta lo que no es una IP (inet rechazaría la fila entera).
func ipValida(s string) string {
	s = strings.TrimSpace(s)
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F' || c == '.' || c == ':') {
			return ""
		}
	}
	return s
}

// Pesos escribe un monto en pesos como lo lee una persona: «$ 12.000».
func Pesos(n int64) string {
	signo := ""
	if n < 0 {
		signo, n = "-", -n
	}
	s := fmt.Sprint(n)
	var b strings.Builder
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			b.WriteByte('.')
		}
		b.WriteRune(c)
	}
	return signo + "$ " + b.String()
}

// Centavos escribe un monto en centavos: «$ 12.000» o «$ 12.000,50».
func Centavos(n int64) string {
	p := Pesos(n / 100)
	if c := n % 100; c != 0 {
		if c < 0 {
			c = -c
		}
		p += fmt.Sprintf(",%02d", c)
	}
	return p
}

// Final deja ver sólo los últimos 4 caracteres de un dato sensible (CBU, CVU): «terminado en 1234».
// Nunca guarda el número entero.
func Final(s string) string {
	s = strings.TrimSpace(s)
	r := []rune(s)
	if len(r) <= 4 {
		return strings.Repeat("•", len(r))
	}
	return "terminado en " + string(r[len(r)-4:])
}

// Purgar borra lo de hace más de un año, de todas las tiendas (app.purgar_auditoria). La corre el
// worker. Devuelve cuántos registros borró.
func Purgar(ctx context.Context, pool interface {
	EnTx(context.Context, func(pgx.Tx) error) error
}) (int64, error) {
	var n int64
	err := pool.EnTx(ctx, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT app.purgar_auditoria()`).Scan(&n)
	})
	return n, err
}
