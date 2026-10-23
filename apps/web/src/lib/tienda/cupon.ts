/**
 * Un cupón en la tienda online: si vale para este carrito y cuánto descuenta. Es la misma regla que
 * `apps/api/internal/dominio/vidriera/cupon.go` (los dos leen los casos de
 * `apps/api/internal/dominio/vidriera/testdata/cupones.json`): el navegador la usa para mostrar el
 * descuento, y el servidor la vuelve a correr al comprar (nunca se cree lo que dice el navegador).
 */
export type CuponTienda = {
	codigo: string;
	tipo: 'pct' | 'fijo' | 'envio' | 'regalo';
	valor: number;
	pausado: boolean;
	usos: number;
	/** 0: sin límite. */
	limite: number;
	canales: string[];
	empieza: string | null;
	termina: string | null;
	minimo: number;
	quien: 'todos' | 'nuevos' | 'vip' | 'riesgo';
	porCliente: '1' | '3' | 'libre';
	/** "categoria": la de `categoria` (si no quedó guardada, vale para toda la tienda). */
	alcance: { tipo: 'todo' | 'categoria' | 'productos'; productos: string[]; categoria?: string };
	/** Si se suma al descuento por transferencia. */
	combinable: boolean;
};

export type LineaCupon = { id: string; precio: number; cantidad: number; categoria?: string | null };

/** Lo que se sabe de quien compra, por su email (lo arma el servidor). */
export type Historial = {
	/** Pedidos anteriores no cancelados. */
	pedidos: number;
	/** Veces que ya usó este cupón. */
	usosDelCupon: number;
	vip: boolean;
	/** Fecha de la última compra (ISO) o null. */
	ultimaCompra: string | null;
};

export type ResultadoCupon =
	| { ok: true; descuento: number; envioGratis: boolean; combinable: boolean }
	| { ok: false; motivo: string };

const DIA = 86_400_000;
const no = (motivo: string): ResultadoCupon => ({ ok: false, motivo });

export function aplicarCupon(
	c: CuponTienda,
	lineas: LineaCupon[],
	h: Historial,
	ahora: Date,
	canal = 'web'
): ResultadoCupon {
	const t = ahora.getTime();
	if (c.pausado) return no('Este cupón está pausado.');
	if (c.empieza && t < Date.parse(c.empieza)) return no('Este cupón todavía no empezó.');
	if (c.termina && t > Date.parse(c.termina)) return no('Este cupón ya venció.');
	if (c.limite > 0 && c.usos >= c.limite) return no('Este cupón ya se usó todas las veces posibles.');
	if (!c.canales.includes(canal)) return no('Este cupón no vale en la tienda online.');
	if (c.tipo === 'regalo') return no('Este cupón se canjea en el local.');
	const subtotal = lineas.reduce((s, l) => s + l.precio * l.cantidad, 0);
	if (subtotal < c.minimo)
		return no(`Este cupón es para compras desde $ ${c.minimo.toLocaleString('es-AR')}.`);
	if (c.quien === 'nuevos' && h.pedidos > 0) return no('Este cupón es para la primera compra.');
	if (c.quien === 'vip' && !h.vip) return no('Este cupón es para clientes VIP.');
	if (
		c.quien === 'riesgo' &&
		!(h.pedidos > 0 && h.ultimaCompra && t - Date.parse(h.ultimaCompra) > 60 * DIA)
	)
		return no('Este cupón es para quienes hace un tiempo que no compran.');
	const tope = c.porCliente === 'libre' ? Infinity : Number(c.porCliente);
	if (h.usosDelCupon >= tope)
		return no(
			c.porCliente === '1' ? 'Ya usaste este cupón.' : 'Ya usaste este cupón las veces posibles.'
		);
	const vale = (l: LineaCupon) =>
		c.alcance.tipo === 'productos'
			? c.alcance.productos.includes(l.id)
			: c.alcance.tipo === 'categoria' && c.alcance.categoria
				? l.categoria === c.alcance.categoria
				: true;
	const elegible = lineas.filter(vale).reduce((s, l) => s + l.precio * l.cantidad, 0);
	if (elegible === 0) return no('Este cupón no vale para los productos de tu carrito.');
	if (c.tipo === 'envio') return { ok: true, descuento: 0, envioGratis: true, combinable: c.combinable };
	const descuento =
		c.tipo === 'pct'
			? Math.round((elegible * Math.min(100, Math.max(0, c.valor))) / 100)
			: Math.min(elegible, Math.max(0, c.valor));
	return { ok: true, descuento, envioGratis: false, combinable: c.combinable };
}

export const normalizarCodigo = (s: string) => s.trim().toUpperCase().slice(0, 21);
