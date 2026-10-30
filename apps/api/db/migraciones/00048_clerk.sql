-- +goose Up
-- Ingreso social con Clerk (la identidad compartida de Grupo Vimo): Google, Apple y los que tenga la
-- instancia. Reemplaza al OAuth propio de Google; Mercado Pago y Mercado Libre siguen como estaban.

-- Las identidades de Google propias se borran: al entrar con Google por Clerk, la cuenta se vuelve a
-- vincular sola por el email verificado (sólo existe la cuenta de QA).
DELETE FROM auth.identidades WHERE proveedor = 'google';
ALTER TABLE auth.identidades DROP CONSTRAINT identidades_proveedor_check;
ALTER TABLE auth.identidades ADD CONSTRAINT identidades_proveedor_check
	CHECK (proveedor IN ('clerk', 'mercadopago', 'mercadolibre'));

-- La foto de perfil que da el proveedor (Google, Apple). Sólo https.
ALTER TABLE auth.usuarios ADD COLUMN foto text CHECK (foto ~ '^https://' AND length(foto) <= 1000);
GRANT UPDATE (foto) ON auth.usuarios TO mostrador_app;

-- +goose Down
REVOKE UPDATE (foto) ON auth.usuarios FROM mostrador_app;
ALTER TABLE auth.usuarios DROP COLUMN foto;
DELETE FROM auth.identidades WHERE proveedor = 'clerk';
ALTER TABLE auth.identidades DROP CONSTRAINT identidades_proveedor_check;
ALTER TABLE auth.identidades ADD CONSTRAINT identidades_proveedor_check
	CHECK (proveedor IN ('google', 'mercadopago', 'mercadolibre'));
