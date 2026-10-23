/**
 * Reseñas y preguntas que deja quien compra en la tienda online (/t/[tienda]/resena y /pregunta).
 * Mismas reglas que Go (ventas/resenas.go): se validan acá con valibot y se guardan en el sandbox
 * de la demo o en la API. En la vitrina de una plantilla no se guarda nada.
 */
import { env } from 'cloudflare:workers';
import { json, type RequestEvent } from '@sveltejs/kit';
import * as v from 'valibot';
import { ErrorApi } from '../bff.ts';
import { usaSandbox, vidrieraDe } from '../datos/index.ts';
import { golpe } from '../limite.ts';
import { COOKIE, dbSandbox } from '../sandbox/sandbox.ts';
import { NuevaPregunta, NuevaResena } from './esquemas.ts';
import { resolverTienda, type TiendaResuelta } from './publica.ts';

type Evento = RequestEvent<{ tienda: string }>;

const error = (estado: number, mensaje: string) => json({ error: mensaje }, { status: estado });
const listo = () => json({ ok: true }, { status: 201 });
const MUCHAS = 'Mandaste muchas seguidas. Probá de nuevo en un rato.';

/** Diez envíos por hora por IP y por tipo, como Go (en la API, el tope lo pone Go). */
async function dentroDelTope(event: Evento, tipo: string) {
	let ip = 'sin-ip';
	try {
		ip = event.request.headers.get('cf-connecting-ip') ?? event.getClientAddress();
	} catch {
		// Sin dirección (tests): cuenta como una sola.
	}
	return golpe(dbSandbox(env.SANDBOX), `tienda:${tipo}:${ip}`, 10, 3600);
}

/** Lo común: la tienda (no la vitrina), el cuerpo validado y el tope. */
async function preparar<E extends v.GenericSchema>(
	event: Evento,
	tipo: 'resena' | 'pregunta',
	esquema: E
): Promise<Response | { r: TiendaResuelta; datos: v.InferOutput<E> }> {
	const r = await resolverTienda(event.params.tienda, event.cookies.get(COOKIE));
	if (!r) return error(404, 'Esta tienda no existe.');
	if (r.tienda.vitrina)
		return error(
			409,
			`Es una tienda de ejemplo: ${tipo === 'resena' ? 'las reseñas' : 'las preguntas'} no se guardan.`
		);
	const b = v.safeParse(esquema, await event.request.json().catch(() => null));
	if (!b.success) return error(400, b.issues[0].message);
	if (usaSandbox() && !(await dentroDelTope(event, tipo))) return error(429, MUCHAS);
	return { r, datos: b.output };
}

/** Errores de Go, en el mismo formato que los del sandbox. */
function deLaApi(e: unknown) {
	if (!(e instanceof ErrorApi)) throw e;
	if (e.estado === 409) return error(409, e.message);
	if (e.estado === 429) return error(429, MUCHAS);
	return error(e.estado >= 500 ? 503 : 400, e.message);
}

export async function dejarResena(event: Evento) {
	const p = await preparar(event, 'resena', NuevaResena);
	if (p instanceof Response) return p;
	const { r, datos } = p;
	if (r.datos) {
		const res = await r.datos.dejarResena(datos);
		if (res === 'ya-reseno') return error(409, 'Ya dejaste tu reseña de esa compra. ¡Gracias!');
		if (res === 'invalido') return error(400, 'Revisá la reseña.');
		return listo();
	}
	try {
		await vidrieraDe(event).resena(r.tienda.slug, datos);
		return listo();
	} catch (e) {
		return deLaApi(e);
	}
}

export async function preguntar(event: Evento) {
	const p = await preparar(event, 'pregunta', NuevaPregunta);
	if (p instanceof Response) return p;
	const { r, datos } = p;
	if (r.datos) {
		if ((await r.datos.preguntar(datos)) === 'invalido')
			return error(400, 'Ese producto ya no está publicado.');
		return listo();
	}
	try {
		await vidrieraDe(event).pregunta(r.tienda.slug, datos);
		return listo();
	} catch (e) {
		return deLaApi(e);
	}
}
