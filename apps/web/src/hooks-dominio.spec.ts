import { describe, expect, it, vi } from 'vitest';

// El dominio de la tienda resuelve, pero el plan que rige no incluye dominio propio.
vi.mock('#lib/server/tienda/dominio.ts', () => ({
	tiendaDeDominio: vi.fn(async () => ({ slug: 'luna', propio: false })),
	olvidarDominio: vi.fn()
}));

const { handle } = await import('./hooks.server.ts');

function pedir(url: string) {
	const event = {
		url: new URL(url),
		request: new Request(url),
		locals: {} as Record<string, unknown>,
		cookies: { get: () => undefined },
		getClientAddress: () => '127.0.0.1'
	};
	return handle({ event, resolve: vi.fn() } as never);
}

describe('dominio propio sin el plan', () => {
	it('manda a la dirección de Tu Local Web con un 302 sin caché (nunca un 301 que quede guardado)', async () => {
		const r = await pedir('https://lunadeco.com.ar/producto/mate?color=rojo');
		expect(r.status).toBe(302);
		expect(r.headers.get('location')).toBe(
			'https://luna.tulocalweb.com.ar/producto/mate?color=rojo'
		);
		expect(r.headers.get('cache-control')).toBe('no-store');
	});
});
