package correoargentino

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
)

// Errores que el panel y la tienda traducen.
var (
	ErrApagado         = errors.New("correoargentino: falta la clave de cifrado")
	ErrSinCredenciales = errors.New("correoargentino: faltan las credenciales de API")
	ErrSinConexion     = errors.New("correoargentino: la tienda no conectó Correo Argentino")
	ErrCuenta          = errors.New("correoargentino: email o contraseña de MiCorreo incorrectos")
	ErrNoExiste        = errors.New("correoargentino: no existe")
	ErrNoEsCorreo      = errors.New("correoargentino: el pedido no va por Correo Argentino")
)

// ErrInvalido: un dato que no pasa, con qué decirle a la tienda.
type ErrInvalido struct{ Motivo string }

func (e ErrInvalido) Error() string { return "correoargentino: " + e.Motivo }

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
	EnTx(context.Context, func(pgx.Tx) error) error
}

// Servicio de Correo Argentino.
type Servicio struct {
	Pool  tx
	Cifra *cifrado.Cifrador
	// Cliente de la API (Base: producción o pruebas).
	Cliente *Cliente
	// Credenciales de API de Tu Local Web. Vacías: cada tienda tiene que traer las suyas.
	Plataforma Credenciales
	Log        *slog.Logger
	Ahora      func() time.Time

	mu         sync.Mutex
	cotizados  map[string]cotizado
	sucursales map[string]sucursalesGuardadas
}

type cotizado struct {
	c     Cotizacion
	vence time.Time
}

type sucursalesGuardadas struct {
	s     []Sucursal
	vence time.Time
}

func (s *Servicio) ahora() time.Time {
	if s.Ahora != nil {
		return s.Ahora()
	}
	return time.Now()
}

func (s *Servicio) log() *slog.Logger {
	if s.Log != nil {
		return s.Log
	}
	return slog.Default()
}

func (s *Servicio) cliente() *Cliente {
	if s.Cliente == nil {
		s.Cliente = &Cliente{}
	}
	return s.Cliente
}

// Activo: se pueden guardar conexiones (hay clave de cifrado).
func (s *Servicio) Activo() bool { return s != nil && s.Cifra != nil }

// DesdeConfig arma el servicio. pruebas: usa apitest (las credenciales son por ambiente).
func DesdeConfig(pool tx, usuario, clave string, pruebas bool, claveCifrado string, log *slog.Logger) *Servicio {
	s := &Servicio{Pool: pool, Log: log, Plataforma: Credenciales{usuario, clave}, Cliente: &Cliente{Base: BaseProduccion}}
	if pruebas {
		s.Cliente.Base = BasePruebas
	}
	if claveCifrado != "" {
		c, err := cifrado.Nuevo(claveCifrado)
		if err != nil {
			log.Error("correoargentino: MOSTRADOR_CLAVE_CIFRADO inválida; las tiendas no van a poder conectar Correo Argentino")
		}
		s.Cifra = c
	}
	return s
}

// Provincias: los códigos de MiCorreo (una letra, la del CPA).
var Provincias = map[string]string{
	"A": "Salta", "B": "Buenos Aires", "C": "Ciudad Autónoma de Buenos Aires", "D": "San Luis", "E": "Entre Ríos",
	"F": "La Rioja", "G": "Santiago del Estero", "H": "Chaco", "J": "San Juan", "K": "Catamarca", "L": "La Pampa",
	"M": "Mendoza", "N": "Misiones", "P": "Formosa", "Q": "Neuquén", "R": "Río Negro", "S": "Santa Fe",
	"T": "Tucumán", "U": "Chubut", "V": "Tierra del Fuego", "W": "Corrientes", "X": "Córdoba", "Y": "Jujuy",
	"Z": "Santa Cruz",
}

var (
	cpValido     = regexp.MustCompile(`^\d{4}$`)
	emailValido  = regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	clienteValid = regexp.MustCompile(`^[0-9A-Za-z-]{1,40}$`)
)

// Origen: de dónde sale el paquete (remitente). El CP es obligatorio para cotizar.
type Origen struct {
	CP        string `json:"cp"`
	Calle     string `json:"calle"`
	Altura    string `json:"altura"`
	Piso      string `json:"piso"`
	Depto     string `json:"depto"`
	Localidad string `json:"localidad"`
	Provincia string `json:"provincia"`
}

// Paquete típico de la tienda: peso por unidad en gramos y la caja en centímetros.
type Paquete struct {
	Peso  int `json:"peso"`
	Alto  int `json:"alto"`
	Ancho int `json:"ancho"`
	Largo int `json:"largo"`
}

func (o *Origen) limpiar() error {
	o.CP, o.Calle, o.Altura = strings.TrimSpace(o.CP), recortar(o.Calle, 80), recortar(o.Altura, 10)
	o.Piso, o.Depto, o.Localidad = recortar(o.Piso, 3), recortar(o.Depto, 3), recortar(o.Localidad, 80)
	o.Provincia = strings.ToUpper(strings.TrimSpace(o.Provincia))
	if !cpValido.MatchString(o.CP) {
		return ErrInvalido{"El código postal de origen tiene que tener 4 números."}
	}
	if _, ok := Provincias[o.Provincia]; !ok {
		return ErrInvalido{"Elegí la provincia de origen."}
	}
	if o.Calle == "" || o.Altura == "" || o.Localidad == "" {
		return ErrInvalido{"Completá la dirección de donde salen los paquetes."}
	}
	return nil
}

func (p Paquete) valido() error {
	if p.Peso < 1 || p.Peso > 25000 {
		return ErrInvalido{"El peso por unidad va de 1 a 25.000 gramos."}
	}
	for _, l := range []int{p.Alto, p.Ancho, p.Largo} {
		if l < 1 || l > 150 {
			return ErrInvalido{"Cada medida de la caja va de 1 a 150 cm."}
		}
	}
	return nil
}

// medidas del paquete de un pedido: el peso por unidad por las unidades (tope de MiCorreo: 25 kg).
func (p Paquete) medidas(unidades int) Medidas {
	if unidades < 1 {
		unidades = 1
	}
	w := p.Peso * unidades
	if w > 25000 {
		w = 25000
	}
	return Medidas{Weight: w, Height: p.Alto, Width: p.Ancho, Length: p.Largo}
}

// conexion de una tienda, ya descifrada.
type conexion struct {
	cliente string
	email   string
	cred    Credenciales
	propias bool
	origen  Origen
	paquete Paquete
	desde   time.Time
}

func (s *Servicio) conexionDe(ctx context.Context, tiendaID string) (conexion, error) {
	var c conexion
	var cifrada []byte
	err := s.Pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT cliente, email, credencial, origen, paquete, conectada_en FROM ca_conexiones WHERE tienda_id = $1`,
			tiendaID).Scan(&c.cliente, &c.email, &cifrada, &c.origen, &c.paquete, &c.desde)
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return c, ErrSinConexion
	}
	if err != nil {
		return c, err
	}
	c.cred = s.Plataforma
	if cifrada != nil {
		if !s.Activo() {
			return c, ErrApagado
		}
		if c.cred, err = s.descifrar(tiendaID, cifrada); err != nil {
			return c, err
		}
		c.propias = true
	}
	if c.cred.vacias() {
		return c, ErrSinCredenciales
	}
	return c, nil
}

func contexto(tiendaID string) string { return "correo:" + tiendaID }

func (s *Servicio) tiendaDeSlug(ctx context.Context, slug string) (string, error) {
	var id *string
	if err := s.Pool.EnTx(ctx, func(t pgx.Tx) error {
		return t.QueryRow(ctx, `SELECT app.tienda_de_slug($1)::text`, slug).Scan(&id)
	}); err != nil {
		return "", err
	}
	if id == nil {
		return "", ErrNoExiste
	}
	return *id, nil
}
