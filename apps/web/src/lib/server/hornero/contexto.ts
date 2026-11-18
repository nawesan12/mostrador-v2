/**
 * Lo que el motor de Tincho sabe para responder: la tienda, el tablero de hoy y de la semana, ventas
 * por día, productos, clientes y canales. La IA no usa esto: en la API Go arma su propio resumen y
 * pide el detalle con herramientas.
 */
import type { ContextoHornero, Tablero, Tienda } from '../datos/tipos.ts';

export type DatosHornero = {
	tienda: Pick<Tienda, 'nombre' | 'rubro'>;
	hola: string | null;
	hoy: Tablero;
	semana: Tablero;
	/** Los últimos 30 días (para «¿cuánto vendí este mes?»); opcional para Telegram y la demo. */
	mes?: Tablero;
	contexto: ContextoHornero;
};

/** ¿Quien pregunta ve montos? (Ventas y Local no: la API los omite y `montos` viene en false). */
export const conMontos = (d: DatosHornero) => d.hoy.montos && d.semana.montos;
