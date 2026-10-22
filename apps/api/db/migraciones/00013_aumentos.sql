-- +goose Up
-- Aumentos de precio por inflación: el último aplicado guarda los precios de antes (para deshacer)
-- y uno puede quedar programado para el lunes.
CREATE TABLE aumentos (
	id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	pct int NOT NULL CHECK (pct BETWEEN 1 AND 50),
	redondeo text NOT NULL CHECK (redondeo IN ('no', '90', '500')),
	para timestamptz,
	aplicado_en timestamptz,
	deshecho_en timestamptz,
	antes jsonb
);
CREATE INDEX aumentos_tienda ON aumentos (tienda_id, id DESC);
SELECT app.aplicar_rls('aumentos');
GRANT SELECT, INSERT, DELETE ON aumentos TO mostrador_app;
GRANT UPDATE (aplicado_en, deshecho_en, antes) ON aumentos TO mostrador_app;

-- +goose Down
DROP TABLE aumentos;
