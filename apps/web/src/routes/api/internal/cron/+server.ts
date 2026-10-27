import { env } from 'cloudflare:workers';
import { json } from '@sveltejs/kit';
import { borrarDe } from '#lib/server/archivos.ts';
import { safeEqual } from '#lib/server/crypto.ts';
import { purgarLimites } from '#lib/server/limite.ts';
import { purgarOauth } from '#lib/server/mcp/oauth.ts';
import { dbSandbox, purgar } from '#lib/server/sandbox/sandbox.ts';
import type { RequestHandler } from './$types';

/**
 * Lo llama mostrador-jobs (cron, por service binding) con x-internal-secret.
 * Hoy: purga los sandboxes vencidos de la demo, de a tandas chicas (sus conexiones de IA se van
 * con ellos, y sus fotos en R2), y los códigos de OAuth, clientes sin uso y ventanas de límites vencidos.
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
	return json({ purgados });
};
