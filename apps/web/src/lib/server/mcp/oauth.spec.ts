import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';
import { b64url, sha256 } from '../crypto.ts';
import { golpe } from '../limite.ts';
import { d1Local } from '../sandbox/d1-local.ts';
import { crear, dbSandbox, purgar, VIDA_MS } from '../sandbox/sandbox.ts';
import {
	ACCESO_MS,
	autenticar,
	canjearCodigo,
	CODIGO_MS,
	conexionesDe,
	desconectar,
	emitirCodigo,
	permisosPedidos,
	pkceValido,
	purgarOauth,
	registrarCliente,
	renovar,
	revocar
} from './oauth.ts';

const dir = join(import.meta.dirname, '../../../../migraciones-sandbox');
const migraciones = readdirSync(dir)
	.filter((f) => f.endsWith('.sql'))
	.sort()
	.map((f) => readFileSync(join(dir, f), 'utf8'));

let d1: ReturnType<typeof d1Local> | undefined;
afterEach(() => {
	d1?.cerrar();
	d1 = undefined;
});

const T0 = Date.UTC(2026, 9, 23, 12);
const VUELTA = 'https://claude.ai/api/mcp/auth_callback';
const VERIFICADOR = 'v'.repeat(20) + '-verificador-de-prueba-pkce-123';
const desafio = async (v: string) =>
	b64url(await crypto.subtle.digest('SHA-256', new TextEncoder().encode(v)));

/** Un sandbox, un cliente registrado y un código emitido. */
async function armar(
	permisos: ('lectura' | 'acciones')[] = ['lectura', 'acciones'],
	sandboxCreado = T0
) {
	d1 = d1Local(migraciones);
	const db = dbSandbox(d1);
	const sb = (await crear(db, 'moda', sandboxCreado))!;
	const c = (await registrarCliente(db, { client_name: 'Claude', redirect_uris: [VUELTA] }, T0))!;
	const codigo = await emitirCodigo(
		db,
		{
			clienteId: c.client_id,
			sandboxId: sb.id,
			redireccion: VUELTA,
			desafio: await desafio(VERIFICADOR),
			permisos
		},
		T0
	);
	const canje = (o: Partial<Parameters<typeof canjearCodigo>[1]> = {}, ahora = T0 + 1000) =>
		canjearCodigo(
			db,
			{ codigo, clienteId: c.client_id, redireccion: VUELTA, verificador: VERIFICADOR, ...o },
			ahora
		);
	return { db, sb, c, codigo, canje };
}

describe('registro dinámico', () => {
	it('acepta https y localhost, rechaza el resto', async () => {
		d1 = d1Local(migraciones);
		const db = dbSandbox(d1);
		expect(await registrarCliente(db, { redirect_uris: ['http://evil.com/cb'] })).toBeNull();
		expect(await registrarCliente(db, { redirect_uris: ['javascript:alert(1)'] })).toBeNull();
		expect(await registrarCliente(db, {})).toBeNull();
		const c = await registrarCliente(db, {
			redirect_uris: ['http://localhost:3000/cb', 'http://evil.com/cb']
		});
		expect(c?.redirect_uris).toEqual(['http://localhost:3000/cb']);
		expect(c?.client_name).toBe('Otra IA (MCP)');
		expect(c?.token_endpoint_auth_method).toBe('none');
	});
});

describe('PKCE S256', () => {
	it('valida el verificador contra el desafío', async () => {
		const d = await desafio(VERIFICADOR);
		expect(await pkceValido(VERIFICADOR, d)).toBe(true);
		expect(await pkceValido(VERIFICADOR + 'x', d)).toBe(false);
		// Muy corto (RFC 7636: de 43 a 128 caracteres).
		expect(await pkceValido('corto', await desafio('corto'))).toBe(false);
	});

	it('sin el verificador correcto el código no sirve', async () => {
		const { canje } = await armar();
		expect(await canje({ verificador: 'x'.repeat(50) })).toBeNull();
	});
});

describe('código de autorización', () => {
	it('se canjea una sola vez', async () => {
		const { canje } = await armar();
		const t = await canje();
		expect(t?.access_token).toMatch(/^mostrador_mcp_/);
		expect(t?.refresh_token).toMatch(/^mostrador_rt_/);
		expect(t?.expires_in).toBe(ACCESO_MS / 1000);
		expect(t?.scope).toBe('lectura acciones');
		expect(await canje()).toBeNull();
	});

	it('vence a los 5 minutos', async () => {
		const { canje } = await armar();
		expect(await canje({}, T0 + CODIGO_MS + 1)).toBeNull();
	});

	it('otro cliente u otra vuelta no lo pueden canjear (y queda gastado)', async () => {
		const { canje } = await armar();
		expect(await canje({ clienteId: 'mcp_otro' })).toBeNull();
		expect(await canje()).toBeNull();
	});

	it('los tokens se guardan hasheados, nunca en claro', async () => {
		const { canje, codigo } = await armar();
		const t = (await canje())!;
		const filas = JSON.stringify(
			await d1!
				.prepare('select * from oauth_conexiones')
				.all()
				.then((r) => r.results)
		);
		expect(filas).not.toContain(t.access_token);
		expect(filas).not.toContain(t.refresh_token);
		expect(filas).toContain(await sha256(t.access_token));
		expect(filas).not.toContain(codigo);
	});
});

describe('access token', () => {
	it('autentica, respeta los permisos y vence a la hora', async () => {
		const { db, sb, canje } = await armar(['lectura']);
		const t = (await canje())!;
		const a = await autenticar(db, `Bearer ${t.access_token}`, T0 + 2000);
		expect(a).toMatchObject({ sandboxId: sb.id, nombre: 'Claude', permisos: ['lectura'] });
		expect(await autenticar(db, `Bearer ${t.access_token}`, T0 + ACCESO_MS + 1001)).toBeNull();
		expect(await autenticar(db, `Bearer ${t.refresh_token}`, T0 + 2000)).toBeNull();
		expect(await autenticar(db, null)).toBeNull();
	});

	it('cuenta los pedidos por minuto', async () => {
		const { db, canje } = await armar();
		const t = (await canje())!;
		const h = `Bearer ${t.access_token}`;
		expect((await autenticar(db, h, T0 + 2000))?.pedidos).toBe(1);
		expect((await autenticar(db, h, T0 + 3000))?.pedidos).toBe(2);
		// Pasado el minuto, la ventana arranca de nuevo.
		expect((await autenticar(db, h, T0 + 70_000))?.pedidos).toBe(1);
	});

	it('vence con el sandbox, aunque el token sea nuevo', async () => {
		// El sandbox vence un minuto después del canje.
		const { db, c, canje } = await armar(undefined, T0 + 1000 + 60_000 - VIDA_MS);
		const fin = T0 + 1000 + 60_000;
		const t = (await canje())!;
		expect(t.expires_in).toBe(60);
		expect(await autenticar(db, `Bearer ${t.access_token}`, fin + 1)).toBeNull();
		expect(
			await renovar(db, { refresh: t.refresh_token, clienteId: c.client_id }, fin + 1)
		).toBeNull();
	});
});

describe('refresh token', () => {
	it('rota: el viejo deja de servir, y el access token anterior también', async () => {
		const { db, c, canje } = await armar();
		const t1 = (await canje())!;
		const t2 = (await renovar(
			db,
			{ refresh: t1.refresh_token, clienteId: c.client_id },
			T0 + 5000
		))!;
		expect(t2.refresh_token).not.toBe(t1.refresh_token);
		expect(await autenticar(db, `Bearer ${t1.access_token}`, T0 + 6000)).toBeNull();
		expect(await autenticar(db, `Bearer ${t2.access_token}`, T0 + 6000)).not.toBeNull();
		const t3 = await renovar(db, { refresh: t2.refresh_token, clienteId: c.client_id }, T0 + 7000);
		expect(t3).not.toBeNull();
	});

	it('reusar uno ya rotado corta la conexión', async () => {
		const { db, sb, c, canje } = await armar();
		const t1 = (await canje())!;
		const t2 = (await renovar(
			db,
			{ refresh: t1.refresh_token, clienteId: c.client_id },
			T0 + 5000
		))!;
		expect(await renovar(db, { refresh: t1.refresh_token, clienteId: c.client_id })).toBeNull();
		expect(await autenticar(db, `Bearer ${t2.access_token}`, T0 + 6000)).toBeNull();
		expect(await conexionesDe(db, sb.id, T0 + 6000)).toHaveLength(0);
	});

	it('sólo lo renueva su cliente', async () => {
		const { db, canje } = await armar();
		const t = (await canje())!;
		expect(await renovar(db, { refresh: t.refresh_token, clienteId: 'mcp_otro' })).toBeNull();
	});
});

describe('revocar y desconectar', () => {
	it('revocar con el access o el refresh token corta la conexión', async () => {
		for (const cual of ['access_token', 'refresh_token'] as const) {
			const { db, sb, canje } = await armar();
			const t = (await canje())!;
			await revocar(db, t[cual]);
			expect(await autenticar(db, `Bearer ${t.access_token}`, T0 + 2000)).toBeNull();
			expect(await conexionesDe(db, sb.id, T0 + 2000)).toHaveLength(0);
			d1?.cerrar();
			d1 = undefined;
		}
	});

	it('desde el panel, sólo una conexión del propio sandbox', async () => {
		const { db, sb, canje } = await armar();
		await canje();
		const [cx] = await conexionesDe(db, sb.id, T0 + 2000);
		expect(cx.nombre).toBe('Claude');
		expect(await desconectar(db, 'otro-sandbox', cx.id)).toBe(false);
		expect(await desconectar(db, sb.id, cx.id)).toBe(true);
		expect(await conexionesDe(db, sb.id, T0 + 2000)).toHaveLength(0);
	});

	it('la purga del sandbox se lleva sus conexiones y códigos; la del cron, los clientes viejos', async () => {
		const { db, sb, canje } = await armar();
		await canje();
		await purgar(db, T0 + VIDA_MS + 1);
		expect(await conexionesDe(db, sb.id, T0)).toHaveLength(0);
		await purgarOauth(db, T0 + 31 * 86_400_000);
		const n = await d1!.prepare('select count(*) n from oauth_clientes').first<number>('n');
		expect(n).toBe(0);
	});

	it('el cron no borra un cliente registrado hace mucho que se sigue usando', async () => {
		const { db, sb, c } = await armar();
		const DIA = 86_400_000;
		const pedirCodigo = (ahora: number) =>
			emitirCodigo(
				db,
				{
					clienteId: c.client_id,
					sandboxId: sb.id,
					redireccion: VUELTA,
					desafio: 'x'.repeat(43),
					permisos: ['lectura']
				},
				ahora
			);
		const clientes = () => d1!.prepare('select count(*) n from oauth_clientes').first<number>('n');
		await pedirCodigo(T0 + 20 * DIA);
		await purgarOauth(db, T0 + 31 * DIA);
		expect(await clientes()).toBe(1);
		await purgarOauth(db, T0 + 51 * DIA);
		expect(await clientes()).toBe(0);
	});
});

describe('permisos y límites', () => {
	it('acciones sólo si se piden (o si no se pide nada)', () => {
		expect(permisosPedidos(null)).toEqual(['lectura', 'acciones']);
		expect(permisosPedidos('lectura')).toEqual(['lectura']);
		expect(permisosPedidos('lectura acciones')).toEqual(['lectura', 'acciones']);
	});

	it('golpe cuenta por ventana', async () => {
		d1 = d1Local(migraciones);
		const db = dbSandbox(d1);
		expect(await golpe(db, 'x', 2, 60, T0)).toBe(true);
		expect(await golpe(db, 'x', 2, 60, T0)).toBe(true);
		expect(await golpe(db, 'x', 2, 60, T0)).toBe(false);
		expect(await golpe(db, 'x', 2, 60, T0 + 61_000)).toBe(true);
	});
});
