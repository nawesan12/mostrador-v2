package pedidos

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/nawesan12/mostrador-v2/apps/api/internal/plataforma/problema"
)

// ErrSinStock: no alcanza el stock de algún producto.
var ErrSinStock = errors.New("no alcanza el stock")

// ErrProducto: algún producto no existe en la tienda.
var ErrProducto = errors.New("el producto no existe")

// PedidoNuevo es lo que manda el punto de venta o Crear pedido (components/schemas/PedidoNuevo).
type PedidoNuevo struct {
	Items []struct {
		ProductoID string `json:"productoId"`
		Cantidad   int    `json:"cantidad"`
	} `json:"items"`
	Pago    string  `json:"pago"`
	Canal   string  `json:"canal"`
	Envio   string  `json:"envio"`
	Cliente *string `json:"cliente"`
	// ClienteID: en la API los clientes salen de los compradores, así que alcanza con nombre y teléfono.
	ClienteID *string `json:"clienteId"`
	Telefono  *string `json:"telefono"`
	Nota      *string `json:"nota"`
	// Pagado en false: queda Sin pagar (se cobra con un link).
	Pagado *bool `json:"pagado"`
	// Precios cobrados en otro canal (Mercado Libre), por producto. Nunca viene del navegador.
	Precios map[string]int64 `json:"-"`
	// La tienda online: envío, descuento por transferencia, email y el código de la confirmación.
	// Los calcula el servidor (vidriera), nunca vienen del navegador.
	CostoEnvio    int64  `json:"-"`
	Descuento     int64  `json:"-"`
	Email         string `json:"-"`
	CiudadEnvio   string `json:"-"`
	CodigoPublico string `json:"-"`
	// Cupón ya validado por la vidriera y lo que descontó (va dentro de Descuento). El uso se cuenta
	// acá, en la transacción: si se agotó en el medio, ErrCupon.
	Cupon          string `json:"-"`
	DescuentoCupon int64  `json:"-"`
}

// ErrCupon: el cupón se agotó o se pausó entre que se validó y se usó.
var ErrCupon = errors.New("pedidos: cupón agotado")

var (
	pagos   = []string{"mp", "tarjeta", "payway", "modo", "uala", "transf", "efectivo"}
	canales = []string{"web", "wa", "ml", "ig", "pos", "tg"}
	envios  = []string{"andreani", "oca", "correo", "moto", "retiro", "mercadoenvios"}
)

// CrearDesdeProductos arma el pedido con los precios de la tienda y descuenta el stock, todo en una
// transacción: si un producto no alcanza, no se descuenta nada.
func CrearDesdeProductos(ctx context.Context, pool tx, tiendaID string, n PedidoNuevo) (int, int64, error) {
	if n.Envio == "" {
		n.Envio = "retiro"
	}
	if len(n.Items) == 0 || len(n.Items) > 100 || !slices.Contains(pagos, n.Pago) || !slices.Contains(canales, n.Canal) || !slices.Contains(envios, n.Envio) {
		return 0, 0, errors.New("pedido inválido")
	}
	nombre := "Consumidor final"
	if n.Cliente != nil && strings.TrimSpace(*n.Cliente) != "" {
		nombre = strings.TrimSpace(*n.Cliente)
	}
	estado := "Para preparar"
	if n.Canal == "pos" && n.Envio == "retiro" {
		estado = "Entregado"
	}
	estadoPago := map[string]string{"efectivo": "Cobrado", "transf": "Acreditada"}[n.Pago]
	if estadoPago == "" {
		estadoPago = "Aprobado"
	}
	if n.Pagado != nil && !*n.Pagado {
		estado, estadoPago = "Sin pagar", "Pendiente"
	}
	var numero int
	var total int64
	err := pool.EnTienda(ctx, tiendaID, func(t pgx.Tx) error {
		items := []Item{}
		for _, i := range n.Items {
			if i.Cantidad < 1 || i.Cantidad > 10000 {
				return errors.New("cantidad inválida")
			}
			var it Item
			// Descuenta sólo si alcanza: la condición evita vender lo que no hay con dos cajas a la vez.
			err := t.QueryRow(ctx, `UPDATE productos SET stock = stock - $3
				WHERE tienda_id = $1 AND id::text = $2 AND stock >= $3
				RETURNING nombre, foto, precio`, tiendaID, i.ProductoID, i.Cantidad).Scan(&it.Nombre, &it.Foto, &it.Precio)
			if errors.Is(err, pgx.ErrNoRows) {
				var existe bool
				_ = t.QueryRow(ctx, `SELECT true FROM productos WHERE tienda_id = $1 AND id::text = $2`, tiendaID, i.ProductoID).Scan(&existe)
				if existe {
					return ErrSinStock
				}
				return ErrProducto
			}
			if err != nil {
				return err
			}
			if p, ok := n.Precios[i.ProductoID]; ok {
				it.Precio = p
			}
			it.Cantidad = i.Cantidad
			total += it.Precio * int64(it.Cantidad)
			items = append(items, it)
		}
		var tel *string
		if n.Telefono != nil && strings.TrimSpace(*n.Telefono) != "" {
			t := strings.TrimSpace(*n.Telefono)
			tel = &t
		}
		var email *string
		if n.Email != "" {
			email = &n.Email
		}
		comprador, _ := json.Marshal(Comprador{Nombre: nombre, Telefono: tel, Email: email})
		if n.CostoEnvio < 0 || n.Descuento < 0 || n.Descuento > total {
			return errors.New("pedido inválido")
		}
		var codigo *string
		if n.CodigoPublico != "" {
			codigo = &n.CodigoPublico
		}
		js, _ := json.Marshal(items)
		if _, err := t.Exec(ctx, `SELECT 1 FROM tiendas WHERE id = $1 FOR UPDATE`, tiendaID); err != nil {
			return err
		}
		var cupon *string
		if n.Cupon != "" {
			cupon = &n.Cupon
			tag, err := t.Exec(ctx, `UPDATE cupones SET usos = usos + 1, vendio = vendio + $3, desconto = desconto + $4
				WHERE tienda_id = $1 AND codigo = $2 AND NOT pausado AND (limite = 0 OR usos < limite)`,
				tiendaID, n.Cupon, total+n.CostoEnvio-n.Descuento, n.DescuentoCupon)
			if err != nil {
				return err
			}
			if tag.RowsAffected() == 0 {
				return ErrCupon
			}
		}
		return t.QueryRow(ctx, `
			INSERT INTO pedidos (tienda_id, numero, comprador, canal, pago, estado_pago, envio, estado, facturado,
				items, subtotal, costo_envio, total, nota, estados_en, ciudad_envio, codigo_publico, cupon)
			SELECT $1, coalesce(max(numero), 1000) + 1, $2, $3, $4, $5, $6, $7, false, $8, $9, $11::bigint, $9::bigint + $11::bigint - $12::bigint, $10,
				jsonb_build_object($7::text, now()), $13, $14, $15
			FROM pedidos WHERE tienda_id = $1
			RETURNING numero`,
			tiendaID, comprador, n.Canal, n.Pago, estadoPago, n.Envio, estado, js, total, n.Nota,
			n.CostoEnvio, n.Descuento, n.CiudadEnvio, codigo, cupon).Scan(&numero)
	})
	return numero, total + n.CostoEnvio - n.Descuento, err
}

func rutasNuevo(mux *http.ServeMux, pool tx, con func(func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc) {
	mux.HandleFunc("POST /v1/panel/pedidos", con(func(w http.ResponseWriter, r *http.Request, id string) {
		var n PedidoNuevo
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&n); err != nil {
			problema.Escribir(w, problema.Invalido("Mandá los productos del pedido."))
			return
		}
		numero, total, err := CrearDesdeProductos(r.Context(), pool, id, n)
		switch {
		case errors.Is(err, ErrSinStock):
			problema.Escribir(w, problema.Conflicto("sin_stock", "No alcanza el stock de algún producto."))
		case errors.Is(err, ErrProducto):
			problema.Escribir(w, problema.Invalido("Hay un producto que no existe."))
		case err != nil && strings.Contains(err.Error(), "inválid"):
			problema.Escribir(w, problema.Invalido("Revisá el pedido."))
		case err != nil:
			problema.Escribir(w, problema.Interno())
		default:
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(map[string]any{"numero": numero, "total": total})
		}
	}))
}
