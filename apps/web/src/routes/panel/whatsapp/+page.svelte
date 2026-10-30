<script lang="ts">
	import { resolve } from '$app/paths';
	import { deslizable } from '#lib/panel/deslizable.ts';
	import { tick } from 'svelte';
	import { goto, invalidateAll } from '$app/navigation';
	import { page } from '$app/state';
	import Icon from '#lib/components/Icon.svelte';
	import MarcaIcono from '#lib/components/MarcaIcono.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { cara, unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { iniciales } from '#lib/panel/listas/clientes.ts';
	import { ESTADOS } from '#lib/marcas.ts';
	import { rutaDe } from '#lib/panel/navegacion.ts';
	import { SITE } from '#lib/site.ts';
	import { getShell } from '#lib/panel/shell.svelte.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Bandeja (design/PanelVentas.dc.html, WhatsApp): conversaciones, hilo y ficha del cliente. */
	let { data } = $props();
	const ui = getUi();
	const shell = getShell();
	const tienda = $derived(page.data.tienda as { nombre: string; slug: string });

	type Filtro = 'todas' | 'vos' | 'hornero' | 'resuelto';
	let filtro = $state<Filtro>('todas');
	let buscar = $state('');
	const FILTROS: [Filtro, string][] = [
		['todas', 'Todas'],
		['vos', 'Te esperan'],
		['hornero', 'Tincho'],
		['resuelto', 'Resueltas']
	];
	const lista = $derived(
		data.conversaciones.filter(
			(c) =>
				(filtro === 'todas' || c.atiende === filtro) &&
				(!buscar.trim() || c.nombre.toLowerCase().includes(buscar.trim().toLowerCase()))
		)
	);
	const ETIQUETA = {
		hornero: ['Tincho', 'rgba(210,88,47,.12)', '#C24E25'],
		vos: ['Te espera', 'rgba(227,169,59,.2)', '#9A6A10'],
		resuelto: ['Resuelto', 'rgba(47,143,99,.12)', '#2F8F63']
	} as const;

	const c = $derived(data.abierta);
	let ancho = $state(1440);
	const angosto = $derived(ancho < 1000);
	const celular = $derived(ancho < 760);
	// En el celular se ve la lista o el hilo.
	let enHilo = $derived(data.elegida);

	let hilo = $state<HTMLElement>();
	async function bajar() {
		await tick();
		hilo?.scrollTo({ top: hilo.scrollHeight });
	}
	$effect(() => {
		void c?.mensajes.length;
		void bajar();
	});

	async function abrir(id: string) {
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- misma página, cambia la conversación
		await goto(`?c=${id}`, { replace: true, reset: false });
		enHilo = true;
		const conv = data.conversaciones.find((x) => x.id === id);
		if (conv?.sinLeer) await post(id, { leida: true });
	}

	async function post(id: string, cuerpo: object) {
		const r = await fetch(`/panel/whatsapp/${id}`, {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify(cuerpo)
		});
		await invalidateAll();
		return r.ok;
	}

	let borrador = $state('');
	async function mandar(texto: string) {
		if (!c || !texto.trim()) return;
		borrador = '';
		if (!(await post(c.id, { texto })))
			ui.toast({ k: 'err', title: 'No se mandó', desc: 'Probá de nuevo.' });
	}
	async function tomar() {
		if (!c) return;
		const tomada = !c.tomada;
		await post(c.id, { tomada });
		ui.toast({
			k: 'info',
			title: tomada ? 'Tomaste la conversación' : 'Tincho vuelve a atender',
			desc: tomada
				? `Tincho queda en pausa con ${c.nombre.split(' ')[0]}`
				: 'Te avisa si necesita algo'
		});
	}

	const ACCIONES = $derived(
		c
			? ([
					[
						I.phBookOpenTextFill,
						'Catálogo',
						`Te paso nuestro catálogo: ${SITE.replace(/^https?:\/\//, '')}/t/${tienda.slug}`
					],
					[I.phLinkBold, 'Link de pago', ''],
					[
						I.phTruckFill,
						'Calcular envío',
						`El envío a ${c.cliente.ciudad} sale ${pesos(4200)} y llega en 48 h.`
					],
					[I.phLightningFill, 'Respuestas rápidas', '¡Gracias por escribirnos! ¿En qué te ayudo?']
				] as const)
			: []
	);
	const hora = (iso: string) =>
		new Intl.DateTimeFormat('es-AR', {
			timeZone: 'America/Argentina/Buenos_Aires',
			hour: '2-digit',
			minute: '2-digit',
			hourCycle: 'h23'
		}).format(new Date(iso));
	const hace = (iso: string) => {
		const m = Math.max(1, Math.round((Date.now() - Date.parse(iso)) / 60_000));
		return m < 60 ? `${m} min` : m < 1440 ? `${Math.round(m / 60)} h` : `${Math.round(m / 1440)} d`;
	};
	const anio = (iso: string) => new Date(iso).getFullYear();
</script>

<svelte:window bind:innerWidth={ancho} />
<Seo
	title="Bandeja · Panel"
	description="Conversaciones de tu tienda"
	path="/panel/whatsapp"
	noindex
/>

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Bandeja</h1>
			<div class="c-s">Las preguntas de Mercado Libre hoy; WhatsApp e Instagram, pronto</div>
		</div>
		{#if data.conversaciones.length}
			<span class="atendiendo"
				><span class="punto"></span>Tincho atendiendo · {data.atendidasHoy} hoy</span
			>
		{/if}
	</div>

	{#if !data.conversaciones.length}
		<div class="sin-canales" use:anim={{ t: 'rise', delay: 60 }}>
			<Icon icon={I.phChatsCircleBold} size={30} />
			<div class="sc-t">Todavía no hay mensajes</div>
			<p>
				Acá llegan los mensajes de tus canales. Las preguntas de Mercado Libre entran cuando lo
				conectás; WhatsApp e Instagram llegan&nbsp;pronto.
			</p>
			<a class="btn btn-secondary" href={resolve('/panel/integraciones')}>Ver integraciones</a>
		</div>
	{:else}
		<div class="bandeja" class:angosto class:celular use:anim={{ t: 'rise', delay: 60 }}>
			{#if !celular || !enHilo}
				<div class="lista">
					<div class="l-cab">
						<label class="buscar"
							><Icon icon={I.phMagnifyingGlass} size={15} /><input
								bind:value={buscar}
								placeholder="Buscar conversación"
								aria-label="Buscar conversación"
							/></label
						>
						<div class="filtros" role="group" aria-label="Filtrar" use:deslizable>
							{#each FILTROS as [id, t] (id)}
								<button type="button" aria-pressed={filtro === id} onclick={() => (filtro = id)}
									>{t}<span
										>{id === 'todas'
											? data.conversaciones.length
											: data.conversaciones.filter((x) => x.atiende === id).length}</span
									></button
								>
							{/each}
						</div>
					</div>
					<div class="convs" data-lenis-prevent>
						{#each lista as x (x.id)}
							{@const e = ETIQUETA[x.atiende]}
							<button
								type="button"
								class="conv"
								class:on={x.id === c?.id}
								onclick={() => abrir(x.id)}
							>
								<span class="av">
									{#if x.foto}<img src={cara(x.foto)} alt="" />{:else}<span class="ini"
											>{iniciales(x.nombre)}</span
										>{/if}
									<span class="canal"
										><MarcaIcono k={x.canal} size={19} radius={10} fondo="var(--sf)" /></span
									>
								</span>
								<span class="c-cuerpo">
									<span class="c-fila"
										><span class="c-n">{x.nombre}</span><span class="c-h">{hace(x.creado)}</span
										></span
									>
									<span class="c-m">{x.ultimo}</span>
									<span class="c-pie"
										><span class="tag" style:background={e[1]} style:color={e[2]}>{e[0]}</span
										>{#if x.sinLeer}<span class="sin">{x.sinLeer}</span>{/if}</span
									>
								</span>
							</button>
						{:else}
							<div class="vacio">No hay conversaciones con ese filtro.</div>
						{/each}
					</div>
				</div>
			{/if}

			{#if c && (!celular || enHilo)}
				<div class="hilo-col">
					<div class="h-cab">
						{#if celular}<button
								type="button"
								class="volver"
								aria-label="Volver a la lista"
								onclick={() => (enHilo = false)}><Icon icon={I.phCaretLeftBold} size={14} /></button
							>{/if}
						{#if c.foto}<img class="h-av" src={cara(c.foto)} alt="" />{:else}<span class="h-av ini"
								>{iniciales(c.nombre)}</span
							>{/if}
						<span class="h-txt">
							<span class="h-n"
								>{c.nombre}<MarcaIcono k={c.canal} size={16} radius={0} fondo="transparent" /></span
							>
							<span class="h-s"
								>{c.tomada
									? 'Atendés vos'
									: c.atiende === 'hornero'
										? 'Tincho está respondiendo…'
										: 'En línea'}</span
							>
						</span>
						<button type="button" class="tomar" class:devolver={c.tomada} onclick={tomar}
							><Icon icon={c.tomada ? I.phSparkleFill : I.phHandFill} size={14} />{c.tomada
								? 'Devolver a Tincho'
								: 'Tomar'}</button
						>
					</div>
					<div class="hilo" bind:this={hilo} data-lenis-prevent>
						<div class="dia">Hoy</div>
						{#each c.mensajes as m, i (i)}
							{#if m.producto}
								<div class="tarjeta-prod">
									<div class="tp-in">
										{#if m.producto.foto}<img
												src={unsplash(m.producto.foto, 480, '&h=220')}
												alt=""
											/>{/if}
										<div class="tp-pie">
											<span
												><span class="tp-n">{m.producto.nombre}</span><span class="tp-s"
													>Quedan {m.producto.stock}</span
												></span
											>
											<b>{pesos(m.producto.precio)}</b>
										</div>
									</div>
								</div>
							{:else if m.de === 'cliente'}
								<div class="msg cliente">{m.texto}<span class="hora">{hora(m.en)}</span></div>
							{:else}
								<div class="msg-yo">
									{#if m.de === 'hornero'}<span class="ai">✦ Tincho</span>{/if}
									<span class="msg yo">{m.texto}<span class="hora">{hora(m.en)} ✓✓</span></span>
								</div>
							{/if}
						{/each}
					</div>
					{#if c.sugerencia}
						<div class="sugerencia">
							<Icon
								icon={I.phSparkleFill}
								size={15}
								style="color:#b0441d;margin-top:2px;flex:none"
							/>
							<span><b>Sugerencia:</b> {c.sugerencia}</span>
							<button type="button" onclick={() => mandar(c!.sugerencia!)}>Enviar</button>
						</div>
					{/if}
					<div class="escribir">
						<div class="rapidas" use:deslizable>
							{#each ACCIONES as [icono, t, texto] (t)}
								<button
									type="button"
									onclick={() =>
										texto
											? mandar(texto)
											: shell.abrirLink({ para: c!.nombre, concepto: `Pedido de ${c!.nombre}` })}
									><Icon icon={icono} size={14} style="color:#b0441d" />{t}</button
								>
							{/each}
						</div>
						<form
							class="campo"
							onsubmit={(e) => {
								e.preventDefault();
								void mandar(borrador);
							}}
						>
							<input
								bind:value={borrador}
								placeholder="Escribí un mensaje…"
								aria-label="Escribí un mensaje"
								maxlength="2000"
							/>
							<button class="mandar" aria-label="Mandar" disabled={!borrador.trim()}
								><Icon icon={I.phPaperPlaneTiltFill} size={16} /></button
							>
						</form>
					</div>
				</div>
			{/if}

			{#if c && !angosto}
				<div class="ficha" data-lenis-prevent>
					<div class="f-cab">
						{#if c.foto}<img src={cara(c.foto)} alt="" />{:else}<span class="ini grande"
								>{iniciales(c.nombre)}</span
							>{/if}
						<div class="f-n">{c.nombre}</div>
						<div class="f-s">{c.cliente.ciudad} · cliente desde {anio(c.cliente.desde)}</div>
					</div>
					<div class="f-kpis">
						<div><span>Gastó</span><b>{pesos(c.cliente.gastado)}</b></div>
						<div><span>Pedidos</span><b>{c.cliente.pedidos}</b></div>
					</div>
					{#if c.cliente.etiquetas.length}
						<div>
							<div class="f-t">Etiquetas</div>
							<div class="f-tags">
								{#each c.cliente.etiquetas as t (t)}<span>{t}</span>{/each}
							</div>
						</div>
					{/if}
					{#if c.cliente.ultimoPedido}
						{@const u = c.cliente.ultimoPedido}
						<div>
							<div class="f-t">Último pedido</div>
							<!-- eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta del panel con query -->
							<a class="f-ped" href="{rutaDe('Pedidos')}?pedido={u.numero}"
								><b>#{u.numero}</b> · {u.resumen} · {pesos(u.total)}<span
									style:color={ESTADOS[u.estado]?.[1]}>{u.estado}</span
								></a
							>
						</div>
					{/if}
					{#if c.cliente.sabe}
						<div class="f-sabe"><b>✦ Tincho sabe que</b> {c.cliente.sabe}</div>
					{/if}
					<button
						type="button"
						class="f-crear"
						onclick={() =>
							// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta del panel con query
							goto(`/panel/pedidos/nuevo${c!.clienteId ? `?cliente=${c!.clienteId}` : ''}`)}
						>Crear pedido para {c.nombre.split(' ')[0]}</button
					>
				</div>
			{/if}
		</div>
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
		gap: 14px;
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
	.atendiendo {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 9px 14px;
		border-radius: 999px;
		background: var(--sf);
		box-shadow: var(--sh);
		font-size: 13px;
		font-weight: 600;
		white-space: nowrap;
	}
	.punto {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: #2f8f63;
	}
	.bandeja {
		display: grid;
		grid-template-columns: 320px minmax(0, 1fr) 280px;
		height: 680px;
		border-radius: 24px;
		background: var(--sf);
		box-shadow: var(--shd);
		overflow: hidden;
	}
	.bandeja.angosto {
		grid-template-columns: 300px minmax(0, 1fr);
	}
	.bandeja.celular {
		grid-template-columns: minmax(0, 1fr);
		height: calc(100dvh - 240px);
		min-height: 480px;
	}
	.lista {
		border-right: 1px solid var(--ln);
		display: flex;
		flex-direction: column;
		min-height: 0;
	}
	.celular .lista {
		border-right: 0;
	}
	.l-cab {
		padding: 12px;
		display: flex;
		flex-direction: column;
		gap: 10px;
		border-bottom: 1px solid var(--ln);
	}
	.buscar {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 9px 12px;
		border-radius: 999px;
		background: var(--sf3);
		color: var(--mu);
	}
	.buscar input {
		flex: 1;
		min-width: 0;
		border: 0;
		outline: 0;
		background: transparent;
		font-size: 13.5px;
		color: var(--tx);
	}
	.filtros {
		display: flex;
		gap: 6px;
		overflow-x: auto;
		scrollbar-width: none;
	}
	.filtros button {
		border: 0;
		white-space: nowrap;
		font-size: 12.5px;
		font-weight: 600;
		padding: 6px 11px;
		border-radius: 999px;
		background: var(--chip);
		color: var(--mu);
	}
	.filtros button[aria-pressed='true'] {
		background: var(--tx);
		color: var(--sf);
	}
	.filtros span {
		opacity: 0.6;
		margin-left: 4px;
	}
	.convs {
		flex: 1;
		overflow: auto;
	}
	.conv {
		width: 100%;
		display: flex;
		gap: 11px;
		padding: 12px;
		border: 0;
		border-bottom: 1px solid var(--ln);
		background: transparent;
		color: var(--tx);
		text-align: left;
	}
	.conv.on {
		background: var(--chip);
	}
	.av {
		position: relative;
		flex: none;
	}
	.av img,
	.av .ini {
		width: 42px;
		height: 42px;
		border-radius: 50%;
		object-fit: cover;
		display: block;
	}
	.ini {
		display: grid;
		place-items: center;
		background: var(--sf3);
		color: var(--mu);
		font-size: 13px;
		font-weight: 700;
	}
	.canal {
		position: absolute;
		right: -3px;
		bottom: -3px;
		border-radius: 50%;
		box-shadow: var(--sh);
		display: block;
	}
	.c-cuerpo {
		flex: 1;
		min-width: 0;
	}
	.c-fila {
		display: flex;
		justify-content: space-between;
		gap: 8px;
	}
	.c-n {
		font-size: 14px;
		font-weight: 600;
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}
	.c-h {
		font-size: 11.5px;
		color: var(--fa);
		flex: none;
	}
	.c-m {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
		margin-top: 1px;
	}
	.c-pie {
		display: flex;
		gap: 5px;
		margin-top: 5px;
		align-items: center;
	}
	.tag {
		font-size: 11px;
		font-weight: 600;
		padding: 2px 7px;
		border-radius: 999px;
	}
	.sin {
		margin-left: auto;
		min-width: 18px;
		height: 18px;
		padding: 0 5px;
		border-radius: 999px;
		background: #2f8f63;
		color: #fff;
		font-size: 10.5px;
		font-weight: 700;
		display: grid;
		place-items: center;
	}
	.sin-canales {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 10px;
		max-width: 460px;
		margin: 24px auto;
		padding: 32px 24px;
		border-radius: 20px;
		background: var(--sf);
		text-align: center;
		color: var(--mu);
	}
	.sc-t {
		font-size: 18px;
		font-weight: 700;
		color: var(--tx);
	}
	.sin-canales p {
		margin: 0 0 6px;
		font-size: 14.5px;
		line-height: 1.5;
	}
	.vacio {
		padding: 24px;
		font-size: 13.5px;
		color: var(--mu);
		text-align: center;
	}
	.hilo-col {
		display: flex;
		flex-direction: column;
		min-height: 0;
		min-width: 0;
	}
	.h-cab {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 14px;
		border-bottom: 1px solid var(--ln);
	}
	.volver {
		width: 34px;
		height: 34px;
		border: 0;
		border-radius: 50%;
		background: var(--chip);
		color: var(--tx);
		display: grid;
		place-items: center;
		flex: none;
	}
	.h-av {
		width: 38px;
		height: 38px;
		border-radius: 50%;
		object-fit: cover;
		flex: none;
	}
	.h-txt {
		flex: 1;
		min-width: 0;
	}
	.h-n {
		display: flex;
		align-items: center;
		gap: 6px;
		font-size: 15px;
		font-weight: 600;
		white-space: nowrap;
	}
	.h-s {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.tomar {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 8px 13px;
		border-radius: 999px;
		border: 0;
		background: var(--tx);
		color: var(--sf);
		font-size: 12.5px;
		font-weight: 600;
		white-space: nowrap;
	}
	.tomar.devolver {
		background: rgba(210, 88, 47, 0.12);
		color: #b0441d;
	}
	.hilo {
		flex: 1;
		overflow: auto;
		padding: 16px;
		display: flex;
		flex-direction: column;
		gap: 8px;
		background: var(--sf3);
	}
	.dia {
		align-self: center;
		font-size: 11.5px;
		font-weight: 600;
		color: var(--fa);
		padding: 4px 10px;
		border-radius: 999px;
		background: var(--sf);
	}
	.msg {
		font-size: 14px;
		line-height: 1.45;
		padding: 9px 13px;
		overflow-wrap: anywhere;
	}
	.msg.cliente {
		align-self: flex-start;
		max-width: 76%;
		background: var(--sf);
		border-radius: 16px 16px 16px 4px;
		box-shadow: var(--sh);
	}
	.msg-yo {
		align-self: flex-end;
		max-width: 78%;
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: 2px;
	}
	.ai {
		font-size: 10.5px;
		font-weight: 600;
		color: #b0441d;
	}
	.msg.yo {
		background: #fce3d6;
		color: #2a140c;
		border-radius: 16px 16px 4px 16px;
	}
	.hora {
		display: block;
		text-align: right;
		font-size: 10.5px;
		color: var(--fa);
		margin-top: 2px;
	}
	.msg.yo .hora {
		color: #9a7a66;
	}
	.tarjeta-prod {
		align-self: flex-end;
		width: 240px;
		border-radius: 16px;
		background: #fce3d6;
		padding: 6px;
	}
	.tp-in {
		border-radius: 12px;
		background: #fff;
		color: #2a140c;
		overflow: hidden;
	}
	.tp-in img {
		width: 100%;
		height: 100px;
		object-fit: cover;
		display: block;
	}
	.tp-pie {
		padding: 9px 11px;
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 8px;
	}
	.tp-n {
		display: block;
		font-size: 13px;
		font-weight: 600;
	}
	.tp-s {
		display: block;
		font-size: 11.5px;
		color: #7a6252;
	}
	.tp-pie b {
		color: #b0441d;
		font-size: 14px;
		white-space: nowrap;
	}
	.sugerencia {
		margin: 10px 12px 0;
		padding: 10px 12px;
		border-radius: 14px;
		background: rgba(210, 88, 47, 0.09);
		border: 1px solid rgba(210, 88, 47, 0.22);
		display: flex;
		gap: 10px;
		align-items: flex-start;
		font-size: 13.5px;
		line-height: 1.45;
	}
	.sugerencia span {
		flex: 1;
	}
	.sugerencia button {
		padding: 7px 12px;
		border-radius: 999px;
		border: 0;
		background: #b0441d;
		color: #fff;
		font-size: 12.5px;
		font-weight: 600;
		white-space: nowrap;
	}
	.escribir {
		padding: 10px 12px 12px;
		display: flex;
		flex-direction: column;
		gap: 8px;
	}
	.rapidas {
		display: flex;
		gap: 6px;
		overflow-x: auto;
		scrollbar-width: none;
	}
	.rapidas button {
		display: flex;
		align-items: center;
		gap: 6px;
		padding: 7px 11px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--sf);
		color: var(--tx);
		font-size: 12.5px;
		font-weight: 600;
		white-space: nowrap;
	}
	.campo {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 6px 6px 6px 14px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: var(--bg);
	}
	.campo input {
		flex: 1;
		min-width: 0;
		border: 0;
		outline: 0;
		background: transparent;
		font-size: 14px;
		font-weight: 500;
		color: var(--tx);
	}
	.mandar {
		width: 36px;
		height: 36px;
		border: 0;
		border-radius: 50%;
		background: #2f8f63;
		color: #fff;
		display: grid;
		place-items: center;
	}
	.mandar:disabled {
		opacity: 0.5;
	}
	.ficha {
		border-left: 1px solid var(--ln);
		overflow: auto;
		padding: 16px;
		display: flex;
		flex-direction: column;
		gap: 14px;
	}
	.f-cab {
		text-align: center;
	}
	.f-cab img,
	.f-cab .grande {
		width: 72px;
		height: 72px;
		border-radius: 50%;
		object-fit: cover;
		margin: 0 auto;
		font-size: 22px;
	}
	.f-n {
		margin-top: 8px;
		font-size: 16px;
		font-weight: 600;
	}
	.f-s {
		font-size: 12.5px;
		color: var(--mu);
	}
	.f-kpis {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 8px;
	}
	.f-kpis div {
		padding: 10px;
		border-radius: 14px;
		background: var(--sf3);
	}
	.f-kpis span {
		display: block;
		font-size: 11.5px;
		color: var(--mu);
	}
	.f-kpis b {
		display: block;
		font-size: 17px;
		font-weight: 600;
	}
	.f-t {
		font-size: 12.5px;
		font-weight: 600;
		color: var(--mu);
		margin-bottom: 6px;
	}
	.f-tags {
		display: flex;
		flex-wrap: wrap;
		gap: 5px;
	}
	.f-tags span {
		font-size: 12px;
		font-weight: 600;
		padding: 4px 9px;
		border-radius: 999px;
		background: var(--chip);
	}
	.f-ped {
		display: block;
		padding: 10px 12px;
		border-radius: 14px;
		border: 1px solid var(--ln);
		font-size: 13.5px;
		color: var(--tx);
	}
	.f-ped span {
		display: block;
		font-size: 12px;
		font-weight: 600;
		margin-top: 2px;
	}
	.f-sabe {
		padding: 12px;
		border-radius: 14px;
		background: rgba(210, 88, 47, 0.08);
		font-size: 13px;
		line-height: 1.45;
	}
	.f-sabe b {
		color: #b0441d;
	}
	.f-crear {
		padding: 11px;
		border-radius: 999px;
		border: 0;
		background: var(--tx);
		color: var(--sf);
		font-size: 13.5px;
		font-weight: 600;
	}
</style>
