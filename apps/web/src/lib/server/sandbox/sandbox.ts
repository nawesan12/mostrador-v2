/**
 * Sandbox de la demo pública. Cada visitante tiene uno (cookie), que vence a las 24 h.
 *
 * "Copia al escribir": la tienda de ejemplo de cada rubro sale del código, igual para todos; D1
 * guarda sólo el sandbox (una fila) y lo que el visitante cambie (tabla `cambios`). Así crear una
 * demo cuesta una escritura y no cientos, y el cupo gratis de D1 (100.000 escrituras por día)
 * alcanza para muchas visitas.
 */
import { and, eq, lt } from 'drizzle-orm';
import { drizzle, type DrizzleD1Database } from 'drizzle-orm/d1';
import { inArray } from './sql.ts';
import { fromB64url, randomToken, sha256 } from '../crypto.ts';
import { rubro as buscarRubro, slugDe } from './rubros.ts';
import { cambios, sandboxes } from './schema.ts';
import type { Datos, Onboarding, ResumenPanel, Tienda } from '../datos/tipos.ts';
import { demoDe } from './demo.ts';
import { resumenDemo } from './resumen.ts';
import RUBROS from '@mostrador/contrato/rubros.json';
import { plantilla as buscarPlantilla } from '#lib/plantillas.ts';

export type DbSandbox = DrizzleD1Database<Record<string, never>>;

export const COOKIE = 'mostrador_demo';
export const VIDA_MS = 24 * 60 * 60 * 1000;

export const dbSandbox = (d1: D1Database): DbSandbox => drizzle(d1);

export type Sandbox = {
	id: string;
	rubro: string;
	plantilla: string | null;
	creadoEn: number;
	venceEn: number;
};

/**
 * Crea un sandbox del rubro, con su plantilla por defecto o la elegida en la galería (que tiene que
 * ser de ese rubro). Devuelve null si el rubro o la plantilla no existen.
 */
export async function crear(
	db: DbSandbox,
	rubro: string,
	ahora = Date.now(),
	plantilla: string | null = null
): Promise<Sandbox | null> {
	const r = buscarRubro(rubro);
	if (!r) return null;
	if (plantilla && buscarPlantilla(plantilla)?.rubro !== r.nombre) return null;
	const sb = { id: randomToken(24), rubro, plantilla, creadoEn: ahora, venceEn: ahora + VIDA_MS };
	await db.insert(sandboxes).values(sb);
	return sb;
}

/** El sandbox de la cookie, si existe y no venció. */
export async function leer(db: DbSandbox, id: string | undefined, ahora = Date.now()) {
	if (!id || id.length > 64) return null;
	const [sb] = await db.select().from(sandboxes).where(eq(sandboxes.id, id)).limit(1);
	return sb && sb.venceEn > ahora ? sb : null;
}

/**
 * Borra los sandboxes vencidos y sus cambios. Va de a tandas chicas para no pasar el tope de CPU ni
 * de parámetros de D1; lo llama el cron.
 */
export async function purgar(db: DbSandbox, ahora = Date.now(), tanda = 50) {
	const viejos = await db
		.select({ id: sandboxes.id })
		.from(sandboxes)
		.where(lt(sandboxes.venceEn, ahora))
		.limit(tanda);
	if (!viejos.length) return 0;
	const ids = viejos.map((v) => v.id);
	await db.batch([
		db.delete(cambios).where(inArray(cambios.sandboxId, ids)),
		db.delete(sandboxes).where(inArray(sandboxes.id, ids))
	]);
	return ids.length;
}

/** Lo que el visitante cambió de una entidad, por clave. `null` = borrado. */
async function cambiosDe(db: DbSandbox, sandboxId: string, entidad: string) {
	const filas = await db
		.select()
		.from(cambios)
		.where(and(eq(cambios.sandboxId, sandboxId), eq(cambios.entidad, entidad)));
	return new Map(filas.map((f) => [f.clave, f.borrado ? null : JSON.parse(f.datos ?? 'null')]));
}

/** uuid estable a partir de un texto (para ids de ejemplo que no cambian entre pedidos). */
async function uuidDe(texto: string) {
	const h = [...fromB64url(await sha256(texto))]
		.map((b) => b.toString(16).padStart(2, '0'))
		.join('');
	return `${h.slice(0, 8)}-${h.slice(8, 12)}-4${h.slice(13, 16)}-8${h.slice(17, 20)}-${h.slice(20, 32)}`;
}

/** Guarda (o reemplaza) lo que el visitante cambió de una entidad. */
export async function guardarCambio(
	db: DbSandbox,
	sandboxId: string,
	entidad: string,
	clave: string,
	datos: unknown,
	ahora = Date.now()
) {
	const fila = {
		sandboxId,
		entidad,
		clave,
		datos: JSON.stringify(datos),
		borrado: false,
		actualizadoEn: ahora
	};
	await db
		.insert(cambios)
		.values(fila)
		.onConflictDoUpdate({
			target: [cambios.sandboxId, cambios.entidad, cambios.clave],
			set: { datos: fila.datos, borrado: false, actualizadoEn: ahora }
		});
}

export class DatosSandbox implements Datos {
	constructor(
		private db: DbSandbox,
		private sb: Sandbox
	) {}

	async rubros() {
		return RUBROS;
	}

	async tienda(): Promise<Tienda> {
		const r = buscarRubro(this.sb.rubro)!;
		// La tienda de ejemplo es la de la plantilla elegida (o la del rubro).
		const p = buscarPlantilla(this.sb.plantilla ?? r.plantilla);
		const nombre = p?.tienda ?? r.tienda;
		const base: Tienda = {
			id: await uuidDe(`${this.sb.id}:tienda`),
			slug: slugDe(nombre),
			nombre,
			rubro: r.id,
			plantilla: p?.id ?? r.plantilla,
			plan: 'gratis',
			creadaEn: new Date(this.sb.creadoEn).toISOString()
		};
		const cambio = (await cambiosDe(this.db, this.sb.id, 'tienda')).get('tienda');
		return cambio ? { ...base, ...cambio } : base;
	}

	/** La plantilla ya la fija el sandbox al crearse; acá queda el resto de lo elegido. */
	async guardarOnboarding(o: Onboarding) {
		await guardarCambio(this.db, this.sb.id, 'onboarding', 'estado', {
			elecciones: o.elecciones,
			terminado: o.terminado ?? false
		});
	}

	/** La plantilla del sandbox (o la del rubro). */
	private plantilla() {
		return this.sb.plantilla ?? buscarRubro(this.sb.rubro)!.plantilla;
	}

	async resumen(): Promise<ResumenPanel> {
		const d = demoDe(this.plantilla(), this.sb.creadoEn)!;
		const leidas = (await cambiosDe(this.db, this.sb.id, 'notificaciones')).get('leidas') as
			{ hasta?: number } | undefined;
		return resumenDemo(d, this.sb.creadoEn, leidas?.hasta ?? 0, Date.now());
	}

	async marcarNotificacionesLeidas() {
		await guardarCambio(this.db, this.sb.id, 'notificaciones', 'leidas', { hasta: Date.now() });
	}
}
