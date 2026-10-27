<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Segmentado from '#lib/components/Segmentado.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import TarjetaA6 from '#lib/panel/TarjetaA6.svelte';
	import { EJEMPLO, type Variables } from '#lib/panel/tarjeta.ts';

	/**
	 * Tarjetas A6 listas para imprimir (o guardar como PDF desde el diálogo del navegador): una por
	 * hoja A6, o 4 por A4 con marcas de corte. En el A4 los dorsos van espejados, así coinciden al
	 * imprimir doble faz (girando por el borde largo).
	 */
	let { data } = $props();

	type Una = { v: Variables; foto: string | null };
	const tarjetas = $derived<Una[]>(
		data.ejemplo
			? [{ v: EJEMPLO, foto: data.foto }]
			: data.pedidos.map((p) => ({
					v: {
						nombre: p.nombre,
						fecha: new Date(p.creado),
						productos: p.productos,
						pedido: p.numero
					},
					foto: p.foto ?? data.foto
				}))
	);
	// De a 4 por hoja A4.
	const hojas = $derived(
		Array.from({ length: Math.ceil(tarjetas.length / 4) }, (_, i) =>
			tarjetas.slice(i * 4, i * 4 + 4)
		)
	);
	/** El dorso de una hoja A4: cada fila con las columnas cambiadas. */
	const espejo = (h: Una[]) => [h[1], h[0], h[3], h[2]];

	function formato(f: 'a6' | 'a4') {
		const u = new URL(page.url.href);
		u.searchParams.set('f', f);
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- la misma página con otro formato
		goto(u.pathname + u.search, { replaceState: true });
	}
	const FORMATOS = [
		['a6', 'Una por hoja A6'],
		['a4', '4 por A4']
	] as const;
</script>

<Seo
	title="Imprimir tarjetas · Panel"
	description="Tarjetas A6 para los paquetes"
	path="/panel/pedidos/imprimir"
	noindex
/>

<div class="barra">
	<a class="volver" href={resolve('/panel/pedidos')}
		><Icon icon={I.phArrowLeftBold} size={14} />Pedidos</a
	>
	<span class="b-t"
		>{tarjetas.length}
		{tarjetas.length === 1 ? 'tarjeta' : 'tarjetas'}</span
	>
	<Segmentado opciones={FORMATOS} value={data.formato} label="Formato" onchange={formato} />
	<a class="editar" href={resolve('/panel/tarjeta-para-paquetes')}>Editar la tarjeta</a>
	<button type="button" class="imprimir" onclick={() => print()} disabled={!tarjetas.length}
		><Icon icon={I.phPrinterFill} size={15} />Imprimir</button
	>
</div>

{#if !tarjetas.length}
	<p class="vacio">No encontramos esos pedidos.</p>
{:else if data.formato === 'a6'}
	<div class="hojas">
		{#each tarjetas as x, i (i)}
			<div class="hoja a6">
				<TarjetaA6
					t={data.tarjeta}
					cara="frente"
					tienda={data.tienda}
					paleta={data.paleta}
					foto={x.foto}
					v={x.v}
				/>
			</div>
			<div class="hoja a6">
				<TarjetaA6
					t={data.tarjeta}
					cara="dorso"
					tienda={data.tienda}
					paleta={data.paleta}
					foto={x.foto}
					v={x.v}
				/>
			</div>
		{/each}
	</div>
{:else}
	<div class="hojas">
		{#each hojas as h, i (i)}
			{#each ['frente', 'dorso'] as const as cara (cara)}
				<div class="hoja a4">
					{#each cara === 'frente' ? h : espejo(h) as x, j (j)}
						<div class="celda">
							{#if x}<TarjetaA6
									t={data.tarjeta}
									{cara}
									tienda={data.tienda}
									paleta={data.paleta}
									foto={x.foto}
									v={x.v}
								/>{/if}
						</div>
					{/each}
					<!-- Marcas de corte: en el medio de cada borde. -->
					<span class="corte v arriba"></span><span class="corte v abajo"></span><span
						class="corte h izq"
					></span><span class="corte h der"></span>
				</div>
			{/each}
		{/each}
	</div>
{/if}

<style>
	/* Chrome no tiene la palabra A6: van las medidas. */
	@page a6 {
		size: 105mm 148mm;
		margin: 0;
	}
	@page a4 {
		size: A4;
		margin: 0;
	}
	:global(body) {
		background: #e9e2da;
	}
	.barra {
		position: sticky;
		top: 0;
		z-index: 2;
		display: flex;
		align-items: center;
		gap: 12px;
		flex-wrap: wrap;
		padding: 12px 16px;
		background: rgba(251, 247, 242, 0.92);
		backdrop-filter: blur(10px);
		border-bottom: 1px solid rgba(58, 29, 18, 0.09);
		font-family: 'Instrument Sans', system-ui, sans-serif;
		color: #2a140c;
	}
	.volver,
	.editar {
		display: flex;
		align-items: center;
		gap: 6px;
		color: inherit;
		font-size: 13.5px;
		font-weight: 600;
		text-decoration: none;
	}
	.b-t {
		flex: 1;
		font-size: 14px;
		color: #6b5244;
	}
	.imprimir {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 18px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #e06a3e, #d2582f);
		color: #fff;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
	}
	.vacio {
		padding: 40px 16px;
		text-align: center;
		color: #6b5244;
	}
	.hojas {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 20px;
		padding: 24px 16px 60px;
	}
	.hoja {
		position: relative;
		background: #fff;
		box-shadow: 0 20px 40px -24px rgba(42, 20, 12, 0.5);
		overflow: hidden;
	}
	.hoja.a6 {
		page: a6;
		width: 105mm;
		height: 148mm;
	}
	.hoja.a4 {
		page: a4;
		width: 210mm;
		height: 297mm;
		display: grid;
		grid-template-columns: 105mm 105mm;
		grid-template-rows: 148mm 148mm;
	}
	.celda {
		display: flex;
		align-items: center;
	}
	.corte {
		position: absolute;
		background: #2a140c;
	}
	.corte.v {
		left: calc(105mm - 0.1mm);
		width: 0.2mm;
		height: 5mm;
	}
	.corte.arriba {
		top: 0;
	}
	.corte.abajo {
		bottom: 0;
	}
	.corte.h {
		top: calc(148mm - 0.1mm);
		height: 0.2mm;
		width: 5mm;
	}
	.corte.izq {
		left: 0;
	}
	.corte.der {
		right: 0;
	}
	/* En la pantalla, la hoja se achica para que entre (a 375 px, una A6 mide ~400 px). */
	@media screen and (max-width: 860px) {
		.hoja.a4 {
			zoom: 0.42;
		}
		.hoja.a6 {
			zoom: 0.8;
		}
	}
	@media print {
		:global(body) {
			background: #fff;
		}
		.barra {
			display: none;
		}
		.hojas {
			display: block;
			padding: 0;
		}
		.hoja {
			box-shadow: none;
			break-after: page;
		}
		.hoja:last-child {
			break-after: auto;
		}
		:global(*) {
			print-color-adjust: exact;
			-webkit-print-color-adjust: exact;
		}
	}
</style>
