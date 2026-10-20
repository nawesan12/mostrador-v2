/**
 * La fuente de datos de un request: el sandbox de la demo (DATA_SOURCE=sandbox) o la API Go
 * (DATA_SOURCE=api). Las páginas piden `datosDe(locals)` y no saben cuál es.
 */
import { env } from 'cloudflare:workers';
import { redirect } from '@sveltejs/kit';
import { DatosApi } from './api.ts';
import { DatosSandbox, dbSandbox } from '../sandbox/sandbox.ts';
import type { Datos } from './tipos.ts';

export const usaSandbox = () => String(env.DATA_SOURCE) !== 'api';

export function datosDe(locals: App.Locals): Datos {
	if (!usaSandbox()) return new DatosApi(String(env.API_URL), undefined);
	// Sin sandbox (cookie vencida o primera visita) se elige rubro de nuevo.
	if (!locals.sandbox) redirect(303, '/demo');
	return new DatosSandbox(dbSandbox(env.SANDBOX), locals.sandbox);
}
