import { fail } from '@sveltejs/kit';
import {
	cambiarClave,
	firmasYVacaciones,
	guardarFiltros,
	guardarFirma,
	guardarVacaciones,
	leerFiltros,
	validar,
	type Filtros
} from '#lib/server/ajustes.ts';
import { Jmap } from '#lib/server/jmap.ts';
import { marca } from '#lib/server/marca.ts';
import { guardar } from '#lib/server/sesion.ts';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals, cookies }) => {
	const j = new Jmap(locals.cred!);
	const [fv, filtros, cuentas] = await Promise.all([
		firmasYVacaciones(j),
		leerFiltros(j),
		j.cuentas()
	]);
	return {
		...fv,
		filtros,
		casillas: cuentas.map((c) => ({ cuenta: c.id, email: c.email, personal: c.personal })),
		yo: locals.cred!.email,
		marca: marca(),
		tema: cookies.get('correo_tema') === 'claro' ? ('claro' as const) : ('oscuro' as const)
	};
};

const texto = (f: FormData, k: string) => String(f.get(k) ?? '');
const fecha = (s: string) =>
	s ? new Date(`${s}T00:00:00-03:00`).toISOString().replace(/\.\d+Z$/, 'Z') : null;

export const actions: Actions = {
	firma: async ({ locals, request }) => {
		const f = await request.formData();
		try {
			await guardarFirma(new Jmap(locals.cred!), {
				cuenta: texto(f, 'cuenta'),
				identidad: texto(f, 'identidad'),
				nombre: texto(f, 'nombre'),
				texto: texto(f, 'texto')
			});
		} catch (e) {
			return fail(400, { error: e instanceof Error ? e.message : 'No se pudo.' });
		}
		return { ok: 'Firma guardada' };
	},
	vacaciones: async ({ locals, request }) => {
		const f = await request.formData();
		try {
			await guardarVacaciones(new Jmap(locals.cred!), {
				cuenta: texto(f, 'cuenta'),
				activa: f.get('activa') === 'on',
				desde: fecha(texto(f, 'desde')),
				hasta: fecha(texto(f, 'hasta')),
				asunto: texto(f, 'asunto'),
				texto: texto(f, 'texto')
			});
		} catch (e) {
			return fail(400, { error: e instanceof Error ? e.message : 'No se pudo.' });
		}
		return { ok: 'Respuesta automática guardada' };
	},
	filtros: async ({ locals, request }) => {
		const f = await request.formData();
		try {
			const filtros = validar(JSON.parse(texto(f, 'filtros')) as Filtros);
			await guardarFiltros(new Jmap(locals.cred!), filtros);
		} catch (e) {
			return fail(400, { error: e instanceof Error ? e.message : 'No se pudo.' });
		}
		return { ok: 'Reglas guardadas' };
	},
	clave: async ({ locals, request, cookies, url }) => {
		const f = await request.formData();
		const nueva = texto(f, 'nueva');
		if (nueva !== texto(f, 'repetir'))
			return fail(400, { error: 'Las dos claves nuevas no coinciden.' });
		try {
			await cambiarClave(locals.cred!.email, texto(f, 'actual'), nueva);
		} catch (e) {
			return fail(400, { error: e instanceof Error ? e.message : 'No se pudo.' });
		}
		// La sesión guarda la clave: se actualiza para no tener que volver a entrar.
		guardar(cookies, { email: locals.cred!.email, clave: nueva }, url.protocol === 'https:');
		return { ok: 'Clave cambiada' };
	}
};
