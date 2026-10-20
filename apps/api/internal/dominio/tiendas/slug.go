package tiendas

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// Slug arma la dirección de la tienda desde su nombre: "Café Bruma" → "cafe-bruma". Misma regla que
// slugDe en apps/web/src/lib/server/sandbox/rubros.ts.
func Slug(nombre string) string {
	var b strings.Builder
	guion := false
	for _, r := range norm.NFD.String(strings.ToLower(nombre)) {
		switch {
		case unicode.Is(unicode.Mn, r):
			continue // tildes
		case r == '&':
			if b.Len() > 0 && !guion {
				b.WriteByte('-')
			}
			b.WriteByte('y')
			guion = false
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			guion = false
		default:
			if b.Len() > 0 && !guion {
				b.WriteByte('-')
				guion = true
			}
		}
	}
	s := strings.Trim(b.String(), "-")
	if len(s) > 48 {
		s = strings.TrimRight(s[:48], "-")
	}
	return s
}
