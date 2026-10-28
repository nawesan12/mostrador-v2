-- +goose Up
-- Backoffice del equipo de Mostrador (/staff). Las personas del equipo son cuentas aparte de las de
-- las tiendas (equipo.miembros), con su propia sesión (equipo.sesiones): un token de una tienda
-- nunca abre el backoffice, ni uno del equipo el panel de una tienda.
--
-- La API entra con el rol mostrador_staff, que lee todas las tiendas por la política staff_lee que
-- ya tiene cada tabla (app.aplicar_rls) y no puede escribir en ninguna. Lo que el equipo cambia en
-- una tienda pasa por equipo.accion_tienda (SECURITY DEFINER): chequea el permiso del rol, fija la
-- tienda en la transacción (la RLS sigue valiendo para el dueño de la función), cambia sólo lo de
-- esa acción y lo anota en la auditoría de la tienda y en la del equipo, todo junto.
CREATE SCHEMA equipo;
GRANT USAGE ON SCHEMA equipo TO mostrador_staff;
GRANT USAGE ON SCHEMA auth TO mostrador_staff;

CREATE TABLE equipo.miembros (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	email citext NOT NULL UNIQUE CHECK (length(email) <= 254),
	nombre text NOT NULL CHECK (length(nombre) BETWEEN 1 AND 80),
	clave_hash text NOT NULL,
	rol text NOT NULL CHECK (rol IN ('soporte', 'administracion', 'duenio')),
	activo boolean NOT NULL DEFAULT true,
	creado_en timestamptz NOT NULL DEFAULT now()
);
-- Las altas y bajas las hace `mostradorctl equipo-alta` con el migrador: la API sólo lee.
GRANT SELECT ON equipo.miembros TO mostrador_staff;

-- Qué puede hacer cada rol. Es la misma matriz que equipo.Permisos (Go) y lib/staff/permisos.ts;
-- los tests comparan las tres contra apps/api/internal/dominio/equipo/testdata/permisos.json.
CREATE TABLE equipo.permisos (
	rol text NOT NULL CHECK (rol IN ('soporte', 'administracion', 'duenio')),
	permiso text NOT NULL,
	PRIMARY KEY (rol, permiso)
);
INSERT INTO equipo.permisos (rol, permiso)
SELECT r, p FROM unnest(ARRAY['soporte', 'administracion', 'duenio']) r,
	unnest(ARRAY['tiendas.ver', 'tiendas.mirar', 'tiendas.notas', 'tiendas.prueba', 'tiendas.pack',
		'consultas.ver', 'consultas.responder']) p;
INSERT INTO equipo.permisos (rol, permiso)
SELECT r, p FROM unnest(ARRAY['administracion', 'duenio']) r,
	unnest(ARRAY['tiendas.plan', 'tiendas.credito', 'tiendas.pausar', 'partners.ver', 'partners.decidir',
		'facturacion.ver', 'auditoria.ver']) p;
INSERT INTO equipo.permisos (rol, permiso) VALUES ('duenio', 'tiendas.cerrar');
GRANT SELECT ON equipo.permisos TO mostrador_staff;

-- Sesiones del equipo: el sha256 del token, como auth.sesiones.
CREATE TABLE equipo.sesiones (
	token_hash text PRIMARY KEY,
	miembro_id uuid NOT NULL REFERENCES equipo.miembros (id) ON DELETE CASCADE,
	creada_en timestamptz NOT NULL DEFAULT now(),
	vence_en timestamptz NOT NULL,
	visto_en timestamptz NOT NULL DEFAULT now(),
	ip inet,
	agente text
);
CREATE INDEX sesiones_equipo_miembro ON equipo.sesiones (miembro_id);
GRANT SELECT, INSERT, UPDATE (visto_en), DELETE ON equipo.sesiones TO mostrador_staff;

-- "Ver la tienda como la ve la dueña": un token de sólo lectura para el panel de una tienda, corto.
CREATE TABLE equipo.miradas (
	token_hash text PRIMARY KEY,
	miembro_id uuid NOT NULL REFERENCES equipo.miembros (id) ON DELETE CASCADE,
	tienda_id uuid NOT NULL REFERENCES public.tiendas (id) ON DELETE CASCADE,
	creada_en timestamptz NOT NULL DEFAULT now(),
	vence_en timestamptz NOT NULL
);
GRANT SELECT, INSERT, DELETE ON equipo.miradas TO mostrador_staff;

-- Auditoría del equipo: quién hizo qué y cuándo, incluido entrar a ver una tienda. Sólo se agrega:
-- el rol no puede cambiar ni borrar registros.
CREATE TABLE equipo.auditoria (
	id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	miembro_id uuid NOT NULL REFERENCES equipo.miembros (id),
	accion text NOT NULL CHECK (accion ~ '^[a-z_]+$'),
	texto text NOT NULL CHECK (length(texto) <= 300),
	tienda_id uuid REFERENCES public.tiendas (id) ON DELETE SET NULL,
	detalle jsonb,
	ip inet,
	agente text,
	creado_en timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX auditoria_equipo_reciente ON equipo.auditoria (creado_en DESC, id DESC);
CREATE INDEX auditoria_equipo_tienda ON equipo.auditoria (tienda_id, creado_en DESC);
GRANT SELECT, INSERT ON equipo.auditoria TO mostrador_staff;

-- Notas internas del equipo sobre una tienda (la tienda no las ve).
CREATE TABLE equipo.notas (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL REFERENCES public.tiendas (id) ON DELETE CASCADE,
	miembro_id uuid NOT NULL REFERENCES equipo.miembros (id),
	texto text NOT NULL CHECK (length(texto) BETWEEN 1 AND 1000),
	creada_en timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX notas_tienda ON equipo.notas (tienda_id, creada_en DESC);
GRANT SELECT, INSERT ON equipo.notas TO mostrador_staff;

-- Respuestas por mail a las consultas (el mail sale por el outbox).
CREATE TABLE equipo.respuestas (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tipo text NOT NULL CHECK (tipo IN ('contacto', 'arrepentimiento')),
	ref text NOT NULL,
	miembro_id uuid NOT NULL REFERENCES equipo.miembros (id),
	para citext NOT NULL,
	asunto text NOT NULL CHECK (length(asunto) BETWEEN 3 AND 140),
	texto text NOT NULL CHECK (length(texto) BETWEEN 3 AND 4000),
	creada_en timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX respuestas_ref ON equipo.respuestas (tipo, ref, creada_en);
GRANT SELECT, INSERT ON equipo.respuestas TO mostrador_staff;

-- Consultas: a quién está asignada. Arrepentimientos: estado "en curso" (además de recibido,
-- procesado y rechazado). Postulaciones: el motivo y quién la resolvió.
ALTER TABLE soporte.consultas ADD COLUMN asignado_a uuid REFERENCES equipo.miembros (id) ON DELETE SET NULL;
ALTER TABLE soporte.arrepentimientos ADD COLUMN asignado_a uuid REFERENCES equipo.miembros (id) ON DELETE SET NULL;
ALTER TABLE soporte.arrepentimientos DROP CONSTRAINT arrepentimientos_estado_check;
ALTER TABLE soporte.arrepentimientos ADD CONSTRAINT arrepentimientos_estado_check
	CHECK (estado IN ('recibido', 'en_curso', 'procesado', 'rechazado'));
ALTER TABLE soporte.postulaciones_partner
	ADD COLUMN motivo text CHECK (length(motivo) <= 500),
	ADD COLUMN resuelta_en timestamptz,
	ADD COLUMN resuelta_por uuid REFERENCES equipo.miembros (id);
GRANT UPDATE (asignado_a) ON soporte.consultas, soporte.arrepentimientos TO mostrador_staff;
GRANT UPDATE (motivo, resuelta_en, resuelta_por) ON soporte.postulaciones_partner TO mostrador_staff;

-- La titular de cada tienda (nombre y email; nunca la clave).
GRANT SELECT (id, nombre, email, creado_en) ON auth.usuarios TO mostrador_staff;
GRANT SELECT ON auth.membresias TO mostrador_staff;

-- Tiendas: prueba extendida (NULL: la de siempre, 14 días desde el alta), saldo a favor, pausa y
-- motivo de la baja. Una tienda pausada no se ve en /t/ (app.tienda_de_slug).
ALTER TABLE tiendas
	ADD COLUMN prueba_hasta timestamptz,
	ADD COLUMN credito bigint NOT NULL DEFAULT 0 CHECK (credito BETWEEN 0 AND 1000000000),
	ADD COLUMN pausada_en timestamptz,
	ADD COLUMN baja_motivo text CHECK (length(baja_motivo) <= 500);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.tienda_de_slug(p_slug text) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	t uuid;
BEGIN
	PERFORM set_config('app.slug', p_slug, true);
	SELECT id INTO t FROM tiendas WHERE slug = p_slug AND cerrada_en IS NULL AND pausada_en IS NULL;
	PERFORM set_config('app.slug', '', true);
	RETURN t;
END
$$;
-- +goose StatementEnd

-- Mails de la plataforma (respuestas del equipo, avisos a partners): eventos del outbox sin tienda.
-- La app no los ve (su política compara tienda_id); el equipo sólo puede insertar los suyos.
ALTER TABLE outbox ALTER COLUMN tienda_id DROP NOT NULL;
CREATE POLICY equipo_inserta ON outbox FOR INSERT TO mostrador_staff
	WITH CHECK (tienda_id IS NULL AND tipo LIKE 'equipo.%');
GRANT INSERT ON outbox TO mostrador_staff;

-- +goose StatementBegin
-- Una acción del equipo sobre una tienda. Devuelve el id del registro de auditoría del equipo.
-- p_datos trae lo que la acción necesita (dias, plan, monto, cantidad, motivo), ya validado en Go;
-- acá se vuelve a chequear el permiso y lo que haría daño (montos, planes).
CREATE FUNCTION equipo.accion_tienda(p_miembro uuid, p_tienda uuid, p_accion text, p_datos jsonb,
	p_texto text, p_ip text, p_agente text) RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	v_permiso text;
	v_antes jsonb;
	v_despues jsonb;
	v_mes text := to_char(now() AT TIME ZONE 'America/Argentina/Buenos_Aires', 'YYYY-MM');
	v_id bigint;
BEGIN
	v_permiso := CASE p_accion
		WHEN 'extender_prueba' THEN 'tiendas.prueba'
		WHEN 'regalar_pack' THEN 'tiendas.pack'
		WHEN 'cambiar_plan' THEN 'tiendas.plan'
		WHEN 'aplicar_credito' THEN 'tiendas.credito'
		WHEN 'pausar' THEN 'tiendas.pausar'
		WHEN 'reactivar' THEN 'tiendas.pausar'
		WHEN 'cerrar' THEN 'tiendas.cerrar'
	END;
	IF v_permiso IS NULL THEN
		RAISE EXCEPTION 'acción desconocida: %', p_accion USING ERRCODE = '22023';
	END IF;
	IF NOT EXISTS (SELECT 1 FROM equipo.miembros m JOIN equipo.permisos p ON p.rol = m.rol
		WHERE m.id = p_miembro AND m.activo AND p.permiso = v_permiso) THEN
		RAISE EXCEPTION 'sin permiso para %', p_accion USING ERRCODE = '42501';
	END IF;
	PERFORM set_config('app.tienda_id', p_tienda::text, true);
	SELECT jsonb_build_object('plan', plan, 'pruebaHasta', coalesce(prueba_hasta, creada_en + interval '14 days'),
		'credito', credito, 'packs', coalesce((packs->>v_mes)::int, 0), 'pausadaEn', pausada_en, 'cerradaEn', cerrada_en)
		INTO v_antes FROM tiendas WHERE id = p_tienda FOR UPDATE;
	IF v_antes IS NULL THEN
		RAISE EXCEPTION 'no existe la tienda' USING ERRCODE = 'P0002';
	END IF;
	IF v_antes->>'cerradaEn' IS NOT NULL THEN
		RAISE EXCEPTION 'la tienda está cerrada' USING ERRCODE = '55000';
	END IF;
	CASE p_accion
	WHEN 'extender_prueba' THEN
		IF coalesce((p_datos->>'dias')::int, 0) NOT BETWEEN 1 AND 30 THEN
			RAISE EXCEPTION 'días fuera de rango' USING ERRCODE = '22023';
		END IF;
		UPDATE tiendas SET prueba_hasta = greatest(coalesce(prueba_hasta, creada_en + interval '14 days'), now())
			+ make_interval(days => (p_datos->>'dias')::int) WHERE id = p_tienda;
	WHEN 'cambiar_plan' THEN
		IF coalesce(p_datos->>'plan', '') NOT IN ('gratis', 'emprende', 'pro', 'negocio') THEN
			RAISE EXCEPTION 'plan desconocido' USING ERRCODE = '22023';
		END IF;
		UPDATE tiendas SET plan = p_datos->>'plan' WHERE id = p_tienda;
	WHEN 'aplicar_credito' THEN
		IF coalesce((p_datos->>'monto')::bigint, 0) NOT BETWEEN 100 AND 150000000 THEN
			RAISE EXCEPTION 'monto fuera de rango' USING ERRCODE = '22023';
		END IF;
		UPDATE tiendas SET credito = credito + (p_datos->>'monto')::bigint WHERE id = p_tienda;
	WHEN 'regalar_pack' THEN
		IF coalesce((p_datos->>'cantidad')::int, 0) NOT BETWEEN 1 AND 5 THEN
			RAISE EXCEPTION 'cantidad fuera de rango' USING ERRCODE = '22023';
		END IF;
		UPDATE tiendas SET packs = packs || jsonb_build_object(v_mes, coalesce((packs->>v_mes)::int, 0) + (p_datos->>'cantidad')::int)
			WHERE id = p_tienda;
	WHEN 'pausar' THEN
		UPDATE tiendas SET pausada_en = coalesce(pausada_en, now()) WHERE id = p_tienda;
	WHEN 'reactivar' THEN
		UPDATE tiendas SET pausada_en = NULL WHERE id = p_tienda;
	WHEN 'cerrar' THEN
		UPDATE tiendas SET cerrada_en = now(), cancelada_en = coalesce(cancelada_en, now()),
			baja_motivo = left(p_datos->>'motivo', 500) WHERE id = p_tienda;
	END CASE;
	SELECT jsonb_build_object('plan', plan, 'pruebaHasta', coalesce(prueba_hasta, creada_en + interval '14 days'),
		'credito', credito, 'packs', coalesce((packs->>v_mes)::int, 0), 'pausadaEn', pausada_en, 'cerradaEn', cerrada_en)
		INTO v_despues FROM tiendas WHERE id = p_tienda;
	INSERT INTO auditoria (tienda_id, actor_tipo, actor_id, accion, entidad, entidad_id, antes, despues, ip, agente)
		VALUES (p_tienda, 'staff', p_miembro::text, 'equipo.' || p_accion, 'tienda', p_tienda::text, v_antes, v_despues,
			nullif(p_ip, '')::inet, nullif(p_agente, ''));
	INSERT INTO equipo.auditoria (miembro_id, accion, texto, tienda_id, detalle, ip, agente)
		VALUES (p_miembro, p_accion, left(p_texto, 300), p_tienda,
			jsonb_build_object('datos', p_datos, 'antes', v_antes, 'despues', v_despues), nullif(p_ip, '')::inet, nullif(p_agente, ''))
		RETURNING id INTO v_id;
	PERFORM set_config('app.tienda_id', '', true);
	RETURN v_id;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION equipo.accion_tienda(uuid, uuid, text, jsonb, text, text, text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION equipo.accion_tienda(uuid, uuid, text, jsonb, text, text, text) TO mostrador_staff;

-- +goose Down
DROP FUNCTION equipo.accion_tienda(uuid, uuid, text, jsonb, text, text, text);
REVOKE INSERT ON outbox FROM mostrador_staff;
DROP POLICY equipo_inserta ON outbox;
DELETE FROM outbox WHERE tienda_id IS NULL;
ALTER TABLE outbox ALTER COLUMN tienda_id SET NOT NULL;
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.tienda_de_slug(p_slug text) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	t uuid;
BEGIN
	PERFORM set_config('app.slug', p_slug, true);
	SELECT id INTO t FROM tiendas WHERE slug = p_slug AND cerrada_en IS NULL;
	PERFORM set_config('app.slug', '', true);
	RETURN t;
END
$$;
-- +goose StatementEnd
ALTER TABLE tiendas DROP COLUMN prueba_hasta, DROP COLUMN credito, DROP COLUMN pausada_en, DROP COLUMN baja_motivo;
REVOKE SELECT ON auth.membresias FROM mostrador_staff;
REVOKE SELECT (id, nombre, email, creado_en) ON auth.usuarios FROM mostrador_staff;
ALTER TABLE soporte.postulaciones_partner DROP COLUMN motivo, DROP COLUMN resuelta_en, DROP COLUMN resuelta_por;
ALTER TABLE soporte.arrepentimientos DROP CONSTRAINT arrepentimientos_estado_check;
UPDATE soporte.arrepentimientos SET estado = 'recibido' WHERE estado = 'en_curso';
ALTER TABLE soporte.arrepentimientos ADD CONSTRAINT arrepentimientos_estado_check
	CHECK (estado IN ('recibido', 'procesado', 'rechazado'));
ALTER TABLE soporte.arrepentimientos DROP COLUMN asignado_a;
ALTER TABLE soporte.consultas DROP COLUMN asignado_a;
DROP SCHEMA equipo CASCADE;
REVOKE USAGE ON SCHEMA auth FROM mostrador_staff;
