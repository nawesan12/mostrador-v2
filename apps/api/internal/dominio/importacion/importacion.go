// Package importacion trae una tienda desde otra plataforma (Tiendanube, Shopify, WooCommerce,
// Mercado Libre, Empretienda o un CSV): productos, clientes, pedidos viejos y las direcciones de la
// tienda vieja. El trabajo avanza solo por el outbox: cada paso lee una página de la Fuente, la
// guarda y encola el paso siguiente en la misma transacción, hasta terminar. Sin polling ni cron.
package importacion

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"slices"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/cifrado"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/outbox"
)

// Plataformas de las que se puede importar (la base las chequea igual).
var Plataformas = []string{"tiendanube", "shopify", "woocommerce", "mercadolibre", "empretienda", "csv"}

// Etapas, en orden: primero los productos (los pedidos viejos los nombran), después los clientes,
// los pedidos y las direcciones que no son de un producto.
var Etapas = []string{"productos", "clientes", "pedidos", "redirecciones"}

// MaxErrores: cuántos errores guarda la importación (los últimos).
const MaxErrores = 50

// MaxPasos corta una fuente que nunca termina (un cursor que se repite).
const MaxPasos = 20000

// TipoPaso es el evento del outbox que hace un paso.
const TipoPaso = "importacion.paso"

var (
	ErrSinPermiso = errors.New("importacion: sólo dueña o admin")
	ErrInvalido   = errors.New("importacion: datos inválidos")
	ErrSinFuente  = errors.New("importacion: todavía no se puede importar desde esa plataforma")
	ErrActiva     = errors.New("importacion: ya hay una importación en curso")
	ErrNoExiste   = errors.New("importacion: no existe")
	ErrTerminada  = errors.New("importacion: ya terminó")
	ErrSinCifrado = errors.New("importacion: falta la clave de cifrado para guardar la credencial")
)

type tx interface {
	EnTienda(context.Context, string, func(pgx.Tx) error) error
}

// Servicio de importación. Fuentes, por plataforma: la que no está no se puede pedir.
type Servicio struct {
	Pool    tx
	Cifra   *cifrado.Cifrador
	Fuentes map[string]Abrir
}

// Progreso: lo guardado hasta ahora. Repetidos: lo que ya estaba (o la fuente mandó dos veces);
// recortados: lo que entró pero con algo cortado (textos largos, más de 8 fotos o de 100 líneas).
type Progreso struct {
	Productos     int `json:"productos"`
	Clientes      int `json:"clientes"`
	Pedidos       int `json:"pedidos"`
	Redirecciones int `json:"redirecciones"`
	Repetidos     int `json:"repetidos"`
	Recortados    int `json:"recortados"`
}

func (p *Progreso) sumar(o Progreso) {
	p.Productos += o.Productos
	p.Clientes += o.Clientes
	p.Pedidos += o.Pedidos
	p.Redirecciones += o.Redirecciones
	p.Repetidos += o.Repetidos
	p.Recortados += o.Recortados
}

// Error de un ítem que no entró (o de la fuente, si cortó la importación).
type Error struct {
	Etapa   string `json:"etapa"`
	Externo string `json:"externo"`
	Detalle string `json:"detalle"`
}

// Importacion es lo que ve el panel (components/schemas/Importacion).
type Importacion struct {
	ID         string     `json:"id"`
	Plataforma string     `json:"plataforma"`
	URLOrigen  string     `json:"urlOrigen"`
	Estado     string     `json:"estado"`
	Etapa      string     `json:"etapa"`
	Progreso   Progreso   `json:"progreso"`
	Errores    []Error    `json:"errores"`
	Creada     time.Time  `json:"creada"`
	Terminada  *time.Time `json:"terminada"`
}

// Nueva es el pedido de importación (components/schemas/ImportacionNueva).
type Nueva struct {
	Plataforma string `json:"plataforma"`
	URL        string `json:"url"`
	// El token o la clave de la plataforma vieja, si la fuente la necesita. Nunca vuelve.
	Credencial string `json:"credencial"`
}

func esAdmin(rol string) bool { return rol == "duenio" || rol == "admin" }

const columnas = `id::text, plataforma, url_origen, estado, etapa, progreso, errores, creada, terminada`

func escanear(f pgx.Row) (Importacion, error) {
	var i Importacion
	err := f.Scan(&i.ID, &i.Plataforma, &i.URLOrigen, &i.Estado, &i.Etapa, &i.Progreso, &i.Errores, &i.Creada, &i.Terminada)
	if i.Errores == nil {
		i.Errores = []Error{}
	}
	return i, err
}

func contextoCifrado(tienda string) string { return "importacion:" + tienda }

// Crear arranca una importación y encola su primer paso. Una sola activa por tienda.
func (s *Servicio) Crear(ctx context.Context, tienda, rol string, n Nueva) (Importacion, error) {
	if !esAdmin(rol) {
		return Importacion{}, ErrSinPermiso
	}
	if !slices.Contains(Plataformas, n.Plataforma) {
		return Importacion{}, ErrInvalido
	}
	if s.Fuentes[n.Plataforma] == nil {
		return Importacion{}, ErrSinFuente
	}
	n.URL = strings.TrimSpace(n.URL)
	if n.URL != "" {
		u, err := url.Parse(n.URL)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || len(n.URL) > 300 {
			return Importacion{}, ErrInvalido
		}
	}
	if len(n.Credencial) > 4096 {
		return Importacion{}, ErrInvalido
	}
	var cred []byte
	if n.Credencial != "" {
		if s.Cifra == nil {
			return Importacion{}, ErrSinCifrado
		}
		var err error
		if cred, err = s.Cifra.Cifrar(contextoCifrado(tienda), []byte(n.Credencial)); err != nil {
			return Importacion{}, err
		}
	}
	var imp Importacion
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var err error
		imp, err = escanear(t.QueryRow(ctx, `INSERT INTO importaciones (tienda_id, plataforma, url_origen, credencial)
			VALUES ($1, $2, $3, $4) RETURNING `+columnas, tienda, n.Plataforma, n.URL, cred))
		var pe *pgconn.PgError
		if errors.As(err, &pe) && pe.Code == "23505" {
			return ErrActiva
		}
		if err != nil {
			return err
		}
		return outbox.Publicar(ctx, t, tienda, TipoPaso, evento{ID: imp.ID, Paso: 0})
	})
	return imp, err
}

// Actual devuelve la importación en curso o, si no hay, la última; nil si nunca hubo.
func (s *Servicio) Actual(ctx context.Context, tienda string) (*Importacion, error) {
	var imp *Importacion
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		i, err := escanear(t.QueryRow(ctx, `SELECT `+columnas+` FROM importaciones WHERE tienda_id = $1
			ORDER BY estado IN ('pendiente', 'corriendo') DESC, creada DESC LIMIT 1`, tienda))
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		imp = &i
		return err
	})
	return imp, err
}

// Cancelar frena una importación en curso: lo ya guardado queda, el paso que estaba encolado no
// hace nada y la credencial se borra.
func (s *Servicio) Cancelar(ctx context.Context, tienda, rol, id string) (Importacion, error) {
	if !esAdmin(rol) {
		return Importacion{}, ErrSinPermiso
	}
	var imp Importacion
	err := s.Pool.EnTienda(ctx, tienda, func(t pgx.Tx) error {
		var err error
		imp, err = escanear(t.QueryRow(ctx, `UPDATE importaciones SET estado = 'cancelada', terminada = now(), credencial = NULL
			WHERE tienda_id = $1 AND id::text = $2 AND estado IN ('pendiente', 'corriendo') RETURNING `+columnas, tienda, id))
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		var existe bool
		if err := t.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM importaciones WHERE tienda_id = $1 AND id::text = $2)`,
			tienda, id).Scan(&existe); err != nil {
			return err
		}
		if existe {
			return ErrTerminada
		}
		return ErrNoExiste
	})
	return imp, err
}

// evento es lo que lleva cada paso en el outbox: qué importación y qué número de paso.
type evento struct {
	ID   string `json:"id"`
	Paso int    `json:"paso"`
}

// Manejadores para el worker (cmd/worker).
func (s *Servicio) Manejadores() map[string]outbox.Manejador {
	return map[string]outbox.Manejador{TipoPaso: s.Paso}
}

// marshal para jsonb (nil queda NULL).
func crudo(b json.RawMessage) any {
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	return []byte(b)
}
