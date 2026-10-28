/**
 * Lo que el backoffice del equipo (/staff) le pide a su fuente de datos: la API Go (con la sesión
 * del equipo) o la demo (datos de ejemplo de sólo lectura). Mismo contrato: packages/contrato.
 */
import type { components } from '@mostrador/contrato/esquema';
import type { FiltrosLista } from '../sandbox/listas.ts';

type S = components['schemas'];
export type SesionEquipo = S['SesionEquipo'];
export type InicioEquipo = S['InicioEquipo'];
export type ListaGenerica = S['ListaGenerica'];
export type FilaTienda = S['FilaTiendaEquipo'];
export type FichaTienda = S['FichaTiendaEquipo'];
export type AccionTienda = S['AccionTiendaEquipo'];
export type NotaEquipo = S['NotaEquipo'];
export type BandejaEquipo = S['BandejaEquipo'];
export type ConsultaEquipo = S['ConsultaEquipo'];
export type ConsultaDetalle = S['ConsultaDetalle'];
export type TipoConsulta = S['TipoConsulta'];
export type EstadoConsulta = S['EstadoConsulta'];
export type RespuestaEquipo = S['RespuestaEquipo'];
export type PartnersEquipo = S['PartnersEquipo'];
export type FacturacionEquipo = S['FacturacionEquipo'];
export type AuditoriaEquipo = S['AuditoriaEquipo'];
export type SesionNueva = S['SesionNueva'];

export type FiltrosAuditoria = { miembro?: string; accion?: string; tienda?: string };

export interface DatosEquipo {
	/** La demo: datos de ejemplo, ninguna acción se aplica. */
	readonly demo: boolean;
	sesion(): Promise<SesionEquipo>;
	inicio(): Promise<InicioEquipo>;
	tiendas(f: FiltrosLista): Promise<ListaGenerica>;
	ficha(id: string): Promise<FichaTienda | null>;
	accion(id: string, a: AccionTienda): Promise<void>;
	nota(id: string, texto: string): Promise<NotaEquipo>;
	mirar(id: string): Promise<SesionNueva>;
	consultas(): Promise<BandejaEquipo>;
	consulta(tipo: TipoConsulta, id: string): Promise<ConsultaDetalle | null>;
	actualizarConsulta(
		tipo: TipoConsulta,
		id: string,
		cambio: { estado?: EstadoConsulta; asignadoA?: string | null }
	): Promise<void>;
	responder(
		tipo: TipoConsulta,
		id: string,
		r: { asunto: string; texto: string }
	): Promise<RespuestaEquipo>;
	partners(): Promise<PartnersEquipo>;
	decidir(id: string, decision: 'aprobar' | 'rechazar', motivo: string): Promise<void>;
	facturacion(): Promise<FacturacionEquipo>;
	auditoria(f: FiltrosAuditoria): Promise<AuditoriaEquipo>;
}
