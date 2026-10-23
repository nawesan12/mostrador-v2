import { describe, expect, it } from 'vitest';
import { crc32, csv, zip } from './zip.ts';

describe('zip y csv', () => {
	it('crc32 conocido', () => {
		expect(crc32(new TextEncoder().encode('123456789')).toString(16)).toBe('cbf43926');
	});
	it('arma un zip con la firma y el directorio central', () => {
		const z = zip([{ nombre: 'productos.csv', datos: new TextEncoder().encode('a,b\n1,2') }]);
		const v = new DataView(z.buffer);
		expect(v.getUint32(0, true)).toBe(0x04034b50);
		expect(v.getUint32(z.length - 22, true)).toBe(0x06054b50);
		expect(v.getUint16(z.length - 22 + 10, true)).toBe(1);
	});
	it('csv escapa y Excel lleva BOM y punto y coma', () => {
		expect(
			csv(
				[
					['a', 'b,c'],
					[1, null]
				],
				'csv'
			)
		).toBe('a,"b,c"\r\n1,');
		expect(csv([['x', 'y']], 'excel').slice(1)).toBe('x;y');
	});
});
