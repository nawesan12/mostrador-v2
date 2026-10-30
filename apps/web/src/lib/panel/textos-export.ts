import { NOMBRE_CANAL, NOMBRE_ENVIO, NOMBRE_PAGO } from './pedidos/nombres.ts';

/**
 * Lo que se exporta (CSV y Excel) sale en castellano: los códigos internos de canal, pago, envío,
 * estado y segmento se cambian por el texto que muestra el panel. La columna se reconoce por el
 * encabezado; un valor que no está en la tabla queda como vino.
 */
const ESTADO: Record<string, string> = {
	// Productos
	active: 'Publicado',
	out: 'Sin stock',
	paused: 'Pausado',
	draft: 'Borrador',
	// Facturación
	ok: 'Autorizada por ARCA',
	pend: 'Esperando a ARCA',
	rej: 'Rechazada',
	void: 'Anulada con nota de crédito',
	nofac: 'No facturada',
	pendiente: 'Esperando a ARCA',
	emitido: 'Autorizada por ARCA',
	rechazado: 'Rechazada',
	no_facturado: 'No facturada',
	// Envíos
	todo: 'Sin despachar',
	way: 'En camino',
	late: 'Demorado',
	done: 'Entregado',
	ret: 'Devuelto al remitente'
};
const SEGMENTO: Record<string, string> = {
	vip: 'VIP',
	risk: 'En riesgo',
	sleep: 'Dormido',
	new: 'Nuevo'
};
const CONVERSACION: Record<string, string> = { cliente: 'Cliente', hornero: 'Hornero', vos: 'Vos' };

const POR_COLUMNA: Record<string, Record<string, string>> = {
	Canal: NOMBRE_CANAL,
	Pago: NOMBRE_PAGO,
	Envío: NOMBRE_ENVIO,
	Transporte: NOMBRE_ENVIO,
	Estado: ESTADO,
	Segmento: SEGMENTO,
	De: CONVERSACION
};

/** Traduce las columnas conocidas de una tabla cuya primera fila es el encabezado. */
export function enCastellano<T>(filas: T[][]): T[][] {
	const [cab, ...resto] = filas;
	if (!cab) return filas;
	const tablas = cab.map((c) => POR_COLUMNA[String(c ?? '')]);
	if (!tablas.some(Boolean)) return filas;
	return [
		cab,
		...resto.map((f) =>
			f.map((v, i) => {
				const t = tablas[i];
				return t && typeof v === 'string' && v in t ? (t[v] as T) : v;
			})
		)
	];
}
