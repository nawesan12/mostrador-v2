/**
 * Backoffice de ejemplo para la demo (DATA_SOURCE=sandbox): tiendas, consultas, postulaciones y
 * auditoría generadas con una semilla fija, relativas al día de hoy. No toca D1 ni las tiendas de
 * los visitantes, y ninguna acción se aplica (las páginas muestran "en la demo no se aplica").
 */
import rubros from '@mostrador/contrato/rubros.json';
import { PLANTILLAS } from '#lib/plantillas.ts';
import { PERMISOS, type Rol } from '#lib/staff/permisos.ts';
import {
	abonoMensual,
	DIAS_PRUEBA,
	estadoDe,
	plazoArrepentimiento,
	type EstadoTienda,
	type Plan
} from '#lib/staff/estado.ts';
import { listarGenerico, type DefLista, type FiltrosLista } from '../sandbox/listas.ts';
import type * as T from './tipos.ts';

const DIA = 86_400_000;
const HORA = 3_600_000;

/** Error de las escrituras en la demo: las páginas lo muestran como toast. */
export class EnLaDemo extends Error {
	constructor() {
		super('En la demo no se aplica');
	}
}

/** PRNG chico y estable (mulberry32). */
function azar(semilla: number) {
	let a = semilla >>> 0;
	return () => {
		a = (a + 0x6d2b79f5) >>> 0;
		let t = a;
		t = Math.imul(t ^ (t >>> 15), t | 1);
		t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
		return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
	};
}

export const MIEMBROS_DEMO: { id: string; nombre: string; email: string; rol: Rol }[] = [
	{ id: 'm-lucia', nombre: 'Lucía Ferreyra', email: 'lucia@tulocalweb.com.ar', rol: 'duenio' },
	{
		id: 'm-tomas',
		nombre: 'Tomás Ibarra',
		email: 'tomas@tulocalweb.com.ar',
		rol: 'administracion'
	},
	{ id: 'm-camila', nombre: 'Camila Ríos', email: 'camila@tulocalweb.com.ar', rol: 'soporte' },
	{ id: 'm-joaquin', nombre: 'Joaquín Paz', email: 'joaquin@tulocalweb.com.ar', rol: 'soporte' }
];

const PRIMERO = [
	'Casa',
	'Taller',
	'Almacén',
	'Estudio',
	'Tienda',
	'Rincón',
	'Mercado',
	'Club',
	'Patio',
	'Feria',
	'Cocina',
	'Atelier'
];
const SEGUNDO = [
	'Lumbre',
	'Ceibo',
	'Bruma',
	'Nido',
	'Arcilla',
	'Pampa',
	'Ombú',
	'Jacarandá',
	'Quebracho',
	'Malva',
	'Totora',
	'Lapacho',
	'Aromo',
	'Tala',
	'Calden',
	'Mburucuyá',
	'Yatay',
	'Espinillo'
];
const NOMBRES = [
	'Valeria',
	'Martina',
	'Sofía',
	'Florencia',
	'Julieta',
	'Agustina',
	'Carla',
	'Paula',
	'Romina',
	'Nicolás',
	'Federico',
	'Matías',
	'Gonzalo',
	'Lucas',
	'Mariana',
	'Daniela'
];
const APELLIDOS = [
	'Gómez',
	'Rodríguez',
	'Fernández',
	'López',
	'Martínez',
	'Sosa',
	'Romero',
	'Acosta',
	'Benítez',
	'Medina',
	'Herrera',
	'Aguirre',
	'Giménez',
	'Molina',
	'Castro',
	'Ortiz'
];
const CIUDADES = [
	'Córdoba',
	'Rosario',
	'Mendoza',
	'La Plata',
	'Mar del Plata',
	'Salta',
	'Neuquén',
	'San Miguel de Tucumán',
	'Paraná',
	'Bahía Blanca'
];
const PARTNERS = ['pampacode', 'estudio-nube', 'sur-digital', 'contadora-lopez'];
const MOTIVOS_BAJA = [
	'Cerró el local y no vende más online',
	'Volvió a Tiendanube por las integraciones que ya tenía',
	'El costo no le cerraba con las ventas de la temporada',
	'Pasó a vender sólo por Mercado Libre'
];

type Tienda = {
	id: string;
	nombre: string;
	slug: string;
	titular: string;
	email: string;
	plan: Plan;
	ciclo: 'mes' | 'anio';
	rubro: string;
	plantilla: string;
	partner: string;
	creadaEn: number;
	pruebaHasta: number | null;
	canceladaEn: number | null;
	cerradaEn: number | null;
	pausadaEn: number | null;
	suscripcion: string | null;
	motivo: string | null;
	credito: number;
	ciudad: string;
	productos: number;
	pedidosMes: number;
	packs: number;
	cuit: string;
};

const slugDe = (s: string) =>
	s
		.normalize('NFD')
		.replace(/[̀-ͯ]/g, '')
		.toLowerCase()
		.replace(/[^a-z0-9]+/g, '-')
		.replace(/^-|-$/g, '');

/** Las tiendas de ejemplo de hoy (cambian de día en día, iguales para todos los pedidos). */
export function tiendasDemo(ahora: number): Tienda[] {
	const r = azar(76);
	const hoy = Math.floor(ahora / DIA) * DIA + 15 * HORA;
	const usados = new Set<string>();
	const out: Tienda[] = [];
	for (let i = 0; i < 168; i++) {
		let nombre: string;
		do {
			nombre = `${PRIMERO[Math.floor(r() * PRIMERO.length)]} ${SEGUNDO[Math.floor(r() * SEGUNDO.length)]}`;
		} while (usados.has(nombre) && usados.size < PRIMERO.length * SEGUNDO.length);
		usados.add(nombre);
		const rubro = rubros[Math.floor(r() * rubros.length)];
		const delRubro = PLANTILLAS.filter((p) => p.rubro === rubro.nombre);
		const plantilla = delRubro.length
			? delRubro[Math.floor(r() * delRubro.length)].id
			: rubro.plantilla;
		// Antigüedad: muchas nuevas, una cola larga de tiendas viejas.
		const dias = Math.floor(Math.pow(r(), 1.8) * 420);
		const creadaEn = hoy - dias * DIA - Math.floor(r() * 10) * HORA;
		const x = r();
		const plan: Plan = x < 0.22 ? 'gratis' : x < 0.58 ? 'emprende' : x < 0.9 ? 'pro' : 'negocio';
		const nom = NOMBRES[Math.floor(r() * NOMBRES.length)];
		const ape = APELLIDOS[Math.floor(r() * APELLIDOS.length)];
		const y = r();
		const cerrada = dias > 40 && y < 0.04;
		const cancelada = !cerrada && dias > 30 && y < 0.09;
		const pausada = !cerrada && !cancelada && dias > 30 && y < 0.11;
		const rechazado = plan !== 'gratis' && dias > 20 && y > 0.95;
		const t: Tienda = {
			id: `demo-${String(i + 1).padStart(3, '0')}`,
			nombre,
			slug: slugDe(nombre),
			titular: `${nom} ${ape}`,
			email: `${slugDe(nom)}.${slugDe(ape)}@gmail.com`,
			plan,
			ciclo: r() < 0.28 ? 'anio' : 'mes',
			rubro: rubro.id,
			plantilla,
			partner: r() < 0.2 ? PARTNERS[Math.floor(r() * PARTNERS.length)] : '',
			creadaEn,
			pruebaHasta: r() < 0.06 ? creadaEn + (DIAS_PRUEBA + 7) * DIA : null,
			canceladaEn: cancelada || cerrada ? hoy - Math.floor(r() * 25) * DIA : null,
			cerradaEn: cerrada ? hoy - Math.floor(r() * 25) * DIA : null,
			pausadaEn: pausada ? hoy - Math.floor(r() * 12) * DIA : null,
			suscripcion: plan === 'gratis' ? null : rechazado ? 'paused' : 'authorized',
			motivo: cancelada || cerrada ? MOTIVOS_BAJA[Math.floor(r() * MOTIVOS_BAJA.length)] : null,
			credito: r() < 0.05 ? Math.round(r() * 20) * 100_000 : 0,
			ciudad: CIUDADES[Math.floor(r() * CIUDADES.length)],
			productos: Math.floor(8 + r() * 420),
			pedidosMes: Math.floor(r() * 260),
			packs: r() < 0.08 ? 1 : 0,
			cuit: `27-${Math.floor(20_000_000 + r() * 19_999_999)}-${Math.floor(r() * 10)}`
		};
		out.push(t);
	}
	return out;
}

const estado = (t: Tienda, ahora: number): EstadoTienda => estadoDe({ ...t, ahora });
const mrrDe = (t: Tienda, ahora: number) =>
	estado(t, ahora) === 'activa' ? abonoMensual(t.plan, t.ciclo) : 0;
const pruebaHasta = (t: Tienda) => t.pruebaHasta ?? t.creadaEn + DIAS_PRUEBA * DIA;
const iso = (ms: number) => new Date(ms).toISOString();
const dia = (ms: number) =>
	new Intl.DateTimeFormat('en-CA', { timeZone: 'America/Argentina/Buenos_Aires' }).format(ms);

function proximoCobro(t: Tienda, ahora: number) {
	if (t.plan === 'gratis' || t.cerradaEn || t.canceladaEn) return null;
	const d = new Date(pruebaHasta(t));
	while (d.getTime() <= ahora) d.setUTCMonth(d.getUTCMonth() + (t.ciclo === 'anio' ? 12 : 1));
	return d.getTime();
}

const DEF_TIENDAS: DefLista<T.FilaTienda> = {
	buscar: (t) => `${t.nombre} ${t.slug} ${t.titular} ${t.email}`,
	segmento: 'estado',
	facetas: ['plan', 'rubro', 'plantilla', 'partner'],
	rango: { clave: 'dias', menor: true },
	banderas: [],
	ordenes: {
		recientes: (a, b) => b.creadaEn.localeCompare(a.creadaEn),
		nombre: (a, b) => a.nombre.localeCompare(b.nombre, 'es'),
		mrr: (a, b) => b.mrr - a.mrr
	},
	resumen: (_, l) => ({
		mrr: l.reduce((s, t) => s + t.mrr, 0),
		prueba: l.filter((t) => t.estado === 'prueba').length,
		activas: l.filter((t) => t.estado === 'activa').length
	})
};

function consultasDemo(ahora: number): (T.ConsultaDetalle & { documento?: string })[] {
	const h = (horas: number) => iso(ahora - horas * HORA);
	const contacto = (
		id: string,
		nombre: string,
		email: string,
		horas: number,
		estado: T.EstadoConsulta,
		mensaje: string,
		extra: Partial<T.ConsultaEquipo> = {},
		asignadoA: string | null = null
	): T.ConsultaDetalle => ({
		id,
		tipo: 'contacto',
		estado,
		nombre,
		email,
		asunto: 'Consulta de ventas',
		mensaje,
		creadaEn: h(horas),
		asignadoA,
		whatsapp: null,
		web: null,
		facturacion: null,
		origen: null,
		...extra,
		respuestas: []
	});
	const arr = (
		id: string,
		codigo: string,
		nombre: string,
		email: string,
		horas: number,
		estado: T.EstadoConsulta,
		plan: string,
		mensaje: string
	): T.ConsultaDetalle => ({
		id,
		tipo: 'arrepentimiento',
		estado,
		nombre,
		email,
		asunto: `Arrepentimiento ${codigo}`,
		mensaje,
		creadaEn: h(horas),
		asignadoA: estado === 'nueva' ? null : 'm-tomas',
		codigo,
		plan,
		plazo: plazoArrepentimiento(ahora - horas * HORA),
		documento: '30987654',
		respuestas: []
	});
	const anio = new Date(ahora).getUTCFullYear();
	const lista = [
		contacto(
			'c-01',
			'Rocío Benítez',
			'rocio@mielesdelsur.com.ar',
			2,
			'nueva',
			'Tengo 380 productos en Tiendanube. ¿Me ayudan a pasar todo, con fotos y variantes? Vendo unos 120 pedidos por mes.',
			{ origen: 'tiendanube', facturacion: '5-a-20', whatsapp: '5493415551234' }
		),
		contacto(
			'c-02',
			'Gustavo Peralta',
			'gperalta@ferreteriaelsol.com',
			7,
			'nueva',
			'¿El plan Negocio incluye varias sucursales con stock separado? Tenemos tres locales en Mendoza.',
			{ facturacion: '20-a-100', web: 'ferreteriaelsol.com' }
		),
		contacto(
			'c-03',
			'Ana Quiroga',
			'ana.quiroga@gmail.com',
			20,
			'en_curso',
			'Quiero saber si puedo facturar con factura A desde la tienda y cómo se conecta con ARCA.',
			{ facturacion: 'hasta-5' },
			'm-camila'
		),
		contacto(
			'c-04',
			'Leandro Vega',
			'leandro@cervezaandina.com.ar',
			30,
			'en_curso',
			'Vendemos cerveza artesanal: ¿cómo manejan la verificación de edad para bebidas con alcohol?',
			{ origen: 'empretienda', web: 'cervezaandina.com.ar' },
			'm-joaquin'
		),
		contacto(
			'c-05',
			'Micaela Torres',
			'mica@estudiotorres.com',
			52,
			'resuelta',
			'¿Tienen descuento para ONG? Somos una cooperativa textil.',
			{},
			'm-camila'
		),
		contacto(
			'c-06',
			'Pablo Duarte',
			'pablo.duarte@outlook.com',
			75,
			'resuelta',
			'¿Se puede usar Tu Local Web sólo para el local, sin tienda online?',
			{ facturacion: 'empiezo' },
			'm-joaquin'
		),
		arr(
			'a-01',
			`ARR-${anio}-100214`,
			'Silvina Cáceres',
			'silvina.caceres@gmail.com',
			26,
			'nueva',
			'Pro mensual',
			'Contraté el plan equivocado: quería Emprende.'
		),
		arr(
			'a-02',
			`ARR-${anio}-100198`,
			'Diego Funes',
			'dfunes@hotmail.com',
			24 * 9,
			'en_curso',
			'Emprende anual',
			'Al final no voy a abrir la tienda este año.'
		),
		arr(
			'a-03',
			`ARR-${anio}-100151`,
			'Lorena Paz',
			'lorepaz@gmail.com',
			24 * 21,
			'resuelta',
			'Pro mensual',
			''
		)
	];
	lista[2].respuestas = [
		{
			id: 'r-01',
			autor: 'Camila Ríos',
			asunto: 'Re: Factura A desde tu tienda',
			texto:
				'¡Hola, Ana! Sí: desde el plan Emprende la tienda emite factura A y B con tu certificado de ARCA. Te paso la guía paso a paso.',
			creadaEn: h(18)
		}
	];
	return lista;
}

const POSTULACIONES = (ahora: number): T.PostulacionEquipo[] => [
	{
		id: 'p-01',
		nombre: 'Ana Ledesma',
		email: 'ana@pampacode.com.ar',
		agencia: 'Pampa Code',
		web: 'pampacode.com.ar',
		especialidades: ['desarrollo', 'diseno'],
		estado: 'nueva',
		motivo: null,
		creadaEn: iso(ahora - 5 * HORA)
	},
	{
		id: 'p-02',
		nombre: 'Martín Suárez',
		email: 'martin@lentefija.com',
		agencia: 'Lente Fija',
		web: 'lentefija.com',
		especialidades: ['fotografia'],
		estado: 'nueva',
		motivo: null,
		creadaEn: iso(ahora - 30 * HORA)
	},
	{
		id: 'p-03',
		nombre: 'Carolina López',
		email: 'estudio@contadoralopez.com.ar',
		agencia: 'Estudio López',
		web: null,
		especialidades: ['contabilidad'],
		estado: 'nueva',
		motivo: null,
		creadaEn: iso(ahora - 3 * DIA)
	},
	{
		id: 'p-04',
		nombre: 'Federico Albornoz',
		email: 'fede@surdigital.agency',
		agencia: 'Sur Digital',
		web: 'surdigital.agency',
		especialidades: ['marketing', 'diseno'],
		estado: 'aprobada',
		motivo: null,
		creadaEn: iso(ahora - 40 * DIA)
	},
	{
		id: 'p-05',
		nombre: 'Javier Molina',
		email: 'javi.molina@gmail.com',
		agencia: 'Freelance',
		web: null,
		especialidades: ['desarrollo'],
		estado: 'rechazada',
		motivo: 'Sin trabajos para mostrar todavía: lo invitamos a postularse de nuevo en seis meses.',
		creadaEn: iso(ahora - 22 * DIA)
	}
];

const NOMBRES_PARTNER: Record<string, string> = {
	pampacode: 'Pampa Code',
	'estudio-nube': 'Estudio Nube',
	'sur-digital': 'Sur Digital',
	'contadora-lopez': 'Estudio López'
};

/** Qué acciones registró el equipo en los últimos días (la auditoría de la demo). */
function auditoriaDemo(ts: Tienda[], ahora: number): T.AuditoriaEquipo['registros'] {
	const r = azar(9);
	const m = (id: string) => {
		const x = MIEMBROS_DEMO.find((y) => y.id === id)!;
		return { id: x.id, nombre: x.nombre, rol: x.rol };
	};
	const plantillas: [string, string, (t: Tienda) => string][] = [
		['ver_ficha', 'm-camila', (t) => `Vio la ficha de ${t.nombre}`],
		['mirar_tienda', 'm-camila', (t) => `Entró a ver la tienda ${t.nombre}`],
		['extender_prueba', 'm-joaquin', (t) => `Extendió la prueba de ${t.nombre} 7 días`],
		['regalar_pack', 'm-camila', (t) => `Regaló un pack de Hornero a ${t.nombre}`],
		['nota', 'm-joaquin', (t) => `Dejó una nota en ${t.nombre}`],
		['cambiar_plan', 'm-tomas', (t) => `Cambió el plan de ${t.nombre} a Pro`],
		['aplicar_credito', 'm-tomas', (t) => `Aplicó $ 19.900 de crédito a ${t.nombre}`],
		['mirar_tienda', 'm-tomas', (t) => `Entró a ver la tienda ${t.nombre}`],
		['pausar', 'm-lucia', (t) => `Pausó ${t.nombre}`]
	];
	const out: T.AuditoriaEquipo['registros'] = [];
	let cuando = ahora - 40 * 60_000;
	for (let i = 0; i < 46; i++) {
		const [accion, quien, texto] = plantillas[Math.floor(r() * plantillas.length)];
		const t = ts[Math.floor(r() * ts.length)];
		out.push({
			id: String(5000 - i),
			cuando: iso(cuando),
			miembro: m(quien),
			accion,
			texto: texto(t),
			tiendaId: t.id,
			tienda: t.nombre
		});
		cuando -= Math.floor((1 + r() * 9) * HORA);
	}
	out.splice(3, 0, {
		id: '5100',
		cuando: iso(ahora - 3 * HORA),
		miembro: m('m-tomas'),
		accion: 'partner_aprobar',
		texto: 'Aprobó a Sur Digital como partner',
		tiendaId: null,
		tienda: null
	});
	out.splice(6, 0, {
		id: '5101',
		cuando: iso(ahora - 6 * HORA),
		miembro: m('m-camila'),
		accion: 'responder',
		texto: 'Respondió la consulta de Ana Quiroga',
		tiendaId: null,
		tienda: null
	});
	return out.sort((a, b) => b.cuando.localeCompare(a.cuando));
}

export class EquipoDemo implements T.DatosEquipo {
	readonly demo = true;
	private ts: Tienda[];
	constructor(
		private rol: Rol,
		private ahora = Date.now()
	) {
		this.ts = tiendasDemo(ahora);
	}

	private fila(t: Tienda): T.FilaTienda {
		return {
			id: t.id,
			nombre: t.nombre,
			slug: t.slug,
			titular: t.titular,
			email: t.email,
			plan: t.plan,
			ciclo: t.ciclo,
			estado: estado(t, this.ahora),
			rubro: t.rubro,
			plantilla: t.plantilla,
			partner: t.partner,
			creadaEn: iso(t.creadaEn),
			dias: Math.floor((this.ahora - t.creadaEn) / DIA),
			mrr: mrrDe(t, this.ahora)
		};
	}

	async sesion(): Promise<T.SesionEquipo> {
		const m = MIEMBROS_DEMO.find((x) => x.rol === this.rol)!;
		return {
			miembro: { id: m.id, nombre: m.nombre, email: m.email },
			rol: this.rol,
			permisos: PERMISOS[this.rol],
			venceEn: iso(this.ahora + 12 * HORA)
		};
	}

	async inicio(): Promise<T.InicioEquipo> {
		const a = this.ahora;
		const est = this.ts.map((t) => [t, estado(t, a)] as const);
		const pruebas = est
			.filter(([t, e]) => e === 'prueba' && pruebaHasta(t) - a <= 3 * DIA)
			.sort(([x], [y]) => pruebaHasta(x) - pruebaHasta(y));
		const mesHoy = dia(a).slice(0, 7);
		const bajas = this.ts.filter((t) => t.canceladaEn && dia(t.canceladaEn).slice(0, 7) === mesHoy);
		const consultas = consultasDemo(a);
		const alertas: T.InicioEquipo['alertas'] = [
			...est
				.filter(([, e]) => e === 'cobro_rechazado')
				.slice(0, 3)
				.map(([t]) => ({
					tipo: 'cobro_rechazado' as const,
					titulo: `Cobro rechazado: ${t.nombre}`,
					detalle: `Mercado Pago pausó el abono ${t.plan === 'negocio' ? 'Negocio' : t.plan === 'pro' ? 'Pro' : 'Emprende'}. Escribile a ${t.titular.split(' ')[0]} para actualizar la tarjeta.`,
					tiendaId: t.id,
					consulta: null
				})),
			{
				tipo: 'integracion',
				titulo: `Mercado Pago para reconectar: ${this.ts[4].nombre}`,
				detalle: 'El permiso venció y no pudimos renovarlo. Hasta que lo conecte, no cobra online.',
				tiendaId: this.ts[4].id,
				consulta: null
			},
			{
				tipo: 'arrepentimiento',
				titulo: `Devolución por vencer: ${consultas[7].codigo}`,
				detalle: `${consultas[7].nombre} pidió la baja. El plazo legal para devolver vence el ${consultas[7].plazo?.split('-').reverse().join('/')}.`,
				tiendaId: null,
				consulta: `arrepentimiento/${consultas[7].id}`
			}
		];
		return {
			kpis: {
				tiendasActivas: est.filter(([, e]) => e !== 'cerrada' && e !== 'pausada').length,
				nuevasSemana: this.ts.filter((t) => a - t.creadaEn < 7 * DIA).length,
				pruebasPorVencer: pruebas.length,
				mrr: this.ts.reduce((s, t) => s + mrrDe(t, a), 0),
				bajasMes: bajas.length,
				consultasSinResponder: consultas.filter((c) => c.estado === 'nueva').length,
				postulacionesNuevas: POSTULACIONES(a).filter((p) => p.estado === 'nueva').length
			},
			pruebas: pruebas.slice(0, 8).map(([t]) => ({
				id: t.id,
				nombre: t.nombre,
				slug: t.slug,
				fecha: dia(pruebaHasta(t))
			})),
			bajas: bajas.slice(0, 8).map((t) => ({
				id: t.id,
				nombre: t.nombre,
				slug: t.slug,
				fecha: dia(t.canceladaEn!),
				motivo: t.motivo
			})),
			alertas
		};
	}

	async tiendas(f: FiltrosLista) {
		return listarGenerico(
			this.ts.map((t) => this.fila(t)),
			DEF_TIENDAS,
			f
		);
	}

	async ficha(id: string): Promise<T.FichaTienda | null> {
		const t = this.ts.find((x) => x.id === id);
		if (!t) return null;
		const a = this.ahora;
		const r = azar(Number(id.slice(5)) || 1);
		const paga = t.plan !== 'gratis';
		return {
			tienda: {
				id: t.id,
				nombre: t.nombre,
				slug: t.slug,
				rubro: t.rubro,
				plantilla: t.plantilla,
				creadaEn: iso(t.creadaEn),
				estado: estado(t, a),
				partner: t.partner,
				motivoBaja: t.motivo
			},
			negocio: {
				razonSocial: t.titular,
				cuit: t.cuit,
				condicion: 'Monotributo',
				telefono: `+54 9 ${Math.floor(r() * 9000 + 1000)} ${Math.floor(r() * 900000 + 100000)}`,
				direccion: `${t.ciudad}`
			},
			titular: { nombre: t.titular, email: t.email, desde: iso(t.creadaEn) },
			personas: t.plan === 'gratis' ? 1 : 1 + Math.floor(r() * 3),
			plan: {
				actual: t.plan,
				ciclo: t.ciclo,
				precio: abonoMensual(t.plan, t.ciclo),
				enPrueba: estado(t, a) === 'prueba',
				pruebaHasta: iso(pruebaHasta(t)),
				proximoCobro: proximoCobro(t, a) ? iso(proximoCobro(t, a)!) : null,
				suscripcion: t.suscripcion,
				credito: t.credito
			},
			uso: { productos: t.productos, pedidosMes: t.pedidosMes, packsMes: t.packs },
			integraciones: [
				...(paga
					? [
							{
								id: 'mercadopago',
								nombre: 'Mercado Pago',
								salud: (t.id === this.ts[4].id ? 'reconectar' : 'ok') as 'ok' | 'reconectar'
							}
						]
					: []),
				...(r() < 0.5
					? [{ id: 'mercadolibre', nombre: 'Mercado Libre', salud: 'ok' as const }]
					: []),
				...(r() < 0.4 ? [{ id: 'ga4', nombre: 'Google Analytics', salud: 'ok' as const }] : [])
			],
			dominios:
				paga && r() < 0.6
					? [
							{
								nombre: `${t.slug.replace(/-/g, '')}.com.ar`,
								verificado: r() < 0.8,
								principal: true
							}
						]
					: [],
			actividad: [
				['persona', 'Publicó 3 productos nuevos', 2],
				['hornero', 'Respondió 5 preguntas de Mercado Libre', 6],
				['sistema', 'Cobró el abono del mes', 30],
				['persona', 'Cambió los colores de la tienda', 52],
				['integracion', 'Llegó un pedido desde Mercado Libre', 75]
			].map(([quien, accion, horas], i) => ({
				id: `${t.id}-a${i}`,
				cuando: iso(a - Number(horas) * HORA),
				quien: quien as 'persona',
				accion: String(accion)
			})),
			notas:
				r() < 0.5
					? [
							{
								id: `${t.id}-n1`,
								texto: `${t.titular.split(' ')[0]} prefiere que la llamen después de las 16. Vende sobre todo por Instagram.`,
								autor: 'Camila Ríos',
								creadaEn: iso(a - 4 * DIA)
							}
						]
					: []
		};
	}

	async consultas(): Promise<T.BandejaEquipo> {
		return {
			consultas: consultasDemo(this.ahora).map((c) => {
				const { respuestas, documento, ...resto } = c;
				void respuestas;
				void documento;
				return resto;
			}),
			miembros: MIEMBROS_DEMO.map(({ id, nombre, rol }) => ({ id, nombre, rol }))
		};
	}

	async consulta(tipo: T.TipoConsulta, id: string) {
		return consultasDemo(this.ahora).find((c) => c.tipo === tipo && c.id === id) ?? null;
	}

	async partners(): Promise<T.PartnersEquipo> {
		const activos = PARTNERS.map((codigo) => {
			const suyas = this.ts.filter((t) => t.partner === codigo);
			const activas = suyas.filter((t) => mrrDe(t, this.ahora) > 0);
			const nivel = activas.length >= 20 ? 'oro' : activas.length >= 5 ? 'plata' : 'partner';
			const pct = nivel === 'oro' ? 0.3 : nivel === 'plata' ? 0.25 : 0.2;
			return {
				codigo,
				nombre: NOMBRES_PARTNER[codigo],
				nivel: nivel as 'partner',
				tiendas: suyas.length,
				activas: activas.length,
				comision: Math.round(activas.reduce((s, t) => s + mrrDe(t, this.ahora), 0) * pct)
			};
		}).sort((a, b) => b.comision - a.comision);
		return { postulaciones: POSTULACIONES(this.ahora), activos };
	}

	async facturacion(): Promise<T.FacturacionEquipo> {
		const a = this.ahora;
		const subs = this.ts
			.filter((t) => t.plan !== 'gratis' && !t.cerradaEn)
			.map((t) => {
				const p = proximoCobro(t, a);
				return {
					tiendaId: t.id,
					nombre: t.nombre,
					plan: t.plan as 'pro',
					ciclo: t.ciclo,
					monto: abonoMensual(t.plan, t.ciclo) * (t.ciclo === 'anio' ? 12 : 1),
					estado: estado(t, a),
					proximoCobro: p ? iso(p) : null,
					credito: t.credito
				};
			})
			.sort((x, y) => (x.proximoCobro ?? 'z').localeCompare(y.proximoCobro ?? 'z'));
		const finMes = dia(a).slice(0, 7);
		return {
			kpis: {
				mrr: this.ts.reduce((s, t) => s + mrrDe(t, a), 0),
				cobrosMes: subs
					.filter(
						(s) =>
							s.proximoCobro &&
							dia(Date.parse(s.proximoCobro)).slice(0, 7) === finMes &&
							s.estado === 'activa'
					)
					.reduce((x, s) => x + s.monto, 0),
				activas: subs.filter((s) => s.estado === 'activa').length,
				conProblemas: subs.filter((s) => s.estado === 'cobro_rechazado').length,
				creditos: subs.reduce((x, s) => x + s.credito, 0)
			},
			suscripciones: subs,
			comprobantes: []
		};
	}

	async auditoria(f: T.FiltrosAuditoria): Promise<T.AuditoriaEquipo> {
		const todos = auditoriaDemo(this.ts, this.ahora);
		return {
			registros: todos.filter(
				(x) =>
					(!f.miembro || x.miembro.id === f.miembro) &&
					(!f.accion || x.accion === f.accion) &&
					(!f.tienda || x.tiendaId === f.tienda)
			),
			miembros: MIEMBROS_DEMO.map(({ id, nombre, rol }) => ({ id, nombre, rol }))
		};
	}

	// Ninguna escritura se aplica en la demo.
	async accion(): Promise<void> {
		throw new EnLaDemo();
	}
	async nota(): Promise<T.NotaEquipo> {
		throw new EnLaDemo();
	}
	async mirar(): Promise<T.SesionNueva> {
		throw new EnLaDemo();
	}
	async actualizarConsulta(): Promise<void> {
		throw new EnLaDemo();
	}
	async responder(): Promise<T.RespuestaEquipo> {
		throw new EnLaDemo();
	}
	async decidir(): Promise<void> {
		throw new EnLaDemo();
	}
}
