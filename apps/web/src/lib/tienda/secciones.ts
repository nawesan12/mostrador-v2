import type { TiendaPublica } from './tipos.ts';

type Seccion = TiendaPublica['secciones'][number];

/** El orden con el que arranca toda tienda (mitienda.Inicial en Go, DISENO_INICIAL en la demo). */
const INICIAL: readonly Seccion[] = ['hero', 'cats', 'prods', 'revs', 'ig', 'wa'];

/**
 * Las secciones de la portada en el orden en que se dibujan. Si la tienda nunca las reordenó (siguen
 * en el orden inicial), va el orden del diseño de la plantilla (Malbec pone los productos antes que
 * las categorías); si las reordenó en Mi tienda, el suyo.
 */
export function enOrden(visibles: readonly Seccion[], plantilla: readonly Seccion[]): Seccion[] {
	const pos = visibles.map((s) => INICIAL.indexOf(s));
	const inicial = pos.every((p, i) => i === 0 || p > pos[i - 1]);
	return inicial ? plantilla.filter((s) => visibles.includes(s)) : [...visibles];
}
