/**
 * Botón de arrepentimiento (Res. 424/2020): el formulario de /legales/arrepentimiento.
 * Con DATA_SOURCE=api el pedido va a la API Go (que lo registra y manda el mail). En la demo no hay
 * a quién revocarle nada: se valida igual que Go y se da un código correlativo, sin guardar datos.
 */
import * as v from 'valibot';
import type { components } from '@mostrador/contrato/esquema';

export type PedidoArrepentimiento = components['schemas']['PedidoArrepentimiento'];

const recortar = (s: string) => s.trim().replace(/\s+/g, ' ');

/** Las mismas reglas y mensajes que apps/api/internal/dominio/arrepentimiento. */
export const esquema = v.object({
	email: v.pipe(
		v.string(),
		v.trim(),
		v.toLowerCase(),
		v.maxLength(254, 'Revisá el email: tiene que ser el de la cuenta.'),
		v.email('Revisá el email: tiene que ser el de la cuenta.')
	),
	nombre: v.pipe(
		v.string(),
		v.transform(recortar),
		v.nonEmpty('Escribí tu nombre y apellido.'),
		v.maxLength(80, 'Escribí tu nombre y apellido.')
	),
	documento: v.pipe(
		v.string(),
		v.transform((s) => s.replace(/\D/g, '')),
		v.check((s) => [7, 8, 11].includes(s.length), 'Poné tu DNI o tu CUIT, con o sin guiones.')
	),
	plan: v.pipe(
		v.string(),
		v.trim(),
		v.nonEmpty('Contanos qué plan contrataste y cuándo.'),
		v.maxLength(120, 'Contanos qué plan contrataste y cuándo.')
	),
	motivo: v.optional(
		v.pipe(
			v.string(),
			v.trim(),
			v.maxLength(1000, 'El motivo puede tener hasta 1.000 caracteres.')
		),
		''
	)
});

/** Año en la Argentina (el del código de trámite). */
export const anioArgentina = (ahora = new Date()) =>
	Number(
		new Intl.DateTimeFormat('es-AR', {
			year: 'numeric',
			timeZone: 'America/Argentina/Buenos_Aires'
		}).format(ahora)
	);

/** Demo: el próximo número del contador, en una sola escritura atómica. */
export async function codigoDemo(d1: D1Database, ahora = new Date()) {
	const fila = await d1
		.prepare(
			`INSERT INTO contadores (nombre, valor) VALUES ('arrepentimiento', 100000)
			 ON CONFLICT (nombre) DO UPDATE SET valor = valor + 1 RETURNING valor`
		)
		.first<{ valor: number }>();
	return `ARR-${anioArgentina(ahora)}-${fila!.valor}`;
}
