-- +goose Up
-- Tiempo real (docs/decisiones.md §42). Los cambios que importan avisan por NOTIFY en el canal
-- `tiempo_real`, que Postgres entrega recién al confirmarse la transacción (si se revierte, no sale
-- nada). La API escucha con una conexión dedicada y reparte por tienda y por rol a los navegadores.
--
-- El aviso dice sólo QUÉ cambió, nunca datos de negocio: {"t": tienda, "k": tipo, "id": id} y,
-- cuando hace falta para rutear, "u" (la persona afectada) o "s" (el id público de una sesión, el
-- mismo prefijo del hash que ya muestra Configuración › Seguridad). Ni montos, ni nombres, ni textos.
--
-- Todos los triggers son por sentencia con tablas de transición: un aumento de precios de 5.000
-- productos o una importación es UNA sentencia y, si toca más de 20 filas de una tienda, sale un
-- solo aviso por lote ({"lote": true}) en vez de uno por fila. Además Postgres junta los avisos
-- idénticos de una misma transacción, y la API vuelve a juntar lo que llega en ráfaga.

-- +goose StatementBegin
CREATE FUNCTION app.avisar_tiempo_real(p_tienda uuid, p_tipo text, p_id text, p_usuario uuid DEFAULT NULL,
	p_extra jsonb DEFAULT NULL) RETURNS void
LANGUAGE plpgsql SET search_path = public, app, pg_temp AS $$
BEGIN
	-- Un aviso mide menos de 200 bytes (el tope de NOTIFY es 8000): el id se corta por las dudas.
	PERFORM pg_notify('tiempo_real', jsonb_strip_nulls(
		jsonb_build_object('t', p_tienda, 'k', p_tipo, 'id', left(p_id, 80), 'u', p_usuario)
		|| coalesce(p_extra, '{}'::jsonb))::text);
END
$$;
-- +goose StatementEnd

-- Trigger genérico por sentencia. Argumentos:
--   0 tipo: 'producto' (le suma .nuevo/.cambio/.borrado según la operación) o fijo ('producto.stock')
--   1 columna del id que viaja en el aviso ('' = sin id)
--   2 columna de la tienda ('tienda_id'; en tiendas, 'id')
--   3 columnas que cuentan como cambio en un UPDATE, separadas por coma ('' = cualquier columna)
--   4 columna de la persona afectada ('' = ninguna)
--   5 columna para emparejar filas viejas y nuevas en un UPDATE ('' = la del id)
-- +goose StatementBegin
CREATE FUNCTION app.tr_tiempo_real() RETURNS trigger
LANGUAGE plpgsql SET search_path = public, app, pg_temp AS $$
DECLARE
	v_tipo text := TG_ARGV[0];
	v_id text := nullif(TG_ARGV[1], '');
	v_tienda text := TG_ARGV[2];
	v_cols text := nullif(TG_ARGV[3], '');
	v_usuario text := nullif(TG_ARGV[4], '');
	v_clave text := coalesce(nullif(TG_ARGV[5], ''), nullif(TG_ARGV[1], ''));
	v_rel text := CASE TG_OP WHEN 'DELETE' THEN 'viejas' ELSE 'nuevas' END;
	v_cruce text := '';
	v_viejo text;
	v_nuevo text;
	r record;
BEGIN
	IF position('.' IN v_tipo) = 0 THEN
		v_tipo := v_tipo || CASE TG_OP WHEN 'INSERT' THEN '.nuevo' WHEN 'DELETE' THEN '.borrado' ELSE '.cambio' END;
	END IF;
	-- En un UPDATE sólo cuentan las filas en que cambió algo de lo que se mira (no el «visto», un
	-- token renovado o un contador interno).
	IF TG_OP = 'UPDATE' THEN
		v_cruce := format(' JOIN viejas o ON o.%1$I IS NOT DISTINCT FROM n.%1$I', v_tienda);
		IF v_clave IS NOT NULL THEN
			v_cruce := v_cruce || format(' AND o.%1$I = n.%1$I', v_clave);
		END IF;
		IF v_cols IS NULL THEN
			v_cruce := v_cruce || ' WHERE o IS DISTINCT FROM n';
		ELSE
			SELECT string_agg('o.' || quote_ident(trim(c)), ', '), string_agg('n.' || quote_ident(trim(c)), ', ')
			INTO v_viejo, v_nuevo FROM unnest(string_to_array(v_cols, ',')) c;
			v_cruce := v_cruce || format(' WHERE (%s) IS DISTINCT FROM (%s)', v_viejo, v_nuevo);
		END IF;
	END IF;
	FOR r IN EXECUTE format($q$
		WITH filas AS (
			SELECT DISTINCT n.%1$I::uuid AS tienda, %2$s AS id, %3$s AS usuario FROM %4$I n %5$s
		), por AS (
			SELECT tienda, count(*) AS cuantas FROM filas WHERE tienda IS NOT NULL GROUP BY tienda
		)
		SELECT f.tienda, f.id, f.usuario, false AS lote FROM filas f JOIN por p USING (tienda) WHERE p.cuantas <= 20
		UNION ALL
		SELECT tienda, NULL, NULL, true FROM por WHERE cuantas > 20
		$q$,
		v_tienda,
		CASE WHEN v_id IS NULL THEN 'NULL::text' ELSE format('n.%I::text', v_id) END,
		CASE WHEN v_usuario IS NULL THEN 'NULL::uuid' ELSE format('n.%I', v_usuario) END,
		v_rel, v_cruce)
	LOOP
		IF r.lote THEN
			PERFORM app.avisar_tiempo_real(r.tienda, v_tipo, NULL, NULL, '{"lote": true}');
		ELSE
			PERFORM app.avisar_tiempo_real(r.tienda, v_tipo, r.id, r.usuario);
		END IF;
	END LOOP;
	RETURN NULL;
END
$$;
-- +goose StatementEnd

-- El pago acreditado de un pedido: la página del pedido de quien compra («Recibimos tu pago»), el
-- detalle en el panel y el Inicio. Mismos estados que los mails a quien compra (00034).
-- +goose StatementBegin
CREATE FUNCTION app.tr_tiempo_real_pago() RETURNS trigger
LANGUAGE plpgsql SET search_path = public, app, pg_temp AS $$
DECLARE
	pagado constant text[] := ARRAY['Aprobado', 'Acreditada', 'Cobrado'];
	r record;
BEGIN
	FOR r IN SELECT n.tienda_id, n.numero FROM nuevas n JOIN viejas o USING (tienda_id, numero)
		WHERE n.estado_pago = ANY (pagado) AND NOT (o.estado_pago = ANY (pagado))
	LOOP
		PERFORM app.avisar_tiempo_real(r.tienda_id, 'pago.acreditado', r.numero::text);
	END LOOP;
	RETURN NULL;
END
$$;
-- +goose StatementEnd

-- Las sesiones son de la persona, no de la tienda: el aviso va por "u". Al cerrarse una sesión
-- vigente (salir, «Cerrar», contraseña nueva, quitar del equipo) o cambiar de tienda, la API corta
-- las conexiones de esa sesión ("s"). Borrar sesiones ya vencidas (limpieza) no avisa nada.
-- +goose StatementBegin
CREATE FUNCTION app.tr_tiempo_real_sesiones() RETURNS trigger
LANGUAGE plpgsql SET search_path = public, app, auth, pg_temp AS $$
DECLARE
	r record;
BEGIN
	IF TG_OP = 'INSERT' THEN
		FOR r IN SELECT DISTINCT usuario_id FROM nuevas LOOP
			PERFORM app.avisar_tiempo_real(NULL, 'sesion.cambio', NULL, r.usuario_id);
		END LOOP;
	ELSIF TG_OP = 'DELETE' THEN
		FOR r IN SELECT usuario_id, tienda_id, left(token_hash, 16) AS s FROM viejas WHERE vence_en > now() LOOP
			PERFORM app.avisar_tiempo_real(r.tienda_id, 'sesion.cerrada', NULL, r.usuario_id, jsonb_build_object('s', r.s));
		END LOOP;
	ELSE
		FOR r IN SELECT n.usuario_id, n.tienda_id, left(n.token_hash, 16) AS s
			FROM nuevas n JOIN viejas o USING (token_hash) WHERE n.tienda_id IS DISTINCT FROM o.tienda_id
		LOOP
			PERFORM app.avisar_tiempo_real(r.tienda_id, 'sesion.tienda', NULL, r.usuario_id, jsonb_build_object('s', r.s));
		END LOOP;
	END IF;
	RETURN NULL;
END
$$;
-- +goose StatementEnd

-- Un trigger por operación (las tablas de transición no admiten varias en uno).
-- +goose StatementBegin
CREATE FUNCTION app.crear_tiempo_real(p_tabla regclass, p_ops text[], p_nombre text, VARIADIC p_args text[])
RETURNS void LANGUAGE plpgsql AS $$
DECLARE
	v_op text;
	v_args text;
BEGIN
	SELECT string_agg(quote_literal(a), ', ') INTO v_args FROM unnest(p_args) a;
	FOREACH v_op IN ARRAY p_ops LOOP
		EXECUTE format('CREATE TRIGGER %I AFTER %s ON %s REFERENCING %s FOR EACH STATEMENT EXECUTE FUNCTION app.tr_tiempo_real(%s)',
			'tiempo_real_' || p_nombre || '_' || lower(left(v_op, 3)), v_op, p_tabla,
			CASE v_op WHEN 'INSERT' THEN 'NEW TABLE AS nuevas' WHEN 'DELETE' THEN 'OLD TABLE AS viejas'
				ELSE 'OLD TABLE AS viejas NEW TABLE AS nuevas' END,
			v_args);
	END LOOP;
END
$$;
-- +goose StatementEnd

-- Pedidos: nuevo, cambio (estado, pago, envío, factura… no los avisos internos) y borrado.
SELECT app.crear_tiempo_real('pedidos', '{INSERT}', 'pedido', 'pedido.nuevo', 'numero', 'tienda_id', '', '', '');
SELECT app.crear_tiempo_real('pedidos', '{UPDATE}', 'pedido', 'pedido.cambio', 'numero', 'tienda_id',
	'estado,estado_pago,pago,envio,ciudad_envio,seguimiento,facturado,factura,items,subtotal,costo_envio,total,nota,entrega,comprador,vip,descuentos,cupon,recordado_en', '', '');
SELECT app.crear_tiempo_real('pedidos', '{DELETE}', 'pedido', 'pedido', 'numero', 'tienda_id', '', '', '');
CREATE TRIGGER tiempo_real_pago AFTER UPDATE ON pedidos REFERENCING OLD TABLE AS viejas NEW TABLE AS nuevas
	FOR EACH STATEMENT EXECUTE FUNCTION app.tr_tiempo_real_pago();

-- Productos y variantes: el stock (y si está publicado) por separado del resto, que es lo que mira
-- la tienda online y el punto de venta.
SELECT app.crear_tiempo_real('productos', '{INSERT,DELETE}', 'producto', 'producto', 'id', 'tienda_id', '', '', '');
SELECT app.crear_tiempo_real('productos', '{UPDATE}', 'producto', 'producto.cambio', 'id', 'tienda_id',
	'nombre,descripcion,categoria,precio,costo,sku,foto,canales,opciones,fotos,slug', '', '');
SELECT app.crear_tiempo_real('productos', '{UPDATE}', 'producto_stock', 'producto.stock', 'id', 'tienda_id', 'stock,publicado', '', '');
SELECT app.crear_tiempo_real('variantes', '{INSERT,DELETE}', 'variante', 'producto.cambio', 'producto_id', 'tienda_id', '', '', '');
SELECT app.crear_tiempo_real('variantes', '{UPDATE}', 'variante', 'producto.cambio', 'producto_id', 'tienda_id',
	'valores,sku,precio,foto', '', 'id');
SELECT app.crear_tiempo_real('variantes', '{UPDATE}', 'variante_stock', 'producto.stock', 'producto_id', 'tienda_id',
	'stock,activa', '', 'id');

-- Clientes, carritos abandonados, cupones, envíos de Correo, comprobantes de ARCA, reseñas y
-- preguntas (las de la tienda y las de Mercado Libre), turnos y devoluciones.
SELECT app.crear_tiempo_real('clientes', '{INSERT,UPDATE,DELETE}', 'cliente', 'cliente', 'id', 'tienda_id', '', '', '');
SELECT app.crear_tiempo_real('carritos', '{INSERT,UPDATE,DELETE}', 'carrito', 'carrito', 'id', 'tienda_id',
	'email,nombre,items,total,recordado_en,recuperado_en', '', '');
SELECT app.crear_tiempo_real('cupones', '{INSERT,UPDATE,DELETE}', 'cupon', 'cupon', 'id', 'tienda_id', '', '', '');
SELECT app.crear_tiempo_real('ca_envios', '{INSERT,UPDATE,DELETE}', 'envio', 'envio', 'pedido', 'tienda_id', '', '', '');
SELECT app.crear_tiempo_real('comprobantes', '{INSERT,UPDATE}', 'comprobante', 'comprobante', 'id', 'tienda_id',
	'estado,cae,numero,motivo,emitido_en', '', '');
SELECT app.crear_tiempo_real('resenas', '{INSERT,UPDATE,DELETE}', 'resena', 'resena', 'id', 'tienda_id', '', '', '');
SELECT app.crear_tiempo_real('preguntas', '{INSERT,UPDATE,DELETE}', 'pregunta', 'pregunta', 'id', 'tienda_id', '', '', '');
SELECT app.crear_tiempo_real('ml_preguntas', '{INSERT,UPDATE,DELETE}', 'pregunta_ml', 'pregunta', 'pregunta_id', 'tienda_id', '', '', '');
SELECT app.crear_tiempo_real('turnos', '{INSERT,UPDATE,DELETE}', 'turno', 'turno', 'id', 'tienda_id', '', '', '');
SELECT app.crear_tiempo_real('devoluciones', '{INSERT,UPDATE,DELETE}', 'devolucion', 'devolucion', 'id', 'tienda_id', '', '', '');

-- Integraciones conectadas o desconectadas (también cuando Mercado Pago o Mercado Libre piden
-- reautorizar). Renovar el token cada pocas horas no cuenta como cambio.
SELECT app.crear_tiempo_real('mp_conexiones', '{INSERT,UPDATE,DELETE}', 'mp', 'integracion.cambio', '', 'tienda_id',
	'mp_usuario,alcances,conectada_en,reautorizar,sucursal_mp,caja_externa', '', '');
SELECT app.crear_tiempo_real('ml_conexiones', '{INSERT,UPDATE,DELETE}', 'ml', 'integracion.cambio', '', 'tienda_id',
	'ml_usuario,apodo,conectada_en,reautorizar,recargo,pausar_sin_stock', '', '');
SELECT app.crear_tiempo_real('ca_conexiones', '{INSERT,UPDATE,DELETE}', 'ca', 'integracion.cambio', '', 'tienda_id',
	'cliente,email,origen,paquete,conectada_en', '', '');
SELECT app.crear_tiempo_real('arca_config', '{INSERT,UPDATE,DELETE}', 'arca', 'integracion.cambio', '', 'tienda_id',
	'cuit,razon_social,condicion,punto_venta,ambiente,certificado,certificado_vence,facturar_al_cobrar', '', '');
SELECT app.crear_tiempo_real('telegram_vinculos', '{INSERT,UPDATE,DELETE}', 'telegram', 'integracion.cambio', '', 'tienda_id',
	'chat,nombre,vinculado_en', '', '');

-- Cobros (Mercado Pago, links de pago): sólo para quien ve la plata (lo filtra la API).
SELECT app.crear_tiempo_real('mp_cobros', '{INSERT,UPDATE}', 'mp_cobro', 'finanzas.cambio', '', 'tienda_id', 'estado,pago', '', 'referencia');
SELECT app.crear_tiempo_real('links', '{INSERT,UPDATE,DELETE}', 'link', 'finanzas.cambio', '', 'tienda_id', 'pagado_en,monto,concepto', '', 'codigo');

-- Importación del catálogo: el avance (cada paso del worker).
SELECT app.crear_tiempo_real('importaciones', '{INSERT,UPDATE}', 'importacion', 'importacion', 'id', 'tienda_id',
	'estado,etapa,progreso,errores,terminada', '', '');

-- Actividad: un registro nuevo o deshecho.
SELECT app.crear_tiempo_real('auditoria', '{INSERT}', 'actividad', 'actividad', 'id', 'tienda_id', '', '', '');
SELECT app.crear_tiempo_real('auditoria', '{UPDATE}', 'actividad', 'actividad.cambio', 'id', 'tienda_id', 'deshecho_en', '', '');

-- La tienda: datos generales, el diseño publicado de Mi tienda y el plan (pago del abono incluido).
-- El borrador del diseño no avisa: lo guarda el editor mientras se escribe.
SELECT app.crear_tiempo_real('tiendas', '{UPDATE}', 'tienda', 'tienda.cambio', '', 'id',
	'nombre,slug,plantilla,rubro,negocio,preferencias,canales_pausados,meta_mensual,automaticas,carritos_automaticos,tarjeta,ga4,pixel_meta,feeds_cargados,ia_permisos,onboarding_terminado_en,pausada_en,cerrada_en', '', '');
SELECT app.crear_tiempo_real('tiendas', '{UPDATE}', 'diseno', 'mitienda.publicada', '', 'id', 'diseno_publicado', '', '');
SELECT app.crear_tiempo_real('tiendas', '{UPDATE}', 'plan', 'plan.cambio', '', 'id',
	'plan,plan_ciclo,suscripcion_mp,suscripcion_estado,plan_pendiente,ciclo_pendiente,packs,credito,prueba_hasta,cancelada_en', '', '');
SELECT app.crear_tiempo_real('entregas', '{INSERT,UPDATE}', 'entregas', 'tienda.cambio', '', 'tienda_id', '', '', '');
SELECT app.crear_tiempo_real('dominios', '{INSERT,UPDATE,DELETE}', 'dominio', 'dominio', 'nombre', 'tienda_id', '', '', '');

-- Visitas (lotes de la web cada 5 s; la API además las espacia) y tareas de «Para hacer hoy».
SELECT app.crear_tiempo_real('visitas', '{INSERT,UPDATE}', 'visitas', 'visitas.cambio', '', 'tienda_id', '', '', '');
SELECT app.crear_tiempo_real('tareas_hechas', '{INSERT,DELETE}', 'tarea', 'tarea.cambio', '', 'tienda_id', '', '', '');

-- El equipo: membresías (con la persona afectada, para refrescarle el panel si le cambian o le
-- quitan el rol) e invitaciones.
SELECT app.crear_tiempo_real('auth.membresias', '{INSERT,UPDATE,DELETE}', 'membresia', 'equipo.cambio', 'usuario_id', 'tienda_id', '',
	'usuario_id', '');
SELECT app.crear_tiempo_real('invitaciones', '{INSERT,UPDATE,DELETE}', 'invitacion', 'equipo.cambio', '', 'tienda_id', '', '', 'id');

CREATE TRIGGER tiempo_real_sesion_ins AFTER INSERT ON auth.sesiones REFERENCING NEW TABLE AS nuevas
	FOR EACH STATEMENT EXECUTE FUNCTION app.tr_tiempo_real_sesiones();
CREATE TRIGGER tiempo_real_sesion_upd AFTER UPDATE ON auth.sesiones REFERENCING OLD TABLE AS viejas NEW TABLE AS nuevas
	FOR EACH STATEMENT EXECUTE FUNCTION app.tr_tiempo_real_sesiones();
CREATE TRIGGER tiempo_real_sesion_del AFTER DELETE ON auth.sesiones REFERENCING OLD TABLE AS viejas
	FOR EACH STATEMENT EXECUTE FUNCTION app.tr_tiempo_real_sesiones();

DROP FUNCTION app.crear_tiempo_real(regclass, text[], text, text[]);

-- +goose Down
-- +goose StatementBegin
DO $$
DECLARE
	r record;
BEGIN
	FOR r IN SELECT tgname, tgrelid::regclass AS tabla FROM pg_trigger WHERE tgname LIKE 'tiempo\_real\_%' AND NOT tgisinternal LOOP
		EXECUTE format('DROP TRIGGER %I ON %s', r.tgname, r.tabla);
	END LOOP;
END
$$;
-- +goose StatementEnd
DROP FUNCTION app.tr_tiempo_real_sesiones();
DROP FUNCTION app.tr_tiempo_real_pago();
DROP FUNCTION app.tr_tiempo_real();
DROP FUNCTION app.avisar_tiempo_real(uuid, text, text, uuid, jsonb);
