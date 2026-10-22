-- +goose Up
-- Lo que se marcó como hecho en "Para hacer hoy", por día de Argentina.
CREATE TABLE tareas_hechas (
	tienda_id uuid NOT NULL REFERENCES tiendas (id) ON DELETE CASCADE,
	dia date NOT NULL,
	tarea text NOT NULL,
	en timestamptz NOT NULL DEFAULT now(),
	PRIMARY KEY (tienda_id, dia, tarea)
);
SELECT app.aplicar_rls('tareas_hechas');
GRANT SELECT, INSERT, DELETE ON tareas_hechas TO mostrador_app;

-- +goose Down
DROP TABLE tareas_hechas;
