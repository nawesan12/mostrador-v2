-- +goose Up
-- Un pago de Mercado Pago que llega con el pedido ya cancelado no se factura: el comprobante se
-- encola sólo si el pedido sigue en pie, y se encola al reabrirlo si el pago ya estaba aprobado.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.encolar_factura() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	c arca_config%ROWTYPE;
	v_tipo int;
	v_neto bigint;
	v_estado text := 'pendiente';
	v_motivo text;
	v_receptor jsonb;
BEGIN
	-- Un pedido cancelado no es una venta: el pago que llega tarde lo devuelve la tienda (aviso
	-- pago.tarde). Si después lo reabre con el pago ya aprobado, ahí se encola.
	IF NEW.estado_pago NOT IN ('Aprobado', 'Acreditada', 'Cobrado') OR NEW.estado = 'Cancelado'
		OR (TG_OP = 'UPDATE' AND OLD.estado_pago IN ('Aprobado', 'Acreditada', 'Cobrado') AND OLD.estado <> 'Cancelado') THEN
		RETURN NEW;
	END IF;
	SELECT * INTO c FROM arca_config WHERE tienda_id = NEW.tienda_id;
	IF NOT FOUND OR c.certificado IS NULL THEN
		v_estado := 'no_facturado';
		v_motivo := 'La tienda todavía no conectó ARCA.';
	ELSIF NOT c.facturar_al_cobrar THEN
		v_estado := 'no_facturado';
		v_motivo := 'Facturar al cobrar está apagado: se factura a mano.';
	END IF;
	v_receptor := jsonb_build_object('nombre', coalesce(NEW.comprador->>'nombre', 'Consumidor final'), 'docTipo', 99, 'docNro', 0, 'condicion', 5);
	v_tipo := CASE WHEN c.condicion = 'responsable_inscripto' THEN 6 ELSE 11 END;
	-- Pidió factura A en el checkout (la vidriera validó el CUIT): a su nombre, como responsable inscripta.
	IF v_tipo = 6 AND NEW.comprador->'factura'->>'cuit' ~ '^\d{11}$' THEN
		v_tipo := 1;
		v_receptor := jsonb_build_object('nombre', left(coalesce(nullif(NEW.comprador->'factura'->>'razonSocial', ''), NEW.comprador->>'nombre'), 120),
			'docTipo', 80, 'docNro', NEW.comprador->'factura'->>'cuit', 'condicion', 1);
	END IF;
	v_neto := CASE WHEN v_tipo = 11 THEN NEW.total * 100 ELSE (NEW.total * 10000 + 60) / 121 END;
	INSERT INTO comprobantes (tienda_id, pedido_numero, tipo, punto_venta, estado, motivo, receptor, total, neto, iva)
	VALUES (NEW.tienda_id, NEW.numero, v_tipo, c.punto_venta, v_estado, v_motivo,
		v_receptor,
		NEW.total, v_neto, NEW.total * 100 - v_neto)
	ON CONFLICT DO NOTHING;
	RETURN NEW;
END
$$;
-- +goose StatementEnd

DROP TRIGGER encolar_factura ON pedidos;
CREATE TRIGGER encolar_factura AFTER INSERT OR UPDATE OF estado_pago, estado ON pedidos
	FOR EACH ROW WHEN (NEW.total > 0) EXECUTE FUNCTION app.encolar_factura();

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.encolar_factura() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	c arca_config%ROWTYPE;
	v_tipo int;
	v_neto bigint;
	v_estado text := 'pendiente';
	v_motivo text;
	v_receptor jsonb;
BEGIN
	IF NEW.estado_pago NOT IN ('Aprobado', 'Acreditada', 'Cobrado')
		OR (TG_OP = 'UPDATE' AND OLD.estado_pago IN ('Aprobado', 'Acreditada', 'Cobrado')) THEN
		RETURN NEW;
	END IF;
	SELECT * INTO c FROM arca_config WHERE tienda_id = NEW.tienda_id;
	IF NOT FOUND OR c.certificado IS NULL THEN
		v_estado := 'no_facturado';
		v_motivo := 'La tienda todavía no conectó ARCA.';
	ELSIF NOT c.facturar_al_cobrar THEN
		v_estado := 'no_facturado';
		v_motivo := 'Facturar al cobrar está apagado: se factura a mano.';
	END IF;
	v_receptor := jsonb_build_object('nombre', coalesce(NEW.comprador->>'nombre', 'Consumidor final'), 'docTipo', 99, 'docNro', 0, 'condicion', 5);
	v_tipo := CASE WHEN c.condicion = 'responsable_inscripto' THEN 6 ELSE 11 END;
	-- Pidió factura A en el checkout (la vidriera validó el CUIT): a su nombre, como responsable inscripta.
	IF v_tipo = 6 AND NEW.comprador->'factura'->>'cuit' ~ '^\d{11}$' THEN
		v_tipo := 1;
		v_receptor := jsonb_build_object('nombre', left(coalesce(nullif(NEW.comprador->'factura'->>'razonSocial', ''), NEW.comprador->>'nombre'), 120),
			'docTipo', 80, 'docNro', NEW.comprador->'factura'->>'cuit', 'condicion', 1);
	END IF;
	v_neto := CASE WHEN v_tipo = 11 THEN NEW.total * 100 ELSE (NEW.total * 10000 + 60) / 121 END;
	INSERT INTO comprobantes (tienda_id, pedido_numero, tipo, punto_venta, estado, motivo, receptor, total, neto, iva)
	VALUES (NEW.tienda_id, NEW.numero, v_tipo, c.punto_venta, v_estado, v_motivo,
		v_receptor,
		NEW.total, v_neto, NEW.total * 100 - v_neto)
	ON CONFLICT DO NOTHING;
	RETURN NEW;
END
$$;
-- +goose StatementEnd

DROP TRIGGER encolar_factura ON pedidos;
CREATE TRIGGER encolar_factura AFTER INSERT OR UPDATE OF estado_pago ON pedidos
	FOR EACH ROW WHEN (NEW.total > 0) EXECUTE FUNCTION app.encolar_factura();
