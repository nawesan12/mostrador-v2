/**
 * Lo que comparten el armador de cupones (panel) y quien los guarda (sandbox / API): el texto del
 * descuento, cuándo vence y la estimación de Tincho (design/PanelCupones.dc.html).
 */
import type { components } from '@mostrador/contrato/esquema';

type Cupon = components['schemas']['Cupon'];
export type CuponNuevo = components['schemas']['CuponNuevo'];

const DIA = 86_400_000;
const pesos = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');
const AR = 3 * 3_600_000;
const fecha = (t: number) =>
	new Date(t).toLocaleDateString('es-AR', {
		day: 'numeric',
		month: 'long',
		timeZone: 'America/Argentina/Buenos_Aires'
	});

export function encabezado(tipo: Cupon['tipo'], valor: number) {
	if (tipo === 'pct') return `${valor} % off`;
	if (tipo === 'fijo') return `${pesos(valor)} off`;
	if (tipo === 'envio') return 'Envío gratis';
	return 'Llevá 2, pagá 1';
}

/** El número grande del ticket. */
export function grande(c: Pick<Cupon, 'tipo' | 'valor'>) {
	if (c.tipo === 'pct') return `${c.valor} %`;
	if (c.tipo === 'fijo') return `$ ${(c.valor / 1000).toLocaleString('es-AR')}k`;
	return c.tipo === 'envio' ? 'Envío' : '2x1';
}

export function alcanceTexto(
	alcance: CuponNuevo['alcance'],
	categoria: string | null,
	productos: string[]
) {
	if (alcance === 'todo') return 'en toda la tienda';
	if (alcance === 'categoria') return `en ${(categoria ?? 'una categoría').toLowerCase()}`;
	return productos.length ? `en ${productos.join(', ')}` : 'en los productos que elijas';
}

/** El domingo que viene a las 23:59 de Argentina. */
function domingo(ahora: number) {
	const ar = new Date(ahora - AR);
	const dias = (7 - ar.getUTCDay()) % 7;
	return Date.UTC(ar.getUTCFullYear(), ar.getUTCMonth(), ar.getUTCDate() + dias, 26, 59);
}

/** El próximo Hot Sale (del 11 al 13 de mayo). */
function hotSale(ahora: number) {
	const y = new Date(ahora - AR).getUTCFullYear();
	const este = Date.UTC(y, 4, 11, 3);
	const desde = este > ahora ? este : Date.UTC(y + 1, 4, 11, 3);
	return { desde, hasta: desde + 3 * DIA - 60_000 };
}

/** El lunes que viene a las 9 de Argentina (para "Programar para más adelante"). */
export function lunes9(ahora: number) {
	const ar = new Date(ahora - AR);
	const dias = (8 - ar.getUTCDay()) % 7 || 7;
	return Date.UTC(ar.getUTCFullYear(), ar.getUTCMonth(), ar.getUTCDate() + dias, 12);
}

/** Cuándo arranca y vence según la vigencia elegida. */
export function vigenciaDe(
	v: CuponNuevo['vigencia'],
	ahora: number
): { empieza: number | null; termina: number | null } {
	if (v === '7') return { empieza: null, termina: ahora + 7 * DIA };
	if (v === '30') return { empieza: null, termina: ahora + 30 * DIA };
	if (v === 'finde') return { empieza: null, termina: domingo(ahora) };
	if (v === 'hotsale') {
		const h = hotSale(ahora);
		return { empieza: h.desde, termina: h.hasta };
	}
	return { empieza: null, termina: null };
}

export function vigenciaTexto(v: CuponNuevo['vigencia'], ahora: number) {
	if (v === 'hotsale') {
		const h = hotSale(ahora);
		return `Del ${fecha(h.desde).split(' ')[0]} al ${fecha(h.hasta)}`;
	}
	return {
		'7': 'Vence en 7 días',
		finde: 'Vence el domingo',
		'30': 'Vence en 30 días',
		sin: 'Sin vencimiento'
	}[v];
}

/** La línea de estado de una tarjeta: "Vence en 12 días", "Arranca el 11 de mayo"… */
export function cuando(c: Pick<Cupon, 'estado' | 'empieza' | 'termina'>, ahora: number) {
	if (c.estado === 'pausado') return 'Pausado por vos';
	if (c.estado === 'programado')
		return c.empieza ? `Arranca el ${fecha(Date.parse(c.empieza))}` : 'Programado';
	if (c.estado === 'terminado')
		return c.termina ? `Terminó el ${fecha(Date.parse(c.termina))}` : 'Terminado';
	if (!c.termina) return 'Sin vencimiento';
	const dias = Math.ceil((Date.parse(c.termina) - ahora) / DIA);
	if (dias <= 1) return 'Vence hoy';
	if (dias <= 7 && new Date(Date.parse(c.termina) - AR).getUTCDay() === 0)
		return 'Vence el domingo';
	return `Vence en ${dias} días`;
}

/** Lo que calcula Tincho en el armador (la fórmula del handoff). */
export function estimar(n: Pick<CuponNuevo, 'tipo' | 'valor' | 'quien' | 'minimo' | 'limite'>) {
	const v = n.valor;
	const pctEq =
		n.tipo === 'pct' ? v : n.tipo === 'fijo' ? Math.min(60, v / 400) : n.tipo === 'envio' ? 8 : 33;
	const alcance = { todos: 1, nuevos: 0.55, vip: 0.25, riesgo: 0.4 }[n.quien];
	const usos = Math.round(
		(n.limite || 260) * Math.min(1, pctEq / 20 + 0.3) * alcance * (n.minimo > 50_000 ? 0.6 : 1)
	);
	const ticket = Math.max(n.minimo, 38_000);
	const ventas = usos * ticket;
	const costo =
		n.tipo === 'fijo' ? usos * v : n.tipo === 'envio' ? usos * 5200 : (ventas * pctEq) / 100;
	const margen = Math.round(42 - pctEq * 0.95 + n.minimo / 10_000);
	return { ventas, costo, margen };
}

/** La descripción que queda guardada. */
export function descripcionDe(n: CuponNuevo, categoria: string | null, productos: string[]) {
	const a = alcanceTexto(n.alcance, categoria, productos);
	return `${encabezado(n.tipo, n.valor)} ${a}${n.minimo ? ` · mín. ${pesos(n.minimo)}` : ''}`;
}
