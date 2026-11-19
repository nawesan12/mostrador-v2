/**
 * Volver a verificarse antes de cambiar adónde llega la plata (alias, CBU, titular, la cuenta de
 * Mercado Pago). La API responde 403 «reverificar»; `conVerificacion` abre `Reverificar.svelte`
 * (montado una vez en el layout del panel), espera el permiso y repite el pedido con él.
 */
let pendiente = $state<((token: string | null) => void) | null>(null);

export const reverificacion = {
	get abierta() {
		return pendiente !== null;
	},
	terminar(token: string | null) {
		const r = pendiente;
		pendiente = null;
		r?.(token);
	}
};

/** Abre la verificación y devuelve el permiso (o null si la persona la cerró). */
export function pedirVerificacion(): Promise<string | null> {
	reverificacion.terminar(null);
	return new Promise((resolver) => (pendiente = resolver));
}

/** Hace el pedido; si la API pide verificarse, la pide y lo repite con el permiso. */
export async function conVerificacion(
	hacer: (encabezados: Record<string, string>) => Promise<Response>
): Promise<Response> {
	const r = await hacer({});
	if (r.status !== 403) return r;
	const p = (await r
		.clone()
		.json()
		.catch(() => null)) as { code?: string } | null;
	if (p?.code !== 'reverificar') return r;
	const token = await pedirVerificacion();
	if (!token) return r;
	return hacer({ 'x-reverificacion': token });
}
