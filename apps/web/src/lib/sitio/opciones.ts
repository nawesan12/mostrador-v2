/** Opciones de los formularios del sitio (los chips del handoff). Las valida también Go. */
export const FACTURACION = [
	['empiezo', 'Recién empiezo'],
	['hasta-5', 'Hasta $ 5 M'],
	['5-a-20', '$ 5 a 20 M'],
	['20-a-100', '$ 20 a 100 M'],
	['mas-de-100', 'Más de $ 100 M']
] as const;
export const ORIGEN = [
	['tiendanube', 'Tiendanube'],
	['empretienda', 'Empretienda'],
	['mercadolibre', 'Mercado Libre'],
	['shopify', 'Shopify'],
	['woocommerce', 'WooCommerce'],
	['ninguna', 'Ninguna']
] as const;
export const ESPECIALIDADES = [
	['diseno', 'Diseño'],
	['desarrollo', 'Desarrollo'],
	['marketing', 'Marketing'],
	['contabilidad', 'Contabilidad'],
	['fotografia', 'Fotografía']
] as const;
