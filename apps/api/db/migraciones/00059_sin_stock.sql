-- +goose Up
-- «Te quedaste sin stock» (docs/decisiones.md §41). Cuando una venta deja en 0 un producto publicado
-- (o una de sus variantes), Go encola el aviso en la misma transacción que descuenta. Uno por
-- producto por día de Argentina como mucho: esta columna guarda el día del último aviso y el UPDATE
-- condicionado (con la fila ya bloqueada por el descuento) decide quién avisa si hay dos ventas a
-- la vez. Sin tabla aparte: no hay nada que purgar.
ALTER TABLE productos ADD COLUMN sin_stock_avisado date;
GRANT UPDATE (sin_stock_avisado) ON productos TO mostrador_app;

-- +goose Down
DELETE FROM outbox WHERE tipo IN ('tienda.sin_stock', 'push.sin_stock');
REVOKE UPDATE (sin_stock_avisado) ON productos FROM mostrador_app;
ALTER TABLE productos DROP COLUMN sin_stock_avisado;
