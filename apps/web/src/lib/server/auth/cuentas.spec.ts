import { describe, expect, it } from 'vitest';
import { CuentasApi, ErrorCuenta } from './cuentas.ts';

const cliente = { ip: '190.2.3.4', agente: 'Chrome' };

describe('CuentasApi', () => {
	it('manda la clave del BFF y la IP del visitante', async () => {
		let h: Headers | undefined;
		const f = (async (_u: RequestInfo | URL, init?: RequestInit) => {
			h = new Headers(init?.headers);
			return new Response(
				JSON.stringify({ token: 't', venceEn: '2026-11-19T00:00:00Z', tiendaId: 'x' }),
				{ status: 200 }
			);
		}) as typeof fetch;
		const s = await new CuentasApi('http://api', 'clave-bff', cliente, f).ingresar(
			'a@b.co',
			'x',
			true
		);
		expect(s.token).toBe('t');
		expect(h?.get('x-mostrador-bff')).toBe('clave-bff');
		expect(h?.get('x-cliente-ip')).toBe('190.2.3.4');
	});

	it('traduce problem+json a ErrorCuenta con los campos', async () => {
		const f = (async () =>
			new Response(
				JSON.stringify({
					status: 400,
					code: 'invalido',
					detail: 'Revisá',
					issues: [{ path: 'clave', message: 'Muy corta' }]
				}),
				{ status: 400 }
			)) as typeof fetch;
		const err = await new CuentasApi('http://api', 'c', cliente, f)
			.registrar({ nombre: 'L', tienda: 'X', email: 'a@b.co', clave: 'x' })
			.catch((e) => e);
		expect(err).toBeInstanceOf(ErrorCuenta);
		expect(err).toMatchObject({ estado: 400, codigo: 'invalido', campos: { clave: 'Muy corta' } });
	});

	it('202 y 204 no tienen cuerpo', async () => {
		const f = (async () => new Response(null, { status: 202 })) as typeof fetch;
		await expect(
			new CuentasApi('http://api', 'c', cliente, f).recuperar('a@b.co')
		).resolves.toBeUndefined();
	});
});
