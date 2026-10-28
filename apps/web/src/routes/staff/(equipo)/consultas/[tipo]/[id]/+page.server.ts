import { error, fail } from '@sveltejs/kit';
import { ErrorApi } from '#lib/server/bff.ts';
import { EnLaDemo } from '#lib/server/staff/demo.ts';
import { equipoDe, exigir } from '#lib/server/staff/index.ts';
import type { EstadoConsulta, TipoConsulta } from '#lib/server/staff/tipos.ts';
import type { Actions, PageServerLoad } from './$types';

const tipoDe = (t: string): TipoConsulta => {
	if (t !== 'contacto' && t !== 'arrepentimiento') error(404, 'Esa consulta no existe.');
	return t;
};

/** Una consulta con sus respuestas (Go deja en la auditoría que alguien la abrió). */
export const load: PageServerLoad = async (event) => {
	const { sesion } = await event.parent();
	exigir(sesion.permisos, 'consultas.ver');
	const datos = equipoDe(event);
	const [consulta, bandeja] = await Promise.all([
		datos.consulta(tipoDe(event.params.tipo), event.params.id),
		datos.consultas()
	]);
	if (!consulta) error(404, 'Esa consulta no existe.');
	return { consulta, miembros: bandeja.miembros };
};

async function intentar(f: () => Promise<unknown>, ok: Record<string, unknown>) {
	try {
		await f();
		return ok;
	} catch (e) {
		if (e instanceof EnLaDemo) return fail(409, { demo: true });
		if (e instanceof ErrorApi) return fail(e.estado, { error: e.message });
		throw e;
	}
}

export const actions: Actions = {
	estado: async (event) => {
		const f = await event.request.formData();
		const estado = String(f.get('estado') ?? '') as EstadoConsulta;
		const asignado = String(f.get('asignadoA') ?? '');
		return intentar(
			() =>
				equipoDe(event).actualizarConsulta(tipoDe(event.params.tipo), event.params.id, {
					...(estado ? { estado } : {}),
					...(f.has('asignadoA') ? { asignadoA: asignado || null } : {})
				}),
			{ cambio: true }
		);
	},
	responder: async (event) => {
		const f = await event.request.formData();
		const asunto = String(f.get('asunto') ?? '').trim();
		const texto = String(f.get('texto') ?? '').trim();
		if (asunto.length < 3 || texto.length < 10)
			return fail(400, { error: 'Escribí el asunto y la respuesta.' });
		return intentar(
			() =>
				equipoDe(event).responder(tipoDe(event.params.tipo), event.params.id, { asunto, texto }),
			{ respondida: true }
		);
	}
};
