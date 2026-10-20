import { datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import { plantilla } from '#lib/plantillas.ts';
import type { LayoutServerLoad } from './$types';

/**
 * El armazón del panel: la tienda, el resumen (insignias, notificaciones, prueba) y las preferencias
 * de la persona (tema y modo, en cookies que escribe el navegador, así el primer render ya sale bien).
 */
export const load: LayoutServerLoad = async ({ locals, cookies }) => {
	const datos = datosDe(locals);
	const [tienda, resumen] = await Promise.all([datos.tienda(), datos.resumen()]);
	return {
		tienda,
		resumen,
		// La foto de la tienda: la portada de su plantilla.
		fotoTienda: plantilla(tienda.plantilla)?.img ?? null,
		demo: usaSandbox(),
		tema: cookies.get('panel_tema') === 'dark' ? ('dark' as const) : ('light' as const),
		modo: cookies.get('panel_modo') === 'simple' ? ('simple' as const) : ('completo' as const)
	};
};
