import { redirect } from '@sveltejs/kit';
import { borrarMirada } from '#lib/server/staff/index.ts';
import type { RequestHandler } from './$types';

/** Termina de ver una tienda en modo lectura y vuelve al backoffice. */
export const POST: RequestHandler = async ({ cookies }) => {
	borrarMirada(cookies);
	redirect(303, '/staff/tiendas');
};
