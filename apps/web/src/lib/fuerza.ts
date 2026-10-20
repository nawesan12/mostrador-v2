/** Fuerza de una contraseña para el medidor del registro (0 a 4) y su texto. */
export function fuerza(pw: string): { nivel: 0 | 1 | 2 | 3 | 4; texto: string } {
	if (!pw) return { nivel: 0, texto: 'Mínimo 8 caracteres' };
	let n = 0;
	if (pw.length >= 8) n++;
	if (pw.length >= 12) n++;
	if (/[a-z]/.test(pw) && /[A-Z]/.test(pw)) n++;
	if (/\d/.test(pw) || /[^A-Za-z0-9]/.test(pw)) n++;
	if (pw.length < 8) n = Math.min(n, 1);
	const nivel = Math.min(n, 4) as 0 | 1 | 2 | 3 | 4;
	const textos = [
		'Mínimo 8 caracteres',
		'Muy corta · usá al menos 8 caracteres',
		'Débil · sumá mayúsculas y un número',
		'Bastante segura · sumá un número o un símbolo',
		'Muy segura'
	];
	return { nivel, texto: textos[nivel] };
}
