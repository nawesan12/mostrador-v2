/**
 * Datos desde la API Go (DATA_SOURCE=api). El navegador nunca habla con Go: lo hace este servidor
 * (BFF), con el token de la sesión. Los errores problem+json se convierten en errores de SvelteKit.
 */
import { error } from '@sveltejs/kit';
import type { Datos, Onboarding, Periodo, ResumenPanel, Rubro, Tablero, Tienda } from './tipos.ts';

export class DatosApi implements Datos {
	constructor(
		private base: string,
		private token: string | undefined,
		// Envuelto: en Workers, `fetch` llamado como método de otro objeto da "Illegal invocation".
		private f: typeof fetch = (i, init) => fetch(i, init)
	) {}

	private async pedir<T>(
		ruta: string,
		metodo: 'GET' | 'PUT' | 'POST' = 'GET',
		cuerpo?: unknown
	): Promise<T> {
		const res = await this.f(this.base + ruta, {
			method: metodo,
			headers: {
				accept: 'application/json',
				...(cuerpo === undefined ? {} : { 'content-type': 'application/json' }),
				...(this.token ? { authorization: `Bearer ${this.token}` } : {})
			},
			body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo)
		});
		if (!res.ok) {
			const p = (await res.json().catch(() => null)) as { detail?: string; code?: string } | null;
			error(res.status, { message: p?.detail ?? 'La API no respondió bien.', code: p?.code });
		}
		if (res.status === 204) return undefined as T;
		return (await res.json()) as T;
	}

	rubros() {
		return this.pedir<Rubro[]>('/v1/rubros');
	}

	tienda() {
		return this.pedir<Tienda>('/v1/panel/tienda');
	}

	guardarOnboarding(o: Onboarding) {
		return this.pedir<void>('/v1/panel/onboarding', 'PUT', o);
	}

	resumen() {
		return this.pedir<ResumenPanel>('/v1/panel/resumen');
	}
	marcarNotificacionesLeidas() {
		return this.pedir<void>('/v1/panel/notificaciones/leidas', 'POST', {});
	}

	tablero(periodo: Periodo) {
		return this.pedir<Tablero>(`/v1/panel/tablero?periodo=${periodo}`);
	}
	guardarMeta(objetivo: number) {
		return this.pedir<void>('/v1/panel/meta', 'PUT', { objetivo });
	}
	ocultarAviso(aviso: 'certificado') {
		return this.pedir<void>(`/v1/panel/avisos/${aviso}/ocultar`, 'POST', {});
	}
	async recordarCarritos() {
		return (await this.pedir<{ enviados: number }>('/v1/panel/carritos/recordar', 'POST', {}))
			.enviados;
	}
}
