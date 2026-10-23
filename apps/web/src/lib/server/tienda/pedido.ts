import { sha256 } from '../crypto.ts';

/** Código del pedido para la página de confirmación: el número firmado (no se adivina el de otro). */
export async function codigoPedido(sandbox: string, numero: number) {
	return `${numero}-${(await sha256(`pedido:${sandbox}:${numero}`)).replace(/[^a-z0-9]/gi, '').slice(0, 10)}`;
}
