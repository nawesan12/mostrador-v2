import { defineConfig } from 'vitest/config';
import adapterCloudflare from '@sveltejs/adapter-cloudflare';
import adapterNode from '@sveltejs/adapter-node';
import { fileURLToPath } from 'node:url';

// DESTINO=vps: Node en el VPS (Docker + Caddy); `cloudflare:workers` lo da plataforma/vps.ts.
const vps = process.env.DESTINO === 'vps';
import { sveltekit } from '@sveltejs/kit/vite';

export default defineConfig({
	// En el VPS todo va en el bundle: el contenedor no instala node_modules.
	ssr: vps ? { noExternal: true } : undefined,
	resolve: vps
		? {
				alias: {
					'cloudflare:workers': fileURLToPath(
						new URL('./src/lib/server/plataforma/vps.ts', import.meta.url)
					)
				}
			}
		: undefined,
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
			adapter: vps
				? adapterNode({ precompress: true })
				: adapterCloudflare({ platformProxy: { remoteBindings: !process.env.VITEST } }),
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
