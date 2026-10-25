-- +goose Up
-- Mails a quien compra: la confirmación, el pago recibido, el envío en camino y el retiro listo.
-- Un trigger los encola en el outbox en la misma transacción que el cambio (por cualquier camino:
-- tienda online, panel, lote, Mercado Pago) y el worker los manda. `avisos` guarda cuáles salieron:
-- la entrega del outbox es al menos una vez, el mail sale una sola.
ALTER TABLE pedidos ADD COLUMN avisos jsonb NOT NULL DEFAULT '{}';
GRANT UPDATE (avisos) ON pedidos TO mostrador_app;

-- +goose StatementBegin
CREATE FUNCTION app.avisar_comprador() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	v_aviso text;
	v_avisos text[] := '{}';
	pagado constant text[] := ARRAY['Aprobado', 'Acreditada', 'Cobrado'];
BEGIN
	-- Sólo pedidos con email (los de la tienda online) y que no son del punto de venta.
	IF coalesce(NEW.comprador->>'email', '') = '' OR NEW.canal = 'pos' THEN
		RETURN NEW;
	END IF;
	-- Cada transición por separado: un mismo cambio puede cobrar y despachar a la vez.
	IF TG_OP = 'INSERT' THEN
		v_avisos := ARRAY['confirmacion'];
	ELSE
		IF NEW.estado_pago = ANY (pagado) AND NOT (OLD.estado_pago = ANY (pagado)) THEN
			v_avisos := v_avisos || 'pago'::text;
		END IF;
		IF NEW.estado = 'En camino' AND OLD.estado <> 'En camino' THEN
			v_avisos := v_avisos || 'en_camino'::text;
		END IF;
		IF NEW.estado = 'Listo para despachar' AND OLD.estado <> 'Listo para despachar' AND NEW.envio = 'retiro' THEN
			v_avisos := v_avisos || 'retiro'::text;
		END IF;
	END IF;
	FOREACH v_aviso IN ARRAY v_avisos LOOP
		INSERT INTO outbox (tienda_id, tipo, datos)
		VALUES (NEW.tienda_id, 'comprador.aviso', jsonb_build_object('numero', NEW.numero, 'aviso', v_aviso));
	END LOOP;
	RETURN NEW;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.avisar_comprador() FROM PUBLIC;
CREATE TRIGGER avisar_comprador AFTER INSERT OR UPDATE OF estado, estado_pago ON pedidos
	FOR EACH ROW EXECUTE FUNCTION app.avisar_comprador();

-- +goose Down
DROP TRIGGER avisar_comprador ON pedidos;
DROP FUNCTION app.avisar_comprador();
ALTER TABLE pedidos DROP COLUMN avisos;
