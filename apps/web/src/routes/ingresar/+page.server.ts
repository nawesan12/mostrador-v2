import { fail, redirect } from '@sveltejs/kit';
import * as v from 'valibot';
import { usaSandbox } from '#lib/server/datos/index.ts';
import { errores, Ingreso, Recuperar, Registro } from '#lib/server/auth/esquemas.ts';
import type { Actions } from './$types';

const datos = async (request: Request) => Object.fromEntries(await request.formData());

/*
 * En la demo (sandbox) no hay cuentas: entrar o registrarse lleva a la demo, y recuperar muestra la
 * pantalla de "Revisá tu email" sin mandar nada. Con la API Go (DATA_SOURCE=api) estas mismas
 * acciones autentican de verdad (PR de cuentas y sesiones).
 */
export const actions: Actions = {
	ingresar: async ({ request }) => {
		const f = await datos(request);
		const r = v.safeParse(Ingreso, f);
		if (!r.success)
			return fail(400, { modo: 'login', email: String(f.email ?? ''), errores: errores(r.issues) });
		if (usaSandbox()) redirect(303, '/demo');
		return fail(503, {
			modo: 'login',
			email: r.output.email,
			errores: { form: 'El ingreso todavía no está disponible.' }
		});
	},
	registro: async ({ request }) => {
		const f = await datos(request);
		const r = v.safeParse(Registro, f);
		if (!r.success)
			return fail(400, {
				modo: 'registro',
				nombre: String(f.nombre ?? ''),
				tienda: String(f.tienda ?? ''),
				email: String(f.email ?? ''),
				errores: errores(r.issues)
			});
		// El registro sigue en el onboarding (decisión: el onboarding es el registro).
		if (usaSandbox()) redirect(303, '/demo');
		return fail(503, {
			modo: 'registro',
			errores: { form: 'El registro todavía no está disponible.' }
		});
	},
	/** Google o WhatsApp: en la API, inicia el OAuth o el código por WhatsApp (PR de ingreso social). */
	social: async ({ request }) => {
		const proveedor = String((await request.formData()).get('proveedor') ?? '');
		if (!['google', 'whatsapp'].includes(proveedor))
			return fail(400, { modo: 'login', errores: { form: 'Elegí cómo querés ingresar.' } });
		if (usaSandbox()) redirect(303, '/demo');
		return fail(503, {
			modo: 'login',
			errores: { form: 'El ingreso con ' + proveedor + ' todavía no está disponible.' }
		});
	},
	recuperar: async ({ request }) => {
		const f = await datos(request);
		const r = v.safeParse(Recuperar, f);
		if (!r.success)
			return fail(400, {
				modo: 'recuperar',
				canal: String(f.canal ?? 'email'),
				errores: errores(r.issues)
			});
		// La respuesta es la misma exista o no la cuenta: no se revela quién tiene cuenta.
		// El destino vuelve en la respuesta, no en la URL: así no queda en el historial ni en logs.
		const destino = r.output.canal === 'email' ? r.output.email : r.output.telefono;
		return { modo: 'enviado' as const, canal: r.output.canal, a: destino };
	}
};
