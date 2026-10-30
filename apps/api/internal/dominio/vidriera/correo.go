package vidriera

import (
	"context"
	"errors"
	"strings"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/dominio/pedidos"
	"github.com/nawesan12/mostrador-v2/apps/api/internal/integraciones/correoargentino"
)

// Correo cotiza con Correo Argentino (integraciones/correoargentino). nil: no se ofrece.
type Correo interface {
	Cotizar(ctx context.Context, tiendaID, cp string, unidades int) (correoargentino.Cotizacion, error)
	Sucursal(ctx context.Context, tiendaID, provincia, codigo string) (correoargentino.Sucursal, error)
}

// ErrCorreo: Correo Argentino no cotizó al confirmar (caído o sin respuesta). Nunca se cobra un envío
// sin la cotización real.
var ErrCorreo = errors.New("vidriera: sin cotización de Correo Argentino")

// Las dos opciones de Correo en el checkout. En el pedido las dos son envio = "correo"; la modalidad
// va en la entrega (tipo D o S).
const (
	CorreoDomicilio = "correo-domicilio"
	CorreoSucursal  = "correo-sucursal"
)

func esCorreo(envio string) bool { return envio == CorreoDomicilio || envio == CorreoSucursal }

// envioCorreo vuelve a cotizar con Correo (lo que mostró el checkout no vale: el precio sale de acá) y
// arma la entrega del pedido. gratis: la tienda regala el envío desde cierto monto.
func envioCorreo(ctx context.Context, correo Correo, tiendaID string, c Compra, unidades int, gratis bool) (int64, *pedidos.Entrega, error) {
	if correo == nil {
		return 0, nil, ErrInvalido
	}
	prov := strings.ToUpper(strings.TrimSpace(c.Provincia))
	if _, ok := correoargentino.Provincias[prov]; !ok {
		return 0, nil, ErrInvalido
	}
	cot, err := correo.Cotizar(ctx, tiendaID, c.CP, unidades)
	var inv correoargentino.ErrInvalido
	if errors.As(err, &inv) || errors.Is(err, correoargentino.ErrSinConexion) {
		return 0, nil, ErrInvalido
	}
	if err != nil {
		return 0, nil, ErrCorreo
	}
	e := &pedidos.Entrega{Direccion: strings.TrimSpace(c.Direccion), CP: strings.TrimSpace(c.CP), Provincia: prov,
		Localidad: strings.TrimSpace(c.Localidad)}
	if len(e.Localidad) > 80 {
		e.Localidad = e.Localidad[:80]
	}
	var o *correoargentino.Opcion
	if c.Envio == CorreoDomicilio {
		o, e.Tipo = cot.Domicilio, "D"
		if len(e.Direccion) < 4 || e.Localidad == "" {
			return 0, nil, ErrInvalido
		}
	} else {
		o, e.Tipo = cot.Sucursal, "S"
		s, err := correo.Sucursal(ctx, tiendaID, prov, strings.TrimSpace(c.Sucursal))
		if errors.As(err, &inv) {
			return 0, nil, ErrInvalido
		}
		if err != nil {
			return 0, nil, ErrCorreo
		}
		e.Sucursal, e.SucursalNombre = s.Codigo, s.Nombre
	}
	if o == nil {
		return 0, nil, ErrInvalido
	}
	if gratis {
		return 0, e, nil
	}
	return o.Precio, e, nil
}
