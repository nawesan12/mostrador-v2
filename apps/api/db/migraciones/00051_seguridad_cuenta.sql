-- +goose Up
-- Seguridad de la cuenta (docs/decisiones.md §33). Las opciones son de la persona, no de la tienda:
-- las sesiones son suyas. {"avisoIngreso": bool, "inactivas": bool}; lo que falta, por defecto
-- (aviso prendido, inactivas apagado).
ALTER TABLE auth.usuarios ADD COLUMN seguridad jsonb NOT NULL DEFAULT '{}';
GRANT UPDATE (seguridad) ON auth.usuarios TO mostrador_app;

-- El mail «Entraron a tu cuenta» sale por el outbox sin tienda (es de la persona, y así ninguna
-- tienda ve las IP de quien entra). La app sólo puede insertar los de cuentas; no los ve.
CREATE POLICY cuentas_inserta ON outbox FOR INSERT TO mostrador_app
	WITH CHECK (tienda_id IS NULL AND tipo LIKE 'cuenta.%');

-- +goose Down
DROP POLICY cuentas_inserta ON outbox;
DELETE FROM outbox WHERE tienda_id IS NULL AND tipo LIKE 'cuenta.%';
ALTER TABLE auth.usuarios DROP COLUMN seguridad;
