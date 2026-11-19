<script lang="ts">
	import { resolve } from '$app/paths';
	import { page } from '$app/state';
	import { planDe } from '#lib/panel/planes.ts';
	import { hrefPlan } from '#lib/panel/tope.ts';

	/**
	 * Error dentro del panel (con el menú alrededor). El 403 es una sección que el rol de la persona
	 * no puede ver (docs/decisiones.md §37): se dice por qué, sin «algo salió mal».
	 */
	const sinPermiso = $derived(page.status === 403);
	// 402 tope_plan: la sección es de otro plan (Informes, por ejemplo). Se dice cuál, con el botón.
	const tope = $derived(page.status === 402 && page.error?.code === 'tope_plan');
	const sube = $derived(page.error?.plan ? planDe(page.error.plan) : null);
	// La sesión venció navegando el panel: a ingresar, que después vuelve acá.
	$effect(() => {
		if (page.status === 401)
			location.replace(
				`/ingresar?volver=${encodeURIComponent(page.url.pathname + page.url.search)}`
			);
	});
</script>

<svelte:head>
	<title>{tope ? 'Otro plan' : sinPermiso ? 'Sin permiso' : 'Algo salió mal'} · Panel</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<div class="error">
	<img src="/hornero/horno-vacio.webp" alt="" width="160" height="160" />
	{#if tope}
		<h1>Esto está en {sube ? `el plan ${sube.nombre}` : 'otro plan'}</h1>
		<p>{page.error?.message}</p>
		<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- hrefPlan ya pasa por resolve -->
		<a class="btn btn-primary" href={hrefPlan(page.error?.plan)}
			>{sube ? `Pasar a ${sube.nombre}` : 'Ver planes'}</a
		>
		<a class="btn" href={resolve('/panel')}>Volver al resumen</a>
	{:else}
		<h1>{sinPermiso ? 'Esta parte no es para tu rol' : 'No pudimos mostrar esta página'}</h1>
		<p>
			{sinPermiso
				? (page.error?.message ?? 'Pedíselo a la dueña de la tienda.')
				: 'Probá de nuevo en un rato. Si sigue, escribinos desde Ayuda.'}
		</p>
		<a class="btn btn-primary" href={resolve('/panel')}>Volver al resumen</a>
	{/if}
</div>

<style>
	.error {
		max-width: 460px;
		margin: 40px auto;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 12px;
		text-align: center;
	}
	h1 {
		margin: 0;
		font-size: 26px;
		letter-spacing: -0.03em;
		font-weight: 600;
		text-wrap: balance;
	}
	p {
		margin: 0;
		font-size: 15px;
		line-height: 1.5;
		color: var(--mu);
		text-wrap: pretty;
	}
</style>
