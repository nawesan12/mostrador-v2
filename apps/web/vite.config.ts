import { defineConfig } from 'vitest/config';
import adapter from '@sveltejs/adapter-cloudflare';
import { sveltekit } from '@sveltejs/kit/vite';

export default defineConfig({
	plugins: [
		sveltekit({
			compilerOptions: {
				// Runes en todo el proyecto, salvo librerías. Se puede sacar en Svelte 6.
				runes: ({ filename }) =>
					filename.split(/[/\\]/).includes('node_modules') ? undefined : true,
				experimental: { async: true }
			},
			// En los tests no se abren bindings remotos (Workers AI): en CI no hay credenciales y
			// las pruebas nunca llaman al modelo.
			adapter: adapter({ platformProxy: { remoteBindings: !process.env.VITEST } }),
			// Igual que en Caja: el chequeo de origen lo hace hooks.server.ts, porque el OAuth del
			// conector MCP recibe POST sin `Origin` desde los servidores de Claude o ChatGPT.
			csrf: { trustedOrigins: ['*'] },
			experimental: { remoteFunctions: true }
		})
	],
	test: {
		expect: { requireAssertions: true },
		projects: [
			{
				extends: './vite.config.ts',
				test: {
					name: 'server',
					environment: 'node',
					include: ['src/**/*.{test,spec}.{js,ts}'],
					exclude: ['src/**/*.svelte.{test,spec}.{js,ts}']
				}
			}
		]
	}
});
