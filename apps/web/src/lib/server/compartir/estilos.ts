/**
 * El estilo de cada plantilla para la imagen al compartir: los mismos colores, tipografías y radios
 * que usa la plantilla en la tienda (lib/tienda/plantillas y el motor de Motor.svelte). El acento
 * es el elegido en Mi tienda (lib/tienda/acentos.ts).
 */
import { acentos } from '#lib/tienda/acentos.ts';

export type Composicion = 'ed' | 'grid' | 'cat';
/** Lo que distingue a cada plantilla además de colores y letras. */
export type Adorno =
	'marco' | 'arco' | 'linea' | 'franja' | 'circulo' | 'rombo' | 'bloque' | 'sol' | 'nota' | 'menu';

export type Tipografia = {
	familia: string;
	peso: number;
	italica: boolean;
	mayusculas: boolean;
	/** Interletrado en em. */
	tracking: number;
	/** Cuánto más ancha que una de texto común (para calcular el tamaño). */
	ancho: number;
};

export type Estilo = {
	id: string;
	fondo: string;
	tinta: string;
	suave: string;
	acento: string;
	sobre: string;
	titulo: Tipografia;
	/** La frase: itálica en las editoriales que la tienen. */
	frase: Tipografia;
	texto: string;
	radio: number;
	comp: Composicion;
	adorno: Adorno;
	oscura: boolean;
};

const t = (
	familia: string,
	peso: number,
	extra: Partial<Omit<Tipografia, 'familia' | 'peso'>> = {}
): Tipografia => ({
	familia,
	peso,
	italica: false,
	mayusculas: false,
	tracking: 0,
	ancho: 1,
	...extra
});

const SANS = t('Instrument Sans', 500);

type Base = Omit<Estilo, 'id' | 'acento' | 'sobre' | 'frase' | 'texto' | 'oscura'> & {
	frase?: Tipografia;
	texto?: string;
	oscura?: boolean;
};

// prettier-ignore
const BASE: Record<string, Base> = {
	andes: { fondo: '#F3F5F8', tinta: '#0E1116', suave: '#69717E', titulo: t('Sora', 800, { tracking: -0.04, ancho: 1.12 }), radio: 20, comp: 'grid', adorno: 'bloque' },
	ceibo: { fondo: '#F4EFEA', tinta: '#1C1A17', suave: '#6E655C', titulo: t('Instrument Serif', 400, { tracking: -0.02, ancho: 0.82 }), frase: t('Instrument Serif', 400, { italica: true, ancho: 0.8 }), radio: 2, comp: 'ed', adorno: 'marco' },
	malbec: { fondo: '#1E0E12', tinta: '#F4E9E1', suave: '#BFA9A3', titulo: t('Cormorant Garamond', 600, { italica: true, tracking: -0.01, ancho: 0.8 }), frase: t('Cormorant Garamond', 600, { italica: true, ancho: 0.8 }), radio: 2, comp: 'ed', adorno: 'marco', oscura: true },
	miga: { fondo: '#F4EADB', tinta: '#2A1C12', suave: '#7A5F4C', titulo: t('Schibsted Grotesk', 900, { tracking: -0.06, ancho: 1.02 }), radio: 18, comp: 'cat', adorno: 'nota' },
	'miga-clasica': { fondo: '#FBF1E6', tinta: '#3A2414', suave: '#7A5F4C', titulo: t('Young Serif', 400, { tracking: -0.01, ancho: 1.1 }), radio: 12, comp: 'cat', adorno: 'menu' },
	mostaza: { fondo: '#FBF3E4', tinta: '#3B2416', suave: '#78604C', titulo: t('Bricolage Grotesque', 800, { tracking: -0.045, ancho: 1 }), radio: 16, comp: 'grid', adorno: 'circulo' },
	pampa: { fondo: '#0B0B0B', tinta: '#F2F2F2', suave: '#A3A3A3', titulo: t('Archivo Black', 400, { mayusculas: true, tracking: -0.01, ancho: 1.12 }), texto: 'Archivo', radio: 4, comp: 'grid', adorno: 'franja', oscura: true },
	quebracho: { fondo: '#EFE9E1', tinta: '#2B2520', suave: '#6F6358', titulo: t('Newsreader', 500, { tracking: -0.025, ancho: 0.9 }), frase: t('Newsreader', 500, { italica: true, ancho: 0.9 }), radio: 3, comp: 'ed', adorno: 'linea' },
	rosedal: { fondo: '#FBF4F1', tinta: '#3A1F25', suave: '#8C6A70', titulo: t('Gloock', 400, { tracking: -0.02, ancho: 1 }), radio: 22, comp: 'ed', adorno: 'arco' },
	tijera: { fondo: '#F6F4F1', tinta: '#1E1B2E', suave: '#5A566B', titulo: t('Syne', 800, { tracking: -0.03, ancho: 1.45 }), radio: 14, comp: 'grid', adorno: 'rombo' },
	yerbal: { fondo: '#EEF1E8', tinta: '#1E2B22', suave: '#5B6B5F', titulo: t('Schibsted Grotesk', 800, { tracking: -0.045, ancho: 1 }), radio: 14, comp: 'cat', adorno: 'menu' },
	patio: { fondo: '#FFF7E8', tinta: '#2A1E10', suave: '#75644F', titulo: t('Fredoka', 700, { tracking: -0.02, ancho: 1.02 }), radio: 28, comp: 'grid', adorno: 'circulo' },
	taller: { fondo: '#F3EEE6', tinta: '#2A231C', suave: '#71665A', titulo: t('Bricolage Grotesque', 800, { tracking: -0.045, ancho: 1 }), radio: 8, comp: 'ed', adorno: 'linea' },
	hora: { fondo: '#F5F5F2', tinta: '#161616', suave: '#66665F', titulo: t('Bodoni Moda', 500, { italica: true, tracking: -0.02, ancho: 0.95 }), frase: t('Bodoni Moda', 500, { italica: true, ancho: 0.95 }), radio: 0, comp: 'grid', adorno: 'marco' },
	litoral: { fondo: '#F3E9DC', tinta: '#3B2416', suave: '#78604C', titulo: t('Bricolage Grotesque', 800, { tracking: -0.04, ancho: 1 }), radio: 18, comp: 'cat', adorno: 'menu' },
	torno: { fondo: '#EDE6DD', tinta: '#2C2621', suave: '#73685E', titulo: t('Instrument Serif', 400, { italica: true, tracking: -0.015, ancho: 0.8 }), frase: t('Instrument Serif', 400, { italica: true, ancho: 0.8 }), radio: 28, comp: 'ed', adorno: 'arco' },
	lente: { fondo: '#FFF8EC', tinta: '#1F1A12', suave: '#736650', titulo: t('Unbounded', 700, { tracking: -0.04, ancho: 1.45 }), radio: 22, comp: 'grid', adorno: 'sol' }
};

export const PLANTILLAS_CON_ESTILO = Object.keys(BASE);

/** El estilo de la plantilla con el acento elegido (una plantilla desconocida usa Litoral). */
export function estiloDe(plantilla: string, paleta: number): Estilo {
	const id = plantilla in BASE ? plantilla : 'litoral';
	const b = BASE[id];
	const opciones = acentos(id);
	const a = opciones[paleta] ?? opciones[0];
	return {
		...b,
		id,
		acento: a.color,
		sobre: a.sobre,
		frase: b.frase ?? t(b.texto ?? SANS.familia, 500),
		texto: b.texto ?? SANS.familia,
		oscura: !!b.oscura
	};
}

/** #RRGGBB con transparencia. */
export function alfa(hex: string, a: number) {
	const n = parseInt(hex.slice(1), 16);
	return `rgba(${(n >> 16) & 255},${(n >> 8) & 255},${n & 255},${a})`;
}

/** Las fuentes (archivos WOFF de fuentes/, todas con licencia OFL) que usa cada familia. */
export const FUENTES: { familia: string; archivo: string; peso: number; italica: boolean }[] = [
	['Sora', 'sora-latin-800-normal', 800],
	['Instrument Serif', 'instrument-serif-latin-400-normal', 400],
	['Instrument Serif', 'instrument-serif-latin-400-italic', 400, true],
	['Cormorant Garamond', 'cormorant-garamond-latin-600-normal', 600],
	['Cormorant Garamond', 'cormorant-garamond-latin-600-italic', 600, true],
	['Schibsted Grotesk', 'schibsted-grotesk-latin-800-normal', 800],
	['Schibsted Grotesk', 'schibsted-grotesk-latin-900-normal', 900],
	['Caveat', 'caveat-latin-700-normal', 700],
	['Young Serif', 'young-serif-latin-400-normal', 400],
	['Bricolage Grotesque', 'bricolage-grotesque-latin-800-normal', 800],
	['Archivo Black', 'archivo-black-latin-400-normal', 400],
	['Archivo', 'archivo-latin-600-normal', 600],
	['Newsreader', 'newsreader-latin-500-normal', 500],
	['Newsreader', 'newsreader-latin-500-italic', 500, true],
	['Gloock', 'gloock-latin-400-normal', 400],
	['Syne', 'syne-latin-800-normal', 800],
	['Fredoka', 'fredoka-latin-700-normal', 700],
	['Bodoni Moda', 'bodoni-moda-latin-500-normal', 500],
	['Bodoni Moda', 'bodoni-moda-latin-500-italic', 500, true],
	['Unbounded', 'unbounded-latin-700-normal', 700],
	['Instrument Sans', 'instrument-sans-latin-500-normal', 500],
	['Instrument Sans', 'instrument-sans-latin-600-normal', 600],
	['Instrument Sans', 'instrument-sans-latin-700-normal', 700]
].map(([familia, archivo, peso, italica]) => ({
	familia: familia as string,
	archivo: archivo as string,
	peso: peso as number,
	italica: !!italica
}));

/** Las fuentes que hacen falta para un estilo (satori no busca las que no le pasan). */
export function fuentesDe(e: Estilo) {
	const familias = new Set([
		e.titulo.familia,
		e.frase.familia,
		e.texto,
		'Instrument Sans',
		'Caveat'
	]);
	return FUENTES.filter((f) => familias.has(f.familia));
}
