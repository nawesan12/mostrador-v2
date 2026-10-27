import { describe, expect, it } from 'vitest';
import { completar, EJEMPLO, errorTarjeta, linkQr, TARJETA_INICIAL } from './tarjeta.ts';

describe('tarjeta A6', () => {
	it('completa las variables con los datos del pedido', () => {
		expect(completar(TARJETA_INICIAL.titulo, EJEMPLO)).toBe('¡Gracias, Lucía!');
		expect(completar(TARJETA_INICIAL.mensaje, EJEMPLO)).toBe(
			'Lo preparamos a mano el 20 de octubre. Que lo disfrutes.'
		);
		expect(completar('{productos} · {pedido}', EJEMPLO)).toBe(
			'Mate de calabaza y Bombilla de alpaca · #1048'
		);
		expect(completar('{productos}', { ...EJEMPLO, productos: ['A', 'B', 'C', 'D'] })).toBe(
			'A, B y 2 más'
		);
		expect(completar('Hola {nombre}', { ...EJEMPLO, nombre: '  ' })).toBe('Hola vos');
	});

	it('arma el link del QR, o null si falta el dato', () => {
		const tienda = { url: 'https://mostrador.iamnaw.dev/t/mate-y-co', whatsapp: '11 5555-0101' };
		const t = TARJETA_INICIAL;
		expect(linkQr(t, tienda)).toBe('https://mostrador.iamnaw.dev/t/mate-y-co#resenas');
		expect(linkQr({ ...t, qr: 'cupon' }, tienda)).toBeNull();
		expect(linkQr({ ...t, qr: 'cupon', cupon: 'LUCIA15' }, tienda)).toBe(
			'https://mostrador.iamnaw.dev/t/mate-y-co/c/LUCIA15'
		);
		expect(linkQr({ ...t, qr: 'whatsapp' }, tienda)).toBe('https://wa.me/541155550101');
		expect(linkQr({ ...t, qr: 'whatsapp' }, { ...tienda, whatsapp: '' })).toBeNull();
		expect(linkQr({ ...t, qr: 'instagram', instagram: 'mateyco' }, tienda)).toBe(
			'https://instagram.com/mateyco'
		);
	});

	it('valida lo que se guarda', () => {
		expect(errorTarjeta(TARJETA_INICIAL)).toBeNull();
		expect(errorTarjeta({ ...TARJETA_INICIAL, titulo: 'x'.repeat(41) })).toMatch(/40/);
		expect(errorTarjeta({ ...TARJETA_INICIAL, titulo: ' ' })).toBe('Poné un título.');
		expect(errorTarjeta({ ...TARJETA_INICIAL, cupon: 'mal cupón' })).toBe('Ese cupón no existe.');
		expect(errorTarjeta({ ...TARJETA_INICIAL, modelo: 'otro' as 'arco' })).toBe('Elegí un modelo.');
	});
});
