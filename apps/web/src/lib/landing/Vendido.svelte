<script lang="ts">
	import { anim } from '#lib/anim.ts';
	import { getLanding } from './estado.svelte.ts';
	import { fmt } from './formato.ts';

	/** Tarjeta "Vendido hoy" del hero, con el total que sube en cada tick y sus 12 barras. */
	let { fondo = '#2A140C', borde = 'none' }: { fondo?: string; borde?: string } = $props();
	const l = getLanding();
	const barras = $derived([30, 42, 36, 55, 48, 62, 58, 70, 66, 80, 76, 60 + (l.tick % 4) * 10]);
</script>

<div
	class="vendido"
	style:background={fondo}
	style:border={borde}
	use:anim={{ t: 'float', dur: 4800 }}
>
	<div class="et">Vendido hoy</div>
	<div class="total">{fmt(l.heroTotal)}</div>
	<div class="barras">
		{#each barras as h, i (i)}
			<span
				style:height="{h}%"
				style:background={i === barras.length - 1 ? '#E8875F' : 'rgba(251,247,242,.22)'}
			></span>
		{/each}
	</div>
</div>

<style>
	.vendido {
		width: 230px;
		padding: 16px 18px;
		border-radius: 20px;
		color: #fbf7f2;
		box-shadow: 0 30px 50px -20px rgba(42, 20, 12, 0.7);
		text-align: left;
		backdrop-filter: blur(10px);
	}
	.et {
		font-size: 12.5px;
		color: #d9c8ba;
	}
	.total {
		font-size: 30px;
		font-weight: 600;
		letter-spacing: -0.04em;
		margin-top: 2px;
	}
	.barras {
		margin-top: 10px;
		display: flex;
		gap: 3px;
		align-items: end;
		height: 30px;
	}
	.barras span {
		flex: 1;
		border-radius: 3px;
		transition: height 0.6s;
	}
</style>
