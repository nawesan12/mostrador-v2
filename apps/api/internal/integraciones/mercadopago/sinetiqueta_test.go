package mercadopago

import "testing"

func TestSinEtiqueta(t *testing.T) {
	for entra, sale := range map[string]string{
		"ana+tienda@x.com": "ana@x.com",
		"ana@x.com":        "ana@x.com",
		"+raro@x.com":      "+raro@x.com",
		"sin-arroba":       "sin-arroba",
	} {
		if got := sinEtiqueta(entra); got != sale {
			t.Errorf("sinEtiqueta(%q) = %q, quería %q", entra, got, sale)
		}
	}
}
