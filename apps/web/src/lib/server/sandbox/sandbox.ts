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
import type {
	AccionLote,
	FiltrosLista,
	ListaGenerica,
	NombreLista,
	Producto,
	ProductoEditable,
	ContextoHornero,
	Datos,
	ListaPedidos,
	PedidoDetalle,
	Idea,
	Onboarding,
	Periodo,
	ResumenPanel,
	Tablero,
	Tienda
} from '../datos/tipos.ts';
import { demoDe, type Demo, type EstadoPedido } from './demo.ts';
import {
	DEF_CLIENTES,
	DEF_PRODUCTOS,
	listarGenerico,
	registrosClientes,
	registrosProductos,
	slugProducto,
	type CambioCliente,
	type CambioProducto
} from './listas.ts';
import {
	detalleDemo,
	listarDemo,
	pedidosVivos,
	transicionValida,
	type CambioPedido,
	type FiltrosPedidos
} from './pedidos.ts';
import { resumenDemo } from './resumen.ts';
import { contextoDemo, tableroDemo, type CambiosTablero } from './tablero.ts';
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

/** Borra lo que el visitante cambió (vuelve al dato de ejemplo). */
export async function borrarCambio(
	db: DbSandbox,
	sandboxId: string,
	entidad: string,
	clave: string
) {
	await db
		.delete(cambios)
		.where(
			and(eq(cambios.sandboxId, sandboxId), eq(cambios.entidad, entidad), eq(cambios.clave, clave))
		);
}

const IDEAS: Idea[] = ['carritos', 'stock', 'promo'];

/** Suma un mensaje a Hornero en el día del visitante y devuelve cuántos lleva. */
export async function sumarUsoHornero(db: DbSandbox, sandboxId: string, dia: string) {
	const uso = (await cambiosDe(db, sandboxId, 'hornero')).get('uso') as
		{ dia: string; n: number } | undefined;
	const n = uso?.dia === dia ? uso.n + 1 : 1;
	await guardarCambio(db, sandboxId, 'hornero', 'uso', { dia, n });
	return n;
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

	/** Los datos de ejemplo con lo que cambió el visitante (estado y factura de cada pedido). */
	private async demo(): Promise<Demo> {
		const d = demoDe(this.plantilla(), this.sb.creadoEn)!;
		const c = await this.cambiosPedidos();
		if (!c.size) return d;
		return {
			...d,
			// eslint-disable-next-line @typescript-eslint/no-unused-vars -- se descarta `cambio`
			pedidos: pedidosVivos(d, c, Number.MAX_SAFE_INTEGER).map(({ cambio, ...p }) => p)
		};
	}

	private async cambiosPedidos() {
		const c = await cambiosDe(this.db, this.sb.id, 'pedido');
		return new Map([...c].filter(([, v]) => v).map(([k, v]) => [Number(k), v as CambioPedido]));
	}

	async listarPedidos(f: FiltrosPedidos): Promise<ListaPedidos> {
		const d = demoDe(this.plantilla(), this.sb.creadoEn)!;
		return listarDemo(d, await this.cambiosPedidos(), f, Date.now());
	}

	async pedido(numero: number): Promise<PedidoDetalle | null> {
		const d = demoDe(this.plantilla(), this.sb.creadoEn)!;
		return detalleDemo(d, await this.cambiosPedidos(), numero, Date.now());
	}

	/** Cambia el estado de un pedido (avanzar, deshacer o cancelar). */
	async cambiarEstado(
		numero: number,
		estado: EstadoPedido
	): Promise<'ok' | 'no-existe' | 'no-se-puede'> {
		const d = demoDe(this.plantilla(), this.sb.creadoEn)!;
		const cambios = await this.cambiosPedidos();
		const p = pedidosVivos(d, cambios, Date.now()).find((x) => x.numero === numero);
		if (!p) return 'no-existe';
		if (!transicionValida(p.estado, estado)) return 'no-se-puede';
		await guardarCambio(this.db, this.sb.id, 'pedido', String(numero), {
			...p.cambio,
			estado,
			en: { ...p.cambio.en, [estado]: Date.now() }
		} satisfies CambioPedido);
		return 'ok';
	}

	async pedidosEnLote(numeros: number[], accion: AccionLote) {
		const d = demoDe(this.plantilla(), this.sb.creadoEn)!;
		const cambios = await this.cambiosPedidos();
		const pedidos = new Map(pedidosVivos(d, cambios, Date.now()).map((p) => [p.numero, p]));
		const hechos: { numero: number; antes: EstadoPedido }[] = [];
		const ahora = Date.now();
		for (const n of new Set(numeros)) {
			const p = pedidos.get(n);
			if (!p) continue;
			let cambio: CambioPedido | null = null;
			if (accion === 'preparar' && p.estado === 'Para preparar')
				cambio = {
					...p.cambio,
					estado: 'Listo para despachar',
					en: { ...p.cambio.en, 'Listo para despachar': ahora }
				};
			else if (accion === 'etiquetas' && p.estado === 'Listo para despachar')
				cambio = { ...p.cambio, estado: 'En camino', en: { ...p.cambio.en, 'En camino': ahora } };
			else if (
				accion === 'facturar' &&
				!p.facturado &&
				p.estado !== 'Sin pagar' &&
				p.estado !== 'Cancelado'
			)
				cambio = { ...p.cambio, facturado: true };
			else if (accion === 'avisar' && p.estado !== 'Cancelado')
				cambio = { ...p.cambio, recordado: ahora };
			if (!cambio) continue;
			await guardarCambio(this.db, this.sb.id, 'pedido', String(n), cambio);
			hechos.push({ numero: n, antes: p.estado });
		}
		return { hechos: hechos.length, cambios: hechos };
	}

	async recordarPago(numero: number): Promise<'ok' | 'no-existe' | 'no-se-puede'> {
		const d = demoDe(this.plantilla(), this.sb.creadoEn)!;
		const cambios = await this.cambiosPedidos();
		const p = pedidosVivos(d, cambios, Date.now()).find((x) => x.numero === numero);
		if (!p) return 'no-existe';
		if (p.estado !== 'Sin pagar') return 'no-se-puede';
		await guardarCambio(this.db, this.sb.id, 'pedido', String(numero), {
			...p.cambio,
			recordado: Date.now()
		});
		return 'ok';
	}

	// ---------------------------------------------------------------- Listas

	private async cambiosProductos() {
		const c = await cambiosDe(this.db, this.sb.id, 'producto');
		return new Map([...c].filter(([, v]) => v).map(([k, v]) => [k, v as CambioProducto]));
	}
	private async ajustesStock() {
		const c = await cambiosDe(this.db, this.sb.id, 'stock');
		return new Map(
			[...c].map(([k, v]) => [k, Number((v as { delta: number } | null)?.delta ?? 0)])
		);
	}
	private async regProductos() {
		const d = await this.demo();
		return registrosProductos(
			d,
			Date.now(),
			await this.cambiosProductos(),
			await this.ajustesStock()
		);
	}
	private async regClientes() {
		const d = await this.demo();
		return registrosClientes(d, this.sb.creadoEn, Date.now(), await this.cambiosClientes());
	}
	private async cambiosClientes() {
		const c = await cambiosDe(this.db, this.sb.id, 'cliente');
		return new Map([...c].filter(([, v]) => v).map(([k, v]) => [k, v as CambioCliente]));
	}

	async listar(lista: NombreLista, f: FiltrosLista): Promise<ListaGenerica> {
		if (lista === 'clientes') return listarGenerico(await this.regClientes(), DEF_CLIENTES, f);
		if (lista === 'productos') return listarGenerico(await this.regProductos(), DEF_PRODUCTOS, f);
		return { total: 0, filas: [], segmentos: {}, facetas: {}, resumen: {} };
	}

	async accionLista(lista: NombreLista, ids: string[], accion: string, valor?: unknown) {
		const ahora = Date.now();
		if (lista === 'clientes') {
			const todos = await this.regClientes();
			// La sugerencia de Hornero: cupón a los que están en riesgo.
			const elegidos =
				accion === 'cupon-riesgo'
					? todos.filter((c) => c.seg === 'risk' && !c.cupon)
					: todos.filter((c) => ids.includes(c.id));
			const cambios = await this.cambiosClientes();
			const guardar = (id: string, c: CambioCliente) =>
				guardarCambio(this.db, this.sb.id, 'cliente', id, { ...cambios.get(id), ...c });
			if (accion === 'cupon' || accion === 'cupon-riesgo') {
				for (const c of elegidos) await guardar(c.id, { cupon: ahora });
				return elegidos.length;
			}
			if (accion === 'sin-cupon') {
				for (const c of elegidos) await guardar(c.id, { cupon: undefined });
				return elegidos.length;
			}
			if (accion === 'etiquetar' && typeof valor === 'string' && valor.trim()) {
				const e = valor.trim().slice(0, 30);
				for (const c of elegidos)
					await guardar(c.id, {
						etiquetas: [...new Set([...(cambios.get(c.id)?.etiquetas ?? []), e])]
					});
				return elegidos.length;
			}
			if (accion === 'crear') {
				const v = valor as {
					nombre?: unknown;
					telefono?: unknown;
					email?: unknown;
					ciudad?: unknown;
				} | null;
				const nombre = typeof v?.nombre === 'string' ? v.nombre.trim().slice(0, 80) : '';
				if (nombre.length < 2) return 0;
				const texto = (x: unknown, max: number) =>
					typeof x === 'string' && x.trim() ? x.trim().slice(0, max) : null;
				await guardar(`n${ahora.toString(36)}`, {
					nuevo: {
						nombre,
						telefono: texto(v?.telefono, 30),
						email: texto(v?.email, 120),
						ciudad: texto(v?.ciudad, 60) ?? '',
						creado: ahora
					}
				});
				return 1;
			}
			// Escribir no cambia datos: abre la conversación.
			return elegidos.length;
		}
		if (lista === 'productos') {
			const todos = await this.regProductos();
			const cambios = await this.cambiosProductos();
			const elegidos =
				accion === 'pausar-sin-stock-ml'
					? todos.filter((p) => p.estado === 'out' && p.canales.includes('ml'))
					: todos.filter((p) => ids.includes(p.id));
			if (accion === 'precio' && typeof valor === 'number' && valor > -90 && valor <= 200) {
				for (const p of elegidos) {
					const precio = Math.max(0, Math.round((p.precio * (1 + valor / 100)) / 10) * 10);
					await guardarCambio(this.db, this.sb.id, 'producto', p.id, {
						...cambios.get(p.id),
						precio
					});
				}
				return elegidos.length;
			}
			const publicado =
				accion === 'publicar'
					? Boolean(valor)
					: accion === 'pausar' || accion === 'pausar-sin-stock-ml'
						? false
						: null;
			if (publicado === null) return elegidos.length;
			for (const p of elegidos)
				await guardarCambio(this.db, this.sb.id, 'producto', p.id, {
					...cambios.get(p.id),
					publicado
				});
			return elegidos.length;
		}
		return 0;
	}

	async producto(id: string): Promise<Producto | null> {
		const p = (await this.regProductos()).find((x) => x.id === id);
		if (!p) return null;
		return {
			id: p.id,
			nombre: p.nombre,
			descripcion: p.descripcion,
			categoria: p.categoria === 'Sin categoría' ? null : p.categoria,
			precio: p.precio,
			costo: p.costo,
			sku: p.sku,
			stock: p.stock,
			foto: p.foto,
			canales: Object.fromEntries(
				['web', 'ml', 'ig', 'google'].map((k) => [k, p.canales.includes(k)])
			),
			publicado: p.publicado,
			vendidos30d: p.vendidos30,
			slug: slugProducto(p.nombre)
		};
	}

	async guardarProducto(id: string | null, p: ProductoEditable): Promise<Producto | null> {
		const cambios = await this.cambiosProductos();
		const clave = id ?? `n${Date.now().toString(36)}`;
		if (id && !(await this.regProductos()).some((x) => x.id === id)) return null;
		const antes = cambios.get(clave) ?? {};
		// El stock que se escribe en la ficha reemplaza al de ejemplo y a los ajustes.
		if (p.stock !== undefined) await borrarCambio(this.db, this.sb.id, 'stock', clave);
		await guardarCambio(this.db, this.sb.id, 'producto', clave, {
			...antes,
			nombre: p.nombre,
			descripcion: p.descripcion ?? antes.descripcion,
			categoria: p.categoria ?? antes.categoria,
			precio: p.precio,
			costo: p.costo ?? antes.costo,
			sku: p.sku ?? antes.sku,
			stock: p.stock ?? antes.stock,
			foto: p.foto !== undefined ? p.foto : antes.foto,
			canales: p.canales ?? antes.canales,
			publicado: true,
			...(id ? {} : { nuevo: true, creado: Date.now() })
		} satisfies CambioProducto);
		return this.producto(clave);
	}

	/** La plantilla del sandbox (o la del rubro). */
	private plantilla() {
		return this.sb.plantilla ?? buscarRubro(this.sb.rubro)!.plantilla;
	}

	async resumen(): Promise<ResumenPanel> {
		const d = await this.demo();
		const leidas = (await cambiosDe(this.db, this.sb.id, 'notificaciones')).get('leidas') as
			{ hasta?: number } | undefined;
		return resumenDemo(d, this.sb.creadoEn, leidas?.hasta ?? 0, Date.now());
	}

	async marcarNotificacionesLeidas() {
		await guardarCambio(this.db, this.sb.id, 'notificaciones', 'leidas', { hasta: Date.now() });
	}

	async tablero(periodo: Periodo): Promise<Tablero> {
		const d = await this.demo();
		const c = await cambiosDe(this.db, this.sb.id, 'tablero');
		const cambios: CambiosTablero = {
			meta: (c.get('meta') as { objetivo: number } | undefined)?.objetivo,
			certificadoOculto: c.has('certificado'),
			ideasHechas: await this.ideasHechas()
		};
		// Quien prueba la demo es la dueña de la tienda de ejemplo (la del onboarding).
		return tableroDemo(d, periodo, {
			base: this.sb.creadoEn,
			ahora: Date.now(),
			hola: 'Valeria',
			cambios
		});
	}

	async guardarMeta(objetivo: number) {
		await guardarCambio(this.db, this.sb.id, 'tablero', 'meta', { objetivo });
	}

	async ocultarAviso(aviso: 'certificado') {
		await guardarCambio(this.db, this.sb.id, 'tablero', aviso, true);
	}

	private async ideasHechas() {
		const c = await cambiosDe(this.db, this.sb.id, 'hornero');
		return IDEAS.filter((k) => c.get(k));
	}

	async recordarCarritos() {
		const d = await this.demo();
		if ((await this.ideasHechas()).includes('carritos')) return 0;
		await guardarCambio(this.db, this.sb.id, 'hornero', 'carritos', { en: Date.now() });
		return d.carritos.length;
	}

	async marcarIdea(idea: Idea, hecha: boolean) {
		if (hecha) await guardarCambio(this.db, this.sb.id, 'hornero', idea, { en: Date.now() });
		else await borrarCambio(this.db, this.sb.id, 'hornero', idea);
	}

	async contextoHornero(): Promise<ContextoHornero> {
		return contextoDemo(await this.demo(), Date.now());
	}
}
