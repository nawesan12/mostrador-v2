<script lang="ts">
	import { untrack } from 'svelte';
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import Segmentado from '#lib/components/Segmentado.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { anim } from '#lib/panel/anim.ts';
	import TarjetaA6 from '#lib/panel/TarjetaA6.svelte';
	import {
		DESTINOS,
		EJEMPLO,
		linkQr,
		MODELOS,
		TOPES,
		VARIABLES,
		type DestinoQr,
		type Tarjeta
	} from '#lib/panel/tarjeta.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/**
	 * Tarjeta para paquetes (A6, pieza 7a de Imprimibles). El handoff no la diseña: se arma con las
	 * piezas de Mi tienda (cabecera, panel de ajustes y vista previa en vivo) y del editor de producto.
	 */
	let { data } = $props();
	const ui = getUi();

	// Lo que se edita: arranca con lo guardado.
	let t = $state<Tarjeta>(untrack(() => ({ ...data.tarjeta })));
	let guardando = $state(false);
	// El último campo de texto tocado: ahí van las variables.
	let campo = $state<'titulo' | 'mensaje' | 'dorsoTitulo' | 'dorsoTexto'>('titulo');

	const ICONOS: Record<DestinoQr, typeof I.phStarFill> = {
		resena: I.phStarFill,
		cupon: I.phTicketFill,
		instagram: I.phInstagramLogoFill,
		whatsapp: I.phWhatsappLogoFill
	};
	const falta = $derived(
		linkQr(t, data.tienda)
			? null
			: t.qr === 'cupon'
				? 'Elegí el cupón.'
				: t.qr === 'instagram'
					? 'Poné la cuenta de Instagram.'
					: 'Cargá el WhatsApp de la tienda en Configuración › Negocio.'
	);

	function variable(v: string) {
		const nuevo = `${t[campo]}${t[campo] && !t[campo].endsWith(' ') ? ' ' : ''}${v}`;
		if (nuevo.length <= TOPES[campo]) t[campo] = nuevo;
	}

	async function guardar() {
		guardando = true;
		const r = await fetch('/panel/tarjeta-para-paquetes', {
			method: 'PUT',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(t)
		});
		guardando = false;
		if (!r.ok)
			return ui.toast({
				k: 'err',
				title: 'No se guardó la tarjeta',
				desc:
					((await r.json().catch(() => null)) as { message?: string } | null)?.message ??
					'Probá de nuevo.'
			});
		ui.toast({ k: 'ok', title: 'Tarjeta guardada', desc: 'Se usa en los pedidos que imprimas' });
	}
</script>

<Seo
	title="Tarjeta para paquetes · Panel"
	description="La tarjeta A6 que va en cada paquete"
	path="/panel/tarjeta-para-paquetes"
	noindex
/>

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Tarjeta para paquetes</h1>
			<div class="c-s">A6 · 105 × 148 mm · se imprime desde cada pedido o en lote</div>
		</div>
		<a class="cambiar" href="{resolve('/panel/pedidos/imprimir')}?ejemplo=1" target="_blank"
			><Icon icon={I.phPrinterFill} size={14} />Imprimir una de prueba</a
		>
		<button type="button" class="publicar" onclick={guardar} disabled={guardando}
			>{guardando ? 'Guardando…' : 'Guardar'}</button
		>
	</div>

	<div class="editor" use:anim={{ t: 'rise', delay: 60 }}>
		<div class="vista" aria-label="Vista previa" role="region">
			<div class="caras">
				<div class="cara">
					<span class="c-et">Frente</span>
					<div class="escala">
						<TarjetaA6
							{t}
							cara="frente"
							tienda={data.tienda}
							paleta={data.paleta}
							foto={data.foto}
							v={EJEMPLO}
						/>
					</div>
				</div>
				<div class="cara">
					<span class="c-et">Dorso</span>
					<div class="escala">
						<TarjetaA6
							{t}
							cara="dorso"
							tienda={data.tienda}
							paleta={data.paleta}
							foto={data.foto}
							v={EJEMPLO}
						/>
					</div>
				</div>
			</div>
			<p class="nota">
				<Icon icon={I.phInfoFill} size={15} />Con el pedido de ejemplo de Lucía. Imprimí al 100 %
				(sin «ajustar a la página»): el texto queda dentro del margen seguro de 10&nbsp;mm.
			</p>
		</div>

		<div class="panelito">
			<div class="p-t">Modelo</div>
			<Segmentado opciones={MODELOS} bind:value={t.modelo} label="Modelo del frente" />

			<div class="p-t sep">Frente</div>
			<label class="campo"
				><span class="c-fila"
					>Título<span class="cuenta">{t.titulo.length}/{TOPES.titulo}</span></span
				><input
					bind:value={t.titulo}
					maxlength={TOPES.titulo}
					onfocus={() => (campo = 'titulo')}
				/></label
			>
			<label class="campo"
				><span class="c-fila"
					>Mensaje<span class="cuenta">{t.mensaje.length}/{TOPES.mensaje}</span></span
				><textarea
					bind:value={t.mensaje}
					maxlength={TOPES.mensaje}
					onfocus={() => (campo = 'mensaje')}></textarea></label
			>

			<div class="p-t sep">Dorso</div>
			<label class="campo"
				><span class="c-fila"
					>Título<span class="cuenta">{t.dorsoTitulo.length}/{TOPES.dorsoTitulo}</span></span
				><input
					bind:value={t.dorsoTitulo}
					maxlength={TOPES.dorsoTitulo}
					onfocus={() => (campo = 'dorsoTitulo')}
				/></label
			>
			<label class="campo"
				><span class="c-fila"
					>Texto<span class="cuenta">{t.dorsoTexto.length}/{TOPES.dorsoTexto}</span></span
				><textarea
					bind:value={t.dorsoTexto}
					maxlength={TOPES.dorsoTexto}
					onfocus={() => (campo = 'dorsoTexto')}></textarea></label
			>
			<div class="vars" role="group" aria-label="Sumar un dato del pedido">
				{#each VARIABLES as [v, d] (v)}<button type="button" title={d} onclick={() => variable(v)}
						>{v}</button
					>{/each}
			</div>

			<div class="p-t sep">El QR lleva a</div>
			{#each DESTINOS as [d, texto] (d)}
				<button type="button" class="sec" aria-pressed={t.qr === d} onclick={() => (t.qr = d)}
					><Icon icon={ICONOS[d]} size={16} style="color:#C24E25" /><span class="s-t">{texto}</span
					>{#if t.qr === d}<Icon icon={I.phCheckBold} size={14} />{/if}</button
				>
			{/each}
			{#if t.qr === 'cupon'}
				{#if data.cupones.length}
					<label class="campo"
						>Cupón<select bind:value={t.cupon}
							><option value="">Elegí uno</option>{#each data.cupones as c (c)}<option value={c}
									>{c}</option
								>{/each}</select
						></label
					>
				{:else}
					<p class="aviso">
						No tenés cupones activos. <a href={resolve('/panel/cupones')}>Creá uno</a> y volvé.
					</p>
				{/if}
			{:else if t.qr === 'instagram'}
				<label class="campo"
					>Cuenta de Instagram<input
						bind:value={t.instagram}
						maxlength="30"
						placeholder="tutienda"
						autocapitalize="off"
					/></label
				>
			{/if}
			{#if falta && !(t.qr === 'cupon' && !data.cupones.length)}<p class="aviso">{falta}</p>{/if}
		</div>
	</div>
</div>

<style>
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	.cab {
		display: flex;
		align-items: end;
		gap: 12px;
		flex-wrap: wrap;
	}
	.c-txt {
		flex: 1;
		min-width: 220px;
	}
	h1 {
		margin: 0;
		font-size: 30px;
		letter-spacing: -0.035em;
		font-weight: 600;
	}
	.c-s {
		margin-top: 2px;
		font-size: 15px;
		color: var(--mu);
	}
	.cambiar {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 16px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font-size: 13.5px;
		font-weight: 600;
		box-shadow: var(--sh);
		white-space: nowrap;
		text-decoration: none;
	}
	.publicar {
		padding: 11px 20px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #e06a3e, #d2582f);
		color: #fff;
		font:
			600 14px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
	}
	.editor {
		display: grid;
		grid-template-columns: minmax(0, 1fr) 340px;
		gap: 16px;
		align-items: start;
	}
	.vista {
		padding: 22px;
		border-radius: 22px;
		background: var(--sf2);
	}
	.caras {
		display: flex;
		justify-content: center;
		gap: 22px;
		flex-wrap: wrap;
	}
	.cara {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.c-et {
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
	}
	.escala {
		zoom: 0.72;
		box-shadow: 0 24px 50px -26px rgba(42, 20, 12, 0.45);
	}
	.nota {
		margin: 16px 0 0;
		display: flex;
		gap: 8px;
		font-size: 12.5px;
		line-height: 1.5;
		color: var(--mu);
		text-wrap: pretty;
	}
	.nota :global(svg) {
		flex: none;
		margin-top: 2px;
	}
	.panelito {
		padding: 14px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.p-t {
		font-size: 14px;
		font-weight: 600;
		padding: 2px 4px 2px;
	}
	.p-t.sep {
		margin-top: 8px;
	}
	.campo {
		display: flex;
		flex-direction: column;
		gap: 4px;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
	}
	.c-fila {
		display: flex;
		justify-content: space-between;
	}
	.cuenta {
		font-weight: 500;
		font-variant-numeric: tabular-nums;
	}
	.campo input,
	.campo textarea,
	.campo select {
		padding: 11px 12px;
		border-radius: 12px;
		border: 1px solid var(--ln2);
		background: var(--bg);
		font:
			500 15px 'Instrument Sans',
			sans-serif;
		color: var(--tx);
		outline: none;
	}
	.campo textarea {
		min-height: 64px;
		font-size: 14px;
		line-height: 1.5;
		resize: vertical;
	}
	.campo input:focus,
	.campo textarea:focus,
	.campo select:focus {
		border-color: #d2582f;
	}
	.vars {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
	}
	.vars button {
		padding: 5px 10px;
		border-radius: 999px;
		border: 0;
		background: var(--chip);
		color: var(--tx);
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
	}
	.sec {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 9px 10px;
		border-radius: 14px;
		background: var(--sf);
		border: 1px solid var(--ln);
		color: var(--tx);
		text-align: left;
		font: inherit;
	}
	.sec[aria-pressed='true'] {
		border-color: #d2582f;
	}
	.s-t {
		flex: 1;
		font-size: 13.5px;
		font-weight: 600;
	}
	.aviso {
		margin: 0;
		padding: 10px 12px;
		border-radius: 12px;
		background: #fcede5;
		color: #a8401c;
		font-size: 13px;
		line-height: 1.45;
	}
	.aviso a {
		color: inherit;
		font-weight: 600;
	}
	@media (max-width: 900px) {
		.editor {
			grid-template-columns: minmax(0, 1fr);
		}
		.vista {
			padding: 14px;
		}
		.escala {
			zoom: 0.4;
		}
		.caras {
			gap: 12px;
			flex-wrap: nowrap;
		}
	}
</style>
