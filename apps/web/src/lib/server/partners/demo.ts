/**
 * El panel de partners de la demo (DATA_SOURCE=sandbox): Pampa Code, Partner Plata de Rosario, el
 * que aparece en el directorio de /partners. Como el resto del sandbox, el ejemplo sale del código
 * (las mismas tiendas que `partners.Sembrar` en Go) y en D1 sólo va lo que cambia el visitante
 * (perfil, datos de cobro y facturas, entidad `partner` en `cambios`).
 */
import {
	comision,
	estado,
	fechaDePago,
	finComision,
	mesDe,
	nivelEfectivo,
	type CobroPartner,
	type PagoPartner,
	type PanelPartner,
	type PerfilPartner,
	type TiendaReferida
} from '#lib/partners/programa.ts';
import { cambiosDe, guardarCambio, type DbSandbox, type Sandbox } from '../sandbox/sandbox.ts';
import { slugDe } from '../sandbox/rubros.ts';
import type { FuentePartners } from './fuente.ts';

const ENTIDAD = 'partner';

/** Nombre, plan, ciclo, hace cuántos días se dio de alta y si se dio de baja. */
export const TIENDAS_DEMO: [string, TiendaReferida['plan'], 'mes' | 'anio', number, boolean][] = [
	['Almacén Verde', 'pro', 'mes', 320, false],
	['Café Bruma', 'pro', 'mes', 290, false],
	['Rosa Mosqueta', 'emprende', 'mes', 260, false],
	['Casa Nido', 'pro', 'anio', 240, false],
	['Sonido Sur', 'negocio', 'mes', 210, false],
	['Vinoteca Tres', 'pro', 'mes', 180, false],
	['Huellitas', 'emprende', 'mes', 150, false],
	['Madera Viva', 'pro', 'mes', 130, false],
	['Relojería Norte', 'emprende', 'anio', 110, false],
	['Mate & Co', 'pro', 'mes', 90, false],
	['Estudio Arcilla', 'pro', 'mes', 70, false],
	['Óptica Faro', 'emprende', 'mes', 50, false],
	['Navaja Barbería', 'pro', 'mes', 40, false],
	['Panadería Lola', 'emprende', 'mes', 30, false],
	['Tienda Lumbre', 'gratis', 'mes', 6, false],
	['PMPA Store', 'gratis', 'mes', 3, false],
	['Kiosco Central', 'gratis', 'mes', 11, false],
	['Librería Faro', 'gratis', 'mes', 400, false],
	['Dulce Taller', 'emprende', 'mes', 200, true]
];

export const PERFIL_DEMO: PerfilPartner = {
	nombre: 'Pampa Code',
	ciudad: 'Rosario',
	descripcion:
		'Armamos tiendas para marcas del Litoral: integraciones, migraciones y fotos de producto.',
	especialidades: ['desarrollo', 'diseno', 'fotografia'],
	casos: [
		{ nombre: 'Café Bruma', url: 'https://tulocalweb.com.ar/t/cafe-bruma' },
		{ nombre: 'Casa Nido', url: 'https://tulocalweb.com.ar/t/casa-nido' }
	],
	email: 'hola@pampacode.com.ar',
	whatsapp: '5493415550123',
	web: 'https://pampacode.com.ar',
	logo: ''
};

const DIA = 86_400_000;
const fechaAr = (ms: number) => new Date(ms - 3 * 3_600_000).toISOString().slice(0, 10);
function mesAnterior(mes: string, n: number) {
	const [a, m] = mes.split('-').map(Number);
	const t = a * 12 + (m - 1) - n;
	return `${Math.floor(t / 12)}-${String((t % 12) + 1).padStart(2, '0')}`;
}

/** Arma el panel de ejemplo a una hora, con lo que cambió el visitante. */
export function panelDemo(ahora: number, cambios: Map<string, unknown> = new Map()): PanelPartner {
	const filas = TIENDAS_DEMO.map(([nombre, plan, ciclo, dias, baja]) => {
		const alta = new Date(ahora - dias * DIA).toISOString();
		return { nombre, plan, ciclo, alta, est: estado(plan, alta, baja, ahora) };
	});
	const nivel = nivelEfectivo('plata', filas.filter((f) => f.est === 'activa').length);
	const comisionDe = (f: (typeof filas)[number], mes: string) =>
		f.est === 'activa' ? comision(nivel, f.plan, f.ciclo, f.alta, mes) : 0;
	const mes = mesDe(ahora);
	const tiendas: TiendaReferida[] = filas.map((f) => ({
		id: slugDe(f.nombre),
		nombre: f.nombre,
		slug: slugDe(f.nombre),
		plan: f.plan,
		ciclo: f.ciclo,
		estado: f.est,
		alta: fechaAr(Date.parse(f.alta)),
		finComision: finComision(f.alta),
		comisionMes: comisionDe(f, mes)
	}));
	// Los tres meses anteriores: el último aprobado (falta la factura), los otros dos pagados.
	const pagos: PagoPartner[] = (['aprobada', 'pagada', 'pagada'] as const).map((est, i) => {
		const m = mesAnterior(mes, i + 1);
		const detalle = filas
			.map((f) => ({
				tiendaId: slugDe(f.nombre),
				tienda: f.nombre,
				plan: f.plan,
				comision: comisionDe(f, m)
			}))
			.filter((l) => l.comision > 0);
		const subida = cambios.get(`factura:${m}`) as string | undefined;
		return {
			mes: m,
			estado: est,
			monto: detalle.reduce((s, l) => s + l.comision, 0),
			detalle,
			facturaUrl:
				est === 'pagada'
					? `https://drive.google.com/file/d/factura-pampa-code-${m}`
					: (subida ?? null),
			fechaPago: fechaDePago(m),
			pagadoEn: est === 'pagada' ? fechaDePago(m) : null
		};
	});
	return {
		hoy: fechaAr(ahora),
		partner: {
			nombre: 'Ana Ferreyra',
			email: 'ana@pampacode.com.ar',
			nivel: 'plata',
			nivelEfectivo: nivel,
			porcentaje: nivel === 'oro' ? 30 : nivel === 'plata' ? 25 : 20,
			codigo: 'pampa-code',
			cupon: 'PAMPACODE20',
			aprobadoEn: fechaAr(ahora - 330 * DIA),
			perfil: (cambios.get('perfil') as PerfilPartner | undefined) ?? PERFIL_DEMO,
			cobro: (cambios.get('cobro') as CobroPartner | undefined) ?? null
		},
		tiendas,
		clics: [212, 187, 164].map((n, i) => ({ mes: mesAnterior(mes, i), n })),
		pagos,
		colaboraciones: [
			{ tiendaId: 'cafe-bruma', nombre: 'Café Bruma', slug: 'cafe-bruma', rol: 'admin' },
			{ tiendaId: 'casa-nido', nombre: 'Casa Nido', slug: 'casa-nido', rol: 'ventas' }
		]
	};
}

export class DemoPartners implements FuentePartners {
	constructor(
		private db: DbSandbox,
		private sb: Sandbox,
		private ahora = Date.now()
	) {}

	private cambios() {
		return cambiosDe(this.db, this.sb.id, ENTIDAD);
	}
	async panel() {
		return panelDemo(this.ahora, await this.cambios());
	}
	async guardarPerfil(p: PerfilPartner) {
		await guardarCambio(this.db, this.sb.id, ENTIDAD, 'perfil', p);
		return p;
	}
	async guardarCobro(c: CobroPartner) {
		await guardarCambio(this.db, this.sb.id, ENTIDAD, 'cobro', c);
		return c;
	}
	async subirFactura(mes: string, url: string) {
		const g = panelDemo(this.ahora).pagos.find((x) => x.mes === mes);
		if (!g || g.estado === 'pagada') throw new Error('sin-pago');
		await guardarCambio(this.db, this.sb.id, ENTIDAD, `factura:${mes}`, url);
	}
}
