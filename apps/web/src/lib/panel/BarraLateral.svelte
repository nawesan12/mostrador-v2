<script lang="ts">
	import { afterNavigate } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import MarcaIcono from '#lib/components/MarcaIcono.svelte';
	import * as I from '#lib/icons.ts';
	import type { components } from '@mostrador/contrato/esquema';
	import { AREAS, SIMPLE, rutaDe, type Area } from './navegacion.ts';
	import { getShell } from './shell.svelte.ts';

	/**
	 * Barra lateral (handoff): la tienda, buscar, las secciones del área (o "Día a día" en Simple),
	 * los canales conectados, la prueba y el cambio Simple/Completo. Fija desde 1100 px; más angosto,
	 * se abre encima.
	 */
	let {
		area,
		seccion,
		tienda,
		foto,
		dominio,
		resumen
	}: {
		area: Area;
		seccion: string;
		tienda: string;
		foto: string | null;
		dominio: string;
		resumen: components['schemas']['ResumenPanel'];
	} = $props();
	const shell = getShell();
	const simple = $derived(shell.modo === 'simple');
	// En el celular no hay rail: arriba del menú van las áreas, y tocar una muestra sus secciones.
	let vista = $state<string | null>(null);
	const mostrada = $derived(AREAS.find((a) => a.id === vista) ?? area);
	const items = $derived(simple ? SIMPLE : mostrada.items);
	// Cualquier navegación (una sección, el "+", Elegir plan…) cierra el menú.
	afterNavigate(() => {
		shell.barra = false;
		vista = null;
	});
	const activo = (label: string) =>
		label === seccion ||
		(label === 'Inicio' && seccion === 'Resumen') ||
		(label === 'Cobros' && seccion === 'Finanzas');
	const NOMBRES: Record<string, string> = {
		wa: 'WhatsApp',
		ml: 'Mercado Libre',
		mp: 'Mercado Pago',
		ig: 'Instagram',
		tg: 'Telegram'
	};
	const dias = $derived(
		Array.from(
			{ length: resumen.prueba.diasTotales },
			(_, i) => i < resumen.prueba.diasTotales - resumen.prueba.diasRestantes
		)
	);
</script>

<!-- eslint-disable svelte/no-navigation-without-resolve -- rutas del panel armadas con rutaDe -->
{#if shell.barra}
	<div class="fondo" onclick={() => (shell.barra = false)} aria-hidden="true"></div>
{/if}
<aside class="barra" class:abierta={shell.barra} aria-label="Secciones">
	<div class="arriba">
		<div class="tienda">
			{#if foto}<img src={foto} alt="" width="34" height="34" />{/if}
			<span class="t-txt"
				><span class="t-n">{tienda}</span><span class="t-d"
					><span class="vivo"></span>{dominio}</span
				></span
			>
			<Icon icon={I.phCaretUpDown} size={14} style="color:var(--mu)" />
		</div>
		<button
			type="button"
			class="buscar"
			onclick={() => ((shell.cmdk = true), (shell.barra = false))}
		>
			<Icon icon={I.phMagnifyingGlass} size={15} /><span>Buscar o preguntar…</span><kbd>⌘K</kbd>
		</button>
	</div>
	<div class="medio">
		{#if !simple}
			<div class="areas" role="group" aria-label="Áreas">
				{#each AREAS as a (a.id)}
					{@const on = mostrada.id === a.id}
					<button type="button" class:on aria-pressed={on} onclick={() => (vista = a.id)}>
						<Icon icon={on ? a.iconFill : a.icon} size={16} />{a.label}
					</button>
				{/each}
			</div>
		{/if}
		<div class="area">{simple ? 'Día a día' : mostrada.label}</div>
		<nav class="items">
			{#each items as i (i.label)}
				{@const on = activo(i.label)}
				{@const n = i.insignia ? resumen.insignias[i.insignia] : 0}
				<a
					href={rutaDe(i.label)}
					class:on
					aria-current={on ? 'page' : undefined}
					onclick={() => (shell.barra = false)}
				>
					<Icon icon={on ? i.iconFill : i.icon} size={18} />
					<span class="i-l">{i.label}</span>
					{#if n}<span class="badge">{n}</span>{/if}
				</a>
			{/each}
		</nav>
		{#if resumen.canales.length}
			<div class="c-t">Canales conectados</div>
			<div class="canales">
				{#each resumen.canales as c (c)}<span title={NOMBRES[c] ?? c} class="canal"
						><MarcaIcono
							k={c}
							size={34}
							radius={10}
							fondo={c === 'ml' ? undefined : 'var(--sf)'}
						/></span
					>{/each}
				<a class="mas" href={rutaDe('Integraciones')} aria-label="Conectar otro canal"
					><Icon icon={I.phPlus} size={15} /></a
				>
			</div>
		{/if}
	</div>
	<div class="abajo">
		{#if resumen.prueba.diasRestantes > 0}
			<div class="prueba">
				<div class="arcos"></div>
				<div class="p-in">
					<div class="p-plan">Plan {resumen.prueba.plan} · prueba</div>
					<div class="p-dias">
						Quedan {resumen.prueba.diasRestantes}
						{resumen.prueba.diasRestantes === 1 ? 'día' : 'días'}
					</div>
					<div class="p-barras">
						{#each dias as usado, i (i)}<span class:usado></span>{/each}
					</div>
					<a href={rutaDe('Plan y facturación')}>Elegir plan</a>
				</div>
			</div>
		{/if}
		<div class="modo" role="group" aria-label="Modo del panel">
			<button type="button" aria-pressed={simple} onclick={() => shell.cambiarModo('simple')}
				>Simple</button
			>
			<button type="button" aria-pressed={!simple} onclick={() => shell.cambiarModo('completo')}
				>Completo</button
			>
		</div>
	</div>
</aside>

<!-- eslint-enable svelte/no-navigation-without-resolve -->

<style>
	.fondo {
		position: fixed;
		inset: 0;
		z-index: 74;
		background: rgba(42, 20, 12, 0.4);
		backdrop-filter: blur(2px);
		animation: aparece 0.25s both;
	}
	@keyframes aparece {
		from {
			opacity: 0;
		}
	}
	.barra {
		width: 260px;
		flex: none;
		position: sticky;
		top: 0;
		height: 100vh;
		height: 100dvh;
		z-index: 75;
		background: var(--sf2);
		border-right: 1px solid var(--ln);
		display: flex;
		flex-direction: column;
		transition: transform 0.4s var(--ease);
	}
	.arriba {
		padding: 16px 14px 8px;
	}
	.tienda {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 8px;
		border-radius: 14px;
		border: 1px solid var(--ln);
		background: var(--sf);
		box-shadow: var(--sh);
	}
	.tienda img {
		border-radius: 10px;
		object-fit: cover;
		flex: none;
	}
	.t-txt {
		flex: 1;
		min-width: 0;
	}
	.t-n {
		display: block;
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.t-d {
		display: flex;
		align-items: center;
		gap: 5px;
		font-size: 12px;
		color: var(--mu);
		white-space: nowrap;
		overflow: hidden;
	}
	.vivo {
		width: 6px;
		height: 6px;
		border-radius: 50%;
		background: #2f8f63;
		flex: none;
	}
	.buscar {
		margin-top: 10px;
		width: 100%;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 9px 12px;
		border-radius: 999px;
		border: 1px solid var(--ln);
		background: var(--sf);
		color: var(--mu);
		font-size: 13.5px;
	}
	.buscar span {
		flex: 1;
		text-align: left;
	}
	kbd {
		font-family: inherit;
		font-size: 11.5px;
		font-weight: 600;
		padding: 2px 7px;
		border-radius: 6px;
		background: var(--chip);
	}
	.medio {
		padding: 10px 12px;
		flex: 1;
		overflow: auto;
	}
	.areas {
		display: none;
	}
	.area {
		padding: 6px 12px 10px;
		font-size: 22px;
		font-weight: 600;
		letter-spacing: -0.03em;
	}
	.items {
		display: flex;
		flex-direction: column;
		gap: 2px;
	}
	.items a {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 9px 12px;
		border-radius: 12px;
		color: var(--mu);
		font-size: 14px;
		font-weight: 500;
	}
	.items a:hover {
		color: var(--tx);
		background: var(--chip);
	}
	.items a.on {
		background: var(--sf);
		color: var(--tx);
		font-weight: 600;
		box-shadow: var(--sh);
	}
	.items a.on :global(svg) {
		color: #d2582f;
	}
	.i-l {
		flex: 1;
	}
	.badge {
		font-size: 11.5px;
		font-weight: 700;
		padding: 1px 8px;
		border-radius: 999px;
		background: rgba(210, 88, 47, 0.13);
		color: #c24e25;
	}
	.on .badge {
		background: #d2582f;
		color: #fff;
	}
	.c-t {
		margin: 22px 12px 8px;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--fa);
	}
	.canales {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		padding: 0 12px;
	}
	.canal {
		border-radius: 10px;
		box-shadow: var(--sh);
	}
	.mas {
		width: 34px;
		height: 34px;
		border-radius: 10px;
		border: 1px dashed var(--ln2);
		display: grid;
		place-items: center;
		color: var(--mu);
	}
	.abajo {
		padding: 14px;
	}
	.prueba {
		position: relative;
		overflow: hidden;
		border-radius: 18px;
		padding: 16px;
		background: linear-gradient(140deg, #d2582f, #b9481f);
		color: #fff;
		box-shadow: 0 14px 30px -14px rgba(210, 88, 47, 0.6);
	}
	.arcos {
		position: absolute;
		inset: 0;
		background-image: url('/textura/arch-pattern-light.svg');
	}
	.p-in {
		position: relative;
	}
	.p-plan {
		font-size: 12.5px;
		font-weight: 600;
		color: #fce3d6;
	}
	.p-dias {
		margin-top: 4px;
		font-size: 15px;
		font-weight: 600;
	}
	.p-barras {
		margin-top: 10px;
		display: flex;
		gap: 3px;
	}
	.p-barras span {
		flex: 1;
		height: 5px;
		border-radius: 2px;
		background: rgba(255, 255, 255, 0.28);
	}
	.p-barras span.usado {
		background: #fff;
	}
	.prueba a {
		margin-top: 12px;
		display: block;
		text-align: center;
		font-size: 13px;
		font-weight: 600;
		color: #c24e25;
		background: #fff;
		padding: 8px;
		border-radius: 999px;
	}
	.modo {
		margin-top: 10px;
		display: flex;
		padding: 3px;
		border-radius: 999px;
		background: var(--chip);
	}
	.modo button {
		flex: 1;
		border: 0;
		font-size: 12.5px;
		font-weight: 600;
		padding: 7px;
		border-radius: 999px;
		background: transparent;
		color: var(--tx);
	}
	.modo button[aria-pressed='true'] {
		background: var(--sf);
		box-shadow: var(--sh);
	}
	/* Menos de 1100 px: se abre encima, a la derecha del rail. */
	@media (max-width: 1099px) {
		.barra {
			position: fixed;
			left: 76px;
			width: 280px;
			transform: translateX(-115%);
		}
		.barra.abierta {
			transform: none;
			box-shadow: 30px 0 60px -20px rgba(42, 20, 12, 0.5);
		}
	}
	/* Celular: sin rail, desde el borde, con las áreas arriba (las piezas del rail en chips). */
	@media (max-width: 759px) {
		.barra {
			left: 0;
			width: min(320px, 86vw);
		}
		.areas {
			display: flex;
			flex-wrap: wrap;
			gap: 6px;
			padding: 4px 8px 14px;
		}
		.areas button {
			display: inline-flex;
			align-items: center;
			gap: 6px;
			border: 0;
			cursor: pointer;
			padding: 8px 12px;
			border-radius: 999px;
			background: var(--sf);
			color: var(--tx);
			box-shadow: var(--sh);
			font:
				600 13px 'Instrument Sans',
				sans-serif;
		}
		.areas button.on {
			background: #2a140c;
			color: #fff;
		}
	}
	@media (min-width: 1100px) {
		.fondo {
			display: none;
		}
	}
</style>
