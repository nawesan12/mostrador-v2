/**
 * Cliente de la API Go desde el BFF (DATA_SOURCE=api). El navegador nunca habla con Go: lo hace
 * este servidor, con la clave del BFF y la IP real del visitante en X-Cliente-IP, que es lo único
 * que hace que Go le crea. Los errores problem+json vuelven como ErrorApi, ya listos para el form.
 */
/** Error de la API, ya traducido para el formulario. */
export class ErrorApi extends Error {
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

export class ClienteBff {
	constructor(
		private base: string,
		private claveBFF: string,
		private cliente: { ip: string; agente: string },
		// Envuelto: en Workers, `fetch` llamado como método de otro objeto da "Illegal invocation".
		private f: typeof fetch = (i, init) => fetch(i, init)
	) {}

	protected async pedir<T>(
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
			throw new ErrorApi(
				res.status,
				json?.code ?? 'error',
				json?.detail ?? 'No pudimos completar el pedido.',
				campos
			);
		}
		return json as T;
	}
}
