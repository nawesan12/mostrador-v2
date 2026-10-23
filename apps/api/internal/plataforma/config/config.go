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
	// Sitio público (links de los mails).
	Sitio string
	// Clave que comparte con el BFF (SvelteKit): habilita creerle X-Cliente-IP.
	ClaveBFF string
	// Correo: Resend en producción; en desarrollo, SMTP de Mailpit.
	ResendClave, CorreoDe, SMTP string
	// Bandeja del equipo: le llegan las consultas de ventas, las postulaciones de partners y los
	// arrepentimientos. Vacía: no se avisa (los pedidos quedan igual en la base).
	CorreoEquipo string
	// Ingreso con proveedores: {google|mercadopago|mercadolibre: [client id, secreto]}. Sin client
	// id, el proveedor queda apagado.
	OAuth map[string][2]string
	// WhatsApp Cloud de la plataforma (links de ingreso y de contraseña nueva).
	WhatsAppToken, WhatsAppNumeroID string
	// Mercado Pago: token de la cuenta de la plataforma (el abono) y clave de los webhooks. El OAuth
	// de las tiendas usa el mismo client id y secreto que el ingreso.
	MPToken, MPSecretoWebhook string
	// Clave maestra (32 bytes en base64) para cifrar tokens de terceros en la base.
	ClaveCifrado string
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
		Sitio:        valor("MOSTRADOR_SITIO", "https://mostrador.iamnaw.dev"),
		ClaveBFF:     os.Getenv("MOSTRADOR_CLAVE_BFF"),
		ResendClave:  os.Getenv("MOSTRADOR_RESEND_CLAVE"),
		CorreoDe:     valor("MOSTRADOR_CORREO_DE", "Mostrador <hola@mostrador.iamnaw.dev>"),
		SMTP:         os.Getenv("MOSTRADOR_SMTP"),
		CorreoEquipo: os.Getenv("MOSTRADOR_CORREO_EQUIPO"),
		OAuth: map[string][2]string{
			"google":       {os.Getenv("MOSTRADOR_GOOGLE_CLIENT_ID"), os.Getenv("MOSTRADOR_GOOGLE_SECRETO")},
			"mercadopago":  {os.Getenv("MOSTRADOR_MP_CLIENT_ID"), os.Getenv("MOSTRADOR_MP_SECRETO")},
			"mercadolibre": {os.Getenv("MOSTRADOR_ML_CLIENT_ID"), os.Getenv("MOSTRADOR_ML_SECRETO")},
		},
		WhatsAppToken:    os.Getenv("MOSTRADOR_WHATSAPP_TOKEN"),
		WhatsAppNumeroID: os.Getenv("MOSTRADOR_WHATSAPP_NUMERO_ID"),
		MPToken:          os.Getenv("MOSTRADOR_MP_TOKEN"),
		MPSecretoWebhook: os.Getenv("MOSTRADOR_MP_SECRETO_WEBHOOK"),
		ClaveCifrado:     os.Getenv("MOSTRADOR_CLAVE_CIFRADO"),
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
		c.SMTP = conDefecto(c.SMTP, "127.0.0.1:1026")
		c.Sitio = conDefecto(os.Getenv("MOSTRADOR_SITIO"), "http://localhost:5173")
	}
	return c, nil
}

// Exigir falla si falta alguna de las variables pedidas (por nombre de campo).
func (c Config) Exigir(campos ...string) error {
	vals := map[string]string{
		"BaseApp": c.BaseApp, "BaseWorker": c.BaseWorker, "BaseMigrador": c.BaseMigrador, "BaseAdmin": c.BaseAdmin,
		"ResendClave": c.ResendClave, "ClaveBFF": c.ClaveBFF,
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
