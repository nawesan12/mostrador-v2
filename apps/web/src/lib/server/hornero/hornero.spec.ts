import { describe, expect, it } from 'vitest';
import { demoDe } from '../sandbox/demo.ts';
import { contextoDemo, tableroDemo } from '../sandbox/tablero.ts';
import { datosParaElModelo, sistema, type DatosHornero } from './contexto.ts';
import { ideasDe } from './ideas.ts';
import { soloTexto } from './modelo.ts';
import { responderSimulado } from './simulado.ts';

const BASE = Date.parse('2026-10-20T15:00:00Z');
const AHORA = BASE + 30 * 60_000;
const d = demoDe('litoral', BASE)!;
const tab = (cambios = {}) =>
	tableroDemo(d, '7d', { base: BASE, ahora: AHORA, hola: 'Valeria', cambios });
const datos = (): DatosHornero => ({
	tienda: { nombre: 'Mate & Co', rubro: 'almacen' },
	hola: 'Valeria',
	hoy: tableroDemo(d, 'hoy', { base: BASE, ahora: AHORA, hola: 'Valeria', cambios: {} }),
	semana: tab(),
	contexto: contextoDemo(d, AHORA)
});

describe('Tincho', () => {
	it('propone las tres ideas y marca las aprobadas', () => {
		const ideas = ideasDe(tab());
		expect(ideas.map((i) => i.id)).toEqual(['carritos', 'stock', 'promo']);
		expect(ideas.every((i) => !i.hecha)).toBe(true);
		expect(ideas[0].titulo).toBe('Recuperar 12 carritos');
		const hechas = ideasDe(tab({ ideasHechas: ['carritos', 'promo'] }));
		expect(hechas.filter((i) => i.hecha).map((i) => i.id)).toEqual(['carritos', 'promo']);
		expect(tab().hornero.ideas).toBe(3);
	});

	it('el contexto trae 14 días que terminan hoy y los productos del catálogo', () => {
		const c = contextoDemo(d, AHORA);
		expect(c.ventasPorDia).toHaveLength(14);
		expect(c.ventasPorDia.at(-1)!.fecha).toBe('2026-10-20');
		expect(c.productos).toHaveLength(d.catalogo.productos.length);
		expect(c.clientes.conWhatsApp).toBe(tab().hornero.audiencia);
	});

	it('el mensaje de sistema lleva la tienda, la regla de no inventar y los datos', () => {
		const s = sistema(datos(), new Date(AHORA));
		expect(s).toContain('«Mate & Co»');
		expect(s).toContain('nunca inventes');
		expect(s).toContain('"ventasPorDia"');
		expect(s.length).toBeLessThan(12_000);
	});

	it('le pasa al modelo ayer completo, no "ayer a esta hora" (lo confundía con el total)', () => {
		const x = datosParaElModelo(datos());
		expect(x.ayerDiaCompleto).toEqual(datos().contexto.ventasPorDia.at(-2));
		expect(JSON.stringify(x)).not.toContain('MismaHora');
	});

	it('simulado responde con los datos', () => {
		const x = datos();
		const ayer = x.contexto.ventasPorDia.at(-2)!;
		expect(responderSimulado('¿Cuánto vendí ayer?', x)).toContain(
			'$ ' + ayer.ventas.toLocaleString('es-AR')
		);
		expect(responderSimulado('Subí precios 8%', x)).toMatch(/8%/);
		expect(responderSimulado('Clientes que no vuelven', x)).toContain(
			String(x.contexto.clientes.sinVolver)
		);
	});

	it('pasa los dos formatos de stream de Workers AI a texto', async () => {
		const enc = new TextEncoder();
		const sse = new ReadableStream<Uint8Array>({
			start(c) {
				c.enqueue(enc.encode('data: {"response":"Hola, "}\n\ndata: {"choices":[{"delta":{"con'));
				c.enqueue(enc.encode('tent":"Valeria"}}]}\n\n'));
				c.enqueue(
					enc.encode('data: {"choices":[{"delta":{"reasoning_content":"x"}}]}\n\ndata: [DONE]\n\n')
				);
				c.close();
			}
		});
		const texto = await new Response(soloTexto(sse)).text();
		expect(texto).toBe('Hola, Valeria');
	});
});
