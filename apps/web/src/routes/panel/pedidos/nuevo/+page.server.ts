import { datosDe } from '#lib/server/datos/index.ts';
import type { VarianteFila } from '#lib/panel/vendibles.ts';
import type { PageServerLoad } from './$types';

/** Crear pedido: productos con stock y clientes para elegir. `?cliente=` llega precargado. */
export const load: PageServerLoad = async ({ locals, url }) => {
	const datos = datosDe(locals);
	const [productos, clientes, entregas] = await Promise.all([
		datos.listar('productos', { cuantos: 200 }),
		datos.listar('clientes', { cuantos: 200 }),
		// Cómo entrega la tienda de verdad (Envíos y retiro): sin esto el pedido ofrecía Andreani y
		// OCA con costos inventados.
		datos.entregas()
	]);
	return {
		productos: productos.filas as {
			id: string;
			nombre: string;
			foto: string | null;
			sku: string;
			precio: number;
			stock: number;
			variantes: VarianteFila[];
		}[],
		clientes: (
			clientes.filas as {
				id: string;
				nombre: string;
				foto: string | null;
				ciudad: string;
				telefono: string | null;
			}[]
		).map(({ id, nombre, foto, ciudad, telefono }) => ({ id, nombre, foto, ciudad, telefono })),
		cliente: url.searchParams.get('cliente'),
		entregas: { moto: entregas.moto, motoCosto: entregas.motoCosto, correo: entregas.correo }
	};
};
