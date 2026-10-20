/**
 * CSRF: la misma regla que trae SvelteKit (un POST/PUT/PATCH/DELETE de formulario tiene que venir de
 * nuestro origen), aplicada en `hooks.server.ts`. SvelteKit la desactiva (`trustedOrigins: ['*']`)
 * porque rechaza todo pedido sin `Origin`, y dos cosas llegan así desde servidores ajenos:
 * el OAuth del conector MCP (Claude, ChatGPT, Gemini) y los webhooks de terceros.
 * Esos endpoints quedan exceptuados y se autentican por su cuenta (PKCE, firma HMAC).
 */
const EXEMPT = /^\/api\/(oauth\/(token|revoke|register)|webhooks\/[a-z0-9-]+)$/;
const FORM_TYPES = [
	'application/x-www-form-urlencoded',
	'multipart/form-data',
	'text/plain',
	'application/x-sveltekit-formdata'
];
const MUTATING = ['POST', 'PUT', 'PATCH', 'DELETE'];

export function csrfForbidden(request: Request, url: URL): boolean {
	if (!MUTATING.includes(request.method)) return false;
	if (EXEMPT.test(url.pathname)) return false;
	const type = (request.headers.get('content-type') ?? '').split(';')[0].trim().toLowerCase();
	if (type && !FORM_TYPES.includes(type)) return false;
	return request.headers.get('origin') !== url.origin;
}
