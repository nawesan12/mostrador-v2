/** Los avisos y por dónde llegan (design/PanelCuenta.dc.html, Notificaciones). */
export const CANALES_AVISO = ['panel', 'whatsapp', 'email', 'celular'] as const;
export type CanalAviso = (typeof CANALES_AVISO)[number];

type Fila = { id: string; t: string; d: string; def: readonly number[] };
export const AVISOS: readonly { grupo: string; filas: readonly Fila[] }[] = [
	{
		grupo: 'Ventas',
		filas: [
			{ id: 'venta', t: 'Venta nueva', d: 'Cada pedido pagado', def: [1, 1, 0, 1] },
			{
				id: 'pago',
				t: 'Pago pendiente o rechazado',
				d: 'Transferencias y tarjetas',
				def: [1, 1, 1, 0]
			},
			{
				id: 'carrito',
				t: 'Carrito abandonado recuperado',
				d: 'Cuando Hornero cierra la venta',
				def: [1, 0, 0, 1]
			}
		]
	},
	{
		grupo: 'Operación',
		filas: [
			{ id: 'stock', t: 'Stock bajo', d: 'Menos de 5 unidades', def: [1, 0, 1, 1] },
			{ id: 'demora', t: 'Envío demorado', d: 'Más de 48 h sin movimiento', def: [1, 1, 0, 1] },
			{ id: 'arca', t: 'Certificado de ARCA por vencer', d: '15 días antes', def: [1, 1, 1, 0] }
		]
	},
	{
		grupo: 'Cuenta y seguridad',
		filas: [
			{ id: 'sesion', t: 'Inicio de sesión nuevo', d: 'Desde otro dispositivo', def: [1, 1, 1, 0] },
			{ id: 'cobro', t: 'Cobro del plan', d: 'Factura de Mostrador', def: [0, 0, 1, 0] },
			{ id: 'novedades', t: 'Novedades de Mostrador', d: 'Una vez por mes', def: [0, 0, 1, 0] }
		]
	}
];

/** Si un aviso llega por un canal, con lo que eligió la persona o lo de entrada. */
export function activo(matriz: Record<string, boolean>, aviso: string, canal: CanalAviso) {
	const k = `${aviso}:${canal}`;
	if (k in matriz) return matriz[k];
	const fila = AVISOS.flatMap((g) => g.filas).find((f) => f.id === aviso);
	return !!fila?.def[CANALES_AVISO.indexOf(canal)];
}
