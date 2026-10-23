import { env } from 'cloudflare:workers';
import { fail, redirect, type RequestEvent } from '@sveltejs/kit';
import { dev } from '$app/env';
import * as v from 'valibot';
import { plantilla as buscarPlantilla } from '#lib/plantillas.ts';
import { datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import { COOKIE_ORIGEN, leerOrigen } from '#lib/server/origen.ts';
import { RUBROS } from '#lib/server/sandbox/rubros.ts';
import { esProveedor, irAlProveedor } from '#lib/server/auth/social.ts';
import { COOKIE, crear, DatosSandbox, dbSandbox, VIDA_MS } from '#lib/server/sandbox/sandbox.ts';
import type { Actions, PageServerLoad } from './$types';

/**
 * El onboarding (design/Onboarding Mostrador.dc.html) es el registro. En la demo no hay cuentas: se
 * recorre con el negocio de ejemplo y al terminar se crea el sandbox con la plantilla elegida. Con
 * la API, el paso 1 crea la cuenta (en /ingresar) y cada paso se guarda en la tienda.
 */
export const load: PageServerLoad = async ({ locals, url, cookies }) => {
	// La cuenta ya existe: el origen de /empezar (si lo hay) queda en la tienda, una sola vez.
	const origen = !usaSandbox() && locals.token ? leerOrigen(cookies) : null;
	if (origen) {
		const r = await fetch(`${String(env.API_URL)}/v1/panel/tienda/origen`, {
			method: 'PUT',
			headers: { 'content-type': 'application/json', authorization: `Bearer ${locals.token}` },
			body: JSON.stringify(origen),
			// Que Go lento no demore el onboarding: si no contesta, la cookie queda para la próxima.
			signal: AbortSignal.timeout(8000)
		}).catch(() => null);
		if (r?.ok || r?.status === 400) cookies.delete(COOKIE_ORIGEN, { path: '/' });
	}
	// Sin ?paso= manda lo guardado en el navegador (ojo: Number(null) es 0).
	const crudo = url.searchParams.get('paso');
	const paso = crudo === null ? NaN : Number(crudo);
	return {
		demo: usaSandbox(),
		conSesion: !!locals.token,
		paso: Number.isInteger(paso) && paso >= 0 && paso <= 7 ? paso : null
	};
};

const Estado = v.object({
	plantilla: v.pipe(
		v.string(),
		v.check((id) => !!buscarPlantilla(id), 'Elegí una de las plantillas.')
	),
	elecciones: v.record(v.string(), v.unknown())
});

async function leerEstado(request: Request) {
	const f = await request.formData();
	let crudo: unknown = null;
	try {
		crudo = JSON.parse(String(f.get('estado') ?? 'null'));
	} catch {
		// Cae en la validación.
	}
	return v.safeParse(Estado, crudo);
}

async function guardarEnApi(event: RequestEvent, terminado: boolean) {
	const r = await leerEstado(event.request);
	if (!r.success) return fail(400, { error: r.issues[0].message });
	await datosDe(event.locals).guardarOnboarding({ ...r.output, terminado });
	return null;
}

export const actions: Actions = {
	/** "Ir a mi panel": en la demo crea el sandbox con lo elegido; con la API lo guarda. */
	terminar: async (event) => {
		if (!usaSandbox()) {
			if (!event.locals.token) redirect(303, '/ingresar?m=registro');
			const error = await guardarEnApi(event, true);
			if (error) return error;
			redirect(303, '/panel');
		}
		const r = await leerEstado(event.request);
		if (!r.success) return fail(400, { error: r.issues[0].message });
		const p = buscarPlantilla(r.output.plantilla)!;
		const rubro = RUBROS.find((x) => x.nombre === p.rubro);
		const db = dbSandbox(env.SANDBOX);
		const sb = rubro && (await crear(db, rubro.id, Date.now(), p.id));
		if (!sb) return fail(400, { error: 'No pudimos armar la demo con esa plantilla.' });
		await new DatosSandbox(db, sb).guardarOnboarding({
			elecciones: r.output.elecciones,
			terminado: true
		});
		event.cookies.set(COOKIE, sb.id, {
			path: '/',
			httpOnly: true,
			secure: !dev,
			sameSite: 'lax',
			maxAge: VIDA_MS / 1000
		});
		redirect(303, '/panel');
	},

	/** Con la API, cada paso se guarda en la tienda (en la demo queda en el navegador). */
	paso: async (event) => {
		if (usaSandbox() || !event.locals.token) return { guardado: false };
		const error = await guardarEnApi(event, false);
		return error ?? { guardado: true };
	},

	/** Paso 1 con la API: la cuenta se crea en /ingresar y se vuelve al paso 2. */
	email: async ({ request }) => {
		const email = String((await request.formData()).get('email') ?? '').trim();
		redirect(303, `/ingresar?m=registro${email ? `&email=${encodeURIComponent(email)}` : ''}`);
	},
	/** Paso 1 con la API: Mercado Pago, Mercado Libre o Google, con OAuth. */
	proveedor: async (event) => {
		const p = String((await event.request.formData()).get('proveedor') ?? '');
		if (!esProveedor(p)) return fail(400, { error: 'Elegí cómo querés entrar.' });
		return irAlProveedor(event, p, '/onboarding?paso=1', 'login');
	}
};
