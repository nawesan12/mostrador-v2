<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { CONFLICTO, IMPORTACION, REGISTRO_HORNERO } from './datos.ts';
	import Etiqueta from './Etiqueta.svelte';
	import { getOnboarding } from './estado.svelte.ts';

	/** Paso 4, Importación: el avance por tipo de dato, lo que va ordenando Hornero y el conflicto de precios. */
	const o = getOnboarding();
	const fmt = (n: number) => Math.round(n).toLocaleString('es-AR');
	const filas = $derived(
		IMPORTACION.map(([ico, t, s, total, bg, fg], i) => {
			const inicio = i * 7;
			const p = Math.max(0, Math.min(1, (o.importado - inicio) / (100 - inicio - 4 + 0.001)));
			return { ico, t, s, bg, fg, v: fmt(total * p), listo: p >= 1 };
		})
	);
	const registro = $derived(
		REGISTRO_HORNERO.slice(0, Math.min(REGISTRO_HORNERO.length, Math.floor(o.importado / 19)))
	);
	const elegido = $derived(CONFLICTO.find(([k]) => k === o.e.conflicto));

	function entra(el: HTMLElement) {
		el.animate(
			[
				{ opacity: 0, transform: 'translateY(10px)' },
				{ opacity: 1, transform: 'none' }
			],
			{ duration: 420, easing: 'cubic-bezier(.2,.7,.1,1)' }
		);
	}
</script>

<div class="dos">
	<div class="izq">
		<Etiqueta
			rotulo="Importación"
			titulo={o.importacionLista
				? 'Listo. Tu negocio ya está adentro.'
				: 'Estamos armando tu mostrador.'}
		/>
		<div class="progreso">
			<div class="p-cab">
				<span>Progreso general</span><span class="pct">{Math.floor(o.importado)}%</span>
			</div>
			<div class="barra"><div style:width="{Math.floor(o.importado)}%"></div></div>
			<div class="filas">
				{#each filas as f (f.t)}
					<div class="fila">
						<span class="f-ico" style:background={f.bg} style:color={f.fg}
							><Icon icon={f.ico} size={17} /></span
						>
						<span class="f-txt"><span class="f-t">{f.t}</span><span class="f-s">{f.s}</span></span>
						<span class="f-v">{f.v}</span>
						{#if f.listo}
							<Icon icon={I.phCheckCircleFill} size={18} style="color:#2F8F63;flex:none" />
						{:else}
							<Icon
								icon={I.phCircleNotchBold}
								size={18}
								class="gira"
								style="color:#C24E25;flex:none"
							/>
						{/if}
					</div>
				{/each}
			</div>
		</div>
		{#if o.importacionLista}
			<div class="conflicto" use:entra>
				<div class="cf-cab">
					<span class="cf-ico"><Icon icon={I.phGitDiffFill} size={19} /></span>
					<div>
						<div class="cf-t">12 productos tienen precio distinto</div>
						<div class="cf-d">
							En Tiendanube el Termo 1 L sale $&nbsp;42.900 y en Mercado Libre $&nbsp;47.190. ¿Cuál
							usamos como precio base?
						</div>
					</div>
				</div>
				<div class="cf-ops">
					{#each CONFLICTO as [k, t] (k)}
						<button
							type="button"
							aria-pressed={o.e.conflicto === k}
							onclick={() => ((o.e.conflicto = k), o.guardar())}>{t}</button
						>
					{/each}
				</div>
				{#if elegido}
					<div class="cf-ok"><Icon icon={I.phCheckCircleFill} size={15} />{elegido[2]}</div>
				{/if}
			</div>
		{/if}
	</div>
	<div class="hornero">
		<div class="arcos"></div>
		<div class="h-cab">
			<span class="h-av"><img src="/hornero/caja.webp" alt="Hornero" width="52" height="52" /></span
			>
			<div>
				<div class="h-t">Hornero está ordenando</div>
				<div class="h-d">Mientras importamos, mejora lo que llega</div>
			</div>
		</div>
		<div class="h-log">
			{#each registro as l (l)}
				<div use:entra>
					<Icon
						icon={I.phSparkleFill}
						size={14}
						style="color:#F5B796;flex:none;margin-top:2px"
					/>{l}
				</div>
			{/each}
			{#if !o.importacionLista}
				<div class="puntos" aria-label="Hornero sigue trabajando">
					<span></span><span></span><span></span>
				</div>
			{/if}
		</div>
	</div>
</div>

<style>
	.dos {
		display: flex;
		gap: 28px;
		flex-wrap: wrap;
	}
	.izq {
		flex: 1 1 440px;
		min-width: 0;
	}
	.progreso {
		margin-top: 22px;
		padding: 20px;
		border-radius: 22px;
		background: #fff;
		box-shadow:
			0 1px 2px rgba(58, 29, 18, 0.06),
			0 24px 50px -30px rgba(58, 29, 18, 0.4);
	}
	.p-cab {
		display: flex;
		justify-content: space-between;
		font-size: 14px;
		font-weight: 600;
	}
	.pct {
		color: #c24e25;
	}
	.barra {
		margin-top: 10px;
		height: 10px;
		border-radius: 10px;
		background: #f1e7dc;
		overflow: hidden;
	}
	.barra div {
		height: 100%;
		border-radius: 10px;
		background: linear-gradient(90deg, #d2582f, #e8875f);
	}
	.filas {
		margin-top: 18px;
		display: flex;
		flex-direction: column;
	}
	.fila {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 10px 0;
		border-top: 1px solid rgba(58, 29, 18, 0.07);
	}
	.f-ico {
		width: 34px;
		height: 34px;
		flex: none;
		border-radius: 10px;
		display: grid;
		place-items: center;
	}
	.f-txt {
		flex: 1;
		min-width: 0;
	}
	.f-t {
		display: block;
		font-size: 14px;
		font-weight: 600;
	}
	.f-s {
		display: block;
		font-size: 12.5px;
		color: #7a6252;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.f-v {
		font-size: 15px;
		font-weight: 600;
		min-width: 64px;
		text-align: right;
	}
	.fila :global(.gira) {
		animation: gira 0.9s linear infinite;
	}
	@keyframes gira {
		to {
			transform: rotate(360deg);
		}
	}
	.conflicto {
		margin-top: 14px;
		padding: 18px;
		border-radius: 22px;
		background: #fff;
		border: 1.5px solid rgba(227, 169, 59, 0.5);
		box-shadow: 0 20px 44px -30px rgba(58, 29, 18, 0.45);
	}
	.cf-cab {
		display: flex;
		gap: 12px;
		align-items: flex-start;
	}
	.cf-ico {
		width: 38px;
		height: 38px;
		flex: none;
		border-radius: 12px;
		background: rgba(227, 169, 59, 0.2);
		color: #9a6a10;
		display: grid;
		place-items: center;
	}
	.cf-t {
		font-size: 15px;
		font-weight: 600;
	}
	.cf-d {
		font-size: 13px;
		color: #7a6252;
		margin-top: 2px;
		line-height: 1.45;
	}
	.cf-ops {
		margin-top: 12px;
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	.cf-ops button {
		padding: 9px 14px;
		border-radius: 999px;
		border: 1.5px solid rgba(58, 29, 18, 0.16);
		background: #fff;
		color: #2a140c;
		font-size: 13px;
		font-weight: 600;
	}
	.cf-ops button[aria-pressed='true'] {
		background: #2a140c;
		border-color: #2a140c;
		color: #fff;
	}
	.cf-ok {
		margin-top: 10px;
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 13px;
		font-weight: 600;
		color: #2f8f63;
	}
	.hornero {
		flex: 1 1 300px;
		align-self: flex-start;
		padding: 20px;
		border-radius: 22px;
		background: #2a140c;
		color: #fbf7f2;
		position: relative;
		overflow: hidden;
		box-shadow: 0 24px 50px -24px rgba(42, 20, 12, 0.6);
	}
	.arcos {
		position: absolute;
		inset: 0;
		background-image: url('/textura/arch-pattern-light.svg');
	}
	.h-cab {
		position: relative;
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.h-av {
		width: 52px;
		height: 52px;
		flex: none;
		border-radius: 50%;
		overflow: hidden;
		background: #fbf7f2;
		box-shadow: 0 0 0 3px #d2582f;
	}
	.h-av img {
		display: block;
		object-fit: contain;
	}
	.h-t {
		font-size: 15px;
		font-weight: 600;
	}
	.h-d {
		font-size: 12.5px;
		color: #d9c8ba;
	}
	.h-log {
		position: relative;
		margin-top: 16px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.h-log > div:not(.puntos) {
		display: flex;
		gap: 10px;
		padding: 10px 12px;
		border-radius: 12px;
		background: rgba(251, 247, 242, 0.07);
		font-size: 13px;
		line-height: 1.4;
	}
	.puntos {
		display: flex;
		gap: 4px;
		padding: 8px 12px;
	}
	.puntos span {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: #f5b796;
		animation: punto 1.1s infinite;
	}
	.puntos span:nth-child(2) {
		animation-delay: 0.18s;
	}
	.puntos span:nth-child(3) {
		animation-delay: 0.36s;
	}
	@keyframes punto {
		0%,
		100% {
			opacity: 0.25;
		}
		50% {
			opacity: 1;
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.puntos span,
		.fila :global(.gira) {
			animation: none;
		}
	}
</style>
