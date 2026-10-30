-- +goose Up
-- Cada descuento del pedido por separado (cupón, transferencia), para que el detalle del panel y el
-- mail a quien compra cierren: subtotal + envío − descuentos = total.
ALTER TABLE pedidos ADD COLUMN descuentos jsonb NOT NULL DEFAULT '[]';
-- Los pedidos de antes: sólo cuando se sabe de qué fue el descuento. Si hubo cupón y transferencia
-- juntos no se puede separar, y el detalle lo muestra como «Descuento».
UPDATE pedidos p SET descuentos = jsonb_build_array(jsonb_build_object(
		'concepto', 'Cupón ' || p.cupon, 'monto', p.subtotal + p.costo_envio - p.total))
	WHERE p.cupon IS NOT NULL AND p.subtotal + p.costo_envio - p.total > 0
		AND (p.pago <> 'transf' OR NOT EXISTS (SELECT 1 FROM cupones c WHERE c.tienda_id = p.tienda_id
			AND c.codigo = p.cupon AND coalesce((c.opciones->>'combinable')::boolean, false)));
UPDATE pedidos SET descuentos = jsonb_build_array(jsonb_build_object(
		'concepto', 'Descuento por transferencia', 'monto', subtotal + costo_envio - total))
	WHERE cupon IS NULL AND canal = 'web' AND pago = 'transf' AND subtotal + costo_envio - total > 0;
-- La nota de la tienda online sin datos («Tienda online · DNI  · Entrega:  ()»): sólo lo que vino.
UPDATE pedidos SET nota = nullif(btrim(regexp_replace(regexp_replace(regexp_replace(nota,
		'^Tienda online( · )?', ''), 'DNI\s*(·|$)', ''), 'Entrega:\s*\(\)', ''), ' ·'), '')
	WHERE canal = 'web' AND nota LIKE 'Tienda online%';

-- Las promociones automáticas arrancan apagadas: las prende el comerciante. Las tiendas que nunca
-- las tocaron (siguen como venían) quedan apagadas también.
ALTER TABLE tiendas ALTER COLUMN automaticas SET DEFAULT
	'{"transf": false, "envio": false, "cantidad": false, "3x2": false, "cuotas": false, "mayorista": false, "carrito": false, "cumple": false}';
UPDATE tiendas SET automaticas =
	'{"transf": false, "envio": false, "cantidad": false, "3x2": false, "cuotas": false, "mayorista": false, "carrito": false, "cumple": false}'
	WHERE automaticas = '{"transf": true, "envio": true, "cantidad": true, "3x2": false, "cuotas": true, "mayorista": true, "carrito": true, "cumple": false}';

-- El retiro en el local no se despacha: ni pasa por «En camino» ni manda ese mail.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.avisar_comprador() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	v_aviso text;
	v_avisos text[] := '{}';
	pagado constant text[] := ARRAY['Aprobado', 'Acreditada', 'Cobrado'];
BEGIN
	IF coalesce(NEW.comprador->>'email', '') = '' OR NEW.canal = 'pos' THEN
		RETURN NEW;
	END IF;
	IF TG_OP = 'INSERT' THEN
		v_avisos := ARRAY['confirmacion'];
	ELSE
		IF NEW.estado_pago = ANY (pagado) AND NOT (OLD.estado_pago = ANY (pagado)) THEN
			v_avisos := v_avisos || 'pago'::text;
		END IF;
		IF NEW.estado = 'En camino' AND OLD.estado <> 'En camino' AND NEW.envio <> 'retiro' THEN
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

-- +goose Down
ALTER TABLE tiendas ALTER COLUMN automaticas SET DEFAULT
	'{"transf": true, "envio": true, "cantidad": true, "3x2": false, "cuotas": true, "mayorista": true, "carrito": true, "cumple": false}';
ALTER TABLE pedidos DROP COLUMN descuentos;
