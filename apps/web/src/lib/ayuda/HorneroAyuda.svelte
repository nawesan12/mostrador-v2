<script lang="ts">
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { buscar, type Articulo } from './articulos.ts';

	/**
	 * "Preguntale a Hornero" del centro de ayuda (handoff). En el sitio público Hornero responde con
	 * los artículos de ayuda: busca el que mejor contesta y lo resume, sin llamar al modelo.
	 */
	let pregunta = $state('');
	let respuesta = $state<{ a: Articulo | null; q: string } | null>(null);
	function preguntar(e: SubmitEvent) {
		e.preventDefault();
		const q = pregunta.trim();
		if (!q) return;
		respuesta = { a: buscar(q, 1)[0] ?? null, q };
		pregunta = '';
	}
</script>

<div class="hornero">
	<div class="h-cab">
		<img src="/hornero/avatar.png" alt="" width="52" height="52" />
		<div>
			<b>Preguntale a Hornero</b>
			<div class="h-s">Busca en la ayuda y te lleva a la respuesta</div>
		</div>
	</div>
	<div class="burbuja" aria-live="polite">
		{#if !respuesta}
			¿Querés saber cómo activar las 6 cuotas sin interés en Mercado Pago? Preguntame lo que sea.
		{:else if respuesta.a}
			{respuesta.a.resumen}
			<a
				href={resolve('/(sitio)/ayuda/[categoria]/[articulo]', {
					categoria: respuesta.a.categoria,
					articulo: respuesta.a.slug
				})}>{respuesta.a.titulo} <Icon icon={I.phArrowRightBold} size={12} /></a
			>
		{:else}
			No encontré algo sobre «{respuesta.q}». Escribinos desde
			<a href={resolve('/(sitio)/contacto')}>Contacto</a> y te respondemos en el día.
		{/if}
	</div>
	<form class="campo" onsubmit={preguntar}>
		<input
			bind:value={pregunta}
			placeholder="Escribile a Hornero…"
			aria-label="Pregunta para Hornero"
			maxlength="200"
		/>
		<button aria-label="Preguntar"><Icon icon={I.phPaperPlaneTiltFill} size={16} /></button>
	</form>
</div>

<style>
	.hornero {
		padding: 22px;
		border-radius: 24px;
		background: #2a140c;
		color: #fbf7f2;
	}
	.h-cab {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.h-cab img {
		border-radius: 50%;
		object-fit: cover;
		background: #fcede5;
	}
	.h-cab b {
		font-size: 16px;
	}
	.h-s {
		font-size: 13px;
		color: #d9c8ba;
	}
	.burbuja {
		margin-top: 14px;
		padding: 12px 14px;
		border-radius: 16px 16px 16px 4px;
		background: rgba(251, 247, 242, 0.08);
		font-size: 14px;
		line-height: 1.5;
	}
	.burbuja a {
		display: inline-flex;
		align-items: center;
		gap: 5px;
		margin-top: 6px;
		color: #f5b796;
		font-weight: 600;
	}
	.campo {
		margin-top: 12px;
		display: flex;
		padding: 5px;
		border-radius: 999px;
		background: #fff;
	}
	.campo input {
		flex: 1;
		min-width: 0;
		padding: 9px 12px;
		border: 0;
		background: transparent;
		font:
			500 14px 'Instrument Sans',
			sans-serif;
		color: #2a140c;
		outline: none;
	}
	.campo button {
		width: 36px;
		height: 36px;
		border-radius: 50%;
		border: 0;
		background: #b0441d;
		color: #fff;
		display: grid;
		place-items: center;
	}
</style>
