-- +goose Up
-- La tienda de un dominio propio, para lo público (la web la pide al recibir mitienda.com.ar): sólo
-- si el dominio está verificado (ya apunta a Tu Local Web) y la tienda no está cerrada. Devuelve el
-- slug, que es con lo que la web arma /t/<slug>. Como app.tienda_de_slug: el migrador, dueño de la
-- función, sólo ve la fila del dominio pedido y después la de su tienda.
CREATE POLICY por_dominio ON dominios FOR SELECT TO mostrador_migrador
	USING (nombre = current_setting('app.dominio', true));
-- +goose StatementBegin
CREATE FUNCTION app.slug_de_dominio(p_dominio text) RETURNS text
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	v_tienda uuid;
	v_slug text;
BEGIN
	PERFORM set_config('app.dominio', p_dominio, true);
	SELECT tienda_id INTO v_tienda FROM dominios WHERE nombre = p_dominio AND verificado_en IS NOT NULL;
	PERFORM set_config('app.dominio', '', true);
	IF v_tienda IS NULL THEN
		RETURN NULL;
	END IF;
	PERFORM set_config('app.tienda_id', v_tienda::text, true);
	SELECT slug INTO v_slug FROM tiendas WHERE id = v_tienda AND cerrada_en IS NULL;
	PERFORM set_config('app.tienda_id', '', true);
	RETURN v_slug;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.slug_de_dominio(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.slug_de_dominio(text) TO mostrador_app;

-- +goose Down
DROP FUNCTION app.slug_de_dominio(text);
DROP POLICY por_dominio ON dominios;
