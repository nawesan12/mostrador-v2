-- +goose Up
-- Fundación: la tienda activa vive en una variable de la transacción, y RLS la usa en cada tabla.
-- Los roles (mostrador_app, mostrador_worker, mostrador_staff, mostrador_migrador) los crea
-- `mostradorctl bootstrap`: son del clúster y una migración no debería tener permiso para eso.

CREATE SCHEMA IF NOT EXISTS app;
GRANT USAGE ON SCHEMA app TO mostrador_app, mostrador_worker, mostrador_staff;

-- +goose StatementBegin
-- La tienda de la transacción en curso. NULL si no se fijó: entonces RLS no deja ver nada.
CREATE FUNCTION app.tienda_actual() RETURNS uuid
LANGUAGE sql STABLE AS $$
	SELECT nullif(current_setting('app.tienda_id', true), '')::uuid
$$;
-- +goose StatementEnd

-- +goose StatementBegin
-- Aplica el aislamiento por tienda a una tabla de negocio: RLS habilitada y FORZADA (alcanza
-- también al dueño), la política que compara tienda_id y la lectura de staff. Toda tabla con
-- tienda_id pasa por acá; el test estructural rompe el CI si alguna se olvida.
CREATE FUNCTION app.aplicar_rls(tabla regclass) RETURNS void
LANGUAGE plpgsql AS $$
BEGIN
	EXECUTE format('ALTER TABLE %s ENABLE ROW LEVEL SECURITY', tabla);
	EXECUTE format('ALTER TABLE %s FORCE ROW LEVEL SECURITY', tabla);
	EXECUTE format(
		'CREATE POLICY tienda ON %s TO mostrador_app, mostrador_migrador USING (tienda_id = app.tienda_actual()) WITH CHECK (tienda_id = app.tienda_actual())',
		tabla);
	EXECUTE format('CREATE POLICY staff_lee ON %s FOR SELECT TO mostrador_staff USING (true)', tabla);
	EXECUTE format('GRANT SELECT ON %s TO mostrador_staff', tabla);
END
$$;
-- +goose StatementEnd

-- Tiendas. La política compara el id (no hay tienda_id): una tienda sólo se ve a sí misma.
CREATE TABLE tiendas (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	slug text NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9]([a-z0-9-]{0,46}[a-z0-9])?$'),
	nombre text NOT NULL CHECK (length(nombre) BETWEEN 1 AND 80),
	plan text NOT NULL DEFAULT 'gratis' CHECK (plan IN ('gratis', 'emprende', 'pro', 'negocio')),
	creada_en timestamptz NOT NULL DEFAULT now()
);
ALTER TABLE tiendas ENABLE ROW LEVEL SECURITY;
ALTER TABLE tiendas FORCE ROW LEVEL SECURITY;
CREATE POLICY tienda ON tiendas TO mostrador_app, mostrador_migrador
	USING (id = app.tienda_actual()) WITH CHECK (id = app.tienda_actual());
CREATE POLICY staff_lee ON tiendas FOR SELECT TO mostrador_staff USING (true);
GRANT SELECT, UPDATE (nombre, plan) ON tiendas TO mostrador_app;
GRANT SELECT ON tiendas TO mostrador_staff;

-- +goose StatementBegin
-- Alta de tienda: corre como dueño de la función para poder insertar antes de que exista una
-- tienda activa, y deja fijada la nueva en la transacción.
CREATE FUNCTION app.crear_tienda(p_slug text, p_nombre text) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	nueva uuid := gen_random_uuid();
BEGIN
	PERFORM set_config('app.tienda_id', nueva::text, true);
	INSERT INTO tiendas (id, slug, nombre) VALUES (nueva, p_slug, p_nombre);
	RETURN nueva;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.crear_tienda(text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.crear_tienda(text, text) TO mostrador_app;

-- Outbox: el evento se escribe en la misma transacción que el cambio que lo causa. La app sólo
-- puede insertar los de su tienda; el worker los toma de todas y los marca.
CREATE TABLE outbox (
	id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	tipo text NOT NULL CHECK (tipo ~ '^[a-z_]+\.[a-z_]+$'),
	datos jsonb NOT NULL DEFAULT '{}',
	creado_en timestamptz NOT NULL DEFAULT now(),
	publicado_en timestamptz,
	intentos int NOT NULL DEFAULT 0,
	proximo_intento timestamptz NOT NULL DEFAULT now(),
	ultimo_error text
);
-- El worker sólo recorre lo pendiente: índice parcial, chico aunque la tabla crezca.
CREATE INDEX outbox_pendientes ON outbox (proximo_intento, id) WHERE publicado_en IS NULL;
CREATE INDEX outbox_tienda ON outbox (tienda_id, id);
SELECT app.aplicar_rls('outbox');
REVOKE UPDATE, DELETE ON outbox FROM mostrador_app;
GRANT INSERT, SELECT ON outbox TO mostrador_app;
CREATE POLICY worker ON outbox TO mostrador_worker USING (true) WITH CHECK (true);
GRANT SELECT, UPDATE (publicado_en, intentos, proximo_intento, ultimo_error) ON outbox TO mostrador_worker;

-- Idempotencia: una escritura con la misma Idempotency-Key devuelve la respuesta guardada.
CREATE TABLE idempotencia (
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	clave text NOT NULL CHECK (length(clave) BETWEEN 8 AND 255),
	metodo text NOT NULL,
	ruta text NOT NULL,
	huella text NOT NULL,
	estado int,
	respuesta bytea,
	tipo_contenido text,
	creada_en timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (tienda_id, clave)
);
CREATE INDEX idempotencia_vieja ON idempotencia (creada_en);
SELECT app.aplicar_rls('idempotencia');
GRANT SELECT, INSERT, UPDATE (estado, respuesta, tipo_contenido), DELETE ON idempotencia TO mostrador_app;

-- Auditoría: quién hizo qué y desde dónde. Es la base de la pantalla Actividad y del Deshacer.
CREATE TABLE auditoria (
	id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	actor_tipo text NOT NULL CHECK (actor_tipo IN ('persona', 'hornero', 'sistema', 'staff', 'integracion')),
	actor_id text,
	accion text NOT NULL,
	entidad text NOT NULL,
	entidad_id text,
	antes jsonb,
	despues jsonb,
	ip inet,
	agente text,
	creado_en timestamptz NOT NULL DEFAULT now(),
	deshecho_en timestamptz
);
CREATE INDEX auditoria_tienda ON auditoria (tienda_id, creado_en DESC, id DESC);
SELECT app.aplicar_rls('auditoria');
GRANT SELECT, INSERT, UPDATE (deshecho_en) ON auditoria TO mostrador_app;

-- +goose Down
DROP TABLE auditoria, idempotencia, outbox, tiendas;
DROP FUNCTION app.crear_tienda(text, text);
DROP FUNCTION app.aplicar_rls(regclass);
DROP FUNCTION app.tienda_actual();
DROP SCHEMA app;
