-- +goose Up
-- Devoluciones de cada tienda. Entran por el botón de arrepentimiento de la tienda online (Ley
-- 24.240, art. 34; Res. 424/2020: código de trámite) o las carga la tienda. El código no se
-- adivina (secuencia propia + año) y es por tienda.
CREATE SEQUENCE devoluciones_numero;
GRANT USAGE ON SEQUENCE devoluciones_numero TO mostrador_app;
CREATE TABLE devoluciones (
	id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	codigo text NOT NULL UNIQUE,
	pedido_numero int NOT NULL,
	nombre text NOT NULL CHECK (length(nombre) BETWEEN 2 AND 80),
	email text NOT NULL CHECK (length(email) <= 120),
	motivo text NOT NULL DEFAULT '' CHECK (length(motivo) <= 500),
	resolucion text NOT NULL DEFAULT 'reintegro' CHECK (resolucion IN ('reintegro', 'reemplazo', 'cambio')),
	estado text NOT NULL DEFAULT 'Pendiente' CHECK (estado IN ('Pendiente', 'Aprobada', 'Cambio ofrecido')),
	estado_anterior text,
	monto bigint NOT NULL CHECK (monto >= 0),
	origen text NOT NULL DEFAULT 'arrepentimiento' CHECK (origen IN ('arrepentimiento', 'panel')),
	creada timestamptz NOT NULL DEFAULT now(),
	FOREIGN KEY (tienda_id, pedido_numero) REFERENCES pedidos (tienda_id, numero) ON DELETE CASCADE
);
-- Un pedido de arrepentimiento por pedido (el segundo devuelve el mismo código).
CREATE UNIQUE INDEX devoluciones_pedido ON devoluciones (tienda_id, pedido_numero) WHERE origen = 'arrepentimiento';
CREATE INDEX devoluciones_recientes ON devoluciones (tienda_id, creada DESC);
SELECT app.aplicar_rls('devoluciones');
GRANT SELECT, INSERT, UPDATE (estado, estado_anterior) ON devoluciones TO mostrador_app;

-- +goose Down
DROP TABLE devoluciones;
DROP SEQUENCE devoluciones_numero;
