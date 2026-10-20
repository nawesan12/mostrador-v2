import { redirect } from '@sveltejs/kit';

// "Legales" en la miga lleva a la primera, como en el handoff.
export const load = () => redirect(308, '/legales/terminos');
