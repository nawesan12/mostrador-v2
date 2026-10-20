import { env } from 'cloudflare:workers';
import { fail } from '@sveltejs/kit';
import { leerEstado } from '#lib/server/estado/lectura.ts';
import { randomToken } from '#lib/server/crypto.ts';
import type { Actions, PageServerLoad } from './$types';

const RANGOS = [90, 30, 7] as const;

export const load: PageServerLoad = async ({ url, setHeaders }) => {
	const pedido = Number(url.searchParams.get('dias'));
	const dias = RANGOS.find((r) => r === pedido) ?? 90;
	// El monitor escribe una vez por hora: un minuto de caché alcanza y ahorra lecturas a D1.
	setHeaders({ 'cache-control': 'public, max-age=60' });
	return { dias, estado: await leerEstado(env.ESTADO, dias), ahora: Date.now() };
};

const EMAIL = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;

export const actions: Actions = {
	/** "Enterate antes que tus clientes": email o WhatsApp para los avisos de incidentes. */
	suscribir: async ({ request }) => {
		const contacto = String((await request.formData()).get('contacto') ?? '').trim();
		let valor = contacto.toLowerCase();
		let canal: 'email' | 'whatsapp' = 'email';
		if (!EMAIL.test(valor) || valor.length > 254) {
			valor = contacto.replace(/\D/g, '');
			canal = 'whatsapp';
			if (valor.length < 8 || valor.length > 15)
				return fail(400, { contacto, error: 'Poné un email o un WhatsApp con característica.' });
		}
		await env.ESTADO.prepare(
			`INSERT INTO suscripciones (contacto, canal, token_baja, creada_en) VALUES (?1, ?2, ?3, ?4)
			 ON CONFLICT (contacto) DO NOTHING`
		)
			.bind(valor, canal, randomToken(24), Date.now())
			.run();
		return { suscripto: canal };
	}
};
