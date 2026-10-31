import { fail, redirect } from '@sveltejs/kit';
import { env } from 'cloudflare:workers';
import { fapiDe } from '#lib/clerk.ts';
import { ErrorApi } from '#lib/server/bff.ts';
import { guardarInvitacion, tomarInvitacion } from '#lib/server/auth/invitacion.ts';
import { borrarSesion, guardarSesion } from '#lib/server/auth/sesion.ts';
import { cuentasDe, usaSandbox } from '#lib/server/datos/index.ts';
import type { Actions, PageServerLoad } from './$types';

/**
 * El link de una invitación al equipo de una tienda (docs/decisiones.md §37). Con sesión, «Sumarme»;
 * sin sesión, ingresar (si el email ya tiene cuenta) o crear la cuenta acá mismo, con clave o con
 * Google/Apple. Para Google y Apple el token viaja en una cookie corta: el ida y vuelta pasa por
 * otro sitio y /ingresar/clerk lo usa al volver.
 */
const tokenValido = (t: string | null): t is string => !!t && t.length <= 200;

export const load: PageServerLoad = async (event) => {
	if (usaSandbox()) redirect(303, '/prueba');
	const token = event.url.searchParams.get('token');
	if (!tokenValido(token)) return { invitacion: null };
	const api = cuentasDe(event);
	try {
		const invitacion = await api.verInvitacion(token);
		guardarInvitacion(event.cookies, token);
		// Con sesión abierta (y no una mirada del equipo de Tu Local Web), se suma esa cuenta.
		const sesion =
			event.locals.token && !event.locals.mirada
				? await api.sesion(event.locals.token).catch(() => null)
				: null;
		return {
			invitacion,
			token,
			sesion: sesion ? { nombre: sesion.usuario.nombre, email: sesion.usuario.email } : null,
			clerk: fapiDe(env.CLERK_PUBLISHABLE_KEY)
		};
	} catch (e) {
		if (e instanceof ErrorApi && e.estado < 500) return { invitacion: null };
		throw e;
	}
};

export const actions: Actions = {
	/** Entró con otra cuenta: sale y vuelve a la invitación para ingresar con la del email invitado. */
	salir: async (event) => {
		const token = String((await event.request.formData()).get('token') ?? '');
		if (event.locals.token)
			await cuentasDe(event)
				.salir(event.locals.token)
				.catch(() => {});
		borrarSesion(event.cookies);
		redirect(303, `/invitacion?token=${encodeURIComponent(token)}`);
	},
	aceptar: async (event) => {
		const token = String((await event.request.formData()).get('token') ?? '');
		if (!event.locals.token) redirect(303, `/invitacion?token=${encodeURIComponent(token)}`);
		try {
			await cuentasDe(event).aceptarInvitacion(token, event.locals.token);
		} catch (e) {
			if (!(e instanceof ErrorApi)) throw e;
			return fail(e.estado === 401 ? 401 : 422, {
				error:
					e.estado === 401
						? 'Tu sesión venció: ingresá de nuevo y volvé a abrir el link.'
						: e.estado === 409
							? 'Esta invitación es para otro email. Salí y entrá con la cuenta de ese email.'
							: 'La invitación venció o ya se usó. Pedile a la dueña que te la reenvíe.'
			});
		}
		tomarInvitacion(event.cookies);
		redirect(303, '/panel');
	},
	crear: async (event) => {
		const f = await event.request.formData();
		const token = String(f.get('token') ?? '');
		const nombre = String(f.get('nombre') ?? '').trim();
		const clave = String(f.get('clave') ?? '');
		if (!nombre) return fail(400, { error: '¿Cómo te llamás?', nombre });
		if (clave.length < 8)
			return fail(400, { error: 'La contraseña tiene que tener al menos 8 caracteres.', nombre });
		try {
			const s = await cuentasDe(event).registrarConInvitacion(token, nombre, clave);
			guardarSesion(event.cookies, s.token, s.venceEn);
		} catch (e) {
			if (!(e instanceof ErrorApi)) throw e;
			const error =
				e.codigo === 'email_en_uso'
					? 'Ese email ya tiene cuenta: ingresá con ella para sumarte.'
					: e.codigo === 'link_invalido'
						? 'La invitación venció o ya se usó. Pedile a la dueña que te la reenvíe.'
						: e.estado === 429
							? 'Hubo demasiados intentos. Probá de nuevo en un rato.'
							: e.message || 'Revisá los datos.';
			return fail(e.estado === 429 ? 429 : 400, { error, nombre });
		}
		tomarInvitacion(event.cookies);
		redirect(303, '/panel');
	}
};
