/**
 * Tiempo real (docs/decisiones.md §44): las reglas puras del cliente, aparte para probarlas sin
 * navegador. El servidor manda sólo QUÉ cambió ({ tipo, id }); acá se decide qué datos de la
 * página volver a pedir (las dependencias `tienda:<cosa>` que declaran los load con `depends`).
 */

/** Lo que llega por el WebSocket. */
export type EventoTR = { tipo: string; id?: string; lote?: boolean };

/** Lo que se recarga, por familia de evento (el prefijo del tipo). */
const RECARGA: Record<string, readonly string[]> = {
	// Un pedido toca el Inicio (pedidos, para hacer, caja), las insignias, la lista y el detalle,
	// los clientes (salen de quien compra), envíos, facturación y finanzas.
	pedido: [
		'pedidos',
		'resumen',
		'tablero',
		'clientes',
		'envios',
		'facturacion',
		'finanzas',
		'parahacer'
	],
	pago: ['pedidos', 'resumen', 'tablero', 'finanzas'],
	producto: ['productos', 'inventario', 'pos', 'tablero'],
	cliente: ['clientes'],
	carrito: ['carritos', 'resumen', 'tablero'],
	cupon: ['cupones'],
	envio: ['envios', 'pedidos'],
	comprobante: ['facturacion', 'pedidos'],
	resena: ['resenas', 'tablero'],
	pregunta: ['resenas', 'tablero'],
	integracion: ['integraciones', 'tablero'],
	importacion: ['importacion', 'productos'],
	actividad: ['actividad'],
	mitienda: ['mitienda'],
	dominio: ['dominio'],
	plan: ['plan', 'resumen'],
	equipo: ['equipo'],
	tienda: ['armazon', 'configuracion'],
	visitas: ['tablero'],
	tarea: ['parahacer', 'tablero'],
	turno: ['turnos', 'tablero'],
	devolucion: ['devoluciones'],
	finanzas: ['finanzas', 'tablero'],
	sesion: ['sesiones']
};

/** Las dependencias (`tienda:<cosa>`) que hay que invalidar para un evento. */
export function claves(e: EventoTR): string[] {
	const familia = e.tipo.split('.')[0];
	return (RECARGA[familia] ?? []).map((c) => `tienda:${c}`);
}

/** Todas las dependencias que existen (para controlar que cada load use una válida). */
export const TODAS = [...new Set(Object.values(RECARGA).flat())].map((c) => `tienda:${c}`);

/**
 * Espera antes de reconectar: exponencial (1 s, 2 s, 4 s… hasta 30 s) con azar ("full jitter"
 * acotado a la mitad), así mil pestañas no vuelven todas en el mismo segundo después de un corte.
 */
export function espera(intento: number, azar: () => number = Math.random, tope = 30_000): number {
	const base = Math.min(tope, 1000 * 2 ** Math.max(0, intento));
	return Math.round(base / 2 + azar() * (base / 2));
}

/** Qué hacer cuando el servidor cierra (los códigos de apps/api/internal/plataforma/tiemporeal). */
export type TrasCierre = 'ya' | 'reintentar' | 'mas-tarde' | 'parar';
export function trasCierre(codigo: number): TrasCierre {
	switch (codigo) {
		case 4000: // cambió la sesión (tienda, rol): volver ya y recargar todo
		case 1012: // la API se reinicia
			return 'ya';
		case 4401: // sin sesión
		case 4403: // sin permiso (una IA)
		case 1008: // la tienda o el pedido no existen
			return 'parar';
		case 4429: // demasiadas conexiones
			return 'mas-tarde';
		default:
			return 'reintentar';
	}
}

/**
 * La URL del WebSocket. `origen` es el de la API si no es el mismo sitio (dominios propios,
 * pruebas); si no, el de la página, con ws/wss según corresponda.
 */
export function urlTiempoReal(
	pagina: { protocol: string; host: string },
	ruta: string,
	origen?: string | null
): string {
	if (origen) {
		const u = new URL(ruta, origen);
		u.protocol = u.protocol === 'https:' || u.protocol === 'wss:' ? 'wss:' : 'ws:';
		return u.href;
	}
	return `${pagina.protocol === 'https:' ? 'wss:' : 'ws:'}//${pagina.host}${ruta}`;
}

/** Lee un mensaje del servidor; null si no se entiende. */
export function leerEvento(datos: unknown): EventoTR | null {
	if (typeof datos !== 'string' || datos.length > 2000) return null;
	try {
		const e = JSON.parse(datos) as EventoTR;
		return e && typeof e.tipo === 'string' && e.tipo ? e : null;
	} catch {
		return null;
	}
}
