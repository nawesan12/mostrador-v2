import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { PERMISOS, seccionesDe } from './permisos.ts';

describe('permisos del equipo', () => {
	it('son los mismos que en Go y en la base', () => {
		const url = new URL(
			'../../../../api/internal/dominio/equipo/testdata/permisos.json',
			import.meta.url
		);
		expect(PERMISOS).toEqual(JSON.parse(readFileSync(url, 'utf8')));
	});
	it('la navegación cambia según el rol', () => {
		expect(seccionesDe(PERMISOS.soporte).map((s) => s.id)).toEqual([
			'inicio',
			'tiendas',
			'consultas'
		]);
		expect(seccionesDe(PERMISOS.administracion).map((s) => s.id)).toEqual([
			'inicio',
			'tiendas',
			'consultas',
			'partners',
			'facturacion',
			'auditoria'
		]);
	});
});
