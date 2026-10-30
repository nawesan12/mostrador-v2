import type { RequestHandler } from './$types';

/** Qué indexan los buscadores: la web y las tiendas sí; el panel, el ingreso y lo privado, no. */
export const GET: RequestHandler = () =>
	new Response(
		[
			'User-agent: *',
			'Allow: /',
			'Disallow: /panel',
			'Disallow: /ingresar',
			'Disallow: /onboarding',
			'Disallow: /staff',
			'Disallow: /partners/panel',
			'Disallow: /api/',
			'Disallow: /oauth/',
			'Disallow: /prueba',
			'Disallow: /integraciones/',
			'Disallow: /t/*/checkout',
			'Disallow: /t/*/pedido/',
			'',
			'Sitemap: https://tulocalweb.com.ar/sitemap.xml',
			''
		].join('\n'),
		{
			headers: {
				'content-type': 'text/plain; charset=utf-8',
				'cache-control': 'public, max-age=3600'
			}
		}
	);
