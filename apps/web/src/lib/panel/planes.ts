/**
 * Los planes de Tu Local Web (acordados el 19/11/2026): precio, lo que incluye y los topes. Es el
 * espejo de la API (apps/api/internal/dominio/planes, con testdata/planes.json que compara los dos):
 * los topes se aplican en el servidor; acá sólo se muestran. Precios en pesos por mes, precio final,
 * sin comisión por venta. El anual se paga como 10 meses y se usa 12. El id `pro` es Crece.
 */
export const PLANES = [
	{
		id: 'gratis',
		nombre: 'Gratis',
		lema: 'Para probar y empezar',
		precio: 0,
		productos: 20,
		personas: 1,
		fotos: 3,
		fotosMB: 15,
		tincho: 50,
		funciones: [],
		incluye: [
			'Hasta 20 productos',
			'Tu dirección de Tu Local Web',
			'Cobrás con Mercado Pago y transferencia',
			'Tincho: 50 respuestas con IA por mes',
			'1 persona'
		]
	},
	{
		id: 'emprende',
		nombre: 'Emprende',
		lema: 'Para vender todos los días',
		precio: 14_900,
		productos: 150,
		personas: 2,
		fotos: 5,
		fotosMB: 150,
		tincho: 500,
		funciones: ['dominio', 'arca', 'correo', 'cupones'],
		incluye: [
			'Hasta 150 productos',
			'Dominio propio y factura de ARCA',
			'Envíos con Correo Argentino',
			'Cupones, combos y carritos abandonados',
			'Tincho: 500 respuestas con IA por mes',
			'2 personas'
		]
	},
	{
		id: 'pro',
		nombre: 'Crece',
		lema: 'Para vender en todos tus canales',
		precio: 34_900,
		productos: 500,
		personas: 5,
		fotos: 8,
		fotosMB: 600,
		tincho: 2_000,
		funciones: [
			'dominio',
			'arca',
			'correo',
			'cupones',
			'roles',
			'local',
			'mercadolibre',
			'newsletter'
		],
		incluye: [
			'Hasta 500 productos',
			'Venta en el local y cobro con QR',
			'Mercado Libre con el mismo stock',
			'Newsletter y estadísticas',
			'Tincho: 2.000 respuestas con IA por mes',
			'5 personas con roles y permisos'
		]
	},
	{
		id: 'negocio',
		nombre: 'Negocio',
		lema: 'Para vender sin topes',
		precio: 79_900,
		productos: 0,
		personas: 0,
		fotos: 10,
		fotosMB: 2_048,
		tincho: 8_000,
		funciones: [
			'dominio',
			'arca',
			'correo',
			'cupones',
			'roles',
			'local',
			'mercadolibre',
			'newsletter',
			'mayorista'
		],
		incluye: [
			'Productos y personas sin tope',
			'Venta mayorista y listas de precio',
			'Todo lo de Crece',
			'Tincho: 8.000 respuestas con IA por mes'
		]
	}
] as const;
export type IdPlan = (typeof PLANES)[number]['id'];
export type Plan = (typeof PLANES)[number];
export type Funcion =
	| 'dominio'
	| 'arca'
	| 'correo'
	| 'cupones'
	| 'roles'
	| 'local'
	| 'mercadolibre'
	| 'newsletter'
	| 'mayorista';

/** El plan destacado: el del medio, el que se usa durante la prueba gratis. */
export const DESTACADO: IdPlan = 'pro';

/** Lo que se ve en el pie de cada plan: sin comisión por venta, en todos. */
export const SIN_COMISION = '0 % de comisión por venta';

/** Pagando por año: 10 meses por 12 (dos de regalo). Devuelve lo que sale por mes, redondeado. */
export const MESES_PAGOS_DEL_ANIO = 10;
export const precioAnual = (mensual: number) => Math.round((mensual * MESES_PAGOS_DEL_ANIO) / 12);
/** Lo que se cobra una vez por año. */
export const totalAnual = (mensual: number) => mensual * MESES_PAGOS_DEL_ANIO;

/** El pack extra de Tincho: +2.000 respuestas con IA este mes, pago con Mercado Pago. */
export const PACK_TINCHO = { respuestas: 2_000, precio: 4_900 } as const;

export const planDe = (id: string): Plan => PLANES.find((p) => p.id === id) ?? PLANES[0];

/** El plan que rige hoy: durante la prueba gratis, una tienda en Gratis usa Crece. */
export const planVigente = (actual: string, enPrueba: boolean) =>
	actual === 'gratis' && enPrueba ? DESTACADO : actual;

/** Días de la prueba gratis, desde el alta de la tienda (abono.DiasDePrueba en la API). */
export const DIAS_DE_PRUEBA = 14;

/** El plan que rige para una tienda (su plan y su alta): en la prueba, Crece. */
export const planDeTienda = (t: { plan: string; creadaEn: string }, ahora = Date.now()) =>
	planDe(planVigente(t.plan, ahora < new Date(t.creadaEn).getTime() + DIAS_DE_PRUEBA * 86_400_000));

/** El plan más barato que incluye la función. */
export const planCon = (f: Funcion): Plan =>
	PLANES.find((p) => (p.funciones as readonly string[]).includes(f)) ?? PLANES[PLANES.length - 1];

/** «hasta 150» o «sin tope» (0 es sin tope, con uso razonable). */
export const tope = (n: number) => (n === 0 ? 'sin tope' : `hasta ${n.toLocaleString('es-AR')}`);

/**
 * El descuento para quedarse (§57, abono.Ofertas de la API) por motivo de baja: «es caro» → 50 % por 3
 * cobros; «otro» → 50 % en el próximo. Se cobra de verdad (baja el débito) y va una vez por tienda.
 */
export type Beneficio = { pct: number; cobros: number };
export const BENEFICIOS: Record<number, Beneficio> = {
	0: { pct: 50, cobros: 3 },
	4: { pct: 50, cobros: 1 }
};
export const conDescuento = (importe: number, b: Beneficio | null) =>
	b && b.cobros > 0 && b.pct > 0 ? Math.round((importe * (100 - b.pct)) / 100) : importe;
