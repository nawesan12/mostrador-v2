-- +goose Up
-- Pedidos de cada tienda. El comprador queda copiado en el pedido (como se compró), y se vincula a
-- la ficha de cliente cuando exista (cliente_id, dominio Clientes). Los productos van en `items`
-- (nombre, foto, cantidad y precio al momento de la compra).
CREATE TABLE pedidos (
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	numero integer NOT NULL CHECK (numero > 0),
	cliente_id uuid,
	comprador jsonb NOT NULL CHECK (jsonb_typeof(comprador) = 'object' AND comprador ? 'nombre'),
	canal text NOT NULL CHECK (canal IN ('web', 'wa', 'ml', 'ig', 'pos', 'tg')),
	pago text NOT NULL CHECK (pago IN ('mp', 'tarjeta', 'payway', 'modo', 'uala', 'transf', 'efectivo')),
	estado_pago text NOT NULL DEFAULT 'Pendiente',
	envio text NOT NULL CHECK (envio IN ('andreani', 'oca', 'correo', 'moto', 'retiro')),
	ciudad_envio text NOT NULL DEFAULT '',
	estado text NOT NULL DEFAULT 'Sin pagar' CHECK (estado IN ('Sin pagar', 'Para preparar', 'Listo para despachar', 'En camino', 'Entregado', 'Cancelado')),
	-- Cuándo llegó a cada estado ({"Para preparar": "2026-10-22T10:00:00Z", ...}).
	estados_en jsonb NOT NULL DEFAULT '{}',
	facturado boolean NOT NULL DEFAULT false,
	factura text,
	seguimiento text,
	items jsonb NOT NULL CHECK (jsonb_typeof(items) = 'array' AND jsonb_array_length(items) > 0),
	subtotal bigint NOT NULL CHECK (subtotal >= 0),
	costo_envio bigint NOT NULL DEFAULT 0 CHECK (costo_envio >= 0),
	total bigint NOT NULL CHECK (total >= 0),
	vip boolean NOT NULL DEFAULT false,
	hornero boolean NOT NULL DEFAULT false,
	mayorista boolean NOT NULL DEFAULT false,
	nota text,
	recordado_en timestamptz,
	creado_en timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (tienda_id, numero)
);
CREATE INDEX pedidos_recientes ON pedidos (tienda_id, creado_en DESC);
CREATE INDEX pedidos_estado ON pedidos (tienda_id, estado);
SELECT app.aplicar_rls('pedidos');
GRANT SELECT, INSERT ON pedidos TO mostrador_app;
GRANT UPDATE (estado, estados_en, facturado, factura, seguimiento, recordado_en, cliente_id, vip) ON pedidos TO mostrador_app;

-- +goose Down
DROP TABLE pedidos;
