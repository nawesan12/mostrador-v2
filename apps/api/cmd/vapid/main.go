// Comando vapid: genera un par de claves VAPID para el push web del panel (docs/decisiones.md §43).
// Las claves van a la configuración del VPS (MOSTRADOR_VAPID_PUBLICA y MOSTRADOR_VAPID_PRIVADA),
// nunca al repo. Cambiarlas deja sin efecto todas las suscripciones: cada celular tiene que volver
// a activar los avisos.
package main

import (
	"fmt"
	"os"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/webpush"
)

func main() {
	pub, priv, err := webpush.GenerarClaves()
	if err != nil {
		fmt.Fprintln(os.Stderr, "no se pudieron generar las claves:", err)
		os.Exit(1)
	}
	fmt.Printf("MOSTRADOR_VAPID_PUBLICA=%s\nMOSTRADOR_VAPID_PRIVADA=%s\n", pub, priv)
}
