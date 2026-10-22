-- +goose Up
-- Cuenta: cambios a la matriz de permisos, packs de Hornero por mes y cancelación del plan.
ALTER TABLE tiendas
	ADD COLUMN permisos jsonb NOT NULL DEFAULT '{}' CHECK (pg_column_size(permisos) <= 4096),
	ADD COLUMN packs jsonb NOT NULL DEFAULT '{}' CHECK (pg_column_size(packs) <= 4096),
	ADD COLUMN cancelada_en timestamptz;
GRANT UPDATE (permisos, packs, cancelada_en) ON tiendas TO mostrador_app;

-- +goose Down
ALTER TABLE tiendas DROP COLUMN permisos, DROP COLUMN packs, DROP COLUMN cancelada_en;
