-- +goose Up
-- Un segundo pago aprobado de la misma preferencia (dos pestañas, o pagar dos veces) ya no pisa al
-- primero: su id queda acá, el pedido conserva el pago original y la tienda recibe «Te pagaron dos
-- veces». Si después se devuelve el extra, sale de la lista.
ALTER TABLE mp_cobros ADD COLUMN pagos_extra bigint[] NOT NULL DEFAULT '{}';
GRANT UPDATE (pagos_extra) ON mp_cobros TO mostrador_app;

-- +goose Down
ALTER TABLE mp_cobros DROP COLUMN pagos_extra;
