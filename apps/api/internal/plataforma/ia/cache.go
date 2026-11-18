package ia

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// VidaRespuesta: cuánto vale una respuesta guardada.
const VidaRespuesta = 10 * time.Minute

// Cache de respuestas: la misma pregunta (normalizada) del mismo negocio, sin herramientas que
// escriban, dentro de 10 minutos, no vuelve a llamar al modelo. Vive en memoria: perderlo en un
// reinicio sólo cuesta una llamada.
type Cache struct {
	mu     sync.Mutex
	m      map[string]guardada
	Vida   time.Duration
	Ahora  func() time.Time
	maximo int
}

type guardada struct {
	valor string
	vence time.Time
}

func NuevoCache() *Cache {
	return &Cache{m: map[string]guardada{}, Vida: VidaRespuesta, Ahora: time.Now, maximo: 5000}
}

func (c *Cache) Leer(clave string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	g, ok := c.m[clave]
	if !ok || !c.Ahora().Before(g.vence) {
		delete(c.m, clave)
		return "", false
	}
	return g.valor, true
}

func (c *Cache) Guardar(clave, valor string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ahora := c.Ahora()
	if len(c.m) >= c.maximo {
		for k, g := range c.m {
			if !ahora.Before(g.vence) {
				delete(c.m, k)
			}
		}
		// Sigue lleno: se vacía (es un caché, no un registro).
		if len(c.m) >= c.maximo {
			c.m = map[string]guardada{}
		}
	}
	c.m[clave] = guardada{valor, ahora.Add(c.Vida)}
}

// Clave arma la clave de una respuesta con sus partes (negocio, rol, canal, pregunta…).
func Clave(partes ...string) string {
	h := sha256.New()
	for _, p := range partes {
		h.Write([]byte(p))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// Normalizar: minúsculas, sin tildes, sin signos y con un solo espacio («¿Cuánto vendí?» =
// «cuanto vendi»).
func Normalizar(q string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	s, _, _ := transform.String(t, strings.ToLower(q))
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return ' '
	}, s)
	return strings.Join(strings.Fields(s), " ")
}
