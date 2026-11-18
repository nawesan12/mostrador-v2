-- +goose Up
-- Tincho con Claude (docs-ia/asistente-ia.md): lo que gasta cada tienda en la IA, por día y por ruta
-- (chat, mejorar, telegram; «motor» cuenta las respuestas sin IA, a costo cero). Con esto se miden y
-- cortan los cupos de los planes y el tope diario. Se suma con cada respuesta, nunca en memoria.
CREATE TABLE ia_uso (
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	-- Día de Argentina.
	dia date NOT NULL,
	ruta text NOT NULL CHECK (ruta IN ('chat', 'mejorar', 'telegram', 'motor')),
	-- Respuestas (una por pregunta, no por vuelta de herramientas).
	pedidos int NOT NULL DEFAULT 0 CHECK (pedidos >= 0),
	entrada bigint NOT NULL DEFAULT 0 CHECK (entrada >= 0),
	cache_leido bigint NOT NULL DEFAULT 0 CHECK (cache_leido >= 0),
	cache_escrito bigint NOT NULL DEFAULT 0 CHECK (cache_escrito >= 0),
	salida bigint NOT NULL DEFAULT 0 CHECK (salida >= 0),
	-- Micro-dólares (USD × 1.000.000), redondeado para arriba.
	costo bigint NOT NULL DEFAULT 0 CHECK (costo >= 0),
	PRIMARY KEY (tienda_id, dia, ruta)
);
SELECT app.aplicar_rls('ia_uso');
GRANT SELECT, INSERT, UPDATE ON ia_uso TO mostrador_app;

-- El gasto de toda la plataforma en el mes: el tope global (la parte del presupuesto de Tu Local
-- Web) se mira acá sin sumar todas las tiendas, que RLS no deja ver juntas. Fuera de public y sin
-- datos de ninguna tienda (como app.slugs): sólo el mes y la plata.
CREATE TABLE app.ia_gasto (
	mes text PRIMARY KEY CHECK (mes ~ '^[0-9]{4}-[0-9]{2}$'),
	costo bigint NOT NULL DEFAULT 0 CHECK (costo >= 0),
	pedidos int NOT NULL DEFAULT 0 CHECK (pedidos >= 0)
);
GRANT SELECT, INSERT, UPDATE ON app.ia_gasto TO mostrador_app;
GRANT SELECT ON app.ia_gasto TO mostrador_staff;

-- +goose Down
DROP TABLE app.ia_gasto;
DROP TABLE ia_uso;
