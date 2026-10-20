import { fail, redirect, type RequestEvent } from '@sveltejs/kit';
import * as v from 'valibot';
import { cuentasDe, usaSandbox } from '#lib/server/datos/index.ts';
import { ErrorCuenta } from '#lib/server/auth/cuentas.ts';
import { errores, Ingreso, Recuperar, Registro, Restablecer } from '#lib/server/auth/esquemas.ts';
import { guardarSesion } from '#lib/server/auth/sesion.ts';
import type { Actions } from './$types';

const datos = async (request: Request) => Object.fromEntries(await request.formData());

/** Un error de la API como errores del formulario: por campo si los trae, si no, general. */
function deApi(e: unknown) {
	if (!(e instanceof ErrorCuenta)) throw e;
	const errs = Object.keys(e.campos).length ? e.campos : { form: e.message };
	return { estado: e.estado >= 500 ? 503 : e.estado, errores: errs };
}

/*
 * En la demo (sandbox) no hay cuentas: entrar o registrarse lleva a la demo, y recuperar muestra la
 * pantalla de "Revisá tu email" sin mandar nada (docs/decisiones.md §12). Con la API Go
 * (DATA_SOURCE=api) autentican de verdad y la sesión queda en una cookie httpOnly.
 */
export const actions: Actions = {
	ingresar: async (event) => {
		const f = await datos(event.request);
		const r = v.safeParse(Ingreso, f);
		if (!r.success)
			return fail(400, { modo: 'login', email: String(f.email ?? ''), errores: errores(r.issues) });
		if (usaSandbox()) redirect(303, '/demo');
		try {
			const s = await cuentasDe(event).ingresar(
				r.output.email,
				r.output.clave,
				r.output.recordarme === 'on'
			);
			guardarSesion(event.cookies, s.token, s.venceEn);
		} catch (e) {
			const x = deApi(e);
			return fail(x.estado, { modo: 'login', email: r.output.email, errores: x.errores });
		}
		redirect(303, '/panel');
	},
	registro: async (event) => {
		const f = await datos(event.request);
		const r = v.safeParse(Registro, f);
		const previo = {
			modo: 'registro',
			nombre: String(f.nombre ?? ''),
			tienda: String(f.tienda ?? ''),
			email: String(f.email ?? '')
		};
		if (!r.success) return fail(400, { ...previo, errores: errores(r.issues) });
		// El registro sigue en el onboarding (decisión: el onboarding es el registro). En la demo no
		// hay cuentas: se sigue igual, con el negocio de ejemplo.
		if (usaSandbox()) redirect(303, '/onboarding?paso=1');
		try {
			const { nombre, tienda, email, clave } = r.output;
			const s = await cuentasDe(event).registrar({ nombre, tienda, email, clave });
			guardarSesion(event.cookies, s.token, s.venceEn);
		} catch (e) {
			const x = deApi(e);
			return fail(x.estado, { ...previo, errores: x.errores });
		}
		redirect(303, '/onboarding?paso=1');
	},
	/** Google o WhatsApp: con la API, inicia el OAuth o el código por WhatsApp (próximo PR). */
	social: async ({ request }) => {
		const proveedor = String((await request.formData()).get('proveedor') ?? '');
		if (!['google', 'whatsapp'].includes(proveedor))
			return fail(400, { modo: 'login', errores: { form: 'Elegí cómo querés ingresar.' } });
		if (usaSandbox()) redirect(303, '/demo');
		return fail(503, {
			modo: 'login',
			errores: {
				form: `El ingreso con ${proveedor === 'google' ? 'Google' : 'WhatsApp'} todavía no está disponible.`
			}
		});
	},
	recuperar: async (event: RequestEvent) => {
		const f = await datos(event.request);
		const r = v.safeParse(Recuperar, f);
		if (!r.success)
			return fail(400, {
				modo: 'recuperar',
				canal: String(f.canal ?? 'email'),
				errores: errores(r.issues)
			});
		if (!usaSandbox()) {
			if (r.output.canal === 'whatsapp')
				return fail(503, {
					modo: 'recuperar',
					canal: 'whatsapp',
					errores: { form: 'Recuperar por WhatsApp todavía no está disponible. Probá por email.' }
				});
			try {
				await cuentasDe(event).recuperar(r.output.email);
			} catch (e) {
				const x = deApi(e);
				return fail(x.estado, { modo: 'recuperar', canal: 'email', errores: x.errores });
			}
		}
		// El destino vuelve en la respuesta, no en la URL: así no queda en el historial ni en logs.
		const destino = r.output.canal === 'email' ? r.output.email : r.output.telefono;
		return { modo: 'enviado' as const, canal: r.output.canal, a: destino };
	},
	restablecer: async (event) => {
		const f = await datos(event.request);
		const r = v.safeParse(Restablecer, f);
		if (!r.success) return fail(400, { modo: 'restablecer', errores: errores(r.issues) });
		// En la demo no hay cuentas ni links reales.
		if (usaSandbox()) redirect(303, '/demo');
		try {
			await cuentasDe(event).restablecer(r.output.token, r.output.clave);
		} catch (e) {
			const x = deApi(e);
			return fail(x.estado, { modo: 'restablecer', errores: x.errores });
		}
		redirect(303, '/ingresar?m=login&listo=clave');
	}
};
