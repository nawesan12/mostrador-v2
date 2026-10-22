// Package links: links de pago (el modal "Crear link de pago" del panel y la página /p/{codigo}).
// Cobrar de verdad necesita Mercado Pago conectado: hasta entonces pagar responde 409.
package links

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
	EnTx(context.Context, func(pgx.Tx) error) error
}

// ErrNoExiste: no hay link con ese código.
var ErrNoExiste = errors.New("el link no existe")

// Nuevo es lo que manda el panel.
type Nuevo struct {
	Monto    int64  `json:"monto"`
	Concepto string `json:"concepto"`
	Cuotas   *bool  `json:"cuotas"`
	Pedido   *int   `json:"pedido"`
}

// Publico es lo que ve quien paga (components/schemas/LinkPublico).
type Publico struct {
	Codigo   string `json:"codigo"`
	Tienda   string `json:"tienda"`
	Concepto string `json:"concepto"`
	Monto    int64  `json:"monto"`
	Cuotas   bool   `json:"cuotas"`
	Pagado   bool   `json:"pagado"`
	Pedido   *int   `json:"pedido"`
}

const alfabeto = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"

var formato = regexp.MustCompile(`^[2-9A-Z]{6}$`)

// codigo: corto y sin letras que se confunden (0/O, 1/I).
func codigo() string {
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	for i := range b {
		b[i] = alfabeto[int(b[i])%len(alfabeto)]
	}
	return string(b)
}

// Crear guarda el link y devuelve su código.
func Crear(ctx context.Context, pool tx, tiendaID string, n Nuevo) (string, error) {
	n.Concepto = strings.TrimSpace(n.Concepto)
	if n.Monto < 1 || n.Monto > 100_000_000 || len(n.Concepto) < 2 || len(n.Concepto) > 140 {
		return "", errors.New("link inválido")
	}
	cuotas := n.Cuotas == nil || *n.Cuotas
	var cod string
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		// 32^6 ≈ mil millones: un choque es rarísimo, pero si pasa se prueba otro código.
		for range 3 {
			cod = codigo()
			tag, err := t.Exec(ctx, `INSERT INTO links (codigo, tienda_id, monto, concepto, cuotas, pedido)
				VALUES ($1, $2, $3, $4, $5, $6) ON CONFLICT (codigo) DO NOTHING`,
				cod, tiendaID, n.Monto, n.Concepto, cuotas, n.Pedido)
			if err != nil || tag.RowsAffected() == 1 {
				return err
			}
		}
		return errors.New("no se pudo generar el código")
	})
	return cod, err
}

// tiendaDe: la tienda del link, sin sesión.
func tiendaDe(ctx context.Context, pool tx, cod string) (string, error) {
	if !formato.MatchString(cod) {
		return "", ErrNoExiste
	}
	var id *string
	err := pool.EnTx(ctx, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT app.tienda_de_link($1)::text`, cod).Scan(&id)
	})
	if err != nil {
		return "", err
	}
	if id == nil {
		return "", ErrNoExiste
	}
	return *id, nil
}

// Ver devuelve el link para la página pública.
func Ver(ctx context.Context, pool tx, cod string) (Publico, error) {
	l := Publico{Codigo: cod}
	id, err := tiendaDe(ctx, pool, cod)
	if err != nil {
		return l, err
	}
	err = pool.EnTienda(ctx, id, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT t.nombre, l.concepto, l.monto, l.cuotas, l.pagado_en IS NOT NULL, l.pedido
			FROM links l JOIN tiendas t ON t.id = l.tienda_id WHERE l.codigo = $1`, cod).
			Scan(&l.Tienda, &l.Concepto, &l.Monto, &l.Cuotas, &l.Pagado, &l.Pedido)
	})
	return l, err
}

// Rutas: crear (panel) y ver/pagar (público).
func Rutas(mux *http.ServeMux, pool tx, tienda func(*http.Request) string) {
	mux.HandleFunc("POST /v1/panel/links", func(w http.ResponseWriter, r *http.Request) {
		id := tienda(r)
		if id == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		var n Nuevo
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<10)).Decode(&n); err != nil {
			problema.Escribir(w, problema.Invalido("Mandá el monto y el concepto."))
			return
		}
		cod, err := Crear(r.Context(), pool, id, n)
		if err != nil {
			problema.Escribir(w, problema.Invalido("Revisá el monto y el concepto."))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"codigo": cod})
	})
	mux.HandleFunc("GET /v1/publico/links/{codigo}", func(w http.ResponseWriter, r *http.Request) {
		l, err := Ver(r.Context(), pool, r.PathValue("codigo"))
		switch {
		case errors.Is(err, ErrNoExiste), errors.Is(err, pgx.ErrNoRows):
			problema.Escribir(w, problema.NoEncontrado("Ese link no existe."))
		case err != nil:
			problema.Escribir(w, problema.Interno())
		default:
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(l)
		}
	})
	mux.HandleFunc("POST /v1/publico/links/{codigo}/pagar", func(w http.ResponseWriter, r *http.Request) {
		if _, err := tiendaDe(r.Context(), pool, r.PathValue("codigo")); err != nil {
			problema.Escribir(w, problema.NoEncontrado("Ese link no existe."))
			return
		}
		problema.Escribir(w, problema.Conflicto("sin_cobros", "La tienda todavía no conectó Mercado Pago."))
	})
}
