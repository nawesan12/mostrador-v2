package tiendas

import "testing"

// Los mismos casos que el test de slugDe del front.
func TestSlug(t *testing.T) {
	for in, want := range map[string]string{
		"Café Bruma":       "cafe-bruma",
		"Mate & Co":        "mate-y-co",
		"  Ópt!ca   Faro ": "opt-ca-faro",
		"PMPA Store":       "pmpa-store",
		"":                 "",
	} {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, quería %q", in, got, want)
		}
	}
}
