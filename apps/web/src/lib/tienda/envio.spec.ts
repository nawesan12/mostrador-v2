import { describe, expect, it } from 'vitest';
import { cotizar } from './envio.ts';

const t = {
	envios: { gratisDesde: 100000, retiro: true, moto: true, correo: true },
	direccion: 'Belgrano 742'
};
const lunes = new Date('2026-10-26T12:00:00-03:00');

describe('cotizar', () => {
	it('por región, con moto sólo en CABA y retiro siempre', () => {
		const o = cotizar('1414', 5000, t, lunes);
		expect(o.map((x) => x.id)).toEqual(['andreani', 'correo', 'moto', 'retiro']);
		expect(o[0]).toMatchObject({ precio: 6900, detalle: 'Llega el miércoles 28' });
		expect(cotizar('5000', 5000, t, lunes).map((x) => x.id)).toEqual([
			'andreani',
			'correo',
			'retiro'
		]);
	});
	it('envío gratis desde el monto de la tienda', () => {
		expect(cotizar('5000', 150000, t, lunes)[0].precio).toBe(0);
	});
	it('un código postal que no existe sólo deja retirar', () => {
		expect(cotizar('0123', 1, t, lunes).map((x) => x.id)).toEqual(['retiro']);
	});
});
