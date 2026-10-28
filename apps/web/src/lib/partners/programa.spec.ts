import { describe, expect, it } from 'vitest';
import casos from '../../../../api/internal/dominio/partners/testdata/comisiones.json';
import {
	comision,
	cuitValido,
	estado,
	fechaDePago,
	finComision,
	nivelEfectivo,
	resumen,
	type PanelPartner
} from './programa.ts';

// Los mismos casos que Go (apps/api/internal/dominio/partners/testdata/comisiones.json).
describe('comisión de partners', () => {
	for (const c of casos.comisiones)
		it(c.caso, () => expect(comision(c.nivel, c.plan, c.ciclo, c.alta, c.mes)).toBe(c.comision));

	for (const c of casos.niveles)
		it(`${c.asignado} con ${c.activas} activas es ${c.nivel}`, () =>
			expect(nivelEfectivo(c.asignado, c.activas)).toBe(c.nivel));

	it('el ejemplo de la landing: 12 tiendas Pro en Plata = $ 149.700', () =>
		expect(12 * comision('plata', 'pro', 'mes', '2026-03-10T15:00:00Z', '2026-03')).toBe(149_700));

	it('fechas: fin de la comisión y día de pago', () => {
		expect(finComision('2026-01-01T02:00:00Z')).toBe('2026-12-01');
		expect(finComision('2026-03-15T15:00:00Z')).toBe('2027-03-01');
		expect(fechaDePago('2026-10')).toBe('2026-11-10');
		expect(fechaDePago('2026-12')).toBe('2027-01-10');
	});

	it('estado de la tienda', () => {
		const ahora = Date.parse('2026-10-26T15:00:00Z');
		const hace = (d: number) => new Date(ahora - d * 86_400_000).toISOString();
		expect(estado('pro', hace(3), false, ahora)).toBe('prueba');
		expect(estado('pro', hace(15), false, ahora)).toBe('activa');
		expect(estado('gratis', hace(15), false, ahora)).toBe('gratis');
		expect(estado('pro', hace(200), true, ahora)).toBe('baja');
	});

	it('CUIT con verificador', () => {
		expect(cuitValido('20123456786')).toBe(true);
		expect(cuitValido('20123456780')).toBe(false);
		expect(cuitValido('2012345678')).toBe(false);
	});
});

describe('resumen del inicio', () => {
	const base: PanelPartner = {
		hoy: '2026-10-26',
		partner: {
			nombre: 'Ana',
			email: 'ana@pampacode.com.ar',
			nivel: 'partner',
			nivelEfectivo: 'plata',
			porcentaje: 25,
			codigo: 'pampa-code',
			cupon: 'PAMPACODE20',
			aprobadoEn: '2026-01-10',
			perfil: {
				nombre: 'Pampa Code',
				ciudad: '',
				descripcion: '',
				especialidades: [],
				casos: [],
				email: '',
				whatsapp: '',
				web: '',
				logo: ''
			},
			cobro: null
		},
		tiendas: [
			...Array.from({ length: 6 }, (_, i) => ({
				id: `a${i}`,
				nombre: `Activa ${i}`,
				slug: `activa-${i}`,
				plan: 'pro' as const,
				ciclo: 'mes' as const,
				estado: 'activa' as const,
				alta: '2026-05-10',
				finComision: i === 0 ? '2026-11-01' : '2027-05-01',
				comisionMes: 12_475
			})),
			{
				id: 'p',
				nombre: 'En prueba',
				slug: 'en-prueba',
				plan: 'gratis',
				ciclo: 'mes',
				estado: 'prueba',
				alta: '2026-10-15',
				finComision: '2027-10-01',
				comisionMes: 0
			},
			{
				id: 'b',
				nombre: 'De baja',
				slug: 'de-baja',
				plan: 'pro',
				ciclo: 'mes',
				estado: 'baja',
				alta: '2026-02-10',
				finComision: '2027-02-01',
				comisionMes: 0
			}
		],
		clics: [],
		pagos: [
			{
				mes: '2026-09',
				estado: 'aprobada',
				monto: 74_850,
				detalle: [],
				facturaUrl: null,
				fechaPago: '2026-10-10',
				pagadoEn: null
			},
			{
				mes: '2026-08',
				estado: 'pagada',
				monto: 62_375,
				detalle: [],
				facturaUrl: 'https://drive.google.com/f',
				fechaPago: '2026-09-10',
				pagadoEn: '2026-09-10'
			}
		],
		colaboraciones: []
	};

	it('suma, cuenta y arma el progreso a Oro', () => {
		const r = resumen(base);
		expect(r.estimado).toBe(6 * 12_475);
		expect([r.activas, r.prueba, r.perdidas]).toEqual([6, 1, 1]);
		expect(r.proximo).toMatchObject({
			monto: 74_850,
			cuando: 'se paga cuando subas tu factura',
			estado: 'aprobada'
		});
		expect(r.progreso).toMatchObject({ faltan: 14, pct: 30 });
		expect(r.progreso?.siguiente.id).toBe('oro');
	});

	it('tareas: prueba que vence, cobro, factura, fin de comisión y perfil', () => {
		expect(resumen(base).tareas.map((t) => t.id)).toEqual([
			'prueba',
			'cobro',
			'factura-2026-09',
			'fin',
			'perfil'
		]);
	});

	it('sin pagos abiertos, el próximo es el estimado del mes', () => {
		const r = resumen({ ...base, pagos: [base.pagos[1]] });
		expect(r.proximo).toMatchObject({
			estado: 'estimado',
			cuando: `se paga el 10 de${String.fromCharCode(160)}noviembre`,
			monto: 6 * 12_475
		});
	});
});
