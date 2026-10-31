/**
 * La campanita del panel (docs/decisiones.md §45): reglas puras, aparte para probarlas sin navegador.
 */

/** Sólo se abren páginas del panel: lo que venga de afuera no se sigue. */
export function esRutaPanel(ruta: string): boolean {
	return (
		/^\/panel(\/[a-z0-9/_?=&.-]*)?$/.test(ruta) && !ruta.includes('//') && !ruta.includes('..')
	);
}

/** El número de la insignia: hasta 9, después «9+». Vacío sin nada sin leer. */
export function insignia(sinLeer: number): string {
	if (!Number.isFinite(sinLeer) || sinLeer <= 0) return '';
	return sinLeer > 9 ? '9+' : String(Math.floor(sinLeer));
}
