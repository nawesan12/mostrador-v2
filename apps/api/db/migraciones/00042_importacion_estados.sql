-- +goose Up
-- El state del OAuth de las plataformas de las que se importa (hoy, Tiendanube): de un solo uso,
-- 10 minutos y sólo su hash, como ml_estados. Tiendanube no usa PKCE, así que no hay verificador.
CREATE TABLE importacion_estados (
	estado_hash text PRIMARY KEY,
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	usuario_id uuid NOT NULL,
	plataforma text NOT NULL CHECK (plataforma IN ('tiendanube')),
	vence_en timestamptz NOT NULL,
	usado_en timestamptz
);
CREATE INDEX importacion_estados_tienda ON importacion_estados (tienda_id, vence_en);
SELECT app.aplicar_rls('importacion_estados');
GRANT SELECT, INSERT, UPDATE (usado_en), DELETE ON importacion_estados TO mostrador_app;

-- +goose Down
DROP TABLE importacion_estados;
