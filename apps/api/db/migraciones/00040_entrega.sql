-- +goose Up
-- Adónde va el pedido (dirección y código postal), para la etiqueta de envío. NULL: retira o no se cargó.
ALTER TABLE pedidos ADD COLUMN entrega jsonb CHECK (entrega IS NULL OR (jsonb_typeof(entrega) = 'object' AND pg_column_size(entrega) <= 1024));

-- +goose Down
ALTER TABLE pedidos DROP COLUMN entrega;
