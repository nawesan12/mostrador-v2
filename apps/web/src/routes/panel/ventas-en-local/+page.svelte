<script lang="ts">
	import { deslizable } from '#lib/panel/deslizable.ts';
	import { SvelteMap } from 'svelte/reactivity';
	import { goto, invalidateAll } from '$app/navigation';
	import Icon from '#lib/components/Icon.svelte';
	import Modal from '#lib/components/Modal.svelte';
	import Seo from '#lib/components/Seo.svelte';
	import * as I from '#lib/icons.ts';
	import { unsplash } from '#lib/landing/formato.ts';
	import { anim } from '#lib/panel/anim.ts';
	import { horaMin, pesos } from '#lib/panel/formato.ts';
	import { rutaDe } from '#lib/panel/navegacion.ts';
	import { vendibles } from '#lib/panel/vendibles.ts';
	import { getUi } from '#lib/ui.svelte.ts';

	/** Ventas en el local (design/PanelVentas.dc.html, POS): grilla, ticket, medios de pago y Cobrar. */
	let { data } = $props();
	const ui = getUi();

	// Con opciones, cada variante es una tarjeta ("Remera · M · Negro").
	const productos = $derived(vendibles(data.productos.filter((p) => p.publicado || p.stock > 0)));
	const categorias = $derived(['Todo', ...new Set(productos.map((p) => p.categoria))]);
	let categoria = $state('Todo');
	let buscar = $state('');
	const visibles = $derived(
		productos.filter(
			(p) =>
				(categoria === 'Todo' || p.categoria === categoria) &&
				(!buscar.trim() ||
					`${p.nombre} ${p.sku}`.toLowerCase().includes(buscar.trim().toLowerCase()))
		)
	);

	// ---- Ticket ----
	const ticket = new SvelteMap<string, number>();
	const renglones = $derived(
		[...ticket]
			.map(([clave, q]) => ({ p: productos.find((x) => x.clave === clave)!, q }))
			.filter((x) => x.p)
	);
	const total = $derived(renglones.reduce((s, x) => s + x.p.precio * x.q, 0));
	function sumar(clave: string, d: number) {
		const p = productos.find((x) => x.clave === clave);
		if (!p) return;
		const q = (ticket.get(clave) ?? 0) + d;
		if (q > p.stock)
			return ui.toast({
				k: 'warn',
				title: `Quedan ${p.stock} de ${p.nombre}`,
				desc: 'No alcanza el stock'
			});
		if (q <= 0) ticket.delete(clave);
		else ticket.set(clave, q);
	}
	/** Con lector: el código entra como texto y Enter. Si coincide un SKU, se suma. */
	function escanear(e: KeyboardEvent) {
		if (e.key !== 'Enter') return;
		const q = buscar.trim().toLowerCase();
		const p =
			productos.find((x) => x.sku.toLowerCase() === q) ??
			(visibles.length === 1 ? visibles[0] : null);
		if (p) {
			sumar(p.clave, 1);
			buscar = '';
		}
	}

	// ---- Cobro ----
	const MEDIOS = [
		['mp', I.phQrCodeBold, 'QR MP', '#00B1EA'],
		['efectivo', I.phMoneyFill, 'Efectivo', '#2F8F63'],
		['tarjeta', I.phCreditCardFill, 'Tarjeta', '#3E6BFD'],
		['modo', I.phDeviceMobileFill, 'MODO', '#00A88F'],
		['transf', I.phBankFill, 'Transfer.', '#6B5244'],
		['dividir', I.phSplitHorizontalFill, 'Dividir', '#C24E25']
	] as const;
	type Medio = Exclude<(typeof MEDIOS)[number][0], 'dividir'>;
	let medio = $state<(typeof MEDIOS)[number][0]>('mp');
	let cobrando = $state(false);
	let dividir = $state(false);
	let parte = $state(0);
	let otro = $state<Medio>('efectivo');
	let primero = $state<Medio>('mp');

	async function cobrar(pago: Medio, nota?: string) {
		if (!total || cobrando) return;
		cobrando = true;
		const r = await fetch('/panel/pedidos/nuevo', {
			method: 'POST',
			headers: { 'content-type': 'application/json' },
			body: JSON.stringify({
				items: renglones.map((x) => ({
					productoId: x.p.id,
					varianteId: x.p.varianteId,
					cantidad: x.q
				})),
				pago,
				canal: 'pos',
				envio: 'retiro',
				nota: nota ?? null
			})
		});
		cobrando = false;
		if (!r.ok) {
			const m = ((await r.json().catch(() => null)) as { message?: string } | null)?.message;
			return ui.toast({ k: 'err', title: 'No se cobró', desc: m ?? 'Probá de nuevo.' });
		}
		const { numero, total: t } = (await r.json()) as { numero: number; total: number };
		ticket.clear();
		dividir = false;
		await invalidateAll();
		ui.toast({
			k: 'ok',
			title: `Cobrado ${pesos(t)}`,
			desc: `${nota ?? MEDIOS.find((m) => m[0] === pago)![2]} · Factura B emitida`,
			action: 'Ver pedido',
			onAction: () => {
				// eslint-disable-next-line svelte/no-navigation-without-resolve -- ruta del panel con query
				void goto(`${rutaDe('Pedidos')}?pedido=${numero}`);
			}
		});
	}
	function alCobrar() {
		if (medio === 'dividir') {
			parte = Math.round(total / 2);
			dividir = true;
			return;
		}
		void cobrar(medio);
	}
	const nombreMedio = (m: Medio) => MEDIOS.find((x) => x[0] === m)![2];
</script>

<Seo
	title="Ventas en el local · Panel"
	description="Punto de venta"
	path="/panel/ventas-en-local"
	noindex
/>

<div class="pagina">
	<div class="cab" use:anim={{ t: 'rise' }}>
		<div class="c-txt">
			<h1>Ventas en el local</h1>
			<div class="c-s">
				{data.caja.abierta
					? `Caja abierta desde las ${horaMin(new Date(data.caja.abierta))}`
					: 'Caja cerrada'} · mismo stock que la web
			</div>
		</div>
		<span class="lector"><Icon icon={I.phBarcode} size={16} />Funciona con lector de códigos</span>
	</div>

	<div class="pos" use:anim={{ t: 'rise', delay: 60 }}>
		<div class="izq">
			<div class="busca">
				<label class="buscar"
					><Icon icon={I.phMagnifyingGlass} size={16} /><input
						bind:value={buscar}
						onkeydown={escanear}
						placeholder="Buscá o escaneá un producto"
						aria-label="Buscá o escaneá un producto"
					/></label
				>
				<div class="cats" role="group" aria-label="Categorías" use:deslizable>
					{#each categorias as c (c)}
						<button type="button" aria-pressed={categoria === c} onclick={() => (categoria = c)}
							>{c}</button
						>
					{/each}
				</div>
			</div>
			<div class="grilla">
				{#each visibles as p (p.clave)}
					{@const en = ticket.get(p.clave) ?? 0}
					<button
						type="button"
						class="prod"
						disabled={!p.stock}
						onclick={() => sumar(p.clave, 1)}
						aria-label="Sumar {p.nombre}"
					>
						<div class="foto">
							{#if p.foto}<img src={unsplash(p.foto, 300, '&h=250')} alt="" />{/if}
							{#if en}<span class="en">{en}</span>{/if}
						</div>
						<div class="p-txt">
							<div class="p-n">{p.nombre}</div>
							<div class="p-pie">
								<b>{pesos(p.precio)}</b><span class:poco={p.stock < 5}>{p.stock} u</span>
							</div>
						</div>
					</button>
				{:else}
					<div class="vacio">Nada con «{buscar}».</div>
				{/each}
			</div>
		</div>

		<div class="ticket">
			<div class="t-cab">
				<span>Ticket</span><button
					type="button"
					onclick={() => ticket.clear()}
					disabled={!ticket.size}>Vaciar</button
				>
			</div>
			<div class="t-items" data-lenis-prevent>
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
								class="mas"
								aria-label="Sumar {x.p.nombre}"
								onclick={() => sumar(x.p.clave, 1)}><Icon icon={I.phPlusBold} size={12} /></button
							>
						</span>
						<b class="i-t">{pesos(x.p.precio * x.q)}</b>
					</div>
				{:else}
					<div class="t-vacio">
						<Icon icon={I.phBarcode} size={28} />Tocá un producto o escanealo
					</div>
				{/each}
			</div>
			<div class="t-pie">
				<div class="total"><span>Total</span><b>{pesos(total)}</b></div>
				<div class="medios" role="group" aria-label="Medio de pago">
					{#each MEDIOS as [id, icono, t, col] (id)}
						<button type="button" aria-pressed={medio === id} onclick={() => (medio = id)}
							><Icon icon={icono} size={20} style="color:{col}" />{t}</button
						>
					{/each}
				</div>
				<button type="button" class="cobrar" disabled={!total || cobrando} onclick={alCobrar}
					>Cobrar {pesos(total)}</button
				>
				<div class="nota">La factura sale sola por ARCA · ticket por WhatsApp opcional</div>
			</div>
		</div>
	</div>
</div>

{#if ticket.size}
	<!-- En el celular el ticket queda debajo de los productos: esta barra lleva hasta él. -->
	<button
		type="button"
		class="ir-ticket"
		onclick={() => document.querySelector('.ticket')?.scrollIntoView({ behavior: 'smooth' })}
		>Ticket · {renglones.reduce((s, x) => s + x.q, 0)}
		{renglones.reduce((s, x) => s + x.q, 0) === 1 ? 'producto' : 'productos'}<b>{pesos(total)}</b
		></button
	>
{/if}

<Modal bind:open={dividir} titulo="Dividir el pago" ancho={440}>
	<form
		onsubmit={(e) => {
			e.preventDefault();
			const resto = total - parte;
			void cobrar(
				parte >= resto ? primero : otro,
				`${pesos(parte)} con ${nombreMedio(primero)} y ${pesos(resto)} con ${nombreMedio(otro)}`
			);
		}}
	>
		<div class="m-ico"><Icon icon={I.phSplitHorizontalFill} size={24} /></div>
		<div class="m-t">Dividir {pesos(total)}</div>
		<div class="m-d">Parte con un medio y el resto con otro.</div>
		<div class="m-filas">
			<label class="field"
				>Primer medio<select class="select" bind:value={primero}
					>{#each MEDIOS.filter((m) => m[0] !== 'dividir') as [id, , t] (id)}<option value={id}
							>{t}</option
						>{/each}</select
				></label
			>
			<label class="field"
				>Monto<input
					class="input"
					type="number"
					min="1"
					max={total - 1}
					bind:value={parte}
				/></label
			>
			<label class="field"
				>El resto ({pesos(Math.max(0, total - parte))}) con<select class="select" bind:value={otro}
					>{#each MEDIOS.filter((m) => m[0] !== 'dividir') as [id, , t] (id)}<option value={id}
							>{t}</option
						>{/each}</select
				></label
			>
		</div>
		<div class="m-acc">
			<button type="button" class="btn btn-secondary" onclick={() => (dividir = false)}
				>Cancelar</button
			>
			<button class="btn btn-primary" disabled={parte <= 0 || parte >= total || primero === otro}
				>Cobrar</button
			>
		</div>
	</form>
</Modal>

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
	.lector {
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
	.pos {
		display: grid;
		grid-template-columns: minmax(0, 1fr) 360px;
		gap: 16px;
		align-items: start;
	}
	.izq {
		display: flex;
		flex-direction: column;
		gap: 12px;
		min-width: 0;
	}
	.busca {
		display: flex;
		gap: 8px;
		flex-wrap: wrap;
	}
	.buscar {
		flex: 1 1 220px;
		display: flex;
		align-items: center;
		gap: 8px;
		padding: 11px 14px;
		border-radius: 999px;
		background: var(--sf);
		box-shadow: var(--sh);
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
	.cats {
		display: flex;
		gap: 6px;
		overflow-x: auto;
		scrollbar-width: none;
	}
	.cats button {
		border: 0;
		white-space: nowrap;
		font-size: 13px;
		font-weight: 600;
		padding: 10px 14px;
		border-radius: 999px;
		background: var(--sf);
		color: var(--tx);
		box-shadow: var(--sh);
	}
	.cats button[aria-pressed='true'] {
		background: var(--tx);
		color: var(--sf);
	}
	.grilla {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
		gap: 10px;
	}
	.prod {
		padding: 8px;
		border-radius: 18px;
		border: 0;
		background: var(--sf);
		box-shadow: var(--shd);
		color: var(--tx);
		text-align: left;
		transition: transform 0.15s;
	}
	.prod:active {
		transform: scale(0.96);
	}
	.prod:disabled {
		opacity: 0.5;
	}
	.foto {
		position: relative;
		aspect-ratio: 1.2;
		border-radius: 12px;
		overflow: hidden;
		background: var(--sf3);
	}
	.foto img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		display: block;
	}
	.en {
		position: absolute;
		top: 6px;
		right: 6px;
		min-width: 24px;
		height: 24px;
		padding: 0 7px;
		border-radius: 999px;
		background: #d2582f;
		color: #fff;
		font-size: 12px;
		font-weight: 700;
		display: grid;
		place-items: center;
	}
	.p-txt {
		padding: 8px 4px 2px;
	}
	.p-n {
		font-size: 13.5px;
		font-weight: 600;
		line-height: 1.25;
	}
	.p-pie {
		display: flex;
		justify-content: space-between;
		margin-top: 3px;
		font-size: 14px;
	}
	.p-pie span {
		font-size: 12px;
		color: var(--mu);
	}
	.p-pie span.poco {
		color: #b83a2e;
	}
	.vacio {
		padding: 28px;
		color: var(--mu);
		font-size: 14px;
	}
	.ticket {
		position: sticky;
		top: 80px;
		border-radius: 24px;
		background: var(--sf);
		box-shadow: var(--shd);
		display: flex;
		flex-direction: column;
		overflow: hidden;
	}
	.t-cab {
		display: flex;
		justify-content: space-between;
		align-items: center;
		padding: 16px 18px;
		border-bottom: 1px solid var(--ln);
		font-size: 16px;
		font-weight: 600;
	}
	.t-cab button {
		border: 0;
		background: transparent;
		font-size: 13px;
		font-weight: 600;
		color: var(--mu);
	}
	.t-items {
		padding: 6px 18px;
		max-height: 280px;
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
	.i-q button.mas {
		border: 0;
		background: var(--tx);
		color: var(--sf);
	}
	.i-q b {
		min-width: 16px;
		text-align: center;
	}
	.i-t {
		min-width: 74px;
		text-align: right;
		white-space: nowrap;
	}
	.t-vacio {
		padding: 28px 0;
		text-align: center;
		color: var(--mu);
		font-size: 14px;
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 6px;
	}
	.t-pie {
		padding: 14px 18px;
		display: flex;
		flex-direction: column;
		gap: 12px;
		border-top: 1px solid var(--ln);
		background: var(--sf2);
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
		font-size: 32px;
		font-weight: 600;
		letter-spacing: -0.04em;
	}
	.medios {
		display: grid;
		grid-template-columns: repeat(3, 1fr);
		gap: 6px;
	}
	.medios button {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 4px;
		padding: 9px 4px;
		border-radius: 14px;
		border: 1.5px solid var(--ln);
		background: var(--sf);
		font-size: 11.5px;
		font-weight: 600;
		color: var(--tx);
	}
	.medios button[aria-pressed='true'] {
		border-color: #d2582f;
	}
	.cobrar {
		padding: 15px;
		border-radius: 999px;
		border: 0;
		background: linear-gradient(180deg, #e06a3e, #d2582f);
		color: #fff;
		font-size: 15.5px;
		font-weight: 600;
		box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.3);
	}
	.cobrar:disabled {
		background: #c9b6a6;
	}
	.nota {
		font-size: 12px;
		color: var(--mu);
		text-align: center;
	}
	.m-ico {
		width: 48px;
		height: 48px;
		display: grid;
		place-items: center;
		border-radius: 15px;
		background: rgba(210, 88, 47, 0.12);
		color: #c24e25;
	}
	.m-t {
		margin-top: 16px;
		font-size: 21px;
		font-weight: 600;
		letter-spacing: -0.02em;
	}
	.m-d {
		margin-top: 6px;
		font-size: 14.5px;
		color: var(--mu);
	}
	.m-filas {
		margin-top: 16px;
		display: grid;
		gap: 12px;
	}
	.m-acc {
		margin-top: 20px;
		display: flex;
		gap: 8px;
		justify-content: flex-end;
	}
	.ir-ticket {
		display: none;
	}
	@media (max-width: 999px) {
		.ir-ticket {
			position: fixed;
			left: 50%;
			bottom: 92px;
			translate: -50% 0;
			z-index: 60;
			display: flex;
			align-items: center;
			gap: 14px;
			padding: 12px 18px;
			border-radius: 999px;
			border: 0;
			background: #2a140c;
			color: #fff;
			font-size: 14px;
			font-weight: 600;
			white-space: nowrap;
			box-shadow: 0 24px 50px -16px rgba(42, 20, 12, 0.7);
		}
		.pos {
			grid-template-columns: minmax(0, 1fr);
		}
		.ticket {
			position: static;
		}
	}
</style>
