/**
 * La fuente de datos de un request: el sandbox de la demo (DATA_SOURCE=sandbox) o la API Go
 * (DATA_SOURCE=api). Las páginas piden `datosDe(locals)` y no saben cuál es.
 */
import { env } from 'cloudflare:workers';
import { redirect, type RequestEvent } from '@sveltejs/kit';
import { CuentasApi } from '../auth/cuentas.ts';
import { DatosApi } from './api.ts';
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

/** Cuentas en la API Go, con la IP y el navegador del visitante. */
export function cuentasDe(event: Pick<RequestEvent, 'request' | 'getClientAddress'>) {
	let ip = '';
	try {
		ip = event.request.headers.get('cf-connecting-ip') ?? event.getClientAddress();
	} catch {
		// Sin dirección (tests): Go usa la de la conexión.
	}
	return new CuentasApi(String(env.API_URL), String(env.CLAVE_BFF ?? ''), {
		ip,
		agente: event.request.headers.get('user-agent') ?? ''
	});
}
