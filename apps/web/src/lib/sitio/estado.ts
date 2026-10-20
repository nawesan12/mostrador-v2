/** "99,96 %" con espacio duro (dos decimales; el 100 va solo). */
export const porcentaje = (p: number | null) =>
	p === null
		? '—'
		: p >= 99.995
			? '100\u00a0%'
			: `${p.toLocaleString('es-AR', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}\u00a0%`;
