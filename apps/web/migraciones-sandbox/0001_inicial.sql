-- Sandbox de la demo: cada visitante tiene el suyo, con vencimiento a las 24 h.
-- "Copia al escribir": los datos de ejemplo de cada rubro viven en el código (iguales para todos);
-- acá sólo se guarda lo que el visitante cambió. Crear un sandbox es una escritura.
CREATE TABLE sandboxes (
	id TEXT PRIMARY KEY,
	rubro TEXT NOT NULL,
	creado_en INTEGER NOT NULL,
	vence_en INTEGER NOT NULL
);
-- La purga busca los vencidos por fecha.
CREATE INDEX sandboxes_vence ON sandboxes (vence_en);

-- Lo que el visitante cambió, por entidad: la fila reemplaza a la de ejemplo (o la borra).
CREATE TABLE cambios (
	sandbox_id TEXT NOT NULL REFERENCES sandboxes (id) ON DELETE CASCADE,
	entidad TEXT NOT NULL,
	clave TEXT NOT NULL,
	datos TEXT,
	borrado INTEGER NOT NULL DEFAULT 0,
	actualizado_en INTEGER NOT NULL,
	PRIMARY KEY (sandbox_id, entidad, clave)
);
