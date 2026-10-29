/** Los planes de Tu Local Web (PanelCuenta.dc.html): los mismos precios que la landing. */
export const PLANES = [
	{
		id: 'gratis',
		nombre: 'Gratis',
		lema: 'Para probar y empezar',
		precio: 0,
		incluye: ['Productos ilimitados', 'Hornero: 100 respuestas con IA', '1 persona · subdominio'],
		personas: 1,
		hornero: 100
	},
	{
		id: 'emprende',
		nombre: 'Emprende',
		lema: 'Para vender todos los días',
		precio: 19_900,
		incluye: [
			'Dominio propio y factura ARCA',
			'Todos los medios de pago y envío',
			'Hornero: 1.000 respuestas con IA',
			'2 personas'
		],
		personas: 2,
		hornero: 1000
	},
	{
		id: 'pro',
		nombre: 'Pro',
		lema: 'El más elegido',
		precio: 49_900,
		incluye: [
			'Mercado Libre, Instagram y local',
			'Acciones masivas y estadísticas',
			'Hornero: 3.000 respuestas con IA',
			'5 personas con permisos'
		],
		personas: 5,
		hornero: 3000
	},
	{
		id: 'negocio',
		nombre: 'Negocio',
		lema: 'Para varias sucursales',
		precio: 149_900,
		incluye: [
			'Sucursales y cajas ilimitadas',
			'API, webhooks y mayoristas',
			'Hornero: 15.000 respuestas con IA',
			'Personas ilimitadas'
		],
		personas: null,
		hornero: 15_000
	}
] as const;
export type IdPlan = (typeof PLANES)[number]['id'];
/** Pagando por año, 25 % menos. */
export const precioAnual = (mensual: number) => Math.round(mensual * 0.75);
