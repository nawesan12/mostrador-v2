/**
 * Avisos de /estado a los suscriptos cuando el monitor abre o cierra un incidente. Por email con
 * Resend (si mostrador-jobs tiene RESEND_API_KEY); por WhatsApp quedan para cuando esté WhatsApp
 * Cloud (por ahora se registran en el log). Van de a 50 por pasada para acotar el cron.
 */
import type { Cambio } from './monitor.ts';

const esc = (s: string) => s.replace(/[&<>"]/g, (c) => `&#${c.charCodeAt(0)};`);

export async function avisar(
	db: D1Database,
	cambios: Cambio[],
	cfg: { resend?: string; de: string; sitio: string },
	pedir: typeof fetch = (i, init) => fetch(i, init)
) {
	if (!cambios.length) return { enviados: 0 };
	const subs = (
		await db
			.prepare('SELECT contacto, canal, token_baja FROM suscripciones ORDER BY creada_en LIMIT 50')
			.all<{ contacto: string; canal: 'email' | 'whatsapp'; token_baja: string }>()
	).results;
	let enviados = 0;
	for (const c of cambios) {
		for (const s of subs) {
			if (s.canal !== 'email' || !cfg.resend) {
				console.log('estado: aviso sin canal configurado', s.canal, c.tipo, c.componente);
				continue;
			}
			const baja = `${cfg.sitio}/estado/baja?token=${s.token_baja}`;
			const r = await pedir('https://api.resend.com/emails', {
				method: 'POST',
				headers: { authorization: `Bearer ${cfg.resend}`, 'content-type': 'application/json' },
				body: JSON.stringify({
					from: cfg.de,
					to: [s.contacto],
					subject: `${c.tipo === 'abre' ? '⚠️' : '✅'} ${c.titulo}`,
					text: `${c.detalle}\n\nSeguilo en ${cfg.sitio}/estado\nPara no recibir más avisos: ${baja}`,
					html: `<p style="font-family:Helvetica,Arial,sans-serif;font-size:15px;color:#2A140C">${esc(c.detalle)}</p><p style="font-family:Helvetica,Arial,sans-serif;font-size:14px"><a href="${cfg.sitio}/estado" style="color:#C24E25">Ver el estado del sistema</a> · <a href="${esc(baja)}" style="color:#7A6252">No recibir más avisos</a></p>`
				})
			});
			if (r.ok) enviados++;
			else console.log('estado: Resend respondió', r.status);
		}
	}
	return { enviados };
}
