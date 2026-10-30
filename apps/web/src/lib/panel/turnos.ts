/** Turnos en el panel: nombres y colores de cada estado (los de los chips de pedidos). */
import { ESTADOS } from '#lib/marcas.ts';
import { linkWhatsapp } from '#lib/whatsapp.ts';
import type { components } from '@mostrador/contrato/esquema';

export type Turno = components['schemas']['Turno'];
export type EstadoTurno = components['schemas']['EstadoTurno'];

export const ESTADO_TURNO: Record<EstadoTurno, { t: string; c: readonly [string, string] }> = {
	reservado: { t: 'Espera la seña', c: ESTADOS['Pago pendiente'] },
	confirmado: { t: 'Confirmado', c: ESTADOS.Pagado },
	atendido: { t: 'Atendido', c: ESTADOS.Entregado },
	ausente: { t: 'No vino', c: ESTADOS['Sin pagar'] },
	cancelado: { t: 'Cancelado', c: ESTADOS.Cancelado }
};

/** Lo que se dice al cambiar el estado (toast). */
export const HECHO: Record<EstadoTurno, string> = {
	reservado: 'Turno de nuevo esperando la seña',
	confirmado: 'Turno confirmado',
	atendido: 'Marcado como atendido',
	ausente: 'Marcado como ausente',
	cancelado: 'Turno cancelado'
};

/** "wa.me" para el teléfono del cliente (549 + característica + número). */
export const waDe = (tel: string) => linkWhatsapp(tel);
