import { env } from 'cloudflare:workers';
import { error, redirect } from '@sveltejs/kit';
import { iaDe } from '#lib/panel/integraciones.ts';
import { usaSandbox } from '#lib/server/datos/index.ts';
import {
	cliente,
	emitirCodigo,
	NO_DISPONIBLE,
	permisosPedidos,
	type Permiso
} from '#lib/server/mcp/oauth.ts';
import { RUBROS } from '#lib/server/sandbox/rubros.ts';
import { dbSandbox } from '#lib/server/sandbox/sandbox.ts';
import { plantilla } from '#lib/plantillas.ts';
import type { Actions, PageServerLoad } from './$types';

/**
 * Consentimiento del OAuth (como /oauth/autorizar de Caja): qué IA pide acceso, qué va a poder hacer,
 * y Permitir / Cancelar. En la demo, el acceso es al sandbox del visitante; si no tiene uno, elige
 * una tienda en /demo y vuelve acá.
 */
async function validar(url: URL) {
	if (!usaSandbox()) error(501, NO_DISPONIBLE);
	const p = {
		clienteId: url.searchParams.get('client_id') ?? '',
		redireccion: url.searchParams.get('redirect_uri') ?? '',
		desafio: url.searchParams.get('code_challenge') ?? '',
		metodo: url.searchParams.get('code_challenge_method') ?? '',
		state: url.searchParams.get('state'),
		scope: url.searchParams.get('scope')
	};
	if (url.searchParams.get('response_type') !== 'code')
		error(400, 'Pedido de autorización inválido.');
	const c = await cliente(dbSandbox(env.SANDBOX), p.clienteId);
	if (!c) error(400, 'Esa IA no está registrada. Volvé a conectar desde la IA.');
	if (!c.redirecciones.includes(p.redireccion)) error(400, 'La dirección de vuelta no coincide.');
	if (p.metodo !== 'S256' || !/^[A-Za-z0-9_-]{43,128}$/.test(p.desafio))
		error(400, 'La IA tiene que usar PKCE (S256).');
	return { p, c };
}

const vuelta = (redireccion: string, params: Record<string, string | null>) => {
	const u = new URL(redireccion);
	for (const [k, v] of Object.entries(params)) if (v) u.searchParams.set(k, v);
	return u.toString();
};

export const load: PageServerLoad = async ({ url, locals }) => {
	const { p, c } = await validar(url);
	if (!locals.sandbox)
		redirect(303, `/demo?volver=${encodeURIComponent(url.pathname + url.search)}`);
	const sb = locals.sandbox;
	const t = plantilla(sb.plantilla ?? RUBROS.find((r) => r.id === sb.rubro)?.plantilla ?? '');
	return {
		ia: { nombre: c.nombre, id: iaDe(c.nombre).id, host: new URL(p.redireccion).host },
		acciones: permisosPedidos(p.scope).includes('acciones'),
		tienda: t?.tienda ?? null
	};
};

export const actions: Actions = {
	default: async ({ request, url, locals }) => {
		const { p, c } = await validar(url);
		if (!locals.sandbox)
			redirect(303, `/demo?volver=${encodeURIComponent(url.pathname + url.search)}`);
		const f = await request.formData();
		if (f.get('decision') !== 'permitir')
			redirect(303, vuelta(p.redireccion, { error: 'access_denied', state: p.state }), {
				external: true
			});
		// Nunca más de lo que pidió la IA.
		const permisos: Permiso[] =
			f.get('acciones') === 'on' && permisosPedidos(p.scope).includes('acciones')
				? ['lectura', 'acciones']
				: ['lectura'];
		const code = await emitirCodigo(dbSandbox(env.SANDBOX), {
			clienteId: c.id,
			sandboxId: locals.sandbox.id,
			redireccion: p.redireccion,
			desafio: p.desafio,
			permisos
		});
		// La vuelta es una dirección registrada por la IA (https o localhost): siempre externa.
		redirect(303, vuelta(p.redireccion, { code, state: p.state, iss: url.origin }), {
			external: true
		});
	}
};
