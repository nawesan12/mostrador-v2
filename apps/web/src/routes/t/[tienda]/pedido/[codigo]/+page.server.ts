import { error } from '@sveltejs/kit';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import { codigoPedido } from '#lib/server/tienda/pedido.ts';
import type { PageServerLoad } from './$types';

/** La confirmación del pedido: el código lleva la firma, así nadie mira el pedido de otro. */
export const load: PageServerLoad = async ({ params, cookies, url }) => {
	const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
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
	return {
		pedido: {
			numero,
			total,
			pago,
			nombre: p.cliente.split(' ')[0],
			alias: pago === 'transferencia' ? `${r.tienda.slug.replace(/-/g, '.')}.mp` : null
		}
	};
};
