/** Textos de turnos (tienda y panel), en hora de Argentina. */
import { diaAr, diaSemana, horaAr, minutos, sumarDias, type Franja } from './libres.ts';

export const DIAS = ['Domingo', 'Lunes', 'Martes', 'Miércoles', 'Jueves', 'Viernes', 'Sábado'];
export const DIAS_CORTOS = ['Do', 'Lu', 'Ma', 'Mi', 'Ju', 'Vi', 'Sá'];
export const MESES = [
	'enero',
	'febrero',
	'marzo',
	'abril',
	'mayo',
	'junio',
	'julio',
	'agosto',
	'septiembre',
	'octubre',
	'noviembre',
	'diciembre'
];

const mesDe = (dia: string) => MESES[Number(dia.slice(5, 7)) - 1];
const numDe = (dia: string) => Number(dia.slice(8, 10));

/** "Jueves 23 de octubre". */
export const diaLargo = (dia: string) => `${DIAS[diaSemana(dia)]} ${numDe(dia)} de ${mesDe(dia)}`;

/** "Jueves 23 de octubre · 17:30 h". */
export const cuandoLargo = (iso: string) => {
	const t = Date.parse(iso);
	return `${diaLargo(diaAr(t))} · ${horaAr(t)} h`;
};

/** "Hoy 17:30", "Mañana 11:00" o "Jue 10:30" (el próximo libre de alguien). */
export function cuandoCorto(iso: string, ahora = Date.now()) {
	const t = Date.parse(iso);
	const d = diaAr(t);
	const hoy = diaAr(ahora);
	const que =
		d === hoy ? 'Hoy' : d === sumarDias(hoy, 1) ? 'Mañana' : DIAS[diaSemana(d)].slice(0, 3);
	return `${que} ${horaAr(t)}`;
}

/** "Octubre" (el mes de un día, con mayúscula). */
export const mesTitulo = (dia: string) => mesDe(dia)[0].toUpperCase() + mesDe(dia).slice(1);

/** "10", "20:30" (las horas en punto sin minutos, como el diseño: "10 a 20:30 h"). */
const corta = (h: string) =>
	h.endsWith(':00') ? String(Number(h.slice(0, 2))) : h.replace(/^0/, '');

/** "40 min", "1 h 30 min". */
export const duracion = (min: number) =>
	min < 60 ? `${min} min` : `${Math.floor(min / 60)} h${min % 60 ? ` ${min % 60} min` : ''}`;

/**
 * El horario por días, agrupando los que son iguales y seguidos (design/Plantilla Tijera `hours`):
 * [["Lunes a viernes", "10 a 20:30 h"], ["Sábado", "9 a 18 h"], ["Domingo", "Cerrado"]].
 */
export function horarioPorDias(
	horario: Franja[]
): { dias: string; horas: string; cerrado: boolean; hoy: boolean }[] {
	const orden = [1, 2, 3, 4, 5, 6, 0];
	const texto = (d: number) => {
		const fr = horario
			.filter((f) => f.dia === d)
			.sort((a, b) => minutos(a.desde) - minutos(b.desde));
		return fr.length
			? fr.map((f) => `${corta(f.desde)} a ${corta(f.hasta)}`).join(' y ') + ' h'
			: '';
	};
	const hoy = diaSemana(diaAr(Date.now()));
	const grupos: { desde: number; hasta: number; horas: string }[] = [];
	for (const d of orden) {
		const h = texto(d);
		const ult = grupos.at(-1);
		if (ult && ult.horas === h) ult.hasta = d;
		else grupos.push({ desde: d, hasta: d, horas: h });
	}
	return grupos.map((g) => {
		const dias =
			g.desde === g.hasta ? DIAS[g.desde] : `${DIAS[g.desde]} a ${DIAS[g.hasta].toLowerCase()}`;
		const de = orden.slice(orden.indexOf(g.desde), orden.indexOf(g.hasta) + 1);
		return { dias, horas: g.horas || 'Cerrado', cerrado: !g.horas, hoy: de.includes(hoy) };
	});
}
