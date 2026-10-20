/** Formatos del panel (es-AR, hora de Argentina), como los escribe el handoff. */
const TZ = 'America/Argentina/Buenos_Aires';
const NB = ' ';

/** "$ 42.900" (espacio duro: un precio no se parte). */
export const pesos = (n: number) => '$' + NB + Math.round(n).toLocaleString('es-AR');

/** "$ 14,1 M" para montos grandes; si no, como `pesos`. */
export const pesosCortos = (n: number) =>
	n >= 1_000_000
		? '$' + NB + (n / 1_000_000).toLocaleString('es-AR', { maximumFractionDigits: 1 }) + NB + 'M'
		: pesos(n);

const signo = (n: number) => (n < 0 ? '−' : n > 0 ? '+' : '');

/** "+21" / "−3": diferencia de cantidades. */
export const diferencia = (a: number, b: number) =>
	signo(a - b) + Math.abs(Math.round(a - b)).toLocaleString('es-AR');

/** "+6,8%": variación porcentual (null si no hay con qué comparar). */
export function variacion(a: number, b: number) {
	if (!b) return null;
	const v = ((a - b) / b) * 100;
	return (
		signo(v) +
		Math.abs(v).toLocaleString('es-AR', { minimumFractionDigits: 1, maximumFractionDigits: 1 }) +
		'%'
	);
}

/** "+0,4": diferencia en puntos. */
export const puntos = (a: number, b: number) =>
	signo(Math.round((a - b) * 10)) +
	Math.abs(a - b).toLocaleString('es-AR', { minimumFractionDigits: 1, maximumFractionDigits: 1 });

/** "3,4%". */
export const porcentaje = (n: number, decimales = 1) =>
	n.toLocaleString('es-AR', {
		minimumFractionDigits: decimales,
		maximumFractionDigits: decimales
	}) + '%';

const partes = (d: Date, o: Intl.DateTimeFormatOptions) =>
	new Intl.DateTimeFormat('es-AR', { timeZone: TZ, ...o }).formatToParts(d);
const parte = (d: Date, o: Intl.DateTimeFormatOptions, tipo: Intl.DateTimeFormatPartTypes) =>
	partes(d, o).find((p) => p.type === tipo)?.value ?? '';

/** Hora en Argentina (0–23). */
export const hora = (d: Date) => Number(parte(d, { hour: 'numeric', hourCycle: 'h23' }, 'hour'));

/** "Buen día" / "Buenas tardes" / "Buenas noches". */
export function saludo(d: Date) {
	const h = hora(d);
	return h >= 6 && h < 13 ? 'Buen día' : h >= 13 && h < 20 ? 'Buenas tardes' : 'Buenas noches';
}

const mayuscula = (s: string) => s.charAt(0).toUpperCase() + s.slice(1);

/** "Lunes 19 de octubre". */
export const fechaLarga = (d: Date) =>
	mayuscula(
		`${parte(d, { weekday: 'long' }, 'weekday')} ${parte(d, { day: 'numeric' }, 'day')} de ${parte(d, { month: 'long' }, 'month')}`
	);

/** "octubre". */
export const mes = (d: Date) => parte(d, { month: 'long' }, 'month');

/** "9:02". */
export const horaMin = (d: Date) =>
	new Intl.DateTimeFormat('es-AR', {
		timeZone: TZ,
		hour: 'numeric',
		minute: '2-digit',
		hourCycle: 'h23'
	}).format(d);

/** Etiquetas del eje del gráfico: "8 h", "mar", "20 sep". */
export function etiquetaEje(d: Date, periodo: 'hoy' | '7d' | '30d') {
	if (periodo === 'hoy') return `${hora(d)}${NB}h`;
	if (periodo === '7d') return parte(d, { weekday: 'short' }, 'weekday').replace('.', '');
	return `${parte(d, { day: 'numeric' }, 'day')}${NB}${parte(d, { month: 'short' }, 'month').replace('.', '')}`;
}

/** "Hoy" o "Mié 21" para una fecha YYYY-MM-DD. */
export function diaCorto(fecha: string, ahora: Date) {
	const hoy = new Intl.DateTimeFormat('en-CA', { timeZone: TZ }).format(ahora);
	if (fecha === hoy) return 'Hoy';
	const d = new Date(`${fecha}T12:00:00-03:00`);
	return `${mayuscula(parte(d, { weekday: 'short' }, 'weekday').replace('.', ''))} ${parte(d, { day: 'numeric' }, 'day')}`;
}

/** "4 min", "2 h", "ayer" (sin el "hace"). */
export function haceCorto(iso: string, ahora = Date.now()) {
	const min = Math.max(1, Math.round((ahora - Date.parse(iso)) / 60_000));
	if (min < 60) return `${min}${NB}min`;
	const h = Math.round(min / 60);
	if (h < 24) return `${h}${NB}h`;
	return h < 48 ? 'ayer' : `${Math.round(h / 24)}${NB}días`;
}

/** Curva del gráfico del handoff (`chart`): línea suave, área y punto final, en un lienzo de 600×200. */
export function curva(s: number[], p: number[]) {
	const todos = s.concat(p);
	let max = Math.max(...todos) * 1.12;
	let min = Math.min(...todos) * 0.7;
	if (max === min) [min, max] = [min - 1, max + 1];
	const pts = (a: number[]) =>
		a.map((v, i) => [(i * 600) / (a.length - 1), 200 - ((v - min) / (max - min)) * 200]);
	const linea = (P: number[][]) =>
		P.reduce((d, q, i, A) => {
			if (!i) return `M${q[0].toFixed(1)},${q[1].toFixed(1)}`;
			const p0 = A[i - 2] || A[i - 1],
				p1 = A[i - 1],
				p3 = A[i + 1] || q;
			const c1 = [p1[0] + (q[0] - p0[0]) / 6, p1[1] + (q[1] - p0[1]) / 6];
			const c2 = [q[0] - (p3[0] - p1[0]) / 6, q[1] - (p3[1] - p1[1]) / 6];
			return (
				d +
				` C${c1[0].toFixed(1)},${c1[1].toFixed(1)} ${c2[0].toFixed(1)},${c2[1].toFixed(1)} ${q[0].toFixed(1)},${q[1].toFixed(1)}`
			);
		}, '');
	const S = pts(s);
	const l = linea(S);
	return {
		linea: l,
		area: l + ' L600,200 L0,200 Z',
		anterior: linea(pts(p)),
		punto: (S[S.length - 1][1] / 200) * 100
	};
}
