// Los planes de la landing y la comparativa: salen del catálogo de lib/panel/planes.ts (el mismo que
// el panel y la API), no del handoff (scripts/datos-landing.ts ya no los genera).
import { DESTACADO, PLANES } from '#lib/panel/planes.ts';

/**
 * Los planes de la landing: nombre, lema, precio por mes, lo que incluye, el botón y si es el
 * destacado. Precios y textos salen de lib/panel/planes.ts (el mismo catálogo que el panel y la API).
 */
export const LPLANS = PLANES.map(
	(p) =>
		[
			p.nombre,
			p.lema,
			p.precio,
			p.incluye,
			p.precio ? 'Probar 14 días' : 'Crear tienda gratis',
			p.id === DESTACADO
		] as const
);

const precio = (id: string) => {
	const p = PLANES.find((x) => x.id === id)!;
	return `${p.nombre} · $ ${p.precio.toLocaleString('es-AR')}`;
};

export const CMP = [
	['Plan para empezar', 'Gratis', 'Gratis (Inicial)', 1, 1],
	['Plan para vender todos los días', precio('emprende'), 'Esencial · $ 27.999', 1, 1],
	['Plan para crecer', precio('pro'), 'Impulso · $ 79.999', 1, 1],
	['Plan sin topes', precio('negocio'), 'Escala · $ 244.999', 1, 1],
	['Comisión por venta', '0 % en todos', 'Según el plan', 1, 2],
	['Usuarios con permisos', 'Desde Crece', 'Desde Escala', 1, 2],
	['Migración asistida', 'Gratis en todos', 'En Evolución (a consultar)', 1, 2],
	['Asistente de IA en tu panel', 'En todos (según plan)', 'No figura en el plan', 1, 0],
	['Factura de ARCA automática', 'Desde Emprende', 'No figura en el plan', 1, 0],
	['Punto de venta en el local', 'Desde Crece', 'No figura en el plan', 1, 0],
	['Copia diaria de tus datos y aviso de cada ingreso', 'En todos', 'No figura en el plan', 1, 0]
] as const;
