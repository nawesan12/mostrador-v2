-- +goose Up
-- Productos de cada tienda. `canales` dice dónde se publica (web, ml, ig, google).
CREATE TABLE productos (
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	id uuid NOT NULL DEFAULT gen_random_uuid(),
	nombre text NOT NULL CHECK (length(nombre) BETWEEN 2 AND 120),
	descripcion text NOT NULL DEFAULT '' CHECK (length(descripcion) <= 4000),
	categoria text CHECK (length(categoria) <= 60),
	precio bigint NOT NULL CHECK (precio >= 0),
	costo bigint NOT NULL DEFAULT 0 CHECK (costo >= 0),
	sku text NOT NULL DEFAULT '' CHECK (length(sku) <= 40),
	stock integer NOT NULL DEFAULT 0 CHECK (stock >= 0),
	foto text,
	canales jsonb NOT NULL DEFAULT '{"web": true}',
	publicado boolean NOT NULL DEFAULT true,
	creado_en timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (tienda_id, id)
);
SELECT app.aplicar_rls('productos');
GRANT SELECT, INSERT ON productos TO mostrador_app;
GRANT UPDATE (nombre, descripcion, categoria, precio, costo, sku, stock, foto, canales, publicado) ON productos TO mostrador_app;

-- +goose Down
DROP TABLE productos;
