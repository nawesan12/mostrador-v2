import { avisar } from '../web/src/lib/server/estado/avisos.ts';
import { chequear, chequeos, registrar } from '../web/src/lib/server/estado/monitor.ts';

interface Env {
	APP: Fetcher;
	ESTADO: D1Database;
	INTERNAL_SECRET: string;
	/** URL de la API Go; vacía en la demo (la API no se chequea). */
	API_URL?: string;
	/** Para los avisos de /estado por mail; sin clave, sólo se registran. */
	RESEND_API_KEY?: string;
}

const SITIO = 'https://mostrador.iamnaw.dev';

/** /estado: un chequeo por componente, se guarda y se avisa a los suscriptos si cambió algo. */
async function monitor(env: Env, ahora: number) {
	const resultados = await chequear(chequeos(SITIO, env.API_URL ?? ''), {
		// Al propio sitio, por service binding (un worker no puede pedirle por URL a otro de la zona).
		sitio: (url, init) => env.APP.fetch(url, init),
		externo: (url, init) => fetch(url, init)
	});
	const cambios = await registrar(env.ESTADO, resultados, ahora);
	const avisos = await avisar(env.ESTADO, cambios, {
		resend: env.RESEND_API_KEY,
		de: 'Estado de Mostrador <estado@mostrador.iamnaw.dev>',
		sitio: SITIO
	});
	console.log('estado', JSON.stringify({ caidos: resultados.filter((r) => !r.ok).map((r) => r.componente), cambios: cambios.length, ...avisos }));
}

export default {
	async scheduled(controller: ScheduledController, env: Env, ctx: ExecutionContext) {
		const now = new Date(controller.scheduledTime).toISOString();
		ctx.waitUntil(monitor(env, controller.scheduledTime).catch((e) => console.error('estado', e)));
		ctx.waitUntil(
			env.APP.fetch(
				`${SITIO}/api/internal/cron?now=${encodeURIComponent(now)}`,
				{
					method: 'POST',
					// SvelteKit rechaza POST sin Content-Type (protección CSRF).
					headers: { 'x-internal-secret': env.INTERNAL_SECRET, 'content-type': 'application/json' },
					body: '{}'
				}
			).then(async (r) => console.log('cron', r.status, (await r.text()).slice(0, 500)))
		);
	}
};
