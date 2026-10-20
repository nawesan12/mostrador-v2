-- +goose Up
-- Botón de arrepentimiento (Res. 424/2020 de la Secretaría de Comercio Interior): cualquiera puede
-- pedir la revocación de un plan contratado a distancia, sin iniciar sesión. Es un trámite con la
-- plataforma, no un dato de una tienda: vive fuera de public, en `soporte`.
CREATE SCHEMA soporte;
GRANT USAGE ON SCHEMA soporte TO mostrador_app, mostrador_staff;

-- El número del código de trámite (ARR-AAAA-NNNNNN). Arranca en 100000 para que tenga 6 cifras.
CREATE SEQUENCE soporte.arrepentimientos_numero START 100000;
GRANT USAGE ON SEQUENCE soporte.arrepentimientos_numero TO mostrador_app;

CREATE TABLE soporte.arrepentimientos (
	codigo text PRIMARY KEY CHECK (codigo ~ '^ARR-\d{4}-\d{6,}$'),
	email citext NOT NULL CHECK (length(email) <= 254),
	nombre text NOT NULL CHECK (length(nombre) BETWEEN 1 AND 80),
	documento text NOT NULL CHECK (documento ~ '^\d{7,11}$'),
	plan text NOT NULL CHECK (length(plan) BETWEEN 1 AND 120),
	motivo text CHECK (length(motivo) <= 1000),
	estado text NOT NULL DEFAULT 'recibido' CHECK (estado IN ('recibido', 'procesado', 'rechazado')),
	creado_en timestamptz NOT NULL DEFAULT now(),
	ip inet
);
CREATE INDEX arrepentimientos_email ON soporte.arrepentimientos (email, creado_en DESC);
-- La app sólo los registra (y lee el código que devuelve el INSERT); los datos personales los ve
-- únicamente el backoffice (staff), que es quien los procesa.
GRANT INSERT, SELECT (codigo) ON soporte.arrepentimientos TO mostrador_app;
GRANT SELECT, UPDATE (estado) ON soporte.arrepentimientos TO mostrador_staff;

-- +goose Down
DROP SCHEMA soporte CASCADE;
