import { fail, redirect, type RequestEvent } from '@sveltejs/kit';
import * as v from 'valibot';
import { cuentasDe, usaSandbox } from '#lib/server/datos/index.ts';
import { ErrorCuenta } from '#lib/server/auth/cuentas.ts';
import {
	Enlace,
	errores,
	Ingreso,
	Recuperar,
	Registro,
	Restablecer
} from '#lib/server/auth/esquemas.ts';
import { esProveedor, irAlProveedor } from '#lib/server/auth/social.ts';
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
	/** Google, Mercado Pago y Mercado Libre van al proveedor; WhatsApp pide el número. */
	social: async (event) => {
		const proveedor = String((await event.request.formData()).get('proveedor') ?? '');
		if (proveedor === 'whatsapp') return { modo: 'whatsapp' as const };
		if (!esProveedor(proveedor))
			return fail(400, { modo: 'login', errores: { form: 'Elegí cómo querés ingresar.' } });
		return irAlProveedor(event, proveedor, '/demo', 'login');
	},
	/** Pide el link para entrar al WhatsApp de la cuenta. */
	enlace: async (event) => {
		const f = await datos(event.request);
		const r = v.safeParse(Enlace, f);
		if (!r.success) return fail(400, { modo: 'whatsapp', errores: errores(r.issues) });
		if (usaSandbox()) redirect(303, '/demo');
		try {
			await cuentasDe(event).pedirEnlace(r.output.telefono);
		} catch (e) {
			const x = deApi(e);
			return fail(x.estado, { modo: 'whatsapp', errores: x.errores });
		}
		return {
			modo: 'enviado' as const,
			canal: 'whatsapp' as const,
			tipo: 'ingreso' as const,
			a: r.output.telefono
		};
	},
	/** Entra con el link del WhatsApp (un clic: así la vista previa del link no lo consume). */
	entrar: async (event) => {
		const token = String((await event.request.formData()).get('token') ?? '');
		if (usaSandbox()) redirect(303, '/demo');
		try {
			const s = await cuentasDe(event).usarEnlace(token);
			guardarSesion(event.cookies, s.token, s.venceEn);
		} catch (e) {
			const x = deApi(e);
			return fail(x.estado, { modo: 'enlace', errores: x.errores });
		}
		redirect(303, '/panel');
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
			try {
				await cuentasDe(event).recuperar(
					r.output.canal === 'email' ? { email: r.output.email } : { telefono: r.output.telefono }
				);
			} catch (e) {
				const x = deApi(e);
				return fail(x.estado, { modo: 'recuperar', canal: r.output.canal, errores: x.errores });
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
