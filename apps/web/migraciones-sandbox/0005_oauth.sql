-- "Manejá Mostrador desde tu IA": OAuth 2.1 para clientes MCP (Claude, ChatGPT…), como en Caja.
-- En la demo, cada conexión queda atada al sandbox del visitante y se va con él (ON DELETE CASCADE
-- y vence_en = el vencimiento del sandbox). Los códigos y los tokens se guardan hasheados (sha256).

-- Clientes por registro dinámico (RFC 7591): son públicos (sin secreto) y valen para cualquier sandbox.
CREATE TABLE oauth_clientes (
	id TEXT PRIMARY KEY,
	nombre TEXT NOT NULL,
	-- JSON con las direcciones de vuelta permitidas.
	redirecciones TEXT NOT NULL,
	creado_en INTEGER NOT NULL
);

-- Códigos de autorización: un solo uso (se borran al canjear) y 5 minutos.
CREATE TABLE oauth_codigos (
	hash TEXT PRIMARY KEY,
	cliente_id TEXT NOT NULL REFERENCES oauth_clientes (id) ON DELETE CASCADE,
	sandbox_id TEXT NOT NULL REFERENCES sandboxes (id) ON DELETE CASCADE,
	redireccion TEXT NOT NULL,
	desafio TEXT NOT NULL,
	permisos TEXT NOT NULL,
	vence_en INTEGER NOT NULL
);
CREATE INDEX oauth_codigos_sandbox ON oauth_codigos (sandbox_id);

-- Una fila por conexión (lo que el panel lista y desconecta): el access token corto y el refresh
-- token, que rota en cada uso (refresh_anterior detecta la reutilización y corta la conexión).
-- La ventana del límite de pedidos va en la misma fila: autenticar y contar es una sola escritura.
CREATE TABLE oauth_conexiones (
	id TEXT PRIMARY KEY,
	sandbox_id TEXT NOT NULL REFERENCES sandboxes (id) ON DELETE CASCADE,
	cliente_id TEXT NOT NULL REFERENCES oauth_clientes (id) ON DELETE CASCADE,
	nombre TEXT NOT NULL,
	permisos TEXT NOT NULL,
	acceso_hash TEXT NOT NULL UNIQUE,
	acceso_vence INTEGER NOT NULL,
	refresh_hash TEXT NOT NULL UNIQUE,
	refresh_anterior TEXT,
	vence_en INTEGER NOT NULL,
	creado_en INTEGER NOT NULL,
	usado_en INTEGER,
	pedidos INTEGER NOT NULL DEFAULT 0,
	ventana_hasta INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX oauth_conexiones_sandbox ON oauth_conexiones (sandbox_id);
CREATE INDEX oauth_conexiones_cliente ON oauth_conexiones (cliente_id);
CREATE INDEX oauth_conexiones_anterior ON oauth_conexiones (refresh_anterior);

-- Límites por ventana (registro de clientes por IP), como rate_limits de Caja.
CREATE TABLE limites (
	clave TEXT PRIMARY KEY,
	cuenta INTEGER NOT NULL,
	reinicia_en INTEGER NOT NULL
);
CREATE INDEX limites_reinicia ON limites (reinicia_en);
