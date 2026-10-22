-- +goose Up
-- Recordatorios automáticos de carritos abandonados que elige la tienda.
ALTER TABLE tiendas ADD COLUMN carritos_automaticos jsonb NOT NULL DEFAULT '{"hora": true, "dia": true, "email": false}';
GRANT UPDATE (carritos_automaticos) ON tiendas TO mostrador_app;

-- +goose Down
ALTER TABLE tiendas DROP COLUMN carritos_automaticos;
