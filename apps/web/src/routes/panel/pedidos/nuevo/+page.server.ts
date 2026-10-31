import { datosDe } from '#lib/server/datos/index.ts';
import type { VarianteFila } from '#lib/panel/vendibles.ts';
import type { PageServerLoad } from './$types';

/**
 * Crear pedido: productos con stock y clientes para elegir. `?cliente=` llega precargado y
 * `?desde=1048` trae ese pedido para «Volver a vender» (se completa en la página, para revisar).
 */
export const load: PageServerLoad = async ({ locals, url }) => {
	const datos = datosDe(locals);
	const desde = Number(url.searchParams.get('desde'));
	const [productos, clientes, entregas, pedido] = await Promise.all([
		datos.listar('productos', { cuantos: 200 }),
		datos.listar('clientes', { cuantos: 200 }),
		// Cómo entrega la tienda de verdad (Envíos y retiro): sin esto el pedido ofrecía Andreani y
		// OCA con costos inventados.
		datos.entregas(),
		Number.isSafeInteger(desde) && desde > 0 ? datos.pedido(desde).catch(() => null) : null
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
		entregas: { moto: entregas.moto, motoCosto: entregas.motoCosto, correo: entregas.correo },
		desde: pedido && {
			numero: pedido.numero,
			canal: pedido.canal,
			envio: pedido.envio,
			comprador: { nombre: pedido.comprador.nombre, telefono: pedido.comprador.telefono },
			items: pedido.items.map(({ productoId, varianteId, nombre, variante, cantidad, precio }) => ({
				productoId,
				varianteId,
				nombre,
				variante,
				cantidad,
				precio
			}))
		}
	};
};
