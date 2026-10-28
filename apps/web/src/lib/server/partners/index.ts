/**
 * La fuente del panel de partners de un request: la demo (sandbox) o la API Go. Como `datosDe`,
 * las páginas no saben cuál es.
 */
import { env } from 'cloudflare:workers';
import { redirect, type RequestEvent } from '@sveltejs/kit';
import { usaSandbox } from '../datos/index.ts';
import { dbSandbox } from '../sandbox/sandbox.ts';
import { DemoPartners } from './demo.ts';
import { PartnersApi, type FuentePartners } from './fuente.ts';

type Evento = Pick<RequestEvent, 'locals' | 'request' | 'getClientAddress'>;

function clienteDe(event: Evento) {
	let ip = '';
	try {
		ip = event.request.headers.get('cf-connecting-ip') ?? event.getClientAddress();
	} catch {
		// Sin dirección (tests): Go usa la de la conexión.
	}
	return { ip, agente: event.request.headers.get('user-agent') ?? '' };
}

export function partnersDe(event: Evento): FuentePartners {
	if (!usaSandbox()) {
		if (!event.locals.token) redirect(303, '/ingresar');
		return new PartnersApi(
			String(env.API_URL),
			String(env.CLAVE_BFF ?? ''),
			clienteDe(event),
			event.locals.token
		);
	}
	// Sin sandbox, al programa: el botón del panel de ejemplo arma uno.
	if (!event.locals.sandbox) redirect(303, '/partners#ejemplo');
	return new DemoPartners(dbSandbox(env.SANDBOX), event.locals.sandbox);
}

/** Cuenta el clic de un link de partner (sólo con la API: la demo no tiene a quién sumárselo). */
export function contarClic(event: Evento, ref: string) {
	if (usaSandbox()) return Promise.resolve();
	return new PartnersApi(String(env.API_URL), String(env.CLAVE_BFF ?? ''), clienteDe(event))
		.clic(ref)
		.catch(() => undefined);
}
