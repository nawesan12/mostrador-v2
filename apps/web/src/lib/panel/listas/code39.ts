/**
 * Código de barras Code 39 en SVG (sin librerías): lo leen todos los lectores y alcanza para SKU
 * (letras, números, guion, punto y espacio). Cada carácter son 9 barras, 3 anchas.
 */
const PATRONES: Record<string, string> = {
	'0': 'nnnwwnwnn',
	'1': 'wnnwnnnnw',
	'2': 'nnwwnnnnw',
	'3': 'wnwwnnnnn',
	'4': 'nnnwwnnnw',
	'5': 'wnnwwnnnn',
	'6': 'nnwwwnnnn',
	'7': 'nnnwnnwnw',
	'8': 'wnnwnnwnn',
	'9': 'nnwwnnwnn',
	A: 'wnnnnwnnw',
	B: 'nnwnnwnnw',
	C: 'wnwnnwnnn',
	D: 'nnnnwwnnw',
	E: 'wnnnwwnnn',
	F: 'nnwnwwnnn',
	G: 'nnnnnwwnw',
	H: 'wnnnnwwnn',
	I: 'nnwnnwwnn',
	J: 'nnnnwwwnn',
	K: 'wnnnnnnww',
	L: 'nnwnnnnww',
	M: 'wnwnnnnwn',
	N: 'nnnnwnnww',
	O: 'wnnnwnnwn',
	P: 'nnwnwnnwn',
	Q: 'nnnnnnwww',
	R: 'wnnnnnwwn',
	S: 'nnwnnnwwn',
	T: 'nnnnwnwwn',
	U: 'wwnnnnnnw',
	V: 'nwwnnnnnw',
	W: 'wwwnnnnnn',
	X: 'nwnnwnnnw',
	Y: 'wwnnwnnnn',
	Z: 'nwwnwnnnn',
	'-': 'nwnnnnwnw',
	'.': 'wwnnnnwnn',
	' ': 'nwwnnnwnn',
	'*': 'nwnnwnwnn'
};

/** El SVG del código (alto en px). Los caracteres que Code 39 no tiene se omiten. */
export function code39(texto: string, alto = 48) {
	const limpio = texto
		.toUpperCase()
		.split('')
		.filter((c) => c in PATRONES && c !== '*')
		.join('');
	const angosta = 2;
	const ancha = 5;
	let x = 0;
	const barras: string[] = [];
	for (const c of `*${limpio}*`) {
		PATRONES[c].split('').forEach((p, i) => {
			const w = p === 'w' ? ancha : angosta;
			if (i % 2 === 0) barras.push(`<rect x="${x}" y="0" width="${w}" height="${alto}"/>`);
			x += w;
		});
		x += angosta;
	}
	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${x} ${alto}" width="${x}" height="${alto}" role="img" aria-label="${limpio}">${barras.join('')}</svg>`;
}
