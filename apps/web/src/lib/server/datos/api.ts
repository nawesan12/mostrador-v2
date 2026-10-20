/**
 * Datos desde la API Go (DATA_SOURCE=api). El navegador nunca habla con Go: lo hace este servidor
 * (BFF), con el token de la sesión. Los errores problem+json se convierten en errores de SvelteKit.
 */
import { error } from '@sveltejs/kit';
import type { Datos, Rubro, Tienda } from './tipos.ts';

export class DatosApi implements Datos {
	constructor(
		private base: string,
		private token: string | undefined,
		private f: typeof fetch = fetch
	) {}

	private async pedir<T>(ruta: string): Promise<T> {
		const res = await this.f(this.base + ruta, {
			headers: {
				accept: 'application/json',
				...(this.token ? { authorization: `Bearer ${this.token}` } : {})
			}
		});
		if (!res.ok) {
			const p = (await res.json().catch(() => null)) as { detail?: string; code?: string } | null;
			error(res.status, { message: p?.detail ?? 'La API no respondió bien.', code: p?.code });
		}
		return (await res.json()) as T;
	}

	rubros() {
		return this.pedir<Rubro[]>('/v1/rubros');
	}

	tienda() {
		return this.pedir<Tienda>('/v1/panel/tienda');
	}
}
