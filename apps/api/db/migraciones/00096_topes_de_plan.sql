-- +goose Up
-- Topes de plan (19/11/2026): la tienda muestra los primeros N productos por antigüedad y el resto
-- queda oculto y en sólo lectura hasta subir de plan o borrar alguno (dominio/planes). Este índice
-- hace barato contar y ordenar los productos de una tienda por creado_en. Sólo agrega un índice: no
-- toca datos.
CREATE INDEX IF NOT EXISTS productos_orden ON productos (tienda_id, creado_en, id);
-- Lo mismo para el equipo: las personas que quedan fuera del plan son las últimas en entrar.
CREATE INDEX IF NOT EXISTS membresias_orden ON auth.membresias (tienda_id, creada_en);

-- +goose Down
DROP INDEX IF EXISTS auth.membresias_orden;
DROP INDEX IF EXISTS productos_orden;
