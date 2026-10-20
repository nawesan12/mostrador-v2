-- +goose Up
-- Rubro y plantilla de la tienda (packages/contrato/rubros.json).
ALTER TABLE tiendas ADD COLUMN rubro text NOT NULL DEFAULT 'regionales';
ALTER TABLE tiendas ADD COLUMN plantilla text NOT NULL DEFAULT 'litoral';
ALTER TABLE tiendas ALTER COLUMN rubro DROP DEFAULT;
ALTER TABLE tiendas ALTER COLUMN plantilla DROP DEFAULT;

DROP FUNCTION app.crear_tienda(text, text);
-- +goose StatementBegin
CREATE FUNCTION app.crear_tienda(p_slug text, p_nombre text, p_rubro text, p_plantilla text) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	nueva uuid := gen_random_uuid();
BEGIN
	PERFORM set_config('app.tienda_id', nueva::text, true);
	INSERT INTO tiendas (id, slug, nombre, rubro, plantilla) VALUES (nueva, p_slug, p_nombre, p_rubro, p_plantilla);
	RETURN nueva;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.crear_tienda(text, text, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.crear_tienda(text, text, text, text) TO mostrador_app;
GRANT UPDATE (plantilla) ON tiendas TO mostrador_app;

-- +goose Down
DROP FUNCTION app.crear_tienda(text, text, text, text);
-- +goose StatementBegin
CREATE FUNCTION app.crear_tienda(p_slug text, p_nombre text) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	nueva uuid := gen_random_uuid();
BEGIN
	PERFORM set_config('app.tienda_id', nueva::text, true);
	INSERT INTO tiendas (id, slug, nombre) VALUES (nueva, p_slug, p_nombre);
	RETURN nueva;
END
$$;
-- +goose StatementEnd
GRANT EXECUTE ON FUNCTION app.crear_tienda(text, text) TO mostrador_app;
ALTER TABLE tiendas DROP COLUMN plantilla;
ALTER TABLE tiendas DROP COLUMN rubro;
