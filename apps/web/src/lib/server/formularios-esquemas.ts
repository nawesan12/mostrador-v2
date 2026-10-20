/**
 * Validación de los formularios públicos: las mismas reglas y mensajes que
 * apps/api/internal/dominio/formularios, para que la demo y la API digan lo mismo.
 */
import * as v from 'valibot';
import { ESPECIALIDADES, FACTURACION, ORIGEN } from '#lib/sitio/opciones.ts';

const recortar = (s: string) => s.trim().replace(/\s+/g, ' ');
const texto = (max: number, vacio?: string) =>
	v.pipe(
		v.optional(v.string(), ''),
		v.transform(recortar),
		v.check((s) => !vacio || s !== '', vacio ?? ''),
		v.maxLength(max, `Puede tener hasta ${max.toLocaleString('es-AR')} caracteres.`)
	);
const email = v.pipe(
	v.string(),
	v.trim(),
	v.toLowerCase(),
	v.maxLength(254, 'Revisá el email.'),
	v.email('Revisá el email.')
);
/** Una opción de chips: vacía (no eligió) o una de las válidas, tipada como tal. */
const opcion = <T extends string>(validas: readonly (readonly [T, string])[]) =>
	v.pipe(
		v.optional(v.string(), ''),
		v.check((s) => s === '' || validas.some(([k]) => k === s), 'Elegí una de las opciones.'),
		v.transform((s) => (s === '' ? undefined : (s as T)))
	);

export const Consulta = v.object({
	nombre: texto(80, 'Escribí tu nombre.'),
	email,
	whatsapp: v.pipe(
		v.optional(v.string(), ''),
		v.transform((s) => s.replace(/\D/g, '')),
		v.check(
			(s) => s === '' || (s.length >= 8 && s.length <= 15),
			'Revisá el número: con característica, sin 0 ni 15.'
		)
	),
	web: texto(200),
	facturacion: opcion(FACTURACION),
	origen: opcion(ORIGEN),
	mensaje: texto(2000)
});

export const Postulacion = v.object({
	nombre: texto(80, 'Escribí tu nombre.'),
	email,
	agencia: texto(120, '¿Cómo se llama tu agencia o estudio?'),
	web: texto(200),
	especialidades: v.pipe(
		v.array(v.string()),
		v.check(
			(xs) => xs.every((x) => ESPECIALIDADES.some(([k]) => k === x)),
			'Elegí de las opciones.'
		),
		v.transform((xs) => [...new Set(xs)] as (typeof ESPECIALIDADES)[number][0][])
	)
});

export const Suscripcion = v.object({ email });
