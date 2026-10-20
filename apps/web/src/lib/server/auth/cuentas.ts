/**
 * Cuentas contra la API Go (DATA_SOURCE=api). El navegador nunca ve el token de Go: lo guarda este
 * servidor en una cookie httpOnly. A Go le llega la IP real del visitante en X-Cliente-IP junto con
 * la clave del BFF, que es lo único que hace que Go le crea.
 */
import type { components } from '@mostrador/contrato/esquema';

export type Sesion = components['schemas']['Sesion'];
export type DatosSesion = components['schemas']['DatosSesion'];

/** Error de la API, ya traducido para el formulario. */
export class ErrorCuenta extends Error {
	constructor(
		public estado: number,
		public codigo: string,
		message: string,
		public campos: Record<string, string> = {}
	) {
		super(message);
	}
}

type Problema = { detail?: string; code?: string; issues?: { path: string; message: string }[] };

export class CuentasApi {
	constructor(
		private base: string,
		private claveBFF: string,
		private cliente: { ip: string; agente: string },
		// Envuelto: en Workers, `fetch` llamado como método de otro objeto da "Illegal invocation".
		private f: typeof fetch = (i, init) => fetch(i, init)
	) {}

	private async pedir<T>(
		metodo: 'GET' | 'POST',
		ruta: string,
		cuerpo?: unknown,
		token?: string
	): Promise<T> {
		const res = await this.f(this.base + ruta, {
			method: metodo,
			headers: {
				accept: 'application/json',
				'content-type': 'application/json',
				'x-mostrador-bff': this.claveBFF,
				'x-cliente-ip': this.cliente.ip,
				'x-cliente-agente': this.cliente.agente.slice(0, 200),
				...(token ? { authorization: `Bearer ${token}` } : {})
			},
			body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo)
		});
		if (res.status === 204 || res.status === 202) return undefined as T;
		const json = (await res.json().catch(() => null)) as (T & Problema) | null;
		if (!res.ok) {
			const campos = Object.fromEntries((json?.issues ?? []).map((i) => [i.path, i.message]));
			throw new ErrorCuenta(
				res.status,
				json?.code ?? 'error',
				json?.detail ?? 'No pudimos completar el pedido.',
				campos
			);
		}
		return json as T;
	}

	registrar(a: components['schemas']['Alta']) {
		return this.pedir<Sesion>('POST', '/v1/auth/registro', a);
	}
	ingresar(email: string, clave: string, recordarme: boolean) {
		return this.pedir<Sesion>('POST', '/v1/auth/ingresar', { email, clave, recordarme });
	}
	sesion(token: string) {
		return this.pedir<DatosSesion>('GET', '/v1/auth/sesion', undefined, token);
	}
	salir(token: string) {
		return this.pedir<void>('POST', '/v1/auth/salir', undefined, token);
	}
	recuperar(email: string) {
		return this.pedir<void>('POST', '/v1/auth/recuperar', { email });
	}
	restablecer(token: string, clave: string) {
		return this.pedir<void>('POST', '/v1/auth/restablecer', { token, clave });
	}
	/** Botón de arrepentimiento: público, pero pasa por el BFF para que Go vea la IP real. */
	arrepentimiento(p: components['schemas']['PedidoArrepentimiento']) {
		return this.pedir<{ codigo: string }>('POST', '/v1/arrepentimientos', p);
	}
}
