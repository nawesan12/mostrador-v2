/**
 * Encuentra la tienda online de /t/[tienda]. En la demo: si quien mira tiene su sandbox y el slug es
 * el de su tienda, ve la suya (con lo que cambió y publicó); si no, ve la vitrina de la plantilla de
 * ese nombre (los datos de ejemplo, sólo para mirar). Con la API real, la pide a Go.
 */
import { env } from 'cloudflare:workers';
import RUBROS from '@mostrador/contrato/rubros.json';
import { PLANTILLAS } from '#lib/plantillas.ts';
import type { TiendaPublica } from '#lib/tienda/tipos.ts';
import { usaSandbox } from '../datos/index.ts';
import { slugDe } from '../sandbox/rubros.ts';
import { DatosSandbox, dbSandbox, leer, type Sandbox } from '../sandbox/sandbox.ts';

const DIA = 86_400_000;

/** El sandbox de mentira de una plantilla: arranca hoy (los datos de ejemplo quedan frescos). */
function vitrinaDe(slug: string): Sandbox | null {
	const p = PLANTILLAS.find((x) => slugDe(x.tienda) === slug);
	const r = p && RUBROS.find((x) => x.nombre === p.rubro);
	if (!p || !r) return null;
	const hoy = Math.floor(Date.now() / DIA) * DIA;
	return {
		id: `vitrina-${p.id}`,
		rubro: r.id,
		plantilla: p.id,
		creadoEn: hoy,
		venceEn: Number.MAX_SAFE_INTEGER
	};
}

export type TiendaResuelta = {
	tienda: TiendaPublica;
	datos: DatosSandbox | null;
	sandboxId: string | null;
};

export async function resolverTienda(
	slug: string,
	cookie: string | undefined
): Promise<TiendaResuelta | null> {
	if (!/^[a-z0-9-]{2,60}$/.test(slug)) return null;
	if (!usaSandbox()) {
		const r = await fetch(`${String(env.API_URL)}/v1/publico/tiendas/${slug}`);
		if (r.status === 404) return null;
		if (!r.ok) throw new Error(`API ${r.status}`);
		const tienda = (await r.json()) as TiendaPublica;
		return { tienda: { ...tienda, turnos: tienda.turnos ?? null }, datos: null, sandboxId: null };
	}
	const db = dbSandbox(env.SANDBOX);
	const propio = await leer(db, cookie);
	if (propio) {
		const datos = new DatosSandbox(db, propio);
		if ((await datos.tienda()).slug === slug)
			return { tienda: await datos.tiendaPublica(false), datos, sandboxId: propio.id };
	}
	const v = vitrinaDe(slug);
	if (!v) return null;
	const datos = new DatosSandbox(db, v);
	return { tienda: await datos.tiendaPublica(true), datos, sandboxId: null };
}
