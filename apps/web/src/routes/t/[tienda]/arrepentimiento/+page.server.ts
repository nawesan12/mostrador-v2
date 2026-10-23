import { env } from 'cloudflare:workers';
import { error, fail } from '@sveltejs/kit';
import * as v from 'valibot';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { codigoPedido } from '#lib/server/tienda/pedido.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import type { Actions, PageServerLoad } from './$types';

/**
 * Botón de arrepentimiento de la tienda (Ley 24.240, art. 34; Res. 424/2020): revocar una compra
 * hasta 10 días después de recibirla, sin dar motivo, con un código de trámite.
 */
export const load: PageServerLoad = async ({ params, cookies, url }) => {
	const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
	if (!r) error(404, 'Esta tienda no existe.');
	return { tienda: r.tienda, codigo: url.searchParams.get('pedido') ?? '' };
};

const Pedido = v.object({
	pedido: v.pipe(
		v.string(),
		v.trim(),
		v.minLength(1, 'Poné el código o el número del pedido.'),
		v.maxLength(80)
	),
	email: v.pipe(v.string(), v.trim(), v.email('Revisá el email.'), v.maxLength(120)),
	nombre: v.pipe(v.string(), v.trim(), v.minLength(2, 'Poné tu nombre.'), v.maxLength(80)),
	motivo: v.optional(v.pipe(v.string(), v.trim(), v.maxLength(500)), '')
});

export const actions: Actions = {
	default: async ({ request, params, cookies }) => {
		const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
		if (!r) error(404, 'Esta tienda no existe.');
		const f = Object.fromEntries(await request.formData());
		if (f.web) return fail(400, { error: 'No pudimos enviarlo.' });
		const b = v.safeParse(Pedido, f);
		if (!b.success) return fail(400, { error: b.issues[0].message, previo: f });
		const p = b.output;
		if (r.tienda.vitrina)
			return fail(409, {
				error: 'Es una tienda de ejemplo: acá no hay compras para revocar.',
				previo: f
			});
		// Con la API: el código público del pedido, o su número con el email de la compra.
		if (!r.datos || !r.sandboxId) {
			const numero = /^\d{1,7}$/.test(p.pedido) ? Number(p.pedido) : 0;
			const res = await fetch(
				`${String(env.API_URL)}/v1/publico/tiendas/${r.tienda.slug}/arrepentimiento`,
				{
					method: 'POST',
					headers: { 'content-type': 'application/json' },
					body: JSON.stringify({
						codigo: numero ? '' : p.pedido,
						numero,
						email: p.email,
						nombre: p.nombre,
						motivo: p.motivo
					})
				}
			);
			const j = (await res.json().catch(() => null)) as { codigo?: string; detail?: string } | null;
			if (!res.ok || !j?.codigo)
				return fail(res.status === 429 ? 429 : 400, {
					error: j?.detail ?? 'No pudimos registrarlo.',
					previo: f
				});
			return { tramite: j.codigo };
		}
		// En la demo: el código de la confirmación (número firmado) o el número del pedido.
		const numero = Number(p.pedido.split('-')[0]);
		if (!Number.isInteger(numero) || numero <= 0)
			return fail(400, { error: 'No encontramos ese pedido.', previo: f });
		if (p.pedido.includes('-') && (await codigoPedido(r.sandboxId, numero)) !== p.pedido)
			return fail(400, { error: 'No encontramos ese pedido.', previo: f });
		const t = await r.datos.arrepentirse(numero, {
			nombre: p.nombre,
			email: p.email,
			motivo: p.motivo
		});
		if (t === 'no-existe') return fail(400, { error: 'No encontramos ese pedido.', previo: f });
		if (t === 'fuera-de-plazo')
			return fail(409, {
				error:
					'Pasaron más de 10 días desde la compra. Escribile a la tienda para ver qué se puede hacer.',
				previo: f
			});
		return { tramite: t };
	}
};
