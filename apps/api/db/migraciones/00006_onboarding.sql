-- +goose Up
-- Lo que se eligió en el onboarding (objetivos, canales, secciones, cobros y envíos, tono de Hornero,
-- equipo) y cuándo se terminó. Es configuración inicial: cada área del panel la va tomando.
ALTER TABLE tiendas ADD COLUMN onboarding jsonb NOT NULL DEFAULT '{}' CHECK (pg_column_size(onboarding) <= 16384);
ALTER TABLE tiendas ADD COLUMN onboarding_terminado_en timestamptz;
GRANT UPDATE (onboarding, onboarding_terminado_en) ON tiendas TO mostrador_app;

-- +goose Down
ALTER TABLE tiendas DROP COLUMN onboarding_terminado_en;
ALTER TABLE tiendas DROP COLUMN onboarding;
