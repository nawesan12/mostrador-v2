import { error, json, type RequestEvent } from '@sveltejs/kit';
import * as v from 'valibot';
import { ErrorApi } from '#lib/server/bff.ts';
import { correoDe } from '#lib/server/correo-argentino.ts';
import { usaSandbox } from '#lib/server/datos/index.ts';
import type { RequestHandler } from './$types';

/** Correo Argentino: los mismos límites que Go (correoargentino.Origen, Paquete y Alta). */
const Origen = v.object({
	cp: v.pipe(v.string(), v.trim(), v.regex(/^\d{4}$/, 'El código postal tiene 4 números.')),
	calle: v.pipe(v.string(), v.trim(), v.minLength(1, 'Poné la calle.'), v.maxLength(80)),
	altura: v.pipe(v.string(), v.trim(), v.minLength(1, 'Poné la altura.'), v.maxLength(10)),
	piso: v.optional(v.pipe(v.string(), v.trim(), v.maxLength(3)), ''),
	depto: v.optional(v.pipe(v.string(), v.trim(), v.maxLength(3)), ''),
	localidad: v.pipe(v.string(), v.trim(), v.minLength(1, 'Poné la localidad.'), v.maxLength(80)),
	provincia: v.pipe(v.string(), v.regex(/^[A-Z]$/, 'Elegí la provincia.'))
});
const lado = v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(150));
const Paquete = v.object({
	peso: v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(25000)),
	alto: lado,
	ancho: lado,
	largo: lado
});
const Alta = v.object({
	email: v.pipe(v.string(), v.trim(), v.email('Revisá el email.'), v.maxLength(120)),
	clave: v.pipe(v.string(), v.minLength(1, 'Poné la contraseña.'), v.maxLength(200)),
	api: v.optional(
		v.nullable(
			v.object({
				usuario: v.pipe(v.string(), v.trim(), v.maxLength(120)),
				clave: v.pipe(v.string(), v.maxLength(200))
			})
		),
		null
	),
	origen: Origen,
	paquete: Paquete
});
const Ajustes = v.object({ origen: Origen, paquete: Paquete });

const DEMO =
	'Es una tienda de ejemplo: Correo Argentino se conecta con tu cuenta de MiCorreo real.';

async function cuerpo<T extends v.GenericSchema>(s: T, r: Request) {
	const b = v.safeParse(s, await r.json().catch(() => null));
	if (!b.success) error(400, b.issues[0]?.message ?? 'Revisá los datos.');
	return b.output as v.InferOutput<T>;
}

/** Un error de Go con su mensaje (datos, cuenta, rechazo de Correo); lo demás, que reintente. */
function falla(e: unknown): never {
	if (e instanceof ErrorApi) error(e.estado, e.message);
	console.error('correo', e instanceof Error ? e.message : e);
	error(503, 'No pudimos hablar con Correo Argentino. Probá de nuevo en un rato.');
}

function sesion(event: RequestEvent) {
	if (usaSandbox()) error(409, DEMO);
	if (!event.locals.token) error(401, 'Tu sesión venció. Volvé a ingresar.');
	return event.locals.token;
}

/** Conectar: valida la cuenta de MiCorreo y guarda el origen y el paquete. */
export const POST: RequestHandler = async (event) => {
	const token = sesion(event);
	const a = await cuerpo(Alta, event.request);
	const api = a.api && (a.api.usuario || a.api.clave) ? a.api : null;
	return json(
		await correoDe(event)
			.conectar(token, { ...a, api })
			.catch(falla)
	);
};

/** Cambiar el origen y el paquete típico. */
export const PUT: RequestHandler = async (event) => {
	const token = sesion(event);
	const a = await cuerpo(Ajustes, event.request);
	return json(await correoDe(event).ajustar(token, a.origen, a.paquete).catch(falla));
};

/** Desconectar. */
export const DELETE: RequestHandler = async (event) => {
	const token = sesion(event);
	await correoDe(event).desconectar(token).catch(falla);
	return new Response(null, { status: 204 });
};
