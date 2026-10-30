/**
 * Correo Argentino (API MiCorreo) desde el BFF: la conexión, los envíos de los pedidos y la
 * cotización del checkout viven en Go (integraciones/correoargentino). En la demo (sandbox) no hay
 * cuenta de MiCorreo: Correo no se ofrece y conectar responde que es una tienda de ejemplo. Nunca se
 * muestra un precio que no haya cotizado Correo.
 */
import { env } from 'cloudflare:workers';
import type { RequestEvent } from '@sveltejs/kit';
import type { components } from '@mostrador/contrato/esquema';
import { ClienteBff } from './bff.ts';
import { clienteDe } from './datos/index.ts';

export type ConexionCorreo = components['schemas']['ConexionCorreo'];
export type AltaCorreo = components['schemas']['AltaCorreo'];
export type OrigenCorreo = components['schemas']['OrigenCorreo'];
export type PaqueteCorreo = components['schemas']['PaqueteCorreo'];
export type EnvioCorreo = components['schemas']['EnvioCorreo'];
export type CotizacionCorreo = components['schemas']['CotizacionCorreo'];
export type SucursalCorreo = components['schemas']['SucursalCorreo'];

export class CorreoApi extends ClienteBff {
	// Panel (sesión de la tienda).
	estado(sesion: string) {
		return this.pedir<ConexionCorreo>('GET', '/v1/panel/integraciones/correo', undefined, sesion);
	}
	conectar(sesion: string, a: AltaCorreo) {
		return this.pedir<ConexionCorreo>('POST', '/v1/panel/integraciones/correo', a, sesion);
	}
	ajustar(sesion: string, origen: OrigenCorreo, paquete: PaqueteCorreo) {
		return this.pedir<ConexionCorreo>(
			'PUT',
			'/v1/panel/integraciones/correo',
			{ origen, paquete },
			sesion
		);
	}
	desconectar(sesion: string) {
		return this.pedir<void>('DELETE', '/v1/panel/integraciones/correo', undefined, sesion);
	}
	envio(sesion: string, numero: number) {
		return this.pedir<EnvioCorreo>('GET', `/v1/panel/pedidos/${numero}/correo`, undefined, sesion);
	}
	generar(sesion: string, numero: number) {
		return this.pedir<EnvioCorreo>('POST', `/v1/panel/pedidos/${numero}/correo`, undefined, sesion);
	}
	seguimiento(sesion: string, numero: number, seguimiento: string) {
		return this.pedir<EnvioCorreo>(
			'PUT',
			`/v1/panel/pedidos/${numero}/correo/seguimiento`,
			{ seguimiento },
			sesion
		);
	}

	// Checkout (público, con la IP de quien compra: Go pone el tope).
	cotizar(slug: string, cp: string, unidades: number) {
		const q = new URLSearchParams({ cp, unidades: String(unidades) });
		return this.pedir<CotizacionCorreo>(
			'GET',
			`/v1/publico/tiendas/${encodeURIComponent(slug)}/correo/cotizacion?${q}`
		);
	}
	sucursales(slug: string, provincia: string, cp: string) {
		const q = new URLSearchParams({ provincia, cp });
		return this.pedir<SucursalCorreo[]>(
			'GET',
			`/v1/publico/tiendas/${encodeURIComponent(slug)}/correo/sucursales?${q}`
		);
	}
}

/** El cliente de Go con la IP y el navegador de quien hace el pedido. */
export const correoDe = (event: Pick<RequestEvent, 'request' | 'getClientAddress'>) =>
	new CorreoApi(String(env.API_URL), String(env.CLAVE_BFF ?? ''), clienteDe(event));
