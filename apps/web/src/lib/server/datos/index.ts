/**
 * La fuente de datos de un request: el sandbox de la demo (DATA_SOURCE=sandbox) o la API Go
 * (DATA_SOURCE=api). Las páginas piden `datosDe(locals)` y no saben cuál es.
 */
import { env } from 'cloudflare:workers';
import { redirect, type RequestEvent } from '@sveltejs/kit';
import { CuentasApi } from '../auth/cuentas.ts';
import { DatosApi } from './api.ts';
import { FormulariosApi } from '../formularios.ts';
import { DatosSandbox, dbSandbox } from '../sandbox/sandbox.ts';
import type { Datos } from './tipos.ts';

export const usaSandbox = () => String(env.DATA_SOURCE) !== 'api';

export function datosDe(locals: App.Locals): Datos {
	if (!usaSandbox()) {
		// Sin sesión, el panel manda a ingresar.
		if (!locals.token) redirect(303, '/ingresar');
		return new DatosApi(String(env.API_URL), locals.token);
	}
	// Sin sandbox (cookie vencida o primera visita) se elige rubro de nuevo.
	if (!locals.sandbox) redirect(303, '/demo');
	return new DatosSandbox(dbSandbox(env.SANDBOX), locals.sandbox);
}

type Evento = Pick<RequestEvent, 'request' | 'getClientAddress'>;

/** IP y navegador del visitante, para que Go aplique los límites a la persona y no al BFF. */
export function clienteDe(event: Evento) {
	let ip = '';
	try {
		ip = event.request.headers.get('cf-connecting-ip') ?? event.getClientAddress();
	} catch {
		// Sin dirección (tests): Go usa la de la conexión.
	}
	return { ip, agente: event.request.headers.get('user-agent') ?? '' };
}

/** Cuentas en la API Go, con la IP y el navegador del visitante. */
export const cuentasDe = (event: Evento) =>
	new CuentasApi(String(env.API_URL), String(env.CLAVE_BFF ?? ''), clienteDe(event));

/** Formularios públicos en la API Go (arrepentimiento, contacto, partners, novedades). */
export const formulariosDe = (event: Evento) =>
	new FormulariosApi(String(env.API_URL), String(env.CLAVE_BFF ?? ''), clienteDe(event));
