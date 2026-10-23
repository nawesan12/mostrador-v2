-- +goose Up
-- Mercado Shops cerró: quien viene de ahí, viene de Mercado Libre.
ALTER TABLE soporte.consultas DROP CONSTRAINT consultas_origen_check;
UPDATE soporte.consultas SET origen = 'mercadolibre' WHERE origen = 'mercadoshops';
ALTER TABLE soporte.consultas ADD CONSTRAINT consultas_origen_check
	CHECK (origen IN ('tiendanube', 'empretienda', 'mercadolibre', 'shopify', 'woocommerce', 'ninguna'));

-- +goose Down
ALTER TABLE soporte.consultas DROP CONSTRAINT consultas_origen_check;
UPDATE soporte.consultas SET origen = 'mercadoshops' WHERE origen = 'mercadolibre';
ALTER TABLE soporte.consultas ADD CONSTRAINT consultas_origen_check
	CHECK (origen IN ('tiendanube', 'empretienda', 'mercadoshops', 'shopify', 'woocommerce', 'ninguna'));
