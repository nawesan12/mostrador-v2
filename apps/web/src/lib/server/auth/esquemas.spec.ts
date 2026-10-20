import * as v from 'valibot';
import { describe, expect, it } from 'vitest';
import { errores, Ingreso, Recuperar, Registro } from './esquemas.ts';

describe('esquemas de ingreso', () => {
	it('normaliza el email', () => {
		const r = v.parse(Ingreso, { email: '  Lucia@LaEsquina.com.ar ', clave: 'x' });
		expect(r.email).toBe('lucia@laesquina.com.ar');
	});

	it('el registro exige términos y 8 caracteres', () => {
		const r = v.safeParse(Registro, {
			nombre: 'Lucía',
			tienda: 'La Esquina',
			email: 'a@b.co',
			clave: 'corta'
		});
		expect(r.success).toBe(false);
		const e = errores(r.issues!);
		expect(e.clave).toContain('8 caracteres');
		expect(e.terminos).toContain('términos');
	});

	it('recuperar por WhatsApp limpia el número', () => {
		const r = v.parse(Recuperar, { canal: 'whatsapp', telefono: '(11) 2345-6789' });
		expect(r).toEqual({ canal: 'whatsapp', telefono: '1123456789' });
		expect(v.safeParse(Recuperar, { canal: 'whatsapp', telefono: '123' }).success).toBe(false);
	});
});
