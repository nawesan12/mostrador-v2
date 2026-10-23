-- +goose Up
-- Variantes y galería. Un producto puede tener hasta 3 opciones (Talle, Color, Presentación…) y
-- entonces se vende por variante (una combinación de valores), cada una con su stock, SKU y,
-- si quiere, su precio y su foto. La galería son hasta 8 fotos; `foto` queda como la primera.
ALTER TABLE productos
	ADD COLUMN opciones jsonb NOT NULL DEFAULT '[]'
		CHECK (jsonb_typeof(opciones) = 'array' AND jsonb_array_length(opciones) <= 3 AND length(opciones::text) <= 4000),
	ADD COLUMN fotos text[] NOT NULL DEFAULT '{}'
		CHECK (cardinality(fotos) <= 8 AND array_position(fotos, NULL) IS NULL);
GRANT UPDATE (opciones, fotos) ON productos TO mostrador_app;

-- La galería arranca con la foto que ya tenían. El migrador es dueño de la tabla pero RLS está
-- forzada (no ve ninguna tienda): se suelta sólo dentro de esta transacción para el relleno.
ALTER TABLE productos NO FORCE ROW LEVEL SECURITY;
UPDATE productos SET fotos = ARRAY[foto] WHERE foto IS NOT NULL;
ALTER TABLE productos FORCE ROW LEVEL SECURITY;

-- Una fila por combinación. `valores` es {"Talle": "M", "Color": "Negro"}: una clave por opción.
-- `precio` NULL = el del producto. El stock vive acá: el del producto es la suma (trigger de abajo).
CREATE TABLE variantes (
	tienda_id uuid NOT NULL,
	producto_id uuid NOT NULL,
	id uuid NOT NULL DEFAULT gen_random_uuid(),
	valores jsonb NOT NULL CHECK (jsonb_typeof(valores) = 'object'),
	sku text NOT NULL DEFAULT '' CHECK (length(sku) <= 40),
	precio bigint CHECK (precio >= 0),
	stock integer NOT NULL DEFAULT 0 CHECK (stock >= 0),
	foto text,
	activa boolean NOT NULL DEFAULT true,
	creada timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (tienda_id, producto_id, id),
	UNIQUE (tienda_id, producto_id, valores),
	FOREIGN KEY (tienda_id, producto_id) REFERENCES productos (tienda_id, id) ON DELETE CASCADE
);
SELECT app.aplicar_rls('variantes');
GRANT SELECT, INSERT, DELETE ON variantes TO mostrador_app;
GRANT UPDATE (sku, precio, stock, foto, activa) ON variantes TO mostrador_app;

-- Con opciones, el stock del producto es la suma de sus variantes activas. Lo mantiene la base para
-- que ningún camino (venta, ajuste, conteo, edición) lo deje desparejo; y como toca productos.stock,
-- el trigger de Mercado Libre se entera solo.
-- +goose StatementBegin
CREATE FUNCTION app.variantes_stock() RETURNS trigger
LANGUAGE plpgsql SET search_path = public, app AS $$
DECLARE
	t uuid;
	p uuid;
BEGIN
	IF TG_OP = 'DELETE' THEN
		t := OLD.tienda_id; p := OLD.producto_id;
	ELSE
		t := NEW.tienda_id; p := NEW.producto_id;
	END IF;
	UPDATE productos SET stock = s.total
	FROM (SELECT coalesce(sum(stock), 0)::int AS total FROM variantes
		WHERE tienda_id = t AND producto_id = p AND activa) s
	WHERE productos.tienda_id = t AND productos.id = p AND productos.opciones <> '[]'::jsonb
		AND productos.stock <> s.total;
	RETURN NULL;
END
$$;
-- +goose StatementEnd
CREATE TRIGGER variantes_stock AFTER INSERT OR DELETE OR UPDATE OF stock, activa ON variantes
	FOR EACH ROW EXECUTE FUNCTION app.variantes_stock();

-- Mercado Libre: qué variación de la publicación es cada variante ({"<variante>": <variation_id>}).
ALTER TABLE ml_publicaciones ADD COLUMN variaciones jsonb NOT NULL DEFAULT '{}';
GRANT UPDATE (variaciones) ON ml_publicaciones TO mostrador_app;

-- +goose Down
ALTER TABLE ml_publicaciones DROP COLUMN variaciones;
DROP TRIGGER variantes_stock ON variantes;
DROP FUNCTION app.variantes_stock();
DROP TABLE variantes;
ALTER TABLE productos DROP COLUMN fotos, DROP COLUMN opciones;
