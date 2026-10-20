// Package config lee la configuración de la API del entorno. Fuera de desarrollo, lo que falte
// hace que el proceso no arranque, en vez de arrancar a medias.
package config

import (
	"fmt"
	"os"
	"strings"
)

// Config de la API y el worker.
type Config struct {
	Entorno   string // desarrollo | staging | produccion
	Direccion string
	// URLs de Postgres por rol.
	BaseApp, BaseWorker, BaseMigrador, BaseAdmin string
}

// Desarrollo dice si se está corriendo en local.
func (c Config) Desarrollo() bool { return c.Entorno == "desarrollo" }

// Leer arma la configuración desde las variables MOSTRADOR_*.
func Leer() (Config, error) {
	c := Config{
		Entorno:      valor("MOSTRADOR_ENTORNO", "desarrollo"),
		Direccion:    valor("MOSTRADOR_DIRECCION", ":4000"),
		BaseApp:      os.Getenv("MOSTRADOR_BASE_DE_DATOS"),
		BaseWorker:   os.Getenv("MOSTRADOR_BASE_DE_DATOS_WORKER"),
		BaseMigrador: os.Getenv("MOSTRADOR_BASE_DE_DATOS_MIGRADOR"),
		BaseAdmin:    os.Getenv("MOSTRADOR_BASE_DE_DATOS_ADMIN"),
	}
	switch c.Entorno {
	case "desarrollo", "staging", "produccion":
	default:
		return c, fmt.Errorf("MOSTRADOR_ENTORNO=%q: tiene que ser desarrollo, staging o produccion", c.Entorno)
	}
	if c.Desarrollo() {
		// Postgres de infra/compose.dev.yml. La base de la app es mostrador_dev (la crea el bootstrap
		// con el migrador como dueño); el admin entra a la base por defecto del contenedor.
		local := "postgres://%s@127.0.0.1:5440/%s?sslmode=disable"
		c.BaseApp = conDefecto(c.BaseApp, fmt.Sprintf(local, "mostrador_app", "mostrador_dev"))
		c.BaseWorker = conDefecto(c.BaseWorker, fmt.Sprintf(local, "mostrador_worker", "mostrador_dev"))
		c.BaseMigrador = conDefecto(c.BaseMigrador, fmt.Sprintf(local, "mostrador_migrador", "mostrador_dev"))
		c.BaseAdmin = conDefecto(c.BaseAdmin, fmt.Sprintf(local, "mostrador", "mostrador"))
	}
	return c, nil
}

// Exigir falla si falta alguna de las variables pedidas (por nombre de campo).
func (c Config) Exigir(campos ...string) error {
	vals := map[string]string{
		"BaseApp": c.BaseApp, "BaseWorker": c.BaseWorker, "BaseMigrador": c.BaseMigrador, "BaseAdmin": c.BaseAdmin,
	}
	var faltan []string
	for _, f := range campos {
		if vals[f] == "" {
			faltan = append(faltan, f)
		}
	}
	if len(faltan) > 0 {
		return fmt.Errorf("falta configurar: %s", strings.Join(faltan, ", "))
	}
	return nil
}

func valor(k, def string) string { return conDefecto(os.Getenv(k), def) }

func conDefecto(v, def string) string {
	if v == "" {
		return def
	}
	return v
}
