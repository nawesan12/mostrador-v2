/**
 * Actividad (docs/decisiones.md §39): quién hizo qué en la tienda. Los tipos (de qué se trata cada
 * evento) con su nombre, para el filtro y el CSV. Es la misma lista que `hoy.Categorias` en Go.
 */
import type { components } from '@mostrador/contrato/esquema';

export type CategoriaEvento = components['schemas']['CategoriaEvento'];
export type Evento = components['schemas']['Evento'];

export const CATEGORIAS: [CategoriaEvento, string][] = [
	['producto', 'Productos'],
	['pedido', 'Pedidos'],
	['venta', 'Punto de venta'],
	['promocion', 'Cupones y promociones'],
	['cobros', 'Cobros'],
	['negocio', 'Datos del negocio'],
	['envios', 'Envíos y retiro'],
	['dominio', 'Dominio'],
	['integracion', 'Integraciones'],
	['personas', 'Equipo'],
	['plan', 'Plan'],
	['datos', 'Exportaciones'],
	['sesion', 'Sesiones'],
	['mitienda', 'Mi tienda'],
	['soporte', 'Soporte de Tu Local Web']
];

const NOMBRE = new Map<string, string>(CATEGORIAS);
export const nombreCategoria = (c: string) => NOMBRE.get(c) ?? 'Otro';
export const esCategoria = (c: string | null): c is CategoriaEvento => !!c && NOMBRE.has(c);

/** Los filtros de Actividad (todos opcionales). Las fechas son YYYY-MM-DD de Argentina. */
export type FiltroActividad = {
	persona?: string;
	tipo?: CategoriaEvento;
	desde?: string;
	hasta?: string;
	cursor?: string;
};

export type PaginaActividad = {
	items: Evento[];
	siguiente: string | null;
	personas: { id: string; nombre: string }[];
};

const DIA = /^\d{4}-\d{2}-\d{2}$/;

/** Lee los filtros de la URL de la página (lo que no vale, se ignora). */
export function filtroDe(q: URLSearchParams): FiltroActividad {
	const f: FiltroActividad = {};
	const persona = q.get('persona');
	if (persona && /^[\w-]{1,64}$/.test(persona)) f.persona = persona;
	const tipo = q.get('tipo');
	if (esCategoria(tipo)) f.tipo = tipo;
	for (const k of ['desde', 'hasta'] as const) {
		const v = q.get(k);
		if (v && DIA.test(v)) f[k] = v;
	}
	const cursor = q.get('cursor');
	if (cursor && /^[\w=-]{1,200}$/.test(cursor)) f.cursor = cursor;
	return f;
}

/** Los filtros como query string (sin el «?»). */
export function consultaDe(f: FiltroActividad) {
	const q = new URLSearchParams();
	for (const [k, v] of Object.entries(f)) if (v) q.set(k, v);
	return q.toString();
}

/** Cuándo, como se lee en la lista: «hace 5 min», «hoy 14:32», «ayer 09:10», «12 oct. 18:40». */
export function cuando(iso: string, ahora = Date.now()) {
	const t = Date.parse(iso);
	const min = Math.round((ahora - t) / 60_000);
	if (min < 1) return 'recién';
	if (min < 60) return `hace ${min} min`;
	const zona = 'America/Argentina/Buenos_Aires';
	const hora = new Date(t).toLocaleTimeString('es-AR', {
		hour: '2-digit',
		minute: '2-digit',
		hourCycle: 'h23',
		timeZone: zona
	});
	const dia = (x: number) => new Date(x - 3 * 3_600_000).toISOString().slice(0, 10);
	if (dia(t) === dia(ahora)) return `hoy ${hora}`;
	if (dia(t) === dia(ahora - 86_400_000)) return `ayer ${hora}`;
	const fecha = new Date(t).toLocaleDateString('es-AR', {
		day: 'numeric',
		month: 'short',
		year: dia(t).slice(0, 4) === dia(ahora).slice(0, 4) ? undefined : 'numeric',
		timeZone: zona
	});
	return `${fecha} ${hora}`;
}
