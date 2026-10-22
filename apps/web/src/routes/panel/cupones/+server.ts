import { error, json } from '@sveltejs/kit';
import * as v from 'valibot';
import { datosDe } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

const Nuevo = v.object({
	codigo: v.pipe(
		v.string(),
		v.regex(/^[A-Z0-9-]{3,20}$/, 'El código va en mayúsculas, de 3 a 20 letras o números.')
	),
	tipo: v.picklist(['pct', 'fijo', 'envio', 'regalo']),
	valor: v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(100_000)),
	alcance: v.picklist(['todo', 'categoria', 'productos']),
	productos: v.optional(v.pipe(v.array(v.pipe(v.string(), v.maxLength(80))), v.maxLength(50))),
	quien: v.picklist(['todos', 'nuevos', 'vip', 'riesgo']),
	minimo: v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(10_000_000)),
	limite: v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(100_000)),
	porCliente: v.picklist(['1', '3', 'libre']),
	vigencia: v.picklist(['7', 'finde', '30', 'hotsale', 'sin']),
	canales: v.pipe(
		v.array(v.picklist(['web', 'wa', 'ig', 'pos'])),
		v.minLength(1, 'Elegí dónde vale.')
	),
	opciones: v.object({
		hornero: v.boolean(),
		combinable: v.boolean(),
		unicos: v.boolean(),
		link: v.boolean()
	}),
	programar: v.optional(v.boolean(), false)
});

/** Un cupón nuevo. */
export const POST: RequestHandler = async ({ locals, request }) => {
	const r = v.safeParse(Nuevo, await request.json().catch(() => null));
	if (!r.success) error(400, r.issues[0].message);
	const n = r.output;
	if (n.tipo === 'pct' && (n.valor < 1 || n.valor > 90)) error(400, 'El porcentaje va de 1 a 90.');
	if (n.tipo === 'fijo' && n.valor < 100) error(400, 'Poné el monto del descuento.');
	const c = await datosDe(locals).crearCupon(n);
	if (c === 'repetido') error(409, `Ya tenés un cupón ${n.codigo}.`);
	return json(c, { status: 201 });
};
