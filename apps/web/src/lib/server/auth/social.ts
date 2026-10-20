/**
 * Ingreso con Google, Mercado Pago y Mercado Libre desde el BFF. La API guarda el state y el PKCE;
 * acá sólo se manda a la persona al proveedor y, a la vuelta, se canjea y se guarda la sesión.
 */
import { fail, redirect, type RequestEvent } from '@sveltejs/kit';
import { cuentasDe, usaSandbox } from '../datos/index.ts';
import { ErrorApi } from '../bff.ts';
import { PROVEEDORES, type ProveedorOAuth } from './esquemas.ts';

export const esProveedor = (p: string): p is ProveedorOAuth =>
	(PROVEEDORES as readonly string[]).includes(p);

/**
 * Manda al proveedor. En la demo no hay cuentas: sigue con el recorrido de ejemplo (`demo`).
 * Si el proveedor está apagado en la API, vuelve al formulario con el motivo.
 */
export async function irAlProveedor(
	event: RequestEvent,
	proveedor: ProveedorOAuth,
	demo: string,
	modo: string
) {
	if (usaSandbox()) redirect(303, demo);
	let url: string;
	try {
		({ url } = await cuentasDe(event).iniciarOAuth(proveedor));
	} catch (e) {
		if (!(e instanceof ErrorApi)) throw e;
		return fail(e.estado >= 500 ? 503 : e.estado, { modo, errores: { form: e.message } });
	}
	redirect(303, url);
}
