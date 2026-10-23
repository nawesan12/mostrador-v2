import { describe, expect, it } from 'vitest';
import {
	agotado,
	combinaciones,
	elegirVariante,
	errorOpciones,
	errorVariantes,
	faltaElegir,
	generarVariantes,
	precioDesde,
	stockTotal,
	textoVariante,
	valoresDisponibles,
	type Variante
} from './variantes.ts';

const opciones = [
	{ nombre: 'Talle', valores: ['S', 'M', 'L'] },
	{ nombre: 'Color', valores: ['Negro', 'Arena'] }
];
const v = (
	talle: string,
	color: string,
	stock: number,
	precio: number | null = null
): Variante => ({
	id: `${talle}-${color}`,
	valores: { Talle: talle, Color: color },
	sku: '',
	precio,
	stock,
	foto: null,
	activa: true
});
const variantes = [
	v('S', 'Negro', 0),
	v('S', 'Arena', 0),
	v('M', 'Negro', 3),
	v('M', 'Arena', 0),
	v('L', 'Negro', 2, 31_000),
	v('L', 'Arena', 1, 31_000)
];

describe('variantes', () => {
	it('arma el texto como la API: nombre y valor de cada opción, en orden', () => {
		expect(textoVariante(opciones, { Color: 'Negro', Talle: 'M' })).toBe('Talle M · Color Negro');
		expect(
			textoVariante([{ nombre: 'Presentación', valores: ['500 g'] }], { Presentación: '500 g' })
		).toBe('Presentación 500 g');
		expect(textoVariante(opciones, { Color: 'Arena' })).toBe('Color Arena');
	});

	it('elige la variante sólo con todas las opciones elegidas', () => {
		expect(elegirVariante(opciones, variantes, { Talle: 'M' })).toBeNull();
		expect(elegirVariante(opciones, variantes, { Talle: 'M', Color: 'Negro' })?.id).toBe('M-Negro');
		expect(elegirVariante([], variantes, {})).toBeNull();
	});

	it('marca disponibles los valores con stock según lo demás elegido', () => {
		expect(valoresDisponibles(opciones[0], variantes, {})).toEqual({ S: false, M: true, L: true });
		expect(valoresDisponibles(opciones[0], variantes, { Color: 'Arena' })).toEqual({
			S: false,
			M: false,
			L: true
		});
		// Lo elegido en la misma opción no cuenta: se puede cambiar de talle.
		expect(valoresDisponibles(opciones[0], variantes, { Talle: 'L' }).M).toBe(true);
		expect(valoresDisponibles(opciones[1], variantes, { Talle: 'M' })).toEqual({
			Negro: true,
			Arena: false
		});
	});

	it('precio "desde": el menor con stock, y sólo si los precios cambian', () => {
		expect(precioDesde({ precio: 28_000, variantes })).toEqual({ precio: 28_000, desde: true });
		expect(precioDesde({ precio: 28_000, variantes: [v('L', 'Negro', 2, 31_000)] })).toEqual({
			precio: 31_000,
			desde: false
		});
		// Si las baratas están agotadas, el "desde" es el de las que quedan.
		expect(
			precioDesde({ precio: 28_000, variantes: [v('S', 'Negro', 0), v('L', 'Negro', 2, 31_000)] })
		).toEqual({ precio: 31_000, desde: true });
		expect(precioDesde({ precio: 5_000, variantes: [] })).toEqual({ precio: 5_000, desde: false });
	});

	it('agotado: sin opciones mira el producto; con opciones, todas las variantes', () => {
		expect(agotado({ stock: 0, opciones: [], variantes: [] })).toBe(true);
		expect(agotado({ stock: 6, opciones, variantes })).toBe(false);
		expect(agotado({ stock: 0, opciones, variantes: variantes.slice(0, 2) })).toBe(true);
		expect(stockTotal([...variantes, { ...v('XL', 'Negro', 9), activa: false }])).toBe(6);
	});

	it('pide lo que falta con el artículo justo', () => {
		expect(faltaElegir(opciones, {})).toBe('Elegí un talle');
		expect(faltaElegir(opciones, { Talle: 'M' })).toBe('Elegí un color');
		expect(faltaElegir(opciones, { Talle: 'M', Color: 'Negro' })).toBeNull();
		expect(faltaElegir([{ nombre: 'Presentación', valores: ['1 kg'] }], {})).toBe(
			'Elegí una presentación'
		);
		expect(faltaElegir([{ nombre: 'Aroma', valores: ['Rosa'] }], {})).toBe('Elegí un aroma');
	});

	it('genera las combinaciones conservando las que ya existían', () => {
		expect(combinaciones(opciones)).toHaveLength(6);
		expect(combinaciones([])).toEqual([]);
		const con = generarVariantes(
			[opciones[0], { nombre: 'Color', valores: ['Negro', 'Blanco'] }],
			variantes,
			(i) => `nueva-${i}`
		);
		expect(con).toHaveLength(6);
		expect(con.find((x) => x.id === 'M-Negro')?.stock).toBe(3);
		const blanca = con.find((x) => x.valores.Color === 'Blanco' && x.valores.Talle === 'M')!;
		expect(blanca).toMatchObject({ id: 'nueva-3', stock: 0, activa: true, precio: null });
		// Sumar una opción hace todas las combinaciones nuevas.
		const otra = generarVariantes([...opciones, { nombre: 'Tela', valores: ['Lino'] }], variantes);
		expect(otra.every((x) => x.stock === 0 && x.id === '')).toBe(true);
	});

	it('valida opciones y variantes como la API', () => {
		expect(errorOpciones(opciones)).toBeNull();
		expect(errorOpciones([...opciones, ...opciones])).toMatch(/Hasta 3/);
		expect(errorOpciones([{ nombre: 'Talle', valores: ['S', 's'] }])).toMatch(/repetido/);
		expect(errorOpciones([{ nombre: 'Talle', valores: [] }])).toMatch(/al menos un valor/);
		expect(errorOpciones([opciones[0], { nombre: 'talle', valores: ['X'] }])).toMatch(/repetida/);
		// Sin mayúsculas ni tildes, como la API.
		expect(errorOpciones([{ nombre: 'Color', valores: ['Limón', 'limon'] }])).toMatch(/repetido/);
		const diez = Array.from({ length: 10 }, (_, i) => `v${i}`);
		expect(
			errorOpciones([
				{ nombre: 'A', valores: diez },
				{ nombre: 'B', valores: diez }
			])
		).toBeNull();
		expect(
			errorOpciones([
				{ nombre: 'A', valores: diez },
				{ nombre: 'B', valores: [...diez, 'v10'] }
			])
		).toMatch(/demasiadas/);
		expect(errorVariantes(opciones, variantes, [])).toBeNull();
		expect(errorVariantes(opciones, [v('XL', 'Negro', 1)], [])).toMatch(/no corresponde/);
		expect(errorVariantes(opciones, [v('M', 'Negro', 1), v('M', 'Negro', 2)], [])).toMatch(
			/mismos valores/
		);
		expect(errorVariantes(opciones, [{ ...v('M', 'Negro', 1), foto: 'otra' }], ['una'])).toMatch(
			/galería/
		);
		expect(errorVariantes([], variantes, [])).toMatch(/Sin opciones/);
	});
});
