/**
 * La fuente de datos del backoffice del equipo: la API Go con la sesión del equipo (DATA_SOURCE=api)
 * o el backoffice de ejemplo de la demo. La sesión del equipo vive en su propia cookie: nunca es la
 * de una tienda (ni al revés).
 */
import { dev } from '$app/env';
import { error, redirect, type Cookies, type RequestEvent } from '@sveltejs/kit';
import { env } from 'cloudflare:workers';
import { ROLES, type Rol } from '#lib/staff/permisos.ts';
import { usaSandbox } from '../datos/index.ts';
import { AccesoEquipoApi, EquipoApi } from './api.ts';
import { EquipoDemo } from './demo.ts';
import type { DatosEquipo } from './tipos.ts';

/** Cookie de la sesión del equipo (`__Host-` en producción, como la del panel). */
export const COOKIE_EQUIPO = dev ? 'mostrador_equipo' : '__Host-mostrador_equipo';
/** Token de sólo lectura para ver el panel de una tienda (30 min). */
export const COOKIE_MIRADA = dev ? 'mostrador_mirada' : '__Host-mostrador_mirada';
/** El rol que se muestra en la demo (para ver cómo cambia la navegación). */
export const COOKIE_ROL_DEMO = 'staff_rol_demo';

const opciones = (vence: string) => ({
	path: '/',
	httpOnly: true,
	secure: !dev,
	sameSite: 'lax' as const,
	expires: new Date(vence)
});
export const guardarSesionEquipo = (c: Cookies, token: string, vence: string) =>
	c.set(COOKIE_EQUIPO, token, opciones(vence));
export const borrarSesionEquipo = (c: Cookies) =>
	c.delete(COOKIE_EQUIPO, { path: '/', secure: !dev });
export const guardarMirada = (c: Cookies, token: string, vence: string) =>
	c.set(COOKIE_MIRADA, token, opciones(vence));
export const borrarMirada = (c: Cookies) => c.delete(COOKIE_MIRADA, { path: '/', secure: !dev });

export function rolDemo(c: Cookies): Rol {
	const r = c.get(COOKIE_ROL_DEMO);
	return ROLES.some((x) => x.id === r) ? (r as Rol) : 'duenio';
}

type Evento = Pick<RequestEvent, 'request' | 'getClientAddress' | 'cookies' | 'locals'>;

function clienteDe(event: Evento) {
	let ip = '';
	try {
		ip = event.request.headers.get('cf-connecting-ip') ?? event.getClientAddress();
	} catch {
		// Sin dirección (tests): Go usa la de la conexión.
	}
	return { ip, agente: event.request.headers.get('user-agent') ?? '' };
}

export const accesoEquipo = (event: Evento) =>
	new AccesoEquipoApi(String(env.API_URL), String(env.CLAVE_BFF ?? ''), clienteDe(event));

export function equipoDe(event: Evento): DatosEquipo {
	if (usaSandbox()) return new EquipoDemo(rolDemo(event.cookies));
	if (!event.locals.equipo) redirect(303, '/staff/ingresar');
	return new EquipoApi(
		String(env.API_URL),
		String(env.CLAVE_BFF ?? ''),
		clienteDe(event),
		event.locals.equipo
	);
}

/** Corta con 403 si la sesión no tiene el permiso (Go lo vuelve a chequear en cada pedido). */
export function exigir(permisos: readonly string[], permiso: string) {
	if (!permisos.includes(permiso))
		error(403, { message: 'Tu rol no tiene acceso a esta sección.', code: 'sin_permiso' });
}
