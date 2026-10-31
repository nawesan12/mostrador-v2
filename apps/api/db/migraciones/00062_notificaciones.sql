-- +goose Up
-- Centro de notificaciones del panel (docs/decisiones.md §45): la campanita. Una fila por evento y
-- por persona del equipo que puede verlo (según su rol, §37). Las escribe el worker desde los mismos
-- eventos del outbox que mandan mail y push; la API sólo las lista y las marca leídas.
--
-- Aislamiento doble: por tienda (app.aplicar_rls, como toda tabla de negocio) y por persona
-- (políticas RESTRICTIVE: además de la tienda, sólo las propias). El worker inserta con la tienda
-- fijada y sin persona por app.notificar (para quien sea del equipo de esa tienda), sin poder leer.
CREATE TABLE notificaciones (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	usuario_id uuid NOT NULL REFERENCES auth.usuarios (id) ON DELETE CASCADE,
	-- De qué hecho sale (venta:12, pago:<referencia>…): reprocesar el evento no la duplica.
	clave text NOT NULL CHECK (length(clave) BETWEEN 1 AND 120),
	-- El ícono (las marcas de lib/marcas.ts: mp, ml, warn, ok, err, star, link…).
	tipo text NOT NULL CHECK (tipo ~ '^[a-z]{2,12}$'),
	grupo text NOT NULL CHECK (grupo IN ('ventas', 'sistema')),
	titulo text NOT NULL CHECK (length(titulo) BETWEEN 1 AND 120),
	detalle text NOT NULL DEFAULT '' CHECK (length(detalle) <= 300),
	-- La página del panel que abre al tocarla.
	ruta text NOT NULL CHECK (length(ruta) <= 200 AND ruta ~ '^/panel(/[a-z0-9/_?=&.-]*)?$'),
	creada_en timestamptz NOT NULL DEFAULT now(),
	leida_en timestamptz,
	UNIQUE (tienda_id, usuario_id, clave)
);
CREATE INDEX notificaciones_persona ON notificaciones (tienda_id, usuario_id, creada_en DESC, id DESC);
CREATE INDEX notificaciones_sin_leer ON notificaciones (tienda_id, usuario_id) WHERE leida_en IS NULL;
CREATE INDEX notificaciones_vieja ON notificaciones (creada_en);
SELECT app.aplicar_rls('notificaciones');
CREATE POLICY persona_lee ON notificaciones AS RESTRICTIVE FOR SELECT TO mostrador_app
	USING (usuario_id = app.usuario_actual());
CREATE POLICY persona_marca ON notificaciones AS RESTRICTIVE FOR UPDATE TO mostrador_app
	USING (usuario_id = app.usuario_actual()) WITH CHECK (usuario_id = app.usuario_actual());
GRANT SELECT, UPDATE (leida_en) ON notificaciones TO mostrador_app;

-- Alta desde el worker: con la tienda fijada y sin persona, la app no puede leer (ni el ON CONFLICT
-- que evita duplicar), así que inserta por esta función. Sólo en la tienda de la transacción y sólo
-- para quien es del equipo de esa tienda; lo demás se descarta. Devuelve cuántas creó.
-- +goose StatementBegin
CREATE FUNCTION app.notificar(p_clave text, p_usuarios uuid[], p_tipos text[], p_grupos text[], p_titulos text[],
	p_detalles text[], p_rutas text[]) RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	v_tienda uuid := app.tienda_actual();
	n bigint;
BEGIN
	IF v_tienda IS NULL THEN
		RAISE EXCEPTION 'notificar sin tienda' USING ERRCODE = '42501';
	END IF;
	INSERT INTO notificaciones (tienda_id, usuario_id, clave, tipo, grupo, titulo, detalle, ruta)
	SELECT v_tienda, x.u, p_clave, x.tp, x.g, x.ti, x.d, x.r
	FROM unnest(p_usuarios, p_tipos, p_grupos, p_titulos, p_detalles, p_rutas) AS x(u, tp, g, ti, d, r)
	JOIN auth.membresias m ON m.tienda_id = v_tienda AND m.usuario_id = x.u
	ON CONFLICT (tienda_id, usuario_id, clave) DO NOTHING;
	GET DIAGNOSTICS n = ROW_COUNT;
	RETURN n;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.notificar(text, uuid[], text[], text[], text[], text[], text[]) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.notificar(text, uuid[], text[], text[], text[], text[], text[]) TO mostrador_app;

-- La purga ve todas las tiendas sólo dentro de app.purgar_notificaciones (como app.purgar_auditoria).
CREATE POLICY purga ON notificaciones TO mostrador_migrador
	USING (current_setting('app.purga_notificaciones', true) = 'si');

-- +goose StatementBegin
-- Retención: 90 días. La corre el worker con el rol de la app una vez por día. Devuelve cuántas borró.
CREATE FUNCTION app.purgar_notificaciones() RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	n bigint;
BEGIN
	PERFORM set_config('app.purga_notificaciones', 'si', true);
	DELETE FROM notificaciones WHERE creada_en < now() - interval '90 days';
	GET DIAGNOSTICS n = ROW_COUNT;
	PERFORM set_config('app.purga_notificaciones', '', true);
	RETURN n;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.purgar_notificaciones() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.purgar_notificaciones() TO mostrador_app;

-- Tiempo real (§44): una notificación nueva o leída avisa sólo a esa persona en esa tienda ("u"),
-- sin id ni texto. Una sentencia que toca muchas (marcar todo) sale como un aviso por persona.
-- +goose StatementBegin
CREATE FUNCTION app.tr_tiempo_real_notificaciones() RETURNS trigger
LANGUAGE plpgsql SET search_path = public, app, pg_temp AS $$
DECLARE
	r record;
BEGIN
	IF TG_OP = 'INSERT' THEN
		FOR r IN SELECT DISTINCT tienda_id, usuario_id FROM nuevas LOOP
			PERFORM app.avisar_tiempo_real(r.tienda_id, 'notificacion.nueva', NULL, r.usuario_id);
		END LOOP;
	ELSE
		FOR r IN SELECT DISTINCT n.tienda_id, n.usuario_id FROM nuevas n JOIN viejas o USING (id)
			WHERE n.leida_en IS DISTINCT FROM o.leida_en
		LOOP
			PERFORM app.avisar_tiempo_real(r.tienda_id, 'notificacion.leida', NULL, r.usuario_id);
		END LOOP;
	END IF;
	RETURN NULL;
END
$$;
-- +goose StatementEnd
CREATE TRIGGER tiempo_real_notificacion_ins AFTER INSERT ON notificaciones REFERENCING NEW TABLE AS nuevas
	FOR EACH STATEMENT EXECUTE FUNCTION app.tr_tiempo_real_notificaciones();
CREATE TRIGGER tiempo_real_notificacion_upd AFTER UPDATE ON notificaciones REFERENCING OLD TABLE AS viejas NEW TABLE AS nuevas
	FOR EACH STATEMENT EXECUTE FUNCTION app.tr_tiempo_real_notificaciones();

-- +goose Down
DELETE FROM outbox WHERE tipo IN ('tienda.resena', 'integracion.reautorizar', 'equipo.aceptada', 'plan.activado', 'plan.pausado');
DROP TABLE notificaciones;
DROP FUNCTION app.tr_tiempo_real_notificaciones();
DROP FUNCTION app.notificar(text, uuid[], text[], text[], text[], text[], text[]);
DROP FUNCTION app.purgar_notificaciones();
