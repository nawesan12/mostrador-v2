-- +goose Up
-- La tienda online pública: se encuentra por su slug (sin sesión) y cada pedido de la tienda tiene un
-- código público, para la página de confirmación sin que nadie adivine el de otro.
ALTER TABLE pedidos ADD COLUMN codigo_publico text UNIQUE CHECK (codigo_publico ~ '^[A-Za-z0-9_-]{16,64}$');

-- La tienda de un slug, para lo público (como app.tienda_de_link): sólo si no está cerrada.
CREATE POLICY por_slug ON tiendas FOR SELECT TO mostrador_migrador
	USING (slug = current_setting('app.slug', true));
-- +goose StatementBegin
CREATE FUNCTION app.tienda_de_slug(p_slug text) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	t uuid;
BEGIN
	PERFORM set_config('app.slug', p_slug, true);
	SELECT id INTO t FROM tiendas WHERE slug = p_slug AND cerrada_en IS NULL;
	PERFORM set_config('app.slug', '', true);
	RETURN t;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.tienda_de_slug(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.tienda_de_slug(text) TO mostrador_app;

-- +goose Down
DROP FUNCTION app.tienda_de_slug(text);
DROP POLICY por_slug ON tiendas;
ALTER TABLE pedidos DROP COLUMN codigo_publico;
