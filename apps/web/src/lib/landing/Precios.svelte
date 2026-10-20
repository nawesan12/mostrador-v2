<script lang="ts">
	import { resolve } from '$app/paths';
	import Icon from '#lib/components/Icon.svelte';
	import { anim } from '#lib/anim.ts';
	import * as I from '#lib/icons.ts';
	import Marca from './Marca.svelte';
	import { CMP, LPLANS } from './datos.ts';
	import { fmt } from './formato.ts';

	/** Precios de la landing (handoff: #precios y #planes): calculadora, planes y comparativa. */
	const registro = `${resolve('/ingresar')}?m=registro`;

	let ventas = $state(6_000_000);
	let anual = $state(false);
	// Tramo según lo que vende por mes (handoff): contra Esencial, Impulso o Escala de Tiendanube.
	const tramo = $derived(ventas < 1_500_000 ? 0 : ventas < 8_000_000 ? 1 : 2);
	const mostrador = $derived([19_900, 49_900, 149_900][tramo]);
	const otro = $derived([27_999, 79_999, 244_999][tramo]);

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

{#snippet valor(t: string)}
	{@const [a, b] = t.split(' · ')}
	<span class="val"
		>{a}{#if b}<span class="sep"> · </span><span class="pr">{b}</span>{/if}</span
	>
{/snippet}

<section id="precios">
	<div class="arcos"></div>
	<div class="luz"></div>
	<div class="calc-zona">
		<div class="calc-txt" use:anim={'reveal'}>
			<div class="etiq"><span></span>Precio</div>
			<h2>Cuanto más vendés, más ahorrás.</h2>
			<p>
				Cada plan de Mostrador cuesta menos que su equivalente en Tiendanube, y trae Hornero,
				factura de ARCA y punto de venta incluidos.
			</p>
			<div class="slider">
				<div class="s-cab">
					<label for="ventas">Vendo por mes</label><span class="s-v">{fmt(ventas)}</span>
				</div>
				<input
					id="ventas"
					type="range"
					min="500000"
					max="30000000"
					step="100000"
					bind:value={ventas}
				/>
				<div class="s-pie"><span>$ 500 mil</span><span>$ 30 millones</span></div>
			</div>
		</div>
		<div class="calc" use:anim={{ t: 'reveal', delay: 120 }}>
			<div class="filas">
				<div>
					<div class="f-cab">
						<span class="f-n"><Marca alto={14} />Mostrador</span><span class="f-v"
							>{fmt(mostrador)} / mes</span
						>
					</div>
					<div class="barra">
						<div class="nuestra" style:width="{((mostrador / otro) * 100).toFixed(1)}%"></div>
					</div>
					<div class="f-pie">
						Plan {['Emprende', 'Pro', 'Negocio'][tramo]} · 0% de comisión de Mostrador
					</div>
				</div>
				<div>
					<div class="f-cab">
						<span class="f-o">Tiendanube</span><span class="f-o b">{fmt(otro)} / mes</span>
					</div>
					<div class="barra"><div class="ellos"></div></div>
					<div class="f-pie">
						Plan {['Esencial', 'Impulso', 'Escala'][tramo]} · precio publicado
					</div>
				</div>
			</div>
			<div class="ahorro">
				<div>
					<div class="a-l">Ahorrás por año</div>
					<div class="a-v">{fmt(Math.max(0, (otro - mostrador) * 12))}</div>
				</div>
				<a href="#planes">Ver planes</a>
			</div>
		</div>
	</div>

	<div id="planes" class="planes-zona">
		<div class="p-cab" use:anim={'reveal'}>
			<h3>Más barato en cada escalón. Y con más cosas adentro.</h3>
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
					<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- registro es /ingresar con query (resolve adentro) -->
					<a class="pl-cta" href={registro}>{p.cta}</a>
				</div>
			{/each}
		</div>
		<div class="notas" use:anim={'reveal'}>
			<div class="nota">
				<img src="/hornero/avatar.png" alt="" width="40" height="40" />
				<div>
					<b>¿Qué cuenta como respuesta con IA?</b>
					<div>
						Solo cuando Hornero redacta algo nuevo. Horarios, costo de envío, stock y estado del
						pedido se responden solos, sin IA y sin límite, en todos los planes.
					</div>
				</div>
			</div>
			<div class="nota">
				<Icon icon={I.phStackPlusFill} size={26} style="color:#F5B796;flex:none" />
				<div>
					<b>¿Necesitás más?</b>
					<div>
						2.000 respuestas con IA extra por $ 4.900, cuando quieras. Las campañas de WhatsApp se
						cobran al costo de Meta, sin recargo. Te avisamos al llegar al 80 %.
					</div>
				</div>
			</div>
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
		<div class="tabla" use:anim={'reveal'}>
			<div class="t-fila cab">
				<span class="t-col">Plan por plan</span><span class="t-m"
					><img src="/mark.svg" alt="" width="18" />Mostrador</span
				><span class="t-col">Tiendanube</span>
			</div>
			{#each CMP as [t, m, o, mOk, oOk] (t)}
				<div class="t-fila">
					<span class="t-t">{t}</span>
					<span class="t-m"
						><Icon
							icon={mOk ? I.phCheckCircleFill : I.phMinusCircle}
							size={16}
							style="color:{mOk ? '#7FD3A6' : '#B8A293'};flex:none"
						/>{@render valor(m)}</span
					>
					<span class="t-o"
						><Icon
							icon={oOk === 1 ? I.phCheckCircleFill : oOk === 2 ? I.phClock : I.phMinusCircle}
							size={16}
							style="color:{oOk === 1 ? '#D9C8BA' : oOk === 2 ? '#E3A93B' : '#8F7B6C'};flex:none"
						/>{@render valor(o)}</span
					>
				</div>
			{/each}
		</div>
		<div class="legal">
			Precios mensuales con IVA. Tiendanube según tiendanube.com/site/tienda-gratis, octubre de
			2026; “No figura en el plan” significa que no aparece entre las funciones que publica para ese
			plan. Verificá precios vigentes antes de publicar.
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
	.arcos {
		position: absolute;
		inset: 0;
		background-image: url('/textura/arch-pattern-light.svg');
	}
	.luz {
		position: absolute;
		inset: 0;
		background: radial-gradient(700px 500px at 100% 0%, rgba(210, 88, 47, 0.3), transparent 70%);
	}
	.calc-zona {
		position: relative;
		max-width: 1240px;
		margin: 0 auto;
		padding: 120px var(--lado);
		display: flex;
		gap: 56px;
		flex-wrap: wrap;
		align-items: center;
	}
	.calc-txt {
		flex: 1 1 420px;
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
	.calc-txt > p {
		margin: 20px 0 0;
		font-size: 17px;
		line-height: 1.55;
		color: #d9c8ba;
		max-width: 440px;
	}
	.slider {
		margin-top: 40px;
		padding: 22px;
		border-radius: 20px;
		background: rgba(251, 247, 242, 0.06);
		border: 1px solid rgba(251, 247, 242, 0.12);
	}
	.s-cab {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
	}
	.s-cab label {
		font-size: 14px;
		color: #d9c8ba;
	}
	.s-v {
		font-size: 28px;
		font-weight: 600;
		letter-spacing: -0.03em;
	}
	.slider input {
		width: 100%;
		margin-top: 14px;
	}
	.s-pie {
		display: flex;
		justify-content: space-between;
		font-size: 12px;
		color: #b8a293;
		margin-top: 6px;
	}
	.calc {
		flex: 1 1 440px;
		border-radius: 26px;
		background: #fbf7f2;
		color: #2a140c;
		padding: 32px;
		box-shadow: 0 40px 80px -30px rgba(0, 0, 0, 0.6);
	}
	.filas {
		display: flex;
		flex-direction: column;
		gap: 22px;
	}
	.f-cab {
		display: flex;
		justify-content: space-between;
		font-size: 15px;
	}
	.f-n {
		font-weight: 600;
		display: flex;
		align-items: center;
		gap: 8px;
	}
	.f-v {
		font-weight: 600;
	}
	.f-o {
		color: #5c4536;
	}
	.f-o.b {
		font-weight: 600;
	}
	.barra {
		margin-top: 10px;
		height: 14px;
		border-radius: 8px;
		background: #efe6da;
		overflow: hidden;
	}
	.nuestra {
		height: 100%;
		border-radius: 8px;
		background: linear-gradient(90deg, #d2582f, #e8875f);
		transition: width 0.5s cubic-bezier(0.2, 0.7, 0.1, 1);
	}
	.ellos {
		height: 100%;
		width: 100%;
		border-radius: 8px;
		background: repeating-linear-gradient(135deg, #c9b6a6 0 6px, #bca796 6px 12px);
	}
	.f-pie {
		margin-top: 6px;
		font-size: 13px;
		color: #7a6252;
	}
	.ahorro {
		margin-top: 28px;
		padding-top: 24px;
		border-top: 1px dashed rgba(58, 29, 18, 0.2);
		display: flex;
		justify-content: space-between;
		align-items: end;
		gap: 16px;
		flex-wrap: wrap;
	}
	.a-l {
		font-size: 14px;
		color: #7a6252;
	}
	.a-v {
		font-size: 46px;
		font-weight: 600;
		letter-spacing: -0.045em;
		color: #c24e25;
		margin-top: 2px;
	}
	.ahorro a {
		color: #fff;
		background: #2a140c;
		font-weight: 600;
		font-size: 15px;
		padding: 13px 20px;
		border-radius: 999px;
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
	h3 {
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
		background: #d2582f;
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
		background: rgba(251, 247, 242, 0.1);
		color: #fff;
		font-size: 14.5px;
		font-weight: 600;
	}
	.hi .pl-cta {
		background: linear-gradient(180deg, #e06a3e, #d2582f);
	}
	.pl-cta:hover {
		color: #fff;
	}
	.notas {
		margin-top: 12px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(min(100%, 280px), 1fr));
		gap: 12px;
	}
	.nota,
	.empresa {
		display: flex;
		gap: 12px;
		padding: 16px 18px;
		border-radius: 22px;
		background: rgba(251, 247, 242, 0.05);
		border: 1px solid rgba(251, 247, 242, 0.12);
	}
	.nota img {
		border-radius: 50%;
		object-fit: cover;
		flex: none;
		background: #fcede5;
	}
	.nota b {
		display: block;
		font-size: 15px;
		text-wrap: balance;
	}
	.nota div div {
		font-size: 13.5px;
		line-height: 1.5;
		color: #d9c8ba;
		margin-top: 3px;
	}
	.empresa {
		margin-top: 12px;
		align-items: center;
		gap: 14px;
		flex-wrap: wrap;
		padding: 16px 20px;
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
		background: #fbf7f2;
		color: #2a140c;
		font-size: 14px;
		font-weight: 600;
	}
	.tabla {
		margin-top: 44px;
		border-radius: 26px;
		background: rgba(251, 247, 242, 0.04);
		border: 1px solid rgba(251, 247, 242, 0.12);
		overflow: hidden;
	}
	.t-fila {
		display: grid;
		grid-template-columns: minmax(0, 1.4fr) minmax(0, 1fr) minmax(0, 1fr);
		gap: 12px;
		padding: 13px 20px;
		border-top: 1px solid rgba(251, 247, 242, 0.07);
		font-size: 14px;
		align-items: center;
	}
	.t-fila.cab {
		padding: 16px 20px;
		border-top: 0;
		border-bottom: 1px solid rgba(251, 247, 242, 0.12);
		font-weight: 600;
	}
	.t-col {
		color: #d9c8ba;
	}
	.t-t {
		color: #f4eae0;
	}
	.t-m {
		display: flex;
		align-items: center;
		gap: 7px;
		font-weight: 600;
		min-width: 0;
	}
	.cab .t-m {
		gap: 8px;
	}
	.t-o {
		display: flex;
		align-items: center;
		gap: 7px;
		color: #d9c8ba;
		min-width: 0;
	}
	.legal {
		margin-top: 10px;
		font-size: 12.5px;
		color: #b8a293;
	}
	/* En mobile, tres columnas no entran: el concepto va arriba y abajo Mostrador contra Tiendanube. */
	@media (max-width: 640px) {
		.t-fila {
			grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
			gap: 6px 12px;
			padding: 14px 16px;
		}
		.t-fila .t-t {
			grid-column: 1 / -1;
			font-size: 13px;
			color: #d9c8ba;
		}
		.t-fila.cab {
			padding: 14px 16px;
		}
		.t-fila.cab > .t-col:first-child {
			display: none;
		}
		.t-m,
		.t-o {
			align-items: flex-start;
			font-size: 14px;
			line-height: 1.35;
		}
		.sep {
			display: none;
		}
		.pr {
			display: block;
			font-weight: 500;
			opacity: 0.8;
			white-space: nowrap;
		}
		.t-m :global(svg),
		.t-o :global(svg) {
			margin-top: 1px;
		}
	}
</style>
