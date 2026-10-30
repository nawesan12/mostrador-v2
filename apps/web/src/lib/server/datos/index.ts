/**
 * La fuente de datos de un request: el sandbox de la demo (DATA_SOURCE=sandbox) o la API Go
 * (DATA_SOURCE=api). Las páginas piden `datosDe(locals)` y no saben cuál es.
 */
import { env } from 'cloudflare:workers';
import { redirect, type RequestEvent } from '@sveltejs/kit';
import { CuentasApi } from '../auth/cuentas.ts';
import { decrypt, encrypt } from '../cifrado.ts';
import { fuentes } from '../importacion/fuentes.ts';
import { localDe } from '../importar/tiendanube.ts';
import { DatosApi } from './api.ts';
import { FormulariosApi } from '../formularios.ts';
import { VidrieraApi } from '../tienda/vidriera.ts';
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
	if (!locals.sandbox) redirect(303, '/prueba');
	return new DatosSandbox(dbSandbox(env.SANDBOX), locals.sandbox, extrasSandbox());
}

/**
 * Lo que la demo toma del Worker: con qué cifrar credenciales (ENCRYPTION_KEY) y la app de socio de
 * Tiendanube (TIENDANUBE_CLIENT_ID y TIENDANUBE_SECRETO, secretos de wrangler). Sin la app, el botón
 * de Tiendanube dice que no está disponible.
 */
function extrasSandbox() {
	const id = String(env.TIENDANUBE_CLIENT_ID ?? '');
	const secreto = String(env.TIENDANUBE_SECRETO ?? '');
	// Sólo en las pruebas de punta a punta: un Tiendanube falso en localhost.
	const local = localDe(env.TIENDANUBE_LOCAL);
	return {
		cifra: env.ENCRYPTION_KEY ? { cifrar: encrypt, descifrar: decrypt } : undefined,
		tiendanube: id && secreto ? { id, secreto, local } : null,
		fuentes: local ? fuentes({ tiendanubeLocal: local }) : undefined
	};
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

/** Reseñas y preguntas de la tienda online en la API Go, con la IP del visitante (tope por IP). */
export const vidrieraDe = (event: Evento) =>
	new VidrieraApi(String(env.API_URL), String(env.CLAVE_BFF ?? ''), clienteDe(event));
