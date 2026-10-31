-- +goose Up
-- Equipo de cada tienda (docs/decisiones.md §37): roles fijos aplicados en Go e invitaciones por
-- mail con link de un solo uso.

-- «Depósito» pasa a «Local» (punto de venta y pedidos). Nadie podía aceptar una invitación, así
-- que no debería haber membresías con ese rol; si las hay, quedan como Local.
ALTER TABLE auth.membresias DROP CONSTRAINT membresias_rol_check;
UPDATE auth.membresias SET rol = 'local' WHERE rol = 'deposito';
ALTER TABLE auth.membresias ADD CONSTRAINT membresias_rol_check CHECK (rol IN ('duenio', 'admin', 'ventas', 'local'));

-- Las invitaciones pendientes guardadas antes quedan «sin mandar»: la dueña las reenvía o las cancela.
-- Como en 00033: RLS forzada no deja ver ninguna tienda al migrador, se suelta sólo para el UPDATE.
ALTER TABLE invitaciones DROP CONSTRAINT invitaciones_rol_check;
ALTER TABLE invitaciones NO FORCE ROW LEVEL SECURITY;
UPDATE invitaciones SET rol = 'local' WHERE rol = 'deposito';
ALTER TABLE invitaciones FORCE ROW LEVEL SECURITY;
ALTER TABLE invitaciones ADD CONSTRAINT invitaciones_rol_check CHECK (rol IN ('admin', 'ventas', 'local'));
ALTER TABLE invitaciones
	ADD COLUMN invitada_por uuid REFERENCES auth.usuarios (id) ON DELETE SET NULL,
	ADD COLUMN enviada_en timestamptz;
-- Cambiar el rol de una invitación pendiente y anotar cuándo se mandó (antes no había UPDATE).
GRANT UPDATE (rol, enviada_en) ON invitaciones TO mostrador_app;

-- El link de cada invitación: el sha256 del token, como las sesiones. Vive en auth (fuera de RLS)
-- porque quien abre el link todavía no es de la tienda: de acá sale la tienda y recién ahí se lee
-- la invitación con la tienda fijada. Uno solo vigente por invitación; vence a los 7 días.
CREATE TABLE auth.invitaciones (
	token_hash text PRIMARY KEY,
	invitacion_id uuid NOT NULL UNIQUE REFERENCES public.invitaciones (id) ON DELETE CASCADE,
	tienda_id uuid NOT NULL REFERENCES public.tiendas (id) ON DELETE CASCADE,
	vence_en timestamptz NOT NULL,
	creado_en timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX invitaciones_link_tienda ON auth.invitaciones (tienda_id);
GRANT SELECT, INSERT, DELETE ON auth.invitaciones TO mostrador_app;

-- +goose StatementBegin
-- Las tiendas de una persona, para elegir cuál ver. Con RLS forzada nadie ve tiendas ajenas: la
-- función fija cada tienda de sus membresías antes de leerla.
CREATE FUNCTION app.tiendas_de_usuario(p_usuario uuid) RETURNS TABLE (id uuid, nombre text, slug text, rol text)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	m record;
	antes text := current_setting('app.tienda_id', true);
BEGIN
	FOR m IN SELECT mm.tienda_id, mm.rol FROM auth.membresias mm WHERE mm.usuario_id = p_usuario ORDER BY mm.creada_en LOOP
		PERFORM set_config('app.tienda_id', m.tienda_id::text, true);
		RETURN QUERY SELECT t.id, t.nombre, t.slug, m.rol FROM tiendas t WHERE t.id = m.tienda_id AND t.cerrada_en IS NULL;
	END LOOP;
	PERFORM set_config('app.tienda_id', coalesce(antes, ''), true);
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.tiendas_de_usuario(uuid) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.tiendas_de_usuario(uuid) TO mostrador_app;

-- La matriz editable de permisos (tiendas.permisos) deja de usarse: los roles son fijos. La columna
-- queda (sin datos que valga la pena borrar en una migración con clientes reales).

-- +goose Down
DROP FUNCTION app.tiendas_de_usuario(uuid);
DROP TABLE auth.invitaciones;
REVOKE UPDATE (rol, enviada_en) ON invitaciones FROM mostrador_app;
ALTER TABLE invitaciones DROP COLUMN invitada_por, DROP COLUMN enviada_en;
ALTER TABLE invitaciones DROP CONSTRAINT invitaciones_rol_check;
ALTER TABLE invitaciones NO FORCE ROW LEVEL SECURITY;
UPDATE invitaciones SET rol = 'deposito' WHERE rol = 'local';
ALTER TABLE invitaciones FORCE ROW LEVEL SECURITY;
ALTER TABLE invitaciones ADD CONSTRAINT invitaciones_rol_check CHECK (rol IN ('admin', 'ventas', 'deposito'));
ALTER TABLE auth.membresias DROP CONSTRAINT membresias_rol_check;
UPDATE auth.membresias SET rol = 'deposito' WHERE rol = 'local';
ALTER TABLE auth.membresias ADD CONSTRAINT membresias_rol_check CHECK (rol IN ('duenio', 'admin', 'ventas', 'deposito'));
