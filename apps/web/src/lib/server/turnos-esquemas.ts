/**
 * Validación de lo que manda el panel de turnos (las mismas reglas que Go: HH:MM múltiplo de 15,
 * desde < hasta, se acepta 24:00; servicios de 15 a 480 min en pasos de 15).
 */
import { error } from '@sveltejs/kit';
import * as v from 'valibot';
import { horaValida, minutos } from '#lib/turnos/libres.ts';

const hora = v.pipe(
	v.string(),
	v.check(
		(h) => horaValida(h) && minutos(h) % 15 === 0,
		'Usá horas en punto, y cuarto, y media o menos cuarto.'
	)
);
export const Franja = v.pipe(
	v.object({
		dia: v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(6)),
		desde: hora,
		hasta: hora
	}),
	v.check(
		(f) => minutos(f.desde) < minutos(f.hasta),
		'Cada franja tiene que terminar después de empezar.'
	)
);
export const Ajustes = v.object({
	activo: v.boolean(),
	sena: v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(100)),
	anticipacionMin: v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(10080)),
	diasVista: v.pipe(v.number(), v.integer(), v.minValue(1), v.maxValue(90)),
	horario: v.pipe(v.array(Franja), v.maxLength(42))
});
const nulo = (max: number) =>
	v.pipe(
		v.nullable(v.pipe(v.string(), v.trim(), v.maxLength(max))),
		v.transform((s) => s || null)
	);
const orden = v.optional(v.pipe(v.number(), v.integer(), v.minValue(0), v.maxValue(10000)));
export const Servicio = v.object({
	nombre: v.pipe(
		v.string(),
		v.trim(),
		v.minLength(2, 'Poné el nombre del servicio.'),
		v.maxLength(60)
	),
	descripcion: v.pipe(v.string(), v.trim(), v.maxLength(200)),
	precio: v.pipe(
		v.number(),
		v.integer(),
		v.minValue(1, 'Poné el precio.'),
		v.maxValue(100_000_000)
	),
	duracion: v.pipe(
		v.number(),
		v.integer(),
		v.minValue(15),
		v.maxValue(480),
		v.check((n) => n % 15 === 0, 'La duración va de a 15 minutos.')
	),
	foto: nulo(2048),
	etiqueta: nulo(20),
	activo: v.boolean(),
	orden
});
export const Profesional = v.object({
	nombre: v.pipe(v.string(), v.trim(), v.minLength(2, 'Poné el nombre.'), v.maxLength(40)),
	rol: v.pipe(v.string(), v.trim(), v.maxLength(60)),
	foto: nulo(2048),
	servicios: v.pipe(v.array(v.pipe(v.string(), v.maxLength(64))), v.maxLength(100)),
	activo: v.boolean(),
	orden
});
export const TurnoNuevo = v.object({
	servicioId: v.pipe(v.string(), v.minLength(1, 'Elegí el servicio.'), v.maxLength(64)),
	profesionalId: v.pipe(v.string(), v.minLength(1, 'Elegí quién atiende.'), v.maxLength(64)),
	inicio: v.pipe(
		v.string(),
		v.regex(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:00-03:00$/, 'Elegí día y hora.'),
		// Que la fecha exista (no un 31 de febrero) y la hora esté en rango.
		v.check((s) => {
			const [a, m, d, h, mi] = [0, 5, 8, 11, 14].map((i) =>
				Number(s.slice(i, i === 0 ? 4 : i + 2))
			);
			const f = new Date(Date.UTC(a, m - 1, d));
			return f.getUTCMonth() === m - 1 && f.getUTCDate() === d && h < 24 && mi < 60;
		}, 'Elegí día y hora.')
	),
	nombre: v.pipe(
		v.string(),
		v.trim(),
		v.minLength(2, 'Poné el nombre del cliente.'),
		v.maxLength(80)
	),
	telefono: v.pipe(v.string(), v.trim(), v.maxLength(20)),
	email: v.pipe(
		v.string(),
		v.trim(),
		v.maxLength(120),
		v.check((e) => !e || /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(e), 'Revisá el email.')
	),
	nota: v.optional(nulo(300), null),
	forzar: v.optional(v.boolean(), false)
});
export const ESTADOS = ['reservado', 'confirmado', 'cancelado', 'atendido', 'ausente'] as const;

/** Lee y valida el cuerpo JSON; 400 con el primer error. */
export async function cuerpo<T extends v.GenericSchema>(request: Request, esquema: T) {
	const r = v.safeParse(esquema, await request.json().catch(() => null));
	if (!r.success) error(400, r.issues[0].message);
	return r.output as v.InferOutput<T>;
}
