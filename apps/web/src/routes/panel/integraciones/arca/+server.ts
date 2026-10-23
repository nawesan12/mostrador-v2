import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** ARCA: los mismos límites que el contrato (ArcaAlta, certificado y ajustes). */
const Alta = v.object({
	cuit: v.pipe(v.string(), v.minLength(11), v.maxLength(13)),
	razonSocial: v.pipe(v.string(), v.trim(), v.minLength(1), v.maxLength(64)),
	condicion: v.picklist(['monotributo', 'responsable_inscripto']),
	puntoVenta: v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(99998)),
	ambiente: v.picklist(['homologacion', 'produccion'])
});
const Certificado = v.object({
	certificado: v.pipe(v.string(), v.minLength(1), v.maxLength(16384))
});
const Ajustes = v.object({
	puntoVenta: v.optional(v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(99998))),
	facturarAlCobrar: v.optional(v.boolean())
});

const cuerpo = async <T extends v.GenericSchema>(s: T, r: Request) => {
	const b = v.safeParse(s, await r.json().catch(() => null));
	if (!b.success) error(400, 'Revisá los datos.');
	return b.output as v.InferOutput<T>;
};

/** Genera el pedido de certificado (en la demo, valida y lo arma el navegador). */
export const POST: RequestHandler = async ({ locals, request }) => {
	const r = await datosDe(locals).arcaPedido(await cuerpo(Alta, request));
	if ('error' in r) error(400, r.error);
	return json(r);
};

/** Guarda el certificado que devolvió ARCA. */
export const PUT: RequestHandler = async ({ locals, request }) => {
	const r = await datosDe(locals).arcaCertificado((await cuerpo(Certificado, request)).certificado);
	if ('error' in r) error(400, r.error);
	return json(r);
};

/** Punto de venta y facturar al cobrar. */
export const PATCH: RequestHandler = async ({ locals, request }) => {
	const r = await datosDe(locals).arcaAjustes(await cuerpo(Ajustes, request));
	if (r) error(400, r.error);
	return new Response(null, { status: 204 });
};
