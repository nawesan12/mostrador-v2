-- +goose Up
-- Dominios propios de cada tienda (un dominio es de una sola tienda) y el cierre de la cuenta.
CREATE TABLE dominios (
	nombre text PRIMARY KEY CHECK (nombre ~ '^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z]{2,}$'),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	principal boolean NOT NULL DEFAULT false,
	creado_en timestamptz NOT NULL DEFAULT now(),
	verificado_en timestamptz
);
CREATE INDEX dominios_tienda ON dominios (tienda_id);
SELECT app.aplicar_rls('dominios');
GRANT SELECT, INSERT, DELETE ON dominios TO mostrador_app;
GRANT UPDATE (principal, verificado_en) ON dominios TO mostrador_app;

ALTER TABLE tiendas
	ADD COLUMN cerrada_en timestamptz,
	ADD COLUMN borrado_inactivos timestamptz;
GRANT UPDATE (cerrada_en, borrado_inactivos) ON tiendas TO mostrador_app;

-- +goose Down
ALTER TABLE tiendas DROP COLUMN cerrada_en, DROP COLUMN borrado_inactivos;
DROP TABLE dominios;
