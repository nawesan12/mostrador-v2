-- +goose Up
-- Push en el iPhone (docs/decisiones.md §40). Cada iPhone donde alguien abrió sesión en la app
-- registra su token de APNs, atado a esa sesión y a esa persona: si la sesión se cierra (salir,
-- «Cerrar», contraseña nueva, quitar del equipo, inactiva) la fila se va con ella (ON DELETE
-- CASCADE); si vence, el worker no la elige. Es de la persona, no de una tienda: vive en auth, con
-- RLS forzada por persona (app.usuario_actual(), db.EnUsuario), como las fichas de partner.
CREATE TABLE auth.dispositivos (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	usuario_id uuid NOT NULL REFERENCES auth.usuarios (id) ON DELETE CASCADE,
	sesion_hash text NOT NULL REFERENCES auth.sesiones (token_hash) ON DELETE CASCADE,
	-- El token de APNs en hexadecimal (hoy 32 bytes; Apple avisa que puede crecer).
	token text NOT NULL UNIQUE CHECK (token ~ '^[0-9a-f]{64,200}$'),
	entorno text NOT NULL CHECK (entorno IN ('sandbox', 'produccion')),
	modelo text NOT NULL DEFAULT '' CHECK (length(modelo) <= 60),
	version_app text NOT NULL DEFAULT '' CHECK (length(version_app) <= 30),
	creado_en timestamptz NOT NULL DEFAULT now(),
	actualizado_en timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX dispositivos_usuario ON auth.dispositivos (usuario_id);
CREATE INDEX dispositivos_sesion ON auth.dispositivos (sesion_hash);
ALTER TABLE auth.dispositivos ENABLE ROW LEVEL SECURITY;
ALTER TABLE auth.dispositivos FORCE ROW LEVEL SECURITY;
-- Cada persona ve y borra sólo los suyos. Las funciones de abajo (del migrador) fijan la persona
-- antes de leer, así que la misma política las cubre.
CREATE POLICY persona ON auth.dispositivos TO mostrador_app, mostrador_migrador
	USING (usuario_id = app.usuario_actual()) WITH CHECK (usuario_id = app.usuario_actual());
-- Un token que pasa a otra persona (otra cuenta en el mismo iPhone) o que APNs da por muerto se
-- borra por el token, sin importar de quién era: sólo desde las funciones, que fijan cuál (un DELETE
-- con WHERE también pide poder leer la fila).
CREATE POLICY por_token_lee ON auth.dispositivos FOR SELECT TO mostrador_migrador
	USING (token = current_setting('app.dispositivo_token', true));
CREATE POLICY por_token ON auth.dispositivos FOR DELETE TO mostrador_migrador
	USING (token = current_setting('app.dispositivo_token', true));
GRANT SELECT, DELETE ON auth.dispositivos TO mostrador_app;

-- Alta o renovación del iPhone de la sesión. La persona sale de app.usuario_actual() y la sesión
-- tiene que ser suya y estar vigente. El token es único: si estaba con otra sesión (la misma persona
-- que volvió a entrar u otra cuenta en ese iPhone), se mueve a esta.
-- +goose StatementBegin
CREATE FUNCTION app.registrar_dispositivo(p_sesion text, p_token text, p_entorno text, p_modelo text, p_version text)
RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	u uuid := app.usuario_actual();
	nuevo uuid;
BEGIN
	IF u IS NULL OR NOT EXISTS (SELECT 1 FROM auth.sesiones s
		WHERE s.token_hash = p_sesion AND s.usuario_id = u AND s.vence_en > now()) THEN
		RAISE EXCEPTION 'sesión ajena o vencida' USING ERRCODE = '42501';
	END IF;
	PERFORM set_config('app.dispositivo_token', p_token, true);
	DELETE FROM auth.dispositivos WHERE token = p_token;
	PERFORM set_config('app.dispositivo_token', '', true);
	INSERT INTO auth.dispositivos (usuario_id, sesion_hash, token, entorno, modelo, version_app)
	VALUES (u, p_sesion, p_token, p_entorno, coalesce(p_modelo, ''), coalesce(p_version, ''))
	RETURNING id INTO nuevo;
	RETURN nuevo;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.registrar_dispositivo(text, text, text, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.registrar_dispositivo(text, text, text, text, text) TO mostrador_app;

-- Lo que APNs responde como inválido (410, BadDeviceToken, Unregistered) se borra solo.
-- +goose StatementBegin
CREATE FUNCTION app.olvidar_dispositivo(p_token text) RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
BEGIN
	PERFORM set_config('app.dispositivo_token', p_token, true);
	DELETE FROM auth.dispositivos WHERE token = p_token;
	PERFORM set_config('app.dispositivo_token', '', true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.olvidar_dispositivo(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.olvidar_dispositivo(text) TO mostrador_app;

-- Los iPhone del equipo de la tienda fijada (app.tienda_actual()) con sesión vigente, con el rol de
-- cada persona: Go decide con el rol qué ve cada una (montos sólo con Finanzas). Una sesión inactiva
-- (con «Cerrar sesiones inactivas» prendido, sin usar hace más de 7 días) no cuenta, igual que en
-- cuentas.Sesion.
-- +goose StatementBegin
CREATE FUNCTION app.dispositivos_de_tienda()
RETURNS TABLE (usuario_id uuid, rol text, token text, entorno text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	m record;
	antes text := current_setting('app.usuario_id', true);
BEGIN
	IF app.tienda_actual() IS NULL THEN
		RETURN;
	END IF;
	FOR m IN SELECT mm.usuario_id, mm.rol FROM auth.membresias mm WHERE mm.tienda_id = app.tienda_actual() LOOP
		PERFORM set_config('app.usuario_id', m.usuario_id::text, true);
		RETURN QUERY SELECT d.usuario_id, m.rol, d.token, d.entorno
			FROM auth.dispositivos d
			JOIN auth.sesiones s ON s.token_hash = d.sesion_hash
			JOIN auth.usuarios u ON u.id = d.usuario_id
			WHERE d.usuario_id = m.usuario_id AND s.vence_en > now()
				AND NOT (coalesce((u.seguridad->>'inactivas')::boolean, false) AND s.visto_en < now() - interval '7 days');
	END LOOP;
	PERFORM set_config('app.usuario_id', coalesce(antes, ''), true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.dispositivos_de_tienda() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.dispositivos_de_tienda() TO mostrador_app;

-- «Nueva venta #N» también por push: el mismo momento que el mail (§35), en un evento aparte para
-- que cada canal se reintente solo.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.avisar_venta_tienda() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	pagado constant text[] := ARRAY['Aprobado', 'Acreditada', 'Cobrado'];
BEGIN
	IF NEW.canal <> 'web' THEN
		RETURN NEW;
	END IF;
	IF (TG_OP = 'INSERT' AND (NEW.pago <> 'mp' OR NEW.estado_pago = ANY (pagado)))
		OR (TG_OP = 'UPDATE' AND NEW.pago = 'mp' AND NEW.estado_pago = ANY (pagado)
			AND NOT (OLD.estado_pago = ANY (pagado))) THEN
		INSERT INTO outbox (tienda_id, tipo, datos)
		VALUES (NEW.tienda_id, 'tienda.venta', jsonb_build_object('numero', NEW.numero)),
			(NEW.tienda_id, 'push.venta', jsonb_build_object('numero', NEW.numero));
	END IF;
	RETURN NEW;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.avisar_venta_tienda() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	pagado constant text[] := ARRAY['Aprobado', 'Acreditada', 'Cobrado'];
BEGIN
	IF NEW.canal <> 'web' THEN
		RETURN NEW;
	END IF;
	IF (TG_OP = 'INSERT' AND (NEW.pago <> 'mp' OR NEW.estado_pago = ANY (pagado)))
		OR (TG_OP = 'UPDATE' AND NEW.pago = 'mp' AND NEW.estado_pago = ANY (pagado)
			AND NOT (OLD.estado_pago = ANY (pagado))) THEN
		INSERT INTO outbox (tienda_id, tipo, datos)
		VALUES (NEW.tienda_id, 'tienda.venta', jsonb_build_object('numero', NEW.numero));
	END IF;
	RETURN NEW;
END
$$;
-- +goose StatementEnd
DELETE FROM outbox WHERE tipo LIKE 'push.%' OR tipo = 'cuenta.ingreso_push';
DROP FUNCTION app.dispositivos_de_tienda();
DROP FUNCTION app.olvidar_dispositivo(text);
DROP FUNCTION app.registrar_dispositivo(text, text, text, text, text);
DROP TABLE auth.dispositivos;
