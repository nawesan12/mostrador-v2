/**
 * Lo que se hace con un pedido sin abrirlo: la acción de un toque de la fila (pasa a la etapa
 * siguiente con el mismo POST /estado del detalle) y el WhatsApp para quien compró, con el mensaje
 * según cómo está el pedido de verdad.
 */
import * as I from '#lib/icons.ts';
import type { IconData } from '#lib/icons.ts';
import { SITE } from '#lib/site.ts';
import { linkWhatsapp } from '#lib/whatsapp.ts';
import { pesos } from '../formato.ts';
import type { EstadoPedido } from './filtros.ts';

type ConEstado = { estado: EstadoPedido; envio: string; pago: string };

export type AccionRapida = {
	/** El texto del botón (corto: entra en la fila). */
	texto: string;
	/** La etapa a la que pasa. */
	a: EstadoPedido;
	icono: IconData;
	fondo: string;
	color: string;
};

/**
 * La acción principal de la fila según el estado y la entrega. Sin pagar con Mercado Pago no se
 * marca a mano: lo confirma el aviso de Mercado Pago (para reclamarlo está el WhatsApp).
 */
export function accionRapida(p: ConEstado): AccionRapida | null {
	const retiro = p.envio === 'retiro';
	switch (p.estado) {
		case 'Sin pagar':
			return p.pago === 'mp'
				? null
				: {
						texto: 'Marcar pagado',
						a: 'Para preparar',
						icono: I.phCurrencyCircleDollarFill,
						fondo: '#E9F8EE',
						color: '#1F6F4A'
					};
		case 'Para preparar':
			return {
				texto: retiro ? 'Listo para retirar' : 'Preparar',
				a: 'Listo para despachar',
				icono: retiro ? I.phStorefrontFill : I.phPackageFill,
				fondo: '#2A140C',
				color: '#fff'
			};
		case 'Listo para despachar':
			return retiro
				? {
						texto: 'Retirado',
						a: 'Entregado',
						icono: I.phCheckCircleFill,
						fondo: 'rgba(47,143,99,.14)',
						color: '#237550'
					}
				: {
						texto: 'Despachado',
						a: 'En camino',
						icono: I.phTruckFill,
						fondo: 'rgba(210,88,47,.14)',
						color: '#C24E25'
					};
		case 'En camino':
			return {
				texto: 'Entregado',
				a: 'Entregado',
				icono: I.phCheckCircleFill,
				fondo: 'rgba(47,143,99,.14)',
				color: '#237550'
			};
		default:
			return null;
	}
}

/** El aviso después de la acción de un toque (con «Deshacer»). */
export function avisoDe(p: ConEstado & { numero: number; cliente: string }) {
	const nombre = primerNombre(p.cliente);
	const quien = nombre ? `a ${nombre}` : 'a quien compró';
	const retiro = p.envio === 'retiro';
	switch (p.estado) {
		case 'Sin pagar':
			return { title: `#${p.numero} pagado`, desc: 'Pasó a preparar' };
		case 'Para preparar':
			return retiro
				? { title: `#${p.numero} listo para retirar`, desc: `Avisale ${quien} por WhatsApp` }
				: { title: `#${p.numero} preparado`, desc: 'Queda listo para despachar' };
		case 'Listo para despachar':
			return retiro
				? { title: `#${p.numero} retirado`, desc: `${nombre || 'Ya'} pasó a buscarlo` }
				: { title: `#${p.numero} despachado`, desc: `Avisale ${quien} por WhatsApp` };
		default:
			return { title: `#${p.numero} entregado`, desc: 'Pedido cerrado' };
	}
}

/** «Ana» de «Ana Pérez»; nada si el pedido no tiene a quién (consumidor final). */
export function primerNombre(cliente: string) {
	const n = cliente.trim();
	if (!n || /^consumidor final$/i.test(n)) return '';
	return n.split(/\s+/)[0];
}

export type Contexto = {
	/** Dónde y cuándo se retira (Envíos y retiro; vacío: no se dice). */
	retiro: { direccion: string; horario: string };
	/** La dirección de la tienda, para el link de la reseña. */
	slug: string;
	/** El número de seguimiento, si se sabe (el detalle lo trae). */
	seguimiento?: string | null;
};

type ParaMensaje = ConEstado & {
	numero: number;
	cliente: string;
	total: number;
	codigo: string | null;
};

/** El link de la confirmación del pedido, donde quien compró deja la reseña. */
export const linkResena = (slug: string, codigo: string) =>
	`${SITE}/t/${slug}/pedido/${encodeURIComponent(codigo)}`;

/** El mensaje de WhatsApp para quien compró, según el estado del pedido. */
export function mensajeWhatsapp(p: ParaMensaje, c: Contexto) {
	const nombre = primerNombre(p.cliente);
	const hola = nombre ? `¡Hola ${nombre}!` : '¡Hola!';
	const n = `#${p.numero}`;
	switch (p.estado) {
		case 'Sin pagar':
			return `${hola} Te escribimos por tu pedido ${n} de ${pesos(p.total)}: todavía no nos llegó el pago. ¿Te ayudamos con algo?`;
		case 'Para preparar':
			return `${hola} Recibimos tu pago del pedido ${n}. Ya lo estamos preparando.`;
		case 'Listo para despachar': {
			if (p.envio !== 'retiro')
				return `${hola} Tu pedido ${n} ya está preparado. Te avisamos cuando salga.`;
			const dir = c.retiro.direccion.trim();
			const hora = c.retiro.horario.trim().replace(/\.$/, '');
			const donde = dir ? ` en ${dir}${hora ? ` (${hora})` : ''}` : '';
			return `${hola} Tu pedido ${n} está listo para retirar${donde}.`;
		}
		case 'En camino': {
			const base =
				p.envio === 'moto'
					? `${hola} Tu pedido ${n} ya salió para tu casa.`
					: `${hola} Tu pedido ${n} ya está en camino.`;
			return c.seguimiento ? `${base} Número de seguimiento: ${c.seguimiento}.` : base;
		}
		case 'Entregado':
			return p.codigo
				? `${hola} ¿Te pudimos ayudar? Si querés, dejanos una reseña: ${linkResena(c.slug, p.codigo)}`
				: `${hola} Gracias por tu compra. ¿Te pudimos ayudar con tu pedido ${n}?`;
		default:
			return `${hola} Te escribimos por tu pedido ${n}.`;
	}
}

/** wa.me con el mensaje armado; null si el pedido no tiene un teléfono que sirva. */
export function whatsappDe(p: ParaMensaje & { telefono: string | null }, c: Contexto) {
	return p.telefono ? linkWhatsapp(p.telefono, mensajeWhatsapp(p, c)) : null;
}
