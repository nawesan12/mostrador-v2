<script lang="ts">
	import MarcaIcono from './MarcaIcono.svelte';

	/** Alerta en página: para lo que no se puede ignorar. Queda hasta que se resuelve o se descarta. */
	type Tipo = 'info' | 'ok' | 'warn' | 'err';
	let {
		tipo,
		titulo,
		texto,
		accion,
		onaccion
	}: { tipo: Tipo; titulo: string; texto?: string; accion?: string; onaccion?: () => void } =
		$props();

	const COLORES: Record<Tipo, [string, string]> = {
		info: ['rgba(210,88,47,.08)', 'rgba(210,88,47,.25)'],
		ok: ['rgba(47,143,99,.08)', 'rgba(47,143,99,.28)'],
		warn: ['rgba(227,169,59,.12)', 'rgba(227,169,59,.4)'],
		err: ['rgba(184,58,46,.07)', 'rgba(184,58,46,.28)']
	};
	const [bg, bd] = $derived(COLORES[tipo]);
</script>

<div
	class="alerta"
	style:background={bg}
	style:border-color={bd}
	role={tipo === 'err' ? 'alert' : 'status'}
>
	<span class="ico"><MarcaIcono k={tipo} size={36} radius={11} /></span>
	<div class="cuerpo">
		<div class="titulo">{titulo}</div>
		{#if texto}<div class="texto">{texto}</div>{/if}
	</div>
	{#if accion}
		<button type="button" class="btn btn-dark btn-sm" onclick={onaccion}>{accion}</button>
	{/if}
</div>

<style>
	.alerta {
		display: flex;
		align-items: flex-start;
		gap: 12px;
		padding: 14px;
		border-radius: 18px;
		border: 1px solid;
		color: var(--tx);
	}
	/* El ícono va sobre blanco (handoff). */
	.ico :global(.marca) {
		background: #fff !important;
	}
	.cuerpo {
		flex: 1;
		min-width: 0;
	}
	.titulo {
		font-size: 14.5px;
		font-weight: 600;
	}
	.texto {
		margin-top: 2px;
		font-size: 13px;
		line-height: 1.45;
		color: var(--mu);
	}
	/* En mobile la acción baja debajo del texto, alineada con él. */
	@media (max-width: 520px) {
		.alerta {
			flex-wrap: wrap;
		}
		.cuerpo {
			flex-basis: calc(100% - 48px);
		}
		.alerta > :global(.btn) {
			margin-left: 48px;
		}
	}
</style>
