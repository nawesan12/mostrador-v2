/**
 * Datos desde la API Go (DATA_SOURCE=api). El navegador nunca habla con Go: lo hace este servidor
 * (BFF), con el token de la sesión. Los errores problem+json se convierten en errores de SvelteKit.
 */
import { error } from '@sveltejs/kit';
import type {
	AccionLote,
	FiltrosLista,
	ListaGenerica,
	NombreLista,
	Producto,
	ProductoEditable,
	ContextoHornero,
	EstadoPedido,
	FiltrosPedidos,
	ListaPedidos,
	PedidoDetalle,
	Resultado,
	Datos,
	Idea,
	Onboarding,
	Periodo,
	ResumenPanel,
	Rubro,
	Tablero,
	Tienda
} from './tipos.ts';

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

	contextoHornero() {
		return this.pedir<ContextoHornero>('/v1/panel/hornero/contexto');
	}
	marcarIdea(idea: Idea, hecha: boolean) {
		return this.pedir<void>(`/v1/panel/hornero/ideas/${idea}`, 'POST', { hecha });
	}

	listarPedidos(f: FiltrosPedidos) {
		const q = new URLSearchParams();
		for (const [k, v] of Object.entries(f)) {
			if (v === undefined || v === '' || v === false || (Array.isArray(v) && !v.length)) continue;
			q.set(k, Array.isArray(v) ? v.join(',') : String(v));
		}
		return this.pedir<ListaPedidos>(`/v1/panel/pedidos?${q}`);
	}
	async pedido(numero: number) {
		const r = await this.intentar(`/v1/panel/pedidos/${numero}`, 'GET');
		return r.ok ? ((await r.json()) as PedidoDetalle) : null;
	}
	async cambiarEstado(numero: number, estado: EstadoPedido) {
		return resultado(await this.intentar(`/v1/panel/pedidos/${numero}/estado`, 'POST', { estado }));
	}
	pedidosEnLote(numeros: number[], accion: AccionLote) {
		return this.pedir<{ hechos: number; cambios: { numero: number; antes: EstadoPedido }[] }>(
			'/v1/panel/pedidos/lote',
			'POST',
			{ numeros, accion }
		);
	}
	async recordarPago(numero: number) {
		return resultado(await this.intentar(`/v1/panel/pedidos/${numero}/recordatorio`, 'POST', {}));
	}

	listar(lista: NombreLista, f: FiltrosLista) {
		const q = new URLSearchParams();
		if (f.q) q.set('q', f.q);
		if (f.seg?.length) q.set('seg', f.seg.join(','));
		const facetas = Object.entries(f.facetas ?? {}).flatMap(([k, vs]) =>
			vs.map((v) => `${k}:${v}`)
		);
		if (facetas.length) q.set('facetas', facetas.join(','));
		if (f.rango) q.set('rango', String(f.rango));
		if (f.banderas?.length) q.set('banderas', f.banderas.join(','));
		if (f.orden) q.set('orden', f.orden);
		q.set('desde', String(f.desde ?? 0));
		q.set('cuantos', String(f.cuantos ?? 60));
		return this.pedir<ListaGenerica>(`/v1/panel/listas/${lista}?${q}`);
	}
	async accionLista(lista: NombreLista, ids: string[], accion: string, valor?: unknown) {
		return (
			await this.pedir<{ hechos: number }>(`/v1/panel/listas/${lista}/acciones`, 'POST', {
				ids,
				accion,
				valor
			})
		).hechos;
	}
	async producto(id: string) {
		const r = await this.intentar(`/v1/panel/productos/${encodeURIComponent(id)}`, 'GET');
		return r.ok ? ((await r.json()) as Producto) : null;
	}
	async guardarProducto(id: string | null, p: ProductoEditable) {
		if (!id) return this.pedir<Producto>('/v1/panel/productos', 'POST', p);
		const r = await this.intentar(`/v1/panel/productos/${encodeURIComponent(id)}`, 'PUT', p);
		return r.ok ? ((await r.json()) as Producto) : null;
	}

	/** Como `pedir`, pero 404 y 409 vuelven como respuesta (son resultados, no errores). */
	private async intentar(ruta: string, metodo: 'GET' | 'POST' | 'PUT', cuerpo?: unknown) {
		const res = await this.f(this.base + ruta, {
			method: metodo,
			headers: {
				accept: 'application/json',
				...(cuerpo === undefined ? {} : { 'content-type': 'application/json' }),
				...(this.token ? { authorization: `Bearer ${this.token}` } : {})
			},
			body: cuerpo === undefined ? undefined : JSON.stringify(cuerpo)
		});
		if (!res.ok && res.status !== 404 && res.status !== 409) {
			const p = (await res.json().catch(() => null)) as { detail?: string; code?: string } | null;
			error(res.status, { message: p?.detail ?? 'La API no respondió bien.', code: p?.code });
		}
		return res;
	}
}

const resultado = (r: Response): Resultado =>
	r.ok ? 'ok' : r.status === 404 ? 'no-existe' : 'no-se-puede';
