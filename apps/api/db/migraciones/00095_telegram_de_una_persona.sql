-- +goose Up
-- El Telegram vinculado es de una persona del equipo, no de la tienda: se guarda quién lo vinculó.
-- Si la sacan del equipo o deja de administrar, el chat pierde el acceso (Tincho y los avisos de
-- venta lo revisan con su membresía) y Quitar o bajar el rol lo borran. Los vínculos que ya existen
-- quedan con la dueña.
ALTER TABLE telegram_codigos ADD COLUMN usuario_id uuid REFERENCES auth.usuarios (id) ON DELETE CASCADE;
ALTER TABLE telegram_vinculos ADD COLUMN usuario_id uuid REFERENCES auth.usuarios (id) ON DELETE CASCADE;
GRANT UPDATE (usuario_id) ON telegram_vinculos TO mostrador_app;
ALTER TABLE telegram_vinculos NO FORCE ROW LEVEL SECURITY;
UPDATE telegram_vinculos v SET usuario_id = m.usuario_id
	FROM auth.membresias m WHERE m.tienda_id = v.tienda_id AND m.rol = 'duenio';
ALTER TABLE telegram_vinculos FORCE ROW LEVEL SECURITY;

-- +goose Down
ALTER TABLE telegram_vinculos DROP COLUMN usuario_id;
ALTER TABLE telegram_codigos DROP COLUMN usuario_id;
