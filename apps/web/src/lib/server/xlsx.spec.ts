import { describe, expect, it } from 'vitest';
import { enCastellano } from '#lib/panel/textos-export.ts';
import { columna, xlsx } from './xlsx.ts';

describe('xlsx', () => {
	it('nombra las columnas como Excel', () => {
		expect([0, 25, 26, 27, 701, 702].map(columna)).toEqual(['A', 'Z', 'AA', 'AB', 'ZZ', 'AAA']);
	});
	it('es un ZIP con las partes de una planilla y los valores escapados', () => {
		const z = xlsx(
			[
				['Nombre', 'Total'],
				['Mate & bombilla <grande>', 15215],
				['Sin total', null]
			],
			'Pedidos'
		);
		expect(new DataView(z.buffer).getUint32(0, true)).toBe(0x04034b50);
		const t = new TextDecoder().decode(z);
		for (const parte of [
			'[Content_Types].xml',
			'_rels/.rels',
			'xl/workbook.xml',
			'xl/_rels/workbook.xml.rels',
			'xl/styles.xml',
			'xl/worksheets/sheet1.xml'
		])
			expect(t).toContain(parte);
		expect(t).toContain('<sheet name="Pedidos"');
		expect(t).toContain('Mate &amp; bombilla &lt;grande&gt;');
		expect(t).toContain('<c r="B2"><v>15215</v></c>');
		expect(t).toContain('<c r="A1" s="1" t="inlineStr">');
	});
	it('una exportación vacía igual tiene encabezado', () => {
		const t = new TextDecoder().decode(xlsx([['Fecha', 'Comprobante']]));
		expect(t).toContain('<row r="1">');
		expect(t).not.toContain('<row r="2">');
	});
});

describe('exportar en castellano', () => {
	it('cambia los códigos de canal, pago, envío, estado y segmento', () => {
		expect(
			enCastellano([
				['Canal', 'Pago', 'Envío', 'Estado', 'Segmento', 'Nombre'],
				['web', 'transf', 'retiro', 'active', 'new', 'web']
			])
		).toEqual([
			['Canal', 'Pago', 'Envío', 'Estado', 'Segmento', 'Nombre'],
			['Tienda', 'Transferencia', 'Retiro', 'Publicado', 'Nuevo', 'web']
		]);
	});
	it('deja como vino lo que no conoce', () => {
		expect(enCastellano([['Estado'], ['Entregado']])).toEqual([['Estado'], ['Entregado']]);
	});
});
