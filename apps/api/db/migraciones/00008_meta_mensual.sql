-- +goose Up
-- La meta de ventas del mes que pone la tienda en el Resumen (en pesos; 0 = sin meta).
ALTER TABLE tiendas ADD COLUMN meta_mensual bigint NOT NULL DEFAULT 0 CHECK (meta_mensual BETWEEN 0 AND 100000000000);
GRANT UPDATE (meta_mensual) ON tiendas TO mostrador_app;

-- +goose Down
ALTER TABLE tiendas DROP COLUMN meta_mensual;
