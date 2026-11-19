-- +goose Up
-- El aviso del fin de la prueba gratis: a 3 días y el día que termina, a la dueña (campanita y push),
-- con lo que se va a ocultar al pasar a Gratis. aviso_prueba es el último umbral avisado (3 o 0) y
-- aviso_prueba_fin la fecha de fin que se avisó: si el equipo extiende la prueba, se vuelve a avisar.
ALTER TABLE tiendas ADD COLUMN aviso_prueba smallint CHECK (aviso_prueba IN (0, 3)),
	ADD COLUMN aviso_prueba_fin timestamptz;
GRANT UPDATE (aviso_prueba, aviso_prueba_fin) ON tiendas TO mostrador_app;

-- Para el worker, que pasa por todas las tiendas (sólo ids): las que están en Gratis con la prueba
-- por terminar (3 días) o recién terminada (hasta 7 días) y sin el aviso de ese momento.
CREATE POLICY pruebas_por_terminar ON tiendas FOR SELECT TO mostrador_migrador
	USING (current_setting('app.pruebas_por_terminar', true) = 'si');
-- +goose StatementBegin
CREATE FUNCTION app.pruebas_por_terminar(p_limite int) RETURNS SETOF uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
BEGIN
	PERFORM set_config('app.pruebas_por_terminar', 'si', true);
	RETURN QUERY SELECT t.id FROM tiendas t, LATERAL (SELECT coalesce(t.prueba_hasta, t.creada_en + interval '14 days') AS fin) f
		WHERE t.plan = 'gratis' AND t.cerrada_en IS NULL AND NOT t.demo
			AND f.fin <= now() + interval '3 days' AND f.fin > now() - interval '7 days'
			AND (t.aviso_prueba IS NULL OR t.aviso_prueba_fin IS DISTINCT FROM f.fin OR (t.aviso_prueba > 0 AND f.fin <= now()))
		ORDER BY f.fin LIMIT p_limite;
	PERFORM set_config('app.pruebas_por_terminar', '', true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.pruebas_por_terminar(int) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.pruebas_por_terminar(int) TO mostrador_app;

-- +goose Down
DROP FUNCTION app.pruebas_por_terminar(int);
DROP POLICY pruebas_por_terminar ON tiendas;
ALTER TABLE tiendas DROP COLUMN aviso_prueba_fin, DROP COLUMN aviso_prueba;
