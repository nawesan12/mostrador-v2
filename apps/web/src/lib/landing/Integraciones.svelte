<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import { anim } from '#lib/anim.ts';
	import { CATS, INTEG } from './datos.ts';

	/** "Conectado con todo el comercio argentino." (handoff: #integraciones). */
	let cat = $state<string>('todas');
	const lista = $derived(INTEG.filter((i) => cat === 'todas' || i[1] === cat));
	const contar = (id: string) =>
		id === 'todas' ? INTEG.length : INTEG.filter((i) => i[1] === id).length;
</script>

<section id="integraciones">
	<div class="cab" use:anim={'reveal'}>
		<div class="tit">
			<div class="etiq"><span></span>Integraciones</div>
			<h2>Conectado con todo el comercio argentino.</h2>
		</div>
		<p>{INTEG.length} integraciones nativas, sin apps pagas aparte. Las activás con un clic.</p>
	</div>
	<div class="cats" use:anim={'reveal'} role="group" aria-label="Categoría">
		{#each CATS as [id, label] (id)}
			<button
				type="button"
				class:on={cat === id}
				aria-pressed={cat === id}
				onclick={() => (cat = id)}>{label}<span>{contar(id)}</span></button
			>
		{/each}
	</div>
	<div class="grilla">
		{#each lista as [n, , m, d] (n)}
			<div class="integ">
				<div class="logo">
					{#if 'img' in m}
						<img src={m.img} alt="" width="40" height="40" />
					{:else if 'icon' in m}
						<Icon icon={m.icon} size={36} style={'col' in m ? `color:${m.col}` : ''} />
					{:else if 'w' in m}
						<span
							class="palabra"
							style:background={m.bg ?? '#fff'}
							style:color={'fg' in m ? m.fg : undefined}>{m.w}</span
						>
					{/if}
				</div>
				<div>
					<div class="n">{n}</div>
					<div class="d">{d}</div>
				</div>
			</div>
		{/each}
	</div>
</section>

<style>
	section {
		max-width: 1240px;
		margin: 0 auto;
		padding: 120px var(--lado);
	}
	.cab {
		display: flex;
		justify-content: space-between;
		align-items: end;
		gap: 40px;
		flex-wrap: wrap;
	}
	.tit {
		max-width: 680px;
	}
	.etiq {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 600;
		color: #d2582f;
	}
	.etiq span {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: #d2582f;
	}
	h2 {
		margin: 14px 0 0;
		font-size: clamp(36px, 4.6vw, 60px);
		line-height: 1;
		letter-spacing: -0.04em;
		font-weight: 600;
		text-wrap: balance;
	}
	.cab p {
		max-width: 360px;
		margin: 0;
		font-size: 17px;
		line-height: 1.55;
		color: #5c4536;
	}
	.cats {
		margin-top: 40px;
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	.cats button {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		white-space: nowrap;
		border: 1px solid rgba(58, 29, 18, 0.14);
		font-size: 14px;
		font-weight: 600;
		padding: 9px 16px;
		border-radius: 999px;
		background: #fff;
		color: #2a140c;
	}
	.cats button.on {
		border-color: #2a140c;
		background: #2a140c;
		color: #fff;
	}
	.cats span {
		opacity: 0.6;
		font-weight: 500;
	}
	.grilla {
		margin-top: 24px;
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
		gap: 12px;
	}
	.integ {
		position: relative;
		padding: 18px;
		border-radius: 18px;
		background: #fff;
		border: 1px solid rgba(58, 29, 18, 0.07);
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.04);
		display: flex;
		flex-direction: column;
		gap: 16px;
		transition:
			transform 0.25s,
			box-shadow 0.25s;
	}
	.integ:hover {
		transform: translateY(-3px);
		box-shadow: 0 18px 36px -20px rgba(58, 29, 18, 0.4);
	}
	.logo {
		height: 40px;
		display: flex;
		align-items: center;
	}
	.logo img {
		object-fit: contain;
	}
	.palabra {
		height: 36px;
		padding: 0 10px;
		border-radius: 9px;
		display: grid;
		place-items: center;
		font-size: 14px;
		font-weight: 800;
		letter-spacing: -0.02em;
	}
	.n {
		font-size: 15px;
		font-weight: 600;
	}
	.d {
		font-size: 12.5px;
		color: #7a6252;
		margin-top: 2px;
	}
	/* En mobile, de a 3: si no, son 70 tarjetas de a una y el scroll no termina más. */
	@media (max-width: 640px) {
		.grilla {
			grid-template-columns: repeat(3, minmax(0, 1fr));
			gap: 8px;
		}
		.integ {
			padding: 12px 8px;
			gap: 8px;
			align-items: center;
			text-align: center;
			border-radius: 14px;
		}
		.integ:hover {
			transform: none;
		}
		.logo {
			height: 32px;
			justify-content: center;
		}
		.logo img {
			width: 30px;
			height: 30px;
		}
		.logo :global(svg) {
			width: 28px;
			height: 28px;
		}
		.palabra {
			height: 26px;
			padding: 0 6px;
			font-size: 11px;
			border-radius: 7px;
			max-width: 100%;
			overflow: hidden;
			white-space: nowrap;
		}
		.n {
			font-size: 12.5px;
			line-height: 1.2;
		}
		.d {
			display: none;
		}
	}
</style>
