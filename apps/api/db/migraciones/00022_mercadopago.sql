-- +goose Up
-- Mercado Pago: cada tienda conecta su cuenta por OAuth (marketplace) y cobra con su propio token.
-- Los tokens se guardan cifrados (AES-GCM, clave por tienda): la base sola no alcanza para usarlos.
CREATE TABLE mp_conexiones (
	tienda_id uuid PRIMARY KEY REFERENCES tiendas (id) ON DELETE CASCADE,
	mp_usuario bigint NOT NULL UNIQUE,
	acceso bytea NOT NULL,
	renovacion bytea NOT NULL,
	vence_en timestamptz NOT NULL,
	alcances text NOT NULL DEFAULT '',
	conectada_en timestamptz NOT NULL DEFAULT now(),
	-- El refresh falló: hay que volver a autorizar (nunca se cobra con el token de la plataforma).
	reautorizar boolean NOT NULL DEFAULT false,
	-- QR: la sucursal y la caja que se crearon en la cuenta de la tienda.
	sucursal_mp text,
	caja_externa text
);
SELECT app.aplicar_rls('mp_conexiones');
GRANT SELECT, INSERT, UPDATE, DELETE ON mp_conexiones TO mostrador_app;

-- Un OAuth de conexión en curso: un solo uso, 10 minutos, atado a la tienda y a quien lo inició.
CREATE TABLE mp_estados (
	estado_hash text PRIMARY KEY,
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	usuario_id uuid NOT NULL,
	-- PKCE: la aplicación lo exige.
	verificador text NOT NULL,
	vence_en timestamptz NOT NULL,
	usado_en timestamptz
);
SELECT app.aplicar_rls('mp_estados');
GRANT SELECT, INSERT, UPDATE (usado_en), DELETE ON mp_estados TO mostrador_app;

-- Cada cobro que se manda a Mercado Pago. referencia es el external_reference: único, para conciliar.
CREATE TABLE mp_cobros (
	referencia text PRIMARY KEY,
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	tipo text NOT NULL CHECK (tipo IN ('link', 'pedido', 'qr')),
	objetivo text NOT NULL,
	monto bigint NOT NULL CHECK (monto > 0),
	preferencia text,
	orden text,
	pago bigint,
	estado text NOT NULL DEFAULT 'pendiente',
	creado_en timestamptz NOT NULL DEFAULT now(),
	actualizado_en timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX mp_cobros_objetivo ON mp_cobros (tienda_id, tipo, objetivo);
SELECT app.aplicar_rls('mp_cobros');
GRANT SELECT, INSERT, UPDATE (preferencia, orden, pago, estado, actualizado_en) ON mp_cobros TO mostrador_app;

-- El abono de Mostrador (preapproval de la cuenta de la plataforma).
ALTER TABLE tiendas
	ADD COLUMN suscripcion_mp text,
	ADD COLUMN suscripcion_estado text,
	ADD COLUMN plan_pendiente text CHECK (plan_pendiente IN ('emprende', 'pro', 'negocio')),
	ADD COLUMN ciclo_pendiente text CHECK (ciclo_pendiente IN ('mes', 'anio'));
GRANT UPDATE (suscripcion_mp, suscripcion_estado, plan_pendiente, ciclo_pendiente) ON tiendas TO mostrador_app;

-- Los avisos de Mercado Pago traen el usuario de la cuenta, no la tienda: esto dice de qué tienda
-- es, como app.tienda_de_link (la RLS forzada también alcanza al dueño de la función).
CREATE POLICY por_usuario_mp ON mp_conexiones FOR SELECT TO mostrador_migrador
	USING (mp_usuario::text = current_setting('app.mp_usuario', true));
-- +goose StatementBegin
CREATE FUNCTION app.tienda_de_mp(p_usuario bigint) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	t uuid;
BEGIN
	PERFORM set_config('app.mp_usuario', p_usuario::text, true);
	SELECT tienda_id INTO t FROM mp_conexiones WHERE mp_usuario = p_usuario;
	PERFORM set_config('app.mp_usuario', '', true);
	RETURN t;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.tienda_de_mp(bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.tienda_de_mp(bigint) TO mostrador_app;

-- +goose Down
DROP FUNCTION app.tienda_de_mp(bigint);
ALTER TABLE tiendas DROP COLUMN suscripcion_mp, DROP COLUMN suscripcion_estado, DROP COLUMN plan_pendiente,
	DROP COLUMN ciclo_pendiente;
DROP TABLE mp_cobros;
DROP TABLE mp_estados;
DROP TABLE mp_conexiones;
