import { datosDe } from '#lib/server/datos/index.ts';
import type { VarianteFila } from '#lib/panel/vendibles.ts';
import type { PageServerLoad } from './$types';

/** El punto de venta: los productos (con su stock) y la caja de hoy. */
export const load: PageServerLoad = async ({ locals, depends }) => {
	// Tiempo real (§44): se recarga con los avisos de lo que muestra.
	depends('tienda:pos');
	const datos = datosDe(locals);
	const [productos, hoy] = await Promise.all([
		datos.listar('productos', { cuantos: 200 }),
		datos.tablero('hoy')
	]);
	return {
		productos: productos.filas as {
			id: string;
			nombre: string;
			foto: string | null;
			sku: string;
			categoria: string;
			precio: number;
			stock: number;
			variantes: VarianteFila[];
			publicado: boolean;
		}[],
		caja: hoy.caja
	};
};
