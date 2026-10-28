/**
 * "Importar tienda": de qué plataforma viene, cuántas páginas tiene y cómo se lee cada una. Lo usan
 * los dos pasos de `/panel/productos/tienda` (descubrir y lote).
 */
import { fichasDelSitemap, leer as leerEmpretienda } from '../empretienda.ts';
import { MAX_PRODUCTOS, origenDe, POR_PAGINA } from './comun.ts';
import { descubrirShopify } from './shopify.ts';
import { descubrirWoo } from './woocommerce.ts';

export const FUENTES = ['shopify', 'woocommerce', 'empretienda'] as const;
export type Fuente = (typeof FUENTES)[number];

/** Con qué se completa un nombre suelto ("mitienda") en cada plataforma. */
const SUFIJO: Partial<Record<Fuente, string>> = {
	shopify: '.myshopify.com',
	empretienda: '.empretienda.com.ar'
};

export const origenPara = (fuente: Fuente, entrada: string) => origenDe(entrada, SUFIJO[fuente]);

/** La plataforma que delata el dominio, si la delata. */
export function plataformaDelDominio(origen: string): Fuente | null {
	if (origen.endsWith('.empretienda.com.ar')) return 'empretienda';
	if (origen.endsWith('.myshopify.com')) return 'shopify';
	return null;
}

export type Descubierta = {
	fuente: Fuente;
	origen: string;
	/** Productos que se van a leer (null: Shopify no lo dijo; se lee hasta una página vacía). */
	total: number | null;
	paginas: number | null;
	/** Sólo Empretienda: las fichas a leer, de a `POR_PAGINA` por lote. */
	fichas?: string[];
	/** Los que pasan del tope de `MAX_PRODUCTOS`. */
	omitidas: number;
};

async function probar(
	fuente: Fuente,
	origen: string,
	f: typeof fetch
): Promise<Descubierta | null | 'vacia'> {
	if (fuente === 'empretienda') {
		const xml = await leerEmpretienda(`${origen}/sitemap.xml`, f);
		if (!xml?.includes('<urlset')) return null;
		const todas = fichasDelSitemap(xml, origen);
		if (!todas.length) return 'vacia';
		const fichas = todas.slice(0, MAX_PRODUCTOS);
		return {
			fuente,
			origen,
			total: fichas.length,
			paginas: Math.ceil(fichas.length / POR_PAGINA),
			fichas,
			omitidas: todas.length - fichas.length
		};
	}
	const d =
		fuente === 'shopify' ? await descubrirShopify(origen, f) : await descubrirWoo(origen, f);
	if (!d) return null;
	if (d.total === 0) return 'vacia';
	return {
		fuente,
		origen,
		total: d.total,
		paginas: d.total === null ? null : Math.ceil(d.total / POR_PAGINA),
		omitidas: 0
	};
}

/**
 * Prueba primero la plataforma que dice el dominio, después la que eligió el usuario y después el
 * resto: si pegó un link de Shopify con WooCommerce elegido, igual la encuentra.
 */
export async function descubrir(
	elegida: Fuente,
	entrada: string,
	f: typeof fetch = fetch
): Promise<Descubierta | 'invalida' | 'vacia' | 'no-encontrada'> {
	const origen = origenPara(elegida, entrada);
	if (!origen) return 'invalida';
	const delDominio = plataformaDelDominio(origen);
	const orden = delDominio ? [delDominio] : [elegida, ...FUENTES.filter((x) => x !== elegida)];
	for (const fuente of orden) {
		const r = await probar(fuente, origen, f);
		if (r) return r;
	}
	return 'no-encontrada';
}
