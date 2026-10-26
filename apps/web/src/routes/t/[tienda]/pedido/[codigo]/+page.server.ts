import { env } from 'cloudflare:workers';
import { error } from '@sveltejs/kit';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import { codigoPedido } from '#lib/server/tienda/pedido.ts';
import type { PageServerLoad } from './$types';

/** La confirmación del pedido: el código lleva la firma, así nadie mira el pedido de otro. */
export const load: PageServerLoad = async ({ params, cookies, url }) => {
	const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
	// Con la API real, el código es el público del pedido.
	if (r && !r.datos) {
		const res = await fetch(
			`${String(env.API_URL)}/v1/publico/tiendas/${r.tienda.slug}/pedidos/${encodeURIComponent(params.codigo)}`
		);
		if (!res.ok) error(404, 'No encontramos ese pedido.');
		const c = (await res.json()) as {
			numero: number;
			total: number;
			pago: string;
			nombre: string;
			estado: string;
			envio: string;
			seguimiento: string | null;
			alias: string | null;
			cbu: string | null;
			titular: string | null;
		};
		// Los datos para transferir son los que cargó la tienda (nunca inventados). Sin alias, el CBU/CVU.
		const tr = c.pago === 'transferencia';
		return {
			pedido: { ...c, alias: tr ? (c.alias ?? c.cbu) : null, esCbu: tr && !c.alias && !!c.cbu }
		};
	}
	const numero = Number(params.codigo.split('-')[0]);
	if (
		!r?.datos ||
		!r.sandboxId ||
		!Number.isInteger(numero) ||
		(await codigoPedido(r.sandboxId, numero)) !== params.codigo
	)
		error(404, 'No encontramos ese pedido.');
	const p = await r.datos.pedido(numero);
	if (!p) error(404, 'No encontramos ese pedido.');
	const pago = url.searchParams.get('pago') === 'mercadopago' ? 'mercadopago' : 'transferencia';
	const total = Number(url.searchParams.get('total')) || p.total;
	const dt = pago === 'transferencia' ? r.tienda.cobros.datosTransferencia : null;
	return {
		pedido: {
			numero,
			total,
			pago,
			nombre: p.cliente.split(' ')[0],
			alias: dt?.alias || dt?.cbu || null,
			esCbu: !!dt && !dt.alias && !!dt.cbu,
			titular: r.tienda.cobros.datosTransferencia?.titular ?? null
		}
	};
};
