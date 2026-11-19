-- +goose Up
-- Los planes en lo que corre solo (19/11/2026). La regla de qué incluye cada plan vive en Go
-- (dominio/planes); acá hay una copia mínima para lo que la base decide sin pasar por Go: encolar
-- facturas, armar las colas del worker y el sitemap. Un test de Go (planes.TestIgualQueLaBase)
-- compara esta función con planes.Todos para cada plan, función y estado de la prueba.
--
-- p_fin_prueba es coalesce(prueba_hasta, creada_en + 14 días): durante la prueba, Gratis rige como
-- Crece (el id «pro»).
-- +goose StatementBegin
CREATE FUNCTION app.plan_incluye(p_plan text, p_fin_prueba timestamptz, p_funcion text) RETURNS boolean
LANGUAGE sql STABLE SET search_path = public, app AS $$
	SELECT CASE CASE WHEN p_plan = 'gratis' AND now() < p_fin_prueba THEN 'pro' ELSE p_plan END
		WHEN 'emprende' THEN p_funcion IN ('dominio', 'arca', 'correo', 'cupones')
		WHEN 'pro' THEN p_funcion IN ('dominio', 'arca', 'correo', 'cupones', 'roles', 'local', 'mercadolibre', 'newsletter')
		WHEN 'negocio' THEN p_funcion IN ('dominio', 'arca', 'correo', 'cupones', 'roles', 'local', 'mercadolibre', 'newsletter', 'mayorista')
		ELSE false
	END
$$;
-- +goose StatementEnd
GRANT EXECUTE ON FUNCTION app.plan_incluye(text, timestamptz, text) TO mostrador_app, mostrador_worker;

-- Negocio permite 10 fotos por producto (antes la galería llegaba a 8).
ALTER TABLE productos DROP CONSTRAINT productos_fotos_check;
ALTER TABLE productos ADD CONSTRAINT productos_fotos_check CHECK (cardinality(fotos) <= 10 AND array_position(fotos, NULL) IS NULL);

-- Facturar con ARCA es desde Emprende: sin el plan, el comprobante nace «no facturado» (a la vista
-- en el panel, se puede emitir a mano después de subir), como cuando falta el certificado. Así no
-- se acumulan pendientes que se emitirían de golpe y fuera de fecha al volver a subir.
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
	v_con_plan boolean;
BEGIN
	-- Un pedido cancelado no es una venta: el pago que llega tarde lo devuelve la tienda (aviso
	-- pago.tarde). Si después lo reabre con el pago ya aprobado, ahí se encola.
	IF NEW.estado_pago NOT IN ('Aprobado', 'Acreditada', 'Cobrado') OR NEW.estado = 'Cancelado'
		OR (TG_OP = 'UPDATE' AND OLD.estado_pago IN ('Aprobado', 'Acreditada', 'Cobrado') AND OLD.estado <> 'Cancelado') THEN
		RETURN NEW;
	END IF;
	SELECT app.plan_incluye(t.plan, coalesce(t.prueba_hasta, t.creada_en + interval '14 days'), 'arca') INTO v_con_plan
		FROM tiendas t WHERE t.id = NEW.tienda_id;
	SELECT * INTO c FROM arca_config WHERE tienda_id = NEW.tienda_id;
	IF NOT FOUND OR c.certificado IS NULL THEN
		v_estado := 'no_facturado';
		v_motivo := 'La tienda todavía no conectó ARCA.';
	ELSIF v_con_plan IS NOT TRUE THEN
		v_estado := 'no_facturado';
		v_motivo := 'Tu plan no incluye la factura con ARCA: se factura a mano.';
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

-- La cola de facturas: primero la tienda que espera hace más, así ninguna traba a las demás. Las
-- tiendas sin ARCA en el plan siguen entrando: el drenador pasa sus pendientes de antes de bajar a
-- «no facturado» en la primera vuelta y salen de la cola.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.tiendas_con_facturas_pendientes(p_limite int) RETURNS SETOF uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
BEGIN
	PERFORM set_config('app.cola_facturas', 'si', true);
	RETURN QUERY SELECT c.tienda_id FROM comprobantes c
		WHERE c.estado = 'pendiente' AND c.reintentar_en <= now() AND c.intentos < 8
		GROUP BY c.tienda_id
		ORDER BY min(c.reintentar_en)
		LIMIT p_limite;
	PERFORM set_config('app.cola_facturas', '', true);
END
$$;
-- +goose StatementEnd

-- Los recordatorios de carritos son de los cupones/combos/carritos (desde Emprende): los carritos de
-- una tienda sin el plan no ocupan lugar en la tanda de las demás.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.carritos_para_recordar(p_limite int) RETURNS TABLE (tienda_id uuid, id uuid)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
BEGIN
	PERFORM set_config('app.carritos_recordar', 'si', true);
	RETURN QUERY SELECT c.tienda_id, c.id FROM carritos c JOIN tiendas t ON t.id = c.tienda_id
		WHERE c.recuperado_en IS NULL AND c.baja_en IS NULL AND c.paso < 3
			AND c.actualizado > now() - interval '8 days'
			AND coalesce((t.carritos_recordatorios->>'activo')::boolean, false)
			AND app.plan_incluye(t.plan, coalesce(t.prueba_hasta, t.creada_en + interval '14 days'), 'cupones')
			AND EXISTS (SELECT 1 FROM jsonb_array_elements(t.carritos_recordatorios->'pasos') WITH ORDINALITY AS p (x, i)
				WHERE p.i > c.paso AND coalesce((p.x->>'activo')::boolean, false)
					AND c.actualizado <= now() - make_interval(hours => coalesce((p.x->>'horas')::int, 72)))
		ORDER BY c.actualizado LIMIT p_limite;
	PERFORM set_config('app.carritos_recordar', '', true);
END
$$;
-- +goose StatementEnd

-- El sitemap: el dominio propio sólo si el plan lo incluye (sin él, la tienda vive en su subdominio).
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.sitemap(p_limite int) RETURNS TABLE (tienda text, dominio text, producto text, actualizado timestamptz)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
BEGIN
	PERFORM set_config('app.sitemap', 'si', true);
	RETURN QUERY
		SELECT t.slug, CASE WHEN app.plan_incluye(t.plan, coalesce(t.prueba_hasta, t.creada_en + interval '14 days'), 'dominio')
			THEN d.nombre END, NULL::text, t.creada_en
		FROM tiendas t LEFT JOIN dominios d ON d.tienda_id = t.id AND d.principal AND d.verificado_en IS NOT NULL
		WHERE t.cerrada_en IS NULL
		UNION ALL
		SELECT t.slug, CASE WHEN app.plan_incluye(t.plan, coalesce(t.prueba_hasta, t.creada_en + interval '14 days'), 'dominio')
			THEN d.nombre END, p.slug, p.creado_en
		FROM productos p JOIN tiendas t ON t.id = p.tienda_id
			LEFT JOIN dominios d ON d.tienda_id = t.id AND d.principal AND d.verificado_en IS NOT NULL
		WHERE t.cerrada_en IS NULL AND p.publicado AND p.slug IS NOT NULL
		LIMIT p_limite;
	PERFORM set_config('app.sitemap', '', true);
END
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.sitemap(p_limite int) RETURNS TABLE (tienda text, dominio text, producto text, actualizado timestamptz)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
BEGIN
	PERFORM set_config('app.sitemap', 'si', true);
	RETURN QUERY
		SELECT t.slug, d.nombre, NULL::text, t.creada_en
		FROM tiendas t LEFT JOIN dominios d ON d.tienda_id = t.id AND d.principal AND d.verificado_en IS NOT NULL
		WHERE t.cerrada_en IS NULL
		UNION ALL
		SELECT t.slug, d.nombre, p.slug, p.creado_en
		FROM productos p JOIN tiendas t ON t.id = p.tienda_id
			LEFT JOIN dominios d ON d.tienda_id = t.id AND d.principal AND d.verificado_en IS NOT NULL
		WHERE t.cerrada_en IS NULL AND p.publicado AND p.slug IS NOT NULL
		LIMIT p_limite;
	PERFORM set_config('app.sitemap', '', true);
END
$$;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.carritos_para_recordar(p_limite int) RETURNS TABLE (tienda_id uuid, id uuid)
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
BEGIN
	PERFORM set_config('app.carritos_recordar', 'si', true);
	RETURN QUERY SELECT c.tienda_id, c.id FROM carritos c JOIN tiendas t ON t.id = c.tienda_id
		WHERE c.recuperado_en IS NULL AND c.baja_en IS NULL AND c.paso < 3
			AND c.actualizado > now() - interval '8 days'
			AND coalesce((t.carritos_recordatorios->>'activo')::boolean, false)
			AND EXISTS (SELECT 1 FROM jsonb_array_elements(t.carritos_recordatorios->'pasos') WITH ORDINALITY AS p (x, i)
				WHERE p.i > c.paso AND coalesce((p.x->>'activo')::boolean, false)
					AND c.actualizado <= now() - make_interval(hours => coalesce((p.x->>'horas')::int, 72)))
		ORDER BY c.actualizado LIMIT p_limite;
	PERFORM set_config('app.carritos_recordar', '', true);
END
$$;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION app.tiendas_con_facturas_pendientes(p_limite int) RETURNS SETOF uuid
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
BEGIN
	PERFORM set_config('app.cola_facturas', 'si', true);
	RETURN QUERY SELECT DISTINCT c.tienda_id FROM comprobantes c
		WHERE c.estado = 'pendiente' AND c.reintentar_en <= now() AND c.intentos < 8
		LIMIT p_limite;
	PERFORM set_config('app.cola_facturas', '', true);
END
$$;
-- +goose StatementEnd
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
ALTER TABLE productos DROP CONSTRAINT productos_fotos_check;
ALTER TABLE productos ADD CONSTRAINT productos_fotos_check CHECK (cardinality(fotos) <= 8 AND array_position(fotos, NULL) IS NULL);
DROP FUNCTION app.plan_incluye(text, timestamptz, text);
