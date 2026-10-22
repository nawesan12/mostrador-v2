/**
 * Motor de listas de la demo (design/PanelListas.dc.html): cada lista arma sus registros con los
 * datos de la tienda (clientes y productos salen de los mismos pedidos que el resto del panel) y un
 * filtro genérico los recorta, ordena y cuenta: cada segmento y cada faceta se cuentan sin su propio
 * filtro, como en el handoff.
 */
import type { components } from '@mostrador/contrato/esquema';
import type { Demo, Pedido } from './demo.ts';

type ListaGenerica = components['schemas']['ListaGenerica'];

const DIA = 86_400_000;

export type FiltrosLista = {
	q?: string;
	seg?: string[];
	facetas?: Record<string, string[]>;
	rango?: number;
	banderas?: string[];
	orden?: string;
	desde?: number;
	cuantos?: number;
};

/** Cómo filtra y ordena una lista (los textos y colores viven en el cliente). */
export type DefLista<R> = {
	buscar: (r: R) => string;
	segmento: keyof R;
	facetas: (keyof R)[];
	/** Campo del deslizador; `menor`: filtra por menor (si no, desde). */
	rango: { clave: keyof R; menor?: boolean };
	banderas: (keyof R)[];
	/** La primera es la de por defecto. */
	ordenes: Record<string, (a: R, b: R) => number>;
	resumen: (todos: R[], filtrados: R[]) => Record<string, number>;
};

const sinAcento = (s: string) => s.normalize('NFD').replace(/[̀-ͯ]/g, '').toLowerCase();

export function listarGenerico<R extends { id: string }>(
	registros: R[],
	def: DefLista<R>,
	f: FiltrosLista
): ListaGenerica {
	const q = sinAcento(f.q?.trim() ?? '');
	const banderas = f.banderas ?? [];
	const base = registros.filter((r) => {
		if (q && !sinAcento(def.buscar(r)).includes(q)) return false;
		if (f.rango) {
			const v = Number(r[def.rango.clave]);
			if (def.rango.menor ? !(v < f.rango) : !(v >= f.rango)) return false;
		}
		return banderas.every((b) => !def.banderas.includes(b as keyof R) || Boolean(r[b as keyof R]));
	});
	const valores = (r: R, k: keyof R) => {
		const v = r[k];
		return Array.isArray(v) ? v.map(String) : [String(v)];
	};
	const elegidos = (k: keyof R) =>
		k === def.segmento ? (f.seg ?? []) : (f.facetas?.[String(k)] ?? []);
	const pasa = (r: R, sin?: keyof R) =>
		[def.segmento, ...def.facetas].every((k) => {
			if (k === sin) return true;
			const e = elegidos(k);
			return !e.length || valores(r, k).some((v) => e.includes(v));
		});
	const contar = (k: keyof R) => {
		const m: Record<string, number> = {};
		for (const r of base) if (pasa(r, k)) for (const v of valores(r, k)) m[v] = (m[v] ?? 0) + 1;
		return m;
	};
	const orden = def.ordenes[f.orden ?? ''] ?? Object.values(def.ordenes)[0];
	const lista = base.filter((r) => pasa(r)).sort(orden);
	const desde = Math.max(0, f.desde ?? 0);
	const cuantos = Math.min(200, Math.max(1, f.cuantos ?? 60));
	return {
		total: lista.length,
		filas: lista.slice(desde, desde + cuantos) as unknown as ListaGenerica['filas'],
		segmentos: contar(def.segmento),
		facetas: Object.fromEntries(def.facetas.map((k) => [String(k), contar(k)])),
		resumen: def.resumen(registros, lista)
	};
}

/** Número estable a partir de un texto (para datos de ejemplo que no cambian entre pedidos). */
const semilla = (s: string) => {
	let h = 2166136261;
	for (const c of s) h = Math.imul(h ^ c.charCodeAt(0), 16777619);
	return (h >>> 0) / 4294967296;
};

// ---------------------------------------------------------------- Clientes

export type RegCliente = {
	id: string;
	nombre: string;
	foto: string | null;
	/** Color de las iniciales (0–4). */
	ini: number;
	ciudad: string;
	canal: string;
	telefono: string | null;
	pedidos: number;
	/** Días desde la última compra. */
	ultima: number;
	gastado: number;
	seg: 'vip' | 'freq' | 'new' | 'risk' | 'sleep';
	/** Compras por mes, los últimos 12 (del más viejo al actual). */
	meses: number[];
	cuit: boolean;
	wa: boolean;
	cupon: boolean;
	etiquetas: string[];
};

/** Lo que el visitante cambió o creó de un cliente. */
export type CambioCliente = {
	cupon?: number;
	etiquetas?: string[];
	nuevo?: {
		nombre: string;
		telefono: string | null;
		email: string | null;
		ciudad: string;
		creado: number;
	};
};

export function registrosClientes(
	d: Demo,
	base: number,
	ahora: number,
	cambios: Map<string, CambioCliente>
): RegCliente[] {
	const suyos = new Map<string, Pedido[]>();
	for (const p of d.pedidos)
		if (p.estado !== 'Cancelado' && p.creado <= ahora)
			suyos.set(p.clienteId, [...(suyos.get(p.clienteId) ?? []), p]);
	const ticket = d.pedidos.reduce((s, p) => s + p.total, 0) / Math.max(1, d.pedidos.length);
	const nuevos: RegCliente[] = [...cambios]
		.filter(([, c]) => c.nuevo)
		.map(([id, c]) => ({
			id,
			nombre: c.nuevo!.nombre,
			foto: null,
			ini: Math.floor(semilla(id) * 5),
			ciudad: c.nuevo!.ciudad,
			canal: 'pos',
			telefono: c.nuevo!.telefono,
			pedidos: 0,
			ultima: Math.max(0, (ahora - c.nuevo!.creado) / DIA),
			gastado: 0,
			seg: 'new',
			meses: Array(12).fill(0),
			cuit: false,
			wa: !!c.nuevo!.telefono,
			cupon: !!c.cupon,
			etiquetas: c.etiquetas ?? []
		}));
	return [
		...nuevos,
		...d.clientes.map((c) => {
			const ps = suyos.get(c.id) ?? [];
			const r = semilla(c.id);
			// Quien ya era cliente antes de la ventana de la demo trae compras anteriores.
			const previos =
				c.alta < base - 60 * DIA ? Math.floor(r * (ps.length ? 10 : 6)) + (ps.length ? 0 : 1) : 0;
			const pedidos = ps.length + previos;
			const ultimaCompra = ps.length
				? Math.max(...ps.map((p) => p.creado))
				: base - (61 + Math.floor(r * 340)) * DIA;
			const ultima = Math.max(0, (ahora - ultimaCompra) / DIA);
			const gastado =
				ps.reduce((s, p) => s + p.total, 0) +
				Math.round((previos * ticket * (0.6 + r)) / 100) * 100;
			const seg: RegCliente['seg'] =
				pedidos >= 12 && ultima < 60
					? 'vip'
					: pedidos >= 4 && ultima < 90
						? 'freq'
						: pedidos <= 1 && ultima < 30
							? 'new'
							: ultima > 180
								? 'sleep'
								: ultima > 75
									? 'risk'
									: 'freq';
			const meses = Array.from({ length: 12 }, (_, k) => {
				const d1 = ahora - (11 - k) * 30 * DIA;
				const d0 = d1 - 30 * DIA;
				const enDatos = ps.filter((p) => p.creado >= d0 && p.creado < d1).length;
				// Antes de la ventana de datos, las compras anteriores se reparten en los meses que tocan.
				const antes =
					d1 < base - 60 * DIA && previos && ultimaCompra >= d0 && semilla(c.id + k) < 0.6 ? 1 : 0;
				return Math.min(3, enDatos + antes);
			});
			const primero = ps.length ? ps.reduce((a, b) => (a.creado < b.creado ? a : b)) : null;
			return {
				id: c.id,
				nombre: c.nombre,
				foto: c.foto,
				ini: Math.floor(r * 5),
				ciudad: c.ciudad,
				canal: primero?.canal ?? 'web',
				telefono: c.telefono,
				pedidos,
				ultima,
				gastado,
				seg,
				meses,
				cuit: c.mayorista,
				wa: !!c.telefono,
				cupon: !!cambios.get(c.id)?.cupon,
				etiquetas: cambios.get(c.id)?.etiquetas ?? []
			};
		})
	];
}

export const DEF_CLIENTES: DefLista<RegCliente> = {
	buscar: (r) => `${r.nombre} ${r.ciudad} ${r.telefono ?? ''} ${r.etiquetas.join(' ')}`,
	segmento: 'seg',
	facetas: ['canal', 'ciudad'],
	rango: { clave: 'gastado' },
	banderas: ['wa', 'cuit'],
	ordenes: {
		ltv: (a, b) => b.gastado - a.gastado,
		last: (a, b) => a.ultima - b.ultima,
		orders: (a, b) => b.pedidos - a.pedidos,
		az: (a, b) => a.nombre.localeCompare(b.nombre, 'es')
	},
	resumen: (todos, l) => ({
		gastado: l.reduce((s, r) => s + r.gastado, 0),
		enRiesgo: todos.filter((r) => r.seg === 'risk' && !r.cupon).length
	})
};

// ---------------------------------------------------------------- Productos

export type CambioProducto = {
	nombre?: string;
	descripcion?: string;
	categoria?: string | null;
	precio?: number;
	costo?: number;
	sku?: string;
	stock?: number;
	foto?: string | null;
	canales?: Record<string, boolean>;
	publicado?: boolean;
	/** Producto creado por el visitante. */
	nuevo?: boolean;
	creado?: number;
};

export type RegProducto = {
	id: string;
	nombre: string;
	descripcion: string;
	foto: string | null;
	sku: string;
	categoria: string;
	precio: number;
	costo: number;
	margen: number;
	stock: number;
	/** Unidades por semana, las últimas 8 (de la más vieja a la actual). */
	ventas: number[];
	vendidos30: number;
	canales: string[];
	publicado: boolean;
	estado: 'active' | 'out' | 'paused' | 'draft';
	lowm: boolean;
	star: boolean;
};

export function registrosProductos(
	d: Demo,
	ahora: number,
	cambios: Map<string, CambioProducto>,
	ajustes: Map<string, number>
): RegProducto[] {
	const vendidos = new Map<string, number[]>();
	for (const p of d.pedidos) {
		if (p.estado === 'Cancelado' || p.creado > ahora) continue;
		const semana = Math.floor((ahora - p.creado) / (7 * DIA));
		if (semana > 7) continue;
		for (const i of p.items) {
			const v = vendidos.get(i.productoId) ?? Array(8).fill(0);
			v[7 - semana] += i.cantidad;
			vendidos.set(i.productoId, v);
		}
	}
	const ultimos30 = new Map<string, number>();
	for (const p of d.pedidos)
		if (p.estado !== 'Cancelado' && p.creado <= ahora && p.creado > ahora - 30 * DIA)
			for (const i of p.items)
				ultimos30.set(i.productoId, (ultimos30.get(i.productoId) ?? 0) + i.cantidad);

	const delCatalogo = d.catalogo.productos.map((p, i) => {
		const r = semilla(p.id);
		const margen = Math.round(18 + r * 42);
		const c = cambios.get(p.id) ?? {};
		const canales = c.canales
			? Object.keys(c.canales).filter((k) => c.canales![k])
			: [
					'web',
					'wa',
					'pos',
					...(d.enMl.includes(p.id) ? ['ml'] : []),
					...(semilla(p.id + 'ig') < 0.7 ? ['ig'] : [])
				];
		const precio = c.precio ?? p.precio;
		return armar({
			id: p.id,
			nombre: c.nombre ?? p.nombre,
			descripcion: c.descripcion ?? p.detalle ?? '',
			foto: c.foto !== undefined ? c.foto : p.foto,
			sku: c.sku ?? `SKU-${1000 + i}`,
			categoria: (c.categoria ?? p.categoria) || 'Sin categoría',
			precio,
			costo: c.costo ?? Math.round((precio * (1 - margen / 100)) / 10) * 10,
			stock: Math.max(0, (c.stock ?? d.stock[p.id] ?? 0) + (ajustes.get(p.id) ?? 0)),
			ventas: vendidos.get(p.id) ?? Array(8).fill(0),
			vendidos30: ultimos30.get(p.id) ?? 0,
			canales,
			publicado: c.publicado ?? true
		});
	});
	const nuevos = [...cambios]
		.filter(([, c]) => c.nuevo)
		.sort((a, b) => (b[1].creado ?? 0) - (a[1].creado ?? 0))
		.map(([id, c]) =>
			armar({
				id,
				nombre: c.nombre ?? 'Producto nuevo',
				descripcion: c.descripcion ?? '',
				foto: c.foto ?? null,
				sku: c.sku ?? '',
				categoria: c.categoria || 'Sin categoría',
				precio: c.precio ?? 0,
				costo: c.costo ?? 0,
				stock: Math.max(0, (c.stock ?? 0) + (ajustes.get(id) ?? 0)),
				ventas: Array(8).fill(0),
				vendidos30: 0,
				canales: Object.keys(c.canales ?? { web: true }).filter((k) => c.canales?.[k] ?? true),
				publicado: c.publicado ?? false
			})
		);
	return [...nuevos, ...delCatalogo];
}

function armar(p: Omit<RegProducto, 'margen' | 'estado' | 'lowm' | 'star'>): RegProducto {
	const margen = p.precio ? Math.round(((p.precio - p.costo) / p.precio) * 100) : 0;
	const estado: RegProducto['estado'] = !p.publicado
		? p.foto
			? 'paused'
			: 'draft'
		: p.stock === 0
			? 'out'
			: 'active';
	return { ...p, margen, estado, lowm: margen < 25, star: p.vendidos30 > 60 };
}

export const DEF_PRODUCTOS: DefLista<RegProducto> = {
	buscar: (r) => `${r.nombre} ${r.sku} ${r.categoria}`,
	segmento: 'estado',
	facetas: ['categoria', 'canales'],
	rango: { clave: 'precio' },
	banderas: ['lowm', 'star'],
	ordenes: {
		sold: (a, b) => b.vendidos30 - a.vendidos30,
		stock: (a, b) => a.stock - b.stock,
		price: (a, b) => b.precio - a.precio,
		margin: (a, b) => b.margen - a.margen
	},
	resumen: (todos, l) => ({
		unidades30: l.reduce((s, r) => s + r.vendidos30, 0),
		sinStockEnMl: todos.filter((r) => r.estado === 'out' && r.canales.includes('ml')).length,
		precioMax: Math.max(10_000, ...todos.map((r) => r.precio))
	})
};

export const slugProducto = (nombre: string) =>
	sinAcento(nombre)
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-|-$/g, '')
		.slice(0, 60);
