/**
 * Genera src/lib/plantillas.ts desde design/templates-data.js (las 16 plantillas del handoff):
 * colores, tipografía, layout, foto, productos de muestra y la tienda de ejemplo.
 * Uso: bun scripts/plantillas.ts
 */
import { readFileSync, writeFileSync } from 'node:fs';
import { join } from 'node:path';
import { runInNewContext } from 'node:vm';

type Fuente = {
	id: string;
	name: string;
	rubro: string;
	layout: string;
	store: string;
	img: string;
	bg: string;
	ink: string;
	acc: string;
	ff: string;
	r: string;
	tag: string;
	dark: boolean;
	prods: { pos: string; price: string }[];
};
const window = {} as { MOSTRADOR_TEMPLATES: Fuente[]; MOSTRADOR_RUBROS: string[] };
const __dirname = import.meta.dirname;
// templates-data.js es un script de navegador que escribe en `window`: se corre aislado.
runInNewContext(readFileSync(join(__dirname, '../../../design/templates-data.js'), 'utf8'), {
	window
});
// Lemas del handoff que prometen algo que la tienda todavía no hace (los turnos online de Tijera):
// se reemplazan hasta que exista la función.
const LEMAS: Record<string, string> = { tijera: 'Cortes, barba y productos' };
const filas = window.MOSTRADOR_TEMPLATES.map((t) => ({
	id: t.id,
	nombre: t.name,
	rubro: t.rubro,
	layout: t.layout,
	tienda: t.store,
	img: t.img,
	bg: t.bg,
	tx: t.ink,
	ac: t.acc,
	ff: t.ff,
	radio: t.r,
	lema: LEMAS[t.id] ?? t.tag,
	oscura: t.dark,
	productos: t.prods.map((p) => ({ pos: p.pos, precio: p.price }))
}));
// Miga tiene dos versiones en el handoff (Plantilla Miga y Plantilla Miga Clasica): la clásica no está
// en templates-data.js, así que se suma con su paleta y su tipografía.
const miga = filas.find((f) => f.id === 'miga');
if (miga)
	filas.splice(filas.indexOf(miga) + 1, 0, {
		...miga,
		id: 'miga-clasica',
		nombre: 'Miga Clásica',
		tienda: 'Panadería Lola Clásica',
		bg: '#FBF1E6',
		tx: '#3A2414',
		ac: '#B5651D',
		ff: "'Young Serif', Georgia, serif",
		radio: '24px'
	});
const ts = `// Generado por scripts/plantillas.ts desde design/templates-data.js — no editar a mano.
/** Una plantilla de tienda del handoff. */
export type Plantilla = {
	id: string;
	nombre: string;
	/** Rubro tal como se muestra ("Gastronomía"). */
	rubro: string;
	layout: 'ed' | 'grid' | 'cat';
	tienda: string;
	img: string;
	bg: string;
	tx: string;
	ac: string;
	ff: string;
	radio: string;
	lema: string;
	oscura: boolean;
	productos: { pos: string; precio: string }[];
};

export const PLANTILLAS: Plantilla[] = ${JSON.stringify(filas, null, '\t')};

export const plantilla = (id: string) => PLANTILLAS.find((p) => p.id === id);

/** Nombre del layout (handoff: TPL_LAYOUT). */
export const LAYOUT = { ed: 'Editorial', grid: 'Grilla', cat: 'Catálogo' } as const;

/** Rubros en el orden del handoff (MOSTRADOR_RUBROS). */
export const RUBROS_PLANTILLAS = ${JSON.stringify(window.MOSTRADOR_RUBROS)} as const;

/**
 * Plantillas con su versión trabajada en el handoff (Plantilla X.dc.html): las tarjetas muestran la
 * captura de su portada (static/plantillas/<id>.jpg) en vez de la miniatura genérica.
 */
export const CON_PORTADA: ReadonlySet<string> = new Set(${JSON.stringify(['ceibo', 'mostaza', 'yerbal', 'tijera', 'pampa', 'malbec', 'miga', 'rosedal', 'andes', 'quebracho'])});

/** Color legible sobre el acento (handoff: lumOf(acc) > 0.62 ? '#141414' : '#FFFFFF'). */
export function sobreAcento(hex: string) {
	const n = parseInt(hex.slice(1), 16);
	const lum = (0.299 * ((n >> 16) & 255) + 0.587 * ((n >> 8) & 255) + 0.114 * (n & 255)) / 255;
	return lum > 0.62 ? '#141414' : '#FFFFFF';
}
`;
writeFileSync(join(__dirname, '../src/lib/plantillas.ts'), ts);
console.log(filas.length + ' plantillas');
