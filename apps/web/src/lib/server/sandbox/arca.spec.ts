import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { afterEach, describe, expect, it } from 'vitest';
import { cuitValido } from '#lib/panel/cuit.ts';
import { d1Local } from './d1-local.ts';
import { alCobrarEn, crear, DatosSandbox, dbSandbox, SIN_FACTURAR_AL_COBRAR } from './sandbox.ts';

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
async function demo() {
	d1 = d1Local(migraciones);
	const db = dbSandbox(d1);
	const sb = (await crear(db, 'moda', Date.now() - 3_600_000))!;
	return new DatosSandbox(db, sb);
}

type Fila = { id: string; estado: string; numero: string; motivo: string; tipo: string };
const filas = async (d: DatosSandbox) =>
	(await d.listar('facturacion', { cuantos: 5000 })).filas as unknown as Fila[];

const ALTA = {
	cuit: '20-44859234-1',
	razonSocial: 'Nahuel Santillán',
	condicion: 'monotributo',
	puntoVenta: 4,
	ambiente: 'homologacion'
} as const;

describe('ARCA en la demo', () => {
	it('arranca conectada en homologación, con un CUIT válido y facturar al cobrar', async () => {
		const d = await demo();
		const a = await d.arca();
		expect(a).toMatchObject({
			conectada: true,
			certificado: true,
			ambiente: 'homologacion',
			condicion: 'responsable_inscripto',
			facturarAlCobrar: true
		});
		expect(cuitValido(a.cuit)).toBe(true);
		expect(Date.parse(a.certificadoVence!)).toBeGreaterThan(Date.now());
		expect((await d.integraciones()).conectadas.arca).toBe(true);
		// No se desconecta con el interruptor genérico: tiene su asistente.
		expect(await d.ajustarIntegracion({ tipo: 'conexion', id: 'arca', valor: false })).toBe(
			'no-se-puede'
		);
	});

	it('el pedido valida como Go y lo arma el navegador; certificado y prueba no mienten', async () => {
		const d = await demo();
		expect(await d.arcaPedido({ ...ALTA, cuit: '20448592342' })).toEqual({
			error: expect.stringContaining('CUIT')
		});
		expect(await d.arcaPedido({ ...ALTA, puntoVenta: 0 })).toEqual({
			error: expect.stringContaining('punto de venta')
		});
		expect(await d.arcaPedido(ALTA)).toEqual({ demo: true });
		expect(await d.arcaCertificado('hola')).toEqual({
			error: expect.stringContaining('no es un certificado')
		});
		const cert = '-----BEGIN CERTIFICATE-----\nMIIB\n-----END CERTIFICATE-----';
		expect(await d.arcaCertificado(cert)).toEqual({
			error: expect.stringContaining('tienda de ejemplo')
		});
		expect(await d.arcaProbar()).toMatchObject({ ok: false });
		// Sigue conectada: el visitante no rompe la demo.
		expect((await d.arca()).conectada).toBe(true);
	});

	it('ajustes: punto de venta y facturar al cobrar con su historial', async () => {
		const d = await demo();
		expect(await d.arcaAjustes({ puntoVenta: 0 })).toEqual({ error: expect.any(String) });
		expect(await d.arcaAjustes({ puntoVenta: 7 })).toBeNull();
		expect(await d.arcaAjustes({ facturarAlCobrar: false })).toBeNull();
		const a = await d.arca();
		expect(a.puntoVenta).toBe(7);
		expect(a.facturarAlCobrar).toBe(false);
		expect((await d.integraciones()).opciones.arca.auto).toBe(false);
		// Las ventas de antes del cambio siguen facturadas.
		expect((await filas(d)).some((f) => f.estado === 'nofac')).toBe(false);
	});

	it('anular con nota de crédito: la factura queda anulada y aparece la nota', async () => {
		const d = await demo();
		const ok = (await filas(d)).find((f) => f.estado === 'ok')!;
		expect(await d.accionLista('facturacion', [ok.id], 'anular')).toBe(1);
		const despues = await filas(d);
		expect(despues.find((f) => f.id === ok.id)!.estado).toBe('void');
		const nota = despues.find((f) => f.id === `NC-${ok.id}`)!;
		expect(nota).toMatchObject({ estado: 'ok', tipo: ok.tipo });
		expect(nota.numero.startsWith('NC ')).toBe(true);
		// Una sola vez, y una nota no se anula.
		expect(await d.accionLista('facturacion', [ok.id], 'anular')).toBe(0);
		expect(await d.accionLista('facturacion', [nota.id], 'anular')).toBe(0);
	});

	it('emitir: A sólo con un CUIT válido', async () => {
		const d = await demo();
		const v = { cliente: 'Estudio Paz', tipo: 'A', concepto: '', total: 5000 };
		expect(await d.accionLista('facturacion', [], 'emitir', { ...v, cuit: '20448592342' })).toBe(0);
		expect(await d.accionLista('facturacion', [], 'emitir', { ...v, cuit: '20-44859234-1' })).toBe(
			1
		);
	});

	it('alCobrarEn sigue el historial y arranca prendido', () => {
		const h: [number, boolean][] = [
			[100, false],
			[200, true]
		];
		expect([50, 100, 150, 200, 250].map((t) => alCobrarEn(h, t))).toEqual([
			true,
			false,
			false,
			true,
			true
		]);
		expect(SIN_FACTURAR_AL_COBRAR).toContain('Facturar al cobrar');
	});
});
