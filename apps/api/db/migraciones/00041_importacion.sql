-- +goose Up
-- Migración desde otras plataformas (Tiendanube, Shopify, WooCommerce, Mercado Libre, Empretienda o
-- un CSV): la ficha propia del cliente, los pedidos viejos, las direcciones viejas de la tienda y el
-- trabajo de importación que avanza por el outbox.

-- La ficha del cliente importado o cargado en el panel. Los clientes que salen de los pedidos siguen
-- derivándose de los pedidos; la lista une los dos por email en minúsculas.
CREATE TABLE clientes (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	nombre text NOT NULL CHECK (length(nombre) BETWEEN 1 AND 120),
	-- Se guarda en minúsculas; el índice igual compara con lower() por si alguien escribe a mano.
	email text CHECK (email IS NULL OR length(email) BETWEEN 3 AND 120),
	telefono text CHECK (telefono IS NULL OR length(telefono) <= 30),
	ciudad text NOT NULL DEFAULT '' CHECK (length(ciudad) <= 80),
	provincia text NOT NULL DEFAULT '' CHECK (length(provincia) <= 60),
	notas text NOT NULL DEFAULT '' CHECK (length(notas) <= 2000),
	origen text NOT NULL CHECK (origen = 'panel'
		OR origen ~ '^importado:(tiendanube|shopify|woocommerce|mercadolibre|empretienda|csv)$'),
	-- El id en la plataforma vieja (para no repetir los que no tienen email).
	externo text CHECK (externo IS NULL OR length(externo) BETWEEN 1 AND 80),
	creado_en timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX clientes_email ON clientes (tienda_id, lower(email)) WHERE email IS NOT NULL;
CREATE INDEX clientes_externo ON clientes (tienda_id, origen, externo) WHERE externo IS NOT NULL;
SELECT app.aplicar_rls('clientes');
-- Un upsert sólo completa lo que falta: nunca pisa el nombre ni el email.
GRANT SELECT, INSERT, UPDATE (telefono, ciudad, provincia, notas) ON clientes TO mostrador_app;

-- Pedidos viejos importados. No son pedidos de Mostrador: no entran en Pedidos ni en el Resumen, no
-- tocan stock, facturas ni numeración. Cuentan para la ficha y el segmento del cliente.
CREATE TABLE historial_pedidos (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	plataforma text NOT NULL CHECK (plataforma IN ('tiendanube', 'shopify', 'woocommerce', 'mercadolibre', 'empretienda', 'csv')),
	externo text NOT NULL CHECK (length(externo) BETWEEN 1 AND 80),
	numero_externo text NOT NULL DEFAULT '' CHECK (length(numero_externo) <= 40),
	fecha timestamptz NOT NULL,
	cliente_email text CHECK (cliente_email IS NULL OR length(cliente_email) BETWEEN 3 AND 120),
	cliente_nombre text NOT NULL DEFAULT '' CHECK (length(cliente_nombre) <= 120),
	-- En pesos enteros, como pedidos.total.
	total bigint NOT NULL CHECK (total >= 0),
	-- [{nombre, variante, cantidad, precio}]
	items jsonb NOT NULL DEFAULT '[]' CHECK (jsonb_typeof(items) = 'array' AND jsonb_array_length(items) <= 100),
	-- Tal como venía (pagado, enviado, cancelled…): es historia, no un estado de Mostrador.
	estado text NOT NULL DEFAULT '' CHECK (length(estado) <= 40),
	UNIQUE (tienda_id, plataforma, externo)
);
CREATE INDEX historial_pedidos_email ON historial_pedidos (tienda_id, lower(cliente_email));
SELECT app.aplicar_rls('historial_pedidos');
GRANT SELECT, INSERT ON historial_pedidos TO mostrador_app;

-- Las direcciones de la tienda vieja: si alguien entra por un link viejo (Google, Instagram, un
-- mail) y ese path no existe en Mostrador, la tienda responde 301 a `hacia`. Los dos son paths:
-- `desde` normalizado (vidriera.NormalizarRuta) y `hacia` relativo a la tienda (/producto/<id>).
CREATE TABLE redirecciones (
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	desde text NOT NULL CHECK (desde ~ '^/.' AND length(desde) <= 500),
	hacia text NOT NULL CHECK (hacia ~ '^/([^/\\]|$)' AND length(hacia) <= 500),
	origen text NOT NULL CHECK (origen = 'panel'
		OR origen ~ '^importado:(tiendanube|shopify|woocommerce|mercadolibre|empretienda|csv)$'),
	creada timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (tienda_id, desde)
);
SELECT app.aplicar_rls('redirecciones');
GRANT SELECT, INSERT, UPDATE (hacia, origen), DELETE ON redirecciones TO mostrador_app;

-- El trabajo de importación. Cada paso lo hace el worker (evento `importacion.paso` del outbox) y
-- encola el siguiente en la misma transacción; `pasos` cuenta los hechos, así un evento entregado
-- dos veces no guarda dos veces la misma página.
CREATE TABLE importaciones (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	plataforma text NOT NULL CHECK (plataforma IN ('tiendanube', 'shopify', 'woocommerce', 'mercadolibre', 'empretienda', 'csv')),
	url_origen text NOT NULL DEFAULT '' CHECK (length(url_origen) <= 300),
	estado text NOT NULL DEFAULT 'pendiente' CHECK (estado IN ('pendiente', 'corriendo', 'lista', 'fallida', 'cancelada')),
	etapa text NOT NULL DEFAULT 'productos' CHECK (etapa IN ('productos', 'clientes', 'pedidos', 'redirecciones')),
	-- Por dónde va la fuente en la etapa (lo que la fuente entienda); NULL: desde el principio.
	cursor jsonb CHECK (cursor IS NULL OR pg_column_size(cursor) <= 2048),
	pasos int NOT NULL DEFAULT 0,
	progreso jsonb NOT NULL DEFAULT '{"productos": 0, "clientes": 0, "pedidos": 0, "redirecciones": 0, "repetidos": 0, "recortados": 0}'
		CHECK (jsonb_typeof(progreso) = 'object'),
	-- Los últimos 50: [{etapa, externo, detalle}]
	errores jsonb NOT NULL DEFAULT '[]' CHECK (jsonb_typeof(errores) = 'array' AND jsonb_array_length(errores) <= 50),
	-- El token o la clave de la plataforma vieja, cifrado con la clave de la tienda (plataforma/cifrado,
	-- como los de Mercado Pago). Se borra al terminar: no hay por qué guardarlo después.
	credencial bytea,
	creada timestamptz NOT NULL DEFAULT now(),
	terminada timestamptz,
	CHECK ((estado IN ('lista', 'fallida', 'cancelada')) = (terminada IS NOT NULL))
);
-- Una sola activa por tienda.
CREATE UNIQUE INDEX importaciones_activa ON importaciones (tienda_id) WHERE estado IN ('pendiente', 'corriendo');
CREATE INDEX importaciones_recientes ON importaciones (tienda_id, creada DESC);
SELECT app.aplicar_rls('importaciones');
GRANT SELECT, INSERT, UPDATE (estado, etapa, cursor, pasos, progreso, errores, credencial, terminada) ON importaciones TO mostrador_app;

-- +goose Down
DROP TABLE importaciones;
DROP TABLE redirecciones;
DROP TABLE historial_pedidos;
DROP TABLE clientes;
