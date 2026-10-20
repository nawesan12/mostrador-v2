<script lang="ts">
	import { anim } from '#lib/anim.ts';
	import LogoInteg from './LogoInteg.svelte';
	import { INTEG } from './datos.ts';

	/** "Se conecta con todo lo que ya usás": dos filas de integraciones que corren en sentidos opuestos. */
	const mitad = Math.ceil(INTEG.length / 2);
	const fila1 = [...INTEG.slice(0, mitad), ...INTEG.slice(0, mitad)];
	const fila2 = [...INTEG.slice(mitad), ...INTEG.slice(mitad)];
</script>

<section id="integraciones-strip" aria-label="Integraciones">
	<div class="tit">Se conecta con todo lo que ya usás</div>
	<div class="filas">
		<div class="fila" use:anim={{ t: 'marquee', dur: 60000 }}>
			{#each fila1 as [n, , m], i (i)}
				<span class="chip" aria-hidden={i >= mitad}
					><LogoInteg {m} /><span class="n">{n}</span></span
				>
			{/each}
		</div>
		<div class="fila" use:anim={{ t: 'marquee-r', dur: 64000 }}>
			{#each fila2 as [n, , m], i (i)}
				<span class="chip" aria-hidden={i >= fila2.length / 2}
					><LogoInteg {m} /><span class="n">{n}</span></span
				>
			{/each}
		</div>
	</div>
</section>

<style>
	section {
		border-top: 1px solid rgba(58, 29, 18, 0.08);
		border-bottom: 1px solid rgba(58, 29, 18, 0.08);
		background: #fff;
		padding: 30px 0;
		overflow: hidden;
	}
	.tit {
		text-align: center;
		font-size: 14px;
		font-weight: 600;
		color: #7a6252;
	}
	.filas {
		margin-top: 20px;
		display: flex;
		flex-direction: column;
		gap: 12px;
		mask-image: linear-gradient(90deg, transparent, #000 10%, #000 90%, transparent);
	}
	.fila {
		display: flex;
		gap: 12px;
		width: max-content;
	}
	.chip {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 18px 10px 12px;
		border-radius: 999px;
		background: #fbf7f2;
		border: 1px solid rgba(58, 29, 18, 0.08);
		white-space: nowrap;
	}
	.n {
		font-size: 14px;
		font-weight: 600;
	}
</style>
