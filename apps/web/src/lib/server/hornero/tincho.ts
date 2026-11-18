/**
 * Tincho con IA: lo que el motor no resuelve lo contesta la API Go con Claude Haiku
 * (docs-ia/asistente-ia.md). Acá se arma la respuesta para el panel, las apps y Telegram: el texto,
 * las propuestas para confirmar (cupón, estado, stock, link de pago) y, si llegó a un tope, el aviso
 * con las preguntas que el motor sí sabe responder.
 */
import type { RespuestaTincho } from '../datos/tipos.ts';
import { ayudaConBotones, type Respuesta } from './motor.ts';

/** Si la IA falla, el motor contesta lo que puede y lo dice. */
export const NO_DISPONIBLE = 'No pude consultar a la IA, probá de nuevo.';

/** La respuesta de la API como la dibujan el panel y las apps. */
export function conIa(r: RespuestaTincho): Respuesta {
	if (r.fuente === 'tope')
		return { ...ayudaConBotones(), texto: r.texto, tope: r.tope, uso: r.uso, propuestas: [] };
	return { texto: r.texto, propuestas: r.propuestas, uso: r.uso };
}

/** En Telegram no hay botón «Confirmar»: las propuestas se nombran y se confirman en la app. */
export function paraTelegram(r: RespuestaTincho) {
	const t = r.texto.replace(/\*\*/g, '');
	if (!r.propuestas.length) return t;
	const lista = r.propuestas
		.map((p) => `• ${p.titulo}${p.detalle ? ` (${p.detalle})` : ''}`)
		.join('\n');
	return `${t}\n\n${lista}\nConfirmalo desde Tincho en la app o el panel.`;
}
