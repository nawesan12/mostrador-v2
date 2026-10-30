-- +goose Up
-- El sitemap: las tiendas abiertas y sus productos publicados, de todas las tiendas, para que los
-- buscadores las encuentren. Sólo lo que ya es público (slug, dominio propio verificado, fechas):
-- se lee con una marca de sesión que prende sólo esta función, como en carritos_para_recordar.
CREATE POLICY sitemap ON tiendas FOR SELECT TO mostrador_migrador
	USING (current_setting('app.sitemap', true) = 'si');
CREATE POLICY sitemap ON productos FOR SELECT TO mostrador_migrador
	USING (current_setting('app.sitemap', true) = 'si');
CREATE POLICY sitemap ON dominios FOR SELECT TO mostrador_migrador
	USING (current_setting('app.sitemap', true) = 'si');
-- +goose StatementBegin
CREATE FUNCTION app.sitemap(p_limite int) RETURNS TABLE (tienda text, dominio text, producto text, actualizado timestamptz)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
BEGIN
	PERFORM set_config('app.sitemap', 'si', true);
	RETURN QUERY
		SELECT t.slug, d.nombre, NULL::text, t.creada_en
		FROM tiendas t LEFT JOIN dominios d ON d.tienda_id = t.id AND d.principal AND d.verificado_en IS NOT NULL
		WHERE t.cerrada_en IS NULL
		UNION ALL
		SELECT t.slug, d.nombre, p.slug, p.creado_en
		FROM productos p JOIN tiendas t ON t.id = p.tienda_id
			LEFT JOIN dominios d ON d.tienda_id = t.id AND d.principal AND d.verificado_en IS NOT NULL
		WHERE t.cerrada_en IS NULL AND p.publicado AND p.slug IS NOT NULL
		LIMIT p_limite;
	PERFORM set_config('app.sitemap', '', true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.sitemap(int) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.sitemap(int) TO mostrador_app;

-- +goose Down
DROP FUNCTION app.sitemap(int);
DROP POLICY sitemap ON dominios;
DROP POLICY sitemap ON productos;
DROP POLICY sitemap ON tiendas;
