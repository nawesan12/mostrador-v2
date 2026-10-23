-- +goose Up
-- De dónde vino la tienda (/empezar: utm, partner o tienda que la recomendó, cupón). Se fija una
-- sola vez, al crear la cuenta, y es la base de las comisiones de partners.
ALTER TABLE tiendas ADD COLUMN origen jsonb CHECK (origen IS NULL OR (jsonb_typeof(origen) = 'object' AND pg_column_size(origen) <= 1024));
GRANT UPDATE (origen) ON tiendas TO mostrador_app;
CREATE INDEX tiendas_ref ON tiendas ((origen->>'ref')) WHERE origen ? 'ref';

-- +goose Down
DROP INDEX tiendas_ref;
ALTER TABLE tiendas DROP COLUMN origen;
