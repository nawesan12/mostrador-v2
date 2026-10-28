-- +goose Up
-- Facturación electrónica con ARCA. La clave privada y los tickets se guardan cifrados
-- (plataforma/cifrado, con la clave de cada tienda); el certificado es público.
CREATE TABLE arca_config (
	tienda_id uuid PRIMARY KEY REFERENCES tiendas (id) ON DELETE CASCADE,
	cuit text NOT NULL CHECK (cuit ~ '^\d{11}$'),
	razon_social text NOT NULL CHECK (length(razon_social) BETWEEN 1 AND 64),
	condicion text NOT NULL CHECK (condicion IN ('monotributo', 'responsable_inscripto')),
	punto_venta int NOT NULL CHECK (punto_venta BETWEEN 1 AND 99998),
	ambiente text NOT NULL DEFAULT 'homologacion' CHECK (ambiente IN ('homologacion', 'produccion')),
	clave bytea NOT NULL,
	certificado text CHECK (length(certificado) <= 16384),
	certificado_vence timestamptz,
	facturar_al_cobrar boolean NOT NULL DEFAULT true,
	actualizada_en timestamptz NOT NULL DEFAULT now()
);
SELECT app.aplicar_rls('arca_config');
GRANT SELECT, INSERT, DELETE ON arca_config TO mostrador_app;
GRANT UPDATE (cuit, razon_social, condicion, punto_venta, ambiente, clave, certificado, certificado_vence,
	facturar_al_cobrar, actualizada_en) ON arca_config TO mostrador_app;

-- Ticket de acceso del WSAA (12 h): ARCA no da otro mientras haya uno vigente.
CREATE TABLE arca_tickets (
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	servicio text NOT NULL,
	ticket bytea NOT NULL,
	vence timestamptz NOT NULL,
	PRIMARY KEY (tienda_id, servicio)
);
SELECT app.aplicar_rls('arca_tickets');
GRANT SELECT, INSERT, UPDATE (ticket, vence), DELETE ON arca_tickets TO mostrador_app;

-- Comprobantes. Una venta cobrada tiene siempre su fila: emitido, pendiente, rechazado o "no
-- facturado" con su motivo. Los importes se congelan al encolar (en centavos; neto + iva = total).
CREATE TABLE comprobantes (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	pedido_numero int,
	tipo int NOT NULL CHECK (tipo IN (1, 3, 6, 8, 11, 13)),
	punto_venta int,
	numero int,
	-- El número que se pidió a ARCA antes de saber la respuesta (para no emitir dos veces).
	numero_intentado int,
	cae text CHECK (cae ~ '^\d{14}$'),
	cae_vence date,
	estado text NOT NULL CHECK (estado IN ('pendiente', 'emitido', 'rechazado', 'no_facturado')),
	motivo text CHECK (length(motivo) <= 300),
	codigo_arca text CHECK (length(codigo_arca) <= 20),
	receptor jsonb NOT NULL CHECK (jsonb_typeof(receptor) = 'object' AND octet_length(receptor::text) <= 1000),
	concepto text CHECK (length(concepto) <= 200),
	total bigint NOT NULL CHECK (total > 0),
	neto bigint NOT NULL,
	iva bigint NOT NULL,
	asociado uuid REFERENCES comprobantes (id),
	intentos int NOT NULL DEFAULT 0,
	reintentar_en timestamptz NOT NULL DEFAULT now(),
	creado_en timestamptz NOT NULL DEFAULT now(),
	emitido_en timestamptz,
	CHECK (neto + iva = total * 100),
	CHECK (estado <> 'emitido' OR (numero IS NOT NULL AND cae IS NOT NULL)),
	CHECK (estado <> 'no_facturado' OR motivo IS NOT NULL)
);
CREATE UNIQUE INDEX comprobantes_numero ON comprobantes (tienda_id, tipo, punto_venta, numero) WHERE numero IS NOT NULL;
CREATE UNIQUE INDEX comprobantes_pedido ON comprobantes (tienda_id, pedido_numero) WHERE pedido_numero IS NOT NULL AND tipo IN (1, 6, 11);
-- Una sola nota de crédito por comprobante, aunque lleguen dos a la vez.
CREATE UNIQUE INDEX comprobantes_asociado ON comprobantes (asociado) WHERE asociado IS NOT NULL;
CREATE INDEX comprobantes_cola ON comprobantes (reintentar_en) WHERE estado = 'pendiente';
CREATE INDEX comprobantes_recientes ON comprobantes (tienda_id, creado_en DESC);
SELECT app.aplicar_rls('comprobantes');
GRANT SELECT, INSERT ON comprobantes TO mostrador_app;
GRANT UPDATE (punto_venta, numero, numero_intentado, cae, cae_vence, estado, motivo, codigo_arca, intentos,
	reintentar_en, emitido_en) ON comprobantes TO mostrador_app;

-- La tienda de las facturas pendientes (para el drenador, que pasa por todas): sólo ids.
CREATE POLICY cola ON comprobantes FOR SELECT TO mostrador_migrador
	USING (current_setting('app.cola_facturas', true) = 'si');
-- +goose StatementBegin
CREATE FUNCTION app.tiendas_con_facturas_pendientes(p_limite int) RETURNS SETOF uuid
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
REVOKE ALL ON FUNCTION app.tiendas_con_facturas_pendientes(int) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.tiendas_con_facturas_pendientes(int) TO mostrador_app;

-- Encolar con el cobro: el comprobante se escribe en la misma transacción que marca el pedido como
-- cobrado (sea por Mercado Pago, transferencia confirmada o el punto de venta). Si el pago se
-- guardó, la factura está encolada.
-- +goose StatementBegin
CREATE FUNCTION app.encolar_factura() RETURNS trigger
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app, pg_temp AS $$
DECLARE
	c arca_config%ROWTYPE;
	v_tipo int;
	v_neto bigint;
	v_estado text := 'pendiente';
	v_motivo text;
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
	v_tipo := CASE WHEN c.condicion = 'responsable_inscripto' THEN 6 ELSE 11 END;
	v_neto := CASE WHEN v_tipo = 11 THEN NEW.total * 100 ELSE (NEW.total * 10000 + 60) / 121 END;
	INSERT INTO comprobantes (tienda_id, pedido_numero, tipo, punto_venta, estado, motivo, receptor, total, neto, iva)
	VALUES (NEW.tienda_id, NEW.numero, v_tipo, c.punto_venta, v_estado, v_motivo,
		jsonb_build_object('nombre', coalesce(NEW.comprador->>'nombre', 'Consumidor final'), 'docTipo', 99, 'docNro', 0, 'condicion', 5),
		NEW.total, v_neto, NEW.total * 100 - v_neto)
	ON CONFLICT DO NOTHING;
	RETURN NEW;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.encolar_factura() FROM PUBLIC;
CREATE TRIGGER encolar_factura AFTER INSERT OR UPDATE OF estado_pago ON pedidos
	FOR EACH ROW WHEN (NEW.total > 0) EXECUTE FUNCTION app.encolar_factura();

-- +goose Down
DROP TRIGGER encolar_factura ON pedidos;
DROP FUNCTION app.encolar_factura();
DROP FUNCTION app.tiendas_con_facturas_pendientes(int);
DROP POLICY cola ON comprobantes;
DROP TABLE comprobantes;
DROP TABLE arca_tickets;
DROP TABLE arca_config;
