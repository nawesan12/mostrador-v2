-- +goose Up
-- Cupones en la tienda online: el pedido guarda el código que usó y el cupón suma usos, ventas y
-- descuento (en la misma transacción que crea el pedido, para que el límite no se pase).
ALTER TABLE pedidos ADD COLUMN cupon text CHECK (cupon ~ '^[A-Z0-9-]{3,20}$');
CREATE INDEX pedidos_cupon ON pedidos (tienda_id, cupon) WHERE cupon IS NOT NULL;
GRANT UPDATE (usos, vendio, desconto) ON cupones TO mostrador_app;

-- +goose Down
REVOKE UPDATE (usos, vendio, desconto) ON cupones FROM mostrador_app;
DROP INDEX pedidos_cupon;
ALTER TABLE pedidos DROP COLUMN cupon;
