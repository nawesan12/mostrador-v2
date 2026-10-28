/**
 * Los links de la tienda vieja (importada de otra plataforma). Cuando un path de /t/<tienda> no
 * existe, el hook pregunta acá si había una dirección vieja guardada y responde 301 a la nueva: así
 * no se pierden Google, Instagram ni los mails viejos. Sólo se pregunta ante un 404, para que una
 * página que sí existe no pague la consulta.
 */
import { env } from 'cloudflare:workers';
import { usaSandbox } from '../datos/index.ts';
import { DatosSandbox, dbSandbox, leer } from '../sandbox/sandbox.ts';
import { redireccionDe } from '../sandbox/importacion.ts';
import { haciaValido } from '../importacion/rutas.ts';

const TIENDA = /^\/t\/([a-z0-9-]{2,60})(\/.+)$/;

/** El path de /t/<tienda>/… adónde redirigir (absoluto en el sitio), o null. */
export async function redireccionTienda(
	pathname: string,
	cookie: string | undefined
): Promise<string | null> {
	const m = TIENDA.exec(pathname);
	if (!m) return null;
	const [, slug, path] = m;
	let hacia: string | null;
	if (!usaSandbox()) {
		const r = await fetch(
			`${String(env.API_URL)}/v1/publico/tiendas/${slug}/redireccion?path=${encodeURIComponent(path)}`,
			{ redirect: 'manual', headers: { accept: 'application/json' } }
		);
		if (r.status !== 301) return null;
		hacia = r.headers.get('location');
	} else {
		// En la demo, sólo la tienda del visitante tiene importaciones (las vitrinas, no).
		const db = dbSandbox(env.SANDBOX);
		const sb = await leer(db, cookie);
		if (!sb || (await new DatosSandbox(db, sb).tienda()).slug !== slug) return null;
		hacia = await redireccionDe(db, sb.id, path);
	}
	// Nunca a otro dominio, aunque la base dijera otra cosa.
	if (!hacia || !haciaValido(hacia)) return null;
	return `/t/${slug}${hacia === '/' ? '' : hacia}`;
}
