-- +goose Up
-- Mi tienda: el diseño que se edita (borrador) y el que ve la gente (publicado).
ALTER TABLE tiendas
	ADD COLUMN diseno_borrador jsonb,
	ADD COLUMN diseno_publicado jsonb;
GRANT UPDATE (diseno_borrador, diseno_publicado) ON tiendas TO mostrador_app;

-- +goose Down
ALTER TABLE tiendas DROP COLUMN diseno_borrador, DROP COLUMN diseno_publicado;
