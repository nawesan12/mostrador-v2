interface Env {
	APP: Fetcher;
	INTERNAL_SECRET: string;
}

export default {
	async scheduled(controller: ScheduledController, env: Env, ctx: ExecutionContext) {
		const now = new Date(controller.scheduledTime).toISOString();
		ctx.waitUntil(
			env.APP.fetch(
				`https://mostrador.iamnaw.dev/api/internal/cron?now=${encodeURIComponent(now)}`,
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
