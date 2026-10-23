import { env } from 'cloudflare:workers';
import { datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import { dbSandbox } from '#lib/server/sandbox/sandbox.ts';
import { BOT, habilitado, vinculoDe } from '#lib/server/telegram.ts';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	const d = datosDe(locals);
	const [estado, mp, ml, tg] = await Promise.all([
		d.integraciones(),
		d.mercadoPago(),
		d.mercadoLibre(),
		locals.sandbox ? vinculoDe(dbSandbox(env.SANDBOX), locals.sandbox.id) : null
	]);
	// Telegram (Hornero para la dueña) está conectado si hay un chat vinculado.
	estado.conectadas.tg = !!tg;
	return {
		estado,
		mp,
		ml,
		telegram: {
			disponible: habilitado() && !!locals.sandbox,
			nombre: tg?.nombre ?? null,
			bot: BOT
		},
		demo: usaSandbox()
	};
};
