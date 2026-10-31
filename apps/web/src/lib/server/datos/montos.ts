/**
 * Los montos del Inicio por rol: Ventas y Local (sin el permiso finanzas) no reciben montos de
 * dinero (ventas, serie, ticket, canales, meta, caja, liquidaciones, montos de transferencias y
 * carritos, totales de los últimos pedidos ni lo cobrado por Mercado Pago). La API Go los saca
 * (tiendas.SinMontos); el sandbox de la demo hace lo mismo con estas funciones. No van en cero:
 * se omiten y `montos` dice si vienen.
 */
import type { ContextoHornero, Tablero } from './tipos.ts';

/** El tablero con todos sus montos (lo que arma el sandbox antes de mirar el rol). */
export type TableroConMontos = Tablero &
	Required<Pick<Tablero, 'ventas' | 'serie' | 'canales' | 'meta' | 'caja' | 'liquidaciones'>> & {
		kpis: Required<Tablero['kpis']>;
		paraHacer: Required<Tablero['paraHacer']>;
		enVivo: Required<Tablero['enVivo']>;
		hornero: Required<Tablero['hornero']>;
		ultimosPedidos: Required<Tablero['ultimosPedidos'][number]>[];
	};

/** El contexto de Tincho con las ventas en pesos. */
export type ContextoConMontos = ContextoHornero & {
	ventasPorDia: Required<ContextoHornero['ventasPorDia'][number]>[];
	canales30d: Required<ContextoHornero['canales30d'][number]>[];
};

export function sinMontos(t: Tablero): Tablero {
	const s: Tablero = structuredClone(t);
	s.montos = false;
	delete s.ventas;
	delete s.serie;
	delete s.canales;
	delete s.meta;
	delete s.caja;
	delete s.liquidaciones;
	delete s.kpis.ticket;
	delete s.paraHacer.transferenciasMonto;
	delete s.enVivo.carritosMonto;
	delete s.hornero.carritosMonto;
	for (const p of s.ultimosPedidos) delete p.total;
	// Mercado Pago muestra lo cobrado en 7 días: es un monto.
	s.integraciones = s.integraciones.filter((i) => i.id !== 'mp');
	return s;
}

/** El contexto de Tincho sin las ventas en pesos (quedan los pedidos). */
export function contextoSinMontos(c: ContextoHornero): ContextoHornero {
	const s: ContextoHornero = structuredClone(c);
	s.montos = false;
	for (const d of s.ventasPorDia) delete d.ventas;
	for (const x of s.canales30d) delete x.ventas;
	return s;
}
