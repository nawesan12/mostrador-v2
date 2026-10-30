import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { PLANTILLAS } from '#lib/plantillas.ts';
import { acentos } from '#lib/tienda/acentos.ts';
import type { TarjetaProducto, TarjetaTienda } from '#lib/tienda/compartir.ts';
import { dibujar, inicial, sinViudas, tamano, type Foto, type Nodo } from './dibujo.ts';
import { estiloDe, FUENTES, fuentesDe, PLANTILLAS_CON_ESTILO } from './estilos.ts';
import { medidas, medidor, svgDe } from './imagen.ts';

const fuentes = async (a: string) => {
	const b = readFileSync(new URL(`./fuentes/${a}.woff`, import.meta.url));
	return b.buffer.slice(b.byteOffset, b.byteOffset + b.byteLength) as ArrayBuffer;
};
const sinBajar = async () => null;

const tarjeta = (t: Partial<TarjetaTienda> = {}): TarjetaTienda => ({
	tipo: 'tienda',
	plantilla: 'miga',
	paleta: 0,
	nombre: 'La tienda de Nahuel',
	logo: '',
	frase: '',
	detalle: '',
	categorias: [],
	direccion: 'tulocalweb.com.ar/t/la-tienda-de-nahuel',
	fotos: [],
	propia: '',
	...t
});
const foto: Foto = { uri: 'data:image/jpeg;base64,AA==', ancho: 900, alto: 600 };

/** Todos los textos del árbol. */
function textos(n: Nodo | string): string[] {
	if (typeof n === 'string') return [n.replaceAll('\u00a0', ' ')];
	const h = n.props.children;
	if (!h) return [];
	return (Array.isArray(h) ? h : [h]).flatMap((x) => (x ? textos(x) : []));
}

describe('estilos', () => {
	it('las 17 plantillas tienen estilo propio, con sus fuentes en fuentes/', () => {
		expect(PLANTILLAS_CON_ESTILO.sort()).toEqual(PLANTILLAS.map((p) => p.id).sort());
		expect(PLANTILLAS_CON_ESTILO).toHaveLength(17);
		for (const id of PLANTILLAS_CON_ESTILO) {
			const e = estiloDe(id, 0);
			const familias = new Set(fuentesDe(e).map((f) => f.familia));
			expect(familias, id).toContain(e.titulo.familia);
			expect(familias, id).toContain(e.frase.familia);
		}
		for (const f of FUENTES)
			expect(() =>
				readFileSync(new URL(`./fuentes/${f.archivo}.woff`, import.meta.url))
			).not.toThrow();
	});
	it('el acento es el elegido en Mi tienda; una plantilla desconocida usa Litoral', () => {
		expect(estiloDe('pampa', 2).acento).toBe(acentos('pampa')[2].color);
		expect(estiloDe('no-existe', 0).id).toBe('litoral');
	});
});

describe('dibujo', () => {
	it('sólo escribe lo que viene de la tienda', () => {
		for (const id of PLANTILLAS_CON_ESTILO) {
			const e = estiloDe(id, 0);
			for (const t of [
				tarjeta({ plantilla: id }),
				tarjeta({
					plantilla: id,
					frase: 'Panes y Tortas',
					detalle: '2 productos',
					categorias: [
						{ nombre: 'Panes', cantidad: 1 },
						{ nombre: 'Tortas', cantidad: 1 }
					]
				})
			]) {
				for (const fotos of [[], [foto], [foto, foto, foto, foto]]) {
					const permitidos = new Set([
						t.nombre,
						t.frase,
						t.detalle,
						t.direccion,
						inicial(t.nombre),
						...t.categorias.flatMap((c) => [c.nombre, String(c.cantidad)])
					]);
					for (const s of textos(dibujar(t, e, { fotos, logo: null })))
						expect(permitidos, `${id}: «${s}»`).toContain(s);
				}
			}
		}
	});
	it('la foto propia va sola', () => {
		const n = dibujar(tarjeta({ propia: 'x' }), estiloDe('ceibo', 0), {
			fotos: [foto],
			logo: null
		});
		expect(textos(n)).toEqual([]);
	});
	it('el producto lleva nombre, precio, precio anterior y la tienda', () => {
		const p: TarjetaProducto = {
			tipo: 'producto',
			plantilla: 'pampa',
			paleta: 0,
			tienda: 'Lili',
			logo: '',
			nombre: 'Medialunas',
			precio: '$ 7.900',
			anterior: '$ 9.200',
			foto: '',
			direccion: 'lili.com.ar'
		};
		expect(textos(dibujar(p, estiloDe('pampa', 0), { fotos: [], logo: null }))).toEqual([
			'M',
			'Lili',
			'Medialunas',
			'$ 7.900',
			'$ 9.200',
			'lili.com.ar'
		]);
	});
	it('tamano achica los nombres largos y nunca parte una palabra', () => {
		const tipo = { ancho: 1, mayusculas: false };
		const caja = { ancho: 600, alto: 300, max: 130 };
		expect(tamano('Lili', caja, tipo)).toBe(130);
		const largo = tamano('Almacén de ramos generales y dietética natural', caja, tipo);
		expect(largo).toBeLessThan(90);
		expect(tamano('Supercalifragilístico', caja, tipo) * 21 * 0.56).toBeLessThanOrEqual(600);
	});
	it('sin renglones viudos: las palabras cortas van pegadas a la que sigue', () => {
		expect(sinViudas('La tienda de Nahuel')).toBe('La\u00a0tienda de\u00a0Nahuel');
		expect(sinViudas('Panes, Tortas y Café')).toBe('Panes, Tortas y\u00a0Café');
		expect(sinViudas('Lumbre')).toBe('Lumbre');
		expect(sinViudas('Panadería Lili')).toBe('Panadería Lili');
		expect(sinViudas('Velas hechas en casa')).toBe('Velas hechas en\u00a0casa');
	});
	it('inicial', () => {
		expect(inicial('  ¡Ñandú!')).toBe('Ñ');
		expect(inicial('123 Hogar')).toBe('1');
	});
});

describe('svgDe', () => {
	it('arma el SVG de cada plantilla sin fotos (las letras van como trazos)', async () => {
		for (const id of ['ceibo', 'pampa', 'yerbal', 'lente']) {
			const svg = await svgDe(tarjeta({ plantilla: id, frase: 'Hecho a mano en Rosario' }), {
				fuentes,
				bytes: sinBajar
			});
			expect(svg.startsWith('<svg width="1200" height="630"'), id).toBe(true);
			expect(svg).not.toContain('<text');
		}
	});
	it('mide las palabras con la fuente de verdad', async () => {
		const syne = estiloDe('tijera', 0).titulo;
		const serif = estiloDe('ceibo', 0).titulo;
		const fonts = await Promise.all(
			[...fuentesDe(estiloDe('tijera', 0)), ...fuentesDe(estiloDe('ceibo', 0))].map(async (f) => ({
				name: f.familia,
				data: await fuentes(f.archivo),
				weight: f.peso as 400,
				style: (f.italica ? 'italic' : 'normal') as 'normal'
			}))
		);
		const m = await medidor(
			[
				{ texto: 'La tienda de Nahuel', tipo: syne },
				{ texto: 'La tienda de Nahuel', tipo: serif }
			],
			fonts
		);
		// Las palabras cortas van pegadas a la que sigue: se miden juntas.
		expect(m('de\u00a0Nahuel', syne)!).toBeGreaterThan(m('de\u00a0Nahuel', serif)! * 1.5);
		expect(m(' ', syne)!).toBeGreaterThan(0.1);
		expect(m('de\u00a0Nahuel', syne)!).toBeGreaterThan(m('La\u00a0tienda', syne)!);
		expect(m('otra', syne)).toBeNull();
	});
	it('medidas de un PNG', () => {
		const png = new Uint8Array(24);
		png.set([0x89, 0x50, 0x4e, 0x47], 0);
		new DataView(png.buffer).setUint32(16, 640);
		new DataView(png.buffer).setUint32(20, 200);
		expect(medidas(png)).toEqual({ ancho: 640, alto: 200 });
		expect(medidas(new Uint8Array([1, 2, 3]))).toBeNull();
	});
});
