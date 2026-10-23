/** CBU/CVU: 22 dígitos con sus dos verificadores (como configuracion.CBUValido en Go). */
export function cbuValido(c: string): boolean {
	if (!/^\d{22}$/.test(c)) return false;
	const d = [...c].map(Number);
	const verif = (desde: number, pesos: number[]) =>
		(10 - (pesos.reduce((s, p, i) => s + d[desde + i] * p, 0) % 10)) % 10;
	return (
		verif(0, [7, 1, 3, 9, 7, 1, 3]) === d[7] &&
		verif(8, [3, 9, 7, 1, 3, 9, 7, 1, 3, 9, 7, 1, 3]) === d[21]
	);
}
