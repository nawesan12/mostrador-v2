package vidriera

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5"
)

// MaxRuta: el largo máximo de un path guardado (desde y hacia).
const MaxRuta = 500

// NormalizarRuta deja un path de la tienda vieja como se guarda y se busca: sólo el path (sin
// dominio, query ni fragmento), decodificado, en minúsculas, sin barras repetidas ni la del final.
// Devuelve "" si no sirve (vacío, la raíz, muy largo o con caracteres de control). La misma regla
// está en la web (lib/server/importacion/rutas.ts), con los casos de testdata/rutas.json.
func NormalizarRuta(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "://"); i >= 0 && i < 10 {
		s = s[i+3:]
		if j := strings.IndexByte(s, '/'); j >= 0 {
			s = s[j:]
		} else {
			s = "/"
		}
	}
	if i := strings.IndexAny(s, "?#"); i >= 0 {
		s = s[:i]
	}
	if d, err := url.PathUnescape(s); err == nil {
		s = d
	}
	s = strings.ToLower(s)
	for strings.Contains(s, "//") {
		s = strings.ReplaceAll(s, "//", "/")
	}
	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}
	s = strings.TrimRight(s, "/")
	if s == "" || len(s) > MaxRuta || strings.ContainsFunc(s, func(r rune) bool { return unicode.IsControl(r) || r == '\\' }) {
		return ""
	}
	return s
}

// HaciaValido: un path de la tienda en Mostrador ("/", "/producto/<id>", "/catalogo"). Nunca "//…"
// ni "/\…", que el navegador leería como otro dominio, ni algo que no empiece con "/".
func HaciaValido(s string) bool {
	if !strings.HasPrefix(s, "/") || len(s) > MaxRuta {
		return false
	}
	if len(s) > 1 && (s[1] == '/' || s[1] == '\\') {
		return false
	}
	return !strings.ContainsFunc(s, func(r rune) bool { return unicode.IsControl(r) || r == '\\' || r == ' ' })
}

// Redireccion devuelve adónde mandar un path viejo de la tienda del slug (relativo a la tienda), o
// ErrNoExiste si no hay.
func Redireccion(ctx context.Context, pool tx, slug, path string) (string, error) {
	desde := NormalizarRuta(path)
	if desde == "" {
		return "", ErrNoExiste
	}
	id, err := tiendaDe(ctx, pool, slug)
	if err != nil {
		return "", err
	}
	var hacia string
	err = pool.EnTienda(ctx, id, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT hacia FROM redirecciones WHERE tienda_id = $1 AND desde = $2`, id, desde).Scan(&hacia)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNoExiste
	}
	return hacia, err
}
