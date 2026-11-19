import { sha256 } from '../crypto.ts';

/** Código del pedido para la página de confirmación: el número firmado (no se adivina el de otro). */
export async function codigoPedido(sandbox: string, numero: number) {
	return `${numero}-${(await sha256(`pedido:${sandbox}:${numero}`)).replace(/[^a-z0-9]/gi, '').slice(0, 10)}`;
}

/** El pedido de Mercado Pago que quedó esperando el pago en este navegador (volvió atrás desde MP). */
export const cookiePendiente = (slug: string) => `tlw-mp-${slug}`;
