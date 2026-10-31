import { env } from 'cloudflare:workers';
import { fail, redirect, type RequestEvent } from '@sveltejs/kit';
import { dev } from '$app/env';
import * as v from 'valibot';
import { plantilla as buscarPlantilla } from '#lib/plantillas.ts';
import { datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import { COOKIE_ORIGEN, leerOrigen } from '#lib/server/origen.ts';
import { RUBROS } from '#lib/server/sandbox/rubros.ts';
import { esProveedor, irAlProveedor } from '#lib/server/auth/social.ts';
import { fapiDe } from '#lib/clerk.ts';
import { cbuValido } from '#lib/panel/cbu.ts';
import { COOKIE, crear, DatosSandbox, dbSandbox, VIDA_MS } from '#lib/server/sandbox/sandbox.ts';
import { descubrir } from '#lib/server/importar/fuentes.ts';
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
	// Con la API, la tienda de verdad (nombre y dirección) para la vista previa del paso 5.
	const tienda =
		!usaSandbox() && locals.token
			? await datosDe(locals)
					.tienda()
					.then((t) => ({ nombre: t.nombre, slug: t.slug }))
					.catch(() => null)
			: null;
	// Paso 3, Canales: si Mercado Pago y Mercado Libre ya están conectados (de verdad).
	const conectadas =
		!usaSandbox() && locals.token
			? await Promise.all([
					datosDe(locals)
						.mercadoPago()
						.then((c) => c.conectada)
						.catch(() => false),
					datosDe(locals)
						.mercadoLibre()
						.then((c) => c.conectada)
						.catch(() => false)
				]).then(([mp, ml]) => ({ mp, ml }))
			: { mp: false, ml: false };
	// Paso 6, Cobros y envíos: lo que ya está guardado, para editarlo ahí mismo.
	const cobros =
		!usaSandbox() && locals.token
			? await Promise.all([
					datosDe(locals)
						.configuracion()
						.then((c) => c.negocio)
						.catch(() => null),
					datosDe(locals)
						.entregas()
						.catch(() => null),
					datosDe(locals)
						.arca()
						.catch(() => null)
				]).then(([negocio, entregas, arca]) => ({ negocio, entregas, arca }))
			: null;
	return {
		demo: usaSandbox(),
		conectadas,
		cobros,
		// Google y Apple van por Clerk desde el navegador, como en /ingresar.
		clerk: usaSandbox() ? null : fapiDe(env.CLERK_PUBLISHABLE_KEY),
		conSesion: !!locals.token,
		tienda,
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
	if (!r.success) return { error: fail(400, { error: r.issues[0].message }), importar: null };
	await datosDe(event.locals).guardarOnboarding({ ...r.output, terminado });
	return {
		error: null,
		importar: v.parse(Importar, r.output.elecciones.importar ?? null),
		elecciones: r.output.elecciones
	};
}

/**
 * Las secciones que se eligieron en el paso 5 pasan a la tienda (las de Mi tienda): antes se guardaban
 * con el resto del onboarding y la tienda las ignoraba. Si Mi tienda tiene cambios sin publicar, sólo
 * quedan en el borrador. Si algo falla, el onboarding termina igual.
 */
async function aplicarSecciones(event: RequestEvent, elecciones: Record<string, unknown>) {
	const s = v.safeParse(v.record(v.string(), v.boolean()), elecciones.secciones);
	if (!s.success) return;
	const ver: Record<string, boolean | undefined> = {
		prods: s.output.dest,
		cats: s.output.cat,
		ig: s.output.ig,
		revs: s.output.res
	};
	try {
		const datos = datosDe(event.locals);
		const mt = await datos.miTienda();
		const secciones = mt.diseno.secciones.map((x) =>
			typeof ver[x.id] === 'boolean' ? { ...x, visible: ver[x.id]! } : x
		);
		if (JSON.stringify(secciones) === JSON.stringify(mt.diseno.secciones)) return;
		await datos.guardarMiTienda({ ...mt.diseno, secciones });
		if (!mt.sinPublicar) await datos.publicarTienda();
	} catch {
		// Se cambian después desde Mi tienda.
	}
}

/** La tienda a traer que se eligió en el paso de importación (inválida: no se importa nada). */
const Importar = v.fallback(
	v.nullable(
		v.object({
			plataforma: v.picklist(['tiendanube', 'shopify', 'woocommerce']),
			url: v.pipe(v.string(), v.trim(), v.maxLength(300))
		})
	),
	null
);

/**
 * Al terminar con la API: si eligió una tienda, arranca la importación de verdad y la deja mirando en
 * Productos. Tiendanube pasa por su OAuth (que vuelve a Productos importando); Shopify y WooCommerce
 * con el link de la tienda pública.
 */
async function arrancarImportacion(
	event: RequestEvent,
	imp: v.InferOutput<typeof Importar>
): Promise<string> {
	if (!imp) return '/panel';
	if (imp.plataforma === 'tiendanube') return '/panel/importar/tiendanube';
	const d = await descubrir(imp.plataforma, imp.url).catch(() => null);
	if (!d || typeof d === 'string') return '/panel/productos?importar=1';
	const res = await datosDe(event.locals)
		.crearImportacion({ plataforma: d.fuente, url: d.origen })
		.catch(() => null);
	return res && typeof res !== 'string'
		? '/panel/productos?importando=1'
		: '/panel/productos?importar=1';
}

const campo = (f: FormData, k: string, max: number) =>
	String(f.get(k) ?? '')
		.trim()
		.slice(0, max);

export const actions: Actions = {
	/**
	 * Paso 2 con la API: el nombre de la tienda y su dirección (decisiones §38). Primero la
	 * dirección, que es lo que puede no estar disponible; el nombre va con los datos del negocio.
	 */
	tienda: async (event) => {
		if (usaSandbox() || !event.locals.token) return fail(400, { error: 'Primero creá tu cuenta.' });
		const f = await event.request.formData();
		const nombre = campo(f, 'nombre', 80);
		const slug = campo(f, 'slug', 60);
		if (!nombre) return fail(400, { error: 'Poné el nombre de tu tienda.' });
		const datos = datosDe(event.locals);
		const [t, conf] = await Promise.all([datos.tienda(), datos.configuracion()]);
		if (slug && slug !== t.slug) {
			const r = await datos.cambiarDireccion(slug);
			if (!r.ok) return fail(400, { error: r.mensaje });
		}
		if (nombre !== conf.negocio.nombre) await datos.guardarNegocio({ ...conf.negocio, nombre });
		return { tienda: { nombre, slug: slug || t.slug } };
	},
	/** Paso 6: alias, CBU y titular para cobrar por transferencia (todo opcional). */
	transferencia: async (event) => {
		if (usaSandbox()) return { guardado: 'transferencia' as const };
		const f = await event.request.formData();
		const alias = campo(f, 'alias', 20);
		const cbu = campo(f, 'cbu', 30).replace(/\D/g, '');
		const titular = campo(f, 'titular', 80);
		if (alias && !/^[a-zA-Z0-9.-]{6,20}$/.test(alias))
			return fail(400, {
				donde: 'transferencia',
				error: 'El alias tiene de 6 a 20 letras, números o puntos.'
			});
		if (cbu && !cbuValido(cbu))
			return fail(400, {
				donde: 'transferencia',
				error: 'Revisá el CBU o CVU: son 22 números y alguno no coincide.'
			});
		const datos = datosDe(event.locals);
		const actual = (await datos.configuracion()).negocio;
		await datos.guardarNegocio({ ...actual, alias, cbu, titular });
		return { guardado: 'transferencia' as const };
	},
	/** Paso 6: retiro en el local y envío en moto (todo opcional). */
	entregas: async (event) => {
		if (usaSandbox()) return { guardado: 'entregas' as const };
		const f = await event.request.formData();
		const moto = f.get('moto') === 'on';
		const e = {
			retiro: f.get('retiro') === 'on',
			retiroDireccion: campo(f, 'retiroDireccion', 160),
			retiroHorario: campo(f, 'retiroHorario', 80),
			moto,
			motoCosto: Math.max(
				0,
				Math.min(1_000_000, Number(campo(f, 'motoCosto', 9).replace(/\D/g, '')) || 0)
			),
			motoZona: campo(f, 'motoZona', 160),
			motoDemora: campo(f, 'motoDemora', 60)
		};
		if (moto && !e.motoZona)
			return fail(400, {
				donde: 'entregas',
				error: 'Contanos adónde llega la moto (barrios o localidades).'
			});
		try {
			await datosDe(event.locals).guardarEntregas(e);
		} catch (x) {
			const m = x instanceof Error ? x.message : '';
			return fail(400, { donde: 'entregas', error: m || 'No se pudo guardar. Probá de nuevo.' });
		}
		return { guardado: 'entregas' as const };
	},
	/** "Ir a mi panel": en la demo crea el sandbox con lo elegido; con la API lo guarda. */
	terminar: async (event) => {
		if (!usaSandbox()) {
			if (!event.locals.token) redirect(303, '/ingresar?m=registro');
			const g = await guardarEnApi(event, true);
			if (g.error) return g.error;
			await aplicarSecciones(event, g.elecciones);
			redirect(303, await arrancarImportacion(event, g.importar));
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
		const g = await guardarEnApi(event, false);
		return g.error ?? { guardado: true };
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
