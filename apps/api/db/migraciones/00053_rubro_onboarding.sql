-- +goose Up
-- El onboarding guarda el rubro que eligió la persona en «Tu negocio» (antes quedaba el de
-- registro, «regionales», para todas).
GRANT UPDATE (rubro) ON tiendas TO mostrador_app;

-- +goose Down
REVOKE UPDATE (rubro) ON tiendas FROM mostrador_app;
