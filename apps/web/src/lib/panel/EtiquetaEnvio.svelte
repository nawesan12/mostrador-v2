<script lang="ts">
	import { code39 } from './listas/code39.ts';
	import { NOMBRE_ENVIO } from './pedidos/nombres.ts';

	/**
	 * Etiqueta de envío de 10 × 15 cm (el tamaño de las impresoras térmicas y de media A6): para quién
	 * va, adónde, quién lo manda y el número del pedido con su código de barras.
	 */
	let {
		p,
		remitente
	}: {
		p: {
			numero: number;
			nombre: string;
			telefono: string | null;
			envio: string;
			ciudad: string;
			entrega: { direccion: string; cp: string } | null;
			unidades: number;
		};
		remitente: { nombre: string; direccion: string; whatsapp: string };
	} = $props();
</script>

<div class="eti-envio">
	<div class="cab">
		<span class="via">{NOMBRE_ENVIO[p.envio] ?? p.envio}</span>
		<span class="num">#{p.numero}</span>
	</div>
	<div class="para">
		<div class="et">Para</div>
		<div class="nombre">{p.nombre}</div>
		{#if p.entrega}
			<div class="dir">{p.entrega.direccion}</div>
			<div class="dir">CP {p.entrega.cp}{p.ciudad ? ` · ${p.ciudad}` : ''}</div>
		{:else}
			<div class="dir">Retira en el local</div>
		{/if}
		{#if p.telefono}<div class="tel">Tel. {p.telefono}</div>{/if}
	</div>
	<div class="de">
		<div class="et">Remite</div>
		<div class="r-n">{remitente.nombre}</div>
		{#if remitente.direccion}<div>{remitente.direccion}</div>{/if}
		{#if remitente.whatsapp}<div>WhatsApp {remitente.whatsapp}</div>{/if}
	</div>
	<div class="pie">
		<!-- eslint-disable-next-line svelte/no-at-html-tags -- SVG armado acá (code39.ts), sin datos de afuera salvo el número -->
		<div class="barras">{@html code39(`PED-${p.numero}`, 56)}</div>
		<div class="bultos">{p.unidades} {p.unidades === 1 ? 'unidad' : 'unidades'}</div>
	</div>
</div>

<style>
	.eti-envio {
		align-items: stretch;
		width: 100mm;
		height: 148mm;
		margin: 0 auto;
		padding: 7mm;
		box-sizing: border-box;
		display: flex;
		flex-direction: column;
		gap: 5mm;
		background: #fff;
		color: #000;
		font-family: 'Instrument Sans', Helvetica, Arial, sans-serif;
	}
	.cab {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding-bottom: 3mm;
		border-bottom: 0.6mm solid #000;
	}
	.via {
		font-size: 22px;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}
	.num {
		font-size: 26px;
		font-weight: 700;
	}
	.et {
		font-size: 10px;
		font-weight: 700;
		letter-spacing: 0.14em;
		text-transform: uppercase;
	}
	.para {
		flex: 1;
		display: flex;
		flex-direction: column;
		gap: 1.5mm;
	}
	.nombre {
		font-size: 24px;
		font-weight: 700;
		line-height: 1.1;
	}
	.dir {
		font-size: 17px;
		line-height: 1.3;
	}
	.tel {
		font-size: 15px;
	}
	.de {
		padding-top: 3mm;
		border-top: 0.3mm dashed #000;
		font-size: 12px;
		line-height: 1.4;
	}
	.r-n {
		font-weight: 700;
	}
	.pie {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 4mm;
	}
	.barras :global(svg) {
		display: block;
		width: 60mm;
		height: 14mm;
	}
	.bultos {
		font-size: 12px;
		font-weight: 700;
	}
</style>
