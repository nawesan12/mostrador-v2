-- +goose Up
-- Programa de partners (/partners/panel). Un partner es una cuenta de usuario aprobada (la
-- postulación vive en soporte.postulaciones_partner; aprobar la hace el backoffice). No son datos de
-- una tienda: el aislamiento es por persona, con la persona de la transacción en `app.usuario_id`
-- (db.EnUsuario), igual que `app.tienda_id` para las tiendas. RLS forzada en las tres tablas.
CREATE SCHEMA partners;
GRANT USAGE ON SCHEMA partners TO mostrador_app, mostrador_staff;

-- +goose StatementBegin
CREATE FUNCTION app.usuario_actual() RETURNS uuid
LANGUAGE sql STABLE AS $$
	SELECT nullif(current_setting('app.usuario_id', true), '')::uuid
$$;
-- +goose StatementEnd

CREATE TABLE partners.partners (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	usuario_id uuid NOT NULL UNIQUE REFERENCES auth.usuarios (id) ON DELETE CASCADE,
	-- El nivel que dio el equipo; el que se usa es el mayor entre este y el que dan las tiendas
	-- activas (Plata desde 5, Oro desde 20).
	nivel text NOT NULL DEFAULT 'partner' CHECK (nivel IN ('partner', 'plata', 'oro')),
	-- El ref del link (/empezar?ref=…), siempre en minúsculas; y el cupón propio, en mayúsculas.
	codigo text NOT NULL UNIQUE CHECK (codigo ~ '^[a-z0-9][a-z0-9-]{1,38}[a-z0-9]$'),
	cupon text NOT NULL UNIQUE CHECK (cupon ~ '^[A-Z0-9][A-Z0-9-]{1,38}[A-Z0-9]$'),
	perfil jsonb NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(perfil) = 'object' AND pg_column_size(perfil) <= 8192),
	cobro jsonb CHECK (cobro IS NULL OR (jsonb_typeof(cobro) = 'object' AND pg_column_size(cobro) <= 1024)),
	postulacion_id uuid REFERENCES soporte.postulaciones_partner (id) ON DELETE SET NULL,
	aprobado_en timestamptz NOT NULL DEFAULT now()
);

-- Clics del link: un contador por mes.
CREATE TABLE partners.clics (
	partner_id uuid NOT NULL REFERENCES partners.partners (id) ON DELETE CASCADE,
	mes date NOT NULL CHECK (extract(day FROM mes) = 1),
	n int NOT NULL DEFAULT 0 CHECK (n >= 0),
	PRIMARY KEY (partner_id, mes)
);

-- Pagos: la liquidación de un mes (la arma el equipo), con el detalle por tienda congelado y la
-- factura que sube el partner.
CREATE TABLE partners.pagos (
	partner_id uuid NOT NULL REFERENCES partners.partners (id) ON DELETE CASCADE,
	mes date NOT NULL CHECK (extract(day FROM mes) = 1),
	estado text NOT NULL DEFAULT 'pendiente' CHECK (estado IN ('pendiente', 'aprobada', 'pagada')),
	monto bigint NOT NULL CHECK (monto >= 0),
	detalle jsonb NOT NULL DEFAULT '[]' CHECK (jsonb_typeof(detalle) = 'array'),
	factura_url text CHECK (factura_url ~ '^https://' AND length(factura_url) <= 500),
	factura_en timestamptz,
	aprobado_en timestamptz,
	pagado_en timestamptz,
	creado_en timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (partner_id, mes)
);

-- Cada persona ve y toca sólo su ficha de partner; staff lee todo.
ALTER TABLE partners.partners ENABLE ROW LEVEL SECURITY;
ALTER TABLE partners.partners FORCE ROW LEVEL SECURITY;
CREATE POLICY persona ON partners.partners TO mostrador_app
	USING (usuario_id = app.usuario_actual()) WITH CHECK (usuario_id = app.usuario_actual());
CREATE POLICY staff_lee ON partners.partners FOR SELECT TO mostrador_staff USING (true);
GRANT SELECT, INSERT, UPDATE (perfil, cobro, nivel) ON partners.partners TO mostrador_app;

-- +goose StatementBegin
DO $$
DECLARE t text;
BEGIN
	FOREACH t IN ARRAY ARRAY['partners.clics', 'partners.pagos'] LOOP
		EXECUTE format('ALTER TABLE %s ENABLE ROW LEVEL SECURITY', t);
		EXECUTE format('ALTER TABLE %s FORCE ROW LEVEL SECURITY', t);
		EXECUTE format('CREATE POLICY persona ON %s TO mostrador_app USING (partner_id IN (SELECT id FROM partners.partners WHERE usuario_id = app.usuario_actual())) WITH CHECK (partner_id IN (SELECT id FROM partners.partners WHERE usuario_id = app.usuario_actual()))', t);
		EXECUTE format('CREATE POLICY staff_lee ON %s FOR SELECT TO mostrador_staff USING (true)', t);
		EXECUTE format('GRANT SELECT ON %s TO mostrador_staff', t);
	END LOOP;
END
$$;
-- +goose StatementEnd
GRANT SELECT ON partners.partners TO mostrador_staff;
GRANT SELECT ON partners.clics TO mostrador_app;
GRANT SELECT, INSERT, UPDATE (estado, monto, detalle, factura_url, factura_en, aprobado_en, pagado_en) ON partners.pagos TO mostrador_app;

-- Las tiendas referidas son de otras personas: RLS no deja ver ninguna. La función (dueña: el
-- migrador) abre sólo las del partner de la transacción, con la misma regla que el ref de /empezar
-- (sin distinguir mayúsculas) o el cupón propio, y sólo las columnas que hacen falta.
CREATE POLICY por_partner ON tiendas FOR SELECT TO mostrador_migrador
	USING (lower(origen->>'ref') = current_setting('app.ref_codigo', true)
		OR upper(origen->>'cupon') = current_setting('app.ref_cupon', true));
CREATE POLICY por_funcion ON partners.partners FOR SELECT TO mostrador_migrador
	USING (usuario_id = app.usuario_actual() OR codigo = current_setting('app.ref_codigo', true));
CREATE POLICY por_funcion ON partners.clics TO mostrador_migrador
	USING (partner_id = nullif(current_setting('app.ref_partner', true), '')::uuid)
	WITH CHECK (partner_id = nullif(current_setting('app.ref_partner', true), '')::uuid);
CREATE INDEX tiendas_ref_min ON tiendas (lower(origen->>'ref')) WHERE origen ? 'ref';
CREATE INDEX tiendas_cupon_origen ON tiendas (upper(origen->>'cupon')) WHERE origen ? 'cupon';

-- +goose StatementBegin
CREATE FUNCTION partners.tiendas_referidas()
RETURNS TABLE (id uuid, nombre text, slug text, plan text, plan_ciclo text, creada_en timestamptz,
	cancelada_en timestamptz, cerrada_en timestamptz)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	v_codigo text;
	v_cupon text;
BEGIN
	SELECT p.codigo, p.cupon INTO v_codigo, v_cupon FROM partners.partners p WHERE p.usuario_id = app.usuario_actual();
	IF v_codigo IS NULL THEN
		RETURN;
	END IF;
	PERFORM set_config('app.ref_codigo', v_codigo, true);
	PERFORM set_config('app.ref_cupon', v_cupon, true);
	RETURN QUERY SELECT t.id, t.nombre, t.slug, t.plan, t.plan_ciclo, t.creada_en, t.cancelada_en, t.cerrada_en
		FROM tiendas t
		WHERE lower(t.origen->>'ref') = v_codigo OR upper(t.origen->>'cupon') = v_cupon
		ORDER BY t.creada_en DESC, t.id;
	PERFORM set_config('app.ref_codigo', '', true);
	PERFORM set_config('app.ref_cupon', '', true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION partners.tiendas_referidas() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION partners.tiendas_referidas() TO mostrador_app;

-- Las tiendas donde la persona colabora (la dueña la sumó al equipo): nombre y dirección, nada más.
CREATE POLICY por_membresia ON tiendas FOR SELECT TO mostrador_migrador
	USING (id IN (SELECT m.tienda_id FROM auth.membresias m WHERE m.usuario_id = app.usuario_actual() AND m.rol <> 'duenio'));
-- +goose StatementBegin
CREATE FUNCTION partners.colaboraciones()
RETURNS TABLE (id uuid, nombre text, slug text, rol text)
LANGUAGE sql STABLE SECURITY DEFINER SET search_path = public, app AS $$
	SELECT t.id, t.nombre, t.slug, m.rol
	FROM auth.membresias m JOIN tiendas t ON t.id = m.tienda_id
	WHERE m.usuario_id = app.usuario_actual() AND m.rol <> 'duenio' AND t.cerrada_en IS NULL
	ORDER BY t.nombre
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION partners.colaboraciones() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION partners.colaboraciones() TO mostrador_app;

-- +goose StatementBegin
-- Un clic en el link (/empezar?ref=…). Sin sesión: lo llama el BFF. Un ref que no es de nadie no
-- hace nada.
CREATE FUNCTION partners.sumar_clic(p_codigo text) RETURNS boolean
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	v_id uuid;
BEGIN
	PERFORM set_config('app.ref_codigo', lower(p_codigo), true);
	SELECT p.id INTO v_id FROM partners.partners p WHERE p.codigo = lower(p_codigo);
	PERFORM set_config('app.ref_codigo', '', true);
	IF v_id IS NULL THEN
		RETURN false;
	END IF;
	PERFORM set_config('app.ref_partner', v_id::text, true);
	INSERT INTO partners.clics (partner_id, mes, n)
		VALUES (v_id, date_trunc('month', now() AT TIME ZONE 'America/Argentina/Buenos_Aires')::date, 1)
		ON CONFLICT (partner_id, mes) DO UPDATE SET n = partners.clics.n + 1;
	PERFORM set_config('app.ref_partner', '', true);
	RETURN true;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION partners.sumar_clic(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION partners.sumar_clic(text) TO mostrador_app;

-- +goose Down
DROP FUNCTION partners.sumar_clic(text);
DROP FUNCTION partners.tiendas_referidas();
DROP FUNCTION partners.colaboraciones();
DROP POLICY por_membresia ON tiendas;
DROP INDEX tiendas_cupon_origen;
DROP INDEX tiendas_ref_min;
DROP POLICY por_partner ON tiendas;
DROP SCHEMA partners CASCADE;
DROP FUNCTION app.usuario_actual();
