import { error } from '@sveltejs/kit';
import { usaSandbox } from '#lib/server/datos/index.ts';
import { COOKIE_ROL_DEMO } from '#lib/server/staff/index.ts';
import { ROLES } from '#lib/staff/permisos.ts';
import type { RequestHandler } from './$types';

/** Demo: elige con qué rol se ve el backoffice de ejemplo. Con la API, el rol es el de la sesión. */
export const POST: RequestHandler = async ({ request, cookies }) => {
	if (!usaSandbox()) error(404, 'No existe');
	const { rol } = (await request.json().catch(() => ({}))) as { rol?: string };
	if (!ROLES.some((r) => r.id === rol)) error(400, 'Rol desconocido');
	cookies.set(COOKIE_ROL_DEMO, String(rol), {
		path: '/staff',
		httpOnly: true,
		sameSite: 'lax',
		maxAge: 60 * 60 * 24 * 30
	});
	return new Response(null, { status: 204 });
};
