/**
 * Sandbox de la demo pública. Cada visitante tiene uno (cookie), que vence a las 24 h.
 *
 * "Copia al escribir": la tienda de ejemplo de cada rubro sale del código, igual para todos; D1
 * guarda sólo el sandbox (una fila) y lo que el visitante cambie (tabla `cambios`). Así crear una
 * demo cuesta una escritura y no cientos, y el cupo gratis de D1 (100.000 escrituras por día)
 * alcanza para muchas visitas.
 */
import { PERMISOS, ROLES, fijo, type Rol } from '#lib/panel/permisos.ts';
import {
	configuracionDemo,
	type EstadoCuenta,
	type Invitado,
	type Negocio,
	type Plan
} from './cuenta.ts';
import { finanzasDemo, type Retiro } from './finanzas.ts';
import { actividadDemo, type FilaCambio, type Nombres } from './actividad.ts';
import { canalesDemo, informesDemo, marketingDemo, type CampanaDemo } from './crecer.ts';
import { cuando } from '#lib/panel/cupones.ts';
import { descripcionDe, lunes9, vigenciaDe } from '#lib/panel/cupones.ts';
import { cuponesDemo, type CambioCupon, type EstadoRelampago } from './cupones.ts';
import { aumentar, proximoLunes, type Redondeo } from '#lib/panel/precios.ts';
import { and, desc, eq, lt } from 'drizzle-orm';
import { drizzle, type DrizzleD1Database } from 'drizzle-orm/d1';
import { inArray } from './sql.ts';
import { fromB64url, randomToken, sha256 } from '../crypto.ts';
import { rubro as buscarRubro, slugDe } from './rubros.ts';
import { cambios, sandboxes } from './schema.ts';
import {
	estadoInicial,
	INTEGRACIONES,
	type AjusteIntegracion,
	type EstadoIntegraciones
} from '#lib/panel/integraciones.ts';
import type {
	ConexionML,
	ConexionMP,
	PublicacionML,
	Dominios,
	MisDatos,
	PreferenciasAvisos,
	Permisos,
	PlanCuenta,
	Configuracion,
	Miembro,
	DisenoTienda,
	MiTienda,
	Finanzas,
	Evento,
	ParaHacer,
	Campana,
	Canales,
	Informes,
	Marketing,
	Cupon,
	Cupones,
	CuponNuevo as CuponNuevoT,
	ListasPrecio,
	AccionLote,
	PedidoNuevo,
	AutomaticosCarritos,
	Resultado,
	MovimientoStock,
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
import { demoDe, type Demo, type EstadoPedido, type Pedido } from './demo.ts';
import {
	COBRADO,
	devolucionesDemo,
	resenasDemo,
	todasLasResenas,
	type CambiosResenas,
	type EstadoDevolucion,
	type PedidoNuevoDemo,
	type PreguntaTiendaDemo,
	type ResenaNuevaDemo
} from './resenas.ts';
import { codigoPedido } from '../tienda/pedido.ts';
import {
	DEF_CLIENTES,
	DEF_ENVIOS,
	DEF_FACTURACION,
	DEF_INVENTARIO,
	DEF_PRODUCTOS,
	listarGenerico,
	registrosClientes,
	registrosEnvios,
	registrosFacturacion,
	registrosInventario,
	registrosProductos,
	slugProducto,
	type CambioCliente,
	type CambioFactura,
	type CambioProducto,
	type FacturaManual,
	type Movimiento
} from './listas.ts';
import { inicioDelDia } from './tablero.ts';
import {
	bandejaDemo,
	carritosDemo,
	conversacionDemo,
	type CambioConversacion,
	type EstadoCarritos
} from './ventas.ts';
import {
	detalleDemo,
	listarDemo,
	pedidosVivos,
	transicionValida,
	type CambioPedido,
	type FiltrosPedidos
} from './pedidos.ts';
import { DIAS_DE_PRUEBA, resumenDemo } from './resumen.ts';
import { COSTO_MENSAJE, contextoDemo, tableroDemo, type CambiosTablero } from './tablero.ts';
import RUBROS from '@mostrador/contrato/rubros.json';
import { plantilla as buscarPlantilla } from '#lib/plantillas.ts';
import { SITE } from '#lib/site.ts';
import type { TiendaPublica } from '#lib/tienda/tipos.ts';

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
export async function cambiosDe(db: DbSandbox, sandboxId: string, entidad: string) {
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

/** Un link de pago de la demo. */
export type LinkDemo = {
	monto: number;
	concepto: string;
	cuotas: boolean;
	pedido: number | null;
	creado: number;
	pagado: number | null;
};

const ALFABETO = '23456789ABCDEFGHJKLMNPQRSTUVWXYZ';
/** Código corto y sin letras que se confunden (0/O, 1/I). */
function codigo() {
	const b = crypto.getRandomValues(new Uint8Array(6));
	return [...b].map((x) => ALFABETO[x % ALFABETO.length]).join('');
}

/** El link por su código (de cualquier visitante: lo abre quien paga, sin cookie). */
export async function buscarLink(db: DbSandbox, cod: string, ahora = Date.now()) {
	if (!/^[2-9A-Z]{6}$/.test(cod)) return null;
	const [f] = await db
		.select()
		.from(cambios)
		.where(and(eq(cambios.entidad, 'link'), eq(cambios.clave, cod)))
		.limit(1);
	if (!f || f.borrado) return null;
	const sb = await leer(db, f.sandboxId, ahora);
	if (!sb) return null;
	return { sb, link: JSON.parse(f.datos ?? 'null') as LinkDemo };
}

/**
 * Un cambio buscado por su clave en todos los sandboxes (el link de pago o el chat de Telegram no
 * traen la cookie de la demo). Sólo devuelve sandboxes que no vencieron.
 */
export async function buscarCambio<T>(
	db: DbSandbox,
	entidad: string,
	clave: string,
	ahora = Date.now()
) {
	const [f] = await db
		.select()
		.from(cambios)
		.where(and(eq(cambios.entidad, entidad), eq(cambios.clave, clave)))
		.limit(1);
	if (!f || f.borrado) return null;
	const sb = await leer(db, f.sandboxId, ahora);
	if (!sb) return null;
	return { sb, datos: JSON.parse(f.datos ?? 'null') as T };
}

/** El sábado que viene a las 10 de Argentina (el mejor horario para campañas). */
function sabado10(ahora: number) {
	const ar = new Date(ahora - 3 * 3_600_000);
	const dias = (6 - ar.getUTCDay() + 7) % 7 || 7;
	return Date.UTC(ar.getUTCFullYear(), ar.getUTCMonth(), ar.getUTCDate() + dias, 13);
}

/** Las secciones de la portada y los colores, como vienen en la plantilla. */
const DISENO_INICIAL: DisenoTienda = {
	secciones: (['hero', 'cats', 'prods', 'revs', 'ig', 'wa'] as const).map((id) => ({
		id,
		visible: true
	})),
	paleta: 0
};

const NOMBRE_PLAN: Record<Plan, string> = {
	gratis: 'Gratis',
	emprende: 'Emprende',
	pro: 'Pro',
	negocio: 'Negocio'
};

type DominioDemo = { nombre: string; en: number };
export type TipoExport =
	'productos' | 'clientes' | 'pedidos' | 'facturas' | 'resenas' | 'conversaciones';

type AumentoProgramado = { pct: number; redondeo: Redondeo; para: number };

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
	/**
	 * Los datos de ejemplo más los pedidos que cargó el visitante (punto de venta, crear pedido), sin
	 * los cambios de estado. Los pedidos nuevos traen su comprador como cliente.
	 */
	private async base(): Promise<Demo> {
		const d = demoDe(this.plantilla(), this.sb.creadoEn)!;
		const nuevos = [...(await cambiosDe(this.db, this.sb.id, 'pedido-nuevo'))]
			.map(([, v]) => v as PedidoNuevoDemo | null)
			.filter((x): x is PedidoNuevoDemo => !!x)
			.sort((a, b) => b.pedido.numero - a.pedido.numero);
		if (!nuevos.length) return d;
		return {
			...d,
			pedidos: [...nuevos.map((x) => x.pedido), ...d.pedidos],
			clientes: [...d.clientes, ...nuevos.filter((x) => x.cliente).map((x) => x.cliente!)]
		};
	}

	/** Los datos de ejemplo con lo que cambió el visitante (pedidos nuevos, estados y facturas). */
	private async demo(): Promise<Demo> {
		const d = await this.base();
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
		const d = await this.base();
		return listarDemo(d, await this.cambiosPedidos(), f, Date.now());
	}

	async pedido(numero: number): Promise<PedidoDetalle | null> {
		const d = await this.base();
		return detalleDemo(d, await this.cambiosPedidos(), numero, Date.now());
	}

	/** Cambia el estado de un pedido (avanzar, deshacer o cancelar). */
	async cambiarEstado(
		numero: number,
		estado: EstadoPedido
	): Promise<'ok' | 'no-existe' | 'no-se-puede'> {
		const d = await this.base();
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
		const d = await this.base();
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
		const d = await this.base();
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
		if (lista === 'inventario')
			return listarGenerico(registrosInventario(await this.regProductos()), DEF_INVENTARIO, f);
		if (lista === 'facturacion')
			return listarGenerico(await this.regFacturacion(), DEF_FACTURACION, f);
		return listarGenerico(await this.regEnvios(), DEF_ENVIOS, f);
	}

	private async regFacturacion() {
		const d = await this.demo();
		const c = await cambiosDe(this.db, this.sb.id, 'factura');
		const cambios = new Map([...c].filter(([, v]) => v).map(([k, v]) => [k, v as CambioFactura]));
		const manuales = [...(await cambiosDe(this.db, this.sb.id, 'factura-manual'))]
			.filter(([, v]) => v)
			.map(([k, v]) => [k, v as FacturaManual] as [string, FacturaManual])
			.sort((a, b) => a[1].en - b[1].en);
		return registrosFacturacion(d, Date.now(), cambios, manuales);
	}

	private async regEnvios() {
		const d = await this.base();
		const vivos = pedidosVivos(d, await this.cambiosPedidos(), Date.now());
		const reclamados = new Set((await cambiosDe(this.db, this.sb.id, 'reclamo')).keys());
		return registrosEnvios(vivos, d.clientes, Date.now(), reclamados);
	}

	/** Movimientos de stock de hoy: ventas (de los pedidos) y ajustes, conteos e ingresos. */
	async movimientosHoy(): Promise<MovimientoStock[]> {
		const d = await this.demo();
		const ahora = Date.now();
		const desde = inicioDelDia(ahora);
		const nombres = new Map((await this.regProductos()).map((p) => [p.id, p.nombre]));
		const ventas: MovimientoStock[] = d.pedidos
			.filter((p) => p.estado !== 'Cancelado' && p.creado >= desde && p.creado <= ahora)
			.flatMap((p) =>
				p.items.map((i) => ({
					tipo: p.canal === 'pos' ? 'local' : p.canal === 'ml' ? 'ml' : 'web',
					texto: `${p.canal === 'pos' ? 'Venta en local' : p.canal === 'ml' ? `Mercado Libre #${p.numero}` : `Venta #${p.numero}`} · ${i.nombre}`,
					cantidad: -i.cantidad,
					en: new Date(p.creado).toISOString()
				}))
			);
		const propios = [...(await cambiosDe(this.db, this.sb.id, 'movimiento'))]
			.map(([, v]) => v as Movimiento | null)
			.filter((m): m is Movimiento => !!m && m.en >= desde)
			.map((m): MovimientoStock => ({
				tipo: m.motivo,
				texto: `${m.motivo === 'ingreso' ? 'Ingreso de proveedor' : m.motivo === 'conteo' ? 'Conteo' : 'Ajuste manual'} · ${nombres.get(m.productoId) ?? 'Producto'}`,
				cantidad: m.q,
				en: new Date(m.en).toISOString()
			}));
		return [...ventas, ...propios].sort((a, b) => b.en.localeCompare(a.en)).slice(0, 20);
	}

	private async ajustar(productoId: string, q: number, motivo: Movimiento['motivo']) {
		const actual = (await this.ajustesStock()).get(productoId) ?? 0;
		const ahora = Date.now();
		await guardarCambio(this.db, this.sb.id, 'stock', productoId, { delta: actual + q });
		await guardarCambio(this.db, this.sb.id, 'movimiento', `${ahora}-${productoId}`, {
			productoId,
			q,
			motivo,
			en: ahora
		} satisfies Movimiento);
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
		if (lista === 'inventario') {
			const productos = await this.regProductos();
			if (
				accion === 'ajustar' &&
				typeof valor === 'number' &&
				Number.isInteger(valor) &&
				Math.abs(valor) <= 10_000
			) {
				let n = 0;
				for (const p of productos.filter((x) => ids.includes(x.id))) {
					// No puede quedar stock negativo.
					const q = Math.max(-p.stock, valor);
					if (!q) continue;
					await this.ajustar(p.id, q, 'ajuste');
					n++;
				}
				return n;
			}
			if (accion === 'conteo' && Array.isArray(valor)) {
				let n = 0;
				const stock = new Map(productos.map((p) => [p.id, p.stock]));
				for (const c of valor as { id?: unknown; cantidad?: unknown }[]) {
					const id = String(c?.id ?? '');
					const cantidad = Number(c?.cantidad);
					if (!stock.has(id) || !Number.isInteger(cantidad) || cantidad < 0 || cantidad > 1_000_000)
						continue;
					const q = cantidad - stock.get(id)!;
					if (!q) continue;
					await this.ajustar(id, q, 'conteo');
					n++;
				}
				return n;
			}
			if (accion === 'ingreso' && typeof valor === 'number' && valor > 0 && valor <= 100_000) {
				for (const id of ids) await this.ajustar(id, Math.round(valor), 'ingreso');
				return ids.length;
			}
			// Pedir al proveedor, mover e imprimir códigos no cambian el stock.
			return productos.filter((p) => ids.includes(p.id)).length;
		}
		if (lista === 'facturacion') {
			const todos = await this.regFacturacion();
			const c = await cambiosDe(this.db, this.sb.id, 'factura');
			const guardar = (id: string, v: CambioFactura) =>
				guardarCambio(this.db, this.sb.id, 'factura', id, {
					...(c.get(id) as CambioFactura | null),
					...v
				});
			const elegidos = todos.filter((f) => ids.includes(f.id));
			if (accion === 'enviar') {
				for (const f of elegidos.filter((x) => x.estado === 'ok'))
					await guardar(f.id, { enviado: true });
				return elegidos.filter((x) => x.estado === 'ok').length;
			}
			if (accion === 'reintentar') {
				const a = elegidos.filter((x) => x.estado === 'pend' || x.estado === 'rej');
				for (const f of a) await guardar(f.id, { autorizada: true });
				return a.length;
			}
			if (accion === 'emitir') {
				const v = valor as Partial<FacturaManual> | null;
				const cliente = typeof v?.cliente === 'string' ? v.cliente.trim().slice(0, 80) : '';
				const total = Number(v?.total);
				const tipo = v?.tipo === 'A' ? 'A' : 'B';
				const cuit = typeof v?.cuit === 'string' ? v.cuit.replace(/\D/g, '') : '';
				if (cliente.length < 2 || !Number.isInteger(total) || total <= 0 || total > 1_000_000_000)
					return 0;
				if (tipo === 'A' && cuit.length !== 11) return 0;
				await guardarCambio(this.db, this.sb.id, 'factura-manual', `M-${ahora.toString(36)}`, {
					cliente,
					tipo,
					cuit: tipo === 'A' ? cuit : null,
					concepto: typeof v?.concepto === 'string' ? v.concepto.slice(0, 200) : '',
					total,
					en: ahora
				} satisfies FacturaManual);
				return 1;
			}
			return elegidos.length;
		}
		if (lista === 'envios') {
			const todos = await this.regEnvios();
			const elegidos =
				accion === 'reclamar-todos'
					? todos.filter((e) => e.estado === 'late' && !e.reclamado)
					: accion === 'etiquetas' && !ids.length
						? todos.filter((e) => e.estado === 'todo')
						: todos.filter((e) => ids.includes(e.id));
			if (accion === 'reclamar' || accion === 'reclamar-todos') {
				for (const e of elegidos)
					await guardarCambio(this.db, this.sb.id, 'reclamo', e.id, { en: ahora });
				return elegidos.length;
			}
			if (accion === 'etiquetas') {
				let n = 0;
				for (const e of elegidos.filter((x) => x.estado === 'todo'))
					if ((await this.cambiarEstado(Number(e.id), 'En camino')) === 'ok') n++;
				return n;
			}
			if (accion === 'sin-etiqueta') {
				let n = 0;
				for (const e of elegidos)
					if ((await this.cambiarEstado(Number(e.id), 'Listo para despachar')) === 'ok') n++;
				return n;
			}
			// Pedir retiro y avisar esperan a los transportes y a WhatsApp.
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

	// ---------------------------------------------------------------- Bandeja y carritos

	private async cambiosConversaciones() {
		const [c, leidas] = await Promise.all([
			cambiosDe(this.db, this.sb.id, 'conversacion'),
			cambiosDe(this.db, this.sb.id, 'conversacion-leida')
		]);
		const m = new Map([...c].filter(([, v]) => v).map(([k, v]) => [k, v as CambioConversacion]));
		for (const [k, v] of leidas) if (v) m.set(k, { ...m.get(k), leida: true });
		return m;
	}

	async bandeja() {
		return bandejaDemo(await this.demo(), await this.cambiosConversaciones());
	}

	async conversacion(id: string) {
		return conversacionDemo(await this.demo(), id, await this.cambiosConversaciones());
	}

	async responder(
		id: string,
		c: { texto?: string; tomada?: boolean; leida?: boolean }
	): Promise<Resultado> {
		const cambios = await this.cambiosConversaciones();
		if (!conversacionDemo(await this.demo(), id, cambios)) return 'no-existe';
		// "Leída" va aparte: llega junto con el primer mensaje y no tiene que pisarlo.
		if (c.leida && !c.texto?.trim() && c.tomada === undefined) {
			await guardarCambio(this.db, this.sb.id, 'conversacion-leida', id, true);
			return 'ok';
		}
		const antes = cambios.get(id) ?? {};
		const texto = c.texto?.trim().slice(0, 2000);
		await guardarCambio(this.db, this.sb.id, 'conversacion', id, {
			...antes,
			...(texto
				? { mensajes: [...(antes.mensajes ?? []), { texto, en: Date.now() }], leida: true }
				: {}),
			...(c.tomada !== undefined ? { tomada: c.tomada } : {}),
			...(c.leida ? { leida: true } : {})
		} satisfies CambioConversacion);
		return 'ok';
	}

	private async estadoCarritos(): Promise<EstadoCarritos> {
		const c = await cambiosDe(this.db, this.sb.id, 'carrito');
		const auto = (await cambiosDe(this.db, this.sb.id, 'carritos')).get('automaticos') as
			EstadoCarritos['automaticos'] | undefined;
		return {
			recordados: new Set([...c].filter(([, v]) => v).map(([k]) => k)),
			todos: (await this.ideasHechas()).includes('carritos'),
			automaticos: auto ?? { hora: true, dia: true, email: false }
		};
	}

	async carritos() {
		return carritosDemo(await this.demo(), Date.now(), await this.estadoCarritos(), 120);
	}

	async guardarAutomaticos(a: AutomaticosCarritos) {
		await guardarCambio(this.db, this.sb.id, 'carritos', 'automaticos', {
			hora: !!a.hora,
			dia: !!a.dia,
			email: !!a.email
		});
	}

	async recordarCarrito(id: string): Promise<Resultado> {
		const d = await this.demo();
		if (!d.carritos.some((k) => k.id === id)) return 'no-existe';
		const e = await this.estadoCarritos();
		if (e.todos || e.recordados.has(id)) return 'no-se-puede';
		await guardarCambio(this.db, this.sb.id, 'carrito', id, { en: Date.now() });
		return 'ok';
	}

	// ---------------------------------------------------------------- Pedidos nuevos

	async crearPedido(
		n: PedidoNuevo
	): Promise<{ numero: number; total: number } | 'sin-stock' | 'invalido'> {
		const d = await this.base();
		const productos = new Map((await this.regProductos()).map((p) => [p.id, p]));
		const items: Pedido['items'] = [];
		for (const i of n.items) {
			const p = productos.get(i.productoId);
			if (!p || !Number.isInteger(i.cantidad) || i.cantidad < 1) return 'invalido';
			if (p.stock < i.cantidad) return 'sin-stock';
			items.push({ productoId: p.id, nombre: p.nombre, cantidad: i.cantidad, precio: p.precio });
		}
		if (!items.length) return 'invalido';
		const ahora = Date.now();
		const numero = Math.max(...d.pedidos.map((p) => p.numero)) + 1;
		const nombre = n.cliente?.trim().slice(0, 80);
		// Un cliente que ya existe: el pedido se suma a su ficha.
		const existente = n.clienteId ? d.clientes.find((c) => c.id === n.clienteId) : undefined;
		const cliente: Demo['clientes'][number] | null = existente
			? null
			: {
					id: `v${numero}`,
					nombre: nombre || 'Consumidor final',
					email: null,
					telefono: n.telefono?.trim() || null,
					ciudad: d.catalogo.ciudad,
					foto: null,
					mayorista: false,
					alta: ahora
				};
		const envio = n.envio ?? 'retiro';
		const pedido: Pedido = {
			numero,
			clienteId: existente?.id ?? cliente!.id,
			canal: n.canal,
			pago: n.pago,
			estadoPago:
				n.pagado === false
					? 'Pendiente'
					: n.pago === 'efectivo'
						? 'Cobrado'
						: n.pago === 'transf'
							? 'Acreditada'
							: 'Aprobado',
			envio,
			ciudadEnvio: d.catalogo.ciudad,
			estado:
				n.pagado === false
					? 'Sin pagar'
					: n.canal === 'pos' && envio === 'retiro'
						? 'Entregado'
						: 'Para preparar',
			facturado: n.pagado !== false,
			items,
			total: items.reduce((s, i) => s + i.precio * i.cantidad, 0),
			creado: ahora
		};
		await guardarCambio(this.db, this.sb.id, 'pedido-nuevo', String(numero), {
			pedido,
			cliente
		} satisfies PedidoNuevoDemo);
		// La venta descuenta stock (el movimiento lo muestra el pedido).
		const ajustes = await this.ajustesStock();
		for (const i of items)
			await guardarCambio(this.db, this.sb.id, 'stock', i.productoId, {
				delta: (ajustes.get(i.productoId) ?? 0) - i.cantidad
			});
		return { numero, total: pedido.total };
	}

	// ---------------------------------------------------------------- Cupones

	private async cuponesNuevos() {
		return [...(await cambiosDe(this.db, this.sb.id, 'cupon-nuevo'))]
			.filter(([, v]) => v)
			.map(([, v]) => v as Cupon)
			.sort((a, b) => Date.parse(b.creado) - Date.parse(a.creado));
	}

	async cupones(): Promise<Cupones> {
		const d = await this.demo();
		const [cambios, nuevos, consejo, autos, rel, clientes] = await Promise.all([
			cambiosDe(this.db, this.sb.id, 'cupon'),
			this.cuponesNuevos(),
			cambiosDe(this.db, this.sb.id, 'cupones'),
			cambiosDe(this.db, this.sb.id, 'automatica'),
			cambiosDe(this.db, this.sb.id, 'relampago'),
			this.regClientes()
		]);
		return cuponesDemo(
			d,
			Date.now(),
			this.sb.creadoEn,
			new Map([...cambios].filter(([, v]) => v).map(([k, v]) => [k, v as CambioCupon])),
			nuevos,
			!!consejo.get('consejo'),
			new Map(
				[...autos].filter(([, v]) => v).map(([k, v]) => [k, (v as { activa: boolean }).activa])
			),
			(rel.get('vivo') as EstadoRelampago | undefined) ?? {
				horas: 0,
				liberadas: 0,
				terminada: false
			},
			{
				vip: clientes.filter((c) => c.seg === 'vip').length,
				riesgo: clientes.filter((c) => c.seg === 'risk').length
			}
		);
	}

	async crearCupon(n: CuponNuevoT): Promise<Cupon | 'repetido'> {
		const todos = await this.cupones();
		if (todos.cupones.some((c) => c.codigo === n.codigo && c.estado !== 'terminado'))
			return 'repetido';
		const ahora = Date.now();
		const v = vigenciaDe(n.vigencia, ahora);
		const empieza = n.programar ? lunes9(ahora) : v.empieza;
		const nombres = todos.productos.filter((p) => n.productos?.includes(p.id)).map((p) => p.nombre);
		const c: Cupon = {
			id: `n${ahora.toString(36)}`,
			codigo: n.codigo,
			tipo: n.tipo,
			valor: n.tipo === 'pct' || n.tipo === 'fijo' ? n.valor : 0,
			descripcion: descripcionDe(n, todos.categoria, nombres),
			estado: empieza && empieza > ahora ? 'programado' : 'activo',
			usos: 0,
			limite: n.limite,
			vendio: 0,
			desconto: 0,
			canales: n.canales,
			empieza: empieza ? new Date(empieza).toISOString() : null,
			termina: v.termina
				? new Date(Math.max(v.termina, (empieza ?? 0) + 86_400_000)).toISOString()
				: null,
			minimo: n.minimo,
			consejo: n.opciones.hornero ? 'Hornero ya lo puede ofrecer en el chat.' : null,
			creado: new Date(ahora).toISOString()
		};
		await guardarCambio(this.db, this.sb.id, 'cupon-nuevo', c.id, c);
		return c;
	}

	async accionCupon(
		id: string,
		accion: 'pausar' | 'activar' | 'duplicar' | 'minimo' | 'deshacer-minimo',
		minimo = 0
	): Promise<Cupon | 'no-existe' | 'no-se-puede'> {
		const c = (await this.cupones()).cupones.find((x) => x.id === id);
		if (!c) return 'no-existe';
		const nuevo = id.startsWith('n');
		const guardar = async (cambio: CambioCupon) => {
			if (nuevo) await guardarCambio(this.db, this.sb.id, 'cupon-nuevo', id, { ...c, ...cambio });
			else {
				const antes = (await cambiosDe(this.db, this.sb.id, 'cupon')).get(id) as
					CambioCupon | undefined;
				await guardarCambio(this.db, this.sb.id, 'cupon', id, { ...antes, ...cambio });
			}
			return { ...c, ...cambio };
		};
		if (accion === 'duplicar') {
			const copia: Cupon = {
				...c,
				id: `n${Date.now().toString(36)}`,
				codigo: `${c.codigo}-2`.slice(0, 20),
				estado: 'pausado',
				usos: 0,
				vendio: 0,
				desconto: 0,
				consejo: null,
				empieza: null,
				termina: null,
				creado: new Date().toISOString()
			};
			await guardarCambio(this.db, this.sb.id, 'cupon-nuevo', copia.id, copia);
			return copia;
		}
		if (c.estado === 'terminado') return 'no-se-puede';
		if (accion === 'pausar') return guardar({ estado: 'pausado' });
		if (accion === 'activar') return guardar({ estado: 'activo' });
		if (accion === 'minimo')
			return guardar({
				minimo,
				descripcion: `${c.descripcion.replace(/ · mín\. .*$/, '')} · mín. $ ${minimo.toLocaleString('es-AR')}`,
				consejo: null
			});
		return guardar({ minimo: 0, descripcion: c.descripcion.replace(/ · mín\. .*$/, '') });
	}

	async descartarConsejo() {
		await guardarCambio(this.db, this.sb.id, 'cupones', 'consejo', { oculto: true });
	}

	async fijarAutomatica(id: string, activa: boolean): Promise<Resultado> {
		if (!(await this.cupones()).automaticas.some((a) => a.id === id)) return 'no-existe';
		await guardarCambio(this.db, this.sb.id, 'automatica', id, { activa });
		return 'ok';
	}

	async accionRelampago(
		accion: 'sumar-hora' | 'liberar' | 'terminar' | 'reanudar'
	): Promise<Resultado> {
		const e = ((await cambiosDe(this.db, this.sb.id, 'relampago')).get('vivo') as
			EstadoRelampago | undefined) ?? { horas: 0, liberadas: 0, terminada: false };
		const r = (await this.cupones()).relampago.vivo;
		if (!r) return 'no-existe';
		if (r.terminada && accion !== 'reanudar') return 'no-se-puede';
		const nuevo: EstadoRelampago =
			accion === 'sumar-hora'
				? { ...e, horas: e.horas + 1 }
				: accion === 'liberar'
					? { ...e, liberadas: e.liberadas + 1 }
					: { ...e, terminada: accion === 'terminar' };
		await guardarCambio(this.db, this.sb.id, 'relampago', 'vivo', nuevo);
		return 'ok';
	}

	// ---------------------------------------------------------------- Crecer

	async marketing(): Promise<Marketing> {
		const d = await this.demo();
		const [clientes, cup, nuevas] = await Promise.all([
			this.regClientes(),
			this.cupones(),
			cambiosDe(this.db, this.sb.id, 'campana')
		]);
		const cat = d.catalogo.categorias[0];
		const deCat = new Set(d.catalogo.productos.filter((p) => p.categoria === cat).map((p) => p.id));
		const compraron = new Set(
			d.pedidos.filter((p) => p.items.some((i) => deCat.has(i.productoId))).map((p) => p.clienteId)
		);
		const conWa = clientes.filter((c) => c.wa);
		const activos = cup.cupones.filter((c) => c.estado === 'activo');
		return marketingDemo(
			d,
			Date.now(),
			{
				todos: conWa.length,
				vip: conWa.filter((c) => c.seg === 'vip').length,
				riesgo: conWa.filter((c) => c.seg === 'risk').length,
				categoria: conWa.filter((c) => compraron.has(c.id)).length
			},
			activos.slice(0, 3).map((c) => ({
				codigo: c.codigo,
				descripcion: c.descripcion,
				usos: `${c.usos} usos${c.termina ? ` · ${cuando(c, Date.now()).toLowerCase()}` : ''}`
			})),
			activos.find((c) => c.tipo === 'pct')?.codigo ?? null,
			[...nuevas]
				.filter(([, v]) => v)
				.map(([, v]) => v as CampanaDemo)
				.sort((a, b) => b.en - a.en),
			COSTO_MENSAJE
		);
	}

	async crearCampana(c: {
		audiencia: 'todos' | 'vip' | 'riesgo' | 'categoria';
		programar: boolean;
	}): Promise<Campana | 'sin-audiencia'> {
		const m = await this.marketing();
		const personas = m.audiencias.find((a) => a.id === c.audiencia)?.cantidad ?? 0;
		if (!personas) return 'sin-audiencia';
		const ahora = Date.now();
		const id = `n${ahora.toString(36)}`;
		await guardarCambio(this.db, this.sb.id, 'campana', id, {
			id,
			audiencia: c.audiencia,
			personas,
			en: ahora,
			programada: c.programar ? sabado10(ahora) : null
		} satisfies CampanaDemo);
		return (await this.marketing()).campanas.find((x) => x.id === id)!;
	}

	async canales(): Promise<Canales> {
		const d = await this.demo();
		const [pausados, productos] = await Promise.all([
			cambiosDe(this.db, this.sb.id, 'canal'),
			this.regProductos()
		]);
		return canalesDemo(
			d,
			Date.now(),
			new Map(
				[...pausados].filter(([, v]) => v).map(([k, v]) => [k, (v as { pausado: boolean }).pausado])
			),
			productos.filter((p) => p.canales.includes('ig')).length
		);
	}

	async fijarCanal(id: string, activo: boolean): Promise<Resultado> {
		const c = (await this.canales()).canales.find((x) => x.id === id);
		if (!c) return 'no-existe';
		if (!c.conectado) return 'no-se-puede';
		await guardarCambio(this.db, this.sb.id, 'canal', id, { pausado: !activo });
		return 'ok';
	}

	async informes(): Promise<Informes> {
		const d = await this.demo();
		const productos = await this.regProductos();
		const vendidos = productos.reduce((s, p) => s + p.vendidos30, 0);
		const margen = vendidos
			? Math.round(productos.reduce((s, p) => s + p.margen * p.vendidos30, 0) / vendidos)
			: 38;
		return informesDemo(d, Date.now(), margen);
	}

	// ---------------------------------------------------------------- Actividad y Para hacer hoy

	/** Todos los cambios del visitante, del más nuevo al más viejo. */
	private async historial(): Promise<FilaCambio[]> {
		const filas = await this.db
			.select()
			.from(cambios)
			.where(and(eq(cambios.sandboxId, this.sb.id), eq(cambios.borrado, false)))
			.orderBy(desc(cambios.actualizadoEn))
			.limit(200);
		return filas.map((f) => ({
			entidad: f.entidad,
			clave: f.clave,
			datos: JSON.parse(f.datos ?? 'null'),
			en: f.actualizadoEn
		}));
	}

	private async nombres(): Promise<Nombres> {
		const [productos, cup, canales] = await Promise.all([
			this.regProductos(),
			this.cupones(),
			this.canales()
		]);
		const d = await this.base();
		return {
			productos: new Map(productos.map((p) => [p.id, p.nombre])),
			cupones: new Map(cup.cupones.map((c) => [c.id, c.codigo])),
			automaticas: new Map(cup.automaticas.map((a) => [a.id, a.titulo])),
			canales: new Map(canales.canales.map((c) => [c.id, c.nombre])),
			totales: new Map(d.pedidos.map((p) => [p.numero, p.total]))
		};
	}

	async actividad(): Promise<Evento[]> {
		const [filas, nombres] = await Promise.all([this.historial(), this.nombres()]);
		const deshechos = filas
			.filter((f) => f.entidad === 'deshecho')
			.map((f) => f.datos as { id: string; texto: string; donde: string; en: number });
		return actividadDemo(
			await this.demo(),
			Date.now(),
			this.sb.creadoEn,
			filas,
			deshechos,
			nombres
		);
	}

	async deshacerEvento(id: string): Promise<Resultado> {
		const e = (await this.actividad()).find((x) => x.id === id);
		if (!e) return 'no-existe';
		if (!e.deshacer) return 'no-se-puede';
		const i = id.indexOf(':');
		const [entidad, clave] = [id.slice(0, i), id.slice(i + 1)];
		await borrarCambio(this.db, this.sb.id, entidad, clave);
		await guardarCambio(this.db, this.sb.id, 'deshecho', `${id}:${Date.parse(e.cuando)}`, {
			id: `${id}:${Date.parse(e.cuando)}`,
			texto: `${e.texto} — deshecho`,
			donde: e.donde,
			en: Date.parse(e.cuando)
		});
		return 'ok';
	}

	/** El día de Argentina (las tareas se marcan por día). */
	private dia() {
		return new Date(Date.now() - 3 * 3_600_000).toISOString().slice(0, 10);
	}

	private async tareasHechas() {
		const hoy = this.dia();
		const c = await cambiosDe(this.db, this.sb.id, 'tarea');
		return new Set(
			[...c].filter(([k, v]) => v && k.startsWith(`${hoy}:`)).map(([k]) => k.slice(11))
		);
	}

	async paraHacer(): Promise<ParaHacer> {
		const d = await this.base();
		const ahora = Date.now();
		const [vivosC, bandeja, carritos, tablero, productos, hechas] = await Promise.all([
			this.cambiosPedidos(),
			this.bandeja(),
			this.carritos(),
			cambiosDe(this.db, this.sb.id, 'tablero'),
			this.regProductos(),
			this.tareasHechas()
		]);
		const vivos = pedidosVivos(d, vivosC, ahora);
		const preparar = vivos.filter((p) => p.estado === 'Para preparar');
		const chats = bandeja.conversaciones.filter((c) => c.sinLeer > 0 || c.atiende === 'vos');
		const pendientes = carritos.carritos.filter((c) => !c.recordado);
		const transf = vivos.filter((p) => p.estado === 'Sin pagar' && p.pago === 'transf');
		const reponer = registrosInventario(productos).filter((p) =>
			['out', 'crit', 'low'].includes(p.st)
		);
		const conTicket = vivos.filter((p) => p.estado !== 'Cancelado');
		const ticket = conTicket.length
			? conTicket.reduce((s, p) => s + p.total, 0) / conTicket.length
			: 0;
		const pesosT = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');
		type Tarea = ParaHacer['grupos'][number]['tareas'][number];
		const t = (x: Omit<Tarea, 'hecha'>, hay: boolean): Tarea[] =>
			hay ? [{ ...x, hecha: hechas.has(x.id) }] : [];
		const plural = (n: number, s: string, p: string) => `${n} ${n === 1 ? s : p}`;
		return {
			grupos: [
				{
					titulo: 'Urgente' as const,
					tareas: [
						...t(
							{
								id: 'preparar',
								tipo: 'preparar',
								titulo: `Despachar ${plural(preparar.length, 'pedido', 'pedidos')}`,
								detalle: 'Andreani retira a las 15:00',
								valor: preparar.reduce((s, p) => s + p.total, 0),
								accion: 'Preparar'
							},
							preparar.length > 0
						),
						...t(
							{
								id: 'chats',
								tipo: 'chats',
								titulo: `Responder ${plural(chats.length, 'chat', 'chats')}`,
								detalle: chats[0] ? `${chats[0].nombre.split(' ')[0]} espera respuesta` : '',
								valor: Math.round(chats.length * ticket),
								aproximado: true,
								accion: 'Abrir bandeja'
							},
							chats.length > 0
						)
					]
				},
				{
					titulo: 'Hoy' as const,
					tareas: [
						...t(
							{
								id: 'carritos',
								tipo: 'carritos',
								titulo: `Recuperar ${plural(pendientes.length, 'carrito', 'carritos')}`,
								detalle: `Recordatorio por WhatsApp · ${pesosT(pendientes.length * carritos.costoMensaje)}`,
								valor: pendientes.reduce((s, c) => s + c.total, 0),
								aproximado: true,
								accion: 'Enviar'
							},
							pendientes.length > 0
						),
						...t(
							{
								id: 'transferencias',
								tipo: 'transferencias',
								titulo: `Conciliar ${plural(transf.length, 'transferencia', 'transferencias')}`,
								detalle:
									transf.length === 1
										? `Pedido #${transf[0].numero}`
										: `Pedidos ${transf
												.slice(0, 3)
												.map((p) => `#${p.numero}`)
												.join(', ')
												.replace(/, ([^,]*)$/, ' y $1')}`,
								valor: transf.reduce((s, p) => s + p.total, 0),
								accion: 'Conciliar'
							},
							transf.length > 0
						)
					]
				},
				{
					titulo: 'Esta semana' as const,
					tareas: [
						...t(
							{
								id: 'certificado',
								tipo: 'certificado',
								titulo: 'Renovar certificado de ARCA',
								detalle: 'Vence en 5 días',
								valor: null,
								accion: 'Renovar'
							},
							!tablero.get('certificado')
						),
						...t(
							{
								id: 'reponer',
								tipo: 'reponer',
								titulo: `Reponer ${plural(reponer.length, 'producto', 'productos')}`,
								detalle: 'Hornero armó el pedido al proveedor',
								valor: null,
								accion: 'Revisar'
							},
							reponer.length > 0
						)
					]
				}
			].filter((g) => g.tareas.length)
		};
	}

	async marcarTarea(id: string, hecha: boolean): Promise<Resultado> {
		if (!['preparar', 'chats', 'carritos', 'transferencias', 'certificado', 'reponer'].includes(id))
			return 'no-existe';
		const clave = `${this.dia()}:${id}`;
		if (hecha) await guardarCambio(this.db, this.sb.id, 'tarea', clave, { en: Date.now() });
		else await borrarCambio(this.db, this.sb.id, 'tarea', clave);
		return 'ok';
	}

	// ---------------------------------------------------------------- Finanzas

	async finanzas(): Promise<Finanzas> {
		const [d, t, productos, retiros, campanas] = await Promise.all([
			this.demo(),
			this.tablero('7d'),
			this.regProductos(),
			cambiosDe(this.db, this.sb.id, 'retiro'),
			cambiosDe(this.db, this.sb.id, 'campana')
		]);
		const vendidos = productos.reduce((s, p) => s + p.vendidos30, 0);
		const margen = vendidos
			? Math.round(productos.reduce((s, p) => s + p.margen * p.vendidos30, 0) / vendidos)
			: 38;
		return finanzasDemo(
			d,
			Date.now(),
			t,
			margen,
			[...retiros.values()].filter(Boolean).map((r) => r as Retiro),
			[...campanas.values()]
				.filter(Boolean)
				.map((c) => c as CampanaDemo)
				.filter((c) => !c.programada || c.programada <= Date.now())
				.map((c) => ({ personas: c.personas, en: c.programada ?? c.en })),
			COSTO_MENSAJE
		);
	}

	async retirar(): Promise<{ monto: number; cuenta: string } | 'nada'> {
		const f = await this.finanzas();
		if (!f.disponible) return 'nada';
		const en = Date.now();
		await guardarCambio(this.db, this.sb.id, 'retiro', String(en), {
			monto: f.disponible,
			en
		} satisfies Retiro);
		return { monto: f.disponible, cuenta: f.banco ?? 'tu banco' };
	}

	// ---------------------------------------------------------------- Configuración

	private async estadoCuenta(): Promise<EstadoCuenta> {
		const c = await cambiosDe(this.db, this.sb.id, 'cuenta');
		const g = <T>(k: string, def: T) => (c.get(k) as T | undefined) ?? def;
		return {
			negocio: g('negocio', {}),
			seguridad: g('seguridad', {}),
			avisos: g('avisos', {}),
			cerradas: new Set(g<string[]>('cerradas', [])),
			quitados: new Set(g<string[]>('quitados', [])),
			roles: new Map(Object.entries(g<Record<string, Miembro['rol']>>('roles', {}))),
			invitados: g<Invitado[]>('invitados', []),
			plan: g('plan', {}),
			permisos: g('permisos', {}),
			packs: g('packs', {}),
			cancelado: g('cancelado', null),
			oferta: g('oferta', null)
		};
	}

	async configuracion(): Promise<Configuracion> {
		const [d, t, e] = await Promise.all([this.demo(), this.tienda(), this.estadoCuenta()]);
		return configuracionDemo(d, t.nombre, Date.now(), DIAS_DE_PRUEBA, e);
	}

	async guardarNegocio(n: Negocio) {
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'negocio', n);
	}

	async fijarOpcion(
		grupo: 'seguridad' | 'avisos',
		clave: string,
		activo: boolean
	): Promise<Resultado> {
		const c = await this.configuracion();
		if (!(clave in c[grupo])) return 'no-existe';
		await guardarCambio(this.db, this.sb.id, 'cuenta', grupo, {
			...(await this.estadoCuenta())[grupo],
			[clave]: activo
		});
		return 'ok';
	}

	async cerrarSesion(id: string): Promise<Resultado> {
		const s = (await this.configuracion()).sesiones.find((x) => x.id === id);
		if (!s) return 'no-existe';
		if (s.actual) return 'no-se-puede';
		const e = await this.estadoCuenta();
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'cerradas', [...e.cerradas, id]);
		return 'ok';
	}

	async invitar(contacto: string, rol: Miembro['rol']): Promise<Miembro | 'repetido'> {
		const c = await this.configuracion();
		const v = contacto.trim().toLowerCase();
		if (c.equipo.some((m) => m.contacto.toLowerCase() === v)) return 'repetido';
		const e = await this.estadoCuenta();
		const nuevo: Invitado = {
			id: `i${Date.now().toString(36)}`,
			contacto: contacto.trim(),
			rol,
			en: Date.now()
		};
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'invitados', [...e.invitados, nuevo]);
		return (await this.configuracion()).equipo.find((m) => m.id === nuevo.id)!;
	}

	async cambiarPlan(plan: Plan, ciclo?: 'mes' | 'anio') {
		const e = await this.estadoCuenta();
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'plan', {
			...e.plan,
			actual: plan,
			...(ciclo ? { ciclo } : {})
		});
	}

	// ---------------------------------------------------------------- Cuenta

	private mes() {
		return new Date(Date.now() - 3 * 3_600_000).toISOString().slice(0, 7);
	}

	async planCuenta(): Promise<PlanCuenta> {
		const [c, e, productos, uso] = await Promise.all([
			this.configuracion(),
			this.estadoCuenta(),
			this.regProductos(),
			cambiosDe(this.db, this.sb.id, 'hornero')
		]);
		const usado = (uso.get('uso') as { cuantos?: number } | undefined)?.cuantos ?? 0;
		const fin = Date.now() + DIAS_DE_PRUEBA * 86_400_000;
		return {
			actual: c.plan.actual,
			ciclo: c.plan.ciclo,
			enPrueba: c.plan.enPrueba,
			diasPrueba: c.plan.diasPrueba,
			proximoCobro: new Date(fin).toISOString(),
			uso: {
				productos: productos.length,
				personas: c.equipo.length,
				// La tienda de ejemplo viene con su historia; se suma lo que le preguntó el visitante.
				hornero: 1842 + usado,
				packs: e.packs[this.mes()] ?? 0,
				sinIa: 3912
			},
			medio: { marca: 'visa', ultimos: '4821', vence: '08/29' },
			facturacion: { razonSocial: c.negocio.razonSocial, cuit: c.negocio.cuit, tipo: 'Factura A' },
			// Está en prueba: la primera factura de Mostrador llega cuando termina.
			facturas: [],
			cancelado: e.cancelado ? { hasta: new Date(fin).toISOString() } : null
		};
	}

	async comprarPack() {
		const e = await this.estadoCuenta();
		const mes = this.mes();
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'packs', {
			...e.packs,
			[mes]: (e.packs[mes] ?? 0) + 1
		});
	}

	async devolverPack(): Promise<Resultado> {
		const e = await this.estadoCuenta();
		const mes = this.mes();
		if (!e.packs[mes]) return 'no-existe';
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'packs', {
			...e.packs,
			[mes]: e.packs[mes] - 1
		});
		return 'ok';
	}

	async cancelarPlan(motivo: number, accion: 'cancelar' | 'oferta') {
		const en = Date.now();
		if (accion === 'oferta')
			await guardarCambio(this.db, this.sb.id, 'cuenta', 'oferta', { motivo, en });
		else await guardarCambio(this.db, this.sb.id, 'cuenta', 'cancelado', { motivo, en });
		return { hasta: new Date(en + DIAS_DE_PRUEBA * 86_400_000).toISOString() };
	}

	async reactivarPlan() {
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'cancelado', null);
	}

	async permisos(): Promise<Permisos> {
		const e = await this.estadoCuenta();
		return {
			permisos: PERMISOS.map((p) => ({
				id: p.id,
				titulo: p.titulo,
				detalle: p.detalle,
				roles: Object.fromEntries(
					ROLES.map((r, i) => [
						r,
						fijo(p.id, r) ? !!p.def[i] : (e.permisos[`${p.id}:${r}`] ?? !!p.def[i])
					])
				) as Permisos['permisos'][number]['roles']
			}))
		};
	}

	async fijarPermiso(permiso: string, rol: Rol, activo: boolean): Promise<Resultado> {
		if (!PERMISOS.some((p) => p.id === permiso)) return 'no-existe';
		if (fijo(permiso, rol)) return 'no-se-puede';
		const e = await this.estadoCuenta();
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'permisos', {
			...e.permisos,
			[`${permiso}:${rol}`]: activo
		});
		return 'ok';
	}

	async cambiarRol(id: string, rol: Miembro['rol']): Promise<Resultado> {
		const m = (await this.configuracion()).equipo.find((x) => x.id === id);
		if (!m) return 'no-existe';
		if (m.vos || rol === 'duenio') return 'no-se-puede';
		const e = await this.estadoCuenta();
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'roles', {
			...Object.fromEntries(e.roles),
			[id]: rol
		});
		return 'ok';
	}

	async quitarDelEquipo(id: string): Promise<Resultado> {
		const m = (await this.configuracion()).equipo.find((x) => x.id === id);
		if (!m) return 'no-existe';
		if (m.vos) return 'no-se-puede';
		const e = await this.estadoCuenta();
		if (m.pendiente)
			await guardarCambio(
				this.db,
				this.sb.id,
				'cuenta',
				'invitados',
				e.invitados.filter((i) => i.id !== id)
			);
		else await guardarCambio(this.db, this.sb.id, 'cuenta', 'quitados', [...e.quitados, id]);
		return 'ok';
	}

	// ---------------------------------------------------------------- Dominio, avisos y mis datos

	async dominios(): Promise<Dominios> {
		const [t, c] = await Promise.all([this.tienda(), cambiosDe(this.db, this.sb.id, 'cuenta')]);
		const lista = (c.get('dominios') as DominioDemo[] | undefined) ?? [];
		const ahora = Date.now();
		const gratis = `${SITE.replace(/^https?:\/\//, '')}/t/${t.slug}`;
		const principal = (c.get('principal') as string | undefined) ?? gratis;
		// En la demo los registros "aparecen" a los 1, 3 y 5 minutos de agregar el dominio.
		const paso = (d: DominioDemo, min: number) => ahora - d.en > min * 60_000;
		return {
			gratis,
			dominios: [
				{
					nombre: gratis,
					detalle: 'La dirección gratis de tu tienda',
					estado: 'activo',
					principal: principal === gratis
				},
				...lista.map((d) => {
					const listo = paso(d, 5);
					return {
						nombre: d.nombre,
						detalle: listo ? 'Conectado' : 'Agregado recién · revisando DNS',
						estado: listo ? ('activo' as const) : ('verificando' as const),
						principal: principal === d.nombre,
						chequeos: [
							{
								titulo: 'Registro CNAME de www',
								estado: paso(d, 1) ? ('ok' as const) : ('esperando' as const)
							},
							{
								titulo: 'Verificación TXT',
								estado: paso(d, 3)
									? ('ok' as const)
									: paso(d, 1)
										? ('esperando' as const)
										: ('pendiente' as const)
							},
							{
								titulo: 'Certificado SSL',
								estado: listo ? ('ok' as const) : ('pendiente' as const)
							}
						]
					};
				})
			],
			registros: [
				{ tipo: 'CNAME', nombre: 'www', valor: `tiendas.${SITE.replace(/^https?:\/\//, '')}` },
				{ tipo: 'TXT', nombre: '_mostrador', valor: `mostrador-verificacion=${t.slug}` }
			]
		};
	}

	async conectarDominio(nombre: string): Promise<Resultado> {
		const c = await cambiosDe(this.db, this.sb.id, 'cuenta');
		const lista = (c.get('dominios') as DominioDemo[] | undefined) ?? [];
		if (lista.some((d) => d.nombre === nombre)) return 'no-se-puede';
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'dominios', [
			...lista,
			{ nombre, en: Date.now() }
		]);
		return 'ok';
	}

	async dominioPrincipal(nombre: string): Promise<Resultado> {
		const d = (await this.dominios()).dominios.find((x) => x.nombre === nombre);
		if (!d) return 'no-existe';
		if (d.estado !== 'activo') return 'no-se-puede';
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'principal', nombre);
		return 'ok';
	}

	async quitarDominio(nombre: string): Promise<Resultado> {
		const c = await cambiosDe(this.db, this.sb.id, 'cuenta');
		const lista = (c.get('dominios') as DominioDemo[] | undefined) ?? [];
		if (!lista.some((d) => d.nombre === nombre)) return 'no-existe';
		await guardarCambio(
			this.db,
			this.sb.id,
			'cuenta',
			'dominios',
			lista.filter((d) => d.nombre !== nombre)
		);
		if (c.get('principal') === nombre)
			await guardarCambio(this.db, this.sb.id, 'cuenta', 'principal', null);
		return 'ok';
	}

	// ---------------------------------------------------------------- Mercado Pago

	async integraciones(): Promise<EstadoIntegraciones> {
		const [c, mp] = await Promise.all([
			cambiosDe(this.db, this.sb.id, 'integraciones'),
			this.mercadoPago()
		]);
		const e = estadoInicial(true);
		const g = (k: keyof EstadoIntegraciones) =>
			(c.get(k) as Record<string, never> | undefined) ?? {};
		Object.assign(e.conectadas, g('conectadas'));
		for (const [id, o] of Object.entries(g('opciones') as Record<string, Record<string, boolean>>))
			Object.assign((e.opciones[id] ??= {}), o);
		Object.assign(e.ia, g('ia'));
		Object.assign(e.permisosIa, g('permisosIa'));
		e.conectadas.mp = mp.conectada;
		e.conectadas.ml = (await this.mercadoLibre()).conectada;
		// Telegram cuenta si hay un chat vinculado (lib/server/telegram.ts).
		e.conectadas.tg = !!(await cambiosDe(this.db, this.sb.id, 'cuenta')).get('telegram');
		return e;
	}
	async ajustarIntegracion(a: AjusteIntegracion): Promise<Resultado> {
		const e = await this.integraciones();
		if (a.tipo === 'conexion') {
			if (
				a.id === 'mp' ||
				a.id === 'ml' ||
				a.id === 'tg' ||
				!INTEGRACIONES.some((i) => i.id === a.id && !i.pronto)
			)
				return 'no-se-puede';
			e.conectadas[a.id] = a.valor;
			await guardarCambio(this.db, this.sb.id, 'integraciones', 'conectadas', e.conectadas);
		} else if (a.tipo === 'opcion') {
			if (!(a.clave in (e.opciones[a.id] ?? {}))) return 'no-existe';
			e.opciones[a.id][a.clave] = a.valor;
			await guardarCambio(this.db, this.sb.id, 'integraciones', 'opciones', e.opciones);
		} else if (a.tipo === 'ia') {
			if (!(a.id in e.ia)) return 'no-existe';
			e.ia[a.id] = a.valor;
			await guardarCambio(this.db, this.sb.id, 'integraciones', 'ia', e.ia);
		} else {
			if (!(a.id in e.permisosIa) || a.id === 'leer') return 'no-se-puede';
			e.permisosIa[a.id] = a.valor;
			await guardarCambio(this.db, this.sb.id, 'integraciones', 'permisosIa', e.permisosIa);
		}
		return 'ok';
	}

	/** En la demo Mercado Libre arranca conectado (como en el handoff) y todo se simula. */
	async mercadoLibre(): Promise<ConexionML> {
		const c = await cambiosDe(this.db, this.sb.id, 'cuenta');
		const ml = c.get('ml') as
			{ conectada: boolean; desde: number; recargo?: number; pausar?: boolean } | undefined;
		const conectada = ml?.conectada ?? true;
		const productos = await this.regProductos();
		return {
			disponible: true,
			conectada,
			...(conectada
				? {
						apodo: 'MATEYCO_OFICIAL',
						desde: new Date(ml?.desde ?? Date.now() - 60 * 86_400_000).toISOString()
					}
				: {}),
			reautorizar: false,
			recargo: ml?.recargo ?? 10,
			pausarSinStock: ml?.pausar ?? true,
			publicaciones: conectada ? productos.filter((p) => p.canales.includes('ml')).length : 0,
			preguntas: conectada ? 3 : 0
		};
	}
	async conectarMercadoLibre() {
		return '/integraciones/mercadolibre/callback?code=demo&state=demo';
	}
	async volverMercadoLibre(codigo: string, estado: string): Promise<Resultado> {
		if (codigo !== 'demo' || estado !== 'demo') return 'no-se-puede';
		const c = await cambiosDe(this.db, this.sb.id, 'cuenta');
		const ml = (c.get('ml') as Record<string, unknown> | undefined) ?? {};
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'ml', {
			...ml,
			conectada: true,
			desde: Date.now()
		});
		return 'ok';
	}
	async desconectarMercadoLibre() {
		const c = await cambiosDe(this.db, this.sb.id, 'cuenta');
		const ml = (c.get('ml') as Record<string, unknown> | undefined) ?? {};
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'ml', { ...ml, conectada: false });
	}
	async ajustarMercadoLibre(recargo: number, pausar: boolean): Promise<Resultado> {
		if (!(await this.mercadoLibre()).conectada) return 'no-se-puede';
		const c = await cambiosDe(this.db, this.sb.id, 'cuenta');
		const ml = (c.get('ml') as Record<string, unknown> | undefined) ?? {};
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'ml', { ...ml, recargo, pausar });
		return 'ok';
	}
	async publicarEnMercadoLibre(productoId: string): Promise<PublicacionML | { error: string }> {
		if (!(await this.mercadoLibre()).conectada) return { error: 'Conectá Mercado Libre primero.' };
		const n = parseInt(productoId.replace(/\D/g, '').slice(-6) || '1', 10);
		const item = `MLA${1_400_000_000 + n}`;
		return { itemId: item, permalink: `https://articulo.mercadolibre.com.ar/MLA-${item.slice(3)}` };
	}
	async despublicarDeMercadoLibre() {}

	/** En la demo la tienda arranca con Mercado Pago conectado (el onboarding lo conecta). */
	async mercadoPago(): Promise<ConexionMP> {
		const c = await cambiosDe(this.db, this.sb.id, 'cuenta');
		const mp = c.get('mp') as { conectada: boolean; desde: number } | undefined;
		const conectada = mp?.conectada ?? true;
		return {
			disponible: true,
			conectada,
			...(conectada
				? {
						usuario: '3666528660',
						desde: new Date(mp?.desde ?? Date.now() - 30 * 86_400_000).toISOString()
					}
				: {}),
			reautorizar: false,
			cajaQR: conectada
		};
	}
	/** La demo no sale a Mercado Pago: vuelve directo con un código de mentira. */
	async conectarMercadoPago() {
		return '/integraciones/mercadopago/callback?code=demo&state=demo';
	}
	async volverMercadoPago(codigo: string, estado: string): Promise<Resultado> {
		if (codigo !== 'demo' || estado !== 'demo') return 'no-se-puede';
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'mp', {
			conectada: true,
			desde: Date.now()
		});
		return 'ok';
	}
	async desconectarMercadoPago() {
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'mp', { conectada: false, desde: 0 });
	}

	async preferenciasAvisos(): Promise<PreferenciasAvisos> {
		const c = await cambiosDe(this.db, this.sb.id, 'cuenta');
		return (
			(c.get('avisos-canales') as PreferenciasAvisos | undefined) ?? {
				matriz: {},
				silencio: true,
				resumen: 'whatsapp'
			}
		);
	}

	async guardarPreferenciasAvisos(p: PreferenciasAvisos) {
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'avisos-canales', p);
	}

	private async inactivos() {
		// Sin compras en 3 años: en la demo, con dos meses de historia, no hay.
		return (await this.regClientes()).filter((c) => c.ultima > 3 * 365);
	}

	async misDatos(): Promise<MisDatos> {
		const [productos, clientes, d, facturas, conv, exps, c] = await Promise.all([
			this.regProductos(),
			this.regClientes(),
			this.base(),
			this.regFacturacion(),
			this.bandeja(),
			cambiosDe(this.db, this.sb.id, 'exportacion'),
			cambiosDe(this.db, this.sb.id, 'cuenta')
		]);
		return {
			conteos: {
				productos: productos.length,
				clientes: clientes.length,
				pedidos: d.pedidos.filter((p) => p.creado <= Date.now()).length,
				facturas: facturas.length,
				resenas: d.resenas.length,
				conversaciones: conv.conversaciones.length
			},
			recientes: [...exps.values()]
				.filter(Boolean)
				.map((x) => x as { tipo: string; formato: string; en: number })
				.sort((a, b) => b.en - a.en)
				.slice(0, 5)
				.map((x) => ({ ...x, en: new Date(x.en).toISOString() })),
			inactivos: clientes.filter((x) => x.ultima > 3 * 365).length,
			borrado: (c.get('borrado') as MisDatos['borrado'] | undefined) ?? null
		};
	}

	/** Filas de cada exportación. */
	async exportar(tipo: TipoExport): Promise<(string | number | null)[][]> {
		const fecha = (t: number) => new Date(t).toISOString().slice(0, 10);
		await guardarCambio(this.db, this.sb.id, 'exportacion', String(Date.now()), {
			tipo,
			formato: '',
			en: Date.now()
		});
		if (tipo === 'productos')
			return [
				['SKU', 'Nombre', 'Categoría', 'Precio', 'Costo', 'Stock', 'Publicado'],
				...(await this.regProductos()).map((p) => [
					p.sku,
					p.nombre,
					p.categoria,
					p.precio,
					p.costo,
					p.stock,
					p.publicado ? 'Sí' : 'No'
				])
			];
		if (tipo === 'clientes')
			return [
				['Nombre', 'Ciudad', 'Teléfono', 'Pedidos', 'Gastado'],
				...(await this.regClientes()).map((c) => [
					c.nombre,
					c.ciudad,
					c.telefono,
					c.pedidos,
					c.gastado
				])
			];
		if (tipo === 'facturas')
			return [
				['Número', 'Tipo', 'Cliente', 'Total', 'Estado', 'Pedido'],
				...(await this.regFacturacion()).map((f) => [
					f.numero,
					f.tipo,
					f.cliente,
					f.total,
					f.estado,
					f.pedido
				])
			];
		const d = await this.base();
		const nombre = new Map(d.clientes.map((c) => [c.id, c.nombre]));
		if (tipo === 'pedidos')
			return [
				['Número', 'Fecha', 'Cliente', 'Canal', 'Pago', 'Envío', 'Estado', 'Total'],
				...d.pedidos
					.filter((p) => p.creado <= Date.now())
					.map((p) => [
						p.numero,
						fecha(p.creado),
						nombre.get(p.clienteId) ?? '',
						p.canal,
						p.pago,
						p.envio,
						p.estado,
						p.total
					])
			];
		if (tipo === 'resenas')
			return [
				['Fecha', 'Cliente', 'Producto', 'Estrellas', 'Reseña'],
				...d.resenas.map((r) => [
					fecha(r.creada),
					nombre.get(r.clienteId) ?? '',
					r.producto,
					r.estrellas,
					r.texto
				])
			];
		const filas: (string | number | null)[][] = [
			['Conversación', 'Cliente', 'De', 'Mensaje', 'Fecha']
		];
		for (const c of (await this.bandeja()).conversaciones) {
			const conv = await this.conversacion(c.id);
			for (const m of conv?.mensajes ?? [])
				if (m.texto) filas.push([c.id, c.nombre, m.de, m.texto, m.en.slice(0, 16)]);
		}
		return filas;
	}

	async borrarInactivos(): Promise<number> {
		const n = (await this.inactivos()).length;
		await guardarCambio(
			this.db,
			this.sb.id,
			'cuenta',
			'borrado',
			n ? { clientes: n, el: new Date(Date.now() + 7 * 86_400_000).toISOString() } : null
		);
		return n;
	}

	async frenarBorrado() {
		await guardarCambio(this.db, this.sb.id, 'cuenta', 'borrado', null);
	}

	/** Cerrar la cuenta en la demo: se borra todo lo del visitante y vuelve a empezar. */
	async cerrarCuenta(confirmacion: string): Promise<Resultado> {
		const t = await this.tienda();
		if (confirmacion.trim().toLowerCase() !== t.nombre.toLowerCase()) return 'no-se-puede';
		await this.db.batch([
			this.db.delete(cambios).where(eq(cambios.sandboxId, this.sb.id)),
			this.db.delete(sandboxes).where(eq(sandboxes.id, this.sb.id))
		]);
		return 'ok';
	}

	// ---------------------------------------------------------------- Mi tienda

	async miTienda(): Promise<MiTienda> {
		const [d, t, c] = await Promise.all([
			this.demo(),
			this.tienda(),
			cambiosDe(this.db, this.sb.id, 'mi-tienda')
		]);
		const borrador = (c.get('borrador') as DisenoTienda | undefined) ?? DISENO_INICIAL;
		const publicado = (c.get('publicado') as DisenoTienda | undefined) ?? DISENO_INICIAL;
		const p = buscarPlantilla(t.plantilla);
		const vendidos = new Map<string, number>();
		for (const x of d.pedidos)
			for (const i of x.items)
				vendidos.set(i.productoId, (vendidos.get(i.productoId) ?? 0) + i.cantidad);
		const top = [...d.catalogo.productos].sort(
			(a, b) => (vendidos.get(b.id) ?? 0) - (vendidos.get(a.id) ?? 0)
		);
		const buenas = d.resenas.filter((r) => r.estrellas === 5 && r.texto.length > 30);
		const promedio = d.resenas.length
			? d.resenas.reduce((s, r) => s + r.estrellas, 0) / d.resenas.length
			: 0;
		return {
			plantilla: p?.nombre ?? t.plantilla,
			direccion: `${SITE.replace(/^https?:\/\//, '')}/t/${t.slug}`,
			diseno: borrador,
			sinPublicar: JSON.stringify(borrador) !== JSON.stringify(publicado),
			vista: {
				tienda: t.nombre,
				categorias: d.catalogo.categorias,
				productos: top
					.slice(0, 4)
					.map((x) => ({ nombre: x.nombre, foto: x.foto, precio: x.precio })),
				resena: d.resenas.length
					? { promedio: Math.round(promedio * 10) / 10, texto: (buenas[0] ?? d.resenas[0]).texto }
					: null,
				fotos: [...new Set(top.flatMap((x) => (x.foto ? [x.foto] : [])))].slice(0, 6),
				portada: { foto: p?.img ?? top[0]?.foto ?? null, titulo: p?.lema ?? t.nombre }
			}
		};
	}

	/**
	 * La tienda online como la ve quien compra (lo publicado en Mi tienda). vitrina: es la tienda de
	 * ejemplo de una plantilla para mirar, sin carrito real.
	 */
	async tiendaPublica(vitrina = false): Promise<TiendaPublica> {
		const [d, t, c, conf, productos, mp, cr, respuestasQ] = await Promise.all([
			this.demo(),
			this.tienda(),
			cambiosDe(this.db, this.sb.id, 'mi-tienda'),
			this.configuracion(),
			this.regProductos(),
			this.mercadoPago(),
			this.cambiosResenas(),
			this.respuestasPreguntas()
		]);
		const diseno = (c.get('publicado') as DisenoTienda | undefined) ?? DISENO_INICIAL;
		const p = buscarPlantilla(t.plantilla);
		const visibles = productos.filter((x) => x.publicado && x.canales.includes('web'));
		const vendidos = new Map(visibles.map((x) => [x.id, x.vendidos30]));
		const masVendidos = [...visibles]
			.sort((a, b) => (vendidos.get(b.id) ?? 0) - (vendidos.get(a.id) ?? 0))
			.map((x) => x.id);
		const delCatalogo = new Map(d.catalogo.productos.map((x) => [x.id, x]));
		const categorias = [...new Set(visibles.map((x) => x.categoria).filter(Boolean))].map(
			(nombre) => {
				const de = visibles.filter((x) => x.categoria === nombre);
				return { nombre, cantidad: de.length, foto: de.find((x) => x.foto)?.foto ?? null };
			}
		);
		// Las del visitante se ven todas (como en Go); de las de ejemplo, las seis mejores escritas.
		cr.productos = new Map(productos.map((x) => [x.id, { nombre: x.nombre, foto: x.foto }]));
		const aLaVista = todasLasResenas(d, cr).filter((r) => !r.oculta);
		const resenas = [
			...aLaVista.filter((r) => !r.ejemplo),
			...aLaVista.filter((r) => r.ejemplo && r.estrellas >= 4 && r.texto.length > 25).slice(0, 6)
		].map((r) => ({
			texto: r.texto,
			autor: r.firma,
			producto: r.producto ?? '',
			estrellas: r.estrellas,
			conCompra: r.conCompra,
			respuesta: r.respuesta
		}));
		const promedio = aLaVista.length
			? Math.round((aLaVista.reduce((s, r) => s + r.estrellas, 0) / aLaVista.length) * 10) / 10
			: null;
		const publicados = new Set(visibles.map((x) => x.id));
		const preguntas = cr.preguntas
			.filter((q) => publicados.has(q.productoId) && respuestasQ.get(q.id))
			.map((q) => ({
				productoId: q.productoId,
				pregunta: q.pregunta,
				respuesta: respuestasQ.get(q.id)!
			}));
		return {
			slug: t.slug,
			nombre: t.nombre,
			plantilla: t.plantilla,
			paleta: diseno.paleta,
			secciones: diseno.secciones.filter((s) => s.visible).map((s) => s.id),
			portada: {
				titulo: p?.lema ?? t.nombre,
				lema: d.catalogo.ciudad ? `Desde ${d.catalogo.ciudad}` : '',
				foto: p?.img ?? visibles.find((x) => x.foto)?.foto ?? null
			},
			ciudad: d.catalogo.ciudad,
			direccion: conf.negocio.direccion,
			whatsapp: conf.negocio.whatsapp,
			categorias,
			productos: visibles.map((x) => ({
				id: x.id,
				nombre: x.nombre,
				descripcion: x.descripcion || delCatalogo.get(x.id)?.detalle || '',
				categoria: x.categoria || null,
				precio: x.precio,
				anterior: delCatalogo.get(x.id)?.anterior ?? null,
				foto: x.foto,
				etiqueta:
					delCatalogo.get(x.id)?.etiqueta ?? (x.stock > 0 && x.stock <= 3 ? 'Últimos' : null),
				stock: x.stock
			})),
			masVendidos,
			resenas,
			promedio,
			preguntas,
			fotos: [...new Set(visibles.flatMap((x) => (x.foto ? [x.foto] : [])))].slice(0, 6),
			cobros: { transferencia: 15, cuotas: 6, mercadoPago: mp.conectada },
			envios: { gratisDesde: 240_000, retiro: true, moto: true, correo: true },
			vitrina
		};
	}

	async guardarMiTienda(diseno: DisenoTienda) {
		await guardarCambio(this.db, this.sb.id, 'mi-tienda', 'borrador', diseno);
	}

	async publicarTienda() {
		const c = await cambiosDe(this.db, this.sb.id, 'mi-tienda');
		await guardarCambio(
			this.db,
			this.sb.id,
			'mi-tienda',
			'publicado',
			c.get('borrador') ?? DISENO_INICIAL
		);
	}

	// ---------------------------------------------------------------- Listas de precio

	/** Aplica lo programado si ya llegó el día (la demo no tiene worker: se mira al entrar). */
	private async programadoVigente() {
		const p = (await cambiosDe(this.db, this.sb.id, 'aumento')).get('programado') as
			AumentoProgramado | null | undefined;
		if (!p) return null;
		if (p.para > Date.now()) return p;
		await borrarCambio(this.db, this.sb.id, 'aumento', 'programado');
		await this.aplicarAumento(p.pct, p.redondeo);
		return null;
	}

	async listasPrecio(): Promise<ListasPrecio> {
		const programado = await this.programadoVigente();
		const d = await this.demo();
		const productos = await this.regProductos();
		const mayoristas = d.clientes.filter((c) => c.mayorista).length;
		return {
			listas: [
				{
					id: 'minorista',
					nombre: 'Minorista',
					quien: 'Tienda y local',
					ajuste: 0,
					detalle: 'Precio de referencia'
				},
				{
					id: 'mayorista',
					nombre: 'Mayorista',
					quien: 'Desde 10 unidades',
					ajuste: -25,
					detalle: `${mayoristas} ${mayoristas === 1 ? 'cliente mayorista' : 'clientes mayoristas'}`
				},
				{
					id: 'ml',
					nombre: 'Mercado Libre',
					quien: 'Cubre la comisión',
					ajuste: 10,
					detalle: 'Se actualiza solo'
				},
				{
					id: 'revendedoras',
					nombre: 'Revendedoras',
					quien: 'Para quienes revenden',
					ajuste: -15,
					detalle: 'Con link propio'
				}
			],
			productos: productos.length,
			muestra: [...productos]
				.sort((a, b) => b.vendidos30 - a.vendidos30)
				.slice(0, 4)
				.map((p) => ({ id: p.id, nombre: p.nombre, precio: p.precio })),
			programado: programado
				? {
						pct: programado.pct,
						redondeo: programado.redondeo,
						para: new Date(programado.para).toISOString()
					}
				: null
		};
	}

	/** Sube todos los precios y guarda los de antes para deshacer. */
	private async aplicarAumento(pct: number, redondeo: Redondeo) {
		const productos = await this.regProductos();
		const cambios = await this.cambiosProductos();
		const antes: Record<string, number | null> = {};
		for (const p of productos) {
			const c = cambios.get(p.id) ?? {};
			antes[p.id] = c.precio ?? null;
			await guardarCambio(this.db, this.sb.id, 'producto', p.id, {
				...c,
				precio: aumentar(p.precio, pct, redondeo)
			} satisfies CambioProducto);
		}
		await guardarCambio(this.db, this.sb.id, 'aumento', 'ultimo', { antes, en: Date.now() });
		return productos.length;
	}

	async aumentarPrecios(pct: number, redondeo: Redondeo, programar: boolean) {
		if (programar) {
			const para = proximoLunes(Date.now());
			await guardarCambio(this.db, this.sb.id, 'aumento', 'programado', {
				pct,
				redondeo,
				para
			} satisfies AumentoProgramado);
			return { productos: (await this.regProductos()).length, para: new Date(para).toISOString() };
		}
		return { productos: await this.aplicarAumento(pct, redondeo), para: null };
	}

	async deshacerAumento(): Promise<Resultado> {
		const u = (await cambiosDe(this.db, this.sb.id, 'aumento')).get('ultimo') as
			{ antes: Record<string, number | null> } | null | undefined;
		if (!u) return 'no-existe';
		const cambios = await this.cambiosProductos();
		for (const [id, precio] of Object.entries(u.antes)) {
			const resto: CambioProducto = { ...cambios.get(id) };
			if (precio === null) delete resto.precio;
			else resto.precio = precio;
			await guardarCambio(this.db, this.sb.id, 'producto', id, resto);
		}
		await borrarCambio(this.db, this.sb.id, 'aumento', 'ultimo');
		return 'ok';
	}

	async cancelarAumento(): Promise<Resultado> {
		if (!(await this.programadoVigente())) return 'no-existe';
		await borrarCambio(this.db, this.sb.id, 'aumento', 'programado');
		return 'ok';
	}

	// ---------------------------------------------------------------- Links de pago

	async crearLink(l: { monto: number; concepto: string; cuotas: boolean; pedido: number | null }) {
		const cod = codigo();
		await guardarCambio(this.db, this.sb.id, 'link', cod, {
			...l,
			creado: Date.now(),
			pagado: null
		} satisfies LinkDemo);
		return cod;
	}

	/** Pagar un link (en la demo se simula): queda pagado y, si cobra un pedido, el pedido pasa a preparar. */
	async pagarLink(cod: string): Promise<Resultado> {
		const c = (await cambiosDe(this.db, this.sb.id, 'link')).get(cod) as
			LinkDemo | null | undefined;
		if (!c) return 'no-existe';
		if (c.pagado) return 'no-se-puede';
		await guardarCambio(this.db, this.sb.id, 'link', cod, {
			...c,
			pagado: Date.now()
		} satisfies LinkDemo);
		if (c.pedido) {
			await this.cambiarEstado(c.pedido, 'Para preparar');
			// Si es un pedido cargado acá, queda cobrado y facturado.
			const n = (await cambiosDe(this.db, this.sb.id, 'pedido-nuevo')).get(String(c.pedido)) as
				PedidoNuevoDemo | undefined;
			if (n)
				await guardarCambio(this.db, this.sb.id, 'pedido-nuevo', String(c.pedido), {
					...n,
					pedido: { ...n.pedido, estadoPago: 'Aprobado', pago: 'mp', facturado: true }
				} satisfies PedidoNuevoDemo);
		}
		return 'ok';
	}

	// ---------------------------------------------------------------- Devoluciones y reseñas

	async devoluciones() {
		const d = await this.demo();
		const c = await cambiosDe(this.db, this.sb.id, 'devolucion');
		const estados = new Map(
			[...c].filter(([, v]) => v).map(([k, v]) => [k, (v as { estado: EstadoDevolucion }).estado])
		);
		return devolucionesDemo(d, estados);
	}

	async resolverDevolucion(
		id: string,
		accion: 'aprobar' | 'cambio' | 'deshacer'
	): Promise<Resultado> {
		const dev = (await this.devoluciones()).devoluciones.find((x) => x.id === id);
		if (!dev) return 'no-existe';
		if (accion === 'deshacer') {
			await borrarCambio(this.db, this.sb.id, 'devolucion', id);
			return 'ok';
		}
		await guardarCambio(this.db, this.sb.id, 'devolucion', id, {
			estado: accion === 'aprobar' ? 'Aprobada' : 'Cambio ofrecido'
		});
		return 'ok';
	}

	/** Respuestas publicadas a las preguntas (de Mercado Libre y de la tienda), por id. */
	private async respuestasPreguntas() {
		const c = await cambiosDe(this.db, this.sb.id, 'pregunta');
		return new Map(
			[...c].filter(([, v]) => v).map(([k, v]) => [k, (v as { respuesta: string }).respuesta])
		);
	}

	/**
	 * Reseñas y preguntas que dejó el visitante en su tienda online, más lo que hizo con ellas. Cada
	 * acción va en su propio cambio (ocultar y responder no se pisan si llegan juntas).
	 */
	private async cambiosResenas(): Promise<CambiosResenas> {
		const [nuevas, ocultas, respuestas, preguntas] = await Promise.all([
			cambiosDe(this.db, this.sb.id, 'resena-nueva'),
			cambiosDe(this.db, this.sb.id, 'resena-oculta'),
			cambiosDe(this.db, this.sb.id, 'resena-respuesta'),
			cambiosDe(this.db, this.sb.id, 'pregunta-tienda')
		]);
		return {
			nuevas: [...nuevas.values()].filter(Boolean) as ResenaNuevaDemo[],
			ocultas: new Set([...ocultas].filter(([, v]) => v).map(([k]) => k)),
			respuestas: new Map(
				[...respuestas]
					.filter(([, v]) => v)
					.map(([k, v]) => [k, (v as { respuesta: string }).respuesta])
			),
			preguntas: [...preguntas.values()].filter(Boolean) as PreguntaTiendaDemo[]
		};
	}

	async resenas() {
		const [d, respuestas, cambios, productos] = await Promise.all([
			this.demo(),
			this.respuestasPreguntas(),
			this.cambiosResenas(),
			this.regProductos()
		]);
		// Con los productos que sumó el visitante (el catálogo de ejemplo no los tiene).
		cambios.productos = new Map(productos.map((p) => [p.id, { nombre: p.nombre, foto: p.foto }]));
		return resenasDemo(d, Date.now(), respuestas, cambios);
	}

	async responderPregunta(id: string, respuesta: string): Promise<Resultado> {
		if (!(await this.resenas()).preguntas.some((p) => p.id === id)) return 'no-existe';
		await guardarCambio(this.db, this.sb.id, 'pregunta', id, {
			respuesta: respuesta.trim().slice(0, id.startsWith('t-') ? 1000 : 2000)
		});
		return 'ok';
	}

	/** Ocultar o mostrar una reseña (nunca se borra), o responderla en público (vacía, la saca). */
	async accionResena(
		id: string,
		accion: 'ocultar' | 'mostrar' | 'responder',
		respuesta = ''
	): Promise<Resultado> {
		const d = await this.demo();
		if (!todasLasResenas(d, await this.cambiosResenas()).some((r) => r.id === id))
			return 'no-existe';
		if (accion === 'ocultar') await guardarCambio(this.db, this.sb.id, 'resena-oculta', id, true);
		else if (accion === 'mostrar') await borrarCambio(this.db, this.sb.id, 'resena-oculta', id);
		else {
			const r = respuesta.trim();
			if (r.length > 1000) return 'no-se-puede';
			if (r) await guardarCambio(this.db, this.sb.id, 'resena-respuesta', id, { respuesta: r });
			else await borrarCambio(this.db, this.sb.id, 'resena-respuesta', id);
		}
		return 'ok';
	}

	/**
	 * La reseña que deja alguien en la tienda online. Con el código de un pedido cobrado (el firmado
	 * de la confirmación) queda con compra, una por pedido y producto; con un código que no es de un
	 * pedido cobrado, se guarda igual pero sin compra (como Go).
	 */
	async dejarResena(n: {
		codigo?: string;
		productoId?: string;
		autor: string;
		estrellas: number;
		texto?: string;
	}): Promise<'ok' | 'ya-reseno' | 'invalido'> {
		const autor = n.autor.trim();
		const texto = (n.texto ?? '').trim();
		if (
			autor.length < 2 ||
			autor.length > 60 ||
			!Number.isInteger(n.estrellas) ||
			n.estrellas < 1 ||
			n.estrellas > 5 ||
			texto.length > 1000
		)
			return 'invalido';
		let pedido: number | null = null;
		const numero = Number(n.codigo?.split('-')[0]);
		if (
			n.codigo &&
			Number.isInteger(numero) &&
			(await codigoPedido(this.sb.id, numero)) === n.codigo
		) {
			const p = await this.pedido(numero);
			if (p && COBRADO.includes(p.estadoPago)) pedido = numero;
		}
		const productoId =
			n.productoId && (await this.regProductos()).some((p) => p.id === n.productoId)
				? n.productoId
				: null;
		const ahora = Date.now();
		// Con compra, la clave es el pedido y el producto: así hay una sola por compra.
		const id =
			pedido !== null
				? `p${pedido}-${productoId ?? 'tienda'}`
				: `n${ahora.toString(36)}${randomToken(4).replace(/[^a-z0-9]/gi, '')}`;
		if (pedido !== null && (await cambiosDe(this.db, this.sb.id, 'resena-nueva')).get(id))
			return 'ya-reseno';
		await guardarCambio(this.db, this.sb.id, 'resena-nueva', id, {
			id,
			autor,
			estrellas: n.estrellas,
			texto,
			productoId,
			pedido,
			creada: ahora
		} satisfies ResenaNuevaDemo);
		return 'ok';
	}

	/** Una pregunta sobre un producto publicado; aparece en el panel junto a las de Mercado Libre. */
	async preguntar(q: {
		productoId: string;
		nombre: string;
		email?: string;
		pregunta: string;
	}): Promise<'ok' | 'invalido'> {
		const nombre = q.nombre.trim();
		const pregunta = q.pregunta.trim();
		const email = (q.email ?? '').trim();
		if (
			nombre.length < 2 ||
			nombre.length > 60 ||
			pregunta.length < 3 ||
			pregunta.length > 500 ||
			email.length > 120
		)
			return 'invalido';
		const p = (await this.regProductos()).find((x) => x.id === q.productoId);
		if (!p?.publicado) return 'invalido';
		const ahora = Date.now();
		const id = `t-${ahora.toString(36)}${randomToken(4).replace(/[^a-z0-9]/gi, '')}`;
		await guardarCambio(this.db, this.sb.id, 'pregunta-tienda', id, {
			id,
			productoId: p.id,
			nombre,
			email: email || null,
			pregunta,
			creada: ahora
		} satisfies PreguntaTiendaDemo);
		return 'ok';
	}

	/** La plantilla del sandbox (o la del rubro). */
	private plantilla() {
		return this.sb.plantilla ?? buscarRubro(this.sb.rubro)!.plantilla;
	}

	async resumen(): Promise<ResumenPanel> {
		const d = await this.demo();
		const leidas = (await cambiosDe(this.db, this.sb.id, 'notificaciones')).get('leidas') as
			{ hasta?: number } | undefined;
		const r = resumenDemo(d, this.sb.creadoEn, leidas?.hasta ?? 0, Date.now());
		const plan = (await this.estadoCuenta()).plan.actual ?? 'pro';
		r.prueba.plan = NOMBRE_PLAN[plan];
		// La insignia cuenta lo mismo que la pantalla: las tareas de hoy sin marcar.
		const tareas = (await this.paraHacer()).grupos.flatMap((g) => g.tareas);
		r.insignias.paraHacer = tareas.filter((t) => !t.hecha).length;
		// Los canales de la barra son los que están conectados en Integraciones.
		const { conectadas } = await this.integraciones();
		r.canales = r.canales.filter((c) => conectadas[c]);
		return r;
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
