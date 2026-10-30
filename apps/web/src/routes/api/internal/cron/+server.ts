import { env } from 'cloudflare:workers';
import { json } from '@sveltejs/kit';
import { borrarDe } from '#lib/server/archivos.ts';
import { safeEqual } from '#lib/server/crypto.ts';
import { purgarLimites } from '#lib/server/limite.ts';
import { purgarOauth } from '#lib/server/mcp/oauth.ts';
import { dbSandbox, purgar } from '#lib/server/sandbox/sandbox.ts';
import { avisar } from '#lib/server/estado/avisos.ts';
import { chequear, chequeos, registrar } from '#lib/server/estado/monitor.ts';
import { request as https } from 'node:https';
import type { RequestHandler } from './$types';

/**
 * ARCA todavía negocia con una clave Diffie-Hellman chica que el OpenSSL de Node rechaza (Go, que
 * es el que factura, usa ECDHE y no tiene el problema). Para el chequeo de /estado alcanza con ver si
 * responde, así que ahí se acepta su cifrado; no viaja ningún dato.
 */
function pedirArca(url: string, init: RequestInit): Promise<Response> {
	return new Promise((ok, mal) => {
		const req = https(
			url,
			{ method: init.method ?? 'GET', ciphers: 'DEFAULT@SECLEVEL=0', timeout: 10_000 },
			(res) => {
				res.resume();
				ok(new Response(null, { status: res.statusCode ?? 0 }));
			}
		);
		req.on('timeout', () => req.destroy(new Error('timeout')));
		req.on('error', mal);
		req.end();
	});
}

/**
 * Lo llama cada hora el timer tulocalweb-cron del VPS, con x-internal-secret.
 * Chequea los componentes de /estado (y avisa a los suscriptos si algo cambió), purga los datos de
 * prueba vencidos y los códigos de OAuth, clientes sin uso y ventanas de límites vencidos.
 */
export const POST: RequestHandler = async ({ request }) => {
	if (!safeEqual(request.headers.get('x-internal-secret') ?? '', String(env.INTERNAL_SECRET ?? '')))
		return json({ message: 'No autorizado' }, { status: 401 });
	const db = dbSandbox(env.SANDBOX);
	let purgados = 0;
	// Hasta 4 tandas de 50 por pasada: alcanza de sobra y acota CPU y escrituras por request.
	for (let i = 0; i < 4; i++) {
		const n = await purgar(db, Date.now(), 50, (ids) => borrarDe(db, ids));
		purgados += n;
		if (n < 50) break;
	}
	await Promise.all([purgarOauth(db), purgarLimites(db)]);
	const sitio = String(env.PUBLIC_SITE_URL ?? 'https://tulocalweb.com.ar');
	const resultados = await chequear(chequeos(sitio, String(env.API_URL ?? '')), {
		sitio: (url, init) => fetch(url, init),
		externo: (url, init) =>
			new URL(url).hostname.endsWith('afip.gov.ar') ? pedirArca(url, init) : fetch(url, init)
	});
	const cambios = await registrar(env.ESTADO, resultados, Date.now());
	await avisar(env.ESTADO, cambios, {
		resend: undefined,
		de: 'Estado de Tu Local Web <hola@tulocalweb.com.ar>',
		sitio
	});
	return json({ purgados, caidos: resultados.filter((x) => !x.ok).map((x) => x.componente) });
};
