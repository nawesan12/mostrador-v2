-- +goose Up
-- Mail «Nueva venta #N» a la dueña (docs/decisiones.md §35). Lo encola un trigger en la misma
-- transacción que el pedido, cuando la venta de la tienda online queda confirmada según el medio:
-- por transferencia, al crearse (pendiente de comprobante); por Mercado Pago, cuando se acredita.
-- El worker mira la preferencia (Notificaciones › Venta nueva › Email) y lo manda una sola vez
-- (pedidos.avisos->'tienda_venta').

-- +goose StatementBegin
CREATE FUNCTION app.avisar_venta_tienda() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	pagado constant text[] := ARRAY['Aprobado', 'Acreditada', 'Cobrado'];
BEGIN
	IF NEW.canal <> 'web' THEN
		RETURN NEW;
	END IF;
	IF (TG_OP = 'INSERT' AND (NEW.pago <> 'mp' OR NEW.estado_pago = ANY (pagado)))
		OR (TG_OP = 'UPDATE' AND NEW.pago = 'mp' AND NEW.estado_pago = ANY (pagado)
			AND NOT (OLD.estado_pago = ANY (pagado))) THEN
		INSERT INTO outbox (tienda_id, tipo, datos)
		VALUES (NEW.tienda_id, 'tienda.venta', jsonb_build_object('numero', NEW.numero));
	END IF;
	RETURN NEW;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.avisar_venta_tienda() FROM PUBLIC;
CREATE TRIGGER avisar_venta_tienda AFTER INSERT OR UPDATE OF estado_pago ON pedidos
	FOR EACH ROW EXECUTE FUNCTION app.avisar_venta_tienda();

-- +goose Down
DROP TRIGGER avisar_venta_tienda ON pedidos;
DROP FUNCTION app.avisar_venta_tienda();
DELETE FROM outbox WHERE tipo = 'tienda.venta';
