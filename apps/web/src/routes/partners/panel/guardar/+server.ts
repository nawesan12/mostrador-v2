import { error, json } from '@sveltejs/kit';
import { ErrorApi } from '#lib/server/bff.ts';
import { partnersDe } from '#lib/server/partners/index.ts';
import { validarCobro, validarFactura, validarPerfil } from '#lib/partners/validar.ts';
import type { RequestHandler } from './$types';

/**
 * Lo que el partner cambia desde el panel: perfil público, datos de cobro y la factura de un pago.
 * Se valida acá con la misma regla que Go (que vuelve a validar).
 */
export const POST: RequestHandler = async (event) => {
	const cuerpo = (await event.request.json().catch(() => null)) as Record<string, unknown> | null;
	if (!cuerpo || typeof cuerpo !== 'object') error(400, 'Pedido inválido.');
	const fuente = partnersDe(event);
	try {
		switch (cuerpo.que) {
			case 'perfil': {
				const r = validarPerfil(cuerpo.perfil as never);
				if (!r.ok) return json({ message: r.mensaje, campo: r.campo }, { status: 400 });
				return json(await fuente.guardarPerfil(r.valor));
			}
			case 'cobro': {
				const r = validarCobro(cuerpo.cobro as never);
				if (!r.ok) return json({ message: r.mensaje, campo: r.campo }, { status: 400 });
				return json(await fuente.guardarCobro(r.valor));
			}
			case 'factura': {
				const mes = String(cuerpo.mes ?? '');
				const r = validarFactura(cuerpo.url);
				if (!/^\d{4}-\d{2}$/.test(mes)) return json({ message: 'Elegí el mes.' }, { status: 400 });
				if (!r.ok) return json({ message: r.mensaje, campo: r.campo }, { status: 400 });
				try {
					await fuente.subirFactura(mes, r.valor);
				} catch (e) {
					if (e instanceof Error && e.message === 'sin-pago')
						return json({ message: 'Ese pago ya está cerrado.' }, { status: 404 });
					throw e;
				}
				return json({ ok: true });
			}
		}
	} catch (e) {
		if (e instanceof ErrorApi)
			return json(
				{ message: e.message, campo: Object.keys(e.campos)[0] },
				{ status: e.estado >= 500 ? 503 : e.estado }
			);
		throw e;
	}
	error(400, 'Pedido inválido.');
};
