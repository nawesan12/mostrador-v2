-- +goose Up
-- Volver a verificarse (contraseña, app autenticadora o un código por mail) antes de cambiar adónde
-- llega la plata: el alias, el CBU y el titular, y la cuenta de Mercado Pago. Vale 5 minutos y sólo
-- para la sesión que lo pidió.
CREATE TABLE auth.reverificaciones (
	sesion_hash text PRIMARY KEY REFERENCES auth.sesiones (token_hash) ON DELETE CASCADE,
	codigo_hash text,
	codigo_vence timestamptz,
	intentos integer NOT NULL DEFAULT 0,
	token_hash text,
	token_vence timestamptz
);
GRANT SELECT, INSERT, UPDATE, DELETE ON auth.reverificaciones TO mostrador_app;

-- Cada cambio de adónde llega la plata, con lo que había antes: el mail a la dueña trae un link
-- (token de un solo uso, 7 días) para volverlo atrás y cerrar todas las sesiones del equipo.
CREATE TABLE auth.cobros_cambios (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	tipo text NOT NULL CHECK (tipo IN ('transferencia', 'mercadopago')),
	antes jsonb NOT NULL,
	despues jsonb NOT NULL,
	usuario_id uuid REFERENCES auth.usuarios (id) ON DELETE SET NULL,
	token_hash text NOT NULL UNIQUE,
	creado_en timestamptz NOT NULL DEFAULT now(),
	revertido_en timestamptz
);
CREATE INDEX cobros_cambios_tienda ON auth.cobros_cambios (tienda_id, creado_en DESC);
GRANT SELECT, INSERT, UPDATE (revertido_en) ON auth.cobros_cambios TO mostrador_app;

-- +goose Down
DROP TABLE auth.cobros_cambios;
DROP TABLE auth.reverificaciones;
