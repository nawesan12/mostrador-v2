<script lang="ts">
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import TplMini from '#lib/components/TplMini.svelte';
	import { anim } from '#lib/anim.ts';
	import { phArrowRightBold, phSparkleFill } from '#lib/icons.ts';
	import { PLANTILLAS } from '#lib/plantillas.ts';
	import { getLanding } from './estado.svelte.ts';

	/** Sección "Plantillas" de la landing (handoff): rubros destacados, paletas y la vitrina. */
	const l = getLanding();
	// Los CTA llevan al onboarding, que es el registro (handoff: Onboarding Mostrador.dc.html).
	const registro = resolve('/onboarding');

	// Los primeros 4 rubros, con la primera plantilla de cada uno (handoff: tplFeat).
	const DESTACADAS = PLANTILLAS.filter(
		(t, i) => PLANTILLAS.findIndex((o) => o.rubro === t.rubro) === i
	).slice(0, 4);
	const PALETAS = [
		[null, 'Los de la plantilla'],
		[{ bg: '#FBF7F2', tx: '#2A140C', ac: '#D2582F' }, 'Terracota, sacado de tu Instagram'],
		[{ bg: '#0F1A14', tx: '#EDE7DA', ac: '#C9A227' }, 'Verde noche y dorado'],
		[{ bg: '#F4F1FA', tx: '#231942', ac: '#7B5CD6' }, 'Lavanda']
	] as const;

	let fi = $state(0);
	let pi = $state(0);
	let ancho = $state(1440);
	const t0 = $derived(DESTACADAS[fi]);
	const paleta = $derived(PALETAS[pi][0]);
	// Escala de la vitrina según el ancho (handoff: tplShowZoom).
	const zoom = $derived.by(() => {
		const pad = Math.min(56, Math.max(28, ancho * 0.05));
		const inner = Math.min(ancho, 1240) - 2 * pad - 16;
		const col = inner >= 808 ? (inner - 48) / 2 : inner;
		return Math.max(0.6, Math.min(1.65, (col - 8) / 340));
	});
</script>

<svelte:window bind:innerWidth={ancho} />

<section id="plantillas">
	<div class="arcos"></div>
	<div class="luz"></div>
	<div class="in">
		<div class="txt" use:anim={'reveal'}>
			<div>
				<div class="etiq"><span></span>Plantillas</div>
				<h2>Elegí un estilo. Hornero le pone tu marca.</h2>
			</div>
			<div class="controles">
				<p>
					{PLANTILLAS.length} plantillas por rubro. Toma los colores y el logo de tu Instagram, y se adapta
					sola al celular.
				</p>
				<div>
					<div class="sub">Probá un rubro</div>
					<div class="chips" role="group" aria-label="Probá un rubro">
						{#each DESTACADAS as t, i (t.id)}
							<button
								type="button"
								class:on={i === fi}
								aria-pressed={i === fi}
								onclick={() => (fi = i)}>{t.rubro}</button
							>
						{/each}
					</div>
				</div>
				<div>
					<div class="sub">Con tus colores</div>
					<div class="paletas">
						{#each PALETAS as [p, nombre], i (nombre)}
							{@const v = p ?? { bg: t0.bg, ac: t0.ac }}
							<button
								type="button"
								class="pal"
								class:on={i === pi}
								aria-label={nombre}
								aria-pressed={i === pi}
								title={nombre}
								onclick={() => (pi = i)}
							>
								<span style:background={v.bg}></span><span class="ac" style:background={v.ac}
								></span>
							</button>
						{/each}
						<span class="pal-n">{PALETAS[pi][1]}</span>
					</div>
				</div>
				<button type="button" class="ver" onclick={() => (l.tplI = 0)}
					>Ver las {PLANTILLAS.length} plantillas<Icon icon={phArrowRightBold} size={13} /></button
				>
			</div>
		</div>

		<div class="vitrina">
			<button type="button" class="tarjeta" onclick={() => (l.tplI = PLANTILLAS.indexOf(t0))}>
				<div class="marco" aria-hidden="true" style:zoom>
					<div class="barra">
						<span style:background="#E0A595"></span><span style:background="#E7CB8F"></span><span
							style:background="#A9CBA8"
						></span>
					</div>
					<div class="mini"><TplMini t={t0} {paleta} /></div>
				</div>
				<div class="pie">
					<span class="nom"
						><span class="n">{t0.nombre}</span> <span class="r">{t0.rubro}</span></span
					>
					<span class="muestras">
						<span style:background={paleta?.bg ?? t0.bg}></span><span
							style:background={paleta?.ac ?? t0.ac}
						></span><span style:background={paleta?.tx ?? t0.tx}></span>
					</span>
				</div>
			</button>
		</div>
		<div class="armar" use:anim={'reveal'}>
			<span>¿Ninguna te convence? Hornero arma una a medida con tu Instagram.</span>
			<a href={registro}
				><Icon icon={phSparkleFill} size={15} style="color:#D2582F" />Armar la mía</a
			>
		</div>
	</div>
</section>

<style>
	section {
		position: relative;
		background: #2a140c;
		color: #fbf7f2;
		overflow: hidden;
		border-radius: 36px;
		max-width: 1240px;
		width: calc(100% - 2 * var(--lado));
		margin: 40px auto 0;
	}
	.arcos {
		position: absolute;
		inset: 0;
		background-image: url('/textura/arch-pattern-light.svg');
	}
	.luz {
		position: absolute;
		inset: 0;
		background: radial-gradient(800px 500px at 0% 0%, rgba(210, 88, 47, 0.3), transparent 70%);
	}
	.in {
		position: relative;
		max-width: 1240px;
		margin: 0 auto;
		padding: clamp(20px, 5vw, 56px);
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 380px), 1fr));
		gap: 48px;
		align-items: center;
	}
	.txt {
		display: flex;
		flex-direction: column;
		gap: 22px;
	}
	.etiq {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 600;
		color: #f5b796;
	}
	.etiq span {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: #f5b796;
	}
	h2 {
		margin: 12px 0 0;
		font-size: clamp(32px, 3.6vw, 48px);
		line-height: 1.02;
		letter-spacing: -0.04em;
		font-weight: 600;
		text-wrap: balance;
	}
	.controles {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		gap: 22px;
	}
	.controles p {
		margin: 0;
		font-size: 17px;
		line-height: 1.55;
		color: #d9c8ba;
		max-width: 440px;
		text-wrap: pretty;
	}
	.sub {
		font-size: 12.5px;
		font-weight: 600;
		color: #c9b4a4;
		margin-bottom: 8px;
	}
	.chips {
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	.chips button {
		border: 1px solid rgba(251, 247, 242, 0.25);
		font-size: 14px;
		font-weight: 600;
		padding: 9px 15px;
		border-radius: 999px;
		background: transparent;
		color: #fbf7f2;
		white-space: nowrap;
		transition: background 0.2s;
	}
	.chips button.on {
		border-color: #fbf7f2;
		background: #fbf7f2;
		color: #2a140c;
	}
	.paletas {
		display: flex;
		gap: 10px;
		align-items: center;
	}
	.pal {
		display: flex;
		padding: 4px;
		border-radius: 999px;
		border: 2px solid transparent;
		background: transparent;
	}
	.pal.on {
		border-color: #fbf7f2;
	}
	.pal span {
		width: 20px;
		height: 20px;
		border-radius: 50%;
		box-shadow: inset 0 0 0 1px rgba(0, 0, 0, 0.15);
	}
	.pal .ac {
		margin-left: -6px;
		box-shadow: 0 0 0 2px #2a140c;
	}
	.pal-n {
		font-size: 12.5px;
		color: #c9b4a4;
	}
	.ver {
		display: flex;
		align-items: center;
		gap: 7px;
		padding: 11px 18px;
		border-radius: 999px;
		border: 1px solid rgba(251, 247, 242, 0.3);
		background: transparent;
		color: #fbf7f2;
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
	}
	.vitrina {
		display: flex;
		justify-content: center;
		min-width: 0;
	}
	.tarjeta {
		padding: 0;
		border: 0;
		text-align: left;
		font-family: inherit;
		color: #fbf7f2;
		background: transparent;
		display: flex;
		flex-direction: column;
		gap: 12px;
		transition: transform 0.35s var(--ease);
	}
	.tarjeta:hover {
		transform: translateY(-6px);
	}
	.marco {
		position: relative;
		width: 330px;
		height: 260px;
		border-radius: 14px;
		overflow: hidden;
		background: #fff;
		box-shadow:
			0 30px 60px -30px rgba(0, 0, 0, 0.7),
			0 0 0 1px rgba(251, 247, 242, 0.12);
	}
	.barra {
		height: 20px;
		display: flex;
		align-items: center;
		gap: 4px;
		padding: 0 9px;
		background: rgba(0, 0, 0, 0.06);
	}
	.barra span {
		width: 6px;
		height: 6px;
		border-radius: 50%;
	}
	.mini {
		height: 240px;
	}
	.pie {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 0 2px;
	}
	.nom {
		flex: 1;
		min-width: 0;
	}
	.nom .n {
		display: block;
		font-size: 16px;
		font-weight: 600;
	}
	.nom .r {
		display: block;
		font-size: 13px;
		color: #c9b4a4;
	}
	.muestras {
		display: flex;
	}
	.muestras span {
		width: 16px;
		height: 16px;
		border-radius: 50%;
		box-shadow: 0 0 0 2px #2a140c;
	}
	.muestras span + span {
		margin-left: -4px;
	}
	.armar {
		margin-top: 40px;
		display: flex;
		gap: 14px;
		align-items: center;
		justify-content: center;
		flex-wrap: wrap;
	}
	.armar span {
		font-size: 15px;
		color: #d9c8ba;
	}
	.armar a {
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 15px;
		font-weight: 600;
		color: #2a140c;
		background: #fbf7f2;
		padding: 12px 20px;
		border-radius: 999px;
		white-space: nowrap;
	}
</style>
