import { datosDe, usaSandbox } from '#lib/server/datos/index.ts';
import { plantilla } from '#lib/plantillas.ts';
import type { LayoutServerLoad } from './$types';

/**
 * El armazón del panel: la tienda, el resumen (insignias, notificaciones, prueba) y las preferencias
 * de la persona (tema y modo, en cookies que escribe el navegador, así el primer render ya sale bien).
 */
export const load: LayoutServerLoad = async ({ locals, cookies }) => {
	const datos = datosDe(locals);
	const [tienda, resumen, ajustes] = await Promise.all([
		datos.tienda(),
		datos.resumen(),
		// Si la API todavía no responde turnos, la sección queda como la del rubro.
		datos.ajustesTurnos().catch(() => null)
	]);
	return {
		tienda,
		resumen,
		// La foto de la tienda: la portada de su plantilla.
		fotoTienda: plantilla(tienda.plantilla)?.img ?? null,
		demo: usaSandbox(),
		// Alguien del equipo de Tu Local Web viendo la tienda en sólo lectura (aviso fijo).
		mirada: locals.mirada,
		// La sección Turnos: con turnos activos o si la tienda es de servicios.
		turnos: !!ajustes?.activo || tienda.rubro === 'servicios',
		tema: cookies.get('panel_tema') === 'dark' ? ('dark' as const) : ('light' as const),
		modo: cookies.get('panel_modo') === 'simple' ? ('simple' as const) : ('completo' as const)
	};
};
