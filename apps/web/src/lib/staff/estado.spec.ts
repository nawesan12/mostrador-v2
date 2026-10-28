import { describe, expect, it } from 'vitest';
import { abonoMensual, estadoDe, habilesHasta, plazoArrepentimiento } from './estado.ts';

const D = 86_400_000;
const ahora = Date.parse('2026-10-28T15:00:00-03:00');

describe('estado de una tienda', () => {
	it('sigue el mismo orden que Go', () => {
		const base = { plan: 'pro' as const, creadaEn: ahora - 40 * D, ahora };
		expect(estadoDe(base)).toBe('activa');
		expect(estadoDe({ ...base, creadaEn: ahora - 3 * D })).toBe('prueba');
		expect(estadoDe({ ...base, pruebaHasta: ahora + D })).toBe('prueba');
		expect(estadoDe({ ...base, suscripcion: 'paused' })).toBe('cobro_rechazado');
		expect(estadoDe({ ...base, plan: 'gratis', suscripcion: 'paused' })).toBe('activa');
		expect(estadoDe({ ...base, canceladaEn: ahora })).toBe('cancelada');
		expect(estadoDe({ ...base, pausadaEn: ahora, canceladaEn: ahora })).toBe('pausada');
		expect(estadoDe({ ...base, cerradaEn: ahora, pausadaEn: ahora })).toBe('cerrada');
	});
	it('el abono anual tiene el 25 % menos y va en centavos', () => {
		expect(abonoMensual('pro', 'mes')).toBe(4_990_000);
		expect(abonoMensual('pro', 'anio')).toBe(3_742_500);
		expect(abonoMensual('gratis', 'mes')).toBe(0);
	});
});

describe('plazo del arrepentimiento', () => {
	it('son 10 días hábiles, sin sábados ni domingos', () => {
		// Miércoles 28/10 → miércoles 11/11.
		expect(plazoArrepentimiento(ahora)).toBe('2026-11-11');
		// Viernes 30/10 a la noche (hora argentina) → viernes 13/11.
		expect(plazoArrepentimiento(Date.parse('2026-10-30T23:30:00-03:00'))).toBe('2026-11-13');
	});
	it('cuenta los hábiles que faltan', () => {
		expect(habilesHasta('2026-11-11', ahora)).toBe(10);
		expect(habilesHasta('2026-10-28', ahora)).toBe(0);
		expect(habilesHasta('2026-10-26', ahora)).toBe(-2);
	});
});
