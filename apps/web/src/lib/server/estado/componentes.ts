/** Los componentes de /estado (design/Mostrador Estado.dc.html), en el orden del diseño. */
export const COMPONENTES = [
	['tiendas', 'Tiendas online'],
	['panel', 'Panel de administración'],
	['pagos', 'Checkout y pagos'],
	['whatsapp', 'WhatsApp y Tincho (IA)'],
	['arca', 'Facturación electrónica (ARCA)'],
	['mercadolibre', 'Sincronización con Mercado Libre'],
	['api', 'API y webhooks']
] as const;

export type Componente = (typeof COMPONENTES)[number][0];
export const nombreDe = (c: string) => COMPONENTES.find(([k]) => k === c)?.[1] ?? c;

/** Día de la Argentina (YYYY-MM-DD) de un instante. */
export const diaAR = (ms: number) =>
	new Intl.DateTimeFormat('en-CA', { timeZone: 'America/Argentina/Buenos_Aires' }).format(ms);
