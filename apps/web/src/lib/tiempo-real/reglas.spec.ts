import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join } from 'node:path';
import { describe, expect, it } from 'vitest';
import { claves, espera, leerEvento, TODAS, trasCierre, urlTiempoReal } from './reglas.ts';

describe('claves', () => {
	it('un pedido refresca el Inicio, las insignias, la lista y la ficha del cliente', () => {
		const c = claves({ tipo: 'pedido.nuevo', id: '12' });
		for (const k of ['tienda:pedidos', 'tienda:resumen', 'tienda:tablero', 'tienda:clientes'])
			expect(c).toContain(k);
	});
	it('el stock refresca productos, inventario y el punto de venta', () => {
		expect(claves({ tipo: 'producto.stock', id: 'p' })).toEqual(
			expect.arrayContaining(['tienda:productos', 'tienda:inventario', 'tienda:pos'])
		);
	});
	it('un tipo desconocido no recarga nada', () => {
		expect(claves({ tipo: 'secreto.cambio' })).toEqual([]);
	});
	it('las sesiones sólo recargan la lista de sesiones', () => {
		expect(claves({ tipo: 'sesion.cambio' })).toEqual(['tienda:sesiones']);
	});
});

describe('espera', () => {
	it('crece al doble hasta el tope, con azar entre la mitad y el total', () => {
		expect(espera(0, () => 0)).toBe(500);
		expect(espera(0, () => 1)).toBe(1000);
		expect(espera(3, () => 1)).toBe(8000);
		expect(espera(20, () => 1)).toBe(30_000);
		expect(espera(20, () => 0)).toBe(15_000);
	});
});

describe('trasCierre', () => {
	it('decide según el código del servidor', () => {
		expect(trasCierre(4000)).toBe('ya');
		expect(trasCierre(1012)).toBe('ya');
		expect(trasCierre(4401)).toBe('parar');
		expect(trasCierre(4403)).toBe('parar');
		expect(trasCierre(1008)).toBe('parar');
		expect(trasCierre(4429)).toBe('mas-tarde');
		expect(trasCierre(1006)).toBe('reintentar');
		expect(trasCierre(4408)).toBe('reintentar');
	});
});

describe('urlTiempoReal', () => {
	it('mismo sitio con wss en https y ws en http', () => {
		expect(
			urlTiempoReal({ protocol: 'https:', host: 'tulocalweb.com.ar' }, '/v1/tiempo-real')
		).toBe('wss://tulocalweb.com.ar/v1/tiempo-real');
		expect(urlTiempoReal({ protocol: 'http:', host: 'localhost:5173' }, '/v1/tiempo-real')).toBe(
			'ws://localhost:5173/v1/tiempo-real'
		);
	});
	it('con otro origen (dominio propio, pruebas)', () => {
		expect(
			urlTiempoReal(
				{ protocol: 'https:', host: 'almacensol.com.ar' },
				'/v1/tiempo-real/publico?tienda=sol',
				'https://tulocalweb.com.ar'
			)
		).toBe('wss://tulocalweb.com.ar/v1/tiempo-real/publico?tienda=sol');
		expect(
			urlTiempoReal(
				{ protocol: 'http:', host: 'localhost:4174' },
				'/v1/tiempo-real',
				'ws://localhost:4000'
			)
		).toBe('ws://localhost:4000/v1/tiempo-real');
	});
});

describe('leerEvento', () => {
	it('lee lo que manda el servidor y descarta lo que no entiende', () => {
		expect(leerEvento('{"tipo":"pedido.nuevo","id":"12"}')).toEqual({
			tipo: 'pedido.nuevo',
			id: '12'
		});
		expect(leerEvento('nada')).toBeNull();
		expect(leerEvento('{"id":"1"}')).toBeNull();
		expect(leerEvento(new ArrayBuffer(2))).toBeNull();
	});
});

/** Todos los archivos de un directorio, recursivo. */
function archivos(dir: string): string[] {
	return readdirSync(dir).flatMap((n) => {
		const p = join(dir, n);
		return statSync(p).isDirectory() ? archivos(p) : [p];
	});
}

describe('depends', () => {
	it('cada load depende de claves que algún aviso invalida (si no, nunca se recargaría)', () => {
		const usadas = new Set<string>();
		for (const f of archivos('src/routes').filter((f) =>
			/\+(page|layout)(\.server)?\.ts$/.test(f)
		)) {
			for (const m of readFileSync(f, 'utf8').matchAll(/depends\(([^)]*)\)/g))
				for (const k of m[1].matchAll(/'(tienda:[\w-]+)'/g)) usadas.add(k[1]);
		}
		expect(usadas.size).toBeGreaterThan(15);
		const publicas = ['tienda:vidriera', 'tienda:pedido-publico'];
		for (const k of usadas) expect([...TODAS, ...publicas]).toContain(k);
	});
});
