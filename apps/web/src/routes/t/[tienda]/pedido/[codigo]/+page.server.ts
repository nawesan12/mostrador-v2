import { env } from 'cloudflare:workers';
import { error, fail } from '@sveltejs/kit';
import { COOKIE } from '#lib/server/sandbox/sandbox.ts';
import { resolverTienda } from '#lib/server/tienda/publica.ts';
import { usaSandbox } from '#lib/server/datos/index.ts';
import { codigoPedido, cookiePendiente } from '#lib/server/tienda/pedido.ts';
import { irAfuera } from '#lib/server/afuera.ts';
import type { Actions, PageServerLoad } from './$types';

/** La confirmación del pedido: el código lleva la firma, así nadie mira el pedido de otro. */
export const load: PageServerLoad = async ({ params, cookies, url, depends, parent }) => {
	// Tiempo real (§44): al acreditarse el pago (o cambiar el estado) la página se actualiza sola.
	depends('tienda:pedido-publico');
	// Con la API, la tienda ya la resolvió el layout (antes se pedía dos veces); la demo necesita sus datos.
	const r = usaSandbox()
		? await resolverTienda(params.tienda, cookies.get(COOKIE))
		: { tienda: (await parent()).tienda, datos: null, sandboxId: null };
	// Con la API real, el código es el público del pedido.
	if (r && !r.datos) {
		// Con tope: es la página a la que vuelve Mercado Pago. Si la API no responde, un estado mínimo
		// (el número sale del código) y el tiempo real lo completa.
		const res = await fetch(
			`${String(env.API_URL)}/v1/publico/tiendas/${r.tienda.slug}/pedidos/${encodeURIComponent(params.codigo)}`,
			{ signal: AbortSignal.timeout(5000) }
		).catch(() => null);
		if (res?.status === 404) error(404, 'No encontramos ese pedido.');
		if (!res?.ok) {
			const numero = Number(params.codigo.split('-')[0]);
			if (!Number.isInteger(numero) || numero < 1) error(503, 'No pudimos cargar tu pedido.');
			return {
				pedido: {
					numero,
					codigo: params.codigo,
					total: 0,
					pago: 'mercadopago',
					nombre: '',
					alias: null,
					estado: 'Para preparar',
					sinDatos: true
				}
			};
		}
		const c = (await res.json()) as {
			numero: number;
			total: number;
			pago: string;
			nombre: string;
			estado: string;
			estadoPago?: string;
			envio: string;
			seguimiento: string | null;
			items?: { nombre: string; foto: string | null; variante: string | null; cantidad: number; precio: number }[];
			subtotal?: number;
			costoEnvio?: number;
			entrega?: { direccion: string; cp: string; localidad?: string; sucursal?: string } | null;
			alias: string | null;
			cbu: string | null;
			titular: string | null;
			/** Reservado o encargado con seña (§52). */
			sena?: {
				tipo: 'reserva' | 'encargo';
				monto: number;
				falta: number;
				pagada: boolean;
				saldado: boolean;
			} | null;
		};
		// Vuelve de Mercado Pago con el pago aprobado y el pedido todavía sin pagar: se le pide a la API
		// que lo confirme (la API lo consulta en Mercado Pago; no confía en lo que dice la URL).
		const idPago = url.searchParams.get('payment_id') ?? url.searchParams.get('collection_id');
		const aprobado =
			(url.searchParams.get('status') ?? url.searchParams.get('collection_status')) === 'approved';
		if (
			c.pago === 'mercadopago' &&
			c.estado === 'Sin pagar' &&
			aprobado &&
			idPago &&
			/^\d{1,20}$/.test(idPago)
		)
			await fetch(`${String(env.API_URL)}/v1/publico/tiendas/${r.tienda.slug}/pagos/${idPago}`, {
				method: 'POST',
				signal: AbortSignal.timeout(5000)
			}).catch(() => null);
		// Los datos para transferir son los que cargó la tienda (nunca inventados). Sin alias, el CBU/CVU.
		const tr = c.pago === 'transferencia';
		const pagado = ['Aprobado', 'Acreditada', 'Cobrado'].includes(c.estadoPago ?? '');
		// Ya pagó (o se canceló): el checkout deja de ofrecer «Tenés un pedido esperando el pago».
		if (pagado || c.estado === 'Cancelado')
			cookies.delete(cookiePendiente(r.tienda.slug), { path: `/t/${r.tienda.slug}` });
		return {
			pedido: {
				...c,
				codigo: params.codigo,
				alias: tr ? c.alias : null,
				cbu: tr ? c.cbu : null,
				titular: tr ? c.titular : null,
				// Los mismos estados de «pagado» que los mails a quien compra (Go, 00034).
				pagado,
				// Mercado Pago sin pagar todavía: la página ofrece pagarlo (si no, quedaba sin forma de pagar).
				porPagar: c.pago === 'mercadopago' && c.estado === 'Sin pagar' && c.estadoPago === 'Pendiente' && !c.sena,
				rechazado: ['rejected', 'failure', 'null'].includes(
					url.searchParams.get('status') ?? url.searchParams.get('collection_status') ?? ''
				)
			}
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
	// Con seña (§52), se paga por Mercado Pago (en la demo, simulado).
	const pago =
		p.sena || url.searchParams.get('pago') === 'mercadopago' ? 'mercadopago' : 'transferencia';
	const total = Number(url.searchParams.get('total')) || p.total;
	const dt = pago === 'transferencia' ? r.tienda.cobros.datosTransferencia : null;
	return {
		pedido: {
			numero,
			codigo: params.codigo,
			total,
			pago,
			nombre: p.cliente.split(' ')[0],
			alias: dt?.alias || null,
			cbu: dt?.cbu || null,
			estado: p.estado,
			envio: p.envio,
			titular: dt?.titular || null,
			items: p.items.map((i) => ({
				nombre: i.nombre,
				foto: i.foto ?? null,
				variante: i.variante ?? null,
				cantidad: i.cantidad,
				precio: i.precio
			})),
			pagado: false,
			sena: p.sena ?? null
		},
		// La demo no cobra la seña: la confirmación lo dice.
		senaDemo: !!p.sena
	};
};

export const actions: Actions = {
	/** Pagar después: arma un pago nuevo de Mercado Pago para el pedido y lleva a pagarlo. */
	pagar: async ({ params, cookies, request, getClientAddress }) => {
		const r = await resolverTienda(params.tienda, cookies.get(COOKIE));
		if (!r || r.datos) return fail(400, { error: 'Esta tienda no cobra con Mercado Pago.' });
		const res = await fetch(
			`${String(env.API_URL)}/v1/publico/tiendas/${r.tienda.slug}/pedidos/${encodeURIComponent(params.codigo)}/pagar`,
			{
				method: 'POST',
				headers: {
					'x-mostrador-bff': String(env.CLAVE_BFF ?? ''),
					'x-cliente-ip': request.headers.get('cf-connecting-ip') ?? getClientAddress()
				}
			}
		).catch(() => null);
		const g = res?.ok ? ((await res.json()) as { url?: string }) : null;
		if (!g?.url) return fail(409, { error: 'Mercado Pago no respondió. Probá de nuevo en un rato.' });
		irAfuera(g.url);
	}
};
