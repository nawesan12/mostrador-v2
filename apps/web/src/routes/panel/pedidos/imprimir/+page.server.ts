import { datosDe } from '#lib/server/datos/index.ts';
import { contextoTarjeta } from '#lib/server/tarjeta.ts';
import type { PageServerLoad } from './$types';

/**
 * Imprimir tarjetas A6 o etiquetas de envío (`?que=etiquetas`): `?n=1047,1046` (hasta 40 pedidos)
 * y `?f=a6|a4` (una por hoja, o 4 por A4 con marcas de corte). `?ejemplo=1` imprime la del pedido
 * de ejemplo.
 */
export const load: PageServerLoad = async ({ locals, url, setHeaders }) => {
	setHeaders({ 'cache-control': 'private, no-store' });
	const datos = datosDe(locals);
	const numeros = [
		...new Set(
			(url.searchParams.get('n') ?? '')
				.split(',')
				.map(Number)
				.filter((n) => Number.isInteger(n) && n > 0)
		)
	].slice(0, 40);
	const [ctx, pedidos, conf] = await Promise.all([
		contextoTarjeta(datos, url.origin),
		Promise.all(numeros.map((n) => datos.pedido(n))),
		datos.configuracion()
	]);
	const formato = url.searchParams.get('f');
	return {
		...ctx,
		ejemplo: url.searchParams.get('ejemplo') === '1',
		que:
			url.searchParams.get('que') === 'etiquetas' ? ('etiquetas' as const) : ('tarjetas' as const),
		remitente: {
			nombre: conf.negocio.nombre || ctx.tienda.nombre,
			direccion: conf.negocio.direccion,
			whatsapp: conf.negocio.whatsapp
		},
		formato: (formato === 'a6' || formato === 'a4' ? formato : numeros.length > 1 ? 'a4' : 'a6') as
			'a6' | 'a4',
		pedidos: pedidos
			.filter((p) => p !== null)
			.map((p) => ({
				numero: p.numero,
				nombre: p.comprador.nombre || p.cliente,
				creado: p.creado,
				productos: p.items.map((i) => i.nombre),
				foto: p.items.find((i) => i.foto)?.foto ?? null,
				telefono: p.comprador.telefono,
				envio: p.envio,
				ciudad: p.ciudad,
				entrega: p.entrega,
				unidades: p.items.reduce((s, i) => s + i.cantidad, 0)
			}))
	};
};
