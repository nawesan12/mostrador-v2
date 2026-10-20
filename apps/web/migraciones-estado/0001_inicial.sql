-- /estado: el monitor de mostrador-jobs (cron horario) escribe acá; el sitio sólo lee. Vive aparte
-- del sandbox y de la API: el estado tiene que poder mostrarse aunque lo demás esté caído.

-- Resumen por componente y día (de la Argentina): cuántos chequeos y cuántos fallaron.
CREATE TABLE dias (
	componente TEXT NOT NULL,
	dia TEXT NOT NULL,
	chequeos INTEGER NOT NULL DEFAULT 0,
	fallas INTEGER NOT NULL DEFAULT 0,
	PRIMARY KEY (componente, dia)
);

-- El último chequeo de cada componente y cuántas fallas seguidas lleva.
CREATE TABLE ultimo (
	componente TEXT PRIMARY KEY,
	ok INTEGER NOT NULL,
	fallas_seguidas INTEGER NOT NULL DEFAULT 0,
	latencia_ms INTEGER,
	detalle TEXT,
	chequeado_en INTEGER NOT NULL
);

-- Incidentes: los abre el monitor (dos fallas seguidas) y los cierra cuando vuelve; los
-- mantenimientos se cargan a mano.
CREATE TABLE incidentes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	componente TEXT NOT NULL,
	tipo TEXT NOT NULL DEFAULT 'incidente' CHECK (tipo IN ('incidente', 'mantenimiento')),
	titulo TEXT NOT NULL,
	detalle TEXT NOT NULL DEFAULT '',
	inicio INTEGER NOT NULL,
	fin INTEGER
);
CREATE INDEX incidentes_inicio ON incidentes (inicio DESC);
CREATE INDEX incidentes_abiertos ON incidentes (componente) WHERE fin IS NULL;

-- Quién quiere que le avisen (email o WhatsApp). La baja es con el token del aviso.
CREATE TABLE suscripciones (
	contacto TEXT PRIMARY KEY,
	canal TEXT NOT NULL CHECK (canal IN ('email', 'whatsapp')),
	token_baja TEXT NOT NULL UNIQUE,
	creada_en INTEGER NOT NULL
);
