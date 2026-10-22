<script lang="ts">
	import Seo from '#lib/components/Seo.svelte';
	import { pesos } from '#lib/panel/formato.ts';
	import { code39 } from '#lib/panel/listas/code39.ts';

	/** Hoja de etiquetas para imprimir (A4, 3 por fila). */
	let { data } = $props();
</script>

<Seo
	title="Códigos para imprimir · Panel"
	description="Etiquetas con código de barras"
	path="/panel/inventario"
	noindex
/>

<div class="hoja">
	<div class="barra no-imprimir">
		<span>{data.productos.length} {data.productos.length === 1 ? 'etiqueta' : 'etiquetas'}</span>
		<button type="button" class="btn btn-primary" onclick={() => window.print()}>Imprimir</button>
	</div>
	<div class="grilla">
		{#each data.productos as p (p.id)}
			<div class="etiqueta">
				<div class="n">{p.nombre}</div>
				<!-- eslint-disable-next-line svelte/no-at-html-tags -- SVG armado acá, sólo con barras -->
				<div class="codigo">{@html code39(p.sku || p.id)}</div>
				<div class="pie"><span>{p.sku}</span><b>{pesos(p.precio)}</b></div>
			</div>
		{/each}
	</div>
</div>

<style>
	.hoja {
		padding: 24px;
		background: #fff;
		color: #111;
	}
	.barra {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: 16px;
		font-weight: 600;
	}
	.grilla {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(62mm, 1fr));
		gap: 4mm;
	}
	.etiqueta {
		border: 1px dashed #bbb;
		border-radius: 6px;
		padding: 3mm 4mm;
		break-inside: avoid;
	}
	.n {
		font-size: 12px;
		font-weight: 600;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.codigo {
		margin: 2mm 0 1mm;
	}
	.codigo :global(svg) {
		width: 100%;
		height: 14mm;
		fill: #000;
	}
	.pie {
		display: flex;
		justify-content: space-between;
		font-size: 11px;
	}
	@media print {
		.no-imprimir {
			display: none;
		}
		.hoja {
			padding: 0;
		}
	}
</style>
