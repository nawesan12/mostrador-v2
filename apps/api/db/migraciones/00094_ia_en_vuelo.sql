-- +goose Up
-- Las preguntas a Tincho que se están respondiendo ahora, por tienda: el cupo se reserva antes de
-- llamar al modelo (con la fila bloqueada) y se libera al terminar. Antes se leía el uso y se
-- anotaba al final, y N preguntas en paralelo pasaban todas. `desde` vence una reserva colgada.
CREATE TABLE ia_en_vuelo (
	tienda_id uuid PRIMARY KEY REFERENCES tiendas (id) ON DELETE CASCADE,
	cantidad int NOT NULL DEFAULT 0 CHECK (cantidad >= 0),
	desde timestamptz NOT NULL DEFAULT now()
);
SELECT app.aplicar_rls('ia_en_vuelo');
GRANT SELECT, INSERT, UPDATE ON ia_en_vuelo TO mostrador_app;

-- +goose Down
DROP TABLE ia_en_vuelo;
