-- +goose Up
-- Hornero en Telegram (@mostradorhornero_bot, para la dueña). Un chat por tienda y una tienda por
-- chat; el código de vínculo es de un solo uso, vence a los 10 minutos y se guarda su sha256.
-- (00036: las migraciones 00027–00035 las toman otras ramas; goose corre fuera de orden.)
CREATE TABLE telegram_vinculos (
	tienda_id uuid PRIMARY KEY REFERENCES tiendas (id) ON DELETE CASCADE,
	-- El id del chat privado con el bot. UNIQUE vale entre tiendas aunque RLS no deje verlas.
	chat bigint NOT NULL UNIQUE,
	nombre text NOT NULL CHECK (length(nombre) BETWEEN 1 AND 60),
	vinculado_en timestamptz NOT NULL DEFAULT now(),
	-- Mensajes a Hornero en el día (hora de Argentina), para los topes de Workers AI.
	uso_dia date,
	uso_n int NOT NULL DEFAULT 0 CHECK (uso_n >= 0)
);
SELECT app.aplicar_rls('telegram_vinculos');
GRANT SELECT, INSERT, DELETE ON telegram_vinculos TO mostrador_app;
GRANT UPDATE (chat, nombre, vinculado_en, uso_dia, uso_n) ON telegram_vinculos TO mostrador_app;

CREATE TABLE telegram_codigos (
	-- sha256 en base64url: 43 caracteres.
	hash text PRIMARY KEY CHECK (length(hash) = 43),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	vence_en timestamptz NOT NULL
);
CREATE INDEX telegram_codigos_tienda ON telegram_codigos (tienda_id);
SELECT app.aplicar_rls('telegram_codigos');
GRANT SELECT, INSERT, DELETE ON telegram_codigos TO mostrador_app;

-- El bot sólo sabe el chat o el código: de qué tienda son (como app.tienda_de_ml). Devuelven el id
-- y nada más; leer y escribir sigue siendo con la tienda fijada.
CREATE POLICY por_chat ON telegram_vinculos FOR SELECT TO mostrador_migrador
	USING (chat::text = current_setting('app.tg_chat', true));
CREATE POLICY por_hash ON telegram_codigos FOR SELECT TO mostrador_migrador
	USING (hash = current_setting('app.tg_codigo', true));
-- +goose StatementBegin
CREATE FUNCTION app.tienda_de_telegram(p_chat bigint) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	t uuid;
BEGIN
	PERFORM set_config('app.tg_chat', p_chat::text, true);
	SELECT tienda_id INTO t FROM telegram_vinculos WHERE chat = p_chat;
	PERFORM set_config('app.tg_chat', '', true);
	RETURN t;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.tienda_de_telegram(bigint) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.tienda_de_telegram(bigint) TO mostrador_app;
-- +goose StatementBegin
CREATE FUNCTION app.tienda_de_codigo_telegram(p_hash text) RETURNS uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	t uuid;
BEGIN
	PERFORM set_config('app.tg_codigo', p_hash, true);
	SELECT tienda_id INTO t FROM telegram_codigos WHERE hash = p_hash;
	PERFORM set_config('app.tg_codigo', '', true);
	RETURN t;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.tienda_de_codigo_telegram(text) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.tienda_de_codigo_telegram(text) TO mostrador_app;

-- +goose StatementBegin
-- Aviso de cada venta al chat vinculado: el trigger encola en la misma transacción que el pedido
-- (por cualquier camino: tienda online, panel, punto de venta, Mercado Libre) y el worker lo manda.
-- Sin chat vinculado no se encola nada.
CREATE FUNCTION app.avisar_venta_telegram() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
BEGIN
	IF EXISTS (SELECT 1 FROM telegram_vinculos WHERE tienda_id = NEW.tienda_id) THEN
		INSERT INTO outbox (tienda_id, tipo, datos)
		VALUES (NEW.tienda_id, 'telegram.venta', jsonb_build_object('numero', NEW.numero));
	END IF;
	RETURN NEW;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.avisar_venta_telegram() FROM PUBLIC;
CREATE TRIGGER avisar_venta_telegram AFTER INSERT ON pedidos
	FOR EACH ROW EXECUTE FUNCTION app.avisar_venta_telegram();

-- +goose Down
DROP TRIGGER avisar_venta_telegram ON pedidos;
DROP FUNCTION app.avisar_venta_telegram();
DROP FUNCTION app.tienda_de_codigo_telegram(text);
DROP FUNCTION app.tienda_de_telegram(bigint);
DROP TABLE telegram_codigos;
DROP TABLE telegram_vinculos;
