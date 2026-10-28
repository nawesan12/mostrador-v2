<script lang="ts">
	import { onMount } from 'svelte';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import type { CuponCheckout } from './cupon.svelte.ts';

	/**
	 * "¿Tenés un cupón?" en el checkout. Cada plantilla le pasa sus colores, radio y tipografía para
	 * que se vea como el resto de su formulario. Manda el código aplicado en el campo `cupon`.
	 */
	let {
		cupon,
		lineas,
		email = '',
		tinta,
		suave,
		borde,
		fondo,
		acento,
		sobreAcento,
		radio = '12px',
		fuente = 'inherit',
		error = '#b3261e'
	}: {
		cupon: CuponCheckout;
		lineas: { id: string; cantidad: number }[];
		email?: string;
		tinta: string;
		suave: string;
		borde: string;
		fondo: string;
		acento: string;
		sobreAcento: string;
		radio?: string;
		fuente?: string;
		/** El color del motivo cuando no aplica (en paneles oscuros, uno claro). */
		error?: string;
	} = $props();

	const fmt = (n: number) => '$\u00a0' + Math.round(n).toLocaleString('es-AR');
	// Si llegó por el link de un cupón, se carga y se aplica solo.
	onMount(() => {
		const c = cupon.delLink();
		if (c && !cupon.aplicado && raiz) {
			cupon.codigo = c;
			aplicar(raiz);
		}
	});
	// El email del mismo formulario (para "primera compra" y "una vez por cliente"), si ya está.
	function aplicar(el: HTMLElement) {
		const campo = el.closest('form')?.querySelector<HTMLInputElement>('[name="email"]');
		cupon.aplicar(lineas, campo?.value.trim() || email);
	}
	// Si cambia el email con un cupón cargado, se vuelve a evaluar (primera compra, una vez por
	// cliente): no queda el resultado del email anterior.
	let raiz = $state<HTMLElement>();
	$effect(() => {
		const campo = raiz?.closest('form')?.querySelector<HTMLInputElement>('[name="email"]');
		if (!campo || !raiz) return;
		const el = raiz;
		const alCambiar = () => {
			// También el que no aplicó con el email anterior (o sin email, el del link): se reintenta.
			if (cupon.aplicado) cupon.codigo = cupon.aplicado;
			else if (!cupon.codigo || !cupon.motivo) return;
			aplicar(el);
		};
		campo.addEventListener('change', alCambiar);
		return () => campo.removeEventListener('change', alCambiar);
	});
</script>

<div
	bind:this={raiz}
	class="cupon"
	style="--t:{tinta};--s:{suave};--b:{borde};--f:{fondo};--a:{acento};--sa:{sobreAcento};--r:{radio};--e:{error};font-family:{fuente}"
>
	<input type="hidden" name="cupon" value={cupon.aplicado ?? ''} />
	{#if cupon.ok}
		<div class="aplicado" role="status">
			<Icon icon={I.phTagFill} size={16} />
			<span class="txt"
				><b>{cupon.aplicado}</b>
				{cupon.ok.envioGratis ? 'Envío gratis' : `− ${fmt(cupon.ok.descuento)}`}</span
			>
			<button type="button" onclick={() => cupon.quitar()}>Quitar</button>
		</div>
	{:else}
		<label for="cupon-codigo">¿Tenés un cupón?</label>
		<div class="fila">
			<input
				id="cupon-codigo"
				bind:value={cupon.codigo}
				autocomplete="off"
				autocapitalize="characters"
				spellcheck="false"
				maxlength="20"
				placeholder="CÓDIGO"
				onkeydown={(e) => {
					if (e.key === 'Enter') {
						e.preventDefault();
						aplicar(e.currentTarget);
					}
				}}
			/>
			<button
				type="button"
				disabled={!cupon.codigo.trim() || cupon.cargando}
				onclick={(e) => aplicar(e.currentTarget)}
				>{cupon.cargando ? 'Revisando…' : 'Aplicar'}</button
			>
		</div>
		{#if cupon.motivo}<p class="motivo" role="alert">{cupon.motivo}</p>{/if}
	{/if}
</div>

<style>
	.cupon {
		display: grid;
		gap: 6px;
		color: var(--t);
	}
	label {
		font-size: 13px;
		font-weight: 600;
	}
	.fila {
		display: flex;
		gap: 8px;
		min-width: 0;
	}
	input:not([type='hidden']) {
		flex: 1;
		min-width: 0;
		height: 44px;
		padding: 0 14px;
		border: 1px solid var(--b);
		border-radius: var(--r);
		background: var(--f);
		color: var(--t);
		font: inherit;
		font-size: 15px;
		letter-spacing: 0.06em;
		text-transform: uppercase;
	}
	input::placeholder {
		color: var(--s);
		letter-spacing: 0.1em;
	}
	button {
		flex: none;
		height: 44px;
		padding: 0 16px;
		border: 0;
		border-radius: var(--r);
		background: var(--a);
		color: var(--sa);
		font: inherit;
		font-size: 14px;
		font-weight: 600;
		cursor: pointer;
	}
	button:disabled {
		opacity: 0.45;
		cursor: not-allowed;
	}
	.aplicado {
		display: flex;
		align-items: center;
		gap: 8px;
		min-height: 44px;
		padding: 0 6px 0 14px;
		border: 1px dashed var(--a);
		border-radius: var(--r);
		font-size: 14px;
	}
	.aplicado .txt {
		flex: 1;
		min-width: 0;
	}
	.aplicado button {
		height: 34px;
		padding: 0 12px;
		background: transparent;
		color: var(--t);
		text-decoration: underline;
	}
	.motivo {
		margin: 0;
		font-size: 13px;
		color: var(--e);
		text-wrap: pretty;
	}
</style>
