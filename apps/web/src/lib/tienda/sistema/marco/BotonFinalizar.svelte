<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { faltantes } from '../../faltantes.ts';
	import { vende } from '../cobros.ts';
	import { usarTienda } from '../contexto.ts';
	import { waHref } from '../datos.ts';
	import { precio } from '../formato.ts';
	import Boton from '../primitivas/Boton.svelte';

	/**
	 * «Finalizar compra» del carrito, igual en la página, el panel y la bolsa (los 17 presets):
	 * - Con algo agotado o con menos stock: «Revisá el carrito para seguir», que lleva a esa línea
	 *   (pagar fallaría; mejor decirlo antes que recalcular sin la línea).
	 * - Sin Mercado Pago ni transferencia: «Pedir por WhatsApp» con el pedido escrito (o «Consultar»).
	 * - Si no, al checkout.
	 */
	let { alIr }: { alIr?: () => void } = $props();
	const ctx = usarTienda();
	const t = $derived(ctx.tienda);
	const faltan = $derived(faltantes(ctx.carrito.lineas, t.productos));
	const sinCobro = $derived(!t.vitrina && !vende(t));
	const pedidoWa = $derived.by(() => {
		const lineas = ctx.carrito.detalle(t.productos, t.mayorista ?? null);
		const total = lineas.reduce((s, l) => s + l.subtotal, 0);
		const items = lineas
			.map((l) => `• ${l.cantidad} × ${l.producto.nombre}${l.variante ? ` (${l.variante})` : ''}`)
			.join('\n');
		return waHref(t, `¡Hola! Quiero hacer este pedido:\n${items}\nTotal: ${precio(total)}`);
	});
	function revisar() {
		const primera = faltan[0];
		const el = primera
			? document.querySelector(`[data-linea="${CSS.escape(primera.clave)}"]`)
			: null;
		el?.scrollIntoView({ behavior: 'smooth', block: 'center' });
		(el?.querySelector('button') as HTMLButtonElement | null)?.focus({ preventScroll: true });
	}
</script>

{#if faltan.length}
	<Boton bloque onclick={revisar}>Revisá el carrito para seguir</Boton>
{:else if sinCobro}
	<Boton href={pedidoWa || `${ctx.base}/nosotros`} bloque externo={!!pedidoWa} onclick={alIr}
		>{#if pedidoWa}<Icon icon={I.phWhatsappLogo} size={18} />Pedir por WhatsApp{:else}Consultar{/if}</Boton
	>
{:else}
	<Boton href="{ctx.base}/checkout" bloque onclick={alIr}>Finalizar compra</Boton>
{/if}
