-- +goose Up
-- Reseñas y preguntas que dejan los compradores en la tienda online. Una reseña con el código de
-- un pedido es "con compra" (no hay columna aparte: se deriva del pedido, que no se desincroniza).
-- La de una estrella sin compra no se borra sola: la decisión es de la tienda (oculta).
CREATE TABLE resenas (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	pedido_numero int,
	producto_id uuid,
	autor text NOT NULL CHECK (length(autor) BETWEEN 2 AND 60),
	estrellas int NOT NULL CHECK (estrellas BETWEEN 1 AND 5),
	texto text NOT NULL DEFAULT '' CHECK (length(texto) <= 1000),
	origen text NOT NULL DEFAULT 'Tienda',
	oculta boolean NOT NULL DEFAULT false,
	respuesta text CHECK (length(respuesta) <= 1000),
	creada timestamptz NOT NULL DEFAULT now(),
	-- Si se borra el producto, la reseña queda (sin producto).
	FOREIGN KEY (tienda_id, producto_id) REFERENCES productos (tienda_id, id) ON DELETE SET NULL (producto_id)
);
CREATE UNIQUE INDEX resenas_por_pedido ON resenas (tienda_id, pedido_numero, coalesce(producto_id::text, ''))
	WHERE pedido_numero IS NOT NULL;
CREATE INDEX resenas_recientes ON resenas (tienda_id, creada DESC);
SELECT app.aplicar_rls('resenas');
GRANT SELECT, INSERT, UPDATE (oculta, respuesta) ON resenas TO mostrador_app;

CREATE TABLE preguntas (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	producto_id uuid NOT NULL,
	nombre text NOT NULL CHECK (length(nombre) BETWEEN 2 AND 60),
	email text CHECK (length(email) <= 120),
	pregunta text NOT NULL CHECK (length(pregunta) BETWEEN 3 AND 500),
	respuesta text CHECK (length(respuesta) <= 1000),
	respondida_en timestamptz,
	creada timestamptz NOT NULL DEFAULT now(),
	FOREIGN KEY (tienda_id, producto_id) REFERENCES productos (tienda_id, id) ON DELETE CASCADE
);
CREATE INDEX preguntas_producto ON preguntas (tienda_id, producto_id, creada DESC);
SELECT app.aplicar_rls('preguntas');
GRANT SELECT, INSERT, UPDATE (respuesta, respondida_en) ON preguntas TO mostrador_app;

-- +goose Down
DROP TABLE preguntas;
DROP TABLE resenas;
