import { env } from 'cloudflare:workers';
import { error, fail } from '@sveltejs/kit';
import * as v from 'valibot';
import { LEGALES } from '#lib/legales.ts';
import { errores } from '#lib/server/auth/esquemas.ts';
import { ErrorCuenta } from '#lib/server/auth/cuentas.ts';
import { cuentasDe, usaSandbox } from '#lib/server/datos/index.ts';
import { codigoDemo, esquema } from '#lib/server/legales/arrepentimiento.ts';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = ({ params }) => {
	const legal = LEGALES.find((l) => l.slug === params.slug);
	if (!legal) error(404, 'No existe esa página legal.');
	return { legal };
};

export const actions: Actions = {
	/** Botón de arrepentimiento: sólo existe en /legales/arrepentimiento. */
	revocar: async (event) => {
		if (event.params.slug !== 'arrepentimiento') error(404);
		const f = Object.fromEntries(await event.request.formData());
		const previo = {
			email: String(f.email ?? ''),
			nombre: String(f.nombre ?? ''),
			documento: String(f.documento ?? ''),
			plan: String(f.plan ?? ''),
			motivo: String(f.motivo ?? '')
		};
		// Campo trampa: lo completan los bots, no las personas (está oculto).
		if (f.web) return { codigo: 'ARR-0000-000000' };
		const r = v.safeParse(esquema, previo);
		if (!r.success) return fail(400, { ...previo, errores: errores(r.issues) });
		if (usaSandbox()) return { codigo: await codigoDemo(env.SANDBOX) };
		try {
			const { codigo } = await cuentasDe(event).arrepentimiento(r.output);
			return { codigo };
		} catch (e) {
			if (!(e instanceof ErrorCuenta)) throw e;
			const errs = Object.keys(e.campos).length ? e.campos : { form: e.message };
			return fail(e.estado >= 500 ? 503 : e.estado, { ...previo, errores: errs });
		}
	}
};
