-- +goose Up
-- Turnos con seña (plantilla Tijera y cualquier rubro que lo prenda): la tienda publica servicios y
-- profesionales, quien compra reserva un horario y la tienda lo maneja en la agenda del panel.

-- Ajustes de turnos: una fila por tienda (sin fila, valen los de por defecto y está apagado).
CREATE TABLE turnos_ajustes (
	tienda_id uuid PRIMARY KEY REFERENCES tiendas (id) ON DELETE CASCADE,
	activo boolean NOT NULL DEFAULT false,
	-- Porcentaje del precio que se cobra de seña (0 = sin seña).
	sena int NOT NULL DEFAULT 30 CHECK (sena BETWEEN 0 AND 100),
	anticipacion_min int NOT NULL DEFAULT 60 CHECK (anticipacion_min BETWEEN 0 AND 10080),
	dias_vista int NOT NULL DEFAULT 14 CHECK (dias_vista BETWEEN 1 AND 90),
	-- [{dia: 0..6, desde: 'HH:MM', hasta: 'HH:MM'}], hora de Argentina.
	horario jsonb NOT NULL DEFAULT '[]' CHECK (jsonb_typeof(horario) = 'array' AND pg_column_size(horario) <= 8192),
	actualizado_en timestamptz NOT NULL DEFAULT now()
);
SELECT app.aplicar_rls('turnos_ajustes');
GRANT SELECT, INSERT ON turnos_ajustes TO mostrador_app;
GRANT UPDATE (activo, sena, anticipacion_min, dias_vista, horario, actualizado_en) ON turnos_ajustes TO mostrador_app;

CREATE TABLE turnos_servicios (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	nombre text NOT NULL CHECK (length(nombre) BETWEEN 2 AND 60),
	descripcion text NOT NULL DEFAULT '' CHECK (length(descripcion) <= 200),
	precio bigint NOT NULL CHECK (precio > 0),
	duracion int NOT NULL CHECK (duracion BETWEEN 15 AND 480 AND duracion % 15 = 0),
	foto text CHECK (length(foto) <= 2048),
	etiqueta text CHECK (length(etiqueta) <= 30),
	activo boolean NOT NULL DEFAULT true,
	orden int NOT NULL DEFAULT 0,
	creado_en timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX turnos_servicios_tienda ON turnos_servicios (tienda_id, orden);
SELECT app.aplicar_rls('turnos_servicios');
GRANT SELECT, INSERT, DELETE ON turnos_servicios TO mostrador_app;
GRANT UPDATE (nombre, descripcion, precio, duracion, foto, etiqueta, activo, orden) ON turnos_servicios TO mostrador_app;

CREATE TABLE turnos_profesionales (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	nombre text NOT NULL CHECK (length(nombre) BETWEEN 2 AND 40),
	rol text NOT NULL DEFAULT '' CHECK (length(rol) <= 60),
	foto text CHECK (length(foto) <= 2048),
	-- Los servicios que hace (vacío = todos).
	servicios uuid[] NOT NULL DEFAULT '{}' CHECK (cardinality(servicios) <= 100),
	activo boolean NOT NULL DEFAULT true,
	orden int NOT NULL DEFAULT 0,
	creado_en timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX turnos_profesionales_tienda ON turnos_profesionales (tienda_id, orden);
SELECT app.aplicar_rls('turnos_profesionales');
GRANT SELECT, INSERT, DELETE ON turnos_profesionales TO mostrador_app;
-- UPDATE también hace falta para el SELECT ... FOR UPDATE con que se reserva (ver turnos.Reservar).
GRANT UPDATE (nombre, rol, foto, servicios, activo, orden) ON turnos_profesionales TO mostrador_app;

-- Cada turno copia el nombre del servicio y del profesional: borrarlos no deja la agenda en blanco.
-- Por eso servicio_id y profesional_id no son FK.
CREATE TABLE turnos (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	-- Para la página pública del turno: nadie adivina el de otro.
	codigo_publico text NOT NULL UNIQUE CHECK (codigo_publico ~ '^[A-Za-z0-9_-]{16,64}$'),
	servicio_id uuid NOT NULL,
	servicio text NOT NULL,
	profesional_id uuid NOT NULL,
	profesional text NOT NULL,
	inicio timestamptz NOT NULL,
	fin timestamptz NOT NULL,
	cliente_nombre text NOT NULL CHECK (length(cliente_nombre) BETWEEN 1 AND 80),
	cliente_telefono text NOT NULL DEFAULT '' CHECK (length(cliente_telefono) <= 30),
	cliente_email text NOT NULL DEFAULT '' CHECK (length(cliente_email) <= 120),
	precio bigint NOT NULL CHECK (precio >= 0),
	sena bigint NOT NULL DEFAULT 0 CHECK (sena >= 0 AND sena <= precio),
	estado text NOT NULL CHECK (estado IN ('reservado', 'confirmado', 'cancelado', 'atendido', 'ausente')),
	sena_pagada boolean NOT NULL DEFAULT false,
	sena_devuelta boolean NOT NULL DEFAULT false,
	origen text NOT NULL CHECK (origen IN ('tienda', 'panel')),
	nota text CHECK (length(nota) <= 500),
	creado_en timestamptz NOT NULL DEFAULT now(),
	-- Sólo mientras espera la seña: pasado, el horario se libera solo (no hace falta un cron).
	vence_en timestamptz,
	CHECK (fin > inicio)
);
-- Superposiciones por profesional y rango (inicio < $fin AND fin > $inicio), sin los cancelados.
CREATE INDEX turnos_profesional ON turnos (tienda_id, profesional_id, inicio) WHERE estado <> 'cancelado';
CREATE INDEX turnos_agenda ON turnos (tienda_id, inicio);
SELECT app.aplicar_rls('turnos');
GRANT SELECT, INSERT ON turnos TO mostrador_app;
GRANT UPDATE (estado, sena_pagada, sena_devuelta, vence_en, nota) ON turnos TO mostrador_app;

-- La seña de un turno se cobra con Mercado Pago como un pedido.
ALTER TABLE mp_cobros DROP CONSTRAINT mp_cobros_tipo_check;
ALTER TABLE mp_cobros ADD CONSTRAINT mp_cobros_tipo_check CHECK (tipo IN ('link', 'pedido', 'qr', 'turno'));

-- +goose Down
DELETE FROM mp_cobros WHERE tipo = 'turno';
ALTER TABLE mp_cobros DROP CONSTRAINT mp_cobros_tipo_check;
ALTER TABLE mp_cobros ADD CONSTRAINT mp_cobros_tipo_check CHECK (tipo IN ('link', 'pedido', 'qr'));
DROP TABLE turnos;
DROP TABLE turnos_profesionales;
DROP TABLE turnos_servicios;
DROP TABLE turnos_ajustes;
