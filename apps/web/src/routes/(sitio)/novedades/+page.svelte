<script lang="ts">
	import { enhance } from '$app/forms';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import Etiqueta from '#lib/sitio/Etiqueta.svelte';
	import { NOVEDADES, TIPOS, type Tipo } from '#lib/sitio/novedades.ts';
	import type { PageProps } from './$types';

	/** Novedades (design/Mostrador Novedades.dc.html): filtro por tipo y suscripción mensual. */
	let { form }: PageProps = $props();
	let filtro = $state<Tipo | null>(null);
	const meses = $derived(
		NOVEDADES.map(
			([mes, items]) => [mes, items.filter((n) => !filtro || n.tipo === filtro)] as const
		).filter(([, items]) => items.length)
	);
	const err = $derived(form && 'errores' in form ? (form.errores as Record<string, string>) : {});
	const listo = $derived(!!form && 'enviado' in form);
	let enviando = $state(false);
	const conEnvio = () => {
		enviando = true;
		return async ({ update }: { update: (o?: { reset?: boolean }) => Promise<void> }) => {
			await update({ reset: false });
			enviando = false;
		};
	};
	const tipo = (t: Tipo) => TIPOS.find(([k]) => k === t)!;
</script>

<Seo
	title="Novedades · Mostrador"
	description="Cada semana mejoramos Mostrador con lo que nos piden las tiendas: lo nuevo, las mejoras y los arreglos."
	path="/novedades"
/>

<div class="ancho">
	<div class="in">
		<aside>
			<Etiqueta texto="Novedades" />
			<h1>Qué hay de nuevo</h1>
			<p class="bajada">Cada semana mejoramos Mostrador con lo que nos piden las tiendas.</p>
			<div class="filtros" role="group" aria-label="Filtrar por tipo">
				<button type="button" aria-pressed={filtro === null} onclick={() => (filtro = null)}
					>Todo</button
				>
				{#each TIPOS as [k, t] (k)}
					<button type="button" aria-pressed={filtro === k} onclick={() => (filtro = k)}>{t}</button
					>
				{/each}
			</div>
			<div class="suscribir">
				<b>Recibilas por email</b>
				<div class="s-d">Un resumen por mes, nada más.</div>
				{#if listo}
					<div class="s-ok" role="status">
						<Icon icon={I.phCheckCircleFill} size={16} />Listo, te sumamos.
					</div>
				{:else}
					<form method="POST" use:enhance={conEnvio} novalidate>
						<input
							name="email"
							type="email"
							autocomplete="email"
							placeholder="Tu email"
							aria-label="Tu email"
							aria-invalid={err.email ? 'true' : undefined}
							aria-describedby={err.email ? 'nov-err' : undefined}
						/>
						<button type="submit" disabled={enviando}>Sumarme</button>
					</form>
					{#if err.email || err.form}<small id="nov-err">{err.email ?? err.form}</small>{/if}
				{/if}
			</div>
			<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- página pública -->
			<a class="proponer" href="/contacto"
				><Icon icon={I.phLightbulbFill} size={20} style="color:#b0441d" />Proponé una mejora<Icon
					icon={I.phArrowRightBold}
					size={13}
					style="margin-left:auto"
				/></a
			>
		</aside>

		<div class="lista">
			{#each meses as [mes, items] (mes)}
				<div>
					<div class="mes">{mes}</div>
					<div class="items">
						{#each items as n (n.titulo)}
							{@const [, t, bg, fg] = tipo(n.tipo)}
							<article>
								{#if n.foto}
									<div class="foto">
										<img
											src="https://images.unsplash.com/photo-{n.foto}?auto=format&fit=crop&w=1200&h=514&q=72"
											alt=""
											loading="lazy"
										/>
									</div>
								{/if}
								<div class="a-in">
									<span class="tipo" style:background={bg} style:color={fg}>{t}</span>
									<h2>{n.titulo}</h2>
									<p>{n.texto}</p>
								</div>
							</article>
						{/each}
					</div>
				</div>
			{/each}
		</div>
	</div>
</div>

<style>
	.ancho {
		max-width: 1240px;
		margin: 0 auto;
		padding: 0 clamp(16px, 3vw, 28px);
	}
	.in {
		padding: clamp(40px, 6vw, 80px) 0 clamp(56px, 8vw, 100px);
		display: flex;
		gap: clamp(24px, 4vw, 56px);
		align-items: flex-start;
		flex-wrap: wrap;
	}
	aside {
		flex: 0 1 280px;
		min-width: 220px;
		position: sticky;
		top: 88px;
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	h1 {
		margin: 0;
		font-size: clamp(40px, 5vw, 64px);
		line-height: 0.98;
		letter-spacing: -0.045em;
		font-weight: 600;
	}
	.bajada {
		margin: 0;
		font-size: 15.5px;
		line-height: 1.55;
		color: #5c4536;
	}
	.filtros {
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
	}
	.filtros button {
		padding: 8px 13px;
		border: 0;
		border-radius: 999px;
		background: #fff;
		color: #2a140c;
		font-size: 13.5px;
		font-weight: 600;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.08);
	}
	.filtros button[aria-pressed='true'] {
		background: #2a140c;
		color: #fff;
	}
	.suscribir {
		padding: 18px;
		border-radius: 22px;
		background: #2a140c;
		color: #fbf7f2;
	}
	.suscribir b {
		font-size: 15px;
	}
	.s-d {
		font-size: 13.5px;
		color: #d9c8ba;
		margin-top: 4px;
	}
	form {
		margin-top: 12px;
		display: flex;
		padding: 4px;
		border-radius: 999px;
		background: #fff;
	}
	form input {
		flex: 1;
		min-width: 0;
		padding: 8px 12px;
		border: 0;
		background: transparent;
		font-size: 14px;
		font-weight: 500;
		color: #2a140c;
		outline: none;
	}
	form button {
		padding: 8px 14px;
		border: 0;
		border-radius: 999px;
		background: #b0441d;
		color: #fff;
		font-size: 13.5px;
		font-weight: 600;
	}
	small {
		display: block;
		margin-top: 8px;
		font-size: 12.5px;
		color: #f5b796;
	}
	.s-ok {
		margin-top: 12px;
		display: flex;
		align-items: center;
		gap: 8px;
		font-size: 14px;
		font-weight: 600;
		color: #7fd3a6;
	}
	.proponer {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 14px 16px;
		border-radius: 18px;
		background: #fcede5;
		color: #2a140c;
		font-size: 14px;
		font-weight: 600;
	}
	.lista {
		flex: 1 1 560px;
		min-width: 0;
		max-width: 820px;
		display: flex;
		flex-direction: column;
		gap: 36px;
	}
	.mes {
		position: sticky;
		top: 72px;
		z-index: 2;
		padding: 8px 0;
		background: #fbf7f2;
		font-size: 13px;
		font-weight: 700;
		letter-spacing: 0.12em;
		text-transform: uppercase;
		color: #7a6252;
	}
	.items {
		margin-top: 8px;
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	article {
		border-radius: 24px;
		background: #fff;
		box-shadow: 0 1px 2px rgba(58, 29, 18, 0.06);
		overflow: hidden;
	}
	.foto {
		aspect-ratio: 21 / 9;
		background: #e9ddcf;
	}
	.foto img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		display: block;
	}
	.a-in {
		padding: 20px 22px;
	}
	.tipo {
		display: inline-block;
		padding: 4px 10px;
		border-radius: 999px;
		font-size: 12px;
		font-weight: 700;
	}
	h2 {
		margin: 10px 0 0;
		font-size: 20px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	article p {
		margin: 6px 0 0;
		font-size: 15px;
		line-height: 1.6;
		color: #5c4536;
	}
	/* Celular: la columna de la izquierda no puede quedar fija arriba (taparía la lista) y el
	   rótulo del mes se pega debajo del header, que ahí es más alto. */
	@media (max-width: 760px) {
		aside {
			position: static;
			flex: 1 1 100%;
		}
		.mes {
			top: 96px;
		}
	}
</style>
