-- +goose Up
-- El link de cada producto en la tienda: /producto/mate-de-calabaza en vez del uuid. Se arma del
-- nombre al crearlo y no cambia después (un link compartido o indexado sigue andando aunque el
-- producto se renombre). Único por tienda: el segundo "Mate" es mate-2. El uuid sigue sirviendo
-- de link y la web lo redirige (301) al slug.
-- +goose StatementBegin
CREATE FUNCTION app.slug_producto(p_tienda uuid, p_nombre text, p_id uuid) RETURNS text
LANGUAGE plpgsql AS $$
DECLARE
	v_base text;
	v_slug text;
	v_n integer := 1;
BEGIN
	v_base := lower(translate(p_nombre, 'ÁÀÄÂÃÉÈËÊÍÌÏÎÓÒÖÔÕÚÙÜÛÑÇáàäâãéèëêíìïîóòöôõúùüûñç',
		'AAAAAEEEEIIIIOOOOOUUUUNCaaaaaeeeeiiiiooooouuuunc'));
	v_base := regexp_replace(replace(v_base, '&', ' y '), '[^a-z0-9]+', '-', 'g');
	v_base := trim(BOTH '-' FROM left(trim(BOTH '-' FROM v_base), 60));
	IF length(v_base) < 2 THEN
		v_base := 'producto';
	END IF;
	v_slug := v_base;
	WHILE EXISTS (SELECT 1 FROM productos WHERE tienda_id = p_tienda AND slug = v_slug AND id <> p_id) LOOP
		v_n := v_n + 1;
		v_slug := trim(BOTH '-' FROM left(v_base, 60 - length(v_n::text) - 1)) || '-' || v_n;
	END LOOP;
	RETURN v_slug;
END
$$;

CREATE FUNCTION app.productos_slug() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
	IF NEW.slug IS NULL THEN
		NEW.slug := app.slug_producto(NEW.tienda_id, NEW.nombre, NEW.id);
	END IF;
	RETURN NEW;
END
$$;
-- +goose StatementEnd

ALTER TABLE productos ADD COLUMN slug text CHECK (slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$' AND length(slug) <= 60);

-- Los que ya existen, en el orden en que se crearon (el más viejo se queda con el nombre limpio).
-- Como en 00033: RLS forzada no deja ver ninguna tienda al migrador, se suelta sólo para esto.
ALTER TABLE productos NO FORCE ROW LEVEL SECURITY;
-- +goose StatementBegin
DO $$
DECLARE
	r record;
BEGIN
	FOR r IN SELECT tienda_id, id, nombre FROM productos ORDER BY tienda_id, creado_en, id LOOP
		UPDATE productos SET slug = app.slug_producto(r.tienda_id, r.nombre, r.id)
			WHERE tienda_id = r.tienda_id AND id = r.id;
	END LOOP;
END
$$;
-- +goose StatementEnd
ALTER TABLE productos FORCE ROW LEVEL SECURITY;

ALTER TABLE productos ALTER COLUMN slug SET NOT NULL;
CREATE UNIQUE INDEX productos_slug ON productos (tienda_id, slug);
CREATE TRIGGER productos_slug BEFORE INSERT ON productos FOR EACH ROW EXECUTE FUNCTION app.productos_slug();

-- +goose Down
DROP TRIGGER productos_slug ON productos;
DROP FUNCTION app.productos_slug();
ALTER TABLE productos DROP COLUMN slug;
DROP FUNCTION app.slug_producto(uuid, text, uuid);
