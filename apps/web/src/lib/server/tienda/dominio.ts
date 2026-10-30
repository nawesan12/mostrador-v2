import { env } from 'cloudflare:workers';
import { esDominioPropio } from '#lib/tienda/dominio.ts';
import { usaSandbox } from '../datos/index.ts';

/** Lo que ya se preguntó, por isolate: cinco minutos (un dominio recién conectado tarda eso). */
const cache = new Map<string, { slug: string | null; vence: number }>();
const TTL = 5 * 60_000;

/**
 * El slug de la tienda de un dominio propio, o null. En la demo y en local, `<slug>.localhost`
 * muestra esa tienda (así se prueba sin DNS); con la API, sólo un dominio verificado.
 */
export async function slugDeDominio(host: string): Promise<string | null> {
	const h = host.toLowerCase();
	if (!esDominioPropio(h)) return null;
	if (usaSandbox()) {
		const m = /^([a-z0-9-]{2,60})\.localhost$/.exec(h);
		return m ? m[1] : null;
	}
	const ya = cache.get(h);
	if (ya && ya.vence > Date.now()) return ya.slug;
	const r = await fetch(`${String(env.API_URL)}/v1/publico/dominios/${encodeURIComponent(h)}`);
	if (!r.ok && r.status !== 404) throw new Error(`API ${r.status}`);
	const slug = r.ok ? ((await r.json()) as { slug: string }).slug : null;
	cache.set(h, { slug, vence: Date.now() + TTL });
	return slug;
}
