import { existsSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';
import tabla from './libres.casos.json';
import {
	asignar,
	diaAr,
	horariosDelDia,
	isoAr,
	libres,
	senaDe,
	sumarDias,
	type Consulta
} from './libres.ts';

type Caso = (typeof tabla.casos)[number];
const consulta = (e: Caso['entrada']): Consulta =>
	({ ...e, ahora: Date.parse(e.ahora) }) as unknown as Consulta;

// La tabla es la de Go (apps/api/internal/dominio/turnos/testdata/libres.json); si está en el
// repo, la copia tiene que ser idéntica.
const deGo = join(
	import.meta.dirname,
	'../../../../api/internal/dominio/turnos/testdata/libres.json'
);

describe('libres: la tabla de casos de Go', () => {
	it.runIf(existsSync(deGo))('la copia es idéntica a la de Go', () => {
		expect(tabla).toEqual(JSON.parse(readFileSync(deGo, 'utf8')));
	});
	for (const c of tabla.casos)
		it(c.nombre, () => {
			expect(libres(consulta(c.entrada))).toEqual(c.esperado);
		});
});

describe('libres: fechas, seña y asignación', () => {
	it('el día y la hora salen en UTC−3', () => {
		// 02:30 UTC del 24 son las 23:30 del 23 en Argentina.
		const t = Date.UTC(2026, 9, 24, 2, 30);
		expect(diaAr(t)).toBe('2026-10-23');
		expect(isoAr(t)).toBe('2026-10-23T23:30:00-03:00');
		expect(sumarDias('2026-10-31', 1)).toBe('2026-11-01');
	});

	it('la seña se redondea al peso, como en Go', () => {
		expect(senaDe(18900, 30)).toBe(5670);
		expect(senaDe(12345, 10)).toBe(1235);
		expect(senaDe(12500, 0)).toBe(0);
	});

	it('cualquiera asigna el primero libre por orden; lo ocupado vuelve tachado', () => {
		const base = {
			ajustes: {
				anticipacionMin: 0,
				diasVista: 14,
				horario: [{ dia: 1, desde: '10:00', hasta: '11:00' }]
			},
			servicio: { id: 's', duracion: 30, activo: true },
			profesionales: [
				{ id: 'b', servicios: [], activo: true, orden: 1 },
				{ id: 'a', servicios: [], activo: true, orden: 0 }
			],
			turnos: [
				{
					profesionalId: 'a',
					inicio: '2026-10-26T10:00:00-03:00',
					fin: '2026-10-26T10:30:00-03:00',
					estado: 'confirmado' as const,
					senaPagada: false,
					venceEn: null
				}
			],
			profesionalId: null,
			ahora: Date.parse('2026-10-26T08:00:00-03:00')
		};
		expect(asignar(base, '2026-10-26T10:00:00-03:00')).toBe('b');
		expect(asignar(base, '2026-10-26T10:30:00-03:00')).toBe('a');
		// Fuera de la grilla de 15 minutos, no.
		expect(asignar(base, '2026-10-26T10:05:00-03:00')).toBeNull();
		const soloA = { ...base, profesionalId: 'a', dia: '2026-10-26' };
		expect(
			horariosDelDia(soloA)
				.filter((h) => !h.libre)
				.map((h) => h.hora)
		).toEqual(['10:00', '10:15']);
	});
});
