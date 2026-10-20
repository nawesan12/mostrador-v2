-- +goose Up
-- Formularios públicos del sitio (sin sesión): consultas de ventas (/contacto), postulaciones al
-- programa de partners (/partners) y suscripciones a las novedades (/novedades). Como los
-- arrepentimientos, son de la plataforma y no de una tienda: viven en `soporte`.

CREATE TABLE soporte.consultas (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	nombre text NOT NULL CHECK (length(nombre) BETWEEN 1 AND 80),
	email citext NOT NULL CHECK (length(email) <= 254),
	whatsapp text CHECK (whatsapp ~ '^\d{8,15}$'),
	web text CHECK (length(web) <= 200),
	facturacion text CHECK (facturacion IN ('empiezo', 'hasta-5', '5-a-20', '20-a-100', 'mas-de-100')),
	origen text CHECK (origen IN ('tiendanube', 'empretienda', 'mercadoshops', 'shopify', 'woocommerce', 'ninguna')),
	mensaje text CHECK (length(mensaje) <= 2000),
	estado text NOT NULL DEFAULT 'nueva' CHECK (estado IN ('nueva', 'contactada', 'cerrada')),
	creada_en timestamptz NOT NULL DEFAULT now(),
	ip inet
);
CREATE INDEX consultas_creada ON soporte.consultas (creada_en DESC);

CREATE TABLE soporte.postulaciones_partner (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	nombre text NOT NULL CHECK (length(nombre) BETWEEN 1 AND 80),
	email citext NOT NULL UNIQUE CHECK (length(email) <= 254),
	agencia text NOT NULL CHECK (length(agencia) BETWEEN 1 AND 120),
	web text CHECK (length(web) <= 200),
	especialidades text[] NOT NULL DEFAULT '{}'
		CHECK (especialidades <@ ARRAY['diseno', 'desarrollo', 'marketing', 'contabilidad', 'fotografia']),
	estado text NOT NULL DEFAULT 'nueva' CHECK (estado IN ('nueva', 'aprobada', 'rechazada')),
	creada_en timestamptz NOT NULL DEFAULT now(),
	ip inet
);

-- Una fila por email; darse de baja deja la fecha (para no volver a escribirle) y volver a sumarse
-- la borra.
CREATE TABLE soporte.suscripciones_novedades (
	email citext PRIMARY KEY CHECK (length(email) <= 254),
	creada_en timestamptz NOT NULL DEFAULT now(),
	baja_en timestamptz,
	token_baja_hash text NOT NULL UNIQUE
);

-- La app sólo registra (y en las suscripciones, actualiza la propia por email/token). Leer y
-- procesar es del backoffice.
GRANT INSERT, SELECT (id) ON soporte.consultas TO mostrador_app;
GRANT INSERT, SELECT (id) ON soporte.postulaciones_partner TO mostrador_app;
GRANT INSERT, SELECT (email, baja_en, token_baja_hash), UPDATE (creada_en, baja_en, token_baja_hash)
	ON soporte.suscripciones_novedades TO mostrador_app;
GRANT SELECT, UPDATE (estado) ON soporte.consultas, soporte.postulaciones_partner TO mostrador_staff;
GRANT SELECT ON soporte.suscripciones_novedades TO mostrador_staff;

-- +goose Down
DROP TABLE soporte.suscripciones_novedades;
DROP TABLE soporte.postulaciones_partner;
DROP TABLE soporte.consultas;
