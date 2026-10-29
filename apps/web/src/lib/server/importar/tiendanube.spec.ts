import { describe, expect, it } from 'vitest';
import casos from '@mostrador/contrato/casos/importar/tiendanube.json';
import esperado from '@mostrador/contrato/casos/importar/tiendanube.esperado.json';
import { conCabeceras, fetchFalso } from './fetch-falso.ts';
import {
	BASE,
	canjear,
	ErrorLimite,
	ErrorReconectar,
	fechaTN,
	leer,
	primera,
	resultadoTiendanube,
	urlAutorizar,
	type ApiTN,
	type ProductoTN
} from './tiendanube.ts';

describe('importar: Tiendanube (mismos casos que Go)', () => {
	it('productos, clientes y pedidos de la API → lo que guarda Tu Local Web', () => {
		expect(
			resultadoTiendanube(casos as unknown as Parameters<typeof resultadoTiendanube>[0])
		).toEqual(esperado);
	});

	it('lee las fechas con la zona con o sin dos puntos', () => {
		expect(fechaTN('2022-11-15T19:36:59+0000')).toBe(Date.UTC(2022, 10, 15, 19, 36, 59));
		expect(fechaTN('2026-03-02T14:10:00-03:00')).toBe(Date.UTC(2026, 2, 2, 17, 10));
		expect(fechaTN('2026-02-30T10:00:00Z')).toBeNull();
		expect(fechaTN('ayer')).toBeNull();
	});

	it('un producto sin variantes no entra', () => {
		const p = { id: 1, name: { es: 'Algo' }, variants: [] } satisfies ProductoTN;
		expect(resultadoTiendanube({ productos: [p], clientes: [], pedidos: [] }).productos).toEqual(
			[]
		);
	});
});

describe('Tiendanube: OAuth y API', () => {
	const app = { id: '321', secreto: 's3' };

	it('arma la URL de autorización y canjea el código', async () => {
		expect(urlAutorizar(app, 'abc')).toBe(
			'https://www.tiendanube.com/apps/321/authorize?state=abc'
		);
		const cuerpos: unknown[] = [];
		const f = (async (_u: string, init: RequestInit) => {
			cuerpos.push(JSON.parse(String(init.body)));
			return Response.json({ access_token: 'tok', scope: 'read_products', user_id: 4567 });
		}) as typeof fetch;
		expect(await canjear({ ...app, fetch: f }, 'cod')).toEqual({
			acceso: 'tok',
			tiendaId: '4567',
			alcance: ['read_products']
		});
		expect(cuerpos[0]).toMatchObject({
			client_id: '321',
			client_secret: 's3',
			grant_type: 'authorization_code',
			code: 'cod'
		});
		const malo = (async () => Response.json({ error: 'x' }, { status: 400 })) as typeof fetch;
		expect(await canjear({ ...app, fetch: malo }, 'cod')).toBeNull();
	});

	it('sigue el header Link sólo dentro de la API de la tienda', async () => {
		const api: ApiTN = { token: 't', tiendaId: '9', fetch: undefined };
		const u = primera(api, 'products', 5);
		expect(u).toBe(`${BASE}/9/products?page=1&per_page=5`);
		const pedidos: string[] = [];
		const sig = `${BASE}/9/products?page=2&per_page=5`;
		api.fetch = fetchFalso(
			{
				'/2025-03/9/products?page=1&per_page=5': conCabeceras([{ id: 1 }], {
					link: `<${sig}>; rel="next", <${BASE}/9/products?page=9>; rel="last"`,
					'x-main-language': 'pt',
					'x-total-count': '6'
				}),
				'/2025-03/9/products?page=2&per_page=5': conCabeceras([{ id: 2 }], {
					link: '<https://otro.com/9/products?page=3>; rel="next"'
				})
			},
			pedidos
		);
		expect(await leer(api, u)).toEqual({
			registros: [{ id: 1 }],
			siguiente: sig,
			idioma: 'pt',
			total: 6
		});
		expect((await leer(api, sig)).siguiente).toBe('');
		// Pasada la última, 404: no hay más.
		expect((await leer(api, `${BASE}/9/products?page=3`)).registros).toEqual([]);
		await expect(leer(api, 'https://otro.com/9/products')).rejects.toThrow();
		await expect(leer(api, `${BASE}/10/products`)).rejects.toThrow();
		expect(pedidos).toHaveLength(3);
	});

	it('429 espera lo que dice la API y reintenta; 401 pide reconectar', async () => {
		const esperas: number[] = [];
		let n = 0;
		const api: ApiTN = {
			token: 't',
			tiendaId: '9',
			esperaMax: 3000,
			reintentos: 2,
			dormir: async (ms) => void esperas.push(ms),
			fetch: (async () =>
				++n < 3
					? new Response('', { status: 429, headers: { 'x-rate-limit-reset': String(n * 2000) } })
					: Response.json([])) as unknown as typeof fetch
		};
		expect((await leer(api, primera(api, 'orders', 50))).registros).toEqual([]);
		expect(esperas).toEqual([2000, 3000]);
		n = -10;
		await expect(leer(api, primera(api, 'orders', 50))).rejects.toBeInstanceOf(ErrorLimite);
		api.fetch = (async () => new Response('', { status: 401 })) as unknown as typeof fetch;
		await expect(leer(api, primera(api, 'orders', 50))).rejects.toBeInstanceOf(ErrorReconectar);
	});
});
