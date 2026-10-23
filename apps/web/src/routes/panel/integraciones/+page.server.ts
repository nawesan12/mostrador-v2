import { env } from 'cloudflare:workers';
import { datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import { dbSandbox } from '#lib/server/sandbox/sandbox.ts';
import { BOT, habilitado, vinculoDe } from '#lib/server/telegram.ts';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, url }) => {
	const d = datosDe(locals);
	const [estado, mp, ml, tg, medicion, tienda] = await Promise.all([
		d.integraciones(),
		d.mercadoPago(),
		d.mercadoLibre(),
		locals.sandbox ? vinculoDe(dbSandbox(env.SANDBOX), locals.sandbox.id) : null,
		d.medicion(),
		d.tienda()
	]);
	// Telegram (Hornero para la dueña) está conectado si hay un chat vinculado.
	estado.conectadas.tg = !!tg;
	return {
		estado,
		mp,
		ml,
		medicion,
		// Google Merchant Center y Meta Commerce Manager leen el mismo feed.
		feed: `${url.origin}/t/${tienda.slug}/feed.xml`,
		tienda: `${url.origin}/t/${tienda.slug}`,
		telegram: {
			disponible: habilitado() && !!locals.sandbox,
			nombre: tg?.nombre ?? null,
			bot: BOT
		},
		demo: usaSandbox()
	};
};
