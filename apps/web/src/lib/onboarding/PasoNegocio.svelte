<script lang="ts">
	import Icon from '#lib/components/Icon.svelte';
	import * as I from '#lib/icons.ts';
	import { DATOS_NEGOCIO, HALLAZGOS, OBJETIVOS } from './datos.ts';
	import Etiqueta from './Etiqueta.svelte';
	import { getOnboarding } from './estado.svelte.ts';
	import Logo from './Logo.svelte';

	/** Paso 2, Tu negocio: lo que se encontró en ARCA, Instagram y Maps, el rubro y los objetivos. */
	const o = getOnboarding();
	const COLORES = ['#6B3A22', '#E9D8C4', '#2F5D50'];

	function aparece(el: HTMLElement) {
		el.animate(
			[
				{ opacity: 0, transform: 'translateY(6px)' },
				{ opacity: 1, transform: 'none' }
			],
			{ duration: 380, easing: 'cubic-bezier(.2,.7,.1,1)', fill: 'both' }
		);
	}
</script>

<Etiqueta rotulo="Tu negocio" titulo="Encontramos tu negocio. Revisá que esté todo bien." />
<div class="datos">
	{#each DATOS_NEGOCIO as [k, v] (k)}
		<div>
			<div class="d-k">{k}</div>
			<div class="d-v">{v}</div>
		</div>
	{/each}
</div>
<div class="hallazgos">
	{#each HALLAZGOS as h, i (h.src)}
		{@const listo = o.revelado > i}
		<div class="hallazgo">
			<div class="h-cab">
				<Logo k={h.k} tam={32} icono={18} radio={10} />
				<span class="h-src">{h.src}</span>
				{#if listo}
					<span class="h-ok"><Icon icon={I.phCheckCircleFill} size={14} />Importado</span>
				{:else}
					<span class="h-busca"
						><Icon icon={I.phCircleNotchBold} size={14} class="gira" />Buscando</span
					>
				{/if}
			</div>
			<div class="h-filas">
				{#each h.rows as [k, v] (k)}
					{#if listo}
						<div use:aparece>
							<div class="f-k">{k}</div>
							<div class="f-v">{v}</div>
						</div>
					{:else}
						<div>
							<div class="f-ph1"></div>
							<div class="f-ph2 brillo"></div>
						</div>
					{/if}
				{/each}
				{#if 'colores' in h && listo}
					<div class="colores" use:aparece>
						<span class="c-t">Colores</span>
						{#each COLORES as c (c)}<span class="c" style:background={c}></span>{/each}
						<img
							src="https://images.unsplash.com/photo-1556745753-b2904692b3cd?auto=format&fit=crop&w=80&h=80&q=70"
							alt=""
							width="22"
							height="22"
						/>
					</div>
				{/if}
			</div>
		</div>
	{/each}
</div>
<div class="rubro">
	<span class="r-t">Rubro detectado</span><span class="r-1"
		>Cafetería y almacén de especialidad</span
	><span class="r-2">Yerba mate y accesorios</span><span class="r-mas">+ Agregar</span>
</div>
<div class="objetivos">
	<div class="o-t">¿Qué querés resolver primero?</div>
	<div class="o-d">Ordenamos tu panel y tus primeras tareas según esto. Elegí las que quieras.</div>
	<div class="o-g">
		{#each OBJETIVOS as [ico, t], i (t)}
			{@const on = o.e.objetivos.includes(i)}
			<button type="button" class:on aria-pressed={on} onclick={() => o.alternarObjetivo(i)}>
				<span class="o-ico"><Icon icon={ico} size={18} /></span>
				<span class="o-x">{t}</span>
				{#if on}<Icon icon={I.phCheckCircleFill} size={18} style="color:#D2582F;flex:none" />{/if}
			</button>
		{/each}
	</div>
</div>

<style>
	.datos {
		margin-top: 22px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
		gap: 10px;
	}
	.datos > div {
		padding: 12px 14px;
		border-radius: 14px;
		background: #fff;
		border: 1px solid rgba(58, 29, 18, 0.12);
		min-width: 0;
	}
	.d-k {
		font-size: 12px;
		color: #8a7060;
	}
	.d-v {
		font-size: 15.5px;
		font-weight: 600;
		margin-top: 2px;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.hallazgos {
		margin-top: 18px;
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(260px, 1fr));
		gap: 14px;
	}
	.hallazgo {
		border-radius: 22px;
		background: #fff;
		box-shadow:
			0 1px 2px rgba(58, 29, 18, 0.06),
			0 20px 44px -28px rgba(58, 29, 18, 0.45);
		overflow: hidden;
	}
	.h-cab {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 14px 16px;
		border-bottom: 1px solid rgba(58, 29, 18, 0.07);
	}
	.h-src {
		flex: 1;
		font-size: 14.5px;
		font-weight: 600;
	}
	.h-ok,
	.h-busca {
		display: flex;
		align-items: center;
		gap: 4px;
		font-size: 12px;
		font-weight: 600;
		color: #2f8f63;
	}
	.h-busca {
		gap: 6px;
		color: #c24e25;
	}
	.h-busca :global(.gira) {
		animation: gira 0.9s linear infinite;
	}
	@keyframes gira {
		to {
			transform: rotate(360deg);
		}
	}
	.h-filas {
		padding: 12px 16px 16px;
		display: flex;
		flex-direction: column;
		gap: 10px;
	}
	.f-k {
		font-size: 11.5px;
		color: #8a7060;
	}
	.f-v {
		font-size: 14px;
		font-weight: 500;
		margin-top: 1px;
		line-height: 1.35;
	}
	.f-ph1 {
		width: 40%;
		height: 9px;
		border-radius: 5px;
		background: #f1e7dc;
	}
	.f-ph2 {
		width: 85%;
		height: 13px;
		border-radius: 6px;
		margin-top: 6px;
	}
	.brillo {
		background: linear-gradient(90deg, #f1e7dc, #fbf4ec, #f1e7dc);
		background-size: 200% 100%;
		animation: brillo 1.2s linear infinite;
	}
	@keyframes brillo {
		from {
			background-position: 100% 0;
		}
		to {
			background-position: -100% 0;
		}
	}
	@media (prefers-reduced-motion: reduce) {
		.brillo,
		.h-busca :global(.gira) {
			animation: none;
		}
	}
	.colores {
		display: flex;
		gap: 6px;
		align-items: center;
		margin-top: 2px;
	}
	.c-t {
		font-size: 11.5px;
		color: #8a7060;
		margin-right: 4px;
	}
	.c {
		width: 22px;
		height: 22px;
		border-radius: 7px;
	}
	.colores img {
		border-radius: 7px;
		object-fit: cover;
		margin-left: auto;
	}
	.rubro {
		margin-top: 18px;
		display: flex;
		align-items: center;
		gap: 10px;
		flex-wrap: wrap;
		font-size: 13.5px;
	}
	.r-t {
		color: #7a6252;
	}
	.r-1,
	.r-2,
	.r-mas {
		padding: 7px 12px;
		border-radius: 999px;
	}
	.r-1 {
		font-weight: 600;
		background: #2a140c;
		color: #fff;
	}
	.r-2 {
		font-weight: 500;
		background: #f4ece2;
		color: #5c4536;
	}
	.r-mas {
		font-weight: 500;
		border: 1px dashed rgba(58, 29, 18, 0.25);
		color: #7a6252;
	}
	.objetivos {
		margin-top: 26px;
		padding: 18px;
		border-radius: 22px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
	}
	.o-t {
		font-size: 16px;
		font-weight: 600;
	}
	.o-d {
		font-size: 13.5px;
		color: #7a6252;
		margin-top: 2px;
	}
	.o-g {
		margin-top: 14px;
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
		gap: 8px;
	}
	.o-g button {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 12px;
		border-radius: 16px;
		border: 1.5px solid rgba(58, 29, 18, 0.12);
		background: #fff;
		color: #2a140c;
		text-align: left;
		transition:
			border-color 0.2s,
			background 0.2s;
	}
	.o-g button.on {
		border-color: #d2582f;
		background: #fcede5;
	}
	.o-ico {
		width: 34px;
		height: 34px;
		flex: none;
		border-radius: 10px;
		display: grid;
		place-items: center;
		background: #f4ece2;
		color: #7a6252;
	}
	.on .o-ico {
		background: #d2582f;
		color: #fff;
	}
	.o-x {
		flex: 1;
		font-size: 13.5px;
		font-weight: 600;
		line-height: 1.25;
	}
</style>
