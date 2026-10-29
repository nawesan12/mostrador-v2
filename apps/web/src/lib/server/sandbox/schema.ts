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

/** OAuth 2.1 del conector MCP (migración 0005). */
export const oauthClientes = sqliteTable('oauth_clientes', {
	id: text('id').primaryKey(),
	nombre: text('nombre').notNull(),
	redirecciones: text('redirecciones').notNull(),
	creadoEn: integer('creado_en').notNull(),
	usadoEn: integer('usado_en')
});

export const oauthCodigos = sqliteTable(
	'oauth_codigos',
	{
		hash: text('hash').primaryKey(),
		clienteId: text('cliente_id')
			.notNull()
			.references(() => oauthClientes.id, { onDelete: 'cascade' }),
		sandboxId: text('sandbox_id')
			.notNull()
			.references(() => sandboxes.id, { onDelete: 'cascade' }),
		redireccion: text('redireccion').notNull(),
		desafio: text('desafio').notNull(),
		permisos: text('permisos').notNull(),
		venceEn: integer('vence_en').notNull()
	},
	(t) => [index('oauth_codigos_sandbox').on(t.sandboxId)]
);

export const oauthConexiones = sqliteTable(
	'oauth_conexiones',
	{
		id: text('id').primaryKey(),
		sandboxId: text('sandbox_id')
			.notNull()
			.references(() => sandboxes.id, { onDelete: 'cascade' }),
		clienteId: text('cliente_id')
			.notNull()
			.references(() => oauthClientes.id, { onDelete: 'cascade' }),
		nombre: text('nombre').notNull(),
		permisos: text('permisos').notNull(),
		accesoHash: text('acceso_hash').notNull().unique(),
		accesoVence: integer('acceso_vence').notNull(),
		refreshHash: text('refresh_hash').notNull().unique(),
		refreshAnterior: text('refresh_anterior'),
		venceEn: integer('vence_en').notNull(),
		creadoEn: integer('creado_en').notNull(),
		usadoEn: integer('usado_en'),
		pedidos: integer('pedidos').notNull().default(0),
		ventanaHasta: integer('ventana_hasta').notNull().default(0)
	},
	(t) => [
		index('oauth_conexiones_sandbox').on(t.sandboxId),
		index('oauth_conexiones_cliente').on(t.clienteId),
		index('oauth_conexiones_anterior').on(t.refreshAnterior)
	]
);

export const limites = sqliteTable(
	'limites',
	{
		clave: text('clave').primaryKey(),
		cuenta: integer('cuenta').notNull(),
		reiniciaEn: integer('reinicia_en').notNull()
	},
	(t) => [index('limites_reinicia').on(t.reiniciaEn)]
);

/** Fotos subidas a R2: la tienda dueña (y el sandbox en la demo) y el tamaño, para topes y purga. */
export const archivos = sqliteTable(
	'archivos',
	{
		clave: text('clave').primaryKey(),
		duenio: text('duenio').notNull(),
		sandboxId: text('sandbox_id'),
		bytes: integer('bytes').notNull(),
		creadoEn: integer('creado_en').notNull()
	},
	(t) => [index('archivos_duenio').on(t.duenio), index('archivos_sandbox').on(t.sandboxId)]
);

/** Migración desde otras plataformas (migración 0007): ficha de cliente, historial, redirecciones y trabajo. */
export const clientesSandbox = sqliteTable(
	'clientes',
	{
		sandboxId: text('sandbox_id')
			.notNull()
			.references(() => sandboxes.id, { onDelete: 'cascade' }),
		id: text('id').notNull(),
		nombre: text('nombre').notNull(),
		email: text('email'),
		telefono: text('telefono'),
		ciudad: text('ciudad').notNull().default(''),
		provincia: text('provincia').notNull().default(''),
		notas: text('notas').notNull().default(''),
		origen: text('origen').notNull(),
		externo: text('externo'),
		creadoEn: integer('creado_en').notNull()
	},
	(t) => [primaryKey({ columns: [t.sandboxId, t.id] })]
);

export const historialPedidos = sqliteTable(
	'historial_pedidos',
	{
		sandboxId: text('sandbox_id')
			.notNull()
			.references(() => sandboxes.id, { onDelete: 'cascade' }),
		id: text('id').notNull(),
		plataforma: text('plataforma').notNull(),
		externo: text('externo').notNull(),
		numeroExterno: text('numero_externo').notNull().default(''),
		fecha: integer('fecha').notNull(),
		clienteEmail: text('cliente_email'),
		clienteNombre: text('cliente_nombre').notNull().default(''),
		total: integer('total').notNull(),
		items: text('items').notNull().default('[]'),
		estado: text('estado').notNull().default('')
	},
	(t) => [
		primaryKey({ columns: [t.sandboxId, t.id] }),
		index('historial_pedidos_email').on(t.sandboxId, t.clienteEmail)
	]
);

export const redirecciones = sqliteTable(
	'redirecciones',
	{
		sandboxId: text('sandbox_id')
			.notNull()
			.references(() => sandboxes.id, { onDelete: 'cascade' }),
		desde: text('desde').notNull(),
		hacia: text('hacia').notNull(),
		origen: text('origen').notNull(),
		creada: integer('creada').notNull()
	},
	(t) => [primaryKey({ columns: [t.sandboxId, t.desde] })]
);

export const importaciones = sqliteTable(
	'importaciones',
	{
		sandboxId: text('sandbox_id')
			.notNull()
			.references(() => sandboxes.id, { onDelete: 'cascade' }),
		id: text('id').notNull(),
		plataforma: text('plataforma').notNull(),
		urlOrigen: text('url_origen').notNull().default(''),
		estado: text('estado').notNull().default('pendiente'),
		etapa: text('etapa').notNull().default('productos'),
		cursor: text('cursor'),
		pasos: integer('pasos').notNull().default(0),
		progreso: text('progreso').notNull(),
		errores: text('errores').notNull().default('[]'),
		creada: integer('creada').notNull(),
		terminada: integer('terminada'),
		/** El token de la plataforma vieja, cifrado (iv.ciphertext). Se borra al terminar. */
		credencial: text('credencial')
	},
	(t) => [
		primaryKey({ columns: [t.sandboxId, t.id] }),
		index('importaciones_recientes').on(t.sandboxId, t.creada)
	]
);
