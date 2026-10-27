-- Fotos de productos en R2 (bucket mostrador-files). El dueño es la tienda (su id va en la URL
-- pública; en la demo, el uuid derivado del sandbox, nunca el id del sandbox, que es la cookie).
-- Sirve para los topes (por tienda y en total: la cuenta es de pago por uso y el plan es quedar en
-- el gratis) y, con sandbox_id, para borrar las de la demo cuando su sandbox vence.
CREATE TABLE archivos (
	clave TEXT PRIMARY KEY,
	duenio TEXT NOT NULL,
	sandbox_id TEXT,
	bytes INTEGER NOT NULL,
	creado_en INTEGER NOT NULL
);
CREATE INDEX archivos_duenio ON archivos (duenio);
CREATE INDEX archivos_sandbox ON archivos (sandbox_id);
