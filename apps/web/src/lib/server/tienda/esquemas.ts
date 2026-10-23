/** Lo que manda quien compra al dejar una reseña o una pregunta (mismos topes que Go). */
import * as v from 'valibot';

export const NuevaResena = v.object({
	codigo: v.optional(v.pipe(v.string(), v.maxLength(80))),
	productoId: v.optional(v.pipe(v.string(), v.maxLength(80))),
	autor: v.pipe(
		v.string(),
		v.trim(),
		v.minLength(2, 'Poné tu nombre.'),
		v.maxLength(60, 'El nombre es muy largo.')
	),
	estrellas: v.pipe(
		v.number('Elegí cuántas estrellas.'),
		v.integer(),
		v.minValue(1, 'Elegí cuántas estrellas.'),
		v.maxValue(5)
	),
	texto: v.optional(
		v.pipe(v.string(), v.trim(), v.maxLength(1000, 'Contalo en menos de 1000 letras.')),
		''
	)
});

export const NuevaPregunta = v.object({
	productoId: v.pipe(v.string(), v.minLength(1), v.maxLength(80)),
	nombre: v.pipe(
		v.string(),
		v.trim(),
		v.minLength(2, 'Poné tu nombre.'),
		v.maxLength(60, 'El nombre es muy largo.')
	),
	email: v.optional(
		v.pipe(
			v.string(),
			v.trim(),
			v.maxLength(120),
			v.check((e) => !e || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(e), 'Revisá el email.')
		),
		''
	),
	pregunta: v.pipe(
		v.string(),
		v.trim(),
		v.minLength(3, 'Escribí tu pregunta.'),
		v.maxLength(500, 'Hacela en menos de 500 letras.')
	)
});
