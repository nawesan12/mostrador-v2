-- +goose Up
-- Cupones (design/PanelCupones.dc.html). Lo vendido y descontado se suma cuando el checkout los use.
CREATE TABLE cupones (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	codigo text NOT NULL CHECK (codigo ~ '^[A-Z0-9-]{3,20}$'),
	tipo text NOT NULL CHECK (tipo IN ('pct', 'fijo', 'envio', 'regalo')),
	valor int NOT NULL DEFAULT 0,
	descripcion text NOT NULL,
	pausado boolean NOT NULL DEFAULT false,
	usos int NOT NULL DEFAULT 0,
	limite int NOT NULL DEFAULT 0,
	vendio bigint NOT NULL DEFAULT 0,
	desconto bigint NOT NULL DEFAULT 0,
	canales text[] NOT NULL,
	empieza timestamptz,
	termina timestamptz,
	minimo bigint NOT NULL DEFAULT 0,
	quien text NOT NULL DEFAULT 'todos',
	por_cliente text NOT NULL DEFAULT '1',
	alcance jsonb NOT NULL DEFAULT '{"tipo": "todo"}',
	opciones jsonb NOT NULL DEFAULT '{}',
	creado_en timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX cupones_tienda ON cupones (tienda_id, creado_en DESC);
SELECT app.aplicar_rls('cupones');
GRANT SELECT, INSERT ON cupones TO mostrador_app;
GRANT UPDATE (pausado, minimo, descripcion) ON cupones TO mostrador_app;

-- Promociones automáticas: cuáles prendió la tienda.
ALTER TABLE tiendas ADD COLUMN automaticas jsonb NOT NULL DEFAULT
	'{"transf": true, "envio": true, "cantidad": true, "3x2": false, "cuotas": true, "mayorista": true, "carrito": true, "cumple": false}';
GRANT UPDATE (automaticas) ON tiendas TO mostrador_app;

-- +goose Down
ALTER TABLE tiendas DROP COLUMN automaticas;
DROP TABLE cupones;
