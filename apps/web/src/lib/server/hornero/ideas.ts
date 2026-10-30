/**
 * Las ideas de Tincho (handoff: "Encontré 3 cosas para hacer"), armadas con el tablero de la tienda:
 * recuperar carritos, pausar lo que no tiene stock y una promo de fin de semana.
 */
import type { Idea, Tablero } from '../datos/tipos.ts';

export type IdeaHornero = {
	id: Idea;
	titulo: string;
	detalle: string;
	/** Costo o impacto (handoff: la línea ámbar). */
	nota: string;
	hecha: boolean;
	/** Toast al aprobar. */
	listo: { titulo: string; detalle: string };
	/** Sección donde se edita antes de aprobar. */
	seccion: string;
};

const pesos = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');
const plural = (n: number, uno: string, varios: string) => `${n} ${n === 1 ? uno : varios}`;

export function ideasDe(t: Tablero): IdeaHornero[] {
	const ideas: IdeaHornero[] = [];
	const hechas = new Set(t.hornero.ideasHechas);
	const { carritos, carritosMonto, costoMensaje, audiencia } = t.hornero;
	if (carritos)
		ideas.push({
			id: 'carritos',
			titulo: `Recuperar ${plural(carritos, 'carrito', 'carritos')}`,
			detalle: 'Mensaje por WhatsApp a quienes no terminaron la compra',
			nota: `Costo ${pesos(carritos * costoMensaje)} · recupera ~${pesos(carritosMonto * 0.25)}`,
			hecha: hechas.has('carritos'),
			listo: {
				titulo: 'Recordatorios enviados',
				detalle: `${plural(carritos, 'mensaje', 'mensajes')} en cola por WhatsApp`
			},
			seccion: 'Carritos abandonados'
		});
	const { sinStock, sinStockEnMl } = t.paraHacer;
	if (sinStock)
		ideas.push({
			id: 'stock',
			titulo: `Pausar ${plural(sinStock, 'producto sin stock', 'productos sin stock')}`,
			detalle: sinStockEnMl
				? 'Para no vender lo que no tenés en Mercado Libre'
				: 'Para no vender lo que no tenés',
			nota: 'Evita reclamos y baja de reputación',
			hecha: hechas.has('stock'),
			listo: {
				titulo: sinStock === 1 ? 'Producto pausado' : 'Productos pausados',
				detalle: 'Podés deshacerlo durante 30 días'
			},
			seccion: 'Inventario'
		});
	const top = t.masVendidos[0];
	if (top && audiencia)
		ideas.push({
			id: 'promo',
			titulo: 'Promo de fin de semana',
			detalle: `${top.nombre} 10% off para ${audiencia === 1 ? 'tu cliente' : `tus ${audiencia} clientes`} de WhatsApp`,
			nota: `Costo ${pesos(audiencia * costoMensaje)} · sale el sábado a las 10 h`,
			hecha: hechas.has('promo'),
			listo: { titulo: 'Campaña programada', detalle: 'Sale el sábado a las 10 h' },
			seccion: 'Marketing'
		});
	return ideas;
}
