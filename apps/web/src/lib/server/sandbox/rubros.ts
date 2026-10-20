import RUBROS from '@mostrador/contrato/rubros.json';
import type { Rubro } from '../datos/tipos.ts';

export { RUBROS };

export const rubro = (id: string): Rubro | undefined => RUBROS.find((r) => r.id === id);

/** "Café Bruma" → "cafe-bruma". */
export function slugDe(nombre: string): string {
	return nombre
		.normalize('NFD')
		.replace(/[̀-ͯ]/g, '')
		.toLowerCase()
		.replace(/&/g, 'y')
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-+|-+$/g, '')
		.slice(0, 48);
}
