/**
 * Envíos y retiro (Configuración): retiro en el local y envío propio en moto. Los topes y los
 * mensajes son los de Go (configuracion/entregas.go y la migración 00048).
 */
import * as v from 'valibot';
import type { components } from '@mostrador/contrato/esquema';

export type Entregas = components['schemas']['Entregas'];

export const TOPES_ENTREGAS = {
	retiroDireccion: 160,
	retiroHorario: 80,
	motoZona: 160,
	motoDemora: 60,
	motoCosto: 1_000_000
} as const;

const txt = (max: number) => v.pipe(v.string(), v.trim(), v.maxLength(max, 'Es muy largo.'));

export const EsquemaEntregas = v.object({
	retiro: v.boolean(),
	retiroDireccion: txt(TOPES_ENTREGAS.retiroDireccion),
	retiroHorario: txt(TOPES_ENTREGAS.retiroHorario),
	moto: v.boolean(),
	motoCosto: v.pipe(
		v.number('Poné el costo de la moto en pesos.'),
		v.integer('El costo va en pesos, sin centavos.'),
		v.minValue(0, 'El costo no puede ser negativo.'),
		v.maxValue(TOPES_ENTREGAS.motoCosto, 'Revisá el costo de la moto.')
	),
	motoZona: txt(TOPES_ENTREGAS.motoZona),
	motoDemora: txt(TOPES_ENTREGAS.motoDemora)
});

export const SIN_ENTREGA = 'Dejá al menos una forma de entrega: retiro, moto o Correo Argentino.';

/**
 * Lo que no deja guardar (o null): todo apagado sin Correo conectado, la moto sin zona o el retiro
 * sin ninguna dirección (ni la propia ni la del negocio).
 */
export function problemaEntregas(
	e: Entregas,
	ctx: { correo: boolean; direccionNegocio: string }
): string | null {
	if (e.moto && e.motoZona.trim().length < 3)
		return 'Poné hasta dónde llega la moto (barrios o localidades).';
	if (!e.retiro && !e.moto && !ctx.correo) return SIN_ENTREGA;
	if (e.retiro && !e.retiroDireccion.trim() && !ctx.direccionNegocio.trim())
		return 'Poné la dirección donde retiran.';
	return null;
}
