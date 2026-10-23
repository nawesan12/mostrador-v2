-- +goose Up
-- Carritos abandonados: cuando quien compra deja su email en el checkout y no termina, el carrito
-- queda guardado (uno abierto por email y tienda). Se cierra solo si después compra (recuperado).
-- El token del link de recordatorio no se adivina.
CREATE TABLE carritos (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	email text NOT NULL CHECK (length(email) BETWEEN 3 AND 120),
	nombre text NOT NULL DEFAULT '' CHECK (length(nombre) <= 80),
	items jsonb NOT NULL CHECK (jsonb_typeof(items) = 'array' AND jsonb_array_length(items) BETWEEN 1 AND 50),
	total bigint NOT NULL CHECK (total >= 0),
	token text NOT NULL UNIQUE CHECK (token ~ '^[A-Za-z0-9_-]{20,64}$'),
	creado timestamptz NOT NULL DEFAULT now(),
	actualizado timestamptz NOT NULL DEFAULT now(),
	recordado_en timestamptz,
	recordatorios int NOT NULL DEFAULT 0,
	recuperado_en timestamptz
);
CREATE UNIQUE INDEX carritos_abierto ON carritos (tienda_id, lower(email)) WHERE recuperado_en IS NULL;
CREATE INDEX carritos_recientes ON carritos (tienda_id, actualizado DESC);
SELECT app.aplicar_rls('carritos');
GRANT SELECT, INSERT, UPDATE (nombre, items, total, actualizado, recordado_en, recordatorios, recuperado_en) ON carritos TO mostrador_app;

-- El recordatorio automático por email al tercer día (si la tienda lo prendió): los carritos que
-- toca recordar, de todas las tiendas, para el worker. Sólo ids.
CREATE POLICY para_recordar ON carritos FOR SELECT TO mostrador_migrador
	USING (current_setting('app.carritos_recordar', true) = 'si');
CREATE POLICY carritos_recordar ON tiendas FOR SELECT TO mostrador_migrador
	USING (current_setting('app.carritos_recordar', true) = 'si');
-- +goose StatementBegin
CREATE FUNCTION app.carritos_para_recordar(p_limite int) RETURNS TABLE (tienda_id uuid, id uuid)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
BEGIN
	PERFORM set_config('app.carritos_recordar', 'si', true);
	RETURN QUERY SELECT c.tienda_id, c.id FROM carritos c JOIN tiendas t ON t.id = c.tienda_id
		WHERE c.recuperado_en IS NULL AND c.recordado_en IS NULL
			AND c.actualizado BETWEEN now() - interval '4 days' AND now() - interval '3 days'
			AND coalesce((t.carritos_automaticos->>'email')::boolean, false)
		LIMIT p_limite;
	PERFORM set_config('app.carritos_recordar', '', true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.carritos_para_recordar(int) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.carritos_para_recordar(int) TO mostrador_app;

-- +goose Down
DROP FUNCTION app.carritos_para_recordar(int);
DROP POLICY carritos_recordar ON tiendas;
DROP POLICY para_recordar ON carritos;
DROP TABLE carritos;
