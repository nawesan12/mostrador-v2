import { defineConfig } from '@playwright/test';

// Corre contra el build servido por wrangler (bun run build && bun run preview), o contra
// E2E_BASE si se le pasa otra URL (por ejemplo, la de producción después de un deploy).
export default defineConfig({
	testDir: 'e2e',
	testMatch: '**/*.e2e.ts',
	timeout: 60_000,
	use: { baseURL: process.env.E2E_BASE ?? 'http://localhost:4173' },
	webServer: process.env.E2E_BASE
		? undefined
		: { command: 'bun run preview', url: 'http://localhost:4173', reuseExistingServer: true },
	projects: [
		{ name: 'desktop', use: { viewport: { width: 1440, height: 900 } } },
		{
			name: 'mobile',
			use: { viewport: { width: 375, height: 812 }, isMobile: true, hasTouch: true }
		}
	]
});
