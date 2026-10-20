<script lang="ts">
	import MarcaIcono from '#lib/components/MarcaIcono.svelte';
	import { FEED } from './datos.ts';
	import { getLanding } from './estado.svelte.ts';
	import { fmt } from './formato.ts';

	/**
	 * Feed de ventas en vivo del hero: tres avisos que rotan sobre los 7 eventos del handoff (FEED),
	 * el de arriba entra animado en cada tick.
	 */
	let {
		origen = 'left top',
		grande = false,
		sombra = '0 18px 40px -16px rgba(58,29,18,.45),0 0 0 1px rgba(58,29,18,.06)'
	}: { origen?: string; grande?: boolean; sombra?: string } = $props();
	const l = getLanding();
	const OP = [1, 0.85, 0.6];
	const SC = [1, 0.97, 0.94];
	const items = $derived(
		[0, 1, 2].map((j) => {
			const i = l.tick + 2 - j;
			const [k, t, s, v] = FEED[i % FEED.length];
			return { i, k, t, s, v: fmt(v), op: OP[j], sc: SC[j] };
		})
	);
</script>

{#each items as f, j (f.i)}
	<div
		class="aviso"
		class:grande
		class:nuevo={j === 0}
		style:opacity={f.op}
		style:transform="scale({f.sc})"
		style:transform-origin={origen}
		style:box-shadow={sombra}
	>
		<MarcaIcono k={f.k} size={grande ? 36 : 34} radius={11} />
		<span class="txt"><span class="t">{f.t}</span><span class="s">{f.s}</span></span>
		<span class="v">{f.v}</span>
	</div>
{/each}

<style>
	.aviso {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 11px 13px;
		border-radius: 16px;
		background: rgba(255, 255, 255, 0.95);
		backdrop-filter: blur(10px);
		color: #2a140c;
		text-align: left;
		transition:
			opacity 0.5s,
			transform 0.5s;
	}
	.aviso.grande {
		padding: 12px 14px;
		background: rgba(255, 255, 255, 0.94);
	}
	.nuevo {
		animation: entra 520ms cubic-bezier(0.2, 0.7, 0.1, 1);
	}
	@keyframes entra {
		from {
			opacity: 0;
			transform: translateY(-14px) scale(0.96);
		}
	}
	.txt {
		flex: 1;
		min-width: 0;
	}
	.t {
		display: block;
		font-size: 13px;
		font-weight: 600;
	}
	.s {
		display: block;
		font-size: 11.5px;
		color: #7a6252;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.v {
		font-size: 13.5px;
		font-weight: 600;
	}
	.grande .t {
		font-size: 13.5px;
	}
	.grande .s {
		font-size: 12px;
	}
	.grande .v {
		font-size: 14px;
		letter-spacing: -0.02em;
	}
</style>
