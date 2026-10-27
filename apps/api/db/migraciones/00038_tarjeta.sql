-- +goose Up
-- La tarjeta A6 para los paquetes que arma cada tienda (modelo, textos y QR del dorso). NULL: la de
-- fábrica.
ALTER TABLE tiendas ADD COLUMN tarjeta jsonb CHECK (tarjeta IS NULL OR jsonb_typeof(tarjeta) = 'object');
GRANT UPDATE (tarjeta) ON tiendas TO mostrador_app;

-- +goose Down
ALTER TABLE tiendas DROP COLUMN tarjeta;
