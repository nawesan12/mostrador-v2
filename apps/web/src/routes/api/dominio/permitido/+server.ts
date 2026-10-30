import { slugDeDominio } from '#lib/server/tienda/dominio.ts';
import type { RequestHandler } from './$types';

/**
 * Lo pregunta Caddy (on_demand_tls ask) antes de hacerle certificado a un dominio que no conoce:
 * 200 sólo si es el dominio verificado de una tienda. Así nadie consigue certificados apuntando
 * cualquier dominio al VPS.
 */
export const GET: RequestHandler = async ({ url }) => {
	const d = (url.searchParams.get('domain') ?? '').toLowerCase();
	const slug = d ? await slugDeDominio(d).catch(() => null) : null;
	return new Response(null, { status: slug ? 200 : 404 });
};
