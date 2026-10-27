import { describe, expect, it } from 'vitest';
import { CLAVE, tipoDe } from './archivos.ts';

const bytes = (...xs: (number | string)[]) =>
	new Uint8Array(
		xs.flatMap((x) => (typeof x === 'string' ? [...x].map((c) => c.charCodeAt(0)) : [x]))
	);

describe('archivos', () => {
	it('reconoce el formato por la firma, no por lo que dice el navegador', () => {
		expect(tipoDe(bytes('RIFF', 0, 0, 0, 0, 'WEBPVP8 '))).toBe('image/webp');
		expect(tipoDe(bytes(0xff, 0xd8, 0xff, 0xe0, 0))).toBe('image/jpeg');
		expect(tipoDe(bytes(0x89, 'PNG', 13, 10, 26, 10, 0))).toBe('image/png');
		expect(tipoDe(bytes('<svg xmlns="http://www.w3.org/2000/svg">'))).toBeNull();
		expect(tipoDe(bytes('GIF89a', 0, 0, 0, 0))).toBeNull();
	});

	it('sólo sirve claves de fotos, sin salir de la carpeta', () => {
		const ok =
			'fotos/7d3e2f10-aaaa-4bbb-8ccc-0123456789ab/0b7f3c1e-1a2b-4c3d-8e9f-001122334455.webp';
		expect(CLAVE.test(ok)).toBe(true);
		expect(CLAVE.test('fotos/../secreto.webp')).toBe(false);
		expect(CLAVE.test(ok.replace('fotos/', 'otra/'))).toBe(false);
		// El id del sandbox (la cookie, base64url) nunca puede ser parte de una clave.
		expect(CLAVE.test('fotos/Ab_cD-ef/0b7f3c1e-1a2b-4c3d-8e9f-001122334455.webp')).toBe(false);
		expect(CLAVE.test(ok.replace('.webp', '.svg'))).toBe(false);
	});
});
