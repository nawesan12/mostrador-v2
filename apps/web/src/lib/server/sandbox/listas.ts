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

// ---------------------------------------------------------------- Inventario

export type Movimiento = {
	productoId: string;
	q: number;
	motivo: 'ajuste' | 'conteo' | 'ingreso';
	en: number;
};

export type RegInventario = {
	id: string;
	nombre: string;
	foto: string | null;
	sku: string;
	stock: number;
	min: number;
	porDia: number;
	cover: number;
	st: 'out' | 'crit' | 'low' | 'ok' | 'over';
	fast: boolean;
	loc: string;
	proveedor: string;
	costo: number;
	publicadoEnMl: boolean;
};

const UBICACIONES = ['Local', 'Depósito', 'Local + Depósito'];
const PROVEEDORES = [
	'Distribuidora Norte',
	'Mayorista Sur',
	'Importadora Andes',
	'Fábrica del Litoral'
];

export function registrosInventario(productos: RegProducto[]): RegInventario[] {
	return productos.map((p) => {
		const r = semilla(p.id + 'inv');
		const porDia = Math.max(0.1, Math.round((p.vendidos30 / 30) * 10) / 10);
		const min = Math.max(5, Math.round(porDia * 10 + r * 10));
		const st: RegInventario['st'] =
			p.stock === 0
				? 'out'
				: p.stock < min * 0.5
					? 'crit'
					: p.stock < min
						? 'low'
						: p.stock > min * 3.2
							? 'over'
							: 'ok';
		return {
			id: p.id,
			nombre: p.nombre,
			foto: p.foto,
			sku: p.sku,
			stock: p.stock,
			min,
			porDia,
			cover: Math.floor(p.stock / porDia),
			st,
			fast: porDia > 2,
			loc: UBICACIONES[Math.floor(r * UBICACIONES.length)],
			// Mismo proveedor para toda una categoría.
			proveedor: PROVEEDORES[Math.floor(semilla(p.categoria) * PROVEEDORES.length)],
			costo: p.costo,
			publicadoEnMl: p.canales.includes('ml') && p.publicado
		};
	});
}

export const DEF_INVENTARIO: DefLista<RegInventario> = {
	buscar: (r) => `${r.nombre} ${r.sku} ${r.proveedor}`,
	segmento: 'st',
	facetas: ['loc', 'proveedor'],
	rango: { clave: 'cover', menor: true },
	banderas: ['fast'],
	ordenes: {
		cover: (a, b) => a.cover - b.cover,
		stock: (a, b) => a.stock - b.stock,
		value: (a, b) => b.stock * b.costo - a.stock * a.costo
	},
	resumen: (todos, l) => {
		const reponer = todos.filter((r) => r.st === 'out' || r.st === 'crit' || r.st === 'low');
		const semana = todos.filter((r) => r.st === 'out' || r.cover < 7);
		return {
			unidades: l.reduce((s, r) => s + r.stock, 0),
			valor: l.reduce((s, r) => s + r.stock * r.costo, 0),
			valorTotal: todos.reduce((s, r) => s + r.stock * r.costo, 0),
			sinStock: todos.filter((r) => r.st === 'out').length,
			sinStockEnMl: todos.filter((r) => r.st === 'out' && r.publicadoEnMl).length,
			bajo: todos.filter((r) => r.st === 'crit' || r.st === 'low').length,
			rotacion: Math.round(
				todos.reduce((s, r) => s + Math.min(120, r.cover), 0) / Math.max(1, todos.length)
			),
			reponer: reponer.length,
			terminanSemana: semana.length,
			proveedores: new Set(semana.map((r) => r.proveedor)).size
		};
	}
};

// ---------------------------------------------------------------- Facturación

/** Lo que el visitante hizo con un comprobante: enviarlo, reintentarlo o anularlo (nota de crédito). */
export type CambioFactura = { enviado?: boolean; autorizada?: boolean; anulada?: number };
export type FacturaManual = {
	cliente: string;
	tipo: 'A' | 'B' | 'C';
	cuit: string | null;
	concepto: string;
	total: number;
	en: number;
};

export type RegComprobante = {
	id: string;
	/** "0004-00012812"; las notas de crédito, "NC 0004-…"; sin número, "—". */
	numero: string;
	tipo: 'A' | 'B' | 'C';
	cliente: string;
	dias: number;
	estado: 'ok' | 'pend' | 'rej' | 'void' | 'nofac';
	/** Por qué ARCA la rechazó o por qué no se facturó (vacío si no hace falta). */
	motivo: string;
	total: number;
	canal: string;
	enviado: boolean;
	per: 'm' | 'q' | 'y';
	nomail: boolean;
	pedido: number | null;
};

/** El rechazo de ejemplo de la demo (facturas A con el CUIT del cliente mal cargado). */
export const MOTIVO_CUIT = 'El CUIT del receptor no es válido.';

/**
 * Los comprobantes de la demo. `sinFacturar` dice, para el momento de un cobro, si esa venta quedó
 * "no facturada" y por qué (Facturar al cobrar apagado), como el trigger de Go.
 */
export function registrosFacturacion(
	d: Demo,
	ahora: number,
	cambios: Map<string, CambioFactura>,
	manuales: [string, FacturaManual][],
	sinFacturar: (en: number) => string | null = () => null
): RegComprobante[] {
	const clientes = new Map(d.clientes.map((c) => [c.id, c]));
	const facturados = d.pedidos.filter((p) => p.facturado && p.creado <= ahora);
	const periodo = (dias: number) => (dias < 30 ? 'm' : dias < 90 ? 'q' : 'y');
	const regs: RegComprobante[] = facturados.map((p, i) => {
		const c = clientes.get(p.clienteId)!;
		const tipo = c.mayorista ? 'A' : 'B';
		const id = `${tipo}-${p.numero}`;
		const cambio = cambios.get(id) ?? {};
		const dias = (ahora - p.creado) / DIA;
		const motivoNo = sinFacturar(p.creado);
		const base: RegComprobante['estado'] = motivoNo
			? 'nofac'
			: p.estado === 'Cancelado'
				? 'void'
				: dias < 0.015
					? 'pend'
					: tipo === 'A' && p.numero % 7 === 0
						? 'rej'
						: 'ok';
		const autorizada = cambio.autorizada && (base === 'pend' || base === 'rej') ? 'ok' : base;
		const estado = cambio.anulada && autorizada === 'ok' ? 'void' : autorizada;
		const enviado = cambio.enviado ?? semilla(id) < 0.85;
		return {
			id,
			numero: estado === 'nofac' ? '—' : `0004-${String(12812 - i).padStart(8, '0')}`,
			tipo,
			cliente:
				tipo === 'A' ? `${c.nombre} SRL` : semilla(id + 'cf') < 0.4 ? 'Consumidor final' : c.nombre,
			dias,
			estado,
			motivo: estado === 'nofac' ? motivoNo! : estado === 'rej' ? MOTIVO_CUIT : '',
			total: p.total,
			canal: p.canal,
			enviado,
			per: periodo(dias),
			nomail: !enviado,
			pedido: p.numero
		};
	});
	const manualesReg = manuales.map(([id, m], k): RegComprobante => {
		const dias = (ahora - m.en) / DIA;
		const enviado = cambios.get(id)?.enviado ?? false;
		return {
			id,
			numero: `0004-${String(12813 + k).padStart(8, '0')}`,
			tipo: m.tipo,
			cliente: m.cliente,
			dias,
			estado: cambios.get(id)?.anulada ? 'void' : 'ok',
			motivo: '',
			total: m.total,
			canal: 'pos',
			enviado,
			per: 'm',
			nomail: !enviado,
			pedido: null
		};
	});
	// Cada anulación es una nota de crédito autorizada, con su propio número (como en Go).
	const notas = [...manualesReg, ...regs]
		.filter((r) => r.estado === 'void' && cambios.get(r.id)?.anulada)
		.map((r) => [r, cambios.get(r.id)!.anulada!] as const)
		.sort((a, b) => a[1] - b[1])
		.map(([r, en], k): RegComprobante => {
			const dias = (ahora - en) / DIA;
			const enviado = cambios.get(`NC-${r.id}`)?.enviado ?? false;
			return {
				...r,
				id: `NC-${r.id}`,
				numero: `NC 0004-${String(301 + k).padStart(8, '0')}`,
				dias,
				estado: 'ok',
				motivo: '',
				enviado,
				per: periodo(dias),
				nomail: !enviado
			};
		});
	return [...notas.reverse(), ...manualesReg.reverse(), ...regs];
}

export const DEF_FACTURACION: DefLista<RegComprobante> = {
	buscar: (r) => `${r.cliente} ${r.numero} ${r.pedido ?? ''}`,
	segmento: 'estado',
	facetas: ['tipo', 'canal', 'per'],
	rango: { clave: 'total' },
	banderas: ['nomail'],
	ordenes: {
		new: (a, b) => a.dias - b.dias,
		total: (a, b) => b.total - a.total
	},
	resumen: (todos, l) => ({
		facturado: l
			.filter((r) => r.estado === 'ok' && !r.numero.startsWith('NC'))
			.reduce((s, r) => s + r.total, 0),
		ultimos30: todos
			.filter((r) => r.estado === 'ok' && !r.numero.startsWith('NC') && r.dias < 30)
			.reduce((s, r) => s + r.total, 0),
		porResolver: todos.filter((r) => r.estado === 'pend' || r.estado === 'rej').length
	})
};

// ---------------------------------------------------------------- Envíos

export type RegEnvio = {
	id: string;
	pedido: string;
	cliente: string;
	ciudad: string;
	transporte: 'andreani' | 'oca' | 'correo' | 'moto';
	step: number;
	estado: 'todo' | 'way' | 'late' | 'done' | 'ret';
	dias: number;
	costo: number;
	seguimiento: string;
	reclamado: boolean;
	wk: boolean;
};

const COSTO_ENVIO: Record<RegEnvio['transporte'], [number, number]> = {
	andreani: [4200, 3600],
	oca: [3800, 2700],
	correo: [3500, 2500],
	moto: [2500, 1000]
};

export function registrosEnvios(
	pedidos: (Pedido & { cambio?: { en?: Partial<Record<string, number>> } })[],
	clientes: Demo['clientes'],
	ahora: number,
	reclamados: Set<string>
): RegEnvio[] {
	const porId = new Map(clientes.map((c) => [c.id, c]));
	return pedidos
		.filter(
			(p) =>
				p.envio !== 'retiro' &&
				p.creado <= ahora &&
				(p.estado === 'Listo para despachar' ||
					p.estado === 'En camino' ||
					p.estado === 'Entregado')
		)
		.map((p) => {
			const t = p.envio as RegEnvio['transporte'];
			const id = String(p.numero);
			const despachado = p.cambio?.en?.['En camino'] ?? p.creado + DIA;
			const enViaje = Math.max(0, (ahora - despachado) / DIA);
			const step =
				p.estado === 'Listo para despachar'
					? 0
					: p.estado === 'Entregado'
						? 4
						: enViaje < 1
							? 1
							: enViaje < 2
								? 2
								: 3;
			const estado: RegEnvio['estado'] =
				step === 0
					? 'todo'
					: step === 4
						? p.numero % 41 === 0
							? 'ret'
							: 'done'
						: enViaje > 3 && t !== 'moto'
							? 'late'
							: 'way';
			const [base, extra] = COSTO_ENVIO[t];
			const dias = (ahora - p.creado) / DIA;
			return {
				id,
				pedido: `#${p.numero}`,
				cliente: porId.get(p.clienteId)?.nombre ?? '',
				ciudad: p.ciudadEnvio,
				transporte: t,
				step,
				estado,
				dias,
				costo: Math.round((base + semilla(id) * extra) / 100) * 100,
				seguimiento: t === 'moto' ? 'Moto propia' : `AR${300000000 + p.numero * 37}`,
				reclamado: reclamados.has(id),
				wk: dias < 7
			};
		});
}

export const DEF_ENVIOS: DefLista<RegEnvio> = {
	buscar: (r) => `${r.cliente} ${r.pedido} ${r.ciudad} ${r.seguimiento}`,
	segmento: 'estado',
	facetas: ['transporte', 'ciudad'],
	rango: { clave: 'costo' },
	banderas: ['wk'],
	ordenes: {
		new: (a, b) => a.dias - b.dias,
		late: (a, b) => Number(b.estado === 'late') - Number(a.estado === 'late') || a.dias - b.dias,
		cost: (a, b) => b.costo - a.costo
	},
	resumen: (todos, l) => ({
		costos: l.reduce((s, r) => s + r.costo, 0),
		demorados: todos.filter((r) => r.estado === 'late' && !r.reclamado).length,
		porDespachar: todos.filter((r) => r.estado === 'todo').length
	})
};
