import RUBROS from '@mostrador/contrato/rubros.json';
import type { Rubro } from '../datos/tipos.ts';

export { RUBROS };

export const rubro = (id: string): Rubro | undefined => RUBROS.find((r) => r.id === id);

export { slugDe } from '../../slug.ts';
