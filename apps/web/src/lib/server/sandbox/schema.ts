import { sql } from 'drizzle-orm';
import { index, integer, primaryKey, sqliteTable, text } from 'drizzle-orm/sqlite-core';

/** Espejo de migraciones-sandbox/*.sql (el SQL manda; el test de esquema compara los dos). */
export const sandboxes = sqliteTable(
	'sandboxes',
	{
		id: text('id').primaryKey(),
		rubro: text('rubro').notNull(),
		plantilla: text('plantilla'),
		creadoEn: integer('creado_en').notNull(),
		venceEn: integer('vence_en').notNull()
	},
	(t) => [index('sandboxes_vence').on(t.venceEn)]
);

export const cambios = sqliteTable(
	'cambios',
	{
		sandboxId: text('sandbox_id')
			.notNull()
			.references(() => sandboxes.id, { onDelete: 'cascade' }),
		entidad: text('entidad').notNull(),
		clave: text('clave').notNull(),
		datos: text('datos'),
		borrado: integer('borrado', { mode: 'boolean' })
			.notNull()
			.default(sql`0`),
		actualizadoEn: integer('actualizado_en').notNull()
	},
	(t) => [
		primaryKey({ columns: [t.sandboxId, t.entidad, t.clave] }),
		index('cambios_por_clave').on(t.entidad, t.clave)
	]
);
