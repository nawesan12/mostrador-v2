/**
 * Filtros del motor de listas en la URL: `?q=…&seg=vip,risk&f=canal:wa,ciudad:Rosario&rango=50000&b=wa&orden=last`.
 */
import type { FiltrosUrl } from './tipos.ts';

const lista = (v: string | null) => (v ? v.split(',').filter(Boolean) : []);

export function leer(p: Pick<URLSearchParams, 'get'>): FiltrosUrl {
	const facetas: Record<string, string[]> = {};
	for (const kv of lista(p.get('f'))) {
		const i = kv.indexOf(':');
		if (i < 1) continue;
		const k = kv.slice(0, i);
		(facetas[k] ??= []).push(kv.slice(i + 1));
	}
	return {
		q: (p.get('q') ?? '').slice(0, 100),
		seg: lista(p.get('seg')),
		facetas,
		rango: Math.max(0, Number(p.get('rango')) || 0),
		banderas: lista(p.get('b')),
		orden: p.get('orden') ?? ''
	};
}

export function escribir(f: Partial<FiltrosUrl>) {
	const p = new URLSearchParams();
	if (f.q?.trim()) p.set('q', f.q.trim());
	if (f.seg?.length) p.set('seg', f.seg.join(','));
	const fs = Object.entries(f.facetas ?? {}).flatMap(([k, vs]) => vs.map((v) => `${k}:${v}`));
	if (fs.length) p.set('f', fs.join(','));
	if (f.rango) p.set('rango', String(f.rango));
	if (f.banderas?.length) p.set('b', f.banderas.join(','));
	if (f.orden) p.set('orden', f.orden);
	return p;
}

/** Para comparar con las vistas guardadas (sin el orden). */
export function firma(f: Partial<FiltrosUrl>) {
	const p = escribir({ ...f, orden: '' });
	p.sort();
	return p.toString();
}

export const VACIO: FiltrosUrl = { q: '', seg: [], facetas: {}, rango: 0, banderas: [], orden: '' };
