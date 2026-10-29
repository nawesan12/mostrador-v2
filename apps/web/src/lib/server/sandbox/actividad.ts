/**
 * Actividad de la demo (design/PanelCrecer.dc.html, sAct). Lo que hizo el visitante sale de sus
 * cambios guardados (son su historial real); lo de Tincho, el equipo y el sistema, de la tienda de
 * ejemplo. Se puede deshacer lo que vuelve a quedar como estaba con sólo borrar el cambio.
 */
import type { components } from '@mostrador/contrato/esquema';
import type { Demo } from './demo.ts';

type Evento = components['schemas']['Evento'];

const MIN = 60_000;
const iso = (t: number) => new Date(t).toISOString();
const pesos = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');

/** Una fila de la tabla de cambios. */
export type FilaCambio = { entidad: string; clave: string; datos: unknown; en: number };

/** Lo que se necesita para nombrar las cosas. */
export type Nombres = {
	productos: Map<string, string>;
	cupones: Map<string, string>;
	automaticas: Map<string, string>;
	canales: Map<string, string>;
	totales: Map<number, number>;
};

/** Entidades que se deshacen borrando el cambio. */
const SE_DESHACEN = new Set([
	'producto',
	'stock',
	'cupon',
	'cupon-nuevo',
	'automatica',
	'canal',
	'cliente'
]);

/** Quien atiende la caja del local en la demo (la dueña, Valeria, es quien la prueba). */
export const EQUIPO = { nombre: 'Carla', foto: '1520333789090-1afc82db536a' };

function texto(f: FilaCambio, n: Nombres): [string, string] | null {
	const d = (f.datos ?? {}) as Record<string, unknown>;
	const num = Number(f.clave);
	switch (f.entidad) {
		case 'pedido':
			if (d.recordado && !d.estado) return [`le recordó el pago del pedido #${f.clave}`, 'Pedidos'];
			return d.estado
				? [`pasó el pedido #${f.clave} a ${String(d.estado).toLowerCase()}`, 'Pedidos']
				: null;
		case 'pedido-nuevo': {
			const p = (d.pedido ?? {}) as { canal?: string; total?: number };
			return p.canal === 'pos'
				? [
						`cobró ${pesos(p.total ?? n.totales.get(num) ?? 0)} en la caja del local`,
						'Punto de venta'
					]
				: [`cargó el pedido #${f.clave} por ${pesos(p.total ?? 0)}`, 'Pedidos'];
		}
		case 'producto': {
			const nombre = (d.nombre as string) ?? n.productos.get(f.clave) ?? 'un producto';
			if (d.nuevo) return [`creó el producto ${nombre}`, 'Catálogo'];
			if (d.precio !== undefined)
				return [`cambió el precio de ${nombre} a ${pesos(d.precio as number)}`, 'Catálogo'];
			return [`editó ${nombre}`, 'Catálogo'];
		}
		case 'stock':
			return [`ajustó el stock de ${n.productos.get(f.clave) ?? 'un producto'}`, 'Inventario'];
		case 'cupon': {
			const codigo = n.cupones.get(f.clave) ?? 'un cupón';
			if (d.minimo)
				return [`le puso un mínimo de ${pesos(d.minimo as number)} a ${codigo}`, 'Cupones'];
			if (d.estado === 'pausado') return [`pausó el cupón ${codigo}`, 'Cupones'];
			if (d.estado === 'activo') return [`activó el cupón ${codigo}`, 'Cupones'];
			return [`editó el cupón ${codigo}`, 'Cupones'];
		}
		case 'cupon-nuevo':
			return [`creó el cupón ${(d.codigo as string) ?? ''}`.trim(), 'Cupones'];
		case 'automatica':
			return [
				`${d.activa ? 'prendió' : 'apagó'} ${(n.automaticas.get(f.clave) ?? 'una promoción').toLowerCase()}`,
				'Cupones'
			];
		case 'canal':
			return [
				`${d.pausado ? 'pausó' : 'activó'} ${n.canales.get(f.clave) ?? 'un canal'}`,
				'Canales'
			];
		case 'campana':
			return [
				d.programada
					? `programó una campaña para ${d.personas} personas`
					: `mandó una campaña a ${d.personas} personas`,
				'Marketing'
			];
		case 'link':
			return [
				d.pagado
					? `cobró un link de pago por ${pesos(d.monto as number)}`
					: `armó un link de pago por ${pesos(d.monto as number)}`,
				'Link de pago'
			];
		case 'aumento':
			return f.clave === 'programado'
				? [`programó un aumento del ${d.pct}% para el lunes`, 'Listas de precio']
				: [
						`subió los precios de ${Object.keys((d.antes as object) ?? {}).length} productos`,
						'Listas de precio'
					];
		case 'devolucion':
			return [
				`${d.estado === 'Aprobada' ? 'aprobó' : 'ofreció cambio en'} una devolución`,
				'Devoluciones'
			];
		case 'pregunta':
			return ['respondió una pregunta en Mercado Libre', 'Reseñas'];
		case 'conversacion':
			return (d.mensajes as unknown[] | undefined)?.length
				? ['respondió un chat', 'WhatsApp']
				: null;
		case 'factura-manual':
			return ['emitió una factura manual', 'Facturación'];
		case 'cliente':
			return ['editó la ficha de un cliente', 'Clientes'];
		case 'carrito':
			return ['le recordó el carrito a un cliente', 'Carritos'];
	}
	return null;
}

/** La última copia de las 04:00 de Argentina (07:00 UTC). */
function copiaDiaria(ahora: number) {
	const dia = 86_400_000;
	const hoy = Math.floor(ahora / dia) * dia + 7 * 3_600_000;
	return hoy <= ahora ? hoy : hoy - dia;
}

export function actividadDemo(
	d: Demo,
	ahora: number,
	creado: number,
	filas: FilaCambio[],
	deshechos: { id: string; texto: string; donde: string; en: number }[],
	n: Nombres
): Evento[] {
	const propios: Evento[] = filas.flatMap((f) => {
		const t = texto(f, n);
		if (!t) return [];
		// Un producto creado o un cupón propio se "deshace" borrándolo: también vale.
		return [
			{
				id: `${f.entidad}:${f.clave}`,
				tipo: 'equipo' as const,
				quien: 'Vos',
				foto: null,
				texto: t[0],
				cuando: iso(f.en),
				donde: t[1],
				deshacer: SE_DESHACEN.has(f.entidad),
				deshecho: false
			}
		];
	});
	const viejos: Evento[] = deshechos.map((x) => ({
		id: x.id,
		tipo: 'equipo',
		quien: 'Vos',
		foto: null,
		texto: x.texto,
		cuando: iso(x.en),
		donde: x.donde,
		deshacer: false,
		deshecho: true
	}));
	const hoyLocal = d.pedidos.filter((p) => p.canal === 'pos' && p.creado <= ahora).slice(0, 2);
	const ventasHornero = d.pedidos.filter(
		(p) => p.canal === 'wa' && p.creado > ahora - 1440 * MIN && p.creado <= ahora
	);
	const demo: Evento[] = [
		{
			id: 'd-hornero',
			tipo: 'hornero',
			quien: 'Tincho',
			foto: null,
			texto: `respondió ${d.conversaciones.length * 9 + 14} consultas y cerró ${ventasHornero.length} ventas por ${pesos(ventasHornero.reduce((s, p) => s + p.total, 0))}`,
			cuando: iso(ahora - 5 * MIN),
			donde: 'WhatsApp e Instagram',
			deshacer: false,
			deshecho: false
		},
		...hoyLocal.map((p, i): Evento => ({
			id: `d-caja-${p.numero}`,
			tipo: 'equipo',
			quien: EQUIPO.nombre,
			foto: EQUIPO.foto,
			texto: `cobró ${pesos(p.total)} en ${p.pago === 'efectivo' ? 'efectivo' : 'la caja'} del local`,
			cuando: iso(Math.min(p.creado, ahora - (60 + i * 90) * MIN)),
			donde: 'Punto de venta',
			deshacer: false,
			deshecho: false
		})),
		{
			id: 'd-ml',
			tipo: 'sistema',
			quien: 'Sistema',
			foto: null,
			texto: `sincronizó ${d.enMl.length} publicaciones con Mercado Libre`,
			cuando: iso(ahora - 120 * MIN),
			donde: 'Automático',
			deshacer: false,
			deshecho: false
		},
		{
			id: 'd-fotos',
			tipo: 'hornero',
			quien: 'Tincho',
			foto: null,
			texto: `mejoró las fotos de ${Math.min(12, d.catalogo.productos.length)} productos`,
			cuando: iso(ahora - 300 * MIN),
			donde: 'Catálogo',
			deshacer: false,
			deshecho: false
		},
		{
			id: 'd-ingreso',
			tipo: 'sistema',
			quien: 'Seguridad',
			foto: null,
			texto: 'nuevo inicio de sesión en la demo',
			cuando: iso(creado),
			donde: 'Este navegador',
			deshacer: false,
			deshecho: false
		},
		{
			id: 'd-copia',
			tipo: 'sistema',
			quien: 'Sistema',
			foto: null,
			texto: 'hizo una copia de seguridad completa',
			cuando: iso(copiaDiaria(ahora)),
			donde: 'Copia diaria a las 04:00',
			deshacer: false,
			deshecho: false
		}
	];
	return [...propios, ...viejos, ...demo].sort(
		(a, b) => Date.parse(b.cuando) - Date.parse(a.cuando)
	);
}
