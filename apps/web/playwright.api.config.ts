import { defineConfig } from '@playwright/test';

/**
 * E2E contra la app real: la web en modo API (DATA_SOURCE=api) hablando con la API Go, Postgres y
 * Mailpit (infra/compose.dev.yml). Necesita la API corriendo en :4000 (bun run api:dev) con su base
 * migrada; la web la levanta este config.
 */
const API = process.env.E2E_API ?? 'http://127.0.0.1:4000';
export default defineConfig({
	testDir: 'e2e-api',
	testMatch: '**/*.e2e.ts',
	timeout: 60_000,
	workers: 1,
	use: { baseURL: 'http://localhost:4174', viewport: { width: 1440, height: 900 } },
	webServer: {
		// La base local del monitor (/estado) también en modo API: el sitio la lee igual.
		command: `wrangler d1 migrations apply mostrador-estado --local >/dev/null && wrangler dev --local --var HORNERO:simulado .svelte-kit/cloudflare/_worker.js --port 4174 --var DATA_SOURCE:api --var API_URL:${API} --var CLAVE_BFF:clave-bff-e2e`,
		url: 'http://localhost:4174',
		reuseExistingServer: false
	}
});
