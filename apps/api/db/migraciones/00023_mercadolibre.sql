-- +goose Up
-- Mercado Libre como canal: cada tienda conecta su cuenta por OAuth (PKCE) y Mostrador publica,
-- sincroniza stock y precio, trae los pedidos y las preguntas. Tokens cifrados como los de Mercado Pago.
CREATE TABLE ml_conexiones (
	tienda_id uuid PRIMARY KEY REFERENCES tiendas (id) ON DELETE CASCADE,
	ml_usuario bigint NOT NULL UNIQUE,
	apodo text NOT NULL DEFAULT '',
	acceso bytea NOT NULL,
	renovacion bytea NOT NULL,
	vence_en timestamptz NOT NULL,
	conectada_en timestamptz NOT NULL DEFAULT now(),
	reautorizar boolean NOT NULL DEFAULT false,
	-- Recargo sobre el precio de la tienda para cubrir la comisión (la opción "lista Mercado Libre").
	recargo int NOT NULL DEFAULT 10 CHECK (recargo BETWEEN 0 AND 60),
	-- Pausar la publicación cuando el stock llega a cero.
	pausar_sin_stock boolean NOT NULL DEFAULT true
);
SELECT app.aplicar_rls('ml_conexiones');
GRANT SELECT, INSERT, UPDATE, DELETE ON ml_conexiones TO mostrador_app;

CREATE TABLE ml_estados (
	estado_hash text PRIMARY KEY,
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	usuario_id uuid NOT NULL,
	verificador text NOT NULL,
	vence_en timestamptz NOT NULL,
	usado_en timestamptz
);
SELECT app.aplicar_rls('ml_estados');
GRANT SELECT, INSERT, UPDATE (usado_en), DELETE ON ml_estados TO mostrador_app;

-- Qué producto está publicado en qué publicación.
CREATE TABLE ml_publicaciones (
	tienda_id uuid NOT NULL,
	producto_id uuid NOT NULL,
	item_id text NOT NULL UNIQUE,
	permalink text NOT NULL DEFAULT '',
	estado text NOT NULL DEFAULT 'active',
	sincronizado_en timestamptz,
	error text,
	PRIMARY KEY (tienda_id, producto_id),
	FOREIGN KEY (tienda_id, producto_id) REFERENCES productos (tienda_id, id) ON DELETE CASCADE
);
SELECT app.aplicar_rls('ml_publicaciones');
GRANT SELECT, INSERT, UPDATE (estado, sincronizado_en, error, permalink), DELETE ON ml_publicaciones TO mostrador_app;

-- Pedidos que ya entraron (una orden de Mercado Libre es un solo pedido aunque avise varias veces).
CREATE TABLE ml_pedidos (
	orden_ml bigint PRIMARY KEY,
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	numero int NOT NULL,
	creado_en timestamptz NOT NULL DEFAULT now()
);
SELECT app.aplicar_rls('ml_pedidos');
GRANT SELECT, INSERT ON ml_pedidos TO mostrador_app;

-- Preguntas de las publicaciones, con la respuesta (de la tienda o la sugerida por Hornero).
CREATE TABLE ml_preguntas (
	pregunta_id bigint PRIMARY KEY,
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	item_id text NOT NULL,
	texto text NOT NULL,
	respuesta text,
	estado text NOT NULL DEFAULT 'UNANSWERED',
	creada_en timestamptz NOT NULL DEFAULT now(),
	respondida_en timestamptz
);
CREATE INDEX ml_preguntas_tienda ON ml_preguntas (tienda_id, creada_en DESC);
SELECT app.aplicar_rls('ml_preguntas');
GRANT SELECT, INSERT, UPDATE (respuesta, estado, respondida_en) ON ml_preguntas TO mostrador_app;

-- Un producto publicado que cambia de stock, precio o visibilidad se encola para sincronizar. El
-- trigger agarra todos los caminos (punto de venta, tienda, ajustes, Mercado Libre mismo).
-- +goose StatementBegin
CREATE FUNCTION app.ml_encolar() RETURNS trigger
LANGUAGE plpgsql SET search_path = public, app AS $$
BEGIN
	IF (NEW.stock, NEW.precio, NEW.publicado) IS DISTINCT FROM (OLD.stock, OLD.precio, OLD.publicado)
		AND EXISTS (SELECT 1 FROM ml_publicaciones WHERE tienda_id = NEW.tienda_id AND producto_id = NEW.id) THEN
		INSERT INTO outbox (tienda_id, tipo, datos)
		VALUES (NEW.tienda_id, 'ml.sincronizar', jsonb_build_object('producto', NEW.id));
	END IF;
	RETURN NEW;
END
$$;
-- +goose StatementEnd
CREATE TRIGGER ml_sincronizar AFTER UPDATE OF stock, precio, publicado ON productos
	FOR EACH ROW EXECUTE FUNCTION app.ml_encolar();

-- Los avisos traen el usuario de Mercado Libre: de qué tienda es (como app.tienda_de_mp).
CREATE POLICY por_usuario_ml ON ml_conexiones FOR SELECT TO mostrador_migrador
	USING (ml_usuario::text = current_setting('app.ml_usuario', true));
-- +goose StatementBegin
CREATE FUNCTION app.tienda_de_ml(p_usuario bigint) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	t uuid;
BEGIN
	PERFORM set_config('app.ml_usuario', p_usuario::text, true);
	SELECT tienda_id INTO t FROM ml_conexiones WHERE ml_usuario = p_usuario;
	PERFORM set_config('app.ml_usuario', '', true);
	RETURN t;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.tienda_de_ml(bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.tienda_de_ml(bigint) TO mostrador_app;

-- Los pedidos de Mercado Libre se despachan con Mercado Envíos.
ALTER TABLE pedidos DROP CONSTRAINT pedidos_envio_check;
ALTER TABLE pedidos ADD CONSTRAINT pedidos_envio_check
	CHECK (envio IN ('andreani', 'oca', 'correo', 'moto', 'retiro', 'mercadoenvios'));
GRANT UPDATE (numero) ON ml_pedidos TO mostrador_app;
GRANT DELETE ON ml_pedidos TO mostrador_app;

-- +goose Down
ALTER TABLE pedidos DROP CONSTRAINT pedidos_envio_check;
ALTER TABLE pedidos ADD CONSTRAINT pedidos_envio_check CHECK (envio IN ('andreani', 'oca', 'correo', 'moto', 'retiro'));
DROP FUNCTION app.tienda_de_ml(bigint);
DROP TRIGGER ml_sincronizar ON productos;
DROP FUNCTION app.ml_encolar();
DROP TABLE ml_preguntas;
DROP TABLE ml_pedidos;
DROP TABLE ml_publicaciones;
DROP TABLE ml_estados;
DROP TABLE ml_conexiones;
