import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { PLANES, planCon, planDeTienda, precioAnual, totalAnual } from './planes.ts';

// La misma tabla que la API (apps/api/internal/dominio/planes/testdata/planes.json): lo que la web
// muestra es lo que el servidor aplica.
const api = JSON.parse(
	readFileSync(
		new URL('../../../../api/internal/dominio/planes/testdata/planes.json', import.meta.url),
		'utf8'
	)
) as Array<Record<string, unknown>>;

describe('planes', () => {
	it('igual que la API', () => {
		expect(
			PLANES.map((p) => ({
				id: p.id,
				nombre: p.nombre,
				mensual: p.precio,
				productos: p.productos,
				personas: p.personas,
				fotos: p.fotos,
				fotosMB: p.fotosMB,
				tincho: p.tincho,
				funciones: [...p.funciones]
			}))
		).toEqual(api);
	});
	it('el anual son 10 meses', () => {
		expect(totalAnual(34_900)).toBe(349_000);
		expect(precioAnual(34_900)).toBe(29_083);
	});
	it('la prueba rige como Crece y después vuelve a Gratis', () => {
		const ahora = Date.parse('2026-11-19T12:00:00Z');
		expect(planDeTienda({ plan: 'gratis', creadaEn: '2026-11-10T12:00:00Z' }, ahora).nombre).toBe(
			'Crece'
		);
		expect(planDeTienda({ plan: 'gratis', creadaEn: '2026-10-01T12:00:00Z' }, ahora).id).toBe(
			'gratis'
		);
		expect(planDeTienda({ plan: 'emprende', creadaEn: '2026-11-18T12:00:00Z' }, ahora).id).toBe(
			'emprende'
		);
	});
	it('el plan más barato con cada función', () => {
		expect(planCon('dominio').id).toBe('emprende');
		expect(planCon('mercadolibre').id).toBe('pro');
		expect(planCon('mayorista').id).toBe('negocio');
	});
	it('no promete lo que no existe', () => {
		const todo = PLANES.flatMap((p) => p.incluye).join(' ');
		expect(todo).not.toMatch(/sucursal|API|webhook|Instagram|ilimitad/i);
	});
});
