package tiendas

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// La dirección de la tienda (tulocalweb.com.ar/t/<slug>) se cambia desde el onboarding y desde
// Configuración › Negocio (decisiones §38). La vieja queda redirigida para siempre (301) y ninguna
// otra tienda puede tomarla: lo garantizan app.cambiar_slug y app.slugs en Postgres.

// SlugValido: de 2 a 60 letras minúsculas, números o guiones, sin guion al principio ni al final
// (el CHECK de tiendas.slug). Misma regla en la web (#lib/tienda/direccion.ts).
var SlugValido = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,58}[a-z0-9]$`)

// MaxCambiosPorDia: cuántas veces se puede cambiar la dirección en 24 h (lo aplica app.cambiar_slug).
const MaxCambiosPorDia = 5

// Reservados: los slugs que ninguna tienda puede elegir. Tiene que coincidir con
// packages/contrato/direccion.json (lo verifica TestReservadosIgualAlContrato).
var Reservados = []string{
	"admin", "api", "app", "ayuda", "blog", "carrito", "checkout", "compartir", "contacto", "cuenta",
	"demo", "ejemplo", "email", "estado", "facturacion", "google", "apple", "help", "ingresar", "legales",
	"login", "mail", "mercadopago", "mercado-pago", "mercadolibre", "mercado-libre", "mi-tienda", "mitienda",
	"mostrador", "null", "oficial", "onboarding", "panel", "partners", "precios", "prueba", "registro",
	"soporte", "staff", "test", "tienda", "tiendas", "tulocalweb", "tu-local-web", "undefined", "www",
	// Las vitrinas: las tiendas de ejemplo de las plantillas (la web las muestra en /t/<slug>).
	"tienda-lumbre", "cafe-bruma", "almacen-verde", "pmpa-store", "rosa-mosqueta", "casa-nido", "sonido-sur",
	"vinoteca-tres", "huellitas", "panaderia-lola", "panaderia-lola-clasica", "madera-viva",
	"relojeria-norte", "mate-y-co", "estudio-arcilla", "navaja", "optica-faro",
}

// Motivos de Disponibilidad. Libre y Propio se pueden elegir; Actual es la que ya tiene.
const (
	MotivoLibre     = "libre"
	MotivoActual    = "actual"
	MotivoPropio    = "propio"
	MotivoInvalido  = "invalido"
	MotivoReservado = "reservado"
	MotivoUsado     = "usado"
)

// Disponibilidad de un slug para la tienda (components/schemas/DisponibilidadDireccion).
type Disponibilidad struct {
	Slug       string `json:"slug"`
	Disponible bool   `json:"disponible"`
	Motivo     string `json:"motivo"`
	Mensaje    string `json:"mensaje"`
}

var mensajes = map[string]string{
	MotivoLibre:     "Está libre.",
	MotivoActual:    "Es la dirección que tenés ahora.",
	MotivoPropio:    "Es una dirección que ya usaste: podés volver a ella.",
	MotivoInvalido:  "Usá de 2 a 60 letras minúsculas, números o guiones (sin guion al principio ni al final).",
	MotivoReservado: "Esa dirección está reservada. Probá con otra.",
	MotivoUsado:     "Esa dirección ya la usa otra tienda. Probá con otra.",
}

// Errores de CambiarDireccion.
var (
	ErrDireccionInvalida  = errors.New(mensajes[MotivoInvalido])
	ErrDireccionReservada = errors.New(mensajes[MotivoReservado])
	ErrDireccionUsada     = errors.New(mensajes[MotivoUsado])
	ErrDireccionTope      = errors.New("Ya cambiaste la dirección 5 veces hoy. Probá de nuevo mañana.")
)

// motivoDeFormato mira lo que no necesita la base: formato y reservados ("" si pasa).
func motivoDeFormato(slug string) string {
	switch {
	case !SlugValido.MatchString(slug):
		return MotivoInvalido
	case slices.Contains(Reservados, slug):
		return MotivoReservado
	}
	return ""
}

type enTienda interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}

// DisponibilidadDireccion dice si la tienda puede pasar a usar slug.
func DisponibilidadDireccion(ctx context.Context, pool enTienda, tiendaID, slug string) (Disponibilidad, error) {
	d := Disponibilidad{Slug: slug}
	var motivo string
	err := pool.EnTienda(ctx, tiendaID, func(tx pgx.Tx) error {
		var est *string
		if err := tx.QueryRow(ctx, `SELECT app.estado_slug($1)`, slug).Scan(&est); err != nil {
			return err
		}
		if est == nil {
			return pgx.ErrNoRows
		}
		motivo = *est
		return nil
	})
	if err != nil {
		return d, err
	}
	// La actual pasa aunque hoy no fuera válida (una tienda vieja con un slug reservado).
	if motivo != MotivoActual {
		if m := motivoDeFormato(slug); m != "" {
			motivo = m
		}
	}
	d.Motivo = motivo
	d.Disponible = motivo == MotivoLibre || motivo == MotivoPropio
	d.Mensaje = mensajes[motivo]
	return d, nil
}

// CambiarDireccion pasa la tienda a slug y devuelve la dirección anterior (igual a slug si ya era
// esa). La anterior queda redirigida para siempre.
func CambiarDireccion(ctx context.Context, pool enTienda, tiendaID, slug string) (string, error) {
	switch motivoDeFormato(slug) {
	case MotivoInvalido:
		return "", ErrDireccionInvalida
	case MotivoReservado:
		return "", ErrDireccionReservada
	}
	var anterior, res string
	err := pool.EnTienda(ctx, tiendaID, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `SELECT slug FROM tiendas WHERE id = $1`, tiendaID).Scan(&anterior); err != nil {
			return err
		}
		return tx.QueryRow(ctx, `SELECT app.cambiar_slug($1)`, slug).Scan(&res)
	})
	var pg *pgconn.PgError
	if errors.As(err, &pg) && pg.Code == "23505" {
		// Una tienda vieja con ese slug que no estaba en app.slugs.
		return "", ErrDireccionUsada
	}
	if err != nil {
		return "", err
	}
	switch res {
	case "ok", "actual":
		return anterior, nil
	case "usado":
		return "", ErrDireccionUsada
	case "tope":
		return "", ErrDireccionTope
	}
	return "", errors.New("cambiar_slug: respuesta inesperada " + res)
}

func rutasDireccion(mux *http.ServeMux, pool enTienda, tienda func(*http.Request) string) {
	mux.HandleFunc("GET /v1/panel/tienda/direccion", func(w http.ResponseWriter, r *http.Request) {
		id := tienda(r)
		if id == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		d, err := DisponibilidadDireccion(r.Context(), pool, id, r.URL.Query().Get("slug"))
		if err != nil {
			problema.Escribir(w, problema.Interno())
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		escribir(w, d)
	})
	mux.HandleFunc("PUT /v1/panel/tienda/direccion", func(w http.ResponseWriter, r *http.Request) {
		id := tienda(r)
		if id == "" {
			problema.Escribir(w, problema.SinSesion())
			return
		}
		var c struct {
			Slug string `json:"slug"`
		}
		if json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<10)).Decode(&c) != nil {
			problema.Escribir(w, problema.Invalido("El pedido no es un JSON válido."))
			return
		}
		anterior, err := CambiarDireccion(r.Context(), pool, id, c.Slug)
		switch {
		case err == nil:
			escribir(w, map[string]string{"slug": c.Slug, "anterior": anterior})
		case errors.Is(err, ErrDireccionInvalida):
			problema.Escribir(w, problema.Invalido(err.Error(), problema.Campo{Ruta: "slug", Mensaje: err.Error()}))
		case errors.Is(err, ErrDireccionReservada):
			problema.Escribir(w, problema.Conflicto(MotivoReservado, err.Error()))
		case errors.Is(err, ErrDireccionUsada):
			problema.Escribir(w, problema.Conflicto(MotivoUsado, err.Error()))
		case errors.Is(err, ErrDireccionTope):
			problema.Escribir(w, problema.Conflicto("tope", err.Error()))
		default:
			problema.Escribir(w, problema.Interno())
		}
	})
}
