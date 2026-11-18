package tincho_test

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/tincho"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/ia"
)

// Medir antes de prender (docs-ia/asistente-ia.md): dos pedidos iguales a la API de verdad y el
// segundo tiene que leer del caché. Se corre a mano, nunca en el CI:
//
//	TINCHO_PRUEBA_REAL=1 ANTHROPIC_API_KEY=… go test -run TestCacheReal -v ./internal/dominio/tincho/
func TestCacheReal(t *testing.T) {
	clave := os.Getenv("ANTHROPIC_API_KEY")
	if os.Getenv("TINCHO_PRUEBA_REAL") != "1" || clave == "" || os.Getenv("CI") != "" {
		t.Skip("prueba contra la API real: TINCHO_PRUEBA_REAL=1 y ANTHROPIC_API_KEY")
	}
	m := ia.Nuevo(clave)
	datos := "DATOS DEL NEGOCIO (al 18/11/2026 12:00, hora de Argentina)\n" +
		`{"tienda":{"nombre":"La Esquina","rubro":"almacen"},"veMontos":true,"catalogo":{"masVendidos30d":[` +
		strings.Repeat(`{"nombre":"Yerba 1 kg","precio":4500,"stock":12,"vendidos30d":40},`, 30) + `{}]}}`
	for i, pregunta := range []string{"¿Qué me conviene reponer?", "¿Qué me conviene reponer primero?"} {
		p := ia.Armar(ia.Pedido{Sistema: tincho.Sistema, Herramientas: tincho.Herramientas(), Datos: datos,
			Pregunta: pregunta, MaxTokens: ia.TokensChat})
		r, err := m.Crear(context.Background(), p)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("pedido %d: entrada %d, caché leído %d, caché escrito %d, salida %d", i+1, r.Usage.InputTokens,
			r.Usage.CacheReadInputTokens, r.Usage.CacheCreationInputTokens, r.Usage.OutputTokens)
		if i == 1 && r.Usage.CacheReadInputTokens == 0 {
			t.Fatal("el segundo pedido no leyó del caché: hay un invalidador en el prefijo")
		}
	}
}
