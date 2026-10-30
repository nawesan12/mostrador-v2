import { describe, expect, it } from 'vitest';
import { cotizar } from './envio.ts';

const t = {
	envios: {
		gratisDesde: 100000,
		retiro: true,
		retiroDireccion: 'Belgrano 742',
		retiroHorario: 'Lun a vie de 10 a 19',
		moto: true,
		motoCosto: 3500,
		motoZona: 'Palermo y Villa Crespo',
		motoDemora: 'En el día',
		correo: true
	}
};
const correo = {
	domicilio: { precio: 4981, producto: 'Paq.ar Clásico', diasMin: '2', diasMax: '5' },
	sucursal: { precio: 3981, producto: 'Paq.ar Clásico' }
};

describe('cotizar', () => {
	it('sin la cotización de Correo, Correo no aparece (nunca un precio inventado)', () => {
		expect(cotizar('1414', 5000, t).map((x) => x.id)).toEqual(['moto', 'retiro']);
		expect(cotizar('01', 5000, t).map((x) => x.id)).toEqual(['retiro']);
	});
	it('la moto con el costo y la zona de la tienda; el retiro con dirección y horario', () => {
		const [moto, retiro] = cotizar('1414', 5000, t);
		expect(moto).toEqual({
			id: 'moto',
			nombre: 'Envío en moto',
			detalle: 'Llega a Palermo y Villa Crespo · En el día',
			precio: 3500
		});
		expect(retiro.detalle).toBe('Belgrano 742 · Lun a vie de 10 a 19');
		expect(cotizar('1414', 150000, t)[0].precio).toBe(0);
	});
	it('sólo lo que la tienda prendió', () => {
		const sinNada = { envios: { ...t.envios, retiro: false, moto: false } };
		expect(cotizar('1414', 5000, sinNada)).toEqual([]);
		const sinDatos = {
			envios: { ...t.envios, moto: false, retiroDireccion: '', retiroHorario: '' }
		};
		expect(cotizar('1414', 5000, sinDatos)[0].detalle).toBe('Te avisamos cuando esté listo');
	});
	it('con la cotización real: a domicilio y a sucursal, con el plazo si lo informa', () => {
		const o = cotizar('5000', 5000, t, correo);
		expect(o.map((x) => x.id)).toEqual(['correo-domicilio', 'correo-sucursal', 'moto', 'retiro']);
		expect(o[0]).toMatchObject({ precio: 4981, detalle: expect.stringContaining('2 a 5 días') });
		expect(o[1]).toMatchObject({ precio: 3981, detalle: expect.stringContaining('Paq.ar') });
	});
	it('sólo si la tienda conectó Correo', () => {
		const sin = { ...t, envios: { ...t.envios, correo: false } };
		expect(cotizar('5000', 5000, sin, correo).map((x) => x.id)).toEqual(['moto', 'retiro']);
	});
	it('envío gratis desde el monto de la tienda', () => {
		expect(cotizar('5000', 150000, t, correo)[0].precio).toBe(0);
	});
	it('un código postal inválido sólo deja retirar', () => {
		expect(cotizar('01', 1, t, correo).map((x) => x.id)).toEqual(['retiro']);
	});
});
