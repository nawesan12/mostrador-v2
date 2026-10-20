/**
 * Resumen del armazón del panel en la demo: insignias y notificaciones, sacadas de los datos de la
 * tienda de ejemplo (demo.ts), así cada demo habla de sus propios productos y clientes.
 */
import type { components } from '@mostrador/contrato/esquema';
import type { Demo } from './demo.ts';

type ResumenPanel = components['schemas']['ResumenPanel'];
type Notificacion = components['schemas']['Notificacion'];

const MIN = 60_000;
const fmt = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');

export function resumenDemo(
	d: Demo,
	base: number,
	leidasHasta: number,
	ahora: number
): ResumenPanel {
	const { catalogo, clientes, pedidos } = d;
	const cliente = (id: string) => clientes.find((c) => c.id === id)!;
	const primerNombre = (id: string) => cliente(id).nombre.split(' ')[0];
	const ultimoWa = pedidos.find((p) => p.canal === 'wa') ?? pedidos[0];
	const ultimoMl = pedidos.find((p) => p.canal === 'ml') ?? pedidos[0];
	// Lo que Mercado Pago liquidó: lo cobrado ayer por MP.
	const liquidado = pedidos
		.filter(
			(p) => p.pago === 'mp' && ahora - p.creado > 1440 * MIN && ahora - p.creado < 2880 * MIN
		)
		.reduce((s, p) => s + p.total, 0);
	const masVendido = catalogo.productos[0];
	const conStockBajo = catalogo.productos[2] ?? masVendido;
	const n = (
		id: string,
		tipo: string,
		titulo: string,
		detalle: string,
		haceMin: number,
		grupo: 'ventas' | 'sistema'
	): Notificacion => {
		const creada = base - haceMin * MIN;
		return {
			id,
			tipo,
			titulo,
			detalle,
			creada: new Date(creada).toISOString(),
			leida: creada <= leidasHasta,
			grupo
		};
	};
	return {
		insignias: {
			pedidos: pedidos.filter((p) => p.estado === 'Para preparar').length,
			carritos: 12,
			whatsapp: 3,
			paraHacer: 4
		},
		notificaciones: [
			n(
				'wa',
				'wa',
				`Nuevo mensaje de ${primerNombre(ultimoWa.clienteId)}`,
				`“Hola, ¿tienen stock de ${ultimoWa.items[0].nombre}?” · Hornero ya respondió`,
				4,
				'ventas'
			),
			n(
				'mp',
				'mp',
				'Liquidación disponible',
				`${fmt(liquidado || 1284500)} de Mercado Pago ya están en tu cuenta`,
				32,
				'ventas'
			),
			n(
				'ml',
				'ml',
				'Pregunta en Mercado Libre',
				`“¿Hacen envíos a ${cliente(ultimoMl.clienteId).ciudad}?” en ${ultimoMl.items[0].nombre}`,
				60,
				'ventas'
			),
			n('stock', 'warn', 'Stock bajo', `${conStockBajo.nombre}: quedan 3 unidades`, 120, 'sistema'),
			n(
				'arca',
				'arca',
				'Certificado por vencer',
				'Tu certificado de ARCA vence en 5 días',
				1440,
				'sistema'
			),
			n(
				'resena',
				'star',
				'Nueva reseña de 5 estrellas',
				'“Llegó rapidísimo y súper bien embalado”',
				1500,
				'ventas'
			)
		],
		canales: ['wa', 'ml', 'mp', 'ig', 'tg'],
		prueba: { plan: 'Pro', diasRestantes: 9, diasTotales: 14 }
	};
}
