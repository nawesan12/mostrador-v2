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
	/**
	 * Si Tincho la puede hacer desde el chat. Con la API real sólo se aprueba lo que existe de verdad
	 * (recordar carritos por mail); lo demás lleva a la sección donde se hace a mano.
	 */
	aprobable: boolean;
};

const pesos = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');
const plural = (n: number, uno: string, varios: string) => `${n} ${n === 1 ? uno : varios}`;

/**
 * `real`: la tienda usa la API (no la demo). Ahí los recordatorios de carritos van por mail (gratis) y
 * no hay WhatsApp para campañas, así que Tincho no promete costos ni recuperos que no existen.
 */
export function ideasDe(t: Tablero, real = false): IdeaHornero[] {
	const ideas: IdeaHornero[] = [];
	const hechas = new Set(t.hornero.ideasHechas);
	const { carritos, carritosMonto, costoMensaje, audiencia } = t.hornero;
	if (carritos)
		ideas.push(
			real
				? {
						id: 'carritos',
						titulo: `Recordar ${plural(carritos, 'carrito', 'carritos')}`,
						detalle: `Un mail con el link a su carrito a quienes dejaron ${pesos(carritosMonto)} sin comprar`,
						nota: 'Sin costo · a cada uno, como mucho uno por día',
						hecha: t.hornero.recordados,
						listo: {
							titulo: 'Recordatorios enviados',
							detalle: 'Si alguien vuelve y compra, lo ves en Carritos abandonados'
						},
						seccion: 'Carritos abandonados',
						aprobable: true
					}
				: {
						id: 'carritos',
						titulo: `Recuperar ${plural(carritos, 'carrito', 'carritos')}`,
						detalle: 'Mensaje por WhatsApp a quienes no terminaron la compra',
						nota: `Costo ${pesos(carritos * costoMensaje)}`,
						hecha: hechas.has('carritos'),
						listo: {
							titulo: 'Recordatorios enviados',
							detalle: `${plural(carritos, 'mensaje', 'mensajes')} en cola por WhatsApp`
						},
						seccion: 'Carritos abandonados',
						aprobable: true
					}
		);
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
			seccion: 'Inventario',
			aprobable: !real
		});
	const top = t.masVendidos[0];
	// La promo sale por WhatsApp: con la API real todavía no hay cómo mandarla.
	if (top && audiencia && !real)
		ideas.push({
			id: 'promo',
			titulo: 'Promo de fin de semana',
			detalle: `${top.nombre} 10% off para ${audiencia === 1 ? 'tu cliente' : `tus ${audiencia} clientes`} de WhatsApp`,
			nota: `Costo ${pesos(audiencia * costoMensaje)} · sale el sábado a las 10 h`,
			hecha: hechas.has('promo'),
			listo: { titulo: 'Campaña programada', detalle: 'Sale el sábado a las 10 h' },
			seccion: 'Marketing',
			aprobable: true
		});
	return ideas;
}
