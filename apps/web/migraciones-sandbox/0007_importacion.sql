-- Migración desde otras plataformas, en la demo: las mismas tablas que Postgres (00041_importacion),
-- con el sandbox en lugar de la tienda. Todo se va con el sandbox (ON DELETE CASCADE) cuando la purga
-- del cron lo borra a las 24 h. Los índices empiezan por sandbox_id, como pide D1.

-- Ficha del cliente importado o cargado. El email se guarda en minúsculas: el único compara directo.
CREATE TABLE clientes (
	sandbox_id TEXT NOT NULL REFERENCES sandboxes (id) ON DELETE CASCADE,
	id TEXT NOT NULL,
	nombre TEXT NOT NULL,
	email TEXT,
	telefono TEXT,
	ciudad TEXT NOT NULL DEFAULT '',
	provincia TEXT NOT NULL DEFAULT '',
	notas TEXT NOT NULL DEFAULT '',
	origen TEXT NOT NULL,
	externo TEXT,
	creado_en INTEGER NOT NULL,
	PRIMARY KEY (sandbox_id, id)
);
CREATE UNIQUE INDEX clientes_email ON clientes (sandbox_id, email) WHERE email IS NOT NULL;
CREATE INDEX clientes_externo ON clientes (sandbox_id, origen, externo) WHERE externo IS NOT NULL;

-- Pedidos viejos: no entran en Pedidos ni en el Resumen; cuentan para la ficha y el segmento.
CREATE TABLE historial_pedidos (
	sandbox_id TEXT NOT NULL REFERENCES sandboxes (id) ON DELETE CASCADE,
	id TEXT NOT NULL,
	plataforma TEXT NOT NULL,
	externo TEXT NOT NULL,
	numero_externo TEXT NOT NULL DEFAULT '',
	fecha INTEGER NOT NULL,
	cliente_email TEXT,
	cliente_nombre TEXT NOT NULL DEFAULT '',
	total INTEGER NOT NULL,
	-- JSON: [{nombre, variante, cantidad, precio}]
	items TEXT NOT NULL DEFAULT '[]',
	estado TEXT NOT NULL DEFAULT '',
	PRIMARY KEY (sandbox_id, id),
	UNIQUE (sandbox_id, plataforma, externo)
);
CREATE INDEX historial_pedidos_email ON historial_pedidos (sandbox_id, cliente_email);

-- Links de la tienda vieja → path en Mostrador (relativo a la tienda).
CREATE TABLE redirecciones (
	sandbox_id TEXT NOT NULL REFERENCES sandboxes (id) ON DELETE CASCADE,
	desde TEXT NOT NULL,
	hacia TEXT NOT NULL,
	origen TEXT NOT NULL,
	creada INTEGER NOT NULL,
	PRIMARY KEY (sandbox_id, desde)
);

-- El trabajo de importación. En la demo no hay outbox: avanza un paso por cada pedido del navegador
-- y `pasos` hace de candado (un paso sólo se toma si nadie lo tomó). La credencial no se guarda.
CREATE TABLE importaciones (
	sandbox_id TEXT NOT NULL REFERENCES sandboxes (id) ON DELETE CASCADE,
	id TEXT NOT NULL,
	plataforma TEXT NOT NULL,
	url_origen TEXT NOT NULL DEFAULT '',
	estado TEXT NOT NULL DEFAULT 'pendiente',
	etapa TEXT NOT NULL DEFAULT 'productos',
	cursor TEXT,
	pasos INTEGER NOT NULL DEFAULT 0,
	progreso TEXT NOT NULL,
	errores TEXT NOT NULL DEFAULT '[]',
	creada INTEGER NOT NULL,
	terminada INTEGER,
	PRIMARY KEY (sandbox_id, id)
);
CREATE UNIQUE INDEX importaciones_activa ON importaciones (sandbox_id) WHERE estado IN ('pendiente', 'corriendo');
CREATE INDEX importaciones_recientes ON importaciones (sandbox_id, creada);
