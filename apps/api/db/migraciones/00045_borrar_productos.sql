-- +goose Up
-- Borrar un producto desde el panel. Sus variantes, preguntas y la fila de Mercado Libre se van
-- en cascada; las reseñas quedan, sin el producto. Los pedidos guardan su copia del producto.
GRANT DELETE ON productos TO mostrador_app;

-- +goose Down
REVOKE DELETE ON productos FROM mostrador_app;
