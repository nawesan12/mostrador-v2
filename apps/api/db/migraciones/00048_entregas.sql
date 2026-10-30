-- +goose Up
-- Cómo entrega cada tienda (Configuración › Envíos y retiro). Sin fila: retiro en el local con la
-- dirección y el horario del negocio, y sin moto (lo que hacía la tienda antes de esta pantalla).
-- Dirección y horario vacíos: los del negocio. Correo Argentino no va acá: se ofrece si hay fila en
-- ca_conexiones. El costo de la moto es en pesos, fijo.
CREATE TABLE entregas (
	tienda_id uuid PRIMARY KEY REFERENCES tiendas (id) ON DELETE CASCADE,
	retiro boolean NOT NULL DEFAULT true,
	retiro_direccion text NOT NULL DEFAULT '' CHECK (length(retiro_direccion) <= 160),
	retiro_horario text NOT NULL DEFAULT '' CHECK (length(retiro_horario) <= 80),
	moto boolean NOT NULL DEFAULT false,
	moto_costo integer NOT NULL DEFAULT 0 CHECK (moto_costo BETWEEN 0 AND 1000000),
	moto_zona text NOT NULL DEFAULT '' CHECK (length(moto_zona) <= 160),
	moto_demora text NOT NULL DEFAULT '' CHECK (length(moto_demora) <= 60),
	actualizada_en timestamptz NOT NULL DEFAULT now()
);
SELECT app.aplicar_rls('entregas');
GRANT SELECT, INSERT, UPDATE ON entregas TO mostrador_app;

-- +goose Down
DROP TABLE entregas;
