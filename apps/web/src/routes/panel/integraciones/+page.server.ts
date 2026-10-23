import { env } from 'cloudflare:workers';
import { IAS, iaDe } from '#lib/panel/integraciones.ts';
import { datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import { panelApi } from '#lib/server/mcp/api.ts';
import { conexionesDe } from '#lib/server/mcp/oauth.ts';
import { dbSandbox } from '#lib/server/sandbox/sandbox.ts';
import { BOT, habilitado, vinculoDe } from '#lib/server/telegram.ts';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	const d = datosDe(locals);
	const db = locals.sandbox ? dbSandbox(env.SANDBOX) : null;
	const [estado, mp, ml, tg, conexiones] = await Promise.all([
		d.integraciones(),
		d.mercadoPago(),
		d.mercadoLibre(),
		db && locals.sandbox ? vinculoDe(db, locals.sandbox.id) : null,
		db && locals.sandbox
			? conexionesDe(db, locals.sandbox.id)
			: panelApi(locals.token).then((p) =>
					(p?.conexiones ?? []).map((c) => ({
						id: c.id,
						nombre: c.cliente,
						permisos: c.acciones ? 'lectura acciones' : 'lectura',
						creadoEn: Date.parse(c.creadaEn),
						usadoEn: Date.parse(c.usadaEn)
					}))
				)
	]);
	// Telegram (Hornero para la dueña) está conectado si hay un chat vinculado.
	estado.conectadas.tg = !!tg;
	// Manejá Mostrador desde tu IA: las conexiones reales por OAuth (servidor MCP).
	const ias = conexiones.map((c) => ({
		id: c.id,
		nombre: c.nombre,
		ia: iaDe(c.nombre).id,
		acciones: c.permisos.split(' ').includes('acciones'),
		desde: c.creadoEn,
		usada: c.usadoEn
	}));
	estado.ia = Object.fromEntries(IAS.map((p) => [p.id, ias.some((c) => c.ia === p.id)]));
	return {
		estado,
		mp,
		ml,
		telegram: {
			disponible: habilitado() && !!locals.sandbox,
			nombre: tg?.nombre ?? null,
			bot: BOT
		},
		mcp: {
			url: `${env.PUBLIC_SITE_URL}/api/mcp`,
			disponible: true,
			conexiones: ias
		},
		demo: usaSandbox()
	};
};
