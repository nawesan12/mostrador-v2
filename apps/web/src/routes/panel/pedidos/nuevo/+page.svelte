<script lang="ts">
	import { SvelteMap } from 'svelte/reactivity';
	import { goto } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import Segmentado from '#lib/components/Segmentado.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { pesos } from '#lib/panel/formato.ts';
	import { rutaDe } from '#lib/panel/navegacion.ts';
	import { vendibles } from '#lib/panel/vendibles.ts';
	import { getShell } from '#lib/panel/shell.svelte.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/**
	 * Crear pedido. El handoff no lo diseña: se arma con las piezas del punto de venta (grilla, ticket,
	 * medios de pago) y del modal de link de pago, que es como se cobra lo que no está pagado.
	 */
	let { data } = $props();
	const ui = getUi();
	const shell = getShell();

	// ---- Cliente ----
	let clienteId = $state<string | null>(null);
	let nombre = $state('');
	let telefono = $state('');
	let buscarCliente = $state('');
	$effect.pre(() => {
		const c = data.clientes.find((x) => x.id === data.cliente);
		if (c) elegir(c);
	});
	const elegido = $derived(data.clientes.find((c) => c.id === clienteId) ?? null);
	const sugeridos = $derived(
		buscarCliente.trim().length < 2
			? []
			: data.clientes
					.filter((c) =>
						`${c.nombre} ${c.telefono ?? ''}`
							.toLowerCase()
							.includes(buscarCliente.trim().toLowerCase())
					)
					.slice(0, 5)
	);
	function elegir(c: (typeof data.clientes)[number]) {
		clienteId = c.id;
		nombre = c.nombre;
		telefono = c.telefono ?? '';
		buscarCliente = '';
	}

	// ---- Productos ----
	let buscar = $state('');
	// Con opciones, cada variante es un renglón ("Remera · M · Negro").
	const todos = $derived(vendibles(data.productos));
	const encontrados = $derived(
		todos
			.filter(
				(p) =>
					!buscar.trim() ||
					`${p.nombre} ${p.sku}`.toLowerCase().includes(buscar.trim().toLowerCase())
			)
			.slice(0, 8)
	);
	const items = new SvelteMap<string, number>();
	const renglones = $derived(
		[...items]
			.map(([clave, q]) => ({ p: todos.find((x) => x.clave === clave)!, q }))
			.filter((x) => x.p)
	);
	function sumar(clave: string, d: number) {
		const p = todos.find((x) => x.clave === clave);
		if (!p) return;
		const q = (items.get(clave) ?? 0) + d;
		if (q > p.stock)
			return ui.toast({
				k: 'warn',
				title: `Quedan ${p.stock} de ${p.nombre}`,
				desc: 'No alcanza el stock'
			});
		if (q <= 0) items.delete(clave);
		else items.set(clave, q);
	}

	// ---- Envío y cobro ----
	const CANALES = [
		['wa', 'WhatsApp'],
		['ig', 'Instagram'],
		['pos', 'En el local'],
		['web', 'Otro']
	] as const;
	let canal = $state<(typeof CANALES)[number][0]>('wa');
	const ENVIOS = [
		['retiro', 'Retira', 0],
		['andreani', 'Andreani', 4200],
		['oca', 'OCA', 3900],
		['moto', 'Moto', 2500]
	] as const;
	let envio = $state<(typeof ENVIOS)[number][0]>('retiro');
	const COBROS = [
		['link', 'Mandar link de pago'],
		['pagado', 'Ya está pagado']
	] as const;
	let cobro = $state<(typeof COBROS)[number][0]>('link');
	const MEDIOS = [
		['transf', 'Transferencia'],
		['efectivo', 'Efectivo'],
		['mp', 'Mercado Pago'],
		['tarjeta', 'Tarjeta']
	] as const;
	let medio = $state<(typeof MEDIOS)[number][0]>('transf');
	let nota = $state('');

	const subtotal = $derived(renglones.reduce((s, x) => s + x.p.precio * x.q, 0));
	const costoEnvio = $derived(ENVIOS.find((e) => e[0] === envio)![2]);
	let creando = $state(false);

	async function crear() {
		if (!renglones.length || creando) return;
		creando = true;
		const r = await fetch('/panel/pedidos/nuevo', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({
				items: renglones.map((x) => ({
					productoId: x.p.id,
					varianteId: x.p.varianteId,
					cantidad: x.q
				})),
				pago: cobro === 'link' ? 'mp' : medio,
				canal,
				envio,
				cliente: nombre.trim() || null,
				clienteId,
				telefono: telefono.trim() || null,
				nota: nota.trim() || null,
				pagado: cobro === 'pagado'
			})
		});
		creando = false;
		if (!r.ok) {
			const m = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			return ui.toast({ k: 'err', title: 'No se creó el pedido', desc: m ?? 'Probá de nuevo.' });
		}
		const { numero, total } = (await r.json()) as { numero: number; total: number };
		// Con link, se vuelve a la lista (el pedido queda arriba, sin pagar) y se arma el link; si ya
		// estaba pagado, se abre el pedido.
		// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta del panel con query
		await goto(cobro === 'link' ? rutaDe('Pedidos') : `${rutaDe('Pedidos')}?pedido=${numero}`);
		if (cobro === 'link')
			shell.abrirLink({
				monto: total + costoEnvio,
				concepto: `Pedido #${numero}`,
				para: nombre.trim() || undefined,
				telefono: telefono.trim() || null,
				pedido: numero
			});
		else
			ui.toast({
				k: 'ok',
				title: `Pedido #${numero} creado`,
				desc: `${pesos(total)} · queda para preparar`
			});
	}
</script>

<Seo
	title="Crear pedido · Panel"
	description="Cargar un pedido a mano"
	path="/panel/pedidos/nuevo"
	noindex
/>

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Crear pedido</h1>
			<div class="c-s">Para lo que te piden por WhatsApp, Instagram o en el mostrador</div>
		</div>
	</div>

	<div class="grilla" use:anim={{ t: 'rise', delay: 60 }}>
		<div class="izq">
			<section class="tarjeta">
				<h2>Cliente</h2>
				{#if elegido}
					<div class="elegido">
						<span class="av"
							>{#if elegido.foto}<img
									src={unsplash(elegido.foto, 80, '&h=80')}
									alt=""
								/>{:else}{elegido.nombre[0]}{/if}</span
						>
						<span class="e-txt"
							><span class="e-n">{elegido.nombre}</span><span class="e-s"
								>{elegido.ciudad}{elegido.telefono ? ` · ${elegido.telefono}` : ''}</span
							></span
						>
						<button
							type="button"
							class="cambiar"
							onclick={() => {
								clienteId = null;
								nombre = '';
								telefono = '';
							}}>Cambiar</button
						>
					</div>
				{:else}
					<label class="buscar"
						><Icon icon={I.phMagnifyingGlass} size={16} /><input
							bind:value={buscarCliente}
							placeholder="Buscá un cliente por nombre o teléfono"
							aria-label="Buscar cliente"
						/></label
					>
					{#if sugeridos.length}
						<div class="sugeridos">
							{#each sugeridos as c (c.id)}
								<button type="button" onclick={() => elegir(c)}
									><span class="s-n">{c.nombre}</span><span class="s-s">{c.ciudad}</span></button
								>
							{/each}
						</div>
					{/if}
					<div class="dos">
						<label class="field"
							>O uno nuevo<input
								class="input"
								bind:value={nombre}
								maxlength="80"
								placeholder="Nombre"
							/></label
						>
						<label class="field"
							>WhatsApp<input
								class="input"
								bind:value={telefono}
								maxlength="30"
								inputmode="tel"
								placeholder="11 5555 5555"
							/></label
						>
					</div>
				{/if}
			</section>

			<section class="tarjeta">
				<h2>Productos</h2>
				<label class="buscar"
					><Icon icon={I.phMagnifyingGlass} size={16} /><input
						bind:value={buscar}
						placeholder="Buscá por nombre o código"
						aria-label="Buscar producto"
					/></label
				>
				<div class="productos">
					{#each encontrados as p (p.clave)}
						<button
							type="button"
							class="prod"
							disabled={!p.stock}
							onclick={() => sumar(p.clave, 1)}
							aria-label="Sumar {p.nombre}"
						>
							{#if p.foto}<img src={unsplash(p.foto, 96, '&h=96')} alt="" />{/if}
							<span class="p-txt"
								><span class="p-n">{p.nombre}</span><span class="p-s"
									>{pesos(p.precio)} · {p.stock} u</span
								></span
							>
							<span class="mas"><Icon icon={I.phPlusBold} size={13} /></span>
						</button>
					{:else}
						<div class="vacio">
							{buscar.trim()
								? `Nada con «${buscar.trim()}».`
								: 'Todavía no tenés productos para vender.'}
						</div>
					{/each}
				</div>
			</section>

			<section class="tarjeta">
				<h2>Entrega</h2>
				<div class="opcion">
					<span>Por dónde llegó</span><Segmentado
						opciones={CANALES}
						bind:value={canal}
						label="Canal"
					/>
				</div>
				<div class="opcion">
					<span>Envío</span><Segmentado
						opciones={ENVIOS.map(([k, t]) => [k, t] as const)}
						bind:value={envio}
						label="Envío"
					/>
				</div>
				<label class="field"
					>Nota<textarea
						class="input"
						bind:value={nota}
						maxlength="300"
						rows="2"
						placeholder="Para el que arma el paquete"></textarea></label
				>
			</section>
		</div>

		<aside class="resumen">
			<div class="r-cab">Resumen</div>
			<div class="r-items">
				{#each renglones as x (x.p.clave)}
					<div class="item">
						<span class="i-txt"
							><span class="i-n">{x.p.nombre}</span><span class="i-s">{pesos(x.p.precio)} c/u</span
							></span
						>
						<span class="i-q">
							<button
								type="button"
								aria-label="Restar {x.p.nombre}"
								onclick={() => sumar(x.p.clave, -1)}><Icon icon={I.phMinusBold} size={12} /></button
							>
							<b>{x.q}</b>
							<button
								type="button"
								class="q-mas"
								aria-label="Sumar {x.p.nombre}"
								onclick={() => sumar(x.p.clave, 1)}><Icon icon={I.phPlusBold} size={12} /></button
							>
						</span>
					</div>
				{:else}
					<div class="r-vacio">Todavía no sumaste productos</div>
				{/each}
			</div>
			<div class="r-pie">
				<div class="linea"><span>Subtotal</span><span>{pesos(subtotal)}</span></div>
				<div class="linea">
					<span>Envío</span><span>{costoEnvio ? pesos(costoEnvio) : 'Sin cargo'}</span>
				</div>
				<div class="total"><span>Total</span><b>{pesos(subtotal + costoEnvio)}</b></div>
				<Segmentado opciones={COBROS} bind:value={cobro} label="Cómo se cobra" />
				{#if cobro === 'pagado'}
					<label class="field"
						>Cómo pagó<select class="select" bind:value={medio}
							>{#each MEDIOS as [k, t] (k)}<option value={k}>{t}</option>{/each}</select
						></label
					>
				{/if}
				<button type="button" class="crear" disabled={!renglones.length || creando} onclick={crear}
					>{cobro === 'link' ? 'Crear y armar el link' : 'Crear pedido'}</button
				>
				<div class="nota">
					{cobro === 'link'
						? 'Queda sin pagar hasta que use el link; ahí pasa a preparar solo.'
						: 'Queda para preparar y descuenta stock.'}
				</div>
			</div>
		</aside>
	</div>
</div>

<style>
	.pagina {
		display: flex;
		flex-direction: column;
		gap: 18px;
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
	.grilla {
		display: grid;
		grid-template-columns: minmax(0, 1fr) 360px;
		gap: 16px;
		align-items: start;
	}
	.izq {
		display: flex;
		flex-direction: column;
		gap: 14px;
		min-width: 0;
	}
	.tarjeta {
		padding: 18px;
		border-radius: 22px;
		background: var(--sf);
		box-shadow: var(--shd);
		display: flex;
		flex-direction: column;
		gap: 12px;
	}
	h2 {
		margin: 0;
		font-size: 16px;
		font-weight: 600;
	}
	.buscar {
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 11px 14px;
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
		font-size: 14px;
		color: var(--tx);
	}
	.sugeridos {
		display: flex;
		flex-direction: column;
		border: 1px solid var(--ln);
		border-radius: 14px;
		overflow: hidden;
	}
	.sugeridos button {
		display: flex;
		justify-content: space-between;
		gap: 10px;
		padding: 10px 14px;
		border: 0;
		border-bottom: 1px solid var(--ln);
		background: var(--sf);
		color: var(--tx);
		text-align: left;
		font-size: 14px;
	}
	.sugeridos button:last-child {
		border-bottom: 0;
	}
	.s-n {
		font-weight: 600;
	}
	.s-s {
		color: var(--mu);
		font-size: 13px;
	}
	.dos {
		display: grid;
		grid-template-columns: 1fr 1fr;
		gap: 10px;
	}
	.elegido {
		display: flex;
		align-items: center;
		gap: 12px;
	}
	.av {
		width: 42px;
		height: 42px;
		border-radius: 50%;
		overflow: hidden;
		flex: none;
		display: grid;
		place-items: center;
		background: var(--sf3);
		font-weight: 600;
	}
	.av img {
		width: 100%;
		height: 100%;
		object-fit: cover;
	}
	.e-txt {
		flex: 1;
		min-width: 0;
	}
	.e-n {
		display: block;
		font-size: 15px;
		font-weight: 600;
	}
	.e-s {
		display: block;
		font-size: 13px;
		color: var(--mu);
	}
	.cambiar {
		padding: 8px 14px;
		border-radius: 999px;
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		font-size: 13px;
		font-weight: 600;
	}
	.productos {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
		gap: 8px;
	}
	.prod {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 8px;
		border-radius: 16px;
		border: 1px solid var(--ln);
		background: var(--sf);
		color: var(--tx);
		text-align: left;
	}
	.prod:disabled {
		opacity: 0.5;
	}
	.prod img {
		width: 44px;
		height: 44px;
		border-radius: 11px;
		object-fit: cover;
		flex: none;
	}
	.p-txt {
		flex: 1;
		min-width: 0;
	}
	.p-n {
		display: block;
		font-size: 13.5px;
		font-weight: 600;
		overflow: hidden;
		text-overflow: ellipsis;
		white-space: nowrap;
	}
	.p-s {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.mas {
		width: 28px;
		height: 28px;
		border-radius: 50%;
		background: var(--tx);
		color: var(--sf);
		display: grid;
		place-items: center;
		flex: none;
	}
	.vacio {
		padding: 16px;
		color: var(--mu);
		font-size: 14px;
	}
	.opcion {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 10px;
		flex-wrap: wrap;
		font-size: 14px;
	}
	.resumen {
		position: sticky;
		top: 80px;
		border-radius: 24px;
		background: var(--sf);
		box-shadow: var(--shd);
		overflow: hidden;
	}
	.r-cab {
		padding: 16px 18px;
		border-bottom: 1px solid var(--ln);
		font-size: 16px;
		font-weight: 600;
	}
	.r-items {
		padding: 4px 18px;
		max-height: 300px;
		overflow: auto;
	}
	.item {
		display: flex;
		align-items: center;
		gap: 10px;
		padding: 10px 0;
		border-bottom: 1px solid var(--ln);
	}
	.i-txt {
		flex: 1;
		min-width: 0;
	}
	.i-n {
		display: block;
		font-size: 14px;
		font-weight: 600;
	}
	.i-s {
		display: block;
		font-size: 12.5px;
		color: var(--mu);
	}
	.i-q {
		display: flex;
		align-items: center;
		gap: 6px;
	}
	.i-q button {
		width: 28px;
		height: 28px;
		border-radius: 50%;
		border: 1px solid var(--ln2);
		background: transparent;
		color: var(--tx);
		display: grid;
		place-items: center;
	}
	.i-q .q-mas {
		border: 0;
		background: var(--tx);
		color: var(--sf);
	}
	.i-q b {
		min-width: 16px;
		text-align: center;
	}
	.r-vacio {
		padding: 24px 0;
		text-align: center;
		color: var(--mu);
		font-size: 14px;
	}
	.r-pie {
		padding: 14px 18px;
		display: flex;
		flex-direction: column;
		gap: 10px;
		border-top: 1px solid var(--ln);
		background: var(--sf2);
	}
	.linea {
		display: flex;
		justify-content: space-between;
		font-size: 14px;
		color: var(--mu);
	}
	.total {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
	}
	.total span {
		font-size: 14px;
		color: var(--mu);
	}
	.total b {
		font-size: 30px;
		font-weight: 600;
		letter-spacing: -0.04em;
	}
	.crear {
		padding: 15px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #c24e25, #b0441d);
		color: #fff;
		font-size: 15px;
		font-weight: 600;
		box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.3);
	}
	.crear:disabled {
		background: #c9b6a6;
	}
	.nota {
		font-size: 12.5px;
		color: var(--mu);
		text-align: center;
		text-wrap: pretty;
	}
	@media (max-width: 999px) {
		.grilla {
			grid-template-columns: minmax(0, 1fr);
		}
		.resumen {
			position: static;
		}
	}
	@media (max-width: 560px) {
		.dos {
			grid-template-columns: 1fr;
		}
		.opcion {
			flex-direction: column;
			align-items: stretch;
		}
		.tarjeta .opcion :global(.seg) {
			display: flex;
			overflow-x: auto;
			scrollbar-width: none;
		}
		.tarjeta .opcion :global(.seg button) {
			padding: 7px 10px;
			flex: 1;
			white-space: nowrap;
		}
	}
</style>
