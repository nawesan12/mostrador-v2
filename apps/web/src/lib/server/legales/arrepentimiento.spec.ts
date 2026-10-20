import { readdirSync, readFileSync } from 'node:fs';
import { join } from 'node:path';
import * as v from 'valibot';
import { describe, expect, it } from 'vitest';
import { d1Local } from '../sandbox/d1-local.ts';
import { anioArgentina, codigoDemo, esquema } from './arrepentimiento.ts';

const dir = join(import.meta.dirname, '../../../../migraciones-sandbox');
const migraciones = readdirSync(dir)
	.filter((f) => f.endsWith('.sql'))
	.sort()
	.map((f) => readFileSync(join(dir, f), 'utf8'));

const valido = {
	email: ' Lucia@LaEsquina.com.ar ',
	nombre: 'Lucía  Fernández',
	documento: '27-38412905-4',
	plan: 'Pro mensual · 14 de octubre de 2026'
};

describe('validación (igual que Go)', () => {
	it('normaliza como la API', () => {
		const r = v.parse(esquema, valido);
		expect(r).toEqual({
			email: 'lucia@laesquina.com.ar',
			nombre: 'Lucía Fernández',
			documento: '27384129054',
			plan: 'Pro mensual · 14 de octubre de 2026',
			motivo: ''
		});
	});

	it.each([
		['email', 'lucia', 'Revisá el email: tiene que ser el de la cuenta.'],
		['nombre', '   ', 'Escribí tu nombre y apellido.'],
		['documento', '123', 'Poné tu DNI o tu CUIT, con o sin guiones.'],
		['plan', '', 'Contanos qué plan contrataste y cuándo.']
	])('marca %s', (campo, valor, mensaje) => {
		const r = v.safeParse(esquema, { ...valido, [campo]: valor });
		expect(r.success).toBe(false);
		const issue = r.issues![0];
		expect(v.getDotPath(issue)).toBe(campo);
		expect(issue.message).toBe(mensaje);
	});
});

describe('código de la demo', () => {
	it('es correlativo y con el año de la Argentina', async () => {
		const d1 = d1Local(migraciones);
		// 31/12 a las 23:30 en Buenos Aires ya es 1/1 en UTC: el año es el de acá.
		const fin = new Date('2027-01-01T02:30:00Z');
		expect(anioArgentina(fin)).toBe(2026);
		const a = await codigoDemo(d1 as unknown as D1Database, fin);
		const b = await codigoDemo(d1 as unknown as D1Database, fin);
		expect(a).toBe('ARR-2026-100000');
		expect(b).toBe('ARR-2026-100001');
		d1.cerrar();
	});
});
