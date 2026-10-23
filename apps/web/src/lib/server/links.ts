/**
 * La página pública de un link de pago (/p/[codigo]): la abre quien paga, sin sesión ni cookie de la
 * demo. En la demo el link se busca entre los sandboxes; en la real, en la API Go.
 */
import { env } from 'cloudflare:workers';
import type { components } from '@mostrador/contrato/esquema';
import { usaSandbox } from './datos/index.ts';
import { buscarLink, dbSandbox, DatosSandbox } from './sandbox/sandbox.ts';

export type LinkPublico = components['schemas']['LinkPublico'];

export async function verLink(codigo: string): Promise<LinkPublico | null> {
	if (usaSandbox()) {
		const r = await buscarLink(dbSandbox(env.SANDBOX), codigo);
		if (!r) return null;
		const tienda = await new DatosSandbox(dbSandbox(env.SANDBOX), r.sb).tienda();
		return {
			codigo,
			tienda: tienda.nombre,
			concepto: r.link.concepto,
			monto: r.link.monto,
			cuotas: r.link.cuotas,
			pagado: !!r.link.pagado,
			pedido: r.link.pedido
		};
	}
	const r = await fetch(`${String(env.API_URL)}/v1/publico/links/${encodeURIComponent(codigo)}`);
	if (r.status === 404) return null;
	if (!r.ok) throw new Error(`API ${r.status}`);
	return (await r.json()) as LinkPublico;
}

/**
 * Pagar. En la demo se simula; en la real crea la preferencia de Checkout Pro con la cuenta de la
 * tienda y devuelve el init_point para mandar a quien paga.
 */
export async function pagarLink(
	codigo: string
): Promise<'ok' | 'no-existe' | 'ya-pagado' | 'sin-cobros' | 'error' | { url: string }> {
	if (usaSandbox()) {
		const r = await buscarLink(dbSandbox(env.SANDBOX), codigo);
		if (!r) return 'no-existe';
		const res = await new DatosSandbox(dbSandbox(env.SANDBOX), r.sb).pagarLink(codigo);
		return res === 'ok' ? 'ok' : res === 'no-existe' ? 'no-existe' : 'ya-pagado';
	}
	const r = await fetch(
		`${String(env.API_URL)}/v1/publico/links/${encodeURIComponent(codigo)}/pagar`,
		{
			method: 'POST'
		}
	);
	if (r.status === 404) return 'no-existe';
	if (r.status === 409) {
		const p = (await r.json().catch(() => null)) as { codigo?: string } | null;
		return p?.codigo === 'pagado' ? 'ya-pagado' : 'sin-cobros';
	}
	if (r.status === 502) return 'error';
	if (!r.ok) throw new Error(`API ${r.status}`);
	const { url } = (await r.json()) as { url: string };
	// Sólo se redirige al checkout de Mercado Pago.
	if (!/^https:\/\/(www\.)?mercadopago\.com(\.ar)?\//.test(url))
		throw new Error('URL de pago inesperada');
	return { url };
}
