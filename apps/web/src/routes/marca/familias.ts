import { resolve } from '$app/paths';

/** Las cuatro familias de /marca (design/README.md, "7. Marca"). */
export const FAMILIAS = [
	{
		ruta: '/marca/papeleria',
		href: resolve('/marca/papeleria'),
		titulo: 'Papelería',
		texto:
			'Tarjetas, hoja membretada, sobre, firmas de email, factura y remito modelo, carpeta, credenciales y stickers.',
		hornero: 'portapapeles',
		piezas: 17
	},
	{
		ruta: '/marca/redes',
		href: resolve('/marca/redes'),
		titulo: 'Redes',
		texto:
			'Avatares, tapas de destacadas, portadas de LinkedIn, X y YouTube, anuncios para Meta y miniaturas.',
		hornero: 'foto',
		piezas: 24
	},
	{
		ruta: '/marca/imprimibles',
		href: resolve('/marca/imprimibles'),
		titulo: 'Imprimibles',
		texto:
			'Flyers A5, afiche A3, tríptico, one-pager, roll-up, backing de stand y la tarjeta para paquetes.',
		hornero: 'envio',
		piezas: 12
	},
	{
		ruta: '/marca/og',
		href: resolve('/marca/og'),
		titulo: 'Imágenes para compartir',
		texto:
			'Las 9 imágenes de 1200 × 630 que muestran WhatsApp, redes y buscadores al compartir el sitio.',
		hornero: 'whatsapp',
		piezas: 9
	}
] as const;
