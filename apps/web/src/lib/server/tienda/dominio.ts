import { env } from 'cloudflare:workers';
import { esDominioPropio } from '#lib/tienda/dominio.ts';
import { usaSandbox } from '../datos/index.ts';

/**
 * La tienda de un dominio: su slug y si el plan que rige incluye dominio propio. Sin el plan (Gratis,
 * o al terminar la prueba) el dominio sigue resolviendo, pero la web responde 301 a la dirección de
 * Tu Local Web con el mismo camino: los links compartidos no se rompen y al subir de plan vuelve.
 */
export type TiendaDeDominio = { slug: string; propio: boolean };

/** Lo que ya se preguntó, por isolate: cinco minutos (un dominio recién conectado tarda eso). */
const cache = new Map<string, { tienda: TiendaDeDominio | null; vence: number }>();
const TTL = 5 * 60_000;

/**
 * La tienda de un dominio propio, o null. En la demo y en local, `<slug>.localhost` muestra esa
 * tienda (así se prueba sin DNS); con la API, sólo un dominio verificado.
 */
export async function tiendaDeDominio(host: string): Promise<TiendaDeDominio | null> {
	const h = host.toLowerCase();
	if (!esDominioPropio(h)) return null;
	if (usaSandbox()) {
		const m = /^([a-z0-9-]{2,60})\.localhost$/.exec(h);
		return m ? { slug: m[1], propio: true } : null;
	}
	const ya = cache.get(h);
	if (ya && ya.vence > Date.now()) return ya.tienda;
	const r = await fetch(`${String(env.API_URL)}/v1/publico/dominios/${encodeURIComponent(h)}`);
	if (!r.ok && r.status !== 404) throw new Error(`API ${r.status}`);
	const tienda = r.ok ? ((await r.json()) as { slug: string; propio?: boolean }) : null;
	const t = tienda ? { slug: tienda.slug, propio: tienda.propio ?? true } : null;
	cache.set(h, { tienda: t, vence: Date.now() + TTL });
	return t;
}

/** El slug de la tienda de un dominio (propio vigente o no), o null. */
export async function slugDeDominio(host: string): Promise<string | null> {
	return (await tiendaDeDominio(host))?.slug ?? null;
}

/** Olvida lo guardado de un dominio (la tienda cambió de dirección y este isolate no lo sabía). */
export function olvidarDominio(host: string) {
	cache.delete(host.toLowerCase());
}
