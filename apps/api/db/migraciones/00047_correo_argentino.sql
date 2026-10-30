-- +goose Up
-- Correo Argentino (API MiCorreo): cada tienda conecta su cuenta de MiCorreo. De la cuenta sólo se
-- guarda el customerId que devuelve /users/validate (la contraseña de MiCorreo no se guarda). Si la
-- tienda usa credenciales de API propias (las que da Correo para /token), van cifradas con
-- plataforma/cifrado (contexto "correo:<tienda>"); NULL: se usan las de Tu Local Web.
CREATE TABLE ca_conexiones (
	tienda_id uuid PRIMARY KEY REFERENCES tiendas (id) ON DELETE CASCADE,
	cliente text NOT NULL CHECK (cliente ~ '^[0-9A-Za-z-]{1,40}$'),
	email text NOT NULL CHECK (length(email) <= 120),
	credencial bytea,
	-- De dónde sale el paquete: {cp, calle, altura, piso, depto, localidad, provincia}.
	origen jsonb NOT NULL CHECK (jsonb_typeof(origen) = 'object' AND pg_column_size(origen) <= 1024),
	-- El paquete típico: {peso (g por unidad), alto, ancho, largo (cm)}. Los productos no tienen peso.
	paquete jsonb NOT NULL CHECK (jsonb_typeof(paquete) = 'object' AND pg_column_size(paquete) <= 512),
	conectada_en timestamptz NOT NULL DEFAULT now()
);
SELECT app.aplicar_rls('ca_conexiones');
GRANT SELECT, INSERT, UPDATE, DELETE ON ca_conexiones TO mostrador_app;

-- Cada pedido dado de alta en MiCorreo (/shipping/import). referencia es el extOrderId que se mandó:
-- MiCorreo no deja importar dos veces la misma, así que un reintento no duplica el envío. El número
-- de seguimiento lo asigna MiCorreo al pagar el rótulo: la tienda lo carga y queda también en
-- pedidos.seguimiento.
CREATE TABLE ca_envios (
	tienda_id uuid NOT NULL,
	pedido integer NOT NULL,
	tipo text NOT NULL CHECK (tipo IN ('D', 'S')),
	sucursal text,
	referencia text NOT NULL,
	importado_en timestamptz NOT NULL DEFAULT now(),
	seguimiento text CHECK (seguimiento IS NULL OR length(seguimiento) BETWEEN 6 AND 40),
	PRIMARY KEY (tienda_id, pedido),
	FOREIGN KEY (tienda_id, pedido) REFERENCES pedidos (tienda_id, numero) ON DELETE CASCADE
);
SELECT app.aplicar_rls('ca_envios');
GRANT SELECT, INSERT ON ca_envios TO mostrador_app;
GRANT UPDATE (seguimiento) ON ca_envios TO mostrador_app;

-- +goose Down
DROP TABLE ca_envios;
DROP TABLE ca_conexiones;
