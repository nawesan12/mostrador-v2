-- +goose Up
-- Canales que la tienda pausó (las publicaciones quedan ocultas, no se borra nada).
ALTER TABLE tiendas ADD COLUMN canales_pausados jsonb NOT NULL DEFAULT '{}';
GRANT UPDATE (canales_pausados) ON tiendas TO mostrador_app;

-- +goose Down
ALTER TABLE tiendas DROP COLUMN canales_pausados;
