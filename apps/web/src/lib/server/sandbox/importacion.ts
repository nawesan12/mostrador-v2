/**
 * Importación en la demo (D1): las mismas reglas que Go (dominio/importacion), con el sandbox en
 * lugar de la tienda. No hay outbox: el navegador pide `avanzarImportacion` y cada pedido hace un
 * paso (una página de la fuente). `pasos` hace de candado: el paso se toma con un UPDATE condicional
 * antes de guardar, así dos pedidos a la vez no guardan dos veces la misma página. Las escrituras de
 * clientes, pedidos y redirecciones van en un `db.batch` con `json_each` (dos o tres parámetros por
 * consulta, sin importar cuántas filas: el tope de D1 es 100).
 */
import { and, desc, eq, sql } from 'drizzle-orm';
import { randomToken } from '../crypto.ts';
import {
	ErrorDefinitivo,
	ETAPAS,
	FUENTES,
	MAX_ERRORES,
	MAX_PASOS,
	PLATAFORMAS,
	progresoVacio,
	type Abrir,
	type ClienteViejo,
	type ErrorImportacion,
	type Etapa,
	type Importacion,
	type ImportacionNueva,
	type Pagina,
	type PedidoHistorial,
	type PedidoViejo,
	type Plataforma,
	type ProductoViejo,
	type Progreso
} from '../importacion/fuente.ts';
import { haciaValido, normalizarRuta } from '../importacion/rutas.ts';
import { errorOpciones } from '#lib/tienda/variantes.ts';
import type { ProductoEditable } from '../datos/tipos.ts';
import { clientesSandbox, historialPedidos, importaciones, redirecciones } from './schema.ts';
import { inArray } from './sql.ts';
import type { DbSandbox } from './sandbox.ts';

type Fila = typeof importaciones.$inferSelect;
const ACTIVAS = ['pendiente', 'corriendo'];

function aImportacion(f: Fila): Importacion {
	return {
		id: f.id,
		plataforma: f.plataforma as Plataforma,
		urlOrigen: f.urlOrigen,
		estado: f.estado as Importacion['estado'],
		etapa: f.etapa as Etapa,
		progreso: { ...progresoVacio(), ...(JSON.parse(f.progreso) as Partial<Progreso>) },
		errores: JSON.parse(f.errores) as ErrorImportacion[],
		creada: new Date(f.creada).toISOString(),
		terminada: f.terminada === null ? null : new Date(f.terminada).toISOString()
	};
}

/** Lo que el paso necesita del resto del sandbox: guardar con la misma función que el panel. */
export type OpsImportacion = {
	guardarProducto(p: ProductoEditable): Promise<{ id: string } | null>;
	/** Los SKU que ya tiene la tienda (para no repetir productos). */
	skus(): Promise<Set<string>>;
};

export async function importacionActual(db: DbSandbox, sb: string): Promise<Importacion | null> {
	const [f] = await db
		.select()
		.from(importaciones)
		.where(eq(importaciones.sandboxId, sb))
		.orderBy(
			desc(sql`${importaciones.estado} in ('pendiente', 'corriendo')`),
			desc(importaciones.creada)
		)
		.limit(1);
	return f ? aImportacion(f) : null;
}

export async function crearImportacion(
	db: DbSandbox,
	sb: string,
	n: ImportacionNueva,
	fuentes: Partial<Record<Plataforma, Abrir>> = FUENTES,
	ahora = Date.now()
): Promise<Importacion | 'invalido' | 'sin-fuente' | 'activa'> {
	if (!PLATAFORMAS.includes(n.plataforma)) return 'invalido';
	if (!fuentes[n.plataforma]) return 'sin-fuente';
	const url = (n.url ?? '').trim();
	if (url) {
		try {
			const u = new URL(url);
			if (!/^https?:$/.test(u.protocol) || url.length > 300) return 'invalido';
		} catch {
			return 'invalido';
		}
	}
	// La credencial no se guarda en la demo (no hay con qué cifrarla): las fuentes que la piden
	// sólo andan con la API.
	const fila = {
		sandboxId: sb,
		id: randomToken(12),
		plataforma: n.plataforma,
		urlOrigen: url,
		progreso: JSON.stringify(progresoVacio()),
		creada: ahora
	};
	try {
		await db.insert(importaciones).values(fila);
	} catch (e) {
		// drizzle envuelve el error de D1: el UNIQUE viene en la causa.
		const texto = `${String(e)} ${String((e as { cause?: unknown })?.cause ?? '')}`;
		if (texto.includes('UNIQUE')) return 'activa';
		throw e;
	}
	return (await importacionActual(db, sb))!;
}

export async function cancelarImportacion(
	db: DbSandbox,
	sb: string,
	id: string,
	ahora = Date.now()
): Promise<Importacion | 'no-existe' | 'terminada'> {
	const [f] = await db
		.update(importaciones)
		.set({ estado: 'cancelada', terminada: ahora })
		.where(
			and(
				eq(importaciones.sandboxId, sb),
				eq(importaciones.id, id),
				inArray(importaciones.estado, ACTIVAS)
			)
		)
		.returning();
	if (f) return aImportacion(f);
	const [x] = await db
		.select({ id: importaciones.id })
		.from(importaciones)
		.where(and(eq(importaciones.sandboxId, sb), eq(importaciones.id, id)))
		.limit(1);
	return x ? 'terminada' : 'no-existe';
}

/** El email en minúsculas; `malo` si venía algo que no parece un email (entra sin email). */
function emailDe(s: string | undefined): { email: string | null; malo: boolean } {
	const e = (s ?? '').trim().toLowerCase();
	if (!e) return { email: null, malo: false };
	const at = e.indexOf('@');
	if (e.length < 3 || e.length > 120 || at < 1 || at === e.length - 1 || /\s/.test(e))
		return { email: null, malo: true };
	return { email: e, malo: false };
}

/** Recorta a n letras (no unidades de UTF-16, como Go con runas) y avisa si cortó. */
function cortar(s: string | undefined | null, n: number): [string, boolean] {
	const t = (s ?? '').trim();
	const r = [...t];
	return r.length <= n ? [t, false] : [r.slice(0, n).join('').trim(), true];
}

/**
 * Un paso: lee una página de la fuente, toma el paso (si nadie lo tomó) y la guarda. Devuelve la
 * importación como quedó, o null si no hay ninguna. Sin outbox no hay reintentos con espera: un
 * error de la fuente deja la importación fallida con ese error.
 */
export async function avanzarImportacion(
	db: DbSandbox,
	sb: string,
	ops: OpsImportacion,
	fuentes: Partial<Record<Plataforma, Abrir>> = FUENTES,
	ahora = Date.now()
): Promise<Importacion | null> {
	const [f] = await db
		.select()
		.from(importaciones)
		.where(and(eq(importaciones.sandboxId, sb), inArray(importaciones.estado, ACTIVAS)))
		.limit(1);
	if (!f) return importacionActual(db, sb);
	const etapa = f.etapa as Etapa;
	const fallar = async (detalle: string) => {
		const errores = [
			...(JSON.parse(f.errores) as ErrorImportacion[]),
			{ etapa, externo: '', detalle: detalle.slice(0, 300) }
		].slice(-MAX_ERRORES);
		await db
			.update(importaciones)
			.set({ estado: 'fallida', errores: JSON.stringify(errores), terminada: ahora })
			.where(
				and(
					eq(importaciones.sandboxId, sb),
					eq(importaciones.id, f.id),
					eq(importaciones.pasos, f.pasos)
				)
			);
		return importacionActual(db, sb);
	};
	if (f.pasos >= MAX_PASOS)
		return fallar('La tienda de origen no terminaba de mandar datos: frenamos la importación.');
	const abrir = fuentes[f.plataforma as Plataforma];
	if (!abrir) return fallar('Todavía no podemos importar desde esa plataforma.');
	let pag: Pagina;
	try {
		const fuente = await abrir({ plataforma: f.plataforma as Plataforma, url: f.urlOrigen });
		pag = await fuente.pagina(etapa, f.cursor === null ? null : JSON.parse(f.cursor));
	} catch (e) {
		const m = e instanceof ErrorDefinitivo || e instanceof Error ? e.message : '';
		return fallar(m || 'La tienda de origen no respondió.');
	}
	const cursor =
		pag.siguiente === null || pag.siguiente === undefined ? null : JSON.stringify(pag.siguiente);
	if (cursor && cursor.length > 2048)
		return fallar('La tienda de origen mandó un cursor demasiado largo.');
	// Tomar el paso: si otro pedido lo tomó (o la cancelaron), no se guarda nada.
	const [tomada] = await db
		.update(importaciones)
		.set({ pasos: f.pasos + 1, estado: 'corriendo' })
		.where(
			and(
				eq(importaciones.sandboxId, sb),
				eq(importaciones.id, f.id),
				eq(importaciones.pasos, f.pasos),
				inArray(importaciones.estado, ACTIVAS)
			)
		)
		.returning({ id: importaciones.id });
	if (!tomada) return importacionActual(db, sb);
	const origen = `importado:${f.plataforma}`;
	const { prog, errores } = await guardarPagina(
		db,
		sb,
		f.plataforma as Plataforma,
		origen,
		etapa,
		pag,
		ops,
		ahora
	);
	const total = { ...progresoVacio(), ...(JSON.parse(f.progreso) as Partial<Progreso>) };
	for (const k of Object.keys(prog) as (keyof Progreso)[]) total[k] += prog[k];
	const todos = [...(JSON.parse(f.errores) as ErrorImportacion[]), ...errores].slice(-MAX_ERRORES);
	const sig = cursor ? etapa : (ETAPAS[ETAPAS.indexOf(etapa) + 1] ?? null);
	// Si la cancelaron mientras se guardaba, queda cancelada (con lo que alcanzó a entrar).
	await db
		.update(importaciones)
		.set({
			progreso: JSON.stringify(total),
			errores: JSON.stringify(todos),
			...(sig
				? { etapa: sig, cursor }
				: {
						estado: sql`case when ${importaciones.estado} = 'corriendo' then 'lista' else ${importaciones.estado} end`,
						cursor: null,
						terminada: sql`coalesce(${importaciones.terminada}, ${ahora})`
					})
		})
		.where(and(eq(importaciones.sandboxId, sb), eq(importaciones.id, f.id)));
	return importacionActual(db, sb);
}

type Resultado = { prog: Progreso; errores: ErrorImportacion[] };

/** Guarda lo de una etapa. Lo que no pasa las validaciones queda en los errores y el resto entra. */
export async function guardarPagina(
	db: DbSandbox,
	sb: string,
	plataforma: Plataforma,
	origen: string,
	etapa: Etapa,
	p: Pagina,
	ops: OpsImportacion,
	ahora = Date.now()
): Promise<Resultado> {
	const r: Resultado = { prog: progresoVacio(), errores: [] };
	const error = (externo: string, detalle: string) => r.errores.push({ etapa, externo, detalle });
	if (etapa === 'productos')
		await productos(db, sb, origen, p.productos ?? [], ops, r, error, ahora);
	if (etapa === 'clientes') await clientes(db, sb, origen, p.clientes ?? [], r, error, ahora);
	if (etapa === 'pedidos') await pedidos(db, sb, plataforma, p.pedidos ?? [], r, error);
	if (etapa === 'redirecciones') {
		const filas: { desde: string; hacia: string }[] = [];
		for (const x of p.redirecciones ?? []) {
			const desde = normalizarRuta(x.desde);
			if (!desde || !haciaValido(x.hacia))
				error(x.desde, 'La dirección vieja o la nueva no es un path válido.');
			else filas.push({ desde, hacia: x.hacia });
		}
		await redirigir(db, sb, origen, filas, ahora);
		r.prog.redirecciones += new Set(filas.map((f) => f.desde)).size;
	}
	return r;
}

async function redirigir(
	db: DbSandbox,
	sb: string,
	origen: string,
	filas: { desde: string; hacia: string }[],
	ahora: number
) {
	if (!filas.length) return;
	await db.run(sql`
		insert into redirecciones (sandbox_id, desde, hacia, origen, creada)
		select ${sb}, json_extract(value, '$.desde'), json_extract(value, '$.hacia'), ${origen}, ${ahora}
		from json_each(${JSON.stringify(filas)}) where true
		on conflict (sandbox_id, desde) do update set hacia = excluded.hacia, origen = excluded.origen`);
}

/** Como el editor del panel (y Go `Editable.Validar`): nombre de 2 a 120 letras, sin negativos. */
function errorProducto(e: ProductoEditable): string | null {
	const n = [...e.nombre.trim()].length;
	if (n < 2 || n > 120) return 'El nombre tiene que tener entre 2 y 120 letras.';
	if (e.precio < 0 || (e.costo ?? 0) < 0 || (e.stock ?? 0) < 0)
		return 'Precio, costo y stock no pueden ser negativos.';
	if (e.opciones) return errorOpciones(e.opciones);
	if (e.variantes?.length) return 'Para cargar variantes, el producto tiene que tener opciones.';
	return null;
}

async function productos(
	db: DbSandbox,
	sb: string,
	origen: string,
	lista: ProductoViejo[],
	ops: OpsImportacion,
	r: Resultado,
	error: (externo: string, detalle: string) => void,
	ahora: number
) {
	if (!lista.length) return;
	const urls = lista.map((x) => normalizarRuta(x.url)).filter(Boolean);
	const [skus, ya] = await Promise.all([
		ops.skus(),
		db
			.select({ desde: redirecciones.desde })
			.from(redirecciones)
			.where(and(eq(redirecciones.sandboxId, sb), inArray(redirecciones.desde, urls)))
	]);
	const redirigidas = new Set(ya.map((f) => f.desde));
	const nuevas: { desde: string; hacia: string }[] = [];
	for (const x of lista) {
		const desde = normalizarRuta(x.url);
		const sku = (x.producto.sku ?? '').trim();
		if ((desde && redirigidas.has(desde)) || (sku && skus.has(sku))) {
			r.prog.repetidos++;
			continue;
		}
		const e = { ...x.producto };
		const cortes: boolean[] = [];
		let c: boolean;
		[e.nombre, c] = cortar(e.nombre, 120);
		cortes.push(c);
		if (e.descripcion !== undefined) [e.descripcion, c] = cortar(e.descripcion, 4000);
		cortes.push(c);
		if (e.categoria) [e.categoria, c] = cortar(e.categoria, 60);
		cortes.push(c);
		if (e.sku !== undefined) [e.sku, c] = cortar(e.sku, 40);
		cortes.push(c);
		if (e.fotos && e.fotos.length > 8) {
			e.fotos = e.fotos.slice(0, 8);
			cortes.push(true);
		}
		const mal = errorProducto(e);
		if (mal) {
			error(x.externo, mal);
			continue;
		}
		const guardado = await ops.guardarProducto(e);
		if (!guardado) {
			error(x.externo, 'No se pudo guardar el producto.');
			continue;
		}
		r.prog.productos++;
		if (cortes.includes(true)) r.prog.recortados++;
		if (sku) skus.add(sku);
		if (desde) {
			redirigidas.add(desde);
			nuevas.push({ desde, hacia: `/producto/${guardado.id}` });
		}
	}
	await redirigir(db, sb, origen, nuevas, ahora);
	r.prog.redirecciones += nuevas.length;
}

async function clientes(
	db: DbSandbox,
	sb: string,
	origen: string,
	lista: ClienteViejo[],
	r: Resultado,
	error: (externo: string, detalle: string) => void,
	ahora: number
) {
	type Ficha = {
		id: string;
		nombre: string;
		email: string | null;
		telefono: string | null;
		ciudad: string;
		provincia: string;
		notas: string;
		externo: string | null;
	};
	const filas: Ficha[] = [];
	const recortes: boolean[] = [];
	for (const x of lista) {
		const { email, malo } = emailDe(x.email);
		const [cortado, c1] = cortar(x.nombre, 120);
		const nombre = cortado || (email ? email.slice(0, email.indexOf('@')) : '');
		if (!nombre) {
			error(x.externo, 'Un cliente sin nombre ni email no se puede guardar.');
			continue;
		}
		const [telefono, c2] = cortar(x.telefono, 30);
		const [ciudad, c3] = cortar(x.ciudad, 80);
		const [provincia, c4] = cortar(x.provincia, 60);
		const [notas, c5] = cortar(x.notas, 2000);
		const [externo, c6] = cortar(x.externo, 80);
		filas.push({
			id: randomToken(12),
			nombre,
			email,
			telefono: telefono || null,
			ciudad,
			provincia,
			notas,
			externo: externo || null
		});
		recortes.push(malo || c1 || c2 || c3 || c4 || c5 || c6);
	}
	if (!filas.length) return;
	// Cuáles ya estaban (por email o, sin email, por el id de la plataforma vieja).
	const emails = filas.flatMap((f) => (f.email ? [f.email] : []));
	const externos = filas.flatMap((f) => (!f.email && f.externo ? [f.externo] : []));
	const [conEmail, conExterno] = await Promise.all([
		db
			.select({ k: clientesSandbox.email })
			.from(clientesSandbox)
			.where(and(eq(clientesSandbox.sandboxId, sb), inArray(clientesSandbox.email, emails))),
		db
			.select({ k: clientesSandbox.externo })
			.from(clientesSandbox)
			.where(
				and(
					eq(clientesSandbox.sandboxId, sb),
					eq(clientesSandbox.origen, origen),
					inArray(clientesSandbox.externo, externos)
				)
			)
	]);
	const vistos = new Set([...conEmail, ...conExterno].map((f) => f.k));
	const conMail: Ficha[] = [];
	const sinMail: Ficha[] = [];
	filas.forEach((f, i) => {
		const k = f.email ?? f.externo;
		if (k && vistos.has(k)) r.prog.repetidos++;
		else {
			r.prog.clientes++;
			if (recortes[i]) r.prog.recortados++;
		}
		if (k) vistos.add(k);
		(f.email ? conMail : sinMail).push(f);
	});
	const cols = (j: string) => sql.raw(`json_extract(value, '$.${j}')`);
	const insertar = (fs: Ficha[]) => sql`
		insert into clientes (sandbox_id, id, nombre, email, telefono, ciudad, provincia, notas, origen, externo, creado_en)
		select ${sb}, ${cols('id')}, ${cols('nombre')}, ${cols('email')}, ${cols('telefono')}, ${cols('ciudad')},
			${cols('provincia')}, ${cols('notas')}, ${origen}, ${cols('externo')}, ${ahora}
		from json_each(${JSON.stringify(fs)}) where true`;
	// Con email, un upsert que sólo completa lo que falta (nunca pisa nombre ni email).
	if (conMail.length)
		await db.run(sql`${insertar(conMail)}
			on conflict (sandbox_id, email) where email is not null do update set
				telefono = coalesce(clientes.telefono, excluded.telefono),
				ciudad = case when clientes.ciudad = '' then excluded.ciudad else clientes.ciudad end,
				provincia = case when clientes.provincia = '' then excluded.provincia else clientes.provincia end,
				notas = case when clientes.notas = '' then excluded.notas else clientes.notas end`);
	// Sin email, sólo los que no estaban por su id en la plataforma vieja.
	const nuevosSin = sinMail.filter((f) => !f.externo || !conExterno.some((x) => x.k === f.externo));
	const unicos = nuevosSin.filter(
		(f, i) => !f.externo || nuevosSin.findIndex((g) => g.externo === f.externo) === i
	);
	// Cada consulta es idempotente (upsert o "si no estaba"): no hace falta un batch.
	if (unicos.length) await db.run(insertar(unicos));
}

async function pedidos(
	db: DbSandbox,
	sb: string,
	plataforma: Plataforma,
	lista: PedidoViejo[],
	r: Resultado,
	error: (externo: string, detalle: string) => void
) {
	const filas: Record<string, unknown>[] = [];
	const enLaPagina = new Set<string>();
	const recortes: boolean[] = [];
	for (const x of lista) {
		const [externo, c0] = cortar(x.externo, 80);
		if (!externo || !Number.isFinite(x.fecha) || x.fecha <= 0 || !(x.total >= 0)) {
			error(
				x.externo ?? '',
				'Un pedido sin número, sin fecha o con total negativo no se puede guardar.'
			);
			continue;
		}
		if (x.items.some((i) => i.cantidad < 0 || i.precio < 0)) {
			error(externo, 'Un pedido con cantidades o precios negativos no se puede guardar.');
			continue;
		}
		if (enLaPagina.has(externo)) {
			r.prog.repetidos++;
			continue;
		}
		enLaPagina.add(externo);
		const [numero, c1] = cortar(x.numero, 40);
		const [nombre, c2] = cortar(x.nombre, 120);
		const { email, malo } = emailDe(x.email);
		const [estado, c3] = cortar(x.estado, 40);
		let corte = c0 || c1 || c2 || c3 || malo || x.items.length > 100;
		const items = x.items.slice(0, 100).map((i) => {
			const [n, a] = cortar(i.nombre, 120);
			const [v, b] = cortar(i.variante, 120);
			corte ||= a || b;
			return { nombre: n, variante: v, cantidad: i.cantidad, precio: i.precio };
		});
		filas.push({
			id: randomToken(12),
			externo,
			numero,
			fecha: Math.round(x.fecha),
			email,
			nombre,
			total: Math.round(x.total),
			items: JSON.stringify(items),
			estado
		});
		recortes.push(corte);
	}
	if (!filas.length) return;
	const ya = new Set(
		(
			await db
				.select({ k: historialPedidos.externo })
				.from(historialPedidos)
				.where(
					and(
						eq(historialPedidos.sandboxId, sb),
						eq(historialPedidos.plataforma, plataforma),
						inArray(
							historialPedidos.externo,
							filas.map((f) => f.externo)
						)
					)
				)
		).map((f) => f.k)
	);
	filas.forEach((f, i) => {
		if (ya.has(f.externo as string)) r.prog.repetidos++;
		else {
			r.prog.pedidos++;
			if (recortes[i]) r.prog.recortados++;
		}
	});
	const c = (j: string) => sql.raw(`json_extract(value, '$.${j}')`);
	await db.run(sql`
		insert into historial_pedidos (sandbox_id, id, plataforma, externo, numero_externo, fecha, cliente_email,
			cliente_nombre, total, items, estado)
		select ${sb}, ${c('id')}, ${plataforma}, ${c('externo')}, ${c('numero')}, ${c('fecha')}, ${c('email')},
			${c('nombre')}, ${c('total')}, ${c('items')}, ${c('estado')}
		from json_each(${JSON.stringify(filas)}) where true
		on conflict (sandbox_id, plataforma, externo) do nothing`);
}

// ---------------------------------------------------------------- Lectura (clientes y tienda)

export type FichaCliente = typeof clientesSandbox.$inferSelect;
export type FilaHistorial = typeof historialPedidos.$inferSelect;

/** Fichas e historial del sandbox, para unirlos a la lista de Clientes. */
export async function importadosDe(db: DbSandbox, sb: string) {
	const [fichas, historial] = await Promise.all([
		db.select().from(clientesSandbox).where(eq(clientesSandbox.sandboxId, sb)),
		db.select().from(historialPedidos).where(eq(historialPedidos.sandboxId, sb))
	]);
	return { fichas, historial };
}

export function aHistorial(f: FilaHistorial): PedidoHistorial {
	return {
		id: f.id,
		plataforma: f.plataforma as Plataforma,
		numero: f.numeroExterno,
		fecha: new Date(f.fecha).toISOString(),
		total: f.total,
		items: JSON.parse(f.items) as PedidoHistorial['items'],
		estado: f.estado
	};
}

/** Adónde va un path viejo de la tienda del sandbox (relativo a la tienda), o null. */
export async function redireccionDe(db: DbSandbox, sb: string, path: string) {
	const desde = normalizarRuta(path);
	if (!desde) return null;
	const [f] = await db
		.select({ hacia: redirecciones.hacia })
		.from(redirecciones)
		.where(and(eq(redirecciones.sandboxId, sb), eq(redirecciones.desde, desde)))
		.limit(1);
	return f?.hacia ?? null;
}
