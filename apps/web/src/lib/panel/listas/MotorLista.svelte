<script lang="ts" generics="R extends { id: string }">
	import { deslizable } from '#lib/panel/deslizable.ts';
	import { untrack, type Snippet } from 'svelte';
	import { SvelteSet } from 'svelte/reactivity';
	import { goto, invalidateAll } from '$app/navigation';
	import { page } from '$app/state';
	import Icon from '#lib/components/Icon.svelte';
	import Interruptor from '#lib/components/Interruptor.svelte';
	import MarcaIcono from '#lib/components/MarcaIcono.svelte';
	import * as I from '#lib/icons.ts';
	import { getUi } from '#lib/ui.svelte.ts';
	import { anim } from '../anim.ts';
	import { escribir, firma, leer, VACIO } from './filtros.ts';
	import type { ConfigLista, Ctx, FiltrosUrl, ListaGenerica, Opcion } from './tipos.ts';

	/**
	 * Motor de listas del handoff (design/PanelListas.dc.html): encabezado, sugerencia de Tincho,
	 * vistas, segmentos, búsqueda, facetas, rango, banderas, orden, lista virtual, selección y lote.
	 * Los filtros viven en la URL; las filas se piden por ventanas al scrollear.
	 */
	let {
		config,
		lista,
		abrir,
		extra,
		antes,
		despues,
		ruta = undefined,
		exportar = true
	}: {
		config: ConfigLista<R>;
		lista: ListaGenerica;
		/** Al tocar una fila (ficha o editor). */
		abrir?: (r: R) => void;
		/** Botones de más en el encabezado. */
		extra?: Snippet;
		/** Algo entre el encabezado y las vistas (banners). */
		antes?: Snippet;
		/** Algo debajo de la lista. */
		despues?: Snippet;
		/** De dónde se piden las ventanas y adónde van las acciones (por defecto, la del panel). */
		ruta?: string;
		/** Mostrar "Exportar" (el backoffice del equipo no exporta datos personales). */
		exportar?: boolean;
	} = $props();
	const base = $derived(ruta ?? `/panel/listas/${config.lista}`);
	const ui = getUi();
	const f = $derived(leer(page.url.searchParams));
	const vista = $derived(
		config.vistas.find((v) => firma({ ...VACIO, ...v.filtros }) === firma(f))?.id ?? null
	);

	function ir(cambio: Partial<FiltrosUrl>) {
		const q = escribir({ ...f, ...cambio }).toString();
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- misma página, sólo cambia la query
		void goto(q ? `?${q}` : '?', { replace: true, reset: false });
		arriba();
	}
	const alternar = (xs: string[], x: string) =>
		xs.includes(x) ? xs.filter((y) => y !== x) : [...xs, x];

	let q = $state(untrack(() => f.q));
	let espera: ReturnType<typeof setTimeout>;
	function buscar(v: string) {
		q = v;
		clearTimeout(espera);
		espera = setTimeout(() => ir({ q: v }), 250);
	}

	// ---- Lista virtual ----
	const FILA = 76;
	const VENTANA = 60;
	let ancho = $state(1440);
	const amplio = $derived(ancho >= 900);
	const alto = $derived(ancho < 760 ? 520 : 600);
	let scroll = $state(0);
	let caja = $state<HTMLElement>();
	let filas = $state.raw<(R | undefined)[]>([]);
	const pedidas = new SvelteSet<number>();
	$effect(() => {
		const l = lista;
		untrack(() => {
			filas = Array.from({ length: l.total }, (_, i) => l.filas[i] as unknown as R | undefined);
			pedidas.clear();
			pedidas.add(0);
		});
	});
	const inicio = $derived(Math.max(0, Math.floor(scroll / FILA) - 4));
	const fin = $derived(Math.min(lista.total, Math.ceil((scroll + alto) / FILA) + 6));
	const visibles = $derived(
		Array.from({ length: Math.max(0, fin - inicio) }, (_, k) => inicio + k)
	);
	$effect(() => {
		for (let i = inicio; i < fin; i++)
			if (!filas[i]) void pedirVentana(Math.floor(i / VENTANA) * VENTANA);
	});
	async function pedirVentana(desde: number) {
		if (pedidas.has(desde)) return;
		pedidas.add(desde);
		const params = escribir(f);
		params.set('desde', String(desde));
		params.set('cuantos', String(VENTANA));
		const r = await fetch(`${base}?${params}`);
		if (!r.ok) return pedidas.delete(desde);
		const l = (await r.json()) as ListaGenerica;
		const copia = filas.slice();
		l.filas.forEach((x, k) => (copia[desde + k] = x as unknown as R));
		filas = copia;
	}
	function arriba() {
		caja?.scrollTo({ top: 0 });
		scroll = 0;
	}
	let pendiente = 0;
	function alScrollear(e: Event) {
		const t = (e.currentTarget as HTMLElement).scrollTop;
		if (pendiente) return;
		pendiente = requestAnimationFrame(() => {
			pendiente = 0;
			scroll = t;
		});
	}

	// ---- Acciones ----
	const seleccion = new SvelteSet<string>();
	const ctx: Ctx = {
		ui,
		refrescar: () => invalidateAll(),
		async accion(ids, accion, valor) {
			const r = await fetch(base, {
				method: 'POST',
				headers: { 'content-type': 'application/json' },
				body: JSON.stringify({ ids, accion, valor })
			});
			if (!r.ok) {
				// El servidor explica qué pasó (por ejemplo, el rechazo de ARCA): se muestra tal cual.
				const m = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
				ui.toast({ k: 'err', title: 'No se pudo', desc: m ?? 'Probá de nuevo en un rato.' });
				return 0;
			}
			const { hechos } = (await r.json()) as { hechos: number };
			await invalidateAll();
			return hechos;
		}
	};
	function tildarVisibles() {
		for (let i = inicio; i < fin; i++) {
			const r = filas[i];
			if (r) seleccion.add(r.id);
		}
	}
	let tipHecho = $state(false);
	const tip = $derived(config.tip?.(lista) ?? null);

	// ---- Filtros visibles ----
	let mas = $state(false);
	const orden = $derived(config.ordenes.find((o) => o.k === f.orden) ?? config.ordenes[0]);
	function ciclarOrden() {
		const i = config.ordenes.findIndex((o) => o.k === orden.k);
		ir({ orden: config.ordenes[(i + 1) % config.ordenes.length].k });
	}
	const opciones = (k: string): Opcion[] => {
		const fa = config.facetas.find((x) => x.k === k)!;
		if (fa.opciones) return fa.opciones;
		const conteos = lista.facetas[k] ?? {};
		const elegidas = f.facetas[k] ?? [];
		const top = Object.entries(conteos)
			.sort((a, b) => b[1] - a[1])
			.slice(0, 6)
			.map(([v]) => v);
		return [...new Set([...top, ...elegidas])].map((v): Opcion => ({ k: v, t: v }));
	};
	const etiqueta = (k: string, v: string) =>
		config.segmentos.find((s) => s.k === v && k === 'seg')?.t ??
		config.facetas.find((x) => x.k === k)?.opciones?.find((o) => o.k === v)?.t ??
		v;
	const chips = $derived.by(() => {
		const c: [string, Partial<FiltrosUrl>][] = [];
		for (const s of f.seg) c.push([etiqueta('seg', s), { seg: f.seg.filter((x) => x !== s) }]);
		for (const [k, vs] of Object.entries(f.facetas))
			for (const v of vs)
				c.push([etiqueta(k, v), { facetas: { ...f.facetas, [k]: vs.filter((x) => x !== v) } }]);
		if (f.rango) c.push([`${config.rango.t} ${config.rango.formato(f.rango)}`, { rango: 0 }]);
		for (const b of f.banderas)
			c.push([
				config.banderas.find((x) => x.k === b)?.t ?? b,
				{ banderas: f.banderas.filter((x) => x !== b) }
			]);
		if (f.q) c.push([`«${f.q}»`, { q: '' }]);
		return c;
	});
	function limpiar() {
		q = '';
		ir({ ...VACIO, orden: f.orden });
	}
	let rango = $derived(f.rango);
	// Sin búsqueda ni filtros (salvo el orden): si no hay filas, la lista está vacía de verdad.
	const sinFiltros = $derived(firma({ ...VACIO, orden: f.orden }) === firma(f));

	/**
	 * Vacía de verdad (sin filas y sin filtros): no hay nada que buscar, filtrar, exportar ni
	 * seleccionar, así que queda sólo el encabezado y el vacío con su acción.
	 */
	const vacia = $derived(!lista.total && sinFiltros && !vista);
	// Con pocas filas la caja no deja un hueco en blanco abajo; vacía, lo justo para el vacío.
	const altoCaja = $derived(lista.total ? Math.min(alto, lista.total * FILA + 10) : 340);
	const n = $derived(lista.total.toLocaleString('es-AR'));
	const rangoTxt = $derived(
		lista.total
			? `Mostrando ${(inicio + 1).toLocaleString('es-AR')}–${Math.min(lista.total, inicio + Math.ceil(alto / FILA)).toLocaleString('es-AR')} de ${n}`
			: sinFiltros
				? ''
				: 'Sin resultados'
	);
</script>

<svelte:window bind:innerWidth={ancho} />

<div class="motor">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>{config.titulo}</h1>
			<div class="c-s">{config.kpi(lista)}</div>
		</div>
		<!-- eslint-disable svelte/no-navigation-without-resolve -- la descarga del CSV no es una navegación -->
		{#if exportar && !vacia}<a
				class="secundario"
				href="{base}/exportar?{escribir(f)}"
				download
				onclick={() =>
					ui.toast({
						k: 'ok',
						title: `Exportando ${n} ${config.unidad}`,
						desc: 'Con los filtros aplicados · en CSV para Excel'
					})}><Icon icon={I.phDownloadSimpleBold} size={14} />Exportar</a
			>{/if}
		<!-- eslint-enable svelte/no-navigation-without-resolve -->
		{@render extra?.()}
		{#if config.primario && !(vacia && !config.vacio)}
			{@const p = config.primario}
			<button type="button" class="primario" onclick={() => p.hacer(ctx)}
				><Icon icon={p.icon} size={15} />{p.t}</button
			>
		{/if}
	</div>

	{@render antes?.()}

	{#if tip && !tipHecho}
		<div class="tip" use:anim={{ t: 'rise', delay: 40 }}>
			<img src="/hornero/pensando.webp" alt="" width="40" height="40" />
			<div class="tip-t"><b>Tincho:</b> {tip.texto}</div>
			<button
				type="button"
				onclick={async () => {
					tipHecho = true;
					await tip.hacer(ctx);
				}}>{tip.boton}</button
			>
		</div>
	{/if}

	{#if !vacia}
		<div
			class="vistas"
			use:anim={{ t: 'rise', delay: 80 }}
			role="group"
			aria-label="Vistas"
			use:deslizable
		>
			{#each config.vistas as v (v.id)}
				<button
					type="button"
					class="vista"
					aria-pressed={vista === v.id}
					onclick={() => {
						q = v.filtros.q ?? '';
						ir({ ...VACIO, orden: f.orden, ...v.filtros });
					}}><Icon icon={v.icon} size={15} style="color:{v.color}" />{v.t}</button
				>
			{/each}
		</div>

		<div
			class="segs"
			use:anim={{ t: 'rise', delay: 120 }}
			role="group"
			aria-label="Segmentos"
			use:deslizable
		>
			{#each config.segmentos as s (s.k)}
				{@const on = f.seg.includes(s.k)}
				<button
					type="button"
					class="seg"
					class:on
					aria-pressed={on}
					onclick={() => ir({ seg: alternar(f.seg, s.k) })}
				>
					<span class="s-t"><span class="s-p" style:background={s.c}></span>{s.t}</span>
					<span class="s-n">{(lista.segmentos[s.k] ?? 0).toLocaleString('es-AR')}</span>
					<span class="s-b" style:background={s.c} style:opacity={on || !f.seg.length ? 1 : 0.25}
					></span>
				</button>
			{/each}
		</div>

		<div class="filtros" use:anim={{ t: 'rise', delay: 160 }}>
			<div class="f-fila">
				<label class="buscar">
					<Icon icon={I.phMagnifyingGlass} size={16} style="color:var(--mu)" />
					<input
						value={q}
						oninput={(e) => buscar(e.currentTarget.value)}
						placeholder={config.buscar}
						aria-label="Buscar {config.unidad}"
					/>
					{#if q}<button
							type="button"
							class="b-x"
							aria-label="Borrar búsqueda"
							onclick={() => buscar('')}><Icon icon={I.phXCircleFill} size={16} /></button
						>{/if}
				</label>
				<button
					type="button"
					class="pildora"
					class:on={mas}
					aria-expanded={mas}
					onclick={() => (mas = !mas)}
					><Icon icon={I.phSlidersHorizontalBold} size={15} />Filtros{#if chips.length}<span
							class="cuenta">{chips.length}</span
						>{/if}</button
				>
				<button type="button" class="pildora" onclick={ciclarOrden} aria-label="Orden: {orden.t}"
					><Icon icon={I.phSortAscendingBold} size={15} />{orden.t}</button
				>
			</div>
			{#if mas}
				<div class="mas" use:anim={{ t: 'rise' }}>
					{#each config.facetas as fa (fa.k)}
						<div>
							<div class="m-t">{fa.t}</div>
							<div class="opciones">
								{#each opciones(fa.k) as o (o.k)}
									{@const on = (f.facetas[fa.k] ?? []).includes(o.k)}
									<button
										type="button"
										class="op"
										class:on
										aria-pressed={on}
										onclick={() =>
											ir({
												facetas: { ...f.facetas, [fa.k]: alternar(f.facetas[fa.k] ?? [], o.k) }
											})}
										>{#if o.marca}<MarcaIcono
												k={o.marca}
												size={18}
												radius={5}
											/>{:else if o.icon}<Icon
												icon={o.icon}
												size={14}
												style="color:{o.col ?? 'inherit'}"
											/>{/if}{o.t}<span class="op-n">{lista.facetas[fa.k]?.[o.k] ?? 0}</span
										></button
									>
								{/each}
							</div>
						</div>
					{/each}
					<div>
						<label class="m-t m-rango" for="rango-{config.lista}"
							>{config.rango.t}<b>{rango ? config.rango.formato(rango) : 'Cualquiera'}</b></label
						>
						<input
							id="rango-{config.lista}"
							type="range"
							min="0"
							max={config.rango.max(lista)}
							step={config.rango.paso}
							bind:value={rango}
							onchange={() => ir({ rango })}
						/>
					</div>
					<div class="banderas">
						{#each config.banderas as b (b.k)}
							<div class="bandera">
								<span>{b.t}</span>
								<Interruptor
									label={b.t}
									checked={f.banderas.includes(b.k)}
									onchange={() => ir({ banderas: alternar(f.banderas, b.k) })}
								/>
							</div>
						{/each}
					</div>
				</div>
			{/if}
			{#if chips.length}
				<div class="chips">
					{#each chips as [t, cambio] (t)}
						<button
							type="button"
							class="chip"
							onclick={() => {
								if ('q' in cambio) q = '';
								ir(cambio);
							}}>{t}<Icon icon={I.phXBold} size={11} /></button
						>
					{/each}
					<button type="button" class="limpiar" onclick={limpiar}>Limpiar todo</button>
				</div>
			{/if}
		</div>
	{/if}

	<div class="lista" use:anim={{ t: 'rise', delay: 200 }}>
		{#if !vacia}<div class="l-cab">
				<button
					type="button"
					class="tilde"
					aria-label="Seleccionar los visibles"
					onclick={tildarVisibles}
				></button>
				<span class="l-r">{rangoTxt}</span>
				<span class="l-v"
					><Icon icon={I.phLightningFill} size={13} style="color:#b0441d" />{n}
					{lista.total === 1 && config.unidad.endsWith('s')
						? config.unidad.slice(0, -1)
						: config.unidad}</span
				>
			</div>{/if}
		<div
			class="scroll"
			style:height="{altoCaja}px"
			bind:this={caja}
			onscroll={alScrollear}
			data-lenis-prevent
			role="list"
			aria-label={config.titulo}
		>
			<div style:height="{lista.total * FILA}px" class="alto">
				{#each visibles as i (i)}
					{@const r = filas[i]}
					<div class="fila" style:top="{i * FILA}px" role="listitem">
						{#if r}
							{@const v = config.fila(r)}
							{@const sel = seleccion.has(r.id)}
							<div class="f-in" class:sel class:con-stepper={v.stepper !== undefined}>
								<button
									type="button"
									class="tilde"
									class:on={sel}
									aria-label="Seleccionar {v.t}"
									aria-pressed={sel}
									onclick={() => (sel ? seleccion.delete(r.id) : seleccion.add(r.id))}
									>{#if sel}<Icon icon={I.phCheckBold} size={12} />{/if}</button
								>
								<button type="button" class="abrir" onclick={() => abrir?.(r)}>
									<span class="media">
										{#if v.img}<img src={v.img} alt="" style:border-radius={v.radio} />{:else}<span
												class="ini"
												style:border-radius={v.radio}
												style:background={v.iniBg}
												style:color={v.iniFg}>{v.ini}</span
											>{/if}
										{#if v.insignia}
											<span class="insignia"
												>{#if v.insignia.marca}<MarcaIcono
														k={v.insignia.marca}
														size={20}
														radius={7}
													/>{:else if v.insignia.icon}<span class="ins-ico"
														><Icon
															icon={v.insignia.icon}
															size={12}
															style="color:{v.insignia.col ?? 'inherit'}"
														/></span
													>{/if}</span
											>
										{/if}
									</span>
									<span class="quien">
										<span class="q-n"
											><span class="q-t">{v.t}</span>{#each v.tags ?? [] as g (g.t)}<span
													title={g.t}
													class="tag"><Icon icon={g.icon} size={13} style="color:{g.col}" /></span
												>{/each}</span
										>
										<span class="q-s">{v.sub}</span>
									</span>
									{#if amplio}
										<span class="medio">
											{#if v.spark}<span class="spark"
													>{#each v.spark as b, j (j)}<span
															style:height="{b.h}px"
															style:background={b.c}
														></span>{/each}</span
												>{/if}
											{#if v.medidor}<span class="medidor"
													><span style:width="{v.medidor.w}%" style:background={v.medidor.c}
													></span>{#if v.medidor.marca !== undefined}<span
															class="m-marca"
															title="Punto de reposición"
															style:left="{v.medidor.marca}%"
														></span>{/if}</span
												>{/if}
											{#if v.pasos}<span class="pasos"
													>{#each v.pasos as c, j (j)}<span style:background={c}
														></span>{/each}</span
												>{/if}
											<span class="v-l" style:color={v.vC}
												>{#if v.vI}<Icon icon={v.vI} size={13} />{/if}{v.vL}</span
											>
										</span>
									{/if}
									{#if v.chip}<span
											class="chip-f"
											style:background={v.chip.bg}
											style:color={v.chip.fg}>{v.chip.t}</span
										>{/if}
									<span class="valor"
										><span class="v-v">{v.v}</span><span class="v-s">{v.vs}</span></span
									>
								</button>
								{#if v.stepper !== undefined && config.ajustar}
									<span class="stepper">
										<button
											type="button"
											aria-label="Restar uno a {v.t}"
											onclick={() => config.ajustar?.(r, -1, ctx)}
											><Icon icon={I.phMinusBold} size={12} /></button
										>
										<span>{v.stepper}</span>
										<button
											type="button"
											aria-label="Sumar uno a {v.t}"
											onclick={() => config.ajustar?.(r, 1, ctx)}
											><Icon icon={I.phPlusBold} size={12} /></button
										>
									</span>
								{/if}
								{#if v.toggle && config.alternar}
									<span class="toggle" title={v.toggle.t}>
										<Interruptor
											label="{v.toggle.t}: {v.t}"
											checked={v.toggle.on}
											onchange={(on: boolean) => config.alternar?.(r, on, ctx)}
										/>
									</span>
								{/if}
								{#if v.act && config.accionFila}
									<button
										type="button"
										class="accion"
										style:background={v.act.bg}
										style:color={v.act.fg}
										aria-label="{v.act.t}: {v.t}"
										onclick={() => config.accionFila?.(r, ctx)}
										><Icon icon={v.act.icon} size={14} />{#if amplio}{v.act.t}{/if}</button
									>
								{/if}
							</div>
						{:else}
							<div class="f-in esqueleto"><span></span><span></span></div>
						{/if}
					</div>
				{/each}
			</div>
			{#if !lista.total}
				<div class="vacio">
					{#if sinFiltros}
						{@const v = config.vacio}
						<div>
							<img
								src="/hornero/{v?.hornero ?? 'horno-vacio'}.webp"
								alt=""
								width="110"
								height="110"
							/>
							<div class="v-t">{v?.titulo ?? `Todavía no hay ${config.unidad}`}</div>
							{#if v?.texto}<p class="v-x">{v.texto}</p>{/if}
							{#if v?.accion?.href}
								<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta interna del panel -->
								<a class="v-b" href={v.accion.href}>{v.accion.t}</a>
							{:else if v?.accion?.hacer}
								{@const hacer = v.accion.hacer}
								<button type="button" class="v-b" onclick={() => hacer(ctx)}>{v.accion.t}</button>
							{:else if !v && config.primario}
								<!-- Sin un vacío propio, la acción principal de la lista. -->
								{@const p = config.primario}
								<button type="button" class="v-b" onclick={() => p.hacer(ctx)}>{p.t}</button>
							{/if}
						</div>
					{:else}
						<div>
							<img src="/hornero/lupa.webp" alt="" width="110" height="110" />
							<div class="v-t">Nada con esos filtros</div>
							<button type="button" class="v-b" onclick={limpiar}>Limpiar filtros</button>
						</div>
					{/if}
				</div>
			{/if}
		</div>
		{#if scroll > 1200}
			<button
				type="button"
				class="arriba"
				aria-label="Volver arriba"
				onclick={() => caja?.scrollTo({ top: 0, behavior: 'smooth' })}
				><Icon icon={I.phArrowUpBold} size={16} /></button
			>
		{/if}
	</div>
</div>

{@render despues?.()}

{#if seleccion.size}
	<div class="lote" use:anim={{ t: 'rise' }} use:deslizable>
		<span class="lo-t"
			>{seleccion.size} {seleccion.size === 1 ? 'seleccionado' : 'seleccionados'}</span
		>
		{#each config.lote as l (l.t)}
			<button
				type="button"
				onclick={async () => {
					const ids = [...seleccion];
					seleccion.clear();
					await l.hacer(ids, ctx);
				}}><Icon icon={l.icon} size={14} />{l.t}</button
			>
		{/each}
		<button
			type="button"
			class="lo-x"
			aria-label="Quitar la selección"
			onclick={() => seleccion.clear()}><Icon icon={I.phXBold} size={14} /></button
		>
	</div>
{/if}

<style>
	.motor {
		display: flex;
		flex-direction: column;
		gap: 14px;
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
	}
	.secundario,
	.cab :global(.secundario) {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 16px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font-size: 13.5px;
		font-weight: 600;
		box-shadow: var(--sh);
		white-space: nowrap;
	}
	.primario {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 18px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font-size: 13.5px;
		font-weight: 600;
		box-shadow: 0 10px 22px -10px rgba(210, 88, 47, 0.8);
		white-space: nowrap;
	}
	.tip {
		display: flex;
		align-items: center;
		gap: 12px;
		padding: 12px 14px;
		border-radius: 18px;
		background: rgba(210, 88, 47, 0.08);
		border: 1px solid rgba(210, 88, 47, 0.18);
	}
	.tip img {
		object-fit: contain;
		flex: none;
	}
	.tip-t {
		flex: 1;
		min-width: 0;
		font-size: 13.5px;
		line-height: 1.4;
	}
	.tip button {
		flex: none;
		padding: 8px 14px;
		border-radius: 999px;
		border: 0;
		background: #2a140c;
		color: #fff;
		font-size: 12.5px;
		font-weight: 600;
		white-space: nowrap;
	}
	.vistas {
		display: flex;
		gap: 6px;
		overflow-x: auto;
		scrollbar-width: none;
	}
	.vista {
		display: flex;
		align-items: center;
		gap: 7px;
		border: 1.5px solid transparent;
		white-space: nowrap;
		font-size: 13px;
		font-weight: 600;
		padding: 8px 13px;
		border-radius: 999px;
		background: var(--sf);
		color: var(--tx);
	}
	.vista[aria-pressed='true'] {
		border-color: #d2582f;
		background: rgba(210, 88, 47, 0.08);
	}
	.segs {
		display: flex;
		border-radius: 18px;
		overflow-x: auto;
		scrollbar-width: none;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.seg {
		flex: 1 0 120px;
		min-width: 0;
		padding: 12px 14px;
		border: 0;
		border-right: 1px solid var(--ln);
		background: transparent;
		color: var(--tx);
		text-align: left;
		transition: background 0.2s;
	}
	.seg:last-child {
		border-right: 0;
	}
	.seg.on {
		background: var(--chip);
	}
	.s-t {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 12.5px;
		color: var(--mu);
		white-space: nowrap;
		overflow: hidden;
	}
	.s-p {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		flex: none;
	}
	.s-n {
		display: block;
		font-size: 22px;
		font-weight: 600;
		letter-spacing: -0.03em;
		margin-top: 2px;
	}
	.s-b {
		display: block;
		margin-top: 6px;
		height: 4px;
		border-radius: 3px;
	}
	.filtros {
		display: flex;
		flex-direction: column;
		gap: 10px;
		padding: 12px;
		border-radius: 20px;
		background: var(--sf);
		box-shadow: var(--shd);
	}
	.f-fila {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
		align-items: center;
	}
	.buscar {
		flex: 1 1 240px;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 10px 14px;
		border-radius: 999px;
		background: var(--sf3);
	}
	.buscar:focus-within {
		box-shadow: 0 0 0 2px rgba(210, 88, 47, 0.35);
	}
	.buscar input {
		flex: 1;
		min-width: 0;
		border: 0;
		outline: 0;
		background: transparent;
		font-size: 14px;
		font-weight: 500;
		color: var(--tx);
	}
	.b-x {
		border: 0;
		background: transparent;
		color: var(--mu);
		display: grid;
		place-items: center;
		padding: 0;
	}
	.pildora {
		display: flex;
		align-items: center;
		gap: 7px;
		padding: 9px 14px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font-size: 13px;
		font-weight: 600;
		white-space: nowrap;
	}
	.pildora.on {
		border-color: #d2582f;
		background: rgba(210, 88, 47, 0.08);
	}
	.cuenta {
		min-width: 18px;
		height: 18px;
		padding: 0 5px;
		border-radius: 999px;
		background: #b0441d;
		color: #fff;
		font-size: 11px;
		font-weight: 700;
		display: grid;
		place-items: center;
	}
	.mas {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
		gap: 12px;
		padding: 6px 2px 2px;
	}
	.m-t {
		font-size: 12px;
		font-weight: 600;
		color: var(--mu);
		margin-bottom: 6px;
	}
	.m-rango {
		display: flex;
		justify-content: space-between;
	}
	.m-rango b {
		color: var(--tx);
	}
	.mas input[type='range'] {
		width: 100%;
		accent-color: #d2582f;
	}
	.opciones {
		display: flex;
		flex-wrap: wrap;
		gap: 5px;
	}
	.op {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 6px 10px;
		border-radius: 999px;
		border: 1.5px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font-size: 12.5px;
		font-weight: 600;
		white-space: nowrap;
	}
	.op.on {
		border-color: #d2582f;
		background: rgba(210, 88, 47, 0.08);
	}
	.op-n {
		color: var(--fa);
		font-weight: 500;
	}
	.banderas {
		display: flex;
		flex-direction: column;
		gap: 6px;
	}
	.bandera {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 4px 0;
		font-size: 13px;
		font-weight: 500;
	}
	.bandera span {
		flex: 1;
	}
	.chips {
		display: flex;
		gap: 6px;
		flex-wrap: wrap;
		align-items: center;
	}
	.chip {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 5px 8px 5px 11px;
		border-radius: 999px;
		border: 0;
		background: rgba(210, 88, 47, 0.12);
		color: #b0441d;
		font-size: 12.5px;
		font-weight: 600;
	}
	.limpiar {
		border: 0;
		background: transparent;
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
		text-decoration: underline;
	}
	.lista {
		position: relative;
		border-radius: 24px;
		background: var(--sf);
		box-shadow: var(--shd);
		overflow: hidden;
	}
	.l-cab {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 16px;
		border-bottom: 1px solid var(--ln);
		font-size: 12.5px;
		color: var(--mu);
	}
	.l-r {
		flex: 1;
	}
	.l-v {
		display: flex;
		align-items: center;
		gap: 6px;
		white-space: nowrap;
	}
	.tilde {
		width: 20px;
		height: 20px;
		flex: none;
		padding: 0;
		border-radius: 6px;
		border: 1.5px solid var(--ln2);
		background: var(--sf);
		color: #fff;
		display: grid;
		place-items: center;
	}
	.tilde.on {
		border-color: #d2582f;
		background: #d2582f;
	}
	.scroll {
		overflow-y: auto;
		overscroll-behavior: contain;
		position: relative;
	}
	.alto {
		position: relative;
	}
	.fila {
		position: absolute;
		left: 0;
		right: 0;
		height: 76px;
		padding: 5px 10px;
	}
	.f-in {
		height: 100%;
		display: flex;
		align-items: center;
		gap: 14px;
		padding: 0 12px;
		border-radius: 16px;
		border: 1px solid transparent;
		transition: background 0.15s;
	}
	.f-in:hover {
		background: var(--chip);
	}
	.f-in.sel {
		background: rgba(210, 88, 47, 0.06);
		border-color: rgba(210, 88, 47, 0.3);
	}
	.abrir {
		flex: 1;
		min-width: 0;
		height: 100%;
		display: flex;
		align-items: center;
		gap: 14px;
		border: 0;
		padding: 0;
		background: transparent;
		color: var(--tx);
		text-align: left;
	}
	.media {
		position: relative;
		flex: none;
		width: 42px;
		height: 42px;
	}
	.media img,
	.ini {
		width: 42px;
		height: 42px;
		object-fit: cover;
		display: block;
		background: var(--sf3);
	}
	.ini {
		display: grid;
		place-items: center;
		font-size: 14px;
		font-weight: 700;
	}
	.insignia {
		position: absolute;
		right: -4px;
		bottom: -4px;
		border-radius: 7px;
		box-shadow: 0 0 0 2px var(--sf);
		display: block;
	}
	.ins-ico {
		width: 20px;
		height: 20px;
		border-radius: 7px;
		display: grid;
		place-items: center;
		background: var(--sf);
	}
	.quien {
		flex: 1 1 150px;
		min-width: 0;
	}
	.q-n {
		display: flex;
		align-items: center;
		gap: 6px;
	}
	.q-t {
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.tag {
		display: flex;
		flex: none;
	}
	.q-s {
		display: block;
		font-size: 12px;
		color: var(--fa);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.medio {
		width: 180px;
		flex: none;
	}
	.spark {
		display: flex;
		align-items: end;
		gap: 2px;
		height: 22px;
	}
	.spark span {
		flex: 1;
		border-radius: 2px;
	}
	.medidor {
		position: relative;
		display: block;
		height: 8px;
		border-radius: 4px;
		background: var(--sf3);
	}
	.medidor > span:first-child {
		position: absolute;
		left: 0;
		top: 0;
		bottom: 0;
		border-radius: 4px;
	}
	.m-marca {
		position: absolute;
		top: -3px;
		bottom: -3px;
		width: 2px;
		border-radius: 1px;
		background: var(--tx);
		opacity: 0.5;
	}
	.pasos {
		display: flex;
		gap: 3px;
	}
	.pasos span {
		flex: 1;
		height: 5px;
		border-radius: 3px;
	}
	.v-l {
		display: flex;
		align-items: center;
		gap: 5px;
		margin-top: 5px;
		font-size: 12px;
		font-weight: 600;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.chip-f {
		flex: none;
		font-size: 11.5px;
		font-weight: 700;
		padding: 4px 8px;
		border-radius: 999px;
		white-space: nowrap;
	}
	.valor {
		flex: none;
		min-width: 80px;
		text-align: right;
	}
	.v-v {
		display: block;
		font-size: 15px;
		font-weight: 600;
		letter-spacing: -0.01em;
		white-space: nowrap;
	}
	.v-s {
		display: block;
		font-size: 11.5px;
		color: var(--mu);
		white-space: nowrap;
	}
	.stepper {
		flex: none;
		display: flex;
		align-items: center;
		gap: 2px;
		padding: 3px;
		border-radius: 999px;
		background: var(--chip);
	}
	.stepper button {
		width: 28px;
		height: 28px;
		border: 0;
		border-radius: 50%;
		background: var(--sf);
		color: var(--tx);
		display: grid;
		place-items: center;
	}
	.stepper span {
		min-width: 30px;
		text-align: center;
		font-size: 13.5px;
		font-weight: 700;
	}
	.toggle {
		flex: none;
		display: flex;
	}
	.accion {
		flex: none;
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 12px;
		border-radius: 999px;
		border: 0;
		font-size: 12.5px;
		font-weight: 600;
		white-space: nowrap;
	}
	.esqueleto span {
		height: 12px;
		border-radius: 6px;
		background: var(--sf3);
		flex: 1;
	}
	.esqueleto span:first-child {
		flex: none;
		width: 42px;
		height: 42px;
		border-radius: 12px;
	}
	.vacio {
		position: absolute;
		inset: 0;
		display: grid;
		place-items: center;
		text-align: center;
		padding: 30px;
	}
	.vacio img {
		object-fit: contain;
	}
	.v-t {
		font-size: 16px;
		font-weight: 600;
		margin-top: 6px;
	}
	.v-x {
		max-width: 340px;
		margin: 6px auto 0;
		font-size: 13px;
		line-height: 1.45;
		color: var(--fa);
		text-wrap: pretty;
	}
	.v-b {
		display: inline-block;
		text-decoration: none;
		margin-top: 10px;
		padding: 9px 16px;
		border-radius: 999px;
		border: 0;
		background: var(--tx);
		color: var(--sf);
		font-size: 13px;
		font-weight: 600;
	}
	.arriba {
		position: absolute;
		right: 16px;
		bottom: 16px;
		width: 40px;
		height: 40px;
		border-radius: 50%;
		border: 0;
		background: var(--tx);
		color: var(--sf);
		display: grid;
		place-items: center;
		box-shadow: 0 10px 24px -8px rgba(42, 20, 12, 0.5);
	}
	.lote {
		position: fixed;
		left: 50%;
		bottom: 84px;
		translate: -50% 0;
		z-index: 60;
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 8px 8px 18px;
		border-radius: 999px;
		background: #2a140c;
		color: #fff;
		box-shadow: 0 24px 50px -16px rgba(42, 20, 12, 0.7);
		max-width: calc(100vw - 16px);
		overflow-x: auto;
		scrollbar-width: none;
	}
	.lo-t {
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
		margin-right: 6px;
	}
	.lote button {
		display: flex;
		align-items: center;
		gap: 6px;
		border: 0;
		font-size: 13px;
		font-weight: 600;
		padding: 9px 13px;
		border-radius: 999px;
		background: rgba(255, 255, 255, 0.1);
		color: #fff;
		white-space: nowrap;
	}
	.lote .lo-x {
		width: 34px;
		height: 34px;
		flex: none;
		padding: 0;
		justify-content: center;
		background: transparent;
		color: #d9c8ba;
	}
	@media (max-width: 899px) {
		.f-in,
		.abrir {
			gap: 10px;
		}
	}
	@media (max-width: 759px) {
		.l-v {
			display: none;
		}
		/* En el celular el nombre manda: los íconos de canales y etiquetas se ven en la ficha. */
		.tag {
			display: none;
		}
		/* Con contador, las unidades ya se ven en el contador. */
		.con-stepper .valor {
			display: none;
		}
		.lote {
			bottom: 92px;
		}
		.chip-f {
			display: none;
		}
		.tip {
			flex-wrap: wrap;
		}
		.tip-t {
			flex-basis: calc(100% - 52px);
		}
		.tip button {
			margin-left: 52px;
		}
	}
</style>
