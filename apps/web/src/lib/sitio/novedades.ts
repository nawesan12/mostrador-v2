/** Novedades del producto (design/Mostrador Novedades.dc.html), por mes, de la más nueva a la más vieja. */
export type Tipo = 'nuevo' | 'mejora' | 'arreglo';
export type Novedad = { tipo: Tipo; titulo: string; texto: string; foto?: string };

export const TIPOS = [
	['nuevo', 'Nuevo', '#FCEDE5', '#C24E25'],
	['mejora', 'Mejora', '#E8F1FF', '#3E6BFD'],
	['arreglo', 'Arreglo', '#E6F4EC', '#2F8F63']
] as const;

export const NOVEDADES: [string, Novedad[]][] = [
	[
		'Octubre 2026',
		[
			{
				tipo: 'mejora',
				titulo: 'Pedidos 3 veces más rápidos',
				texto:
					'La lista de pedidos ahora carga miles de pedidos al instante, con filtros por canal, pago y envío.'
			},
			{
				tipo: 'arreglo',
				titulo: 'Factura A con CUIT extranjero',
				texto: 'Ya se pueden emitir facturas a clientes del exterior sin error de validación.'
			}
		]
	],
	[
		'Septiembre 2026',
		[
			{
				tipo: 'nuevo',
				titulo: 'Tincho responde en Instagram',
				foto: '1611162617213-7d7a39e9b1d7',
				texto:
					'El asistente ahora contesta mensajes directos de Instagram con el stock y los precios reales de tu tienda.'
			},
			{
				tipo: 'mejora',
				titulo: 'Plantillas con diseño propio',
				texto: 'Sumamos 10 plantillas completamente nuevas, cada una pensada para un rubro.'
			}
		]
	],
	[
		'Agosto 2026',
		[
			{
				tipo: 'nuevo',
				titulo: 'Comparador de productos',
				texto:
					'Las tiendas de tecnología pueden mostrar hasta tres productos lado a lado con sus especificaciones.'
			},
			{
				tipo: 'mejora',
				titulo: 'Onboarding en 10 minutos',
				texto:
					'Ahora importamos productos desde Instagram, Mercado Libre o un Excel mientras creás la cuenta.'
			}
		]
	]
];
