import type { RequestHandler } from './$types';

/** Recuerda el tema elegido (el oscuro es el de siempre). */
export const POST: RequestHandler = async ({ cookies, request }) => {
	const { tema } = (await request.json()) as { tema?: string };
	cookies.set('correo_tema', tema === 'claro' ? 'claro' : 'oscuro', {
		path: '/',
		httpOnly: true,
		sameSite: 'lax',
		maxAge: 365 * 86_400
	});
	return new Response(null, { status: 204 });
};
