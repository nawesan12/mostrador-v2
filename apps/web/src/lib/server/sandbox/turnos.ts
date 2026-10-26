/**
 * Turnos de la demo. La plantilla Tijera (peluquería y barbería) arranca con turnos activos, el
 * horario, los servicios y el equipo de su diseño (design/Plantilla Tijera.dc.html: `hours`,
 * `services`, `team`, con sus fotos de Unsplash) y unos turnos de ejemplo en los próximos días,
 * fechados respecto de cuándo se creó el sandbox, como el resto de la demo. Las demás plantillas
 * arrancan con turnos apagados y sin servicios.
 *
 * Como todo el sandbox: esto sale del código, y D1 guarda sólo lo que cambia el visitante (entidad
 * `turnos`, claves `ajustes`, `s:<id>`, `p:<id>`, `t:<id>` y `c:<codigo>` para buscar por código).
 */
import {
	diaAr,
	diaSemana,
	instante,
	isoAr,
	minutos,
	senaDe,
	sumarDias,
	type AjustesTurnos,
	type Franja,
	type Profesional,
	type Servicio,
	type Turno
} from '#lib/turnos/libres.ts';

export const ENTIDAD = 'turnos';

const franjas = (dias: number[], desde: string, hasta: string): Franja[] =>
	dias.map((dia) => ({ dia, desde, hasta }));

/** Lo que trae una tienda que nunca tocó los turnos. */
export const AJUSTES_INICIALES: AjustesTurnos = {
	activo: false,
	sena: 30,
	anticipacionMin: 60,
	diasVista: 14,
	horario: franjas([1, 2, 3, 4, 5], '10:00', '19:00')
};

// design/Plantilla Tijera.dc.html: hours (lunes a viernes de 10 a 20:30, sábado de 9 a 18).
const HORARIO_TIJERA = [
	...franjas([1, 2, 3, 4, 5], '10:00', '20:30'),
	...franjas([6], '09:00', '18:00')
];

// services: nombre, descripción, precio, duración, foto y etiqueta.
const SERVICIOS_TIJERA: Servicio[] = (
	[
		[
			'corte',
			'Corte',
			'Tijera o máquina, lavado incluido',
			12500,
			40,
			'1599351431202-1e0f0137899a',
			null
		],
		[
			'corte-barba',
			'Corte + barba',
			'Perfilado con navaja y toalla caliente',
			18900,
			60,
			'1621605815971-fbc98d665033',
			'Promo'
		],
		['barba', 'Barba', 'Perfilado, aceite y bálsamo', 8900, 30, '1622286342621-4bd786c2447c', null],
		['color', 'Color', 'Global, mechas o canas', 24000, 90, '1560066984-138dadb4c035', null],
		[
			'corte-infantil',
			'Corte infantil',
			'Hasta 12 años',
			9500,
			30,
			'1595476108010-b4d1f102b1b1',
			null
		],
		[
			'afeitado',
			'Afeitado clásico',
			'Navaja, vapor y masaje',
			11000,
			40,
			'1503951914875-452162b0f3f1',
			'Nuevo'
		]
	] as const
).map(([id, nombre, descripcion, precio, duracion, foto, etiqueta], orden) => ({
	id,
	nombre,
	descripcion,
	precio,
	duracion,
	foto,
	etiqueta,
	activo: true,
	orden
}));

// team: quién atiende y qué hace (el diseño dice el rol; los servicios salen de ahí).
const EQUIPO_TIJERA: Profesional[] = (
	[
		[
			'bruno',
			'Bruno',
			'Barbero · fades y barba',
			'1599566150163-29194dcaad36',
			['corte', 'corte-barba', 'barba', 'corte-infantil', 'afeitado']
		],
		['lara', 'Lara', 'Estilista · color', '1580489944761-15a19d654956', ['corte', 'color']],
		[
			'nico',
			'Nico',
			'Barbero · clásicos',
			'1506794778202-cad84cf45f1d',
			['corte', 'corte-barba', 'barba', 'corte-infantil', 'afeitado']
		],
		[
			'ema',
			'Ema',
			'Estilista · cortes largos',
			'1438761681033-6461ffad8d80',
			['corte', 'color', 'corte-infantil']
		]
	] as const
).map(([id, nombre, rol, foto, servicios], orden) => ({
	id,
	nombre,
	rol,
	foto,
	servicios: [...servicios],
	activo: true,
	orden
}));

const CLIENTES = [
	['Tomás Acosta', '11 5555-0148'],
	['Florencia Ruiz', '11 5555-0172'],
	['Julián Sosa', '11 5555-0119'],
	['Martina Gómez', '11 5555-0163'],
	['Lucas Herrera', '11 5555-0187'],
	['Camila Díaz', '11 5555-0135'],
	['Federico Paz', '11 5555-0156'],
	['Sofía Medina', '11 5555-0124'],
	['Matías Romero', '11 5555-0191'],
	['Paula Castro', '11 5555-0108']
] as const;

// Turnos de ejemplo: [día desde la creación, hora, profesional, servicio].
const EJEMPLOS = [
	[0, '11:00', 'bruno', 'corte-barba'],
	[0, '12:00', 'ema', 'corte'],
	[0, '15:00', 'nico', 'corte'],
	[0, '16:00', 'lara', 'color'],
	[1, '10:00', 'bruno', 'barba'],
	[1, '10:30', 'ema', 'corte-infantil'],
	[1, '11:00', 'lara', 'corte'],
	[1, '17:30', 'bruno', 'corte-barba'],
	[1, '18:15', 'nico', 'afeitado'],
	[2, '11:30', 'nico', 'corte-barba'],
	[2, '14:00', 'lara', 'color'],
	[2, '15:00', 'bruno', 'corte'],
	[3, '12:15', 'bruno', 'corte'],
	[3, '16:45', 'ema', 'corte'],
	[4, '10:00', 'nico', 'barba'],
	[4, '13:00', 'lara', 'corte']
] as const;

const MIN = 60_000;
const hash = (s: string) =>
	[...s].reduce((h, c) => Math.imul(h ^ c.charCodeAt(0), 16777619), 2166136261) >>> 0;

export type TurnosDemo = {
	ajustes: AjustesTurnos;
	servicios: Servicio[];
	profesionales: Profesional[];
	turnos: Turno[];
};

/** Los turnos de ejemplo de una plantilla, fechados desde `base` (la creación del sandbox). */
export function turnosDemo(plantilla: string, base: number): TurnosDemo {
	if (plantilla !== 'tijera')
		return { ajustes: AJUSTES_INICIALES, servicios: [], profesionales: [], turnos: [] };
	const ajustes: AjustesTurnos = { ...AJUSTES_INICIALES, activo: true, horario: HORARIO_TIJERA };
	const hoy = diaAr(base);
	const turnos: Turno[] = [];
	EJEMPLOS.forEach(([d, hora, pid, sid], i) => {
		const dia = sumarDias(hoy, d);
		const s = SERVICIOS_TIJERA.find((x) => x.id === sid)!;
		const p = EQUIPO_TIJERA.find((x) => x.id === pid)!;
		const m = minutos(hora);
		// Sólo los que entran en el horario de ese día (el domingo está cerrado).
		const entra = HORARIO_TIJERA.some(
			(f) => f.dia === diaSemana(dia) && minutos(f.desde) <= m && m + s.duracion <= minutos(f.hasta)
		);
		if (!entra) return;
		const inicio = instante(dia, m);
		const fin = inicio + s.duracion * MIN;
		const [nombre, telefono] = CLIENTES[i % CLIENTES.length];
		const conSena = i % 3 === 0;
		turnos.push({
			id: `ej-${i + 1}`,
			codigo: `ejemplo${hash(`${base}:${i}`).toString(36)}`.padEnd(16, '0'),
			servicioId: s.id,
			servicio: s.nombre,
			profesionalId: p.id,
			profesional: p.nombre,
			inicio: isoAr(inicio),
			fin: isoAr(fin),
			cliente: {
				nombre,
				telefono,
				email: `${nombre.split(' ')[0].toLowerCase().normalize('NFD').replace(/[̀-ͯ]/g, '')}@ejemplo.com`
			},
			precio: s.precio,
			sena: conSena ? senaDe(s.precio, ajustes.sena) : 0,
			estado: fin < base ? 'atendido' : 'confirmado',
			senaPagada: conSena,
			senaDevuelta: false,
			origen: i % 4 === 3 ? 'panel' : 'tienda',
			nota: null,
			creadoEn: new Date(base - (2 + (i % 5)) * 86_400_000).toISOString(),
			venceEn: null,
			vencido: false
		});
	});
	return { ajustes, servicios: SERVICIOS_TIJERA, profesionales: EQUIPO_TIJERA, turnos };
}
