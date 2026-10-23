-- +goose Up
-- "Manejá Mostrador desde tu IA": OAuth 2.1 para conectores MCP (Claude, ChatGPT…). Se guarda el
-- sha256 de códigos y tokens, nunca el valor. Una conexión es de una persona en una tienda; el token
-- de acceso entra a la API como una sesión más, con alcance de lectura o de acciones.
CREATE TABLE auth.ia_clientes (
	id text PRIMARY KEY CHECK (id ~ '^[A-Za-z0-9_-]{16,64}$'),
	nombre text NOT NULL CHECK (length(nombre) BETWEEN 1 AND 80),
	redirecciones text[] NOT NULL CHECK (cardinality(redirecciones) BETWEEN 1 AND 5 AND pg_column_size(redirecciones) <= 2600),
	creado_en timestamptz NOT NULL DEFAULT now(),
	usado_en timestamptz
);
GRANT SELECT, INSERT, UPDATE (usado_en), DELETE ON auth.ia_clientes TO mostrador_app;

-- Códigos de autorización: un solo uso (se borran al canjear) y 5 minutos.
CREATE TABLE auth.ia_codigos (
	-- sha256 en base64url: 43 caracteres.
	hash text PRIMARY KEY CHECK (length(hash) = 43),
	cliente_id text NOT NULL REFERENCES auth.ia_clientes (id) ON DELETE CASCADE,
	usuario_id uuid NOT NULL REFERENCES auth.usuarios (id) ON DELETE CASCADE,
	tienda_id uuid NOT NULL REFERENCES public.tiendas (id) ON DELETE CASCADE,
	redireccion text NOT NULL CHECK (length(redireccion) <= 500),
	desafio text NOT NULL CHECK (desafio ~ '^[A-Za-z0-9_-]{43}$'),
	acciones boolean NOT NULL,
	vence_en timestamptz NOT NULL
);
GRANT SELECT, INSERT, DELETE ON auth.ia_codigos TO mostrador_app;

-- Conexiones: token de acceso (1 h) y de refresco (rota en cada uso; si llega uno ya usado, se corta
-- la conexión entera porque alguien lo copió).
CREATE TABLE auth.ia_conexiones (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	cliente_id text NOT NULL REFERENCES auth.ia_clientes (id) ON DELETE CASCADE,
	usuario_id uuid NOT NULL REFERENCES auth.usuarios (id) ON DELETE CASCADE,
	tienda_id uuid NOT NULL REFERENCES public.tiendas (id) ON DELETE CASCADE,
	acciones boolean NOT NULL,
	acceso_hash text NOT NULL UNIQUE CHECK (length(acceso_hash) = 43),
	acceso_vence timestamptz NOT NULL,
	refresco_hash text NOT NULL UNIQUE CHECK (length(refresco_hash) = 43),
	refresco_previo text CHECK (length(refresco_previo) = 43),
	creada_en timestamptz NOT NULL DEFAULT now(),
	usada_en timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX ia_conexiones_tienda ON auth.ia_conexiones (tienda_id);
CREATE INDEX ia_conexiones_previo ON auth.ia_conexiones (refresco_previo);
GRANT SELECT, INSERT, UPDATE (acceso_hash, acceso_vence, refresco_hash, refresco_previo, usada_en), DELETE
	ON auth.ia_conexiones TO mostrador_app;

-- Qué deja hacer la tienda a las IA conectadas (los mismos interruptores que en el panel).
ALTER TABLE tiendas ADD COLUMN ia_permisos jsonb NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(ia_permisos) = 'object' AND pg_column_size(ia_permisos) <= 1024);
GRANT UPDATE (ia_permisos) ON tiendas TO mostrador_app;

-- +goose Down
ALTER TABLE tiendas DROP COLUMN ia_permisos;
DROP TABLE auth.ia_conexiones;
DROP TABLE auth.ia_codigos;
DROP TABLE auth.ia_clientes;
