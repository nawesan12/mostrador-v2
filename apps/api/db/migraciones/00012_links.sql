-- +goose Up
-- Links de pago: se mandan por WhatsApp y los abre quien paga, sin sesión.
CREATE TABLE links (
	codigo text PRIMARY KEY,
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	monto bigint NOT NULL CHECK (monto > 0),
	concepto text NOT NULL,
	cuotas boolean NOT NULL DEFAULT true,
	pedido int,
	creado_en timestamptz NOT NULL DEFAULT now(),
	pagado_en timestamptz
);
CREATE INDEX links_tienda ON links (tienda_id, creado_en DESC);
SELECT app.aplicar_rls('links');
GRANT SELECT, INSERT ON links TO mostrador_app;
GRANT UPDATE (pagado_en) ON links TO mostrador_app;
GRANT UPDATE (estado_pago) ON pedidos TO mostrador_app;

-- La página pública sólo sabe el código: esto dice de qué tienda es, y el resto se lee con la RLS
-- de esa tienda.
-- La RLS forzada también alcanza al dueño de la función: esta política le deja ver sólo la fila del
-- código que se está buscando, y sólo desde la función.
CREATE POLICY por_codigo ON links FOR SELECT TO mostrador_migrador
	USING (codigo = current_setting('app.link_codigo', true));
-- +goose StatementBegin
CREATE FUNCTION app.tienda_de_link(p_codigo text) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	t uuid;
BEGIN
	PERFORM set_config('app.link_codigo', p_codigo, true);
	SELECT tienda_id INTO t FROM links WHERE codigo = p_codigo;
	PERFORM set_config('app.link_codigo', '', true);
	RETURN t;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.tienda_de_link(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.tienda_de_link(text) TO mostrador_app;

-- +goose Down
DROP FUNCTION app.tienda_de_link(text);
REVOKE UPDATE (estado_pago) ON pedidos FROM mostrador_app;
DROP TABLE links;
