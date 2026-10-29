import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';
import casos from '@mostrador/contrato/casos/importar/tiendanube.json';
import { d1Local } from './d1-local.ts';
import { crear, DatosSandbox, dbSandbox } from './sandbox.ts';
import { FUENTES, fuentes } from '../importacion/fuentes.ts';
import { BASE } from '../importar/tiendanube.ts';

const dir = join(import.meta.dirname, '../../../../migraciones-sandbox');
const migraciones = readdirSync(dir)
	.filter((f) => f.endsWith('.sql'))
	.sort()
	.map((f) => readFileSync(join(dir, f), 'utf8'));

let d1: ReturnType<typeof d1Local> | undefined;
const tiendanubeReal = FUENTES.tiendanube;
afterEach(() => {
	d1?.cerrar();
	d1 = undefined;
	FUENTES.tiendanube = tiendanubeReal;
});

/** Cifra de prueba: reversible y sin el texto a la vista. */
const cifra = {
	cifrar: async (s: string) => 'c.' + Buffer.from(s).toString('base64url'),
	descifrar: async (s: string) => Buffer.from(s.slice(2), 'base64url').toString()
};

/** Tiendanube falso: el canje y la API de la tienda 4567, con un 429 en la segunda de clientes. */
function tiendanubeFalso() {
	const pedidos: string[] = [];
	let limite = 1;
	const api = `${BASE}/4567`;
	const f = (async (u: string, init?: RequestInit) => {
		pedidos.push(u);
		if (u === 'https://www.tiendanube.com/apps/authorize/token') {
			const b = JSON.parse(String(init?.body));
			return b.code === 'bueno' && b.client_secret === 'secreto'
				? Response.json({ access_token: 'tok-123', scope: 'read_products', user_id: 4567 })
				: Response.json({ error: 'invalid_grant' }, { status: 400 });
		}
		if ((init?.headers as Record<string, string>)?.authorization !== 'Bearer tok-123')
			return new Response('', { status: 401 });
		const json = (b: unknown, h: Record<string, string> = {}) =>
			Response.json(b, { headers: { 'x-main-language': 'es', ...h } });
		if (u === `${api}/products?page=1&per_page=5`) return json(casos.productos);
		if (u === `${api}/customers?page=1&per_page=50`)
			return json(casos.clientes, { link: `<${api}/customers?page=2&per_page=50>; rel="next"` });
		if (u === `${api}/customers?page=2&per_page=50`) {
			if (limite-- > 0)
				return new Response('', { status: 429, headers: { 'x-rate-limit-reset': '10' } });
			return json([]);
		}
		if (u === `${api}/orders?page=1&per_page=50`) return json(casos.pedidos);
		return new Response('Last page is 0', { status: 404 });
	}) as typeof fetch;
	return { f, pedidos };
}

async function nueva(f: typeof fetch, conApp = true) {
	d1 = d1Local(migraciones);
	const db = dbSandbox(d1);
	const sb = (await crear(db, 'moda'))!;
	FUENTES.tiendanube = fuentes({ fetch: f, dormir: async () => {} }).tiendanube;
	const app = conApp ? { id: '321', secreto: 'secreto', fetch: f } : null;
	return { db, sb, datos: new DatosSandbox(db, sb, { cifra, tiendanube: app }) };
}

describe('Tiendanube en la demo (OAuth real contra un Tiendanube falso)', () => {
	it('sin la app configurada no está disponible', async () => {
		const { datos } = await nueva(tiendanubeFalso().f, false);
		expect(await datos.tiendanubeDisponible()).toBe(false);
		expect(await datos.conectarTiendanube()).toBeNull();
		expect(await datos.volverTiendanube('bueno', 'x')).toBe('sin-tiendanube');
	});

	it('conecta, importa productos, clientes y pedidos y borra el token al terminar', async () => {
		const { f, pedidos } = tiendanubeFalso();
		const { datos } = await nueva(f);
		expect(await datos.tiendanubeDisponible()).toBe(true);
		const url = (await datos.conectarTiendanube())!;
		expect(url).toMatch(/^https:\/\/www\.tiendanube\.com\/apps\/321\/authorize\?state=/);
		const state = new URL(url).searchParams.get('state')!;

		expect(await datos.volverTiendanube('bueno', 'otro-state')).toBe('estado');
		expect(pedidos).toHaveLength(0);
		const imp = await datos.volverTiendanube('bueno', state);
		if (typeof imp === 'string') throw new Error(imp);
		expect(imp).toMatchObject({ plataforma: 'tiendanube', estado: 'pendiente' });
		// De un solo uso.
		expect(await datos.volverTiendanube('bueno', state)).toBe('estado');

		// El token queda cifrado en la fila mientras dura la importación.
		const fila = () =>
			d1!.prepare('select credencial from importaciones').first<string | null>('credencial');
		const guardada = await fila();
		expect(guardada).toBeTruthy();
		expect(guardada).not.toContain('tok-123');

		let a = await datos.importacionActual();
		for (let i = 0; i < 20 && (a?.estado === 'pendiente' || a?.estado === 'corriendo'); i++)
			a = await datos.avanzarImportacion();
		expect(a?.estado).toBe('lista');
		// Los tres productos del caso, los dos clientes que pasan y el único pedido pago en pesos.
		expect(a?.progreso).toMatchObject({ productos: 3, clientes: 2, pedidos: 1 });
		expect(await fila()).toBeNull();
		// El 429 se reintentó en el mismo paso y el Link de clientes se siguió.
		expect(pedidos.filter((u) => u.includes('customers?page=2'))).toHaveLength(2);
		expect((await datos.historialCliente('lucia.gomez@gmail.com'))[0]).toMatchObject({
			plataforma: 'tiendanube',
			numero: '1203',
			total: 34800
		});
	});

	it('un código que Tiendanube rechaza no crea nada; una tienda desconectada la deja fallida', async () => {
		const { f } = tiendanubeFalso();
		const { datos } = await nueva(f);
		const s1 = new URL((await datos.conectarTiendanube())!).searchParams.get('state')!;
		expect(await datos.volverTiendanube('malo', s1)).toBe('canje');
		expect(await datos.importacionActual()).toBeNull();

		// Un token que ya no sirve (401): fallida, con el aviso de reconectar y sin credencial.
		const imp = await datos.crearImportacion({
			plataforma: 'tiendanube',
			credencial: JSON.stringify({ token: 'viejo', tienda: '4567' })
		});
		expect(typeof imp).toBe('object');
		const a = await datos.avanzarImportacion();
		expect(a?.estado).toBe('fallida');
		expect(a?.errores.at(-1)?.detalle).toMatch(/conectala de nuevo/);
		expect(
			await d1!.prepare('select credencial from importaciones').first<string | null>('credencial')
		).toBeNull();
	});
});
