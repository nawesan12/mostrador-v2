import { defineConfig } from '@playwright/test';

/**
 * E2E contra la app real: la web en modo API (DATA_SOURCE=api) hablando con la API Go, Postgres y
 * Mailpit (infra/compose.dev.yml). Necesita la API corriendo en :4000 (bun run api:dev) con su base
 * migrada; la web la levanta este config.
 */
const API = process.env.E2E_API ?? 'http://127.0.0.1:4000';
// El puerto de la web (otro si hay más de una corrida a la vez en la máquina).
const PUERTO = process.env.E2E_PUERTO ?? '4174';
// El tiempo real va directo a la API (wrangler no rutea el WebSocket como Caddy). Por localhost,
// que es donde vive la cookie de la sesión (las cookies no miran el puerto).
const TIEMPO_REAL = API.replace(/^http/, 'ws').replace('127.0.0.1', 'localhost');
export default defineConfig({
	testDir: 'e2e-api',
	testMatch: '**/*.e2e.ts',
	timeout: 60_000,
	workers: 1,
	use: { baseURL: `http://localhost:${PUERTO}`, viewport: { width: 1440, height: 900 } },
	webServer: {
		// La base local del monitor (/estado) también en modo API: el sitio la lee igual.
		command: `wrangler d1 migrations apply mostrador-estado --local >/dev/null && wrangler dev --local --var HORNERO:simulado .svelte-kit/cloudflare/_worker.js --port ${PUERTO} --var DATA_SOURCE:api --var API_URL:${API} --var CLAVE_BFF:clave-bff-e2e --var TIEMPO_REAL_ORIGEN:${TIEMPO_REAL}`,
		url: `http://localhost:${PUERTO}`,
		reuseExistingServer: false
	}
});
