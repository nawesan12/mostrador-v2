-- +goose Up
-- Actividad de la tienda (docs/decisiones.md §38): cada sección anota en `auditoria` quién hizo qué.
-- Se guarda el nombre y el rol de quien lo hizo en ese momento (si después la quitan del equipo, el
-- registro sigue diciendo quién fue), la IA si fue una IA conectada, y la frase ya escrita.
ALTER TABLE auditoria DROP CONSTRAINT auditoria_actor_tipo_check;
ALTER TABLE auditoria ADD CONSTRAINT auditoria_actor_tipo_check
	CHECK (actor_tipo IN ('persona', 'ia', 'hornero', 'sistema', 'staff', 'integracion'));
ALTER TABLE auditoria
	ADD COLUMN actor_nombre text CHECK (length(actor_nombre) <= 120),
	ADD COLUMN actor_rol text CHECK (actor_rol IN ('duenio', 'admin', 'ventas', 'local')),
	ADD COLUMN actor_ia text CHECK (length(actor_ia) <= 80),
	ADD COLUMN texto text CHECK (length(texto) <= 500);
-- El filtro por persona de Actividad.
CREATE INDEX auditoria_tienda_actor ON auditoria (tienda_id, actor_id, creado_en DESC, id DESC);
-- La purga recorre por fecha, de todas las tiendas.
CREATE INDEX auditoria_vieja ON auditoria (creado_en);

-- La purga ve todas las tiendas sólo dentro de app.purgar_auditoria (como app.purgar_visitas).
CREATE POLICY purga ON auditoria TO mostrador_migrador
	USING (current_setting('app.purga_auditoria', true) = 'si');

-- +goose StatementBegin
-- Retención: un año. La corre el worker con el rol de la app. Devuelve cuántas filas borró.
CREATE FUNCTION app.purgar_auditoria() RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	n bigint;
BEGIN
	PERFORM set_config('app.purga_auditoria', 'si', true);
	DELETE FROM auditoria WHERE creado_en < now() - interval '1 year';
	GET DIAGNOSTICS n = ROW_COUNT;
	PERFORM set_config('app.purga_auditoria', '', true);
	RETURN n;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.purgar_auditoria() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.purgar_auditoria() TO mostrador_app;

-- +goose Down
SELECT set_config('app.purga_auditoria', 'si', true);
DELETE FROM auditoria WHERE actor_tipo = 'ia';
DROP FUNCTION app.purgar_auditoria();
DROP POLICY purga ON auditoria;
DROP INDEX auditoria_vieja;
DROP INDEX auditoria_tienda_actor;
ALTER TABLE auditoria DROP COLUMN actor_nombre, DROP COLUMN actor_rol, DROP COLUMN actor_ia, DROP COLUMN texto;
ALTER TABLE auditoria DROP CONSTRAINT auditoria_actor_tipo_check;
ALTER TABLE auditoria ADD CONSTRAINT auditoria_actor_tipo_check
	CHECK (actor_tipo IN ('persona', 'hornero', 'sistema', 'staff', 'integracion'));
