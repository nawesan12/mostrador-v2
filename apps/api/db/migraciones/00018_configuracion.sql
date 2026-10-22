-- +goose Up
-- Configuración: datos del negocio, preferencias (seguridad y avisos), ciclo del plan e invitaciones.
ALTER TABLE tiendas
	ADD COLUMN negocio jsonb NOT NULL DEFAULT '{}' CHECK (pg_column_size(negocio) <= 4096),
	ADD COLUMN preferencias jsonb NOT NULL DEFAULT '{}' CHECK (pg_column_size(preferencias) <= 4096),
	ADD COLUMN plan_ciclo text NOT NULL DEFAULT 'mes' CHECK (plan_ciclo IN ('mes', 'anio'));
GRANT UPDATE (nombre, negocio, preferencias, plan, plan_ciclo) ON tiendas TO mostrador_app;

CREATE TABLE invitaciones (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	contacto text NOT NULL CHECK (length(contacto) BETWEEN 6 AND 120),
	rol text NOT NULL CHECK (rol IN ('admin', 'ventas', 'deposito')),
	creada_en timestamptz NOT NULL DEFAULT now(),
	UNIQUE (tienda_id, contacto)
);
SELECT app.aplicar_rls('invitaciones');
GRANT SELECT, INSERT, DELETE ON invitaciones TO mostrador_app;

-- +goose Down
DROP TABLE invitaciones;
ALTER TABLE tiendas DROP COLUMN negocio, DROP COLUMN preferencias, DROP COLUMN plan_ciclo;
