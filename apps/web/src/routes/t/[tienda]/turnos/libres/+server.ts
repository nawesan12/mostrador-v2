import { error, json } from '@sveltejs/kit';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import { horariosDe } from '#lib/server/tienda/turnos.ts';
import { fechaValida } from '#lib/turnos/libres.ts';
import type { RequestHandler } from './$types';

const ID = /^[A-Za-z0-9_-]{1,64}$/;

/** Los horarios de un día para un servicio (y un profesional, o cualquiera): libres y tachados. */
export const GET: RequestHandler = async ({ params, url, cookies, setHeaders }) => {
	const servicio = url.searchParams.get('servicio') ?? '';
	const p = url.searchParams.get('profesional') ?? '';
	const dia = url.searchParams.get('dia') ?? '';
	if (!ID.test(servicio) || (p && p !== 'cualquiera' && !ID.test(p)) || !fechaValida(dia))
		error(400, 'Elegí un servicio y un día.');
	const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
	if (!r) error(404, 'Esta tienda no existe.');
	const h = await horariosDe(r, servicio, p && p !== 'cualquiera' ? p : null, dia);
	if (!h) error(404, 'Esta tienda no toma turnos online.');
	setHeaders({ 'cache-control': 'private, no-store' });
	return json(h);
};
