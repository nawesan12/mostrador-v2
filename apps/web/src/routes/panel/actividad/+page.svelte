<script lang="ts">
	import { untrack } from 'svelte';
	import { goto, invalidateAll } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import Vacio from '#lib/components/Vacio.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import {
		CATEGORIAS,
		consultaDe,
		cuando,
		nombreCategoria,
		type Evento,
		type FiltroActividad,
		type PaginaActividad
	} from '#lib/panel/actividad.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/**
	 * Actividad (design/PanelCrecer.dc.html, sAct; docs/decisiones.md §39): quién del equipo hizo qué,
	 * sobre qué y cuándo. Lo anota Go donde se hace cada acción; acá se lee, se filtra y se baja.
	 * Sólo la dueña y administración llegan hasta acá (Go responde 403 al resto).
	 */
	let { data } = $props();
	const ui = getUi();

	let extra = $state<Evento[]>([]);
	let siguiente = $state<string | null>(null);
	let cargando = $state(false);
	// Con filtros nuevos, la lista vuelve a la primera página. Si sólo se recargó (un registro nuevo
	// por tiempo real, §44), lo nuevo aparece arriba y lo que ya se había bajado con «Ver más» queda.
	let filtroVisto = '';
	$effect.pre(() => {
		const clave = JSON.stringify(data.filtro);
		const igual = clave === untrack(() => filtroVisto);
		filtroVisto = clave;
		if (igual && untrack(() => extra.length)) return;
		extra = [];
		siguiente = data.pagina.siguiente;
	});
	const items = $derived.by(() => {
		// eslint-disable-next-line svelte/prefer-svelte-reactivity -- local al cálculo
		const vistos = new Set<string>();
		return [...data.pagina.items, ...extra].filter((e) => !vistos.has(e.id) && !!vistos.add(e.id));
	});
	const filtro = $derived(data.filtro as FiltroActividad);
	const hayFiltros = $derived(Object.values(filtro).some(Boolean));

	function ir(q: string) {
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- misma página, sólo cambia la query
		void goto(q ? `?${q}` : '?', { replace: true, reset: false });
	}
	const filtrar = (cambio: Partial<FiltroActividad>) =>
		ir(consultaDe({ ...filtro, ...cambio, cursor: undefined }));
	const sacarFiltros = () => ir('');
	const valor = (e: Event) => (e.currentTarget as HTMLInputElement | HTMLSelectElement).value;

	async function verMas() {
		if (!siguiente || cargando) return;
		cargando = true;
		try {
			const r = await fetch(
				`/panel/actividad/pagina?${consultaDe({ ...filtro, cursor: siguiente })}`
			);
			if (!r.ok) throw new Error();
			const p = (await r.json()) as PaginaActividad;
			extra = [...extra, ...p.items];
			siguiente = p.siguiente;
		} catch {
			ui.toast({ k: 'err', title: 'No se pudo cargar más', desc: 'Probá de nuevo en un rato.' });
		} finally {
			cargando = false;
		}
	}

	const exportar = $derived(
		`/panel/actividad/exportar${consultaDe(filtro) ? `?${consultaDe(filtro)}` : ''}`
	);

	async function deshacer(id: string) {
		const r = await fetch(`/panel/actividad/${encodeURIComponent(id)}/deshacer`, {
			method: 'POST'
		});
		if (!r.ok) return ui.toast({ k: 'err', title: 'No se pudo deshacer', desc: 'Probá de nuevo.' });
		await invalidateAll();
		ui.toast({ k: 'ok', title: 'Cambio deshecho', desc: 'Volvimos todo a como estaba' });
	}

	const ICONO = {
		hornero: [I.phSparkleFill, 'rgba(210,88,47,.12)', '#C24E25'],
		ia: [I.phSparkleBold, 'rgba(91,76,196,.12)', '#5B4CC4'],
		soporte: [I.phLifebuoyFill, 'var(--chip)', 'var(--mu)'],
		sistema: [I.phGearSixFill, 'var(--chip)', 'var(--mu)'],
		equipo: [I.phUserFill, 'rgba(210,88,47,.12)', '#C24E25']
	} as const;
	const inicial = (q: string) => (q.trim()[0] ?? '·').toUpperCase();
	const hoyAR = new Date(Date.now() - 3 * 3_600_000).toISOString().slice(0, 10);
</script>

<Seo
	title="Actividad · Panel"
	description="Quién hizo qué en tu tienda"
	path="/panel/actividad"
	noindex
/>

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Actividad</h1>
			<div class="c-s">Quién de tu equipo hizo qué y cuándo. Se guarda un año.</div>
		</div>
		<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- descarga (un +server), con la query de los filtros -->
		<a class="btn btn-secondary btn-sm bajar" href={exportar} download data-sveltekit-reload
			><Icon icon={I.phDownloadSimpleBold} size={16} /> Bajar CSV</a
		>
	</div>

	<form class="filtros" use:anim={{ t: 'rise', delay: 40 }} onsubmit={(e) => e.preventDefault()}>
		<label class="field"
			>Persona
			<select
				class="select"
				value={filtro.persona ?? ''}
				onchange={(e) => filtrar({ persona: valor(e) || undefined })}
			>
				<option value="">Todas</option>
				{#each data.pagina.personas as p (p.id)}<option value={p.id}>{p.nombre}</option>{/each}
			</select></label
		>
		<label class="field"
			>Tipo
			<select
				class="select"
				value={filtro.tipo ?? ''}
				onchange={(e) => filtrar({ tipo: (valor(e) || undefined) as FiltroActividad['tipo'] })}
			>
				<option value="">Todo</option>
				{#each CATEGORIAS as [id, nombre] (id)}<option value={id}>{nombre}</option>{/each}
			</select></label
		>
		<label class="field"
			>Desde
			<input
				class="input"
				type="date"
				max={filtro.hasta ?? hoyAR}
				value={filtro.desde ?? ''}
				onchange={(e) => filtrar({ desde: valor(e) || undefined })}
			/></label
		>
		<label class="field"
			>Hasta
			<input
				class="input"
				type="date"
				min={filtro.desde}
				max={hoyAR}
				value={filtro.hasta ?? ''}
				onchange={(e) => filtrar({ hasta: valor(e) || undefined })}
			/></label
		>
		{#if hayFiltros}
			<button type="button" class="limpiar" onclick={sacarFiltros}
				><Icon icon={I.phXBold} size={14} /> Sacar filtros</button
			>
		{/if}
	</form>

	{#if items.length}
		<div class="lista" use:anim={{ t: 'rise', delay: 60 }}>
			{#each items as a (a.id)}
				{@const [icono, bg, fg] = ICONO[a.tipo]}
				<div class="evento" class:deshecho={a.deshecho}>
					{#if a.foto}
						<img src={unsplash(a.foto, 72, '&h=72&crop=faces')} alt="" />
					{:else if a.tipo === 'equipo'}
						<span class="av" class:vos={a.quien === 'Vos'}>{inicial(a.quien)}</span>
					{:else}
						<span class="av" style:background={bg} style:color={fg}
							><Icon icon={icono} size={17} /></span
						>
					{/if}
					<span class="txt">
						<span class="t"><b>{a.quien}</b> {a.texto}</span>
						<!-- Cada dato entero en su renglón (sin «·» sueltos al cortar). -->
						<span class="s"
							>{#each [cuando(a.cuando), a.rol, nombreCategoria(a.categoria), a.donde].filter(Boolean) as x, i (i)}<span
									class="d">{i ? '· ' : ''}{x}</span
								>{/each}</span
						>
					</span>
					{#if a.deshacer}
						<button type="button" class="deshacer" onclick={() => deshacer(a.id)}>Deshacer</button>
					{/if}
				</div>
			{/each}
		</div>
		{#if siguiente}
			<button
				type="button"
				class="btn btn-soft mas"
				aria-busy={cargando}
				disabled={cargando}
				onclick={verMas}>{cargando ? 'Cargando…' : 'Ver más'}</button
			>
		{/if}
	{:else if hayFiltros}
		<Vacio
			hornero="lupa"
			titulo="Nada con esos filtros"
			texto="No hay cambios de esa persona, de ese tipo o en esas fechas. Probá con otros."
		>
			<button type="button" class="btn btn-secondary" onclick={sacarFiltros}>Sacar filtros</button>
		</Vacio>
	{:else}
		<Vacio
			hornero="horno-vacio"
			titulo="Todavía no hay nada registrado"
			texto="Desde ahora, cada precio, pedido, cobro o cambio de configuración queda acá con quién lo hizo. Lo de antes no se guardó."
		/>
	{/if}
</div>

<style>
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 18px;
	}
	.cab {
		display: flex;
		align-items: end;
		gap: 12px;
		flex-wrap: wrap;
	}
	.c-txt {
		flex: 1;
		min-width: 220px;
	}
	h1 {
		margin: 0;
		font-size: 30px;
		letter-spacing: -0.035em;
		font-weight: 600;
	}
	.c-s {
		margin-top: 2px;
		font-size: 15px;
		color: var(--mu);
		text-wrap: pretty;
	}
	.bajar {
		gap: 6px;
		min-height: 40px;
	}
	.filtros {
		display: grid;
		grid-template-columns: repeat(4, minmax(0, 1fr)) auto;
		align-items: end;
		gap: 10px;
	}
	.filtros .select,
	.filtros .input {
		min-height: 44px;
		padding: 9px 12px;
		font-size: 14px;
	}
	.limpiar {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		min-height: 44px;
		padding: 0 12px;
		border: 0;
		border-radius: 999px;
		background: transparent;
		color: var(--mu);
		font:
			600 13px 'Instrument Sans',
			sans-serif;
		white-space: nowrap;
	}
	.lista {
		padding: 8px 20px;
		border-radius: 24px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.evento {
		display: flex;
		gap: 14px;
		padding: 14px 0;
		border-bottom: 1px solid var(--ln);
	}
	.evento:last-child {
		border-bottom: 0;
	}
	.evento img,
	.av {
		width: 36px;
		height: 36px;
		border-radius: 50%;
		flex: none;
	}
	.evento img {
		object-fit: cover;
	}
	.av {
		display: grid;
		place-items: center;
		background: rgba(210, 88, 47, 0.12);
		color: #c24e25;
		font-size: 14px;
		font-weight: 700;
	}
	.av.vos {
		background: #2a140c;
		color: #fbf7f2;
	}
	.txt {
		flex: 1;
		min-width: 0;
	}
	.t {
		display: block;
		font-size: 14px;
		line-height: 1.4;
		overflow-wrap: anywhere;
		text-wrap: pretty;
	}
	.deshecho .t {
		color: var(--mu);
	}
	.s {
		display: block;
		font-size: 12.5px;
		color: var(--fa);
		margin-top: 2px;
	}
	.d {
		display: inline-block;
		max-width: 100%;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
		vertical-align: top;
	}
	.d + .d {
		margin-left: 0.3em;
	}
	.deshacer {
		align-self: center;
		padding: 7px 12px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font:
			600 12.5px 'Instrument Sans',
			sans-serif;
	}
	.mas {
		align-self: center;
		min-height: 44px;
	}
	@media (max-width: 760px) {
		.filtros {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
		.limpiar {
			grid-column: 1 / -1;
			justify-self: start;
		}
		.lista {
			padding: 4px 16px;
		}
	}
</style>
