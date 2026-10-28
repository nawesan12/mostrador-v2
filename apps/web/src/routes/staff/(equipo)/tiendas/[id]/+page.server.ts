import { error, fail, redirect } from '@sveltejs/kit';
import { ErrorApi } from '#lib/server/bff.ts';
import { EnLaDemo } from '#lib/server/staff/demo.ts';
import { equipoDe, exigir, guardarMirada } from '#lib/server/staff/index.ts';
import type { AccionTienda } from '#lib/server/staff/tipos.ts';
import type { Actions, PageServerLoad } from './$types';

/** La ficha de una tienda (Go deja en la auditoría que alguien del equipo la abrió). */
export const load: PageServerLoad = async (event) => {
	const { sesion } = await event.parent();
	exigir(sesion.permisos, 'tiendas.ver');
	const ficha = await equipoDe(event).ficha(event.params.id);
	if (!ficha) error(404, 'Esa tienda no existe.');
	return { ficha };
};

/** Los errores de Go (o el "en la demo no se aplica") vuelven al formulario. */
async function intentar<T>(f: () => Promise<T>) {
	try {
		return { ok: true as const, valor: await f() };
	} catch (e) {
		if (e instanceof EnLaDemo) return { ok: false as const, r: fail(409, { demo: true }) };
		if (e instanceof ErrorApi)
			return { ok: false as const, r: fail(e.estado, { error: e.message }) };
		throw e;
	}
}

const numero = (v: FormDataEntryValue | null) => Math.round(Number(v) || 0);

export const actions: Actions = {
	accion: async (event) => {
		const f = await event.request.formData();
		const a: AccionTienda = {
			accion: String(f.get('accion')) as AccionTienda['accion'],
			dias: numero(f.get('dias')) || undefined,
			plan: (String(f.get('plan') ?? '') || undefined) as AccionTienda['plan'],
			monto: numero(f.get('monto')) || undefined,
			cantidad: numero(f.get('cantidad')) || undefined,
			motivo: String(f.get('motivo') ?? '').trim() || undefined,
			confirmacion: String(f.get('confirmacion') ?? '').trim() || undefined
		};
		const r = await intentar(() => equipoDe(event).accion(event.params.id, a));
		return r.ok ? { hecho: a.accion } : r.r;
	},
	nota: async (event) => {
		const texto = String((await event.request.formData()).get('texto') ?? '').trim();
		if (texto.length < 2 || texto.length > 1000)
			return fail(400, { error: 'Escribí la nota (hasta 1000 caracteres).' });
		const r = await intentar(() => equipoDe(event).nota(event.params.id, texto));
		return r.ok ? { nota: true } : r.r;
	},
	/** Ver la tienda como la ve la dueña: el panel en sólo lectura, con el aviso fijo. */
	mirar: async (event) => {
		const r = await intentar(() => equipoDe(event).mirar(event.params.id));
		if (!r.ok) return r.r;
		guardarMirada(event.cookies, r.valor.token, r.valor.venceEn);
		redirect(303, '/panel');
	}
};
