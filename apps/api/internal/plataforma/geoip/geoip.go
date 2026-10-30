// Package geoip: de una IP a un lugar legible («Rosario, Santa Fe», o el país si no es Argentina),
// con la base DB-IP City Lite (formato MaxMind mmdb, CC BY 4.0) que se monta en el VPS.
//
// Sin base (local, tests) o con una IP privada, el lugar queda vacío: nunca se inventa.
package geoip

import (
	"net/netip"
	"strings"

	"github.com/oschwald/maxminddb-golang/v2"
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
)

// Ubicador dice dónde está una IP. "" cuando no se sabe.
type Ubicador interface {
	Ubicar(ip string) string
}

// Base es la base mmdb abierta (se abre una vez al arrancar y se comparte: el lector es seguro
// para usar desde varias goroutines).
type Base struct{ r *maxminddb.Reader }

// Abrir abre la base de la ruta. Ruta vacía: nil sin error (sin ubicación).
func Abrir(ruta string) (*Base, error) {
	if ruta == "" {
		return nil, nil
	}
	r, err := maxminddb.Open(ruta)
	if err != nil {
		return nil, err
	}
	return &Base{r: r}, nil
}

// Cerrar libera la base.
func (b *Base) Cerrar() {
	if b != nil && b.r != nil {
		_ = b.r.Close()
	}
}

// Registro es lo que se lee de cada red (los nombres por idioma, como los da DB-IP).
type Registro struct {
	Ciudad struct {
		Nombres map[string]string `maxminddb:"names"`
	} `maxminddb:"city"`
	Pais struct {
		ISO     string            `maxminddb:"iso_code"`
		Nombres map[string]string `maxminddb:"names"`
	} `maxminddb:"country"`
	Provincias []struct {
		Nombres map[string]string `maxminddb:"names"`
	} `maxminddb:"subdivisions"`
}

// Ubicar busca la IP. Sin base, con una IP privada, de loopback o que no está en la base: "".
func (b *Base) Ubicar(ip string) string {
	if b == nil || b.r == nil {
		return ""
	}
	a, ok := Publica(ip)
	if !ok {
		return ""
	}
	var r Registro
	if err := b.r.Lookup(a).Decode(&r); err != nil {
		return ""
	}
	return Formatear(r)
}

// Publica parsea la IP y dice si es de internet (no privada, loopback, link-local ni sin especificar).
func Publica(ip string) (netip.Addr, bool) {
	a, err := netip.ParseAddr(strings.TrimSpace(ip))
	if err != nil {
		return a, false
	}
	a = a.Unmap()
	if a.IsPrivate() || a.IsLoopback() || a.IsLinkLocalUnicast() || a.IsUnspecified() || a.IsMulticast() ||
		a.IsInterfaceLocalMulticast() || a.IsLinkLocalMulticast() {
		return a, false
	}
	return a, true
}

// provincias: DB-IP las da en inglés y sin tildes.
var provincias = map[string]string{
	"Buenos Aires F.D.": "CABA",
	"Cordoba":           "Córdoba",
	"Entre Rios":        "Entre Ríos",
	"Neuquen":           "Neuquén",
	"Rio Negro":         "Río Negro",
	"Tucuman":           "Tucumán",
}

func nombre(n map[string]string) string {
	if v := strings.TrimSpace(n["es"]); v != "" {
		return v
	}
	return strings.TrimSpace(n["en"])
}

// Formatear arma el texto: en Argentina «Ciudad, Provincia» (el barrio entre paréntesis se saca:
// esa precisión no es confiable); afuera, el nombre del país en castellano.
func Formatear(r Registro) string {
	if r.Pais.ISO == "" {
		return ""
	}
	if r.Pais.ISO != "AR" {
		if reg, err := language.ParseRegion(r.Pais.ISO); err == nil {
			if n := display.Spanish.Regions().Name(reg); n != "" {
				return n
			}
		}
		return nombre(r.Pais.Nombres)
	}
	ciudad := nombre(r.Ciudad.Nombres)
	if i := strings.Index(ciudad, " ("); i > 0 {
		ciudad = ciudad[:i]
	}
	prov := ""
	if len(r.Provincias) > 0 {
		prov = nombre(r.Provincias[0].Nombres)
		if p, ok := provincias[prov]; ok {
			prov = p
		}
	}
	switch {
	case prov == "CABA" && (ciudad == "" || ciudad == "Buenos Aires"):
		return "Ciudad de Buenos Aires"
	case ciudad == "" && prov == "":
		return "Argentina"
	case ciudad == "":
		return prov
	case prov == "":
		return ciudad
	}
	return ciudad + ", " + prov
}
