/**
 * Tincho en Telegram con la API real (DATA_SOURCE=api): el vínculo del chat vive en Go
 * (integraciones/telegram). Las rutas del panel van con la sesión de la tienda; las del bot, sólo con
 * la clave del BFF (el webhook ya verificó el secreto de Telegram). El modelo sigue en el Worker: Go
 * devuelve el contexto de la tienda del chat y cuántos mensajes lleva hoy.
 */
import type { components } from '@mostrador/contrato/esquema';
import { ClienteBff, ErrorApi } from './bff.ts';

export type EstadoTelegram = components['schemas']['EstadoTelegram'];
export type ContextoTelegram = components['schemas']['ContextoTelegram'];

/** Cómo terminó el /start con un código. */
export type Vinculo =
	{ ok: true; tienda: string } | { ok: false; motivo: 'codigo' | 'ocupado' | 'demasiados' };

const es404 = (e: unknown) => e instanceof ErrorApi && e.estado === 404;

export class TelegramApi extends ClienteBff {
	// Panel (sesión de la tienda).
	estado(sesion: string) {
		return this.pedir<EstadoTelegram>('GET', '/v1/panel/telegram', undefined, sesion);
	}
	/** Código de un solo uso para el link al bot (dueña o admin). */
	codigo(sesion: string) {
		return this.pedir<components['schemas']['CodigoTelegram']>(
			'POST',
			'/v1/panel/telegram/codigo',
			undefined,
			sesion
		);
	}
	/** Desvincula y devuelve el chat que estaba (null si no había), para avisarle. */
	async desvincular(sesion: string) {
		const r = await this.pedir<{ chat: number } | undefined>(
			'DELETE',
			'/v1/panel/telegram',
			undefined,
			sesion
		);
		return r?.chat ?? null;
	}

	// Bot (clave del BFF).
	async vincular(codigo: string, chat: number, nombre: string): Promise<Vinculo> {
		try {
			const r = await this.pedir<{ tienda: string }>('POST', '/v1/telegram/vincular', {
				codigo,
				chat,
				nombre
			});
			return { ok: true, tienda: r.tienda };
		} catch (e) {
			if (e instanceof ErrorApi && e.codigo === 'chat_ocupado')
				return { ok: false, motivo: 'ocupado' };
			if (e instanceof ErrorApi && e.codigo === 'codigo_invalido')
				return { ok: false, motivo: 'codigo' };
			if (e instanceof ErrorApi && e.estado === 429) return { ok: false, motivo: 'demasiados' };
			throw e;
		}
	}
	/** false si el chat no estaba vinculado. */
	async desvincularChat(chat: number) {
		try {
			await this.pedir<void>('POST', '/v1/telegram/desvincular', { chat });
			return true;
		} catch (e) {
			if (es404(e)) return false;
			throw e;
		}
	}
	/** El contexto de Tincho de la tienda del chat (suma el mensaje al día); null si no está vinculado. */
	async contexto(chat: number) {
		try {
			return await this.pedir<ContextoTelegram>('POST', '/v1/telegram/contexto', { chat });
		} catch (e) {
			if (es404(e)) return null;
			throw e;
		}
	}
}
