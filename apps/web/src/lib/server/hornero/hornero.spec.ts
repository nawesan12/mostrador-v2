import { describe, expect, it } from 'vitest';
import { demoDe } from '../sandbox/demo.ts';
import { contextoDemo, tableroDemo } from '../sandbox/tablero.ts';
import type { DatosHornero } from './contexto.ts';
import { ideasDe } from './ideas.ts';
import { sinMontos } from '../datos/montos.ts';
import { responderSimulado } from './simulado.ts';
import { conIa, paraTelegram } from './tincho.ts';

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

	it('simulado responde con los datos', () => {
		const x = datos();
		const ayer = x.contexto.ventasPorDia.at(-2)!;
		expect(responderSimulado('¿Cuánto vendí ayer?', x)).toContain(
			'$ ' + ayer.ventas!.toLocaleString('es-AR')
		);
		expect(responderSimulado('Subí precios 8%', x)).toMatch(/8%/);
		expect(responderSimulado('Clientes que no vuelven', x)).toContain(
			String(x.contexto.clientes.sinVolver)
		);
	});

	it('sin montos, las ideas no muestran plata', () => {
		const carritos = ideasDe(sinMontos(tab()), true).find((i) => i.id === 'carritos');
		expect(carritos?.detalle).not.toContain('$');
	});

	it('la respuesta con IA trae las propuestas para confirmar', () => {
		const p = {
			id: 'p_000000000001',
			herramienta: 'crear_cupon' as const,
			argumentos: { codigo: 'VERANO15', tipo: 'porcentaje', valor: 15 },
			titulo: 'Crear el cupón VERANO15',
			detalle: '15 % de descuento · 7 días'
		};
		const r = conIa({
			texto: 'Te dejé el cupón **VERANO15** para confirmar.',
			propuestas: [p],
			fuente: 'ia'
		});
		expect(r.propuestas).toEqual([p]);
		expect(paraTelegram({ texto: r.texto, propuestas: [p], fuente: 'ia' })).toBe(
			'Te dejé el cupón VERANO15 para confirmar.\n\n• Crear el cupón VERANO15 (15 % de descuento · 7 días)\nConfirmalo desde Tincho en la app o el panel.'
		);
	});

	it('con el cupo agotado avisa y sugiere lo que el motor sí responde', () => {
		const r = conIa({
			texto: 'Ya usaste las 100 respuestas con IA de tu plan este mes.',
			propuestas: [],
			fuente: 'tope',
			tope: 'plan',
			uso: { usadas: 100, tope: 100 }
		});
		expect(r.tope).toBe('plan');
		expect(r.texto).toContain('100 respuestas');
		expect(r.sugerencias?.length).toBeGreaterThan(0);
	});
});
