-- +goose Up
-- Push web del panel instalado como app (docs/decisiones.md §42). Hermana de auth.dispositivos
-- (§40): cada navegador donde alguien activó los avisos guarda su suscripción de Web Push, atada a
-- esa sesión y a esa persona. Si la sesión se cierra, la fila se va con ella (ON DELETE CASCADE); si
-- vence, el worker no la elige. RLS forzada por persona (app.usuario_actual(), db.EnUsuario).
CREATE TABLE auth.suscripciones_web (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	usuario_id uuid NOT NULL REFERENCES auth.usuarios (id) ON DELETE CASCADE,
	sesion_hash text NOT NULL REFERENCES auth.sesiones (token_hash) ON DELETE CASCADE,
	-- La URL del servicio de push (FCM, Apple, Mozilla, Windows); Go valida el host.
	endpoint text NOT NULL UNIQUE CHECK (endpoint ~ '^https://' AND length(endpoint) <= 1000),
	-- Clave pública P-256 del navegador (65 bytes) y secreto auth (16), en base64url sin relleno.
	p256dh text NOT NULL CHECK (p256dh ~ '^[A-Za-z0-9_-]{87}$'),
	auth text NOT NULL CHECK (auth ~ '^[A-Za-z0-9_-]{22}$'),
	aparato text NOT NULL DEFAULT '' CHECK (length(aparato) <= 60),
	creado_en timestamptz NOT NULL DEFAULT now(),
	actualizado_en timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX suscripciones_web_usuario ON auth.suscripciones_web (usuario_id);
CREATE INDEX suscripciones_web_sesion ON auth.suscripciones_web (sesion_hash);
ALTER TABLE auth.suscripciones_web ENABLE ROW LEVEL SECURITY;
ALTER TABLE auth.suscripciones_web FORCE ROW LEVEL SECURITY;
CREATE POLICY persona ON auth.suscripciones_web TO mostrador_app, mostrador_migrador
	USING (usuario_id = app.usuario_actual()) WITH CHECK (usuario_id = app.usuario_actual());
-- Una suscripción que pasa a otra persona (otra cuenta en el mismo navegador) o que el servicio da
-- por vencida (404/410) se borra por el endpoint, sólo desde las funciones.
CREATE POLICY por_endpoint_lee ON auth.suscripciones_web FOR SELECT TO mostrador_migrador
	USING (endpoint = current_setting('app.suscripcion_endpoint', true));
CREATE POLICY por_endpoint ON auth.suscripciones_web FOR DELETE TO mostrador_migrador
	USING (endpoint = current_setting('app.suscripcion_endpoint', true));
GRANT SELECT, DELETE ON auth.suscripciones_web TO mostrador_app;

-- Alta o renovación de la suscripción de la sesión. La sesión tiene que ser de la persona y estar
-- vigente. El endpoint es único: si estaba con otra sesión, se mueve a esta.
-- +goose StatementBegin
CREATE FUNCTION app.registrar_suscripcion_web(p_sesion text, p_endpoint text, p_p256dh text, p_auth text, p_aparato text)
RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	u uuid := app.usuario_actual();
	nueva uuid;
BEGIN
	IF u IS NULL OR NOT EXISTS (SELECT 1 FROM auth.sesiones s
		WHERE s.token_hash = p_sesion AND s.usuario_id = u AND s.vence_en > now()) THEN
		RAISE EXCEPTION 'sesión ajena o vencida' USING ERRCODE = '42501';
	END IF;
	PERFORM set_config('app.suscripcion_endpoint', p_endpoint, true);
	DELETE FROM auth.suscripciones_web WHERE endpoint = p_endpoint;
	PERFORM set_config('app.suscripcion_endpoint', '', true);
	INSERT INTO auth.suscripciones_web (usuario_id, sesion_hash, endpoint, p256dh, auth, aparato)
	VALUES (u, p_sesion, p_endpoint, p_p256dh, p_auth, coalesce(p_aparato, ''))
	RETURNING id INTO nueva;
	RETURN nueva;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.registrar_suscripcion_web(text, text, text, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.registrar_suscripcion_web(text, text, text, text, text) TO mostrador_app;

-- Lo que el servicio de push responde como vencido (404, 410) se borra solo.
-- +goose StatementBegin
CREATE FUNCTION app.olvidar_suscripcion_web(p_endpoint text) RETURNS void
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
BEGIN
	PERFORM set_config('app.suscripcion_endpoint', p_endpoint, true);
	DELETE FROM auth.suscripciones_web WHERE endpoint = p_endpoint;
	PERFORM set_config('app.suscripcion_endpoint', '', true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.olvidar_suscripcion_web(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.olvidar_suscripcion_web(text) TO mostrador_app;

-- Las suscripciones del equipo de la tienda fijada con sesión vigente (y no inactiva), con el rol de
-- cada persona: Go decide con el rol qué ve cada una. Igual que app.dispositivos_de_tienda().
-- +goose StatementBegin
CREATE FUNCTION app.suscripciones_web_de_tienda()
RETURNS TABLE (usuario_id uuid, rol text, endpoint text, p256dh text, auth text)
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
		RETURN QUERY SELECT w.usuario_id, m.rol, w.endpoint, w.p256dh, w.auth
			FROM auth.suscripciones_web w
			JOIN auth.sesiones s ON s.token_hash = w.sesion_hash
			JOIN auth.usuarios u ON u.id = w.usuario_id
			WHERE w.usuario_id = m.usuario_id AND s.vence_en > now()
				AND NOT (coalesce((u.seguridad->>'inactivas')::boolean, false) AND s.visto_en < now() - interval '7 days');
	END LOOP;
	PERFORM set_config('app.usuario_id', coalesce(antes, ''), true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.suscripciones_web_de_tienda() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.suscripciones_web_de_tienda() TO mostrador_app;

-- +goose Down
DELETE FROM outbox WHERE tipo = 'cuenta.push_prueba';
DROP FUNCTION app.suscripciones_web_de_tienda();
DROP FUNCTION app.olvidar_suscripcion_web(text);
DROP FUNCTION app.registrar_suscripcion_web(text, text, text, text, text);
DROP TABLE auth.suscripciones_web;
