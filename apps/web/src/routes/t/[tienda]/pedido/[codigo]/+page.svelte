<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import { imagenTienda, textosCompartir } from '#lib/tienda/compartir.ts';
	import { onMount, untrack } from 'svelte';
	import Plantilla from '#lib/tienda/Plantilla.svelte';
	import { compra, preparar } from '#lib/tienda/medicion.ts';

	import { conectarVidriera } from '#lib/tiempo-real/publico.svelte.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	let { data } = $props();
	const ui = getUi();
	onMount(() => {
		preparar(data.tienda);
		compra(data.pedido);
		// Tiempo real (§44): sólo de este pedido. Al acreditarse el pago, «Recibimos tu pago».
		if (!data.tiempoReal) return;
		const c = conectarVidriera({
			slug: data.tienda.slug,
			codigo: data.pedido.codigo,
			origen: data.tiempoReal.origen
		});
		return () => c?.detener();
	});
	let pagado = untrack(() => data.pedido.pagado);
	$effect(() => {
		if (data.pedido.pagado && !pagado)
			ui.toast({
				k: 'ok',
				title: '¡Recibimos tu pago!',
				desc: `Tu pedido #${data.pedido.numero} ya está pago. Te escribimos con cada novedad.`
			});
		pagado = data.pedido.pagado;
	});
	// Pagado, la plantilla ya no muestra los datos para transferir (es lo único que mira de `pago`).
	const pedido = $derived(
		data.pedido.pagado && data.pedido.pago === 'transferencia'
			? { ...data.pedido, pago: 'pagado' }
			: data.pedido
	);
</script>

<Seo
	title="Pedido #{data.pedido.numero} · {data.tienda.nombre}"
	description={textosCompartir(data.tienda).descripcion}
	path="/t/{data.tienda.slug}"
	noindex
	imagen={imagenTienda(data.tienda)}
/>
<Plantilla tienda={data.tienda} pagina="listo" {pedido} />
