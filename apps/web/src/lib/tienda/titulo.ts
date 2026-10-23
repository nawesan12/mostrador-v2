/**
 * Parte el título de la portada en dos (el diseño pone la segunda parte en itálica o en otro color).
 * Corta cerca de la mitad y nunca deja una palabra corta ("y", "de", "a") colgando al final de la
 * primera parte: pasa a la segunda.
 */
export function partirTitulo(titulo: string): [string, string] {
	const palabras = titulo.trim().split(/\s+/);
	if (palabras.length < 2) return [titulo.trim(), ''];
	let mitad = 0;
	let largo = 0;
	const total = titulo.trim().length;
	while (mitad < palabras.length - 1 && largo + palabras[mitad].length <= total / 2)
		largo += palabras[mitad++].length + 1;
	mitad = Math.max(1, mitad);
	while (mitad > 1 && palabras[mitad - 1].length <= 2) mitad--;
	return [palabras.slice(0, mitad).join(' '), palabras.slice(mitad).join(' ')];
}
