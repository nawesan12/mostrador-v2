/**
 * «Colores» de Mi tienda: el color de acento de la tienda (botones, precios, chips, links). La opción
 * 0 es el de la plantilla; las otras tres, alternativas que se leen bien con el texto que cada
 * plantilla pone encima del acento (contraste AA, 4,5:1; ver acentos.spec.ts). Las plantillas usan
 * `var(--acento, <el suyo>)` y Plantilla.svelte define --acento con el elegido.
 */

/** Oscuros, para llevar texto claro encima. */
const CON_TEXTO_CLARO = ['#B0441D', '#2F5D50', '#1F4E8C', '#7A3B69'];
/** Claros, para llevar texto oscuro encima (plantillas oscuras: Pampa, Malbec). */
const CON_TEXTO_OSCURO = ['#7FE7C4', '#FFB36B', '#B9A4FF'];
/** Medios, para llevar texto oscuro encima sobre un fondo claro (Mostaza, Patio, Hora, Lente). */
const MEDIOS = ['#E8875F', '#D8B23A', '#8FBF9F'];

type Acento = { original: string; sobre: string; familia: string[] };

/** Plantillas con archivo propio: su acento y el color del texto que va encima. */
const PROPIAS: Record<string, Acento> = {
	andes: { original: '#2F6BFF', sobre: '#FFFFFF', familia: CON_TEXTO_CLARO },
	ceibo: { original: '#9C4A2F', sobre: '#F4EFEA', familia: CON_TEXTO_CLARO },
	malbec: { original: '#E8B86A', sobre: '#1E0E12', familia: CON_TEXTO_OSCURO },
	miga: { original: '#D9452B', sobre: '#F4EADB', familia: CON_TEXTO_CLARO },
	'miga-clasica': { original: '#B5651D', sobre: '#FBF1E6', familia: CON_TEXTO_CLARO },
	mostaza: { original: '#C9821A', sobre: '#2A170C', familia: MEDIOS },
	pampa: { original: '#E6FF4D', sobre: '#111111', familia: CON_TEXTO_OSCURO },
	quebracho: { original: '#4F5B3A', sobre: '#FFFFFF', familia: CON_TEXTO_CLARO },
	rosedal: { original: '#B4475B', sobre: '#FFFFFF', familia: CON_TEXTO_CLARO },
	tijera: { original: '#5B3FD6', sobre: '#FFFFFF', familia: CON_TEXTO_CLARO },
	yerbal: { original: '#2F5D50', sobre: '#FFFFFF', familia: CON_TEXTO_CLARO }
};

/** Las del motor de plantillas (Motor.svelte): acento y texto encima, de su tema. */
const DEL_MOTOR: Record<string, Acento> = {
	patio: { original: '#F08A24', sobre: '#2A1E10', familia: MEDIOS },
	taller: { original: '#6B4A2F', sobre: '#FFFFFF', familia: CON_TEXTO_CLARO },
	hora: { original: '#B08D57', sobre: '#161616', familia: MEDIOS },
	litoral: { original: '#6B3A22', sobre: '#FFFFFF', familia: CON_TEXTO_CLARO },
	torno: { original: '#A4623D', sobre: '#FFFFFF', familia: CON_TEXTO_CLARO },
	lente: { original: '#E0A31B', sobre: '#1F1A12', familia: MEDIOS }
};

function deLaPlantilla(plantilla: string): Acento {
	return PROPIAS[plantilla] ?? DEL_MOTOR[plantilla] ?? DEL_MOTOR.litoral;
}

/** Las cuatro opciones de la plantilla: la suya y tres alternativas distintas de ella. */
export function acentos(plantilla: string): { color: string; sobre: string }[] {
	const a = deLaPlantilla(plantilla);
	const otras = a.familia.filter((c) => c.toUpperCase() !== a.original.toUpperCase()).slice(0, 3);
	return [a.original, ...otras].map((color) => ({ color, sobre: a.sobre }));
}

/** El acento elegido (paleta 0–3), o null si es el de la plantilla (no hace falta pisarlo). */
export function acentoElegido(plantilla: string, paleta: number): string | null {
	if (!paleta) return null;
	return acentos(plantilla)[paleta]?.color ?? null;
}

/** Luminancia relativa (WCAG 2). */
function luminancia(hex: string) {
	const [r, g, b] = [1, 3, 5]
		.map((i) => parseInt(hex.slice(i, i + 2), 16) / 255)
		.map((v) => (v <= 0.03928 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4));
	return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

/** Contraste entre dos colores #RRGGBB (1 a 21). */
export function contraste(a: string, b: string) {
	const [x, y] = [luminancia(a), luminancia(b)].sort((m, n) => n - m);
	return (x + 0.05) / (y + 0.05);
}

/** Las plantillas que tienen acento propio (para los tests). */
export const CON_ACENTO = [...Object.keys(PROPIAS), ...Object.keys(DEL_MOTOR)];
