import { defineConfig } from '@playwright/test';

// Corre contra el build servido por wrangler (bun run build && bun run preview), o contra
// E2E_BASE si se le pasa otra URL (por ejemplo, la de producción después de un deploy).
export default defineConfig({
	testDir: 'e2e',
	testMatch: '**/*.e2e.ts',
	timeout: 60_000,
	// Entrar a la demo arma el sandbox y navega al panel con sus chunks en frío, servidos por
	// wrangler dev a varios navegadores a la vez: en CI a veces pasa los 5 s por defecto.
	expect: { timeout: 10_000 },
	use: { baseURL: process.env.E2E_BASE ?? 'http://localhost:4173' },
	// Con el preview, un Tiendanube falso en 4179 (el canje y la API; el preview lo usa por
	// TIENDANUBE_LOCAL): el flujo de importar con OAuth nunca toca la API real.
	webServer: process.env.E2E_BASE
		? undefined
		: [
				{ command: 'bun run preview', url: 'http://localhost:4173', reuseExistingServer: true },
				{
					command: 'bun e2e/tiendanube-falso.ts',
					url: 'http://127.0.0.1:4179/salud',
					reuseExistingServer: true
				}
			],
	projects: [
		{ name: 'desktop', use: { viewport: { width: 1440, height: 900 } } },
		{
			name: 'mobile',
			use: { viewport: { width: 375, height: 812 }, isMobile: true, hasTouch: true }
		}
	]
});
