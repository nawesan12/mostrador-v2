/** Lo que muestra /estado, leído de la base del monitor. */
import { COMPONENTES, diaAR } from './componentes.ts';

export type Barra = { dia: string; estado: 'ok' | 'falla' | 'caida' | 'sin' };
export type EstadoComponente = {
	id: string;
	nombre: string;
	/** Disponibilidad del período, null si todavía no hay chequeos. */
	disponibilidad: number | null;
	barras: Barra[];
	/** null: nunca se chequeó (p. ej. la API en la demo). */
	ok: boolean | null;
};
export type Incidente = {
	id: number;
	componente: string;
	tipo: 'incidente' | 'mantenimiento';
	titulo: string;
	detalle: string;
	inicio: number;
	fin: number | null;
};
export type Estado = {
	componentes: EstadoComponente[];
	disponibilidad: number | null;
	actualizado: number | null;
	incidentes: Incidente[];
};

const DIA_MS = 86_400_000;

/** Los últimos `n` días de la Argentina (sin horario de verano), del más viejo a hoy. */
export const ultimosDias = (n: number, ahora: number) =>
	Array.from({ length: n }, (_, i) => diaAR(ahora - (n - 1 - i) * DIA_MS));

export async function leerEstado(
	db: D1Database,
	dias: 7 | 30 | 90,
	ahora = Date.now()
): Promise<Estado> {
	const rango = ultimosDias(dias, ahora);
	const [filas, ultimos, incidentes] = await Promise.all([
		db
			.prepare('SELECT componente, dia, chequeos, fallas FROM dias WHERE dia >= ?1')
			.bind(rango[0])
			.all<{ componente: string; dia: string; chequeos: number; fallas: number }>(),
		db
			.prepare('SELECT componente, ok, chequeado_en FROM ultimo')
			.all<{ componente: string; ok: number; chequeado_en: number }>(),
		db
			.prepare(
				'SELECT id, componente, tipo, titulo, detalle, inicio, fin FROM incidentes WHERE inicio >= ?1 OR fin IS NULL ORDER BY inicio DESC LIMIT 10'
			)
			.bind(ahora - 90 * DIA_MS)
			.all<Incidente>()
	]);
	const porClave = new Map(filas.results.map((f) => [`${f.componente}|${f.dia}`, f]));
	const ult = new Map(ultimos.results.map((u) => [u.componente, u]));
	let totalCh = 0;
	let totalFa = 0;
	const componentes = COMPONENTES.map(([id, nombre]): EstadoComponente => {
		let ch = 0;
		let fa = 0;
		const barras = rango.map((dia): Barra => {
			const f = porClave.get(`${id}|${dia}`);
			if (!f || !f.chequeos) return { dia, estado: 'sin' };
			ch += f.chequeos;
			fa += f.fallas;
			if (!f.fallas) return { dia, estado: 'ok' };
			return { dia, estado: f.fallas * 2 >= f.chequeos ? 'caida' : 'falla' };
		});
		totalCh += ch;
		totalFa += fa;
		const u = ult.get(id);
		return {
			id,
			nombre,
			disponibilidad: ch ? ((ch - fa) / ch) * 100 : null,
			barras,
			ok: u ? u.ok === 1 : null
		};
	});
	const actualizado = ultimos.results.reduce<number | null>(
		(m, u) => Math.max(m ?? 0, u.chequeado_en),
		null
	);
	return {
		componentes,
		disponibilidad: totalCh ? ((totalCh - totalFa) / totalCh) * 100 : null,
		actualizado,
		incidentes: incidentes.results
	};
}

export { porcentaje } from '#lib/sitio/estado.ts';
