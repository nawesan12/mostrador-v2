-- +goose Up
-- Visitas de la tienda online, medidas por Tu Local Web sin cookies ni datos personales
-- (decisiones §36). Sólo se guarda el agregado por día: nunca una fila por visita.
--
-- visitas: una fila por tienda, día (hora de Argentina) y página. `pagina = 'tienda'` es el total
-- del día (todas las páginas; `visitantes` = personas distintas en la tienda). Las de producto
-- llevan el producto; las demás, no. Crece a lo sumo (6 + productos vistos) filas por tienda y día,
-- y se borra a los 400 días (app.purgar_visitas).
CREATE TABLE visitas (
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	dia date NOT NULL,
	pagina text NOT NULL CHECK (pagina IN ('tienda', 'inicio', 'catalogo', 'producto', 'carrito', 'checkout')),
	producto_id uuid,
	vistas integer NOT NULL DEFAULT 0 CHECK (vistas >= 0),
	visitantes integer NOT NULL DEFAULT 0 CHECK (visitantes >= 0),
	CHECK ((pagina = 'producto') = (producto_id IS NOT NULL)),
	FOREIGN KEY (tienda_id, producto_id) REFERENCES productos (tienda_id, id) ON DELETE CASCADE,
	UNIQUE NULLS NOT DISTINCT (tienda_id, dia, pagina, producto_id)
);
SELECT app.aplicar_rls('visitas');
GRANT SELECT, INSERT, UPDATE ON visitas TO mostrador_app;

-- visitas_vistos: quién ya contó hoy, para los visitantes únicos. `visitante` es un hash de IP,
-- navegador, tienda y una sal del día que vive sólo en la memoria de la web y no se guarda: con la
-- sal perdida no se puede volver a la persona. `clave` es la página ('tienda', 'inicio', …,
-- 'producto:<id>'). Sólo vale el día en curso: lo de días anteriores se borra (app.purgar_visitas).
-- UNLOGGED: si Postgres se cae, se pierde a lo sumo el conteo de únicos del día, no las visitas.
CREATE UNLOGGED TABLE visitas_vistos (
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	dia date NOT NULL,
	clave text NOT NULL CHECK (length(clave) <= 60),
	visitante bytea NOT NULL CHECK (length(visitante) = 16),
	PRIMARY KEY (tienda_id, dia, clave, visitante)
);
SELECT app.aplicar_rls('visitas_vistos');
GRANT SELECT, INSERT ON visitas_vistos TO mostrador_app;

-- La limpieza ve todas las tiendas sólo dentro de app.purgar_visitas (corre como el migrador y
-- prende app.purga_visitas en su transacción), como las funciones de partners.
CREATE POLICY purga ON visitas TO mostrador_migrador
	USING (current_setting('app.purga_visitas', true) = 'si');
CREATE POLICY purga ON visitas_vistos TO mostrador_migrador
	USING (current_setting('app.purga_visitas', true) = 'si');

-- +goose StatementBegin
-- Limpieza diaria (la corre el worker con el rol de la app): los vistos de días anteriores y las
-- visitas de hace más de 400 días, de todas las tiendas. Devuelve cuántas filas borró.
CREATE FUNCTION app.purgar_visitas() RETURNS bigint
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public, app AS $$
DECLARE
	hoy date := (now() AT TIME ZONE 'America/Argentina/Buenos_Aires')::date;
	a bigint;
	b bigint;
BEGIN
	PERFORM set_config('app.purga_visitas', 'si', true);
	DELETE FROM visitas_vistos WHERE dia < hoy;
	GET DIAGNOSTICS a = ROW_COUNT;
	DELETE FROM visitas WHERE dia < hoy - 400;
	GET DIAGNOSTICS b = ROW_COUNT;
	PERFORM set_config('app.purga_visitas', '', true);
	RETURN a + b;
END
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION app.purgar_visitas() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION app.purgar_visitas() TO mostrador_app;

-- +goose Down
DROP FUNCTION app.purgar_visitas();
DROP TABLE visitas_vistos;
DROP TABLE visitas;
