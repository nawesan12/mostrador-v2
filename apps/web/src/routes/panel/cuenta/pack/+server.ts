import { json } from '@sveltejs/kit';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/**
 * El pack extra de Tincho (+2.000 respuestas este mes, $ 4.900): devuelve la URL de Mercado Pago; el
 * pack se suma cuando el pago se aprueba. Un pack pago no se devuelve.
 */
export const POST: RequestHandler = async ({ locals }) => json(await datosDe(locals).comprarPack());
