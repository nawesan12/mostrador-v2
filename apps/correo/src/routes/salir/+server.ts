import { redirect } from '@sveltejs/kit';
import { cerrar } from '#lib/server/sesion.ts';
import type { RequestHandler } from './$types';

export const POST: RequestHandler = ({ cookies }) => {
	cerrar(cookies);
	redirect(303, '/ingresar');
};
