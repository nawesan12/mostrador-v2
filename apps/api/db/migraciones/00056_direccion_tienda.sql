-- +goose Up
-- La dirección de la tienda (tulocalweb.com.ar/t/<slug>) se puede cambiar (decisiones §38). La vieja
-- queda para siempre: /t/<vieja>/… responde 301 a /t/<nueva>/… y ninguna otra tienda puede tomarla
-- (sigue en app.slugs). Como app.slugs, estas tablas viven fuera de public y la app no tiene permisos
-- sobre ellas: sólo las tocan las funciones de abajo, que deciden la tienda con app.tienda_actual().

-- Hasta 60 caracteres (antes 48), siempre empezando y terminando con letra o número.
ALTER TABLE tiendas DROP CONSTRAINT tiendas_slug_check;
ALTER TABLE tiendas ADD CONSTRAINT tiendas_slug_check CHECK (slug ~ '^[a-z0-9]([a-z0-9-]{0,58}[a-z0-9])?$');

-- Las direcciones anteriores de cada tienda. Si la tienda vuelve a una, la fila se borra (vuelve a
-- ser la vigente) pero el slug sigue tomado en app.slugs.
CREATE TABLE app.slugs_anteriores (
	slug text PRIMARY KEY REFERENCES app.slugs (slug),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	cambiado_en timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX slugs_anteriores_tienda ON app.slugs_anteriores (tienda_id);

-- Cada cambio, para el tope por día (5) y para saber qué pasó.
CREATE TABLE app.slugs_cambios (
	id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	desde text NOT NULL,
	hacia text NOT NULL,
	en timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX slugs_cambios_tienda ON app.slugs_cambios (tienda_id, en DESC);

-- +goose StatementBegin
-- Cómo está un slug para la tienda de la transacción: 'actual' (es el suyo), 'propio' (lo usó antes:
-- puede volver), 'usado' (lo tiene o lo tuvo otra tienda) o 'libre'. NULL sin tienda fijada. El
-- formato y los reservados los mira Go antes.
CREATE FUNCTION app.estado_slug(p_slug text) RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	v_tienda uuid := app.tienda_actual();
	v_actual text;
BEGIN
	IF v_tienda IS NULL THEN
		RETURN NULL;
	END IF;
	SELECT slug INTO v_actual FROM tiendas WHERE id = v_tienda;
	IF v_actual = p_slug THEN
		RETURN 'actual';
	END IF;
	IF EXISTS (SELECT 1 FROM app.slugs_anteriores WHERE slug = p_slug AND tienda_id = v_tienda) THEN
		RETURN 'propio';
	END IF;
	IF EXISTS (SELECT 1 FROM app.slugs WHERE slug = p_slug) THEN
		RETURN 'usado';
	END IF;
	RETURN 'libre';
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.estado_slug(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.estado_slug(text) TO mostrador_app;

-- +goose StatementBegin
-- Cambia la dirección de la tienda de la transacción. Devuelve 'ok', 'actual' (ya era esa), 'usado'
-- o 'tope' (5 cambios en las últimas 24 h). La fila de la tienda queda bloqueada: dos cambios a la
-- vez de la misma tienda van en fila; dos tiendas que piden el mismo slug, una sola lo consigue
-- (app.slugs es la que decide).
CREATE FUNCTION app.cambiar_slug(p_nuevo text) RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	v_tienda uuid := app.tienda_actual();
	v_viejo text;
BEGIN
	IF v_tienda IS NULL THEN
		RAISE EXCEPTION 'sin tienda' USING ERRCODE = '42501';
	END IF;
	SELECT slug INTO v_viejo FROM tiendas WHERE id = v_tienda FOR UPDATE;
	IF v_viejo IS NULL THEN
		RAISE EXCEPTION 'no existe la tienda' USING ERRCODE = 'P0002';
	END IF;
	IF v_viejo = p_nuevo THEN
		RETURN 'actual';
	END IF;
	IF (SELECT count(*) FROM app.slugs_cambios WHERE tienda_id = v_tienda AND en > now() - interval '24 hours') >= 5 THEN
		RETURN 'tope';
	END IF;
	IF EXISTS (SELECT 1 FROM app.slugs_anteriores WHERE slug = p_nuevo AND tienda_id = v_tienda) THEN
		DELETE FROM app.slugs_anteriores WHERE slug = p_nuevo;
	ELSE
		INSERT INTO app.slugs (slug) VALUES (p_nuevo) ON CONFLICT DO NOTHING;
		IF NOT FOUND THEN
			RETURN 'usado';
		END IF;
	END IF;
	UPDATE tiendas SET slug = p_nuevo WHERE id = v_tienda;
	-- Las tiendas más viejas pueden no tener su slug en app.slugs.
	INSERT INTO app.slugs (slug) VALUES (v_viejo) ON CONFLICT DO NOTHING;
	INSERT INTO app.slugs_anteriores (slug, tienda_id) VALUES (v_viejo, v_tienda);
	INSERT INTO app.slugs_cambios (tienda_id, desde, hacia) VALUES (v_tienda, v_viejo, p_nuevo);
	RETURN 'ok';
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.cambiar_slug(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.cambiar_slug(text) TO mostrador_app;

-- +goose StatementBegin
-- Para lo público: la dirección vigente de la tienda que antes usaba p_slug, o NULL (no fue de nadie,
-- o la tienda está cerrada o pausada). Como app.slug_de_dominio, el dueño de la función sólo ve la
-- fila de esa tienda.
CREATE FUNCTION app.slug_vigente(p_slug text) RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	v_tienda uuid;
	v_slug text;
BEGIN
	SELECT tienda_id INTO v_tienda FROM app.slugs_anteriores WHERE slug = p_slug;
	IF v_tienda IS NULL THEN
		RETURN NULL;
	END IF;
	PERFORM set_config('app.tienda_id', v_tienda::text, true);
	SELECT slug INTO v_slug FROM tiendas WHERE id = v_tienda AND cerrada_en IS NULL AND pausada_en IS NULL;
	PERFORM set_config('app.tienda_id', '', true);
	RETURN v_slug;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.slug_vigente(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.slug_vigente(text) TO mostrador_app;

-- +goose Down
DROP FUNCTION app.slug_vigente(text);
DROP FUNCTION app.cambiar_slug(text);
DROP FUNCTION app.estado_slug(text);
DROP TABLE app.slugs_cambios;
DROP TABLE app.slugs_anteriores;
ALTER TABLE tiendas DROP CONSTRAINT tiendas_slug_check;
ALTER TABLE tiendas ADD CONSTRAINT tiendas_slug_check CHECK (slug ~ '^[a-z0-9]([a-z0-9-]{0,46}[a-z0-9])?$') NOT VALID;
