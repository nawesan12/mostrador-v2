package importacion

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/listas"
)

// Fuente lee la tienda vieja de a una página. La implementa el paquete de cada plataforma
// (Tiendanube, Shopify, WooCommerce, Mercado Libre, Empretienda, CSV).
type Fuente interface {
	// Pagina devuelve una página de la etapa a partir del cursor (nil: desde el principio). Sólo
	// se lee la lista de la etapa pedida. Con Siguiente nil, la etapa terminó.
	//
	// Un error se reintenta con espera creciente (el outbox); uno que envuelve ErrDefinitivo (la
	// credencial ya no sirve, la tienda no existe) corta la importación en el momento.
	Pagina(ctx context.Context, etapa string, cursor json.RawMessage) (Pagina, error)
}

// Abrir arma la fuente de una importación. La credencial llega descifrada.
type Abrir func(ctx context.Context, o Origen) (Fuente, error)

// Origen: de dónde se importa.
type Origen struct {
	TiendaID   string
	Plataforma string
	URL        string
	Credencial []byte
}

// ErrDefinitivo: lo que no se arregla reintentando.
var ErrDefinitivo = errors.New("importacion: error definitivo de la fuente")

// Pagina es lo que devuelve la fuente en cada paso: sólo viene lleno lo de la etapa.
type Pagina struct {
	Productos     []ProductoViejo
	Clientes      []ClienteViejo
	Pedidos       []PedidoViejo
	Redirecciones []URLVieja
	// Siguiente es el cursor de la próxima página (lo que la fuente entienda, hasta 2 KB); nil: terminó.
	Siguiente json.RawMessage
}

// ProductoViejo es un producto con su id y su URL en la tienda vieja. Se guarda con la misma
// función que el panel (listas.GuardarEnTx) y su URL vieja queda redirigida a la ficha nueva.
type ProductoViejo struct {
	Externo  string
	URL      string
	Producto listas.Editable
}

// ClienteViejo es la ficha de un cliente. Con email se une (en minúsculas) al que ya exista.
type ClienteViejo struct {
	Externo, Nombre, Email, Telefono, Ciudad, Provincia, Notas string
}

// PedidoViejo es un pedido de la tienda vieja: va al historial, no a Pedidos.
type PedidoViejo struct {
	Externo, Numero string
	Fecha           time.Time
	Email, Nombre   string
	// Total en pesos enteros.
	Total  int64
	Items  []ItemViejo
	Estado string
}

// ItemViejo es una línea de un pedido viejo (precio unitario, en pesos).
type ItemViejo struct {
	Nombre   string `json:"nombre"`
	Variante string `json:"variante"`
	Cantidad int    `json:"cantidad"`
	Precio   int64  `json:"precio"`
}

// URLVieja: una dirección de la tienda vieja que no es de un producto (categorías, páginas) y
// adónde va en Tu Local Web (path relativo a la tienda).
type URLVieja struct {
	Desde, Hacia string
}
