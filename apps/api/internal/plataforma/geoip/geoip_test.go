package geoip

import (
	"os"
	"testing"
)

func reg(iso, ciudad, prov string) Registro {
	var r Registro
	r.Pais.ISO = iso
	r.Pais.Nombres = map[string]string{"en": "Somewhere"}
	if ciudad != "" {
		r.Ciudad.Nombres = map[string]string{"en": ciudad}
	}
	if prov != "" {
		r.Provincias = append(r.Provincias, struct {
			Nombres map[string]string `maxminddb:"names"`
		}{map[string]string{"en": prov}})
	}
	return r
}

func TestFormatear(t *testing.T) {
	casos := []struct {
		r    Registro
		want string
	}{
		{reg("AR", "Rosario (La Palmera)", "Santa Fe"), "Rosario, Santa Fe"},
		{reg("AR", "Córdoba (Padre Claret)", "Cordoba"), "Córdoba, Córdoba"},
		{reg("AR", "Buenos Aires (Palermo)", "Buenos Aires F.D."), "Ciudad de Buenos Aires"},
		{reg("AR", "Villa Crespo", "Buenos Aires F.D."), "Villa Crespo, CABA"},
		{reg("AR", "Tandil", "Buenos Aires"), "Tandil, Buenos Aires"},
		{reg("AR", "", "Neuquen"), "Neuquén"},
		{reg("AR", "", ""), "Argentina"},
		{reg("UY", "Montevideo", "Montevideo"), "Uruguay"},
		{reg("US", "Mountain View", "California"), "Estados Unidos"},
		{reg("", "", ""), ""},
	}
	for _, c := range casos {
		if got := Formatear(c.r); got != c.want {
			t.Errorf("Formatear(%v) = %q, quería %q", c.r, got, c.want)
		}
	}
}

func TestSinBaseOPrivadaNoHayLugar(t *testing.T) {
	var b *Base
	if got := b.Ubicar("190.210.32.1"); got != "" {
		t.Fatalf("sin base dio %q", got)
	}
	if b, err := Abrir(""); b != nil || err != nil {
		t.Fatalf("ruta vacía: %v %v", b, err)
	}
	if _, err := Abrir("/no/existe.mmdb"); err == nil {
		t.Fatal("abrió una base que no existe")
	}
	for _, ip := range []string{"::1", "127.0.0.1", "10.1.2.3", "192.168.0.10", "172.17.0.1", "fe80::1", "", "basura", "::ffff:127.0.0.1"} {
		if _, ok := Publica(ip); ok {
			t.Errorf("%q no es pública", ip)
		}
	}
	if _, ok := Publica("190.210.32.1"); !ok {
		t.Error("190.210.32.1 es pública")
	}
}

// Con la base real (MOSTRADOR_GEOIP apuntando al mmdb de DB-IP) se prueba la lectura de punta a punta.
func TestBaseReal(t *testing.T) {
	ruta := os.Getenv("MOSTRADOR_GEOIP")
	if ruta == "" {
		t.Skip("sin MOSTRADOR_GEOIP")
	}
	b, err := Abrir(ruta)
	if err != nil {
		t.Fatal(err)
	}
	defer b.Cerrar()
	if got := b.Ubicar("8.8.8.8"); got != "Estados Unidos" {
		t.Errorf("8.8.8.8 = %q", got)
	}
	if got := b.Ubicar("190.210.32.1"); got == "" || got == "Argentina" {
		t.Errorf("190.210.32.1 = %q", got)
	}
	if got := b.Ubicar("::1"); got != "" {
		t.Errorf("::1 = %q", got)
	}
}
