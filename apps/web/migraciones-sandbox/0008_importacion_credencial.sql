-- La credencial de la plataforma vieja (el token de Tiendanube del OAuth), cifrada con ENCRYPTION_KEY
-- (lib/server/cifrado.ts). Se borra al terminar, fallar o cancelar, como en Go; la purga de 24 h se
-- lleva la fila con el sandbox.
ALTER TABLE importaciones ADD COLUMN credencial TEXT;
