-- +goose Up
-- Medición de la tienda online con las cuentas de la propia tienda: el ID de GA4, el del píxel de
-- Meta y qué feeds de productos ya cargó (Merchant Center, Meta Commerce Manager). Nada de esto es
-- secreto (los IDs se publican en la tienda), así que va en tiendas, con su misma RLS.
ALTER TABLE tiendas
	ADD COLUMN ga4 text CHECK (ga4 ~ '^G-[A-Z0-9]{6,12}$'),
	ADD COLUMN pixel_meta text CHECK (pixel_meta ~ '^[0-9]{15,16}$'),
	ADD COLUMN feeds_cargados text[] NOT NULL DEFAULT '{}' CHECK (feeds_cargados <@ ARRAY['gm', 'meta']);
GRANT UPDATE (ga4, pixel_meta, feeds_cargados) ON tiendas TO mostrador_app;

-- +goose Down
ALTER TABLE tiendas DROP COLUMN feeds_cargados, DROP COLUMN pixel_meta, DROP COLUMN ga4;
