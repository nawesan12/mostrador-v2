-- +goose Up
-- Cuentas, sesiones y recuperación. No son datos de una tienda (una persona puede tener varias),
-- así que viven en el esquema `auth`, fuera de public: el test de RLS no los cubre a propósito.
-- La app accede con permisos de columna mínimos; staff no los ve.
CREATE SCHEMA auth;
GRANT USAGE ON SCHEMA auth TO mostrador_app;

CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE auth.usuarios (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	email citext NOT NULL UNIQUE CHECK (length(email) <= 254),
	nombre text NOT NULL CHECK (length(nombre) BETWEEN 1 AND 80),
	clave_hash text,
	email_verificado_en timestamptz,
	creado_en timestamptz NOT NULL DEFAULT now()
);
GRANT SELECT, INSERT, UPDATE (nombre, clave_hash, email_verificado_en) ON auth.usuarios TO mostrador_app;

-- Qué tiendas puede ver cada persona y con qué rol (handoff: Dueño, Admin, Ventas, Depósito).
CREATE TABLE auth.membresias (
	usuario_id uuid NOT NULL REFERENCES auth.usuarios (id) ON DELETE CASCADE,
	tienda_id uuid NOT NULL REFERENCES public.tiendas (id) ON DELETE CASCADE,
	rol text NOT NULL CHECK (rol IN ('duenio', 'admin', 'ventas', 'deposito')),
	creada_en timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (usuario_id, tienda_id)
);
CREATE INDEX membresias_tienda ON auth.membresias (tienda_id);
GRANT SELECT, INSERT, UPDATE (rol), DELETE ON auth.membresias TO mostrador_app;

-- Sesiones: se guarda el sha256 del token, nunca el token.
CREATE TABLE auth.sesiones (
	token_hash text PRIMARY KEY,
	usuario_id uuid NOT NULL REFERENCES auth.usuarios (id) ON DELETE CASCADE,
	tienda_id uuid REFERENCES public.tiendas (id) ON DELETE SET NULL,
	creada_en timestamptz NOT NULL DEFAULT now(),
	vence_en timestamptz NOT NULL,
	visto_en timestamptz NOT NULL DEFAULT now(),
	ip inet,
	agente text
);
CREATE INDEX sesiones_usuario ON auth.sesiones (usuario_id);
CREATE INDEX sesiones_vencidas ON auth.sesiones (vence_en);
GRANT SELECT, INSERT, UPDATE (tienda_id, visto_en), DELETE ON auth.sesiones TO mostrador_app;

-- Links para crear una contraseña nueva: un solo uso, vencen a los 30 minutos.
CREATE TABLE auth.restablecimientos (
	token_hash text PRIMARY KEY,
	usuario_id uuid NOT NULL REFERENCES auth.usuarios (id) ON DELETE CASCADE,
	vence_en timestamptz NOT NULL,
	usado_en timestamptz,
	creado_en timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX restablecimientos_usuario ON auth.restablecimientos (usuario_id, creado_en DESC);
GRANT SELECT, INSERT, UPDATE (usado_en) ON auth.restablecimientos TO mostrador_app;

-- Límite de intentos por clave (email, IP…) en una ventana: un upsert atómico, como en Caja.
CREATE TABLE auth.intentos (
	clave text PRIMARY KEY,
	cantidad int NOT NULL,
	desde timestamptz NOT NULL
);
GRANT SELECT, INSERT, UPDATE, DELETE ON auth.intentos TO mostrador_app;

-- Registro de slugs tomados, fuera de public y sin datos de la tienda: con RLS forzada ni el dueño
-- ve las tiendas ajenas, así que "¿está libre?" se responde acá. La unicidad real la sigue
-- garantizando el índice único de tiendas.slug.
CREATE TABLE app.slugs (slug text PRIMARY KEY, creado_en timestamptz NOT NULL DEFAULT now());

-- +goose StatementBegin
CREATE FUNCTION app.slug_disponible(p_slug text) RETURNS boolean
LANGUAGE sql STABLE SECURITY DEFINER SET search_path = public, app AS $$
	SELECT NOT EXISTS (SELECT 1 FROM app.slugs WHERE slug = p_slug)
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.slug_disponible(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.slug_disponible(text) TO mostrador_app;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.crear_tienda(p_slug text, p_nombre text, p_rubro text, p_plantilla text) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	nueva uuid := gen_random_uuid();
BEGIN
	INSERT INTO app.slugs (slug) VALUES (p_slug);
	PERFORM set_config('app.tienda_id', nueva::text, true);
	INSERT INTO tiendas (id, slug, nombre, rubro, plantilla) VALUES (nueva, p_slug, p_nombre, p_rubro, p_plantilla);
	RETURN nueva;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.crear_tienda(p_slug text, p_nombre text, p_rubro text, p_plantilla text) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	nueva uuid := gen_random_uuid();
BEGIN
	PERFORM set_config('app.tienda_id', nueva::text, true);
	INSERT INTO tiendas (id, slug, nombre, rubro, plantilla) VALUES (nueva, p_slug, p_nombre, p_rubro, p_plantilla);
	RETURN nueva;
END
$$;
-- +goose StatementEnd
DROP FUNCTION app.slug_disponible(text);
DROP TABLE app.slugs;
DROP SCHEMA auth CASCADE;
