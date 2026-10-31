import { contarVista } from '#lib/server/tienda/vistas.ts';
import type { RequestHandler } from './$types';

/**
 * La navegación dentro de la tienda no pide HTML: el layout avisa acá cada página nueva (con
 * sendBeacon, sin esperar). Se cuenta con los mismos filtros y topes que la carga completa
 * (lib/server/tienda/vistas.ts). Siempre 204: no hay nada que mirar.
 */
export const POST: RequestHandler = async (event) => {
	const texto = await event.request.text().catch(() => '');
	if (texto.length <= 300) {
		try {
			const b = JSON.parse(texto) as { r?: unknown; p?: unknown };
			if (typeof b.r === 'string')
				contarVista(event, b.r, typeof b.p === 'string' ? b.p : undefined);
		} catch {
			// Cuerpo inválido: no cuenta.
		}
	}
	return new Response(null, { status: 204, headers: { 'cache-control': 'no-store' } });
};
