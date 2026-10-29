<script lang="ts">
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import { anim } from '#lib/anim.ts';
	import * as I from '#lib/icons.ts';
	import { LPLANS } from './datos.ts';
	import { fmt } from './formato.ts';

	/** Precios de la landing (handoff: #precios y #planes): los planes. */
	// Los CTA llevan al onboarding, que es el registro (handoff: Onboarding Mostrador.dc.html).
	const registro = resolve('/onboarding');

	let anual = $state(false);

	const planes = $derived(
		LPLANS.map(([n, d, p, f, cta, hi]) => {
			const yr = anual && p > 0;
			return {
				n,
				d,
				f,
				cta,
				hi: !!hi,
				price: p === 0 ? 'Gratis' : fmt(yr ? p * 0.75 : p),
				per: p === 0 ? 'para siempre' : '/ mes',
				note: yr ? fmt(p * 0.75 * 12) + ' por año' : p === 0 ? 'Sin tarjeta' : '14 días gratis'
			};
		})
	);
</script>

<section id="precios">
	<div class="curvas"></div>
	<div class="luz"></div>

	<div id="planes" class="planes-zona">
		<div class="p-cab" use:anim={'reveal'}>
			<div>
				<div class="etiq"><span></span>Precios</div>
				<h2>Más barato en cada escalón. Y con más cosas adentro.</h2>
			</div>
			<div class="ciclo" role="group" aria-label="Ciclo de facturación">
				<button type="button" aria-pressed={!anual} onclick={() => (anual = false)}>Mensual</button>
				<button type="button" aria-pressed={anual} onclick={() => (anual = true)}
					>Anual<span>−25 %</span></button
				>
			</div>
		</div>
		<div class="planes" use:anim={'reveal'}>
			{#each planes as p (p.n)}
				<div class="plan" class:hi={p.hi}>
					{#if p.hi}<span class="elegido">El más elegido</span>{/if}
					<div>
						<div class="pl-n">{p.n}</div>
						<div class="pl-d">{p.d}</div>
					</div>
					<div>
						<span class="pl-p">{p.price}</span> <span class="pl-per">{p.per}</span>
						<div class="pl-note">{p.note}</div>
					</div>
					<div class="pl-f">
						{#each p.f as f (f)}<div>
								<Icon icon={I.phCheckBold} size={15} style="flex:none;margin-top:2px" />{f}
							</div>{/each}
					</div>
					<a class="pl-cta" href={registro}>{p.cta}</a>
				</div>
			{/each}
		</div>
		<div class="empresa" use:anim={'reveal'}>
			<Icon icon={I.phBuildingsFill} size={24} style="color:#F5B796" />
			<div class="e-txt">
				<b>Empresa · a medida</b>
				<div>
					Más de $ 100 millones por mes, integraciones propias o acuerdos de nivel de servicio.
				</div>
			</div>
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- /contacto llega con las páginas públicas -->
			<a href="/contacto">Hablar con ventas</a>
		</div>
	</div>
</section>

<style>
	section {
		background: #2a140c;
		color: #fbf7f2;
		position: relative;
		overflow: hidden;
	}
	/* Las curvas del manual de marca, marrón sobre marrón. */
	.curvas {
		position: absolute;
		inset: 0;
		background: url('/textura/curvas-oscuras.svg') center / cover no-repeat;
	}
	.luz {
		position: absolute;
		inset: 0;
		background: radial-gradient(700px 500px at 100% 0%, rgba(210, 88, 47, 0.3), transparent 70%);
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
		margin: 14px 0 0;
		font-size: clamp(36px, 4.6vw, 60px);
		line-height: 1;
		letter-spacing: -0.04em;
		font-weight: 600;
		text-wrap: balance;
	}
	.planes-zona {
		position: relative;
		max-width: 1240px;
		margin: 0 auto;
		padding: 0 var(--lado) 120px;
		scroll-margin-top: 80px;
	}
	.p-cab {
		display: flex;
		justify-content: space-between;
		align-items: end;
		gap: 18px;
		flex-wrap: wrap;
	}
	h2 {
		margin: 0;
		font-size: clamp(30px, 3.6vw, 46px);
		line-height: 1.02;
		letter-spacing: -0.04em;
		font-weight: 600;
		max-width: 640px;
		text-wrap: balance;
	}
	.ciclo {
		display: flex;
		padding: 4px;
		border-radius: 999px;
		background: rgba(251, 247, 242, 0.08);
		border: 1px solid rgba(251, 247, 242, 0.14);
	}
	.ciclo button {
		border: 0;
		font-size: 14px;
		font-weight: 600;
		padding: 9px 16px;
		border-radius: 999px;
		background: transparent;
		color: #fbf7f2;
		white-space: nowrap;
	}
	.ciclo button[aria-pressed='true'] {
		background: #fbf7f2;
		color: #2a140c;
	}
	.ciclo span {
		margin-left: 6px;
		font-size: 12px;
		font-weight: 700;
		color: #7fd3a6;
	}
	.planes {
		margin-top: 28px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 240px), 1fr));
		gap: 12px;
	}
	.plan {
		position: relative;
		padding: 22px;
		border-radius: 26px;
		background: rgba(251, 247, 242, 0.05);
		color: #fbf7f2;
		border: 1.5px solid rgba(251, 247, 242, 0.12);
		display: flex;
		flex-direction: column;
		gap: 14px;
		box-shadow: 0 30px 60px -36px rgba(0, 0, 0, 0.7);
		--mu: #d9c8ba;
		--ic: #f5b796;
	}
	.plan.hi {
		background: #fbf7f2;
		color: #2a140c;
		border-color: #d2582f;
		--mu: #7a6252;
		--ic: #d2582f;
	}
	.elegido {
		position: absolute;
		top: -12px;
		left: 20px;
		padding: 5px 11px;
		border-radius: 999px;
		background: #b0441d;
		color: #fff;
		font-size: 12px;
		font-weight: 700;
	}
	.pl-n {
		font-size: 20px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.pl-d,
	.pl-per,
	.pl-note {
		font-size: 14px;
		color: var(--mu);
	}
	.pl-d {
		margin-top: 2px;
	}
	.pl-p {
		font-size: 36px;
		font-weight: 600;
		letter-spacing: -0.045em;
	}
	.pl-note {
		font-size: 12.5px;
		margin-top: 2px;
	}
	.pl-f {
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.pl-f div {
		display: flex;
		gap: 9px;
		font-size: 14px;
		line-height: 1.4;
	}
	.pl-f :global(svg) {
		color: var(--ic);
	}
	.pl-cta {
		margin-top: auto;
		display: block;
		text-align: center;
		padding: 13px;
		border-radius: 999px;
		color: #fff;
		font-size: 14.5px;
		font-weight: 600;
		border: 0;
		background: linear-gradient(180deg, rgba(255, 255, 255, 0.18), rgba(255, 255, 255, 0.06));
		box-shadow: var(--inflado-vidrio);
	}
	.hi .pl-cta {
		border: 0;
		background: linear-gradient(180deg, #ee7a4a 0%, #d2582f 55%, #bc4920 100%);
		box-shadow: var(--inflado-naranja);
	}
	.pl-cta:hover {
		color: #fff;
	}
	.empresa {
		margin-top: 12px;
		display: flex;
		align-items: center;
		gap: 14px;
		flex-wrap: wrap;
		padding: 16px 20px;
		border-radius: 22px;
		background: rgba(251, 247, 242, 0.05);
		border: 1px solid rgba(251, 247, 242, 0.12);
	}
	.e-txt {
		flex: 1;
		min-width: 220px;
	}
	.e-txt b {
		font-size: 15.5px;
	}
	.e-txt div {
		font-size: 14px;
		color: #d9c8ba;
	}
	.empresa a {
		padding: 11px 18px;
		border-radius: 999px;
		color: #2a140c;
		font-size: 14px;
		font-weight: 600;
		border: 0;
		background: linear-gradient(180deg, #ffffff 0%, #f5efe9 100%);
		box-shadow: var(--inflado-claro);
	}
</style>
