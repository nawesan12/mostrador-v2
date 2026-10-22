/**
 * El Resumen del panel (dashboard) en la demo, calculado sobre los datos de la tienda de ejemplo
 * (demo.ts): ventas, pedidos, canales, meta, caja, liquidaciones, envíos… Todo sale de los mismos
 * pedidos que después se ven en Pedidos, así los números cierran entre pantallas.
 *
 * Las fechas se cuentan en hora de Argentina (UTC−3, sin horario de verano).
 */
import type { components } from '@mostrador/contrato/esquema';
import type { Demo, Pedido } from './demo.ts';

type Tablero = components['schemas']['Tablero'];
type ContextoHornero = components['schemas']['ContextoHornero'];
export type Periodo = Tablero['periodo'];
export const PERIODOS: Periodo[] = ['hoy', '7d', '30d'];

const MIN = 60_000;
const HORA = 60 * MIN;
const DIA = 24 * HORA;
const AR = -3 * HORA;

/** Comienzo del día (en Argentina) que contiene `t`. */
export const inicioDelDia = (t: number) => t - ((((t + AR) % DIA) + DIA) % DIA);
/** Hora del día en Argentina, con fracción (9.5 = 9:30). */
export const horaAr = (t: number) => (t - inicioDelDia(t)) / HORA;
const fechaAr = (t: number) => new Date(t + AR).toISOString().slice(0, 10);

const vale = (p: Pedido) => p.estado !== 'Cancelado';

/** Lo que el visitante cambió y el tablero respeta. */
export type CambiosTablero = {
	meta?: number;
	certificadoOculto?: boolean;
	/** Ideas de Hornero aprobadas (carritos = recordatorio mandado). */
	ideasHechas?: string[];
};

/** Días que faltan para que venza el certificado de ARCA de la tienda de ejemplo. */
const DIAS_CERTIFICADO = 5;
/** Lo que cuesta un mensaje de marketing por WhatsApp (Meta, Argentina). */
const COSTO_MENSAJE = 120;
/** Día de la semana en que acredita cada medio (null: todos los días). 3 = miércoles. */
const LIQUIDA: Record<string, number | null> = { mp: null, payway: 3, modo: 4, uala: 5 };
/** Horario de retiro de cada transporte. */
const RETIROS = { andreani: '15:00', moto: '16:30', correo: '17:00', oca: '18:00' } as const;

function paraHacer(d: Demo, ahora: number) {
	const transferencias = d.pedidos.filter(
		(p) => p.estado === 'Pago pendiente' && p.pago === 'transf'
	);
	const sinStock = d.catalogo.productos.filter((p) => d.stock[p.id] === 0);
	return {
		preparar: d.pedidos.filter((p) => p.estado === 'Para preparar').length,
		chats: d.conversaciones.length,
		chatMasViejoMin: d.conversaciones.length
			? Math.round((ahora - Math.min(...d.conversaciones.map((c) => c.creado))) / MIN)
			: 0,
		transferencias: transferencias.length,
		transferenciasMonto: transferencias.reduce((s, p) => s + p.total, 0),
		sinStock: sinStock.length,
		sinStockEnMl: sinStock.filter((p) => d.enMl.includes(p.id)).length
	};
}

/** Cuántas tareas de "Para hacer hoy" tienen algo pendiente (la insignia del menú). */
export function tareasDeHoy(d: Demo, ahora = Date.now()) {
	const t = paraHacer(d, ahora);
	return [t.preparar, t.chats, t.transferencias, t.sinStock].filter((n) => n > 0).length;
}

/** Conversión del día (visitas que compran): entre 2,8 % y 4 %, fija para cada tienda y día. */
function conversionDelDia(plantilla: string, dia: number) {
	let h = 2166136261;
	for (const c of `${plantilla}:${dia}`) h = Math.imul(h ^ c.charCodeAt(0), 16777619);
	return 2.8 + (((h >>> 0) % 1000) / 1000) * 1.2;
}

export function tableroDemo(
	d: Demo,
	periodo: Periodo,
	opciones: { base: number; ahora: number; hola: string | null; cambios: CambiosTablero }
): Tablero {
	const { base, ahora, cambios } = opciones;
	const { pedidos, clientes, catalogo } = d;
	const cliente = new Map(clientes.map((c) => [c.id, c]));
	const hoy = inicioDelDia(ahora);
	const entre = (desde: number, hasta: number) =>
		pedidos.filter((p) => vale(p) && p.creado >= desde && p.creado < hasta);
	const suma = (ps: Pedido[]) => ps.reduce((s, p) => s + p.total, 0);

	// La ventana del período y la anterior del mismo largo ("vs. ayer", "vs. semana anterior"…).
	const desde = periodo === 'hoy' ? hoy : ahora - (periodo === '7d' ? 7 : 30) * DIA;
	const hasta = periodo === 'hoy' ? Math.max(ahora, hoy + 2 * HORA) : ahora;
	const largo = hasta - desde;
	const corrimiento = periodo === 'hoy' ? DIA : largo;
	const actual = entre(desde, hasta);
	const anterior = entre(desde - corrimiento, hasta - corrimiento);

	const acumulada = (d0: number) =>
		Array.from({ length: 14 }, (_, k) => suma(entre(d0, d0 + ((k + 1) * largo) / 14)));
	const conversion = (d0: number, d1: number) => {
		const dias: number[] = [];
		for (let t = inicioDelDia(d0); t < d1; t += DIA)
			dias.push(conversionDelDia(catalogo.plantilla, t / DIA));
		return Math.round((dias.reduce((s, x) => s + x, 0) / dias.length) * 10) / 10;
	};
	const nuevos = (d0: number, d1: number) =>
		clientes.filter((c) => !c.mayorista && c.alta >= d0 && c.alta < d1).length;
	const ticket = (ps: Pedido[]) => (ps.length ? Math.round(suma(ps) / ps.length) : 0);

	const porCanal = new Map<string, number>();
	for (const p of actual) porCanal.set(p.canal, (porCanal.get(p.canal) ?? 0) + p.total);
	const totalActual = suma(actual);

	// Meta del mes: la del visitante o, si no puso, un 10 % arriba del ritmo de los últimos 30 días.
	const local = new Date(ahora + AR);
	const mesInicio = Date.UTC(local.getUTCFullYear(), local.getUTCMonth(), 1) - AR;
	const diasMes = new Date(
		Date.UTC(local.getUTCFullYear(), local.getUTCMonth() + 1, 0)
	).getUTCDate();
	const delMes = suma(entre(mesInicio, ahora));
	const transcurrido = Math.max(0.5, (ahora - mesInicio) / DIA);
	const ritmo = (suma(entre(base - 30 * DIA, base)) / 30) * diasMes;
	const objetivo = cambios.meta ?? Math.round((ritmo * 1.1) / 100_000) * 100_000;

	// Caja del local: abre a las 9 (y unos minutos) y cierra a las 21.
	const posHoy = entre(hoy, ahora).filter((p) => p.canal === 'pos');
	const medios = new Map<string, number>();
	for (const p of posHoy) medios.set(p.pago, (medios.get(p.pago) ?? 0) + p.total);
	const apertura = hoy + 9 * HORA + (2 + (catalogo.plantilla.length % 7)) * MIN;
	const abierta = ahora >= apertura && horaAr(ahora) < 21;

	// Liquidaciones de los próximos 7 días, una por medio. Mercado Pago acredita cada día lo del día
	// anterior; los demás, una vez por semana (su día), lo cobrado entre 9 y 2 días antes.
	const liq = new Map<string, { fecha: number; monto: number }>();
	const cobradoCon = (medio: string, d0: number, d1: number) =>
		suma(
			entre(d0, Math.min(d1, ahora)).filter(
				(p) => p.pago === medio && p.estado !== 'Pago pendiente'
			)
		);
	for (const [medio, diaSemana] of Object.entries(LIQUIDA)) {
		let fecha = hoy;
		if (diaSemana !== null)
			while (new Date(fecha + AR + 12 * HORA).getUTCDay() !== diaSemana) fecha += DIA;
		const monto =
			diaSemana === null
				? cobradoCon(medio, hoy - DIA, hoy)
				: cobradoCon(medio, fecha - 9 * DIA, fecha - 2 * DIA);
		if (monto > 0) liq.set(medio, { fecha, monto });
	}

	// Los más vendidos de la semana.
	const unidades = new Map<string, number>();
	for (const p of entre(ahora - 7 * DIA, ahora))
		for (const i of p.items)
			unidades.set(i.productoId, (unidades.get(i.productoId) ?? 0) + i.cantidad);
	const masVendidos = [...unidades]
		.sort((a, b) => b[1] - a[1])
		.slice(0, 4)
		.map(([id, u]) => {
			const p = catalogo.productos.find((x) => x.id === id)!;
			return { productoId: id, nombre: p.nombre, foto: p.foto, unidades: u };
		});

	// Envíos de hoy: lo que hay que despachar, por transporte y horario de retiro.
	const aDespachar = pedidos.filter(
		(p) => (p.estado === 'Para preparar' || p.estado === 'Pagado') && p.envio !== 'retiro'
	);
	const ahoraHm = horaAr(ahora);
	let hayProximo = false;
	const retiros = (Object.keys(RETIROS) as (keyof typeof RETIROS)[])
		.map((t) => ({ t, ps: aDespachar.filter((p) => p.envio === t) }))
		.filter((x) => x.ps.length)
		.sort((a, b) => RETIROS[a.t].localeCompare(RETIROS[b.t]))
		.map(({ t, ps }) => {
			const [h, m] = RETIROS[t].split(':').map(Number);
			const proximo = !hayProximo && h + m / 60 > ahoraHm;
			if (proximo) hayProximo = true;
			const ciudades = new Set(ps.map((p) => p.ciudadEnvio));
			return {
				hora: RETIROS[t],
				transporte: t,
				paquetes: ps.length,
				ciudad: ciudades.size === 1 ? [...ciudades][0] : null,
				proximo
			};
		});

	// En vivo: más gente a la noche que a la madrugada.
	const h = horaAr(ahora);
	const curva = h < 7 ? 0.25 : h < 12 ? 0.6 : h < 19 ? 0.8 : 1;
	const personas = Math.round(8 + 22 * curva);
	const top = masVendidos.length
		? masVendidos
		: [{ nombre: catalogo.productos[0].nombre, foto: null }];
	const paginas = [
		{ nombre: top[0].nombre, personas: Math.round(personas * 0.38) },
		{ nombre: (top[1] ?? top[0]).nombre, personas: Math.round(personas * 0.26) },
		{ nombre: 'Inicio de la tienda', personas: Math.round(personas * 0.22) },
		{ nombre: 'Checkout', personas: Math.round(personas * 0.13) }
	];
	const armandose = Math.max(1, Math.round(personas / 8));
	const armandoseMonto = catalogo.productos.slice(0, armandose).reduce((s, p) => s + p.precio, 0);

	const ph = paraHacer(d, ahora);
	const online = entre(hoy, ahora).filter((p) => p.canal !== 'pos').length;
	const conEstrellas = d.resenas.map((r) => r.estrellas);

	return {
		periodo,
		desde: new Date(desde).toISOString(),
		hasta: new Date(hasta).toISOString(),
		hola: opciones.hola,
		ventas: { actual: totalActual, anterior: suma(anterior) },
		serie: { actual: acumulada(desde), anterior: acumulada(desde - corrimiento) },
		kpis: {
			pedidos: { actual: actual.length, anterior: anterior.length },
			ticket: { actual: ticket(actual), anterior: ticket(anterior) },
			conversion: {
				actual: conversion(desde, hasta),
				anterior: conversion(desde - corrimiento, hasta - corrimiento)
			},
			clientesNuevos: {
				actual: nuevos(desde, hasta),
				anterior: nuevos(desde - corrimiento, hasta - corrimiento)
			}
		},
		canales: [...porCanal]
			.sort((a, b) => b[1] - a[1])
			.map(([canal, v]) => ({ canal, parte: totalActual ? v / totalActual : 0 })),
		meta: {
			objetivo,
			actual: delMes,
			proyeccion: Math.round((delMes / transcurrido) * diasMes)
		},
		avisos: { certificadoDias: cambios.certificadoOculto ? null : DIAS_CERTIFICADO },
		ultimosPedidos: pedidos
			.filter((p) => p.creado <= ahora)
			.slice(0, 5)
			.map((p) => {
				const c = cliente.get(p.clienteId)!;
				return {
					numero: p.numero,
					cliente: c.nombre,
					foto: c.foto,
					canal: p.canal,
					estado: p.estado,
					total: p.total,
					creado: new Date(p.creado).toISOString()
				};
			}),
		paraHacer: ph,
		enVivo: { personas, paginas, carritos: armandose, carritosMonto: armandoseMonto },
		caja: {
			abierta: abierta ? new Date(apertura).toISOString() : null,
			total: suma(posHoy),
			medios: ['efectivo', 'tarjeta', 'mp', 'modo', 'transf']
				.filter((m) => medios.has(m))
				.map((m) => ({ medio: m, monto: medios.get(m)! }))
		},
		liquidaciones: [...liq]
			.sort((a, b) => a[1].fecha - b[1].fecha)
			.map(([medio, l]) => ({ medio, fecha: fechaAr(l.fecha), monto: l.monto })),
		hornero: {
			chatsHoy: online * 2 + d.conversaciones.length,
			ideas: [
				d.carritos.length > 0,
				ph.sinStock > 0,
				masVendidos.length > 0 && clientes.some((c) => c.telefono)
			].filter(Boolean).length,
			carritos: d.carritos.length,
			carritosMonto: d.carritos.reduce((s, c) => s + c.total, 0),
			costoMensaje: COSTO_MENSAJE,
			recordados: cambios.ideasHechas?.includes('carritos') ?? false,
			audiencia: clientes.filter((c) => c.telefono && !c.mayorista).length,
			ideasHechas: cambios.ideasHechas ?? []
		},
		bandeja: d.conversaciones.map((v) => {
			const c = cliente.get(v.clienteId)!;
			return {
				nombre: c.nombre,
				foto: c.foto,
				mensaje: v.mensaje,
				creado: new Date(v.creado).toISOString(),
				atiende: v.atiende,
				sinLeer: v.sinLeer
			};
		}),
		integraciones: [
			{ id: 'mp', valor: suma(entre(ahora - 7 * DIA, ahora).filter((p) => p.pago === 'mp')) },
			{ id: 'ml', valor: d.enMl.length },
			{ id: 'wa', valor: d.conversaciones.length },
			{ id: 'arca', valor: DIAS_CERTIFICADO },
			{
				id: 'andreani',
				valor: pedidos.filter((p) => p.estado === 'Enviado' && p.envio === 'andreani').length
			}
		],
		masVendidos,
		resenas: {
			promedio: conEstrellas.length
				? Math.round((conEstrellas.reduce((s, x) => s + x, 0) / conEstrellas.length) * 10) / 10
				: null,
			nuevas: d.resenas.slice(0, 3).map((r) => {
				const [n, a] = cliente.get(r.clienteId)!.nombre.split(' ');
				return {
					texto: r.texto,
					autor: `${n} ${a?.[0] ?? ''}.`,
					producto: r.producto,
					canal: r.canal,
					estrellas: r.estrellas
				};
			})
		},
		envios: { paquetes: aDespachar.length, retiros }
	};
}

/** Lo que Hornero sabe de la tienda de ejemplo (ver ContextoHornero en el contrato). */
export function contextoDemo(d: Demo, ahora: number): ContextoHornero {
	const hoy = inicioDelDia(ahora);
	const vale30 = d.pedidos.filter(
		(p) => vale(p) && p.creado >= ahora - 30 * DIA && p.creado <= ahora
	);
	const ventasPorDia = Array.from({ length: 14 }, (_, i) => {
		const d0 = hoy - (13 - i) * DIA;
		const ps = d.pedidos.filter(
			(p) => vale(p) && p.creado >= d0 && p.creado < Math.min(d0 + DIA, ahora)
		);
		return { fecha: fechaAr(d0), ventas: ps.reduce((s, p) => s + p.total, 0), pedidos: ps.length };
	});
	const vendidos = new Map<string, number>();
	for (const p of vale30)
		for (const i of p.items)
			vendidos.set(i.productoId, (vendidos.get(i.productoId) ?? 0) + i.cantidad);
	const compras = new Map<string, number[]>();
	for (const p of d.pedidos.filter(vale))
		compras.set(p.clienteId, [...(compras.get(p.clienteId) ?? []), p.creado]);
	const canales = new Map<string, number>();
	for (const p of vale30) canales.set(p.canal, (canales.get(p.canal) ?? 0) + p.total);
	const minoristas = d.clientes.filter((c) => !c.mayorista);
	return {
		ventasPorDia,
		productos: d.catalogo.productos.map((p) => ({
			nombre: p.nombre,
			categoria: p.categoria,
			precio: p.precio,
			stock: d.stock[p.id] ?? 0,
			vendidos30d: vendidos.get(p.id) ?? 0,
			enMl: d.enMl.includes(p.id)
		})),
		clientes: {
			total: minoristas.length,
			nuevos30d: minoristas.filter((c) => c.alta >= ahora - 30 * DIA).length,
			conWhatsApp: minoristas.filter((c) => c.telefono).length,
			sinVolver: [...compras.values()].filter(
				(v) => v.length > 1 && Math.max(...v) < ahora - 30 * DIA
			).length
		},
		canales30d: [...canales]
			.sort((a, b) => b[1] - a[1])
			.map(([canal, ventas]) => ({ canal, ventas }))
	};
}
