package configuracion

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"
)

// Los formatos de lib/tienda/medicion.ts (la migración 00027 los vuelve a chequear).
var (
	ga4Valido   = regexp.MustCompile(`^G-[A-Z0-9]{6,12}$`)
	pixelValido = regexp.MustCompile(`^[0-9]{15,16}$`)
)

// Feeds es si la tienda ya cargó el feed de productos en cada lugar.
type Feeds struct {
	GM   bool `json:"gm"`
	Meta bool `json:"meta"`
}

// Medicion es components/schemas/Medicion: GA4, el píxel de Meta y los feeds, con las cuentas de
// la tienda (Tu Local Web no pone credenciales propias).
type Medicion struct {
	GA4   *string `json:"ga4"`
	Pixel *string `json:"pixel"`
	Feeds Feeds   `json:"feeds"`
}

// LeerMedicion devuelve lo guardado de la tienda.
func LeerMedicion(ctx context.Context, pool tx, tiendaID string) (Medicion, error) {
	var m Medicion
	var feeds []string
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT ga4, pixel_meta, feeds_cargados FROM tiendas WHERE id = $1`, tiendaID).
			Scan(&m.GA4, &m.Pixel, &feeds)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return m, ErrNoExiste
	}
	m.Feeds = Feeds{GM: slices.Contains(feeds, "gm"), Meta: slices.Contains(feeds, "meta")}
	return m, err
}

func limpio(s *string, valido *regexp.Regexp, mayus bool) (*string, bool) {
	if s == nil {
		return nil, true
	}
	v := strings.TrimSpace(*s)
	if mayus {
		v = strings.ToUpper(v)
	}
	if v == "" {
		return nil, true
	}
	return &v, valido.MatchString(v)
}

// GuardarMedicion reemplaza los IDs y los feeds (un ID vacío o null lo borra).
func GuardarMedicion(ctx context.Context, pool tx, tiendaID string, m Medicion) error {
	ga4, ok1 := limpio(m.GA4, ga4Valido, true)
	pixel, ok2 := limpio(m.Pixel, pixelValido, false)
	if !ok1 || !ok2 {
		return ErrInvalido
	}
	feeds := []string{}
	if m.Feeds.GM {
		feeds = append(feeds, "gm")
	}
	if m.Feeds.Meta {
		feeds = append(feeds, "meta")
	}
	return pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		_, err := t.Exec(ctx, `UPDATE tiendas SET ga4 = $2, pixel_meta = $3, feeds_cargados = $4 WHERE id = $1`,
			tiendaID, ga4, pixel, feeds)
		return err
	})
}

func rutasMedicion(mux *http.ServeMux, pool tx, con func(func(http.ResponseWriter, *http.Request, string, Quien)) http.HandlerFunc,
	errorDe func(http.ResponseWriter, error), leer func(http.ResponseWriter, *http.Request, any) bool) {
	mux.HandleFunc("GET /v1/panel/integraciones/medicion", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		m, err := LeerMedicion(r.Context(), pool, id)
		if err != nil {
			errorDe(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(m)
	}))
	mux.HandleFunc("PUT /v1/panel/integraciones/medicion", con(func(w http.ResponseWriter, r *http.Request, id string, _ Quien) {
		var m Medicion
		if leer(w, r, &m) {
			errorDe(w, GuardarMedicion(r.Context(), pool, id, m))
		}
	}))
}
