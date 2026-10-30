/** Fechas, tamaños e iniciales como los muestra el diseño (hora de Argentina). */
const TZ = 'America/Argentina/Buenos_Aires';
const dia = (d: Date) => d.toLocaleDateString('en-CA', { timeZone: TZ });

export function hora(iso: string, ahora = new Date()): string {
	const d = new Date(iso);
	if (dia(d) === dia(ahora) || dia(d) === dia(new Date(+ahora - 86_400_000)))
		return d.toLocaleTimeString('es-AR', {
			timeZone: TZ,
			hour: 'numeric',
			minute: '2-digit',
			hourCycle: 'h23'
		});
	if (+ahora - +d < 6 * 86_400_000)
		return d.toLocaleDateString('es-AR', { timeZone: TZ, weekday: 'short' }).replace('.', '');
	return d
		.toLocaleDateString('es-AR', {
			timeZone: TZ,
			day: 'numeric',
			month: 'short',
			...(d.getFullYear() !== ahora.getFullYear() ? { year: 'numeric' } : {})
		})
		.replace('.', '');
}

/** "Hoy, 9:42" / "Ayer, 18:04" / "lun 20 de oct, 9:15". */
export function cuando(iso: string, ahora = new Date()): string {
	const d = new Date(iso);
	const h = d.toLocaleTimeString('es-AR', {
		timeZone: TZ,
		hour: 'numeric',
		minute: '2-digit',
		hourCycle: 'h23'
	});
	if (dia(d) === dia(ahora)) return `Hoy, ${h}`;
	if (dia(d) === dia(new Date(+ahora - 86_400_000))) return `Ayer, ${h}`;
	return `${d.toLocaleDateString('es-AR', { timeZone: TZ, weekday: 'short', day: 'numeric', month: 'short' }).replace(/\./g, '')}, ${h}`;
}

/** El grupo de la lista: Hoy, Ayer, Esta semana, o el mes. */
export function grupo(iso: string, ahora = new Date()): string {
	const d = new Date(iso);
	if (dia(d) === dia(ahora)) return 'Hoy';
	if (dia(d) === dia(new Date(+ahora - 86_400_000))) return 'Ayer';
	if (+ahora - +d < 7 * 86_400_000) return 'Esta semana';
	const m = d.toLocaleDateString('es-AR', { timeZone: TZ, month: 'long', year: 'numeric' });
	return m.charAt(0).toUpperCase() + m.slice(1);
}

export function tamano(b: number): string {
	const u = ['B', 'KB', 'MB', 'GB'];
	let i = 0;
	let n = b;
	while (n >= 1024 && i < u.length - 1) {
		n /= 1024;
		i++;
	}
	return `${n.toLocaleString('es-AR', { maximumFractionDigits: i >= 2 ? 1 : 0 })} ${u[i]}`;
}

/** "Comercial Pérez" → "CP"; "lucia" → "L". */
export function iniciales(nombre: string): string {
	const p = nombre
		.replace(/[^\p{L}\p{N} ]/gu, ' ')
		.trim()
		.split(/\s+/)
		.filter(Boolean);
	if (!p.length) return '?';
	return (p.length > 1 ? p[0][0] + p[1][0] : p[0][0]).toUpperCase();
}

/** Ícono y color del adjunto por tipo (como el diseño: PDF rojo, planilla verde, foto azul). */
export function tipoAdjunto(tipo: string, nombre: string) {
	const n = nombre.toLowerCase();
	if (tipo === 'application/pdf' || n.endsWith('.pdf'))
		return { ic: 'icon-file-text', c: 'var(--prT)', l: 'PDF' };
	if (/spreadsheet|excel|csv/.test(tipo) || /\.(xlsx?|csv|ods)$/.test(n))
		return { ic: 'icon-file-spreadsheet', c: 'var(--okT)', l: 'Planilla' };
	if (tipo.startsWith('image/')) return { ic: 'icon-image', c: 'var(--infT)', l: 'Foto' };
	return { ic: 'icon-file', c: 'var(--mut)', l: 'Archivo' };
}
