-- +goose Up
-- Ingreso con Google, Mercado Pago y Mercado Libre (OAuth) y con link por WhatsApp.

-- El WhatsApp de la cuenta (E.164 sin "+": 5491155550000). Lo carga la persona en Mis datos.
ALTER TABLE auth.usuarios ADD COLUMN telefono text UNIQUE CHECK (telefono ~ '^\d{10,15}$');
GRANT UPDATE (telefono) ON auth.usuarios TO mostrador_app;

-- Cuentas de proveedores vinculadas a una persona.
CREATE TABLE auth.identidades (
	proveedor text NOT NULL CHECK (proveedor IN ('google', 'mercadopago', 'mercadolibre')),
	sujeto text NOT NULL CHECK (length(sujeto) BETWEEN 1 AND 200),
	usuario_id uuid NOT NULL REFERENCES auth.usuarios (id) ON DELETE CASCADE,
	email citext,
	creada_en timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (proveedor, sujeto)
);
CREATE INDEX identidades_usuario ON auth.identidades (usuario_id);
GRANT SELECT, INSERT, DELETE ON auth.identidades TO mostrador_app;

-- Un OAuth en curso: el state (hasheado) y el verificador PKCE. Un solo uso, 10 minutos.
CREATE TABLE auth.oauth_estados (
	estado_hash text PRIMARY KEY,
	proveedor text NOT NULL,
	verificador text NOT NULL,
	vence_en timestamptz NOT NULL,
	usado_en timestamptz
);
CREATE INDEX oauth_estados_vence ON auth.oauth_estados (vence_en);
GRANT SELECT, INSERT, UPDATE (usado_en), DELETE ON auth.oauth_estados TO mostrador_app;

-- Links para entrar sin contraseña (por WhatsApp): un solo uso, 15 minutos.
CREATE TABLE auth.enlaces (
	token_hash text PRIMARY KEY,
	usuario_id uuid NOT NULL REFERENCES auth.usuarios (id) ON DELETE CASCADE,
	vence_en timestamptz NOT NULL,
	usado_en timestamptz,
	creado_en timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX enlaces_usuario ON auth.enlaces (usuario_id, creado_en DESC);
GRANT SELECT, INSERT, UPDATE (usado_en) ON auth.enlaces TO mostrador_app;

-- +goose Down
DROP TABLE auth.enlaces;
DROP TABLE auth.oauth_estados;
DROP TABLE auth.identidades;
ALTER TABLE auth.usuarios DROP COLUMN telefono;
