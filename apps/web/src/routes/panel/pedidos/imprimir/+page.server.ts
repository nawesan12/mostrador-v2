import { datosDe } from '#lib/server/datos/index.ts';
import { contextoTarjeta } from '#lib/server/tarjeta.ts';
import type { PageServerLoad } from './$types';

/**
 * Imprimir tarjetas A6: `?n=1047,1046` (hasta 40 pedidos) y `?f=a6|a4` (una por hoja, o 4 por A4
 * con marcas de corte). `?ejemplo=1` imprime la del pedido de ejemplo.
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
	const [ctx, pedidos] = await Promise.all([
		contextoTarjeta(datos, url.origin),
		Promise.all(numeros.map((n) => datos.pedido(n)))
	]);
	const formato = url.searchParams.get('f');
	return {
		...ctx,
		ejemplo: url.searchParams.get('ejemplo') === '1',
		formato: (formato === 'a6' || formato === 'a4' ? formato : numeros.length > 1 ? 'a4' : 'a6') as
			'a6' | 'a4',
		pedidos: pedidos
			.filter((p) => p !== null)
			.map((p) => ({
				numero: p.numero,
				nombre: p.comprador.nombre || p.cliente,
				creado: p.creado,
				productos: p.items.map((i) => i.nombre),
				foto: p.items.find((i) => i.foto)?.foto ?? null
			}))
	};
};
