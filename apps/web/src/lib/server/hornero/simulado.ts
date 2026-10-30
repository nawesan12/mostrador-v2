/**
 * Tincho sin modelo (desarrollo y pruebas, HORNERO=simulado): respuestas armadas con los mismos datos,
 * para no llamar a Workers AI desde las pruebas.
 */
import type { DatosHornero } from './contexto.ts';

const pesos = (n: number) => '$ ' + Math.round(n).toLocaleString('es-AR');

export function responderSimulado(pregunta: string, d: DatosHornero) {
	const q = pregunta.toLowerCase();
	const dias = d.contexto.ventasPorDia;
	if (/ayer/.test(q) && dias.length >= 2) {
		const ayer = dias[dias.length - 2];
		return `Ayer vendiste **${pesos(ayer.ventas)}** en ${ayer.pedidos} pedidos.`;
	}
	if (/precio|subí|subi|aument/.test(q)) {
		const pct = Number(q.match(/(\d+)\s*%/)?.[1] ?? 8);
		const top = [...d.contexto.productos].sort((a, b) => b.vendidos30d - a.vendidos30d)[0];
		return top
			? `Con un ${pct}% más, ${top.nombre} pasaría de ${pesos(top.precio)} a **${pesos(Math.round((top.precio * (1 + pct / 100)) / 100) * 100)}**. Lo aplicás desde **Listas de precio** → «Actualizar precios por inflación».`
			: 'Todavía no tenés productos cargados para subirles el precio.';
	}
	if (/client|vuelv/.test(q))
		return `Tenés **${d.contexto.clientes.sinVolver} clientes** que compraron más de una vez y hace un mes que no vuelven. Los ves en **Clientes → Para recuperar**.`;
	const top = d.semana.masVendidos[0];
	return top
		? `Esta semana lo que más salió fue **${top.nombre}** (${top.unidades} unidades). Un cupón para ese producto lo creás en **Cupones**.`
		: 'Todavía no tengo ventas para mirar. Cuando entren los primeros pedidos, te cuento cómo vienen.';
}
