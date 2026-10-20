import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';
import { d1Local } from '../sandbox/d1-local.ts';
import { leerEstado, porcentaje, ultimosDias } from './lectura.ts';
import { chequear, chequeos, registrar, type Resultado } from './monitor.ts';

const dir = join(import.meta.dirname, '../../../../migraciones-estado');
const migraciones = readdirSync(dir)
	.filter((f) => f.endsWith('.sql'))
	.sort()
	.map((f) => readFileSync(join(dir, f), 'utf8'));

let d1: ReturnType<typeof d1Local> | undefined;
const nueva = () => (d1 = d1Local(migraciones)) as unknown as D1Database;
afterEach(() => {
	d1?.cerrar();
	d1 = undefined;
});

const HORA = 3_600_000;
// 20/10/2026 12:00 en Buenos Aires.
const T0 = Date.parse('2026-10-20T15:00:00Z');
const res = (componente: Resultado['componente'], ok: boolean): Resultado => ({
	componente,
	ok,
	latencia: 120,
	detalle: ok ? 'HTTP 200' : 'HTTP 503'
});

describe('chequeos', () => {
	it('sin API_URL no chequea la API', () => {
		expect(chequeos('https://x', '').map((c) => c.componente)).not.toContain('api');
		expect(chequeos('https://x', 'https://api').map((c) => c.componente)).toContain('api');
	});

	it('el sitio tiene que dar 200; un proveedor que contesta 401 está arriba', async () => {
		const r = await chequear(
			[
				{ componente: 'tiendas', url: 'https://x/demo', sitio: true, exige200: true },
				{ componente: 'pagos', url: 'https://mp' },
				{ componente: 'arca', url: 'https://arca', exige200: true }
			],
			{
				sitio: async () => new Response('', { status: 302 }),
				externo: async (url) =>
					url === 'https://mp'
						? new Response('', { status: 401 })
						: Promise.reject(new Error('dns'))
			}
		);
		expect(r.map((x) => [x.componente, x.ok, x.detalle])).toEqual([
			['tiendas', false, 'HTTP 302'],
			['pagos', true, 'HTTP 401'],
			['arca', false, 'No se pudo conectar']
		]);
	});
});

describe('registro e incidentes', () => {
	it('abre un incidente con dos fallas seguidas y lo cierra cuando vuelve', async () => {
		const db = nueva();
		expect(await registrar(db, [res('arca', false)], T0)).toEqual([]);
		const abre = await registrar(db, [res('arca', false)], T0 + HORA);
		expect(abre).toMatchObject([{ tipo: 'abre', componente: 'arca' }]);
		// Una tercera falla no abre otro.
		expect(await registrar(db, [res('arca', false)], T0 + 2 * HORA)).toEqual([]);
		const cierra = await registrar(db, [res('arca', true)], T0 + 3 * HORA);
		expect(cierra[0].tipo).toBe('cierra');
		expect(cierra[0].detalle).toContain('después de 120 minutos');

		const e = await leerEstado(db, 7, T0 + 3 * HORA);
		const arca = e.componentes.find((c) => c.id === 'arca')!;
		expect(arca.ok).toBe(true);
		expect(arca.barras.at(-1)).toEqual({ dia: '2026-10-20', estado: 'caida' });
		expect(arca.disponibilidad).toBe(25);
		expect(e.incidentes).toHaveLength(1);
		expect(e.incidentes[0].fin).not.toBeNull();
	});

	it('los días sin chequeos quedan sin datos y la API sin chequear no tiene estado', async () => {
		const db = nueva();
		await registrar(db, [res('tiendas', true)], T0);
		const e = await leerEstado(db, 30, T0);
		const t = e.componentes.find((c) => c.id === 'tiendas')!;
		expect(t.barras).toHaveLength(30);
		expect(t.barras.filter((b) => b.estado === 'sin')).toHaveLength(29);
		expect(t.disponibilidad).toBe(100);
		expect(e.componentes.find((c) => c.id === 'api')).toMatchObject({
			ok: null,
			disponibilidad: null
		});
		expect(e.actualizado).toBe(T0);
	});
});

describe('formato', () => {
	it('porcentajes y días', () => {
		expect(porcentaje(99.9612)).toBe('99,96\u00a0%');
		expect(porcentaje(100)).toBe('100\u00a0%');
		expect(porcentaje(null)).toBe('—');
		const d = ultimosDias(3, T0);
		expect(d).toEqual(['2026-10-18', '2026-10-19', '2026-10-20']);
	});
});
