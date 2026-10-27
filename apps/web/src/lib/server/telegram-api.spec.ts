import { describe, expect, it } from 'vitest';
import { TelegramApi } from './telegram-api.ts';

const cliente = { ip: '', agente: 'telegram' };
type Pedido = { url: string; metodo: string; headers: Headers; cuerpo: unknown };

/** Go de mentira: responde lo que diga `r` y guarda los pedidos. */
function go(r: (p: Pedido) => Response) {
	const pedidos: Pedido[] = [];
	const f = (async (u: RequestInfo | URL, init?: RequestInit) => {
		const p = {
			url: String(u),
			metodo: init?.method ?? 'GET',
			headers: new Headers(init?.headers),
			cuerpo: init?.body ? JSON.parse(String(init.body)) : undefined
		};
		pedidos.push(p);
		return r(p);
	}) as typeof fetch;
	return { api: new TelegramApi('http://api', 'clave-bff', cliente, f), pedidos };
}

const problema = (status: number, code: string) =>
	new Response(JSON.stringify({ status, code, detail: 'x' }), { status });

describe('TelegramApi', () => {
	it('vincula con la clave del BFF y traduce los rechazos de Go', async () => {
		const respuestas = [
			new Response(JSON.stringify({ tienda: 'La Esquina' }), { status: 200 }),
			problema(409, 'codigo_invalido'),
			problema(409, 'chat_ocupado'),
			problema(429, 'demasiados')
		];
		const { api, pedidos } = go(() => respuestas.shift()!);
		expect(await api.vincular('ABCD2345', 111, 'Lucía')).toEqual({
			ok: true,
			tienda: 'La Esquina'
		});
		expect(pedidos[0]).toMatchObject({
			url: 'http://api/v1/telegram/vincular',
			metodo: 'POST',
			cuerpo: { codigo: 'ABCD2345', chat: 111, nombre: 'Lucía' }
		});
		expect(pedidos[0].headers.get('x-mostrador-bff')).toBe('clave-bff');
		expect(pedidos[0].headers.get('authorization')).toBeNull();
		expect(await api.vincular('ABCD2345', 111, 'L')).toEqual({ ok: false, motivo: 'codigo' });
		expect(await api.vincular('ABCD2345', 111, 'L')).toEqual({ ok: false, motivo: 'ocupado' });
		expect(await api.vincular('ABCD2345', 111, 'L')).toEqual({ ok: false, motivo: 'demasiados' });
	});

	it('contexto y /desvincular: un chat sin vincular es null/false, no un error', async () => {
		const { api } = go(() => problema(404, 'no_encontrado'));
		expect(await api.contexto(5)).toBeNull();
		expect(await api.desvincularChat(5)).toBe(false);
	});

	it('Go caído no se confunde con "sin vincular"', async () => {
		const { api } = go(() => problema(500, 'interno'));
		await expect(api.contexto(5)).rejects.toMatchObject({ estado: 500 });
	});

	it('el panel va con la sesión; desvincular devuelve el chat que estaba', async () => {
		const respuestas = [
			new Response(JSON.stringify({ chat: 333 }), { status: 200 }),
			new Response(null, { status: 204 }),
			new Response(JSON.stringify({ codigo: 'ABCD2345', venceEn: 'x' }), { status: 201 })
		];
		const { api, pedidos } = go(() => respuestas.shift()!);
		expect(await api.desvincular('sesion')).toBe(333);
		expect(await api.desvincular('sesion')).toBeNull();
		expect((await api.codigo('sesion')).codigo).toBe('ABCD2345');
		expect(pedidos.map((p) => `${p.metodo} ${p.url}`)).toEqual([
			'DELETE http://api/v1/panel/telegram',
			'DELETE http://api/v1/panel/telegram',
			'POST http://api/v1/panel/telegram/codigo'
		]);
		expect(pedidos[0].headers.get('authorization')).toBe('Bearer sesion');
	});
});
