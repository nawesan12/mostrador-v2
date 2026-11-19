-- +goose Up
-- «Confirmar» una propuesta de Tincho una sola vez: cada propuesta trae un id y el BFF lo toma antes de
-- ejecutarla y guarda el resultado. Un reintento (se cortó la red, otro dispositivo) devuelve lo
-- guardado en vez de sumar el stock dos veces o crear otro link.
CREATE TABLE tincho_confirmaciones (
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	id text NOT NULL CHECK (length(id) BETWEEN 8 AND 64),
	estado text NOT NULL CHECK (estado IN ('en_curso', 'hecha', 'fallo')),
	resultado jsonb,
	creado_en timestamptz NOT NULL DEFAULT now(),
	actualizado_en timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (tienda_id, id)
);
SELECT app.aplicar_rls('tincho_confirmaciones');
GRANT SELECT, INSERT, UPDATE (estado, resultado, actualizado_en), DELETE ON tincho_confirmaciones TO mostrador_app;

-- +goose Down
DROP TABLE tincho_confirmaciones;
