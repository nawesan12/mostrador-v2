import { env } from 'cloudflare:workers';
import { json } from '@sveltejs/kit';
import { safeEqual } from '#lib/server/crypto.ts';
import { dbSandbox, purgar } from '#lib/server/sandbox/sandbox.ts';
import type { RequestHandler } from './$types';

/**
 * Lo llama mostrador-jobs (cron, por service binding) con x-internal-secret.
 * Hoy: purga los sandboxes vencidos de la demo, de a tandas chicas.
 */
export const POST: RequestHandler = async ({ request }) => {
	if (!safeEqual(request.headers.get('x-internal-secret') ?? '', String(env.INTERNAL_SECRET ?? '')))
		return json({ message: 'No autorizado' }, { status: 401 });
	const db = dbSandbox(env.SANDBOX);
	let purgados = 0;
	// Hasta 4 tandas de 50 por pasada: alcanza de sobra y acota CPU y escrituras por request.
	for (let i = 0; i < 4; i++) {
		const n = await purgar(db);
		purgados += n;
		if (n < 50) break;
	}
	return json({ purgados });
};
