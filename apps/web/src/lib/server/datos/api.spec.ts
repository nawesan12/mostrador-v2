import { describe, expect, it } from 'vitest';
import { DatosApi } from './api.ts';

const respuesta = (status: number, cuerpo: unknown) =>
	(async () => new Response(JSON.stringify(cuerpo), { status })) as typeof fetch;

describe('DatosApi', () => {
	it('manda el token y devuelve el JSON', async () => {
		let auth: string | null = null;
		const f = (async (_: RequestInfo | URL, init?: RequestInit) => {
			auth = new Headers(init?.headers).get('authorization');
			return new Response('[{"id":"moda"}]');
		}) as typeof fetch;
		const r = await new DatosApi('http://api', 'tok', f).rubros();
		expect(r).toEqual([{ id: 'moda' }]);
		expect(auth).toBe('Bearer tok');
	});

	it('un problem+json se convierte en error de SvelteKit con su detalle', async () => {
		const api = new DatosApi(
			'http://api',
			undefined,
			respuesta(401, {
				title: 'Tenés que ingresar',
				status: 401,
				detail: 'Tu sesión venció.',
				code: 'sin_sesion'
			})
		);
		await expect(api.tienda()).rejects.toMatchObject({
			status: 401,
			body: { message: 'Tu sesión venció.', code: 'sin_sesion' }
		});
	});

	it('ARCA: los errores de Go vuelven como texto para mostrar', async () => {
		const api = new DatosApi(
			'http://api',
			'tok',
			respuesta(409, {
				status: 409,
				detail: 'Primero conectá ARCA con tu certificado.',
				code: 'sin_arca'
			})
		);
		expect(await api.arcaProbar()).toEqual({
			ok: false,
			ultimo: 0,
			detalle: 'Primero conectá ARCA con tu certificado.'
		});
		expect(await api.arcaCertificado('x')).toEqual({
			error: 'Primero conectá ARCA con tu certificado.'
		});
		const ok = new DatosApi('http://api', 'tok', respuesta(200, { pedido: '-----BEGIN' }));
		expect(
			await ok.arcaPedido({
				cuit: '20448592341',
				razonSocial: 'X',
				condicion: 'monotributo',
				puntoVenta: 1,
				ambiente: 'homologacion'
			})
		).toEqual({ pedido: '-----BEGIN' });
		const nada = (async () => new Response(null, { status: 204 })) as typeof fetch;
		expect(await new DatosApi('http://api', 'tok', nada).arcaAjustes({ puntoVenta: 2 })).toBeNull();
	});
});
