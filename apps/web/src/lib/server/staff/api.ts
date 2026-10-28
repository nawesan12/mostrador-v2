/** El backoffice del equipo contra la API Go, con el token de la sesión del equipo. */
import { ClienteBff, ErrorApi } from '../bff.ts';
import type { FiltrosLista } from '../sandbox/listas.ts';
import type * as T from './tipos.ts';

const e = encodeURIComponent;

/** Ingreso y salida (sin sesión todavía). */
export class AccesoEquipoApi extends ClienteBff {
	ingresar(email: string, clave: string) {
		return this.pedir<T.SesionNueva>('POST', '/v1/equipo/ingresar', { email, clave });
	}
	salir(token: string) {
		return this.pedir<void>('POST', '/v1/equipo/salir', undefined, token);
	}
}

export class EquipoApi extends ClienteBff implements T.DatosEquipo {
	readonly demo = false;
	constructor(
		base: string,
		claveBFF: string,
		cliente: { ip: string; agente: string },
		private token: string,
		f?: typeof fetch
	) {
		super(base, claveBFF, cliente, f);
	}
	private get<R>(ruta: string) {
		return this.pedir<R>('GET', ruta, undefined, this.token);
	}
	private post<R>(ruta: string, cuerpo: unknown) {
		return this.pedir<R>('POST', ruta, cuerpo, this.token);
	}
	/** null si no existe (404). */
	private async quizas<R>(ruta: string) {
		try {
			return await this.get<R>(ruta);
		} catch (err) {
			if (err instanceof ErrorApi && err.estado === 404) return null;
			throw err;
		}
	}

	sesion() {
		return this.get<T.SesionEquipo>('/v1/equipo/sesion');
	}
	inicio() {
		return this.get<T.InicioEquipo>('/v1/equipo/inicio');
	}
	tiendas(f: FiltrosLista) {
		const q = new URLSearchParams();
		if (f.q) q.set('q', f.q);
		if (f.seg?.length) q.set('seg', f.seg.join(','));
		const fs = Object.entries(f.facetas ?? {}).flatMap(([k, vs]) => vs.map((v) => `${k}:${v}`));
		if (fs.length) q.set('facetas', fs.join(','));
		if (f.rango) q.set('rango', String(f.rango));
		if (f.orden) q.set('orden', f.orden);
		q.set('desde', String(f.desde ?? 0));
		q.set('cuantos', String(f.cuantos ?? 60));
		return this.get<T.ListaGenerica>(`/v1/equipo/tiendas?${q}`);
	}
	ficha(id: string) {
		return this.quizas<T.FichaTienda>(`/v1/equipo/tiendas/${e(id)}`);
	}
	accion(id: string, a: T.AccionTienda) {
		return this.post<void>(`/v1/equipo/tiendas/${e(id)}/acciones`, a);
	}
	nota(id: string, texto: string) {
		return this.post<T.NotaEquipo>(`/v1/equipo/tiendas/${e(id)}/notas`, { texto });
	}
	mirar(id: string) {
		return this.post<T.SesionNueva>(`/v1/equipo/tiendas/${e(id)}/mirar`, {});
	}
	consultas() {
		return this.get<T.BandejaEquipo>('/v1/equipo/consultas');
	}
	consulta(tipo: T.TipoConsulta, id: string) {
		return this.quizas<T.ConsultaDetalle>(`/v1/equipo/consultas/${tipo}/${e(id)}`);
	}
	actualizarConsulta(
		tipo: T.TipoConsulta,
		id: string,
		cambio: { estado?: T.EstadoConsulta; asignadoA?: string | null }
	) {
		return this.post<void>(`/v1/equipo/consultas/${tipo}/${e(id)}`, cambio);
	}
	responder(tipo: T.TipoConsulta, id: string, r: { asunto: string; texto: string }) {
		return this.post<T.RespuestaEquipo>(`/v1/equipo/consultas/${tipo}/${e(id)}/responder`, r);
	}
	partners() {
		return this.get<T.PartnersEquipo>('/v1/equipo/partners');
	}
	decidir(id: string, decision: 'aprobar' | 'rechazar', motivo: string) {
		return this.post<void>(`/v1/equipo/partners/${e(id)}`, { decision, motivo });
	}
	facturacion() {
		return this.get<T.FacturacionEquipo>('/v1/equipo/facturacion');
	}
	auditoria(f: T.FiltrosAuditoria) {
		const q = new URLSearchParams();
		for (const [k, v] of Object.entries(f)) if (v) q.set(k, v);
		return this.get<T.AuditoriaEquipo>(`/v1/equipo/auditoria?${q}`);
	}
}
